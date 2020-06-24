package lib

import (
	"errors"
)

type Database interface {
	Tables() ([]string, error)
	Del(string, string, string) error
	Keys(string) ([]string, error)
	Set(string, string, string, string) error
	Get(string, string, string) (string, error)
	CreateTable(string) error
	Close() error
	Open(string) error
}

func OpenDb(db_file, db_engine string) (Database, error) {
	var db Database
	var err error

	switch db_engine {
	case "bolt":
		db = &BoltDatabase{}
		db.Open(db_file)
		err = db.CreateTable("store")
	case "sqlite3":
		db = &Sqlite3Database{}
		db.Open(db_file)
		err = db.CreateTable("store")
	default:
		err = errors.New("Not Supported")
	}

	return db, err
}
