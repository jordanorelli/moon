package moon

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
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
	case t_comment:
		return "t_comment"
	case t_list_start:
		return "t_list_start"
	case t_list_end:
		return "t_list_end"
	case t_object_start:
		return "t_object_start"
	case t_object_separator:
		return "t_object_separator"
	case t_object_end:
		return "t_object_end"
	case t_real_number:
		return "t_real_number"
	case t_imaginary_number:
		return "t_imaginary_number"
	case t_variable:
		return "t_variable"
	case t_bool:
		return "t_bool"
	case t_duration:
		return "t_duration"
	default:
		panic(fmt.Sprintf("unknown token type: %v", t))
	}
}

const (
	t_error            tokenType = iota // a stored lex error
	t_eof                               // end of file token
	t_string                            // a bare string
	t_string_quoted                     // a quoted string
	t_name                              // a name
	t_comment                           // a comment
	t_list_start                        // [
	t_list_end                          // ]
	t_object_start                      // {
	t_object_end                        // }
	t_object_separator                  // :
	t_real_number                       // a number
	t_imaginary_number                  // an imaginary number
	t_variable                          // e.g. @var_name, a variable name.
	t_bool                              // a boolean token (true|false)
	t_duration                          // a duration (e.g.: 1s, 2h45m, 900ms)
)

type stateFn func(*lexer) stateFn

type token struct {
	t tokenType
	s string
}

func (t token) String() string {
	return fmt.Sprintf("{%s %s}", t.t, t.s)
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

func (l *lexer) peek() rune {
	r := l.next()
	l.unread(r)
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
	switch t {
	case t_variable:
		if !l.bufHasSpaces() {
			break
		}
		msg := fmt.Sprintf(`invalid var name: "%s" (var names cannot contain spaces)`, string(l.buf))
		l.out <- token{t_error, msg}
		return
	case t_name:
		if !l.bufHasSpaces() {
			break
		}
		msg := fmt.Sprintf(`invalid name: "%s" (names cannot contain spaces)`, string(l.buf))
		l.out <- token{t_error, msg}
		return
	case t_string:
		switch string(l.buf) {
		case "true", "false":
			t = t_bool
		}
	case t_string_quoted:
		t = t_string
	}
	l.out <- token{t, string(l.buf)}
	l.buf = l.buf[0:0]
}

func (l *lexer) accept(chars string) bool {
	r := l.next()
	if strings.IndexRune(chars, r) >= 0 {
		l.keep(r)
		return true
	} else {
		l.unread(r)
		return false
	}
}

func (l *lexer) acceptRun(chars string) bool {
	none := true
	for l.accept(chars) {
		none = false
	}
	return !none
}

func (l *lexer) bufHasSpaces() bool {
	for _, r := range l.buf {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
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
	case r == ':':
		l.keep(r)
		l.emit(t_object_separator)
		return lexRoot
	case r == '"', r == '`':
		return lexQuotedString(r)
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
	case r == '.':
		l.keep(r)
		return lexAfterPeriod
	case r == '<':
		if l.peek() == '<' {
			l.next()
			return lexHeredocStart
		}
		fallthrough
	case r == '@':
		return lexVariable
	case strings.IndexRune("+-0123456789", r) >= 0:
		l.unread(r)
		return lexNumber
	case unicode.IsSpace(r):
		return lexRoot
	case unicode.IsPrint(r):
		l.keep(r)
		return lexNameOrString
	default:
		return lexErrorf("unexpected rune in lexRoot: %c", r)
	}
}

func lexAfterPeriod(l *lexer) stateFn {
	r := l.next()
	switch {
	case strings.IndexRune("+-0123456789", r) >= 0:
		l.unread(r)
		return lexNumber
	case unicode.IsLower(r):
		l.keep(r)
		return lexNameOrString
	default:
		return lexErrorf("unexpected rune after period: %c", r)
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

func lexQuotedString(delim rune) stateFn {
	return func(l *lexer) stateFn {
		switch r := l.next(); r {
		case delim:
			l.emit(t_string_quoted)
			return lexRoot
		case '\\':
			switch r := l.next(); r {
			case eof:
				return lexErrorf("unexpected eof in string literal")
			default:
				l.keep(r)
				return lexQuotedString(delim)
			}
		case eof:
			return lexErrorf("unexpected eof in string literal")
		default:
			l.keep(r)
			return lexQuotedString(delim)
		}
	}
}

func lexNameOrString(l *lexer) stateFn {
	r := l.next()
	switch {
	case r == '\n', r == ';':
		l.emit(t_string)
		return lexRoot
	case r == ':':
		l.emit(t_name)
		l.keep(r)
		l.emit(t_object_separator)
		return lexRoot
	case isSpecial(r):
		l.emit(t_string)
		l.unread(r)
		return lexRoot
	case r == '\\':
		rr := l.next()
		if rr == eof {
			return lexErrorf("unexpected eof in string or name")
		}
		l.keep(rr)
		return lexNameOrString
	case r == eof:
		l.emit(t_string)
		return nil
	case unicode.IsGraphic(r):
		l.keep(r)
		return lexNameOrString
	default:
		return lexErrorf("unexpected rune in string or name: %c", r)
	}
}

func lexVariable(l *lexer) stateFn {
	r := l.next()
	switch {
	case unicode.IsSpace(r), r == ';':
		l.emit(t_variable)
		return lexRoot
	case r == '\\':
		rr := l.next()
		if rr == eof {
			return lexErrorf("unexpected eof in variable name")
		}
		l.keep(rr)
		return lexVariable
	case isSpecial(r):
		l.emit(t_variable)
		l.unread(r)
		return lexRoot
	case r == eof:
		l.emit(t_variable)
		return nil
	case unicode.IsGraphic(r):
		l.keep(r)
		return lexVariable
	default:
		return lexErrorf("unexpected rune in var name: %c", r)
	}
}

func lexNumber(l *lexer) stateFn {
	l.accept("+-")
	digits := "0123456789"
	if l.accept("0") {
		if l.accept("xX") {
			digits = "0123456789abcdefABCDEF"
		} else {
			digits = "01234567"
		}
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	imaginary := l.accept("i")
	r := l.next()
	if isAlphaNumeric(r) {
		l.keep(r)
		return lexDuration
	}
	l.unread(r)
	if imaginary {
		l.emit(t_imaginary_number)
	} else {
		l.emit(t_real_number)
	}
	return lexRoot
}

func lexDuration(l *lexer) stateFn {
	r := l.next()
	switch {
	case r == '\n', r == ';':
		_, err := time.ParseDuration(string(l.buf))
		if err == nil {
			l.emit(t_duration)
			return lexRoot
		}
		l.emit(t_string)
		return lexRoot
	case unicode.IsSpace(r):
		_, err := time.ParseDuration(string(l.buf))
		if err == nil {
			l.emit(t_duration)
			return lexRoot
		}
		l.keep(r)
		return lexNameOrString
	case r == ':':
		_, err := time.ParseDuration(string(l.buf))
		if err == nil {
			l.emit(t_duration)
		} else {
			l.emit(t_string)
		}
		l.keep(r)
		l.emit(t_object_separator)
		return lexRoot
	case isSpecial(r):
		_, err := time.ParseDuration(string(l.buf))
		if err == nil {
			l.emit(t_duration)
		} else {
			l.emit(t_string)
		}
		l.unread(r)
		return lexRoot
	case r == '\\':
		l.unread(r)
		return lexNameOrString
	case r == eof:
		_, err := time.ParseDuration(string(l.buf))
		if err == nil {
			l.emit(t_duration)
		} else {
			l.emit(t_string)
		}
		return nil
	case unicode.IsGraphic(r):
		l.keep(r)
		return lexDuration
	default:
		return lexErrorf("unhandled character type in lexDuration: %c", r)
	}
}

func lexHeredocStart(l *lexer) stateFn {
	r := l.next()
	switch {
	case r == '\n':
		if len(l.buf) == 0 {
			return lexErrorf("illegal zero-width heredoc name")
		}
		label := string(l.buf)
		l.buf = l.buf[0:0]
		return lexHeredocBody(label)
	case unicode.IsUpper(r):
		l.keep(r)
		return lexHeredocStart
	case r == eof:
		return lexErrorf("unexpected EOF in lexHeredocStart")
	default:
		return lexErrorf("unexpected rune in lexHeredocStart: %c (only uppercase letters are ok here)", r)
	}
}

func lexHeredocBody(label string) stateFn {
	var body bytes.Buffer
	line := make([]rune, 0, 128)
	return func(l *lexer) stateFn {
		for {
			r := l.next()
			switch r {
			case '\n':
				if string(line) == label {
					l.out <- token{t_string, string(body.Bytes())}
					return lexRoot
				}
				body.WriteString(string(line))
				line = line[0:0]
				body.WriteRune(r)
			case eof:
				return lexErrorf("unexpected eof inside of heredoc %s", label)
			default:
				line = append(line, r)
			}
		}
	}
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpecial(r rune) bool {
	return strings.ContainsRune("[]{}:;#", r)
}
