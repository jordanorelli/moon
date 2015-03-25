package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func input() io.ReadCloser {
	if flag.Arg(1) == "" {
		return os.Stdin
	} else {
		f, err := os.Open(flag.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		return f
	}
}

func check() {
	r := input()
	defer r.Close()

	_, err := parse(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func lexx() {
	r := input()
	defer r.Close()

	c := lex(r)
	for t := range c {
		fmt.Println(t)
	}
}

func parsse() {
	r := input()
	defer r.Close()

	n, err := parse(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	if err := n.pretty(os.Stdout, ""); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
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
	case "":
		fmt.Fprintf(os.Stderr, "specify an action (check)\n")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "no such action: %s\n", flag.Arg(0))
		os.Exit(1)
	}
}
