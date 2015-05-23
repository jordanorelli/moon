/*
Command line utility for the Moon configuration language.

Purpose

The moon utility (moon) is a command line tool for working with moon files.  It
can be used to read values from moon files, evaluate moon files to remove
variables, convert moon files to json where possible, and verify that a given
moon file is syntactically valid.

The following is taken to be the contents of a file named ex.moon:

  # ------------------------------------------------------------------------------
  # example config format
  #
  # this is a working draft of things that are valid in a new config language to
  # replace json as a config language for Go projects.
  #
  # comments are a thing now!
  # ------------------------------------------------------------------------------

  # the whole document is implicitly a namespace, so you can set key value pairs
  # at the top level.
  first_name: jordan
  last_name: orelli

  # the bare strings true and false are boolean values
  bool_true: true
  bool_false: false

  # the quoted strings "true" and "false" are string values.  In the unlikely
  # event you need literal true and false strings, quote them.
  string_true: "true"
  string_false: "false"

  # lists of things should be supported
  items: [
      one
      2
      3.4
      [five; 6 7.8]
  ]

  # objects should be supported
  hash: {key: value; other_key: other_value}

  other_hash: {
      key_1: one
      key_2: 2
      key_3: 3.4
      key_4: [five; 6 7.8]
  }

  # we may reference an item that was defined earlier using a sigil
  repeat_hash: @hash

  # items can be hidden.  i.e., they're only valid in the parse and eval stage as
  # intermediate values internal to the config file; they are *not* visible to
  # the host program.  This is generally useful for composing larger, more
  # complicated things.
  @hidden_item: it has a value
  visible_item: @hidden_item

  @person_one: {
      name: the first name here
      age: 28
      hometown: crooklyn
  }

  @person_two: {
      name: the second name here
      age: 30
      hometown: tha bronx
  }

  people: [@person_one @person_two]

Subcommands

check:  used to syntax check a given file.  To check a given file, in this case ex.moon, to see if it is syntactically valid, one would invoke the following command:

  moon check ex.moon

If the file is syntactically valid, moon will print nothing and exit with a
status of 0.  If the file is invalid, moon will print a cryptic error message
that won't help you fix the file and exit with a status of 1.

eval:  evaluates a given moon file.  The file is parsed and evaluated, and its result is printed on stdout, itself in the moon format.  Invoking the following command:

  moon eval ex.moon

would produce the following output:

  first_name: "jordan"
  items: ["one" 2 3.4 ["five" 6 7.8]]
  other_hash: {key_3: 3.4 key_4: ["five" 6 7.8] key_1: "one" key_2: 2}
  repeat_hash: {key: "value" other_key: "other_value"}
  people: [{age: 28 hometown: "crooklyn" name: "the first name here"} {hometown: "tha bronx" name: "the second name here" age: 30}]
  last_name: "orelli"
  hash: {key: "value" other_key: "other_value"}
  visible_item: "it has a value"

to:  used to convert moon files to other formats.  Right now, the only
supported format is json.  To convert a given moon file to json, one would invoke the following command:

  moon to json ex.moon

If the file is valid and can be converted to json (i.e., only involves types
that are also supported by json or readily convertible to json types), the
file's json representation will be printed on stdout.

The "to" subcommand can also take its input from stdin by omitting a file name, as in one of the following invocations:

  cat ex.moon | moon to json
  moon to json < ex.moon

Either invocation would have the same effect.

get:  evaluates a moon file and retrieves a value, printing it on stdout out.
The file is parsed and evaluated, meaning that get will fail if executed
against an invalid moon file; this is a search on an evaluated moon document,
not a textual search.  The general format of invocation is as follows:

  moon get $search_term $file

Where $search_term refers to a key or a path, and $file refers to the name of a
file to be searched.  Given the moon file ex.moon:

  > moon get first_name ex.moon
  "jordan"

  > moon get visible_item ex.moon
  "it has a value"

  > moon get people ex.moon
  [{hometown: "crooklyn" name: "the first name here" age: 28} {name: "the second name here" age: 30 hometown: "tha bronx"}]

The search term may involve a path, allowing one to reach into an Object or List and retrieve individual items:

  > moon get hash/other_key ex.moon
  "other_value"

  > moon get people/1/name ex.moon
  "the second name here"

*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jordanorelli/moon"
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
	in := input(n)
	defer in.Close()
	doc, err := moon.Read(in)
	if err != nil {
		bail(1, "input error: %s", err)
	}
	b, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		bail(1, "encode error: %s", err)
	}
	os.Stdout.Write(b)
}

func get() {
	docpath := flag.Arg(1)
	in := input(2)
	defer in.Close()

	doc, err := moon.Read(in)
	if err != nil {
		bail(1, "input error: %s", err)
	}
	var v interface{}
	if err := doc.Get(docpath, &v); err != nil {
		bail(1, "error reading value at path %s: %s", docpath, err)
	}
	b, err := moon.Encode(v)
	if err != nil {
		bail(1, "error encoding value: %s", err)
	}
	os.Stdout.Write(b)
}

func eval() {
	in := input(1)
	defer in.Close()

	doc, err := moon.Read(in)
	if err != nil {
		bail(1, "input error: %s", err)
	}
	b, err := moon.Encode(doc)
	if err != nil {
		bail(1, "output error: %s", err)
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
	case "get":
		get()
	case "eval":
		eval()
	case "":
		bail(1, "must specify an action.\nvalid actions: check to get eval")
	default:
		bail(1, "no such action:%s", flag.Arg(0))
	}
}
