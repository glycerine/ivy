package ivy2cpp

import (
	"fmt"
	"strings"
)

type cppWriter struct {
	buf    strings.Builder
	text   *CppText
	indent int
}

func newCPPWriter(text *CppText) cppWriter {
	return cppWriter{text: text}
}

func (w *cppWriter) line(s string) {
	var b strings.Builder
	if s != "" {
		b.WriteString(strings.Repeat("    ", w.indent))
		b.WriteString(s)
	}
	b.WriteByte('\n')
	w.writeString(b.String())
}

func (w *cppWriter) linef(format string, args ...any) {
	w.line(fmt.Sprintf(format, args...))
}

func (w *cppWriter) raw(s string) {
	w.writeString(s)
}

func (w *cppWriter) blank() {
	w.writeString("\n")
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
	if w.text != nil {
		return w.text.GetFile()
	}
	return w.buf.String()
}

func (w *cppWriter) writeString(s string) {
	if w.text != nil {
		w.text.Write(s)
		return
	}
	w.buf.WriteString(s)
}
