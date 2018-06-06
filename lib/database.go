package lib

import (
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	// "flag"
	"fmt"
	"io"
	"os"
	// "strings"
	"time"

	"github.com/boltdb/bolt"
)

const DEFAULT_DB_FILE = "bolt.db"

var DB_FILE string = DEFAULT_DB_FILE

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetSha512Hash(text string) string {
	hasher := sha512.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(key []byte, message string) (encmess string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//returns to base64 encoded string
	encmess = base64.URLEncoding.EncodeToString(cipherText)
	return
}

func decrypt(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess = string(cipherText)
	return
}

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

func (self *Database) hashKey(key string) []byte {
	return []byte(GetSha512Hash(key))
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

		v := b.Get(self.hashKey(key))
		decompressed := self.decompressByte(v)
		garbage := string(decompressed)
		if "" == garbage {
			return errors.New("Not found")
		}
		result, err = Decrypt(passphrase, garbage)
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

		compressed := self.compressByte([]byte(garbage))
		return b.Put(self.hashKey(key), compressed)
	})
}

// compressByte compresses byte
func (self *Database) compressByte(src []byte) []byte {
	compressedData := new(bytes.Buffer)
	self.compress(src, compressedData, 9)
	return compressedData.Bytes()
}

// decompressByte compresses byte
func (self *Database) decompressByte(src []byte) []byte {
	compressedData := bytes.NewBuffer(src)
	deCompressedData := new(bytes.Buffer)
	self.decompress(compressedData, deCompressedData)
	return deCompressedData.Bytes()
}

// compress
func (self *Database) compress(src []byte, dest io.Writer, level int) {
	compressor, _ := flate.NewWriter(dest, level)
	compressor.Write(src)
	compressor.Close()
}

// decompress
func (self *Database) decompress(src io.Reader, dest io.Writer) {
	decompressor := flate.NewReader(src)
	io.Copy(dest, decompressor)
	decompressor.Close()
}

func OpenDb(db_file string) Database {
	bolt_db, err := bolt.Open(DB_FILE, 0600, &bolt.Options{Timeout: 1 * time.Second})
	checkError(err)
	db := Database{db: bolt_db}
	err = db.CreateTable("store")
	checkError(err)
	return db
}

func hashPassphase(passphrase string) []byte {
	return []byte(GetMD5Hash(passphrase))
}

func Encrypt(passphrase, message string) (string, error) {
	return encrypt(hashPassphase(passphrase), message)
}

func Decrypt(passphrase, garbage string) (string, error) {
	return decrypt(hashPassphase(passphrase), garbage)
}

//
//
// func main() {
// 	// get command line args
// 	flag.StringVar(&DB_FILE, "db", DEFAULT_DB_FILE, "database file")
// 	flag.Parse()
// 	args := flag.Args()
//
// 	result, err := (func() (string, error) {
//
// 		switch strings.ToLower(args[0]) {
//
// 		case "get":
// 			db := OpenDb(DB_FILE)
// 			defer db.Close()
// 			return db.Get("store", args[1], args[2])
//
// 		case "set":
// 			db := OpenDb(DB_FILE)
// 			defer db.Close()
// 			return "success", db.Set("store", args[1], args[2], args[3])
//
// 		case "encrypt":
// 			return Encrypt(args[2], args[1])
//
// 		case "decrypt":
// 			return 	Decrypt(args[2], args[1])
//
// 		default:
// 			return "", errors.New("Unknown command")
// 		}
//
// 	})()
//
// 	if nil != err {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
//
// 	fmt.Println(result)
//
// }
//
// // go get github.com/boltdb/bolt/...
// // go get github.com/coreos/bbolt/...
