package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sjsafranek/gopass/lib"
	"github.com/sjsafranek/goutils/cryptic"
)

const (
	DEFAULT_DB_ENGINE = "bolt"
)

var (
	DB_ENGINE string = DEFAULT_DB_ENGINE
)

func main() {
	// get command line args
	flag.StringVar(&DB_ENGINE, "e", DEFAULT_DB_ENGINE, "database engine (bolt, sqlite3)")
	flag.Parse()
	args := flag.Args()

	result, err := (func() (string, error) {

		fmt.Println(DB_ENGINE)
		db, err := lib.OpenDb(fmt.Sprintf("cryptic.%v", DB_ENGINE),DB_ENGINE)
		if nil != err {
			return "", err
		}
		defer db.Close()

		switch strings.ToLower(args[0]) {

		case "get":
			return db.Get("store", args[1], args[2])

		case "set":
			return "success", db.Set("store", args[1], args[2], args[3])

		case "encrypt":
			return cryptic.Encrypt(args[2], args[1])

		case "decrypt":
			return cryptic.Decrypt(args[2], args[1])

		case "ui":
			return lib.Run(db)

		default:
			return "", lib.RunRepl(db)
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
