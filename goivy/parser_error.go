package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// ParseError mirrors Python ivy_parser.ParseError (ivy_parser.py:84-98).
// Its Error() output matches Python ParseError.__repr__.
type ParseError struct {
	Filename string
	Lineno   int
	Token    string
	Message  string
}

func (e *ParseError) Error() string {
	xtracer.Trace("parser.__repr__ ENTER")
	var s string
	if e.Filename != "" {
		s += e.Filename
	}
	if e.Lineno != 0 {
		s += fmt.Sprintf("(%d)", e.Lineno)
	}
	if e.Filename != "" || e.Lineno != 0 {
		s += ": "
	}
	s += "error: "
	if e.Token != "" {
		s += fmt.Sprintf("token '%s': ", e.Token)
	}
	s += e.Message
	return s
}
