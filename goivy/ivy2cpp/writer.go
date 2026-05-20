package ivy2cpp

import (
	"fmt"
	"strings"
)

type cppWriter struct {
	buf    strings.Builder
	indent int
}

func (w *cppWriter) line(s string) {
	if s != "" {
		w.buf.WriteString(strings.Repeat("    ", w.indent))
		w.buf.WriteString(s)
	}
	w.buf.WriteByte('\n')
}

func (w *cppWriter) linef(format string, args ...any) {
	w.line(fmt.Sprintf(format, args...))
}

func (w *cppWriter) raw(s string) {
	w.buf.WriteString(s)
}

func (w *cppWriter) blank() {
	w.buf.WriteByte('\n')
}

func (w *cppWriter) open(s string) {
	w.line(s)
	w.indent++
}

func (w *cppWriter) close(suffix string) {
	if w.indent > 0 {
		w.indent--
	}
	w.line("}" + suffix)
}

func (w *cppWriter) String() string {
	return w.buf.String()
}
