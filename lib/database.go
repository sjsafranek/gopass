package lib

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/boltdb/bolt"

	"github.com/sjsafranek/goutils/cryptic"
	"github.com/sjsafranek/goutils/minify"
	"github.com/sjsafranek/goutils/transformers"

	"github.com/schollz/golock"
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
	db    *bolt.DB
	glock *golock.Lock
}

func (self *Database) Open(db_file string) error {
	if nil != self.db {
		self.Close()
	}

	if !strings.HasSuffix(db_file, ".db") {
		db_file += ".db"
	}

	// first initiate lockfile
	lock_file := strings.Replace(db_file, ".db", ".lock", -1)
	self.glock = golock.New(
		golock.OptionSetName(lock_file),
		golock.OptionSetInterval(1*time.Millisecond),
		golock.OptionSetTimeout(60*time.Second),
	)

	// lock it
	err := self.glock.Lock()
	if err != nil {
		// error means we didn't get the lock
		// handle it
		panic(err)
	}
	//.end

	db, err := bolt.Open(db_file, 0600, &bolt.Options{Timeout: 1 * time.Second})
	self.db = db
	return err
}

func (self *Database) Close() {
	self.db.Close()

	// unlock it
	err := self.glock.Unlock()
	if err != nil {
		panic(err)
	}
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

		// v := b.Get(Sha512HashByte(key))
		v := b.Get(transformers.ToByte(key))
		decompressed := minify.DecompressByte(v)
		garbage := string(decompressed)
		if "" == garbage {
			return errors.New("Not found")
		}
		result, err = cryptic.Decrypt(passphrase, garbage)

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
		garbage, err := cryptic.Encrypt(passphrase, value)
		if nil != err {
			return err
		}

		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}

		compressed := minify.CompressByte([]byte(garbage))
		// return b.Put(Sha512HashByte(key), compressed)
		return b.Put(transformers.ToByte(key), compressed)
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

func (self *Database) Remove(table string, key string, passphrase string) error {
	return self.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", table)
		}

		err := bucket.Delete([]byte(key))
		if err != nil {
			return fmt.Errorf("Could not delete key: %s", err)
		}
		return err
	})
}

func (self *Database) Tables() ([]string, error) {
	var result []string
	if nil == self.db {
		return result, errors.New("Database not opened")
	}
	return result, self.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			result = append(result, string(name))
			return nil
		})
	})
}

func OpenDb(db_file string) Database {
	db := Database{}
	db.Open(db_file)
	err := db.CreateTable("store")
	checkError(err)
	return db
}
