package lib

import (
	"errors"
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/boltdb/bolt"
)

const DEFAULT_DB_FILE = "bolt.db"

var DB_FILE string = DEFAULT_DB_FILE

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type Database struct {
	db *bolt.DB
}

func (self *Database) Open(db_file string) error {
	if nil != self.db {
		self.Close()
	}
	db, err := bolt.Open(db_file, 0600, &bolt.Options{Timeout: 1 * time.Second})
	self.db = db
	return err
}

func (self *Database) Close() {
	self.db.Close()
}

func (self *Database) CreateTable(table_name string) error {
	return self.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(table_name))
		return err
	})
}

func (self *Database) Get(table, key, passphrase string) (string, error) {
	if nil == self.db {
		return "", errors.New("Database not opened")
	}
	var result string
	var err error
	return result, self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}

		v := b.Get(Sha512HashByte(key))
		// v := b.Get(ToByte(key))
		decompressed := DecompressByte(v)
		garbage := string(decompressed)
		if "" == garbage {
			return errors.New("Not found")
		}
		result, err = Decrypt(passphrase, garbage)

		if nil == err && !utf8.ValidString(result) {
			err = errors.New("Not utf-8")
		}

		return err
	})
}

func (self *Database) Set(table, key, value, passphrase string) error {
	if nil == self.db {
		return errors.New("Database not opened")
	}

	return self.db.Update(func(tx *bolt.Tx) error {
		garbage, err := Encrypt(passphrase, value)
		if nil != err {
			return err
		}

		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}

		compressed := CompressByte([]byte(garbage))
		return b.Put(Sha512HashByte(key), compressed)
		// return b.Put(ToByte(key), compressed)
	})
}

func (self *Database) Keys(table string) ([]string, error) {
	var result []string
	if nil == self.db {
		return result, errors.New("Database not opened")
	}
	return result, self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}
		return b.ForEach(func(k, v []byte) error {
			result = append(result, string(k))
			return nil
		})
	})
}

func OpenDb(db_file string) Database {
	bolt_db, err := bolt.Open(DB_FILE, 0600, &bolt.Options{Timeout: 1 * time.Second})
	checkError(err)
	db := Database{db: bolt_db}
	err = db.CreateTable("store")
	checkError(err)
	return db
}
