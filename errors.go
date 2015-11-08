package patcher

import "fmt"

type Error struct {
	Expression string
	Line       int
	Column     int
	Token      *Token
	ErrorMsg   string
}

// Returns a nice formatted error string.
func (e *Error) Error() string {
	s := "[Error"
	if e.Line > 0 {
		s += fmt.Sprintf(" | Line %d Col %d", e.Line, e.Column)
		if e.Token != nil {
			s += fmt.Sprintf(" near '%s'", e.Token.Val)
		}
	}
	s += "] "
	s += e.ErrorMsg
	return s
}

func NewError(msg string, token *Token) *Error {
	var line, col int
	if token != nil {
		// No tokens available
		// TODO: Add location (from where?)
		line = token.Line
		col = token.Col
	}
	return &Error{
		Line:     line,
		Column:   col,
		Token:    token,
		ErrorMsg: msg,
	}
}
