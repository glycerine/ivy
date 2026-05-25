package ivy2go

import (
	"fmt"
	"strings"
)

// goWriter is the Go-emission analogue of ivy2cpp/writer.go's cppWriter.
// It buffers code lines at a current indentation level. The output is
// indented with tabs (the Go convention) rather than spaces; format.Source
// is the final arbiter anyway, so the indentation just keeps emitted
// source readable during debugging.
type goWriter struct {
	buf    strings.Builder
	text   *GoText
	indent int
}

func newGoWriter(text *GoText) goWriter {
	return goWriter{text: text}
}

// line writes s at the current indentation, followed by a newline. An
// empty s emits only the newline (no leading whitespace) to keep blank
// lines truly blank.
func (w *goWriter) line(s string) {
	var b strings.Builder
	if s != "" {
		b.WriteString(strings.Repeat("\t", w.indent))
		b.WriteString(s)
	}
	b.WriteByte('\n')
	w.writeString(b.String())
}

func (w *goWriter) linef(format string, args ...any) {
	w.line(fmt.Sprintf(format, args...))
}

// raw writes s to the underlying buffer without indentation handling.
func (w *goWriter) raw(s string) {
	w.writeString(s)
}

func (w *goWriter) blank() {
	w.writeString("\n")
}

// open writes s as a line then increments the indentation. Convention:
// callers pass the line that ends in `{` (e.g., `if cond {`, `func f() {`).
func (w *goWriter) open(s string) {
	w.line(s)
	w.indent++
}

// close decrements the indentation and writes `}` followed by suffix.
// In Go, the suffix is almost always empty — pass "" — but the API
// mirrors ivy2cpp's writer for cross-package patch parity.
func (w *goWriter) close(suffix string) {
	if w.indent > 0 {
		w.indent--
	}
	w.line("}" + suffix)
}

// closeParen decrements the indentation and writes `)`. Used for Go's
// parenthesized blocks (import, const, var, type) whose open form
// ends in `(`.
func (w *goWriter) closeParen() {
	if w.indent > 0 {
		w.indent--
	}
	w.line(")")
}

func (w *goWriter) String() string {
	if w.text != nil {
		return w.text.GetFile()
	}
	return w.buf.String()
}

func (w *goWriter) writeString(s string) {
	if w.text != nil {
		w.text.Write(s)
		return
	}
	w.buf.WriteString(s)
}
