package moon

import (
	"fmt"
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
				return nil, fmt.Errorf("unrecognized long opt: %s", key)
			}
			if req.t.Kind() == reflect.Bool {
				out[req.name] = true
				continue
			}

			// this is horrible
			d, err := ReadString(fmt.Sprintf("%s: %s", key, val)) // :(
			if err != nil {
				return nil, fmt.Errorf("unable to parse cli argument %s: %s", key, err)
			}
			out[req.name] = d.items[key]
		} else if strings.HasPrefix(arg, "-") {
			arg = strings.TrimPrefix(arg, "-")
			if strings.ContainsRune(arg, '=') {
				runes := []rune(arg)
				if len(runes) == 1 { // -=
					// no clue what to do here
					return nil, fmt.Errorf("unable to parse cli arguments: weird -=?")
				}
				if runes[1] != '=' {
					return nil, fmt.Errorf("you may only use one short flag with an equals sign")
				}
				req, ok := shorts[string(runes[0])]
				if !ok {
					return nil, fmt.Errorf("unrecognized short opt: %c", runes[0])
				}
				d, err := ReadString(fmt.Sprintf("key: %s", runes[2:]))
				if err != nil {
					return nil, fmt.Errorf("unable to parse cli argument %c: %s", runes[0], err)
				}
				out[req.name] = d.items["key"]
			} else {
				runes := []rune(arg)
				for j := 0; j < len(runes); j++ {
					r := runes[j]
					req, ok := shorts[string(r)]
					if !ok {
						return nil, fmt.Errorf("unrecognized short opt: %c", r)
					}
					if req.t.Kind() == reflect.Bool {
						out[req.name] = true
						continue
					}
					if j != len(runes)-1 {
						// what a totally fucking preposterous error message
						return nil, fmt.Errorf("illegal short opt: %c: a "+
							"non-boolean short flag may only appear as the"+
							" terminal option in a run of short opts", r)
					}
					i++
					if i >= len(args) {
						return nil, fmt.Errorf("arg %s is missing a value", req.name)
					}
					val := args[i]
					d, err := ReadString(fmt.Sprintf("key: %s", val))
					if err != nil {
						return nil, fmt.Errorf("error parsing cli arg %s: %s", req.name, err)
					}
					out[req.name] = d.items["key"]
				}
			}
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
