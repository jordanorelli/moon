package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func check(r io.Reader) {
	_, err := parse(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		check(os.Stdin)
	}
}
