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
	t, ok := <-p.input
	if !ok {
		return token{t_eof, "eof"}
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
			return p.parseList(new(list))
		case t_object_start:
			return p.parseObject(make(object))
		default:
			return nil, fmt.Errorf("parse error: unexpected %v token while looking for value", t.t)
		}
	}
}

func (p *parser) parseList(l *list) (*list, error) {
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
		return p.parseList(l)
	case t_list_end:
		return l, nil
	default:
		return nil, fmt.Errorf("parse error: unexpected %v token while scanning for list", t.t)
	}
}

func (p *parser) parseObject(obj object) (object, error) {
	if p.peek().t == t_object_end {
		p.next()
		return obj, nil
	}
	if err := p.ensureNext(t_name, "looking for object field name in parseObject"); err != nil {
		return nil, err
	}
	field_name := p.next().s
	if err := p.ensureNext(t_object_separator, "looking for object separator in parseObject"); err != nil {
		return nil, err
	}
	p.next()

	if v, err := p.parseValue(); err != nil {
		return nil, err
	} else {
		obj[field_name] = v
	}

	switch t := p.next(); t.t {
	case t_list_separator:
		return p.parseObject(obj)
	case t_object_end:
		return obj, nil
	default:
		return nil, fmt.Errorf("parse error: unexpected %v token while scanning for object", t.t)
	}
}
