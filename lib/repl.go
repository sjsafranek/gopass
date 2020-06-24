package lib

import (
	"fmt"
	"io"
	"log"
	"errors"
	"strings"

	"github.com/chzyer/readline"
)

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
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func Repl(db Database) error {

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
	var namespace string = "store"

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

		response := NewResponse()

		switch {

		case strings.HasPrefix(command, "setnamespace"):
			if 2 == len(parts) {
				namespace = parts[1]
				continue
			} else {
				response.SetError(errors.New("Incorrect usage"))
			}

		case strings.HasPrefix(command, "setpassphrase"):
			pswd, err := l.ReadPasswordWithConfig(setPasswordCfg)
			if err == nil {
				passphrase = string(pswd)
			}

		case strings.HasPrefix(command, "getnamespace"):
			response.Data.Namespace = namespace

		case strings.HasPrefix(command, "getpassphrase"):
			response.Data.Passphrase = passphrase

		case strings.HasPrefix(command, "del"):
			var key string

			if 2 == len(parts) {
				key = parts[1]
				err := db.Del(namespace, key, passphrase)
				if nil != err {
					response.SetError(err)
				}
				continue
			}
			log.Println("Error! Incorrect usage")
			log.Println("DEL <key>")

		case strings.HasPrefix(command, "get"):
			var key string

			if 2 == len(parts) {
				if "get" == command {
					key = parts[1]
					value, err := db.Get(namespace, key, passphrase)
					if nil != err {
						response.SetError(err)
					}
					response.Data.Key = key
					response.Data.Value = value
					break
				}
			}

			response.SetError(errors.New("Incorrect usage"))

		case strings.HasPrefix(command, "set"):
			var key string
			var value string

			if "set" == command {
				key = parts[1]

				i1 := strings.Index(line, "'")
				i2 := strings.LastIndex(line, "'")
				value = line[i1+1 : i2]

				err := db.Set(namespace, key, value, passphrase)
				if nil != err {
					response.SetError(err)
				}

				break
			}

			response.SetError(errors.New("Incorrect usage"))

		case command == "help":
			usage(l.Stderr())

		case strings.HasPrefix(command, "keys"):
			results, err := db.Keys(namespace)
			if nil != err {
				response.SetError(err)
			}
			response.Data.Keys = &results

		case strings.HasPrefix(command, "namespaces"):
			results, err := db.Tables()
			if nil != err {
				response.SetError(err)
			}
			response.Data.Namespaces = &results

		case command == "bye":
			goto exit

		case command == "exit":
			goto exit

		case command == "quit":
			goto exit

		case line == "":
			continue

		default:
			// log.Println("you said:", strconv.Quote(line))
		}

		response.Print()
	}
exit:

	return nil
}
