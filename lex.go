package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type tokenType int

const (
	t_error  tokenType = iota // a stored lex error
	t_string                  // a string literal
)

type stateFn func(*lexer) (stateFn, error)

type token struct {
	t tokenType
	s string
}

type lexer struct {
	in  io.RuneReader
	out chan token
	buf []rune // running buffer for current lexeme
}

func (l *lexer) lex() {
	defer close(l.out)
	var err error
	fn := lexRoot
	for {
		fn, err = fn(l)
		switch err {
		case nil:
		case io.EOF:
			return
		default:
			l.out <- token{t_error, err.Error()}
			return
		}
	}
}

func (l *lexer) next() (rune, error) {
	r, _, err := l.in.ReadRune()
	return r, err
}

func (l *lexer) keep(r rune) {
	if l.buf == nil {
		l.buf = make([]rune, 0, 18)
	}
	l.buf = append(l.buf, r)
}

func (l *lexer) emit(t tokenType) {
	l.out <- token{t, string(l.buf)}
	l.buf = l.buf[0:0]
}

func lexString(in string) chan token {
	r := strings.NewReader(in)
	return lex(r)
}

func lex(r io.RuneReader) chan token {
	l := lexer{
		in:  r,
		out: make(chan token),
	}
	go l.lex()
	return l.out
}

func fullTokens(c chan token) ([]token, error) {
	tokens := make([]token, 0, 32)
	for t := range c {
		if t.t == t_error {
			return nil, errors.New(t.s)
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func lexRoot(l *lexer) (stateFn, error) {
	r, err := l.next()
	if err != nil {
		return nil, err
	}
	if unicode.IsSpace(r) {
		return lexRoot, nil
	}
	switch r {
	case '"', '`':
		return lexStringLiteral(r), nil
	default:
		return nil, fmt.Errorf("unexpected rune in lexRoot: %c", r)
	}
}

func lexStringLiteral(delim rune) stateFn {
	return func(l *lexer) (stateFn, error) {
		r, err := l.next()
		if err != nil {
			return nil, err
		}
		switch r {
		case delim:
			l.emit(t_string)
			return lexRoot, nil
		case '\\':
			r, err := l.next()
			if err != nil {
				return nil, err
			}
			l.keep(r)
			return lexStringLiteral(delim), nil
		default:
			l.keep(r)
			return lexStringLiteral(delim), nil
		}
	}
}
