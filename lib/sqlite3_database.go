package lib

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sjsafranek/goutils/cryptic"
)

type Sqlite3Database struct {
	db *sql.DB
}

func (self *Sqlite3Database) Open(dbname string) error {
	connectionString := fmt.Sprintf("%v?cache=shared&mode=rwc&_busy_timeout=50000000", dbname)
	db, err := sql.Open("sqlite3", connectionString)
	self.db = db
	return err
}

func (self *Sqlite3Database) Close() error {
	return self.db.Close()
}

func (self *Sqlite3Database) CreateTable(table_name string) error {
	return self.execWriteQuery(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %v(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			create_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			update_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TRIGGER IF NOT EXISTS %v__update
			AFTER
			UPDATE
				ON %v
					FOR EACH ROW
						BEGIN
							UPDATE %v SET update_at=CURRENT_TIMESTAMP WHERE id=OLD.id;
						END;
		`, table_name, table_name, table_name, table_name))
}

func (self *Sqlite3Database) Get(table_name, key, passphrase string) (string, error) {

	stmt, err := self.db.Prepare(fmt.Sprintf("SELECT value FROM %v WHERE key=? ORDER BY id DESC LIMIT 1;", table_name))
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var garbage string
	err = stmt.QueryRow(key).Scan(&garbage)
	if err != nil {
		return "", err
	}

	return cryptic.Decrypt(passphrase, garbage)
}

func (self *Sqlite3Database) Keys(table_name string) ([]string, error) {
	results := []string{}

	rows, err := self.db.Query(fmt.Sprintf("SELECT DISTINCT key FROM %v;", table_name))
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return results, err
		}
		results = append(results, key)
	}

	return results, rows.Err()
}

func (self *Sqlite3Database) Set(table_name, key, value, passphrase string) error {
	garbage, err := cryptic.Encrypt(passphrase, value)
	if nil != err {
		return err
	}
	return self.execWriteQuery(fmt.Sprintf(`INSERT INTO %v(key, value) VALUES(?, ?);`, table_name), key, garbage)
}

func (self *Sqlite3Database) Del(table, key, passphrase string) error {
	return nil
}

func (self *Sqlite3Database) Tables() ([]string, error) {
	results := []string{}

	rows, err := self.db.Query(`SELECT tbl_name FROM sqlite_master WHERE type = 'table' AND name != 'sqlite_sequence';`)
	if err != nil {
		return results, err
	}

	defer rows.Close()
	for rows.Next() {
		var table_name string
		err = rows.Scan(&table_name)
		if err != nil {
			return results, err
		}

		results = append(results, table_name)
	}

	return results, rows.Err()
}

func (self *Sqlite3Database) execWriteQuery(query string, args ...interface{}) error {
	tx, err := self.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if nil != err {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
