package abc

import (
	"fmt"
	"os"
)

// Port of src/sat/bsat2/ParseUtils.h (namespace Minisat).

// minisatBufferSize is `static const int buffer_size = 1048576`.
const minisatBufferSize = 1048576

// MinisatCharStream is what the parsing templates (B in skipWhitespace, skipLine,
// parseInt, eagerMatch) require of their input: `*in` (Deref) and `++in` (Inc).
// It is implemented by StreamBuffer and by MinisatCharPtr. Deref returns EOF (-1)
// at the end of a stream and 0 at the end of a character string.
type MinisatCharStream interface {
	Deref() int
	Inc()
}

// MinisatCharPtr is a `const char*` cursor into a NUL-terminated string, as used
// with the string instantiations of the templates (match, isEof(const char*)).
type MinisatCharPtr struct {
	s   string
	pos int
}

// NewMinisatCharPtr starts a cursor at the beginning of s.
func NewMinisatCharPtr(s string) *MinisatCharPtr { return &MinisatCharPtr{s: s} }

// Deref is `*in`.
// C: *in
func (p *MinisatCharPtr) Deref() int { return p.At(0) }

// Inc is `++in`.
// C: ++in
func (p *MinisatCharPtr) Inc() { p.pos++ }

// At is `in[i]` (0 past the end: the terminating NUL).
// C: in[i]
func (p *MinisatCharPtr) At(i int) int {
	if p.pos+i >= len(p.s) {
		return 0
	}
	return int(p.s[p.pos+i])
}

// Advance is `in += i`.
// C: in += i
func (p *MinisatCharPtr) Advance(i int) { p.pos += i }

// Rest is the string the cursor points at.
func (p *MinisatCharPtr) Rest() string {
	if p.pos >= len(p.s) {
		return ""
	}
	return p.s[p.pos:]
}

// minisatEOF is the C macro EOF.
const minisatEOF = -1

// MinisatStreamBuffer is Minisat::StreamBuffer: a buffered reader over a gzFile.
type MinisatStreamBuffer struct {
	in   *GZIPReader // gzFile
	buf  [minisatBufferSize]uint8
	pos  int
	size int
}

// assureLookahead is StreamBuffer::assureLookahead (private).
// C: void assureLookahead()
func (b *MinisatStreamBuffer) assureLookahead() {
	if b.pos >= b.size {
		b.pos = 0
		n, _ := b.in.Read(b.buf[:]) // gzread: a negative result (error) is treated as no data
		if n < 0 {
			n = 0
		}
		b.size = n
	}
}

// NewMinisatStreamBuffer is the explicit constructor StreamBuffer(gzFile i).
// C: explicit StreamBuffer(gzFile i)
func NewMinisatStreamBuffer(i *GZIPReader) *MinisatStreamBuffer {
	b := &MinisatStreamBuffer{in: i}
	b.assureLookahead()
	return b
}

// Deref is `int operator*() const`.
// C: int operator*() const
func (b *MinisatStreamBuffer) Deref() int {
	if b.pos >= b.size {
		return minisatEOF
	}
	return int(b.buf[b.pos])
}

// Inc is `void operator++()`.
// C: void operator++()
func (b *MinisatStreamBuffer) Inc() {
	b.pos++
	b.assureLookahead()
}

// Position is StreamBuffer::position.
// C: int position() const
func (b *MinisatStreamBuffer) Position() int { return b.pos }

// minisatIsEofStream is the overload isEof(StreamBuffer&) (static inline).
// C: static inline bool isEof(StreamBuffer& in)
func minisatIsEofStream(in *MinisatStreamBuffer) bool { return in.Deref() == minisatEOF }

// minisatIsEofChars is the overload isEof(const char*) (static inline).
// C: static inline bool isEof(const char* in)
func minisatIsEofChars(in *MinisatCharPtr) bool { return in.Deref() == 0 }

// minisatIsEof dispatches to the isEof overload of the stream type.
func minisatIsEof(in MinisatCharStream) bool {
	switch v := in.(type) {
	case *MinisatStreamBuffer:
		return minisatIsEofStream(v)
	case *MinisatCharPtr:
		return minisatIsEofChars(v)
	}
	panic("minisatIsEof: unsupported stream type")
}

// minisatSkipWhitespace is Minisat::skipWhitespace (static template).
// C: template<class B> static void skipWhitespace(B& in)
func minisatSkipWhitespace[B MinisatCharStream](in B) {
	for (in.Deref() >= 9 && in.Deref() <= 13) || in.Deref() == 32 {
		in.Inc()
	}
}

// minisatSkipLine is Minisat::skipLine (static template; uses isEof).
// C: template<class B> static void skipLine(B& in)
func minisatSkipLine[B MinisatCharStream](in B) {
	for {
		if minisatIsEof(in) {
			return
		}
		if in.Deref() == '\n' {
			in.Inc()
			return
		}
		in.Inc()
	}
}

// minisatParseInt is Minisat::parseInt (static template); an unexpected character
// prints a message and exits with status 3, as in C.
// C: template<class B> static int parseInt(B& in)
func minisatParseInt[B MinisatCharStream](in B) int {
	val := 0
	neg := false
	minisatSkipWhitespace(in)
	if in.Deref() == '-' {
		neg = true
		in.Inc()
	} else if in.Deref() == '+' {
		in.Inc()
	}
	if in.Deref() < '0' || in.Deref() > '9' {
		fmt.Fprintf(os.Stderr, "PARSE ERROR! Unexpected char: %c\n", rune(in.Deref()))
		os.Exit(3)
	}
	for in.Deref() >= '0' && in.Deref() <= '9' {
		val = val*10 + (in.Deref() - '0')
		in.Inc()
	}
	if neg {
		return -val
	}
	return val
}

// minisatMatch is Minisat::match (static template; only instantiated with const char*).
// C: template<class B> static bool match(B& in, const char* str)
func minisatMatch(in *MinisatCharPtr, str string) bool {
	i := 0
	for ; i < len(str); i++ {
		if in.At(i) != int(str[i]) {
			return false
		}
	}
	in.Advance(i)
	return true
}

// minisatEagerMatch is Minisat::eagerMatch (static template).
// C: template<class B> static bool eagerMatch(B& in, const char* str)
func minisatEagerMatch[B MinisatCharStream](in B, str string) bool {
	for i := 0; i < len(str); i++ {
		if int(str[i]) != in.Deref() {
			return false
		}
		in.Inc()
	}
	return true
}
