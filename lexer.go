package el

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	TokenError = iota
	EOF

	TokenKeyword
	TokenIdentifier
	TokenString
	TokenNumber
	TokenSymbol
)

var (
	tokenSpaceChars                = " \n\r\t"
	tokenIdentifierChars           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	tokenIdentifierCharsWithDigits = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789"
	tokenDigits                    = "0123456789"

	TokenSymbols = []string{";", "(", ")", ".", "[", "]"}

	TokenKeywords = []string{"true", "false"}
)

type TokenType int
type Token struct {
	Typ  TokenType
	Val  string
	Line int
	Col  int
}

type lexerStateFn func() lexerStateFn
type lexer struct {
	input     string
	start     int
	pos       int
	width     int
	tokens    []*Token
	errored   bool
	startline int
	startcol  int
	line      int
	col       int
}

// Lex do lexical analysis
func Lex(input string) ([]*Token, *Error) {
	l := &lexer{
		input:     input,
		tokens:    make([]*Token, 0, 100),
		line:      1,
		col:       1,
		startline: 1,
		startcol:  1,
	}
	l.run()
	if l.errored {
		errtoken := l.tokens[len(l.tokens)-1]
		return nil, &Error{
			Line:     errtoken.Line,
			Column:   errtoken.Col,
			ErrorMsg: errtoken.Val,
		}
	}
	return l.tokens, nil
}

func (l *lexer) value() string {
	return l.input[l.start:l.pos]
}

func (l *lexer) length() int {
	return l.pos - l.start
}

func (l *lexer) emit(t TokenType) {
	l.emitWithChange(t, nil)
}

func (l *lexer) emitWithChange(t TokenType, changeToken func(string) string) {
	value := l.value()
	if changeToken != nil {
		value = changeToken(value)
	}
	tok := &Token{
		Typ:  t,
		Val:  value,
		Line: l.startline,
		Col:  l.startcol,
	}

	if t == TokenString {
		tok.Val = strings.Replace(tok.Val, `\"`, `"`, -1)
		tok.Val = strings.Replace(tok.Val, `\\`, `\`, -1)
	}

	l.tokens = append(l.tokens, tok)
	l.start = l.pos
	l.startline = l.line
	l.startcol = l.col
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	l.col += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
	l.col -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
	l.startline = l.line
	l.startcol = l.col
}

func (l *lexer) accept(what string) bool {
	if strings.IndexRune(what, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(what string) {
	for strings.IndexRune(what, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) lexerStateFn {
	t := &Token{
		Typ:  TokenError,
		Val:  fmt.Sprintf(format, args...),
		Line: l.startline,
		Col:  l.startcol,
	}
	l.tokens = append(l.tokens, t)
	l.errored = true
	l.startline = l.line
	l.startcol = l.col
	return nil
}

func (l *lexer) eof() bool {
	return l.start >= len(l.input)-1
}

func (l *lexer) run() {
	for {
		for state := l.stateCode; state != nil; {
			state = state()
		}
		if l.errored {
			return
		}
		if l.next() == EOF {
			break
		}
	}
}

func (l *lexer) stateCode() lexerStateFn {
outer_loop:
	for {
		switch {
		case l.accept(tokenSpaceChars):
			if l.value() == "\n" {
				return l.errorf("Newline not allowed within expression.")
			}
			l.ignore()
			continue
		case l.accept(tokenIdentifierChars):
			return l.stateIdentifier
		case l.accept(tokenDigits):
			return l.stateNumber
		case l.accept(`"`):
			return l.stateString
		}
		for _, sym := range TokenSymbols {
			if strings.HasPrefix(l.input[l.start:], sym) {
				l.pos += len(sym)
				l.col += l.length()
				l.emit(TokenSymbol)

				if sym == ";" {
					// Tag/variable end, return after emit
					return nil
				}

				continue outer_loop
			}
		}

		if l.pos < len(l.input) {
			return l.errorf("Unknown character: %q (%d)", l.peek(), l.peek())
		}

		break
	}
	return nil
}

func (l *lexer) stateNumber() lexerStateFn {
	l.acceptRun(tokenDigits)
	l.emit(TokenNumber)
	return l.stateCode
}

func (l *lexer) stateIdentifier() lexerStateFn {
	l.acceptRun(tokenIdentifierChars)
	l.acceptRun(tokenIdentifierCharsWithDigits)
	for _, kw := range TokenKeywords {
		if kw == l.value() {
			l.emit(TokenKeyword)
			return l.stateCode
		}
	}
	l.emitWithChange(TokenIdentifier, upperFirst)
	return l.stateCode
}

func (l *lexer) stateString() lexerStateFn {
	l.ignore()
	l.startcol-- // we're starting the position at the first "
	for !l.accept(`"`) {
		switch l.next() {
		case '\\':
			// escape sequence
			switch l.peek() {
			case '"', '\\':
				l.next()
			default:
				return l.errorf("Unknown escape sequence: \\%c", l.peek())
			}
		case EOF:
			return l.errorf("Unexpected EOF, string not closed.")
		case '\n':
			return l.errorf("Newline in string is not allowed.")
		}
	}
	l.backup()
	l.emit(TokenString)

	l.next()
	l.ignore()

	return l.stateCode
}
