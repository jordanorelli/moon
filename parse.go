package main

import (
	"fmt"
	"io"
	"strconv"
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
		case t_real_number, t_imaginary_number:
			p.unread(t)
			return p.number()
		case t_list_start:
			return p.parseList(make(list, 0, 4))
		case t_object_start:
			return p.parseObject(make(object))
		default:
			return nil, fmt.Errorf("parse error: unexpected %v token while looking for value", t.t)
		}
	}
}

func (p *parser) parseList(l list) (list, error) {
	if p.peek().t == t_list_end {
		p.next()
		return l, nil
	}

	if v, err := p.parseValue(); err != nil {
		return nil, err
	} else {
		l = append(l, v)
	}

	switch t := p.peek(); t.t {
	case t_list_end:
		p.next()
		return l, nil
	default:
		return p.parseList(l)
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

	switch t := p.peek(); t.t {
	case t_object_end:
		p.next()
		return obj, nil
	default:
		return p.parseObject(obj)
	}
}

func (p *parser) number() (interface{}, error) {
	t := p.next()
	if t.t != t_real_number {
		return nil, fmt.Errorf("unexpected %s token while parsing number", t.t)
	}

	if p.peek().t == t_imaginary_number {
		var c complex128
		s := t.s + p.next().s
		if _, err := fmt.Sscan(s, &c); err != nil {
			return nil, fmt.Errorf("ungood imaginary number format %s: %s", s, err)
		}
		return c, nil
	}

	i, err := strconv.ParseInt(t.s, 0, 64)
	if err == nil {
		return int(i), nil
	}

	f, err := strconv.ParseFloat(t.s, 64)
	if err == nil {
		return f, nil
	}

	return nil, fmt.Errorf("this token broke the number parser: %s", t)
}
