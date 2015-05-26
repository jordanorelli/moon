package moon

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

const ()

var nodes = map[tokenType]func(p *parser) node{
	t_string:           func(p *parser) node { return new(stringNode) },
	t_real_number:      func(p *parser) node { return new(numberNode) },
	t_imaginary_number: func(p *parser) node { return new(numberNode) },
	t_list_start:       func(p *parser) node { p.next(); return &listNode{} },
	t_object_start:     func(p *parser) node { p.next(); return &objectNode{} },
	t_variable:         func(p *parser) node { return new(variableNode) },
	t_bool:             func(p *parser) node { return new(boolNode) },
	t_duration:         func(p *parser) node { return new(durationNode) },
}

// Static path for configuration file. By default, a call to Parse wil look for
// a file at the path specified by Path.
var Path = ""

func bail(status int, t string, args ...interface{}) {
	if status == 0 {
		fmt.Fprintf(os.Stdout, t+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, t+"\n", args...)
	}
	os.Exit(status)
}

// Parse reads in all command line arguments in addition to parsing the Moon
// configuration file found at moon.Path. If moon.Path is not set, Parse does
// not automatically look for a configuration file.
//
// The command-line options as well as the values found in the Moon document
// will be used to fill the destination object pointed to by the dest argument.
// Clients may use struct tags to dictate how values will be filled in the
// destination struct. Within the struct tag, the client may provide a moon
// document describing how the value is to be parsed.
//
// The following tags are currently recognized:
//
//   - name: the value's name inside of the moon document
//   - help: help text to be printed when running your program as "program help"
//   - required: whether or not the specified field is required.
//   - default: default value for the given field
//   - short: single character to be used as a command-line flag
//   - long: a string of characters to be used as a command-line option
//
// Here's an example of a struct definition that is annotated to inform the
// Moon parser how to fill the struct with values from a Moon document.
//
//   var config struct {
//       Host string `
//       name: host
//       help: target host with whom we will connect
//       required: true
//       short: h
//       long: host
//       `
//
//       Port int `
//       name: port
//       help: port to dial on the target host
//       default: 12345
//       short: p
//       long: port
//       `
//   }
//
// .. and here's how we would parse our command-line arguments and config file
// at path "./config.moon" to fill the struct at config:
//
//   moon.Path = "./config"
//   moon.Parse(&config)
//
// Any value provied as a command-line argument will override the value
// supplied by the config file at "./config"
func Parse(dest interface{}) *Doc {
	cliArgs, err := parseArgs(os.Args, dest)
	if err != nil {
		bail(1, "unable to parse cli args: %s", err)
	}

	var doc *Doc

	if Path != "" {
		f, err := os.Open(Path)
		if err == nil {
			defer f.Close()
			d, err := Read(f)
			if err != nil {
				bail(1, "unable to parse moon config file at path %s: %s", Path, err)
			}
			doc = d
		}
	}

	if doc == nil {
		doc = &Doc{items: make(map[string]interface{})}
	}

	for k, v := range cliArgs {
		doc.items[k] = v
	}

	if err := doc.Fill(dest); err != nil {
		bail(1, "unable to fill moon config values: %s", err)
	}
	return doc
}

// Reads a moon document from a given io.Reader. The io.Reader is advanced to
// EOF. The reader is not closed after reading, since it's an io.Reader and not
// an io.ReadCloser. In the event of error, the state that the source reader
// will be left in is undefined.
func Read(r io.Reader) (*Doc, error) {
	tree, err := parse(r)
	if err != nil {
		return nil, err
	}
	ctx := newContext()
	if _, err := tree.eval(ctx); err != nil {
		return nil, fmt.Errorf("eval error: %s\n", err)
	}
	return &Doc{items: ctx.public}, nil
}

// Reads a moon document from a string. This is purely a convenience method;
// all it does is create a buffer and call the moon.Read function.
func ReadString(source string) (*Doc, error) {
	return Read(strings.NewReader(source))
}

// Reads a moon document from a slice of bytes. This is purely a concenience
// method; like ReadString, it simply creates a buffer and calls moon.Read
func ReadBytes(b []byte) (*Doc, error) {
	return Read(bytes.NewBuffer(b))
}

func ReadFile(path string) (*Doc, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Read(f)
}

func parse(r io.Reader) (node, error) {
	p := &parser{
		root:   newRootNode(),
		input:  lex(r),
		backup: make([]token, 0, 8),
	}
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p.root, nil
}

// parser (little p) is an actual parser.  It actually does the parsing of a
// moon document.
type parser struct {
	root   node
	input  chan token
	backup []token
}

func (p *parser) parse() error {
	if p.root == nil {
		p.root = newRootNode()
	}
	return p.root.parse(p)
}

// returns the next token and advances the input stream
func (p *parser) next() token {
	if len(p.backup) > 0 {
		t := p.backup[len(p.backup)-1]
		p.backup = p.backup[:len(p.backup)-1]
		return t
	}
SKIP_COMMENTS:
	t, ok := <-p.input
	if !ok {
		return token{t_eof, "eof"}
	}
	if t.t == t_comment {
		goto SKIP_COMMENTS
	}
	return t
}

func (p *parser) peek() token {
	t := p.next()
	p.unread(t)
	return t
}

func (p *parser) unread(t token) {
	if p.backup == nil {
		p.backup = make([]token, 0, 8)
	}
	p.backup = append(p.backup, t)
}

func (p *parser) ensureNext(tt tokenType, context string) error {
	if p.peek().t != tt {
		return fmt.Errorf("unexpected %v in %s: expected %v", p.peek().t, context, tt)
	}
	return nil
}

// parse the next value.  This is to be executed in a context where we know we
// want something that is a value to come next, such as after an equals sign.
func (p *parser) parseValue() (node, error) {
	for {
		t := p.peek()
		switch t.t {
		case t_error:
			return nil, fmt.Errorf("parse error: saw lex error when looking for value: %v", t.s)
		case t_eof:
			return nil, fmt.Errorf("parse error: unexpected eof when looking for value")
		}

		fn, ok := nodes[t.t]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %v token while looking for value", t.t)
		}
		n := fn(p)
		if err := n.parse(p); err != nil {
			return nil, err
		}
		return n, nil
	}
}
