package moon

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

const ()

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
		case t_string:
			n := new(stringNode)
			if err := n.parse(p); err != nil {
				return nil, err
			}
			return n, nil
		case t_real_number:
			n := new(numberNode)
			if err := n.parse(p); err != nil {
				return nil, err
			}
			return n, nil
		case t_list_start:
			p.next()
			n := new(listNode)
			if err := n.parse(p); err != nil {
				return nil, err
			}
			return n, nil
		case t_object_start:
			p.next()
			n := &objectNode{}
			if err := n.parse(p); err != nil {
				return nil, err
			}
			return n, nil
		case t_variable:
			n := new(variableNode)
			if err := n.parse(p); err != nil {
				return nil, err
			}
			return n, nil
		default:
			return nil, fmt.Errorf("parse error: unexpected %v token while looking for value", t.t)
		}
	}
}
