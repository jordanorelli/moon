package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type tokenType int

func (t tokenType) String() string {
	switch t {
	case t_error:
		return "t_error"
	case t_string:
		return "t_string"
	case t_name:
		return "t_name"
	case t_type:
		return "t_type"
	case t_equals:
		return "t_equals"
	default:
		panic(fmt.Sprintf("unknown token type: %v", t))
	}
}

const (
	t_error   tokenType = iota // a stored lex error
	t_string                   // a string literal
	t_name                     // a name
	t_type                     // a type
	t_equals                   // equals sign
	t_comment                  // a comment
)

type stateFn func(*lexer) (stateFn, error)

type token struct {
	t tokenType
	s string
}

type lexer struct {
	in     io.RuneReader
	out    chan token
	buf    []rune // running buffer for current lexeme
	backup []rune
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
	if len(l.backup) > 0 {
		r := l.backup[len(l.backup)-1]
		l.backup = l.backup[:len(l.backup)-1]
		return r, nil
	}
	r, _, err := l.in.ReadRune()
	return r, err
}

func (l *lexer) keep(r rune) {
	if l.buf == nil {
		l.buf = make([]rune, 0, 18)
	}
	l.buf = append(l.buf, r)
}

func (l *lexer) unread(r rune) {
	l.backup = append(l.backup, r)
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
		in:     r,
		out:    make(chan token),
		backup: make([]rune, 0, 4),
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
	switch {
	case r == '=':
		l.keep(r)
		l.emit(t_equals)
		return lexRoot, nil
	case r == '"', r == '`':
		return lexStringLiteral(r), nil
	case unicode.IsSpace(r):
		return lexRoot, nil
	case unicode.IsLower(r):
		l.keep(r)
		return lexName, nil
	case unicode.IsUpper(r):
		l.keep(r)
		return lexType, nil
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

func lexName(l *lexer) (stateFn, error) {
	r, err := l.next()
	switch err {
	case io.EOF:
		l.emit(t_name)
		return nil, io.EOF
	case nil:
	default:
		return nil, err
	}
	switch {
	case unicode.IsLetter(r), unicode.IsDigit(r), r == '_':
		l.keep(r)
		return lexName, nil
	default:
		l.emit(t_name)
		l.unread(r)
		return lexRoot, nil
	}
}

func lexType(l *lexer) (stateFn, error) {
	r, err := l.next()
	switch err {
	case io.EOF:
		l.emit(t_type)
		return nil, io.EOF
	case nil:
	default:
		return nil, err
	}
	switch {
	case unicode.IsLetter(r), unicode.IsDigit(r), r == '_':
		l.keep(r)
		return lexType, nil
	default:
		l.emit(t_type)
		l.unread(r)
		return lexRoot, nil
	}
}
