package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"./lib"
)

const DEFAULT_DB_FILE = "bolt.db"

var DB_FILE string = DEFAULT_DB_FILE

func main() {
	// get command line args
	flag.StringVar(&DB_FILE, "db", DEFAULT_DB_FILE, "database file")
	flag.Parse()
	args := flag.Args()

	result, err := (func() (string, error) {

		switch strings.ToLower(args[0]) {

		case "get":
			db := lib.OpenDb(DB_FILE)
			defer db.Close()
			return db.Get("store", args[1], args[2])

		case "set":
			db := lib.OpenDb(DB_FILE)
			defer db.Close()
			return "success", db.Set("store", args[1], args[2], args[3])

		case "encrypt":
			return lib.Encrypt(args[2], args[1])

		case "decrypt":
			return lib.Decrypt(args[2], args[1])

		case "ui":
			db := lib.OpenDb(DB_FILE)
			defer db.Close()
			return lib.Run(db)

		default:
			return "", errors.New("Unknown command")
		}

	})()

	if nil != err {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(result)

}

// go get github.com/boltdb/bolt/...
// go get github.com/coreos/bbolt/...
