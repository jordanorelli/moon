package moon

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
)

func parseArgs(args []string, dest interface{}) (map[string]interface{}, error) {
	reqs, err := requirements(dest)
	if err != nil {
		return nil, fmt.Errorf("unable to parse args: bad requirements: %s", err)
	}

	out := make(map[string]interface{})
	shorts := make(map[string]req, len(reqs))
	longs := make(map[string]req, len(reqs))
	for _, req := range reqs {
		if req.short != "" {
			shorts[req.short] = req
		}
		if req.long != "" {
			longs[req.long] = req
		}
	}

	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "help" {
			showHelp(dest)
		}
		if arg == "--" {
			break
		}
		if strings.HasPrefix(arg, "--") {
			arg = strings.TrimPrefix(arg, "--")

			var (
				key string
				val string
			)
			if strings.ContainsRune(arg, '=') {
				parts := strings.SplitN(arg, "=", 2)
				key, val = parts[0], parts[1]
			} else {
				key = arg
				i++
				if i >= len(args) {
					return nil, fmt.Errorf("terminal arg %s is missing a value", key)
				}
				val = args[i]
			}

			req, ok := longs[key]
			if !ok {
				// ignore unknown options silently?
				log.Printf("no such long opt: %s", key)
				continue
			}
			if req.t.Kind() == reflect.Bool {
				out[key] = true
				continue
			}

			d, err := ReadString(fmt.Sprintf("%s: %s", key, val)) // :(
			if err != nil {
				return nil, fmt.Errorf("unable to parse cli argument %s: %s", key, err)
			}
			out[key] = d.items[key]
		} else if strings.HasPrefix(arg, "-") {
			panic("i'm not doing short args yet")
		} else {
			break
		}
	}
	return out, nil
}

func showHelp(dest interface{}) {
	reqs, err := requirements(dest)
	if err != nil {
		panic(err)
	}

	for _, req := range reqs {
		req.writeHelpLine(os.Stdout)
	}
	os.Exit(1)
}
