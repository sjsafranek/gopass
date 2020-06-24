package lib

import (
	"errors"
	"fmt"
	"io"
	"log"
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

	api := NewApi(db)

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
			response.Data = &ResponseData{Namespace: namespace}

		case strings.HasPrefix(command, "getpassphrase"):
			response.Data= &ResponseData{Passphrase: passphrase}

		case "del" == command:

			if 2 != len(parts) {
				response.SetError(errors.New("Incorrect usage"))
				break
			}

			resp, err := api.Do(&Request{
				Method: "delete",
				Params: RequestParams{
					Namespace:  namespace,
					Key:        parts[1],
					Passphrase: passphrase,
				},
			})
			if nil != err {
				panic(err)
			}
			response = resp

		case "get" == command:

			if 2 != len(parts) {
				response.SetError(errors.New("Incorrect usage"))
				break
			}

			resp, err := api.Do(&Request{
				Method: "get",
				Params: RequestParams{
					Namespace:  namespace,
					Key:        parts[1],
					Passphrase: passphrase,
				},
			})
			if nil != err {
				panic(err)
			}
			response = resp

		case "set" == command:


			i1 := strings.Index(line, "'")
			i2 := strings.LastIndex(line, "'")
			if i1 == i2 {
				response.SetError(errors.New("Incorrect usage"))
				break
			}

			resp, err := api.Do(&Request{
				Method: "set",
				Params: RequestParams{
					Namespace:  namespace,
					Key:        parts[1],
					Value:      line[i1+1 : i2],
					Passphrase: passphrase,
				},
			})
			if nil != err {
				panic(err)
			}
			response = resp

		case "do" == command:
			i1 := strings.Index(line, "'")
			i2 := strings.LastIndex(line, "'")

			if i1 == i2 {
				response.SetError(errors.New("Incorrect usage"))
				break
			}

			request := line[i1+1 : i2]
			resp, err := api.DoJSON(request)
			if nil != err {
				resp.SetError(err)
			}
			response = resp

		case command == "help":
			usage(l.Stderr())

		case "keys" == command:
			resp, err := api.Do(&Request{
				Method: "get_keys",
				Params: RequestParams{
					Namespace: namespace,
				},
			})
			if nil != err {
				panic(err)
			}
			response = resp

		case strings.HasPrefix(command, "namespaces"):
			resp, err := api.Do(&Request{
				Method: "get_namespaces",
			})
			if nil != err {
				panic(err)
			}
			response = resp

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
