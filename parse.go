package main

import (
	"fmt"
	"io"
)

const (
	e_no_error parseErrorType = iota
	e_lex_error
	e_unexpected_eof
	e_unexpected_token
)

type parseErrorType int

type parseError struct {
	t parseErrorType
	m string
}

func (p parseError) Error() string {
	return fmt.Sprintf("parse error: %s", p.m)
}

func parseErrorf(t parseErrorType, tpl string, args ...interface{}) error {
	return parseError{t: t, m: fmt.Sprintf(tpl, args...)}
}

type parseFn func(*parser) (parseFn, error)

func parseRoot(p *parser) (parseFn, error) {
	if err := p.next(); err != nil {
		return nil, err
	}
	switch p.cur.t {
	case t_name:
		return parseAfterName(p.cur.s), nil
	default:
		return nil, parseErrorf(e_unexpected_token, "unexpected %s token in parseRoot", p.cur.t)
	}
}

func parseAfterName(name string) parseFn {
	return func(p *parser) (parseFn, error) {
		switch err := p.next(); err {
		case io.EOF:
			return nil, parseErrorf(e_unexpected_eof, "unexpected eof after name %s", name)
		case nil:
		default:
			return nil, err
		}

		switch p.cur.t {
		case t_equals:
			return parseAssign(name), nil
		default:
			return nil, parseErrorf(e_unexpected_token, "unexpected %s token in parseAfterName", p.cur.t)
		}
	}
}

func parseAssign(name string) parseFn {
	return func(p *parser) (parseFn, error) {
		switch err := p.next(); err {
		case io.EOF:
			return nil, parseErrorf(e_unexpected_eof, "unexpected eof when trying to parse value for name %s", name)
		case nil:
		default:
			return nil, err
		}

		switch p.cur.t {
		case t_string:
			p.out.setUnique(name, p.cur.s)
			return parseRoot, nil
		default:
			return nil, parseErrorf(e_unexpected_token, "unexpected %s token in parseAssign", p.cur.t)
		}
	}
}

type parser struct {
	in  chan token
	cur token
	out *Config
}

func (p *parser) next() error {
	t, ok := <-p.in
	if !ok {
		return io.EOF
	}
	if t.t == t_error {
		return parseError{e_lex_error, t.s}
	}
	p.cur = t
	return nil
}

func (p *parser) run() error {
	fn := parseRoot
	var err error
	for {
		fn, err = fn(p)
		switch err {
		case io.EOF:
			return nil
		case nil:
		default:
			return err
		}
	}
}

type assignment struct {
	name  string
	value interface{}
}

func parse(in chan token) (*Config, error) {
	p := &parser{
		in:  in,
		out: new(Config),
	}
	if err := p.run(); err != nil {
		return nil, err
	}
	return p.out, nil
}

func parseString(in string) (*Config, error) {
	return parse(lexString(in))
}
