package lib


import (
	// "flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
)

// const (
// 	DEFAULT_DATABASE_SERVER_ADDRESS = "localhost:9622"
// )
//
// var (
// 	DATABASE_SERVER_ADDRESS = DEFAULT_DATABASE_SERVER_ADDRESS
// )

// func init() {
// 	// get command line args
// 	flag.StringVar(&DATABASE_SERVER_ADDRESS, "a", DEFAULT_DATABASE_SERVER_ADDRESS, "database server address")
// 	flag.Parse()
// }

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("KEYS"),
	readline.PcItem("NAMESPACES"),
	readline.PcItem("SET"),
	readline.PcItem("GET"),
	readline.PcItem("DEL"),
	readline.PcItem("BYE"),
	readline.PcItem("EXIT"),
	readline.PcItem("HELP"),
	readline.PcItem("SETNAMESPACE"),
	readline.PcItem("SETPASSPHRASE"),
	readline.PcItem("GETNAMESPACE"),
	readline.PcItem("GETPASSPHRASE"),
	readline.PcItem("SETCLIENTENCRYPTION"),
	readline.PcItem("CONNECT"),
	readline.PcItem("DISCONNECT"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func RunRepl(db Database) error {

	l, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[31m[skeleton]#\033[0m ",
		HistoryFile:         "history.skeleton",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		return err
	}
	defer l.Close()

	var passphrase string
    var table_name string = "store"

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")
		command := strings.ToLower(parts[0])

		// testing
		setPasswordCfg := l.GenPasswordConfig()
		setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
			l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
			l.Refresh()
			return nil, 0, false
		})
		//.end

		switch {

		// case strings.HasPrefix(command, "connect"):
		// 	address := parts[1]
		// 	err := db.Open(address)
		// 	if nil != err {
		// 		log.Println(err)
		// 	}
        //
		// case strings.HasPrefix(command, "disconnect"):
		// 	db.Close()

		case strings.HasPrefix(command, "setnamespace"):
			if 2 == len(parts) {
				table_name = parts[1]
				continue
			}
			log.Println("Error! Incorrect usage")
			log.Println("SETNAMESPACE <namespace>")

		case strings.HasPrefix(command, "setpassphrase"):
			pswd, err := l.ReadPasswordWithConfig(setPasswordCfg)
			if err == nil {
				passphrase = string(pswd)
			}

		case strings.HasPrefix(command, "getnamespace"):
			log.Println(table_name)

		case strings.HasPrefix(command, "getpassphrase"):
			log.Println(passphrase)

		// case strings.HasPrefix(command, "del"):
		// 	var key string
        //
		// 	if 2 == len(parts) {
		// 		key = parts[1]
		// 		result, err := client.Del(key, passphrase)
		// 		if nil != err {
		// 			log.Println(err)
		// 			continue
		// 		}
		// 		fmt.Println(result)
		// 		continue
		// 	}
		// 	log.Println("Error! Incorrect usage")
		// 	log.Println("DEL <key>")

		case strings.HasPrefix(command, "get"):
			var key string

			if 2 == len(parts) {
				if "get" == command {
					key = parts[1]
					value, err := db.Get(table_name, key, passphrase)
					if nil != err {
						log.Println(err)
						continue
					}
					fmt.Println(value)
					continue
				}
			}
			log.Println("Error! Incorrect usage")
			log.Println("GET <key>")

		case strings.HasPrefix(command, "set"):
			var key string
			var value string

			if "set" == command {
				key = parts[1]

				i1 := strings.Index(line, "'")
				i2 := strings.LastIndex(line, "'")
				value = line[i1+1 : i2]

				err := db.Set(table_name, key, value, passphrase)
				if nil != err {
					log.Println(err)
					continue
				}
				fmt.Println("ok")
				continue
			}

			log.Println("Error! Incorrect usage")
			log.Println("SET <key> <value>")

		case command == "help":
			usage(l.Stderr())

		case strings.HasPrefix(command, "keys"):
			results, err := db.Keys(table_name)
			if nil != err {
				log.Println(err)
				continue
			}

			for i := 0; i < len(results); i++ {
				fmt.Printf("%v) %v\n", i, results[i])
			}

		case strings.HasPrefix(command, "namespaces"):
			results, err := db.Tables()
			if nil != err {
				log.Println(err)
				continue
			}

			for i := 0; i < len(results); i++ {
				fmt.Printf("%v) %v\n", i+1, results[i])
			}

		case command == "bye":
			goto exit

		case command == "exit":
			goto exit

		case command == "quit":
			goto exit

		case line == "":
		default:
			// log.Println("you said:", strconv.Quote(line))
		}
	}
exit:

return nil
}
