package main

import (
	"fmt"
	"io"
)

const ()

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
	return <-p.input
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

// parse the next value.  This is to be executed in a context where we know we
// want something that is a value to come next, such as after an equals sign.
func (p *parser) parseValue() (interface{}, error) {
	for {
		t := p.next()
		switch t.t {
		case t_error:
			return nil, fmt.Errorf("parse error: saw lex error when looking for value: %v", t.s)
		case t_eof:
			return nil, fmt.Errorf("parse error: unexpected eof when looking for value")
		case t_string:
			return t.s, nil
		case t_list_start:
			l := new(list)

		SIN:
			if p.peek().t == t_list_end {
				p.next()
				return l, nil
			}

			if v, err := p.parseValue(); err != nil {
				return nil, err
			} else {
				l.append(v)
			}

			switch t := p.next(); t.t {
			case t_list_separator:
				goto SIN
			case t_list_end:
				return l, nil
			default:
				return nil, fmt.Errorf("parse error: unexpected %v token while scanning for list", t.t)
			}
		default:
			return nil, fmt.Errorf("parse error: unexpected %v token while looking for value", t.t)
		}
	}
}
