package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const eof = -1

type tokenType int

func (t tokenType) String() string {
	switch t {
	case t_error:
		return "t_error"
	case t_eof:
		return "t_eof"
	case t_string:
		return "t_string"
	case t_name:
		return "t_name"
	case t_type:
		return "t_type"
	case t_equals:
		return "t_equals"
	case t_comment:
		return "t_comment"
	case t_list_start:
		return "t_list_start"
	case t_list_end:
		return "t_list_end"
	case t_list_separator:
		return "t_list_separator"
	case t_object_start:
		return "t_object_start"
	case t_object_separator:
		return "t_object_separator"
	case t_object_end:
		return "t_object_end"
	case t_int:
		return "t_int"
	default:
		panic(fmt.Sprintf("unknown token type: %v", t))
	}
}

const (
	t_error            tokenType = iota // a stored lex error
	t_eof                               // end of file token
	t_string                            // a string literal
	t_name                              // a name
	t_type                              // a type
	t_equals                            // equals sign
	t_comment                           // a comment
	t_list_start                        // [
	t_list_end                          // ]
	t_list_separator                    // ,
	t_object_start                      // {
	t_object_end                        // }
	t_object_separator                  // :
	t_int                               // an integer
)

type stateFn func(*lexer) stateFn

type token struct {
	t tokenType
	s string
}

type lexer struct {
	in     io.RuneReader
	out    chan token
	buf    []rune // running buffer for current lexeme
	backup []rune
	err    error
}

func (l *lexer) lex() {
	defer close(l.out)
	for fn := lexRoot; fn != nil; {
		fn = fn(l)
		if l.err != nil {
			fn = lexErrorf("read error: %s", l.err)
		}
	}
}

func (l *lexer) next() rune {
	if len(l.backup) > 0 {
		r := l.backup[len(l.backup)-1]
		l.backup = l.backup[:len(l.backup)-1]
		return r
	}
	r, _, err := l.in.ReadRune()
	switch err {
	case io.EOF:
		return eof
	case nil:
		return r
	default:
		l.err = err
		return eof
	}
	return r
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

func lex(r io.Reader) chan token {
	l := lexer{
		in:     bufio.NewReader(r),
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

func lexErrorf(t string, args ...interface{}) stateFn {
	return func(l *lexer) stateFn {
		l.out <- token{t_error, fmt.Sprintf(t, args...)}
		return nil
	}
}

func lexRoot(l *lexer) stateFn {
	r := l.next()
	switch {
	case r == eof:
		return nil
	case r == '=':
		l.keep(r)
		l.emit(t_equals)
		return lexRoot
	case r == '"', r == '`':
		return lexStringLiteral(r)
	case r == '#':
		return lexComment
	case r == '[':
		l.keep(r)
		l.emit(t_list_start)
		return lexRoot
	case r == ']':
		l.keep(r)
		l.emit(t_list_end)
		return lexRoot
	case r == ',':
		l.keep(r)
		l.emit(t_list_separator)
		return lexRoot
	case r == '{':
		l.keep(r)
		l.emit(t_object_start)
		return lexRoot
	case r == '}':
		l.keep(r)
		l.emit(t_object_end)
		return lexRoot
	case r == ':':
		l.keep(r)
		l.emit(t_object_separator)
		return lexRoot
	// case strings.IndexRune("-0123456789", r) >= 0:
	// 	l.unread(r)
	// 	return lexNumber, nil
	case unicode.IsSpace(r):
		return lexRoot
	case unicode.IsLower(r):
		l.keep(r)
		return lexName
	case unicode.IsUpper(r):
		l.keep(r)
		return lexType
	default:
		return lexErrorf("unexpected rune in lexRoot: %c", r)
	}
}

func lexComment(l *lexer) stateFn {
	switch r := l.next(); r {
	case '\n':
		l.emit(t_comment)
		return lexRoot
	case eof:
		l.emit(t_comment)
		return nil
	default:
		l.keep(r)
		return lexComment
	}
}

func lexStringLiteral(delim rune) stateFn {
	return func(l *lexer) stateFn {
		switch r := l.next(); r {
		case delim:
			l.emit(t_string)
			return lexRoot
		case '\\':
			switch r := l.next(); r {
			case eof:
				return lexErrorf("unexpected eof in string literal")
			default:
				l.keep(r)
				return lexStringLiteral(delim)
			}
		case eof:
			return lexErrorf("unexpected eof in string literal")
		default:
			l.keep(r)
			return lexStringLiteral(delim)
		}
	}
}

func lexName(l *lexer) stateFn {
	r := l.next()
	switch {
	case unicode.IsLetter(r), unicode.IsDigit(r), r == '_':
		l.keep(r)
		return lexName
	case r == eof:
		l.emit(t_name)
		return nil
	default:
		l.emit(t_name)
		l.unread(r)
		return lexRoot
	}
}

func lexType(l *lexer) stateFn {
	r := l.next()
	switch {
	case unicode.IsLetter(r), unicode.IsDigit(r), r == '_':
		l.keep(r)
		return lexType
	case r == eof:
		l.emit(t_type)
		return nil
	default:
		l.emit(t_type)
		l.unread(r)
		return lexRoot
	}
}

// func lexNumber(l *lexer) (stateFn, error) {
// 	r, err := l.next()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	switch {
// 	case r == '-', r == '+':
// 		l.keep(r)
// 		return lexNumber, nil
// 	case r == '0':
// 		l.keep(r)
// 		return lexHexOct, nil
// 	case strings.IndexRune("123456789", r) >= 0:
// 		l.keep(r)
// 		return lexDecimal, nil
// 	default:
// 		return nil, fmt.Errorf("unexpected rune in lexNumber: %c", r)
// 	}
// }
//
// func lexHexOct(l *lexer) (stateFn, error) {
// 	r, err := l.next()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	switch {
// 	case r == 'x', r == 'X':
// 		l.keep(r)
// 		return lexHex, nil
//     case r == 'e', r == 'E':
//         l.keep(r)
//         return lexExponent, nil
// 	case r == '8', r == '9':
// 		return nil, fmt.Errorf("unexpected 8 or 9 in lexHexOct.  there's no 8 or 9 in octal!")
// 	case strings.IndexRune("01234567", r) >= 0:
// 		l.keep(r)
// 		return lexOct, nil
// 	default:
//         // we get here for the literals -0, +0 and 0
// 		l.unread(r)
//         l.emit(t_int)
//         return lexRoot, nil
// 	}
// }
//
// func lexHex(l *lexer) (stateFn, error) {
// 	r, err := l.next()
// 	if err != nil {
// 		return nil, err
// 	}
//
//     switch {
//
//     }
// }
//
// func lexDecimal(l *lexer) (stateFn, error) {
// 	r, err := l.next()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	switch {
//     case strings.IndexRune("0123456789", r) >= 0:
//         l.keep(r)
//         return lexDecimal, nil
//     case r == '.':
//         l.keep(r)
//         return lexFloat, nil
// 	}
// }
