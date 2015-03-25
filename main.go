package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

	_, err := parse(r)
	if err != nil {
		bail(1, "parse error: %s", err)
	}
}

func lexx() {
	r := input(1)
	defer r.Close()

	c := lex(r)
	for t := range c {
		fmt.Println(t)
	}
}

func parsse() {
	r := input(1)
	defer r.Close()

	n, err := parse(r)
	if err != nil {
		bail(1, "parse error: %s", err)
	}
	if err := n.pretty(os.Stdout, ""); err != nil {
		bail(1, "output error: %s", err)
	}
}

func eval(r io.Reader) (map[string]interface{}, error) {
	n, err := parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse error: %s\n", err)
	}

	ctx := make(map[string]interface{})
	if err := n.eval(ctx); err != nil {
		return nil, fmt.Errorf("eval error: %s\n", err)
	}
	return ctx, nil
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
	v, err := eval(input(n))
	if err != nil {
		bail(1, "input error: %s", err)
	}
	b, err := json.MarshalIndent(v, "", "    ")
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
	case "lex":
		lexx()
	case "parse":
		parsse()
	case "to":
		to()
	case "":
		bail(1, "must specify an action.\nvalid actions: check lex parse to")
	default:
		bail(1, "no such action:%s", flag.Arg(0))
	}
}
