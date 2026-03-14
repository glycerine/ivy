// Package gogen generates runnable Go source code from Ivy's internal
// representation (Module, actions, logic formulas, sorts).
package gogen

import (
	"fmt"
	"strings"
)

// CodeWriter provides indented code writing utilities for Go code generation.
type CodeWriter struct {
	buf    strings.Builder
	indent int
}

// NewCodeWriter creates a new CodeWriter with zero indentation.
func NewCodeWriter() *CodeWriter {
	return &CodeWriter{}
}

// Line writes a single line at the current indentation level.
func (w *CodeWriter) Line(s string) {
	w.buf.WriteString(strings.Repeat("\t", w.indent))
	w.buf.WriteString(s)
	w.buf.WriteByte('\n')
}

// Linef writes a formatted line at the current indentation level.
func (w *CodeWriter) Linef(format string, args ...interface{}) {
	w.Line(fmt.Sprintf(format, args...))
}

// OpenBlock writes a header line and increases indentation.
// For example: OpenBlock("func foo() {") writes the header and indents.
func (w *CodeWriter) OpenBlock(header string) {
	w.Line(header)
	w.indent++
}

// CloseBlock decreases indentation and writes a closing brace.
func (w *CodeWriter) CloseBlock() {
	if w.indent > 0 {
		w.indent--
	}
	w.Line("}")
}

// BlankLine writes an empty line (no indentation).
func (w *CodeWriter) BlankLine() {
	w.buf.WriteByte('\n')
}

// String returns all accumulated output.
func (w *CodeWriter) String() string {
	return w.buf.String()
}

// Indent returns the current indentation level.
func (w *CodeWriter) Indent() int {
	return w.indent
}

// Raw writes a string directly with no indentation or newline.
func (w *CodeWriter) Raw(s string) {
	w.buf.WriteString(s)
}
