package ivy2golang

import (
	"fmt"
	"strings"
)

type goWriter struct {
	buf    strings.Builder
	indent int
}

func (w *goWriter) line(s string) {
	if s != "" {
		w.buf.WriteString(strings.Repeat("\t", w.indent))
		w.buf.WriteString(s)
	}
	w.buf.WriteByte('\n')
}

func (w *goWriter) linef(format string, args ...any) {
	w.line(fmt.Sprintf(format, args...))
}

func (w *goWriter) raw(s string) {
	w.buf.WriteString(s)
}

func (w *goWriter) blank() {
	w.buf.WriteByte('\n')
}

func (w *goWriter) open(s string) {
	w.line(s)
	w.indent++
}

func (w *goWriter) close(suffix string) {
	if w.indent > 0 {
		w.indent--
	}
	w.line("}" + suffix)
}

func (w *goWriter) String() string {
	return w.buf.String()
}
