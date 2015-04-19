package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jordanorelli/moon/lib"
	"io"
	"os"
)

func input(n int) io.ReadCloser {
	if flag.Arg(n) == "" {
		return os.Stdin
	} else {
		f, err := os.Open(flag.Arg(n))
		if err != nil {
			bail(1, "input error: %s", err)
		}
		return f
	}
}

func check() {
	r := input(1)
	defer r.Close()

	_, err := moon.Read(r)
	if err != nil {
		bail(1, "parse error: %s", err)
	}
}

func to() {
	switch flag.Arg(1) {
	case "json":
		to_json(2)
	default:
		fmt.Fprintf(os.Stderr, "%s is not a valid output format\n", flag.Arg(1))
		fmt.Fprintln(os.Stderr, "valid output formats: json")
		os.Exit(1)
	}
}

func to_json(n int) {
	conf, err := moon.Read(input(n))
	if err != nil {
		bail(1, "input error: %s", err)
	}
	b, err := json.MarshalIndent(conf, "", "    ")
	if err != nil {
		bail(1, "encode error: %s", err)
	}
	os.Stdout.Write(b)
}

func bail(status int, t string, args ...interface{}) {
	var w io.Writer
	if status == 0 {
		w = os.Stdout
	} else {
		w = os.Stderr
	}
	fmt.Fprintf(w, t+"\n", args...)
	os.Exit(status)
}

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "check":
		check()
	case "to":
		to()
	case "":
		bail(1, "must specify an action.\nvalid actions: check lex parse to")
	default:
		bail(1, "no such action:%s", flag.Arg(0))
	}
}
