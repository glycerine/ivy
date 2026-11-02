package abc

import (
	"fmt"
	"math"
	"os"
)

// Port of src/sat/glucose2/ParseUtils.h (namespace Gluco2).

// gluco2BufferSize is `static const int buffer_size = 1048576`.
const gluco2BufferSize = 1048576

// Gluco2CharStream is what the parsing templates (B in skipWhitespace, skipLine,
// parseInt, eagerMatch) require of their input: `*in` (Deref) and `++in` (Inc).
// It is implemented by StreamBuffer and by Gluco2CharPtr. Deref returns EOF (-1)
// at the end of a stream and 0 at the end of a character string.
type Gluco2CharStream interface {
	Deref() int
	Inc()
}

// Gluco2CharPtr is a `const char*` cursor into a NUL-terminated string, as used
// with the string instantiations of the templates (match, isEof(const char*)).
type Gluco2CharPtr struct {
	s   string
	pos int
}

// NewGluco2CharPtr starts a cursor at the beginning of s.
func NewGluco2CharPtr(s string) *Gluco2CharPtr { return &Gluco2CharPtr{s: s} }

// Deref is `*in`.
// C: *in
func (p *Gluco2CharPtr) Deref() int { return p.At(0) }

// Inc is `++in`.
// C: ++in
func (p *Gluco2CharPtr) Inc() { p.pos++ }

// At is `in[i]` (0 past the end: the terminating NUL).
// C: in[i]
func (p *Gluco2CharPtr) At(i int) int {
	if p.pos+i >= len(p.s) {
		return 0
	}
	return int(p.s[p.pos+i])
}

// Advance is `in += i`.
// C: in += i
func (p *Gluco2CharPtr) Advance(i int) { p.pos += i }

// Rest is the string the cursor points at.
func (p *Gluco2CharPtr) Rest() string {
	if p.pos >= len(p.s) {
		return ""
	}
	return p.s[p.pos:]
}

// gluco2EOF is the C macro EOF.
const gluco2EOF = -1

// Gluco2StreamBuffer is Gluco2::StreamBuffer: a buffered reader over a gzFile.
type Gluco2StreamBuffer struct {
	in   *GZIPReader // gzFile
	buf  [gluco2BufferSize]uint8
	pos  int
	size int
}

// assureLookahead is StreamBuffer::assureLookahead (private).
// C: void assureLookahead()
func (b *Gluco2StreamBuffer) assureLookahead() {
	if b.pos >= b.size {
		b.pos = 0
		n, _ := b.in.Read(b.buf[:]) // gzread: a negative result (error) is treated as no data
		if n < 0 {
			n = 0
		}
		b.size = n
	}
}

// NewGluco2StreamBuffer is the explicit constructor StreamBuffer(gzFile i).
// C: explicit StreamBuffer(gzFile i)
func NewGluco2StreamBuffer(i *GZIPReader) *Gluco2StreamBuffer {
	b := &Gluco2StreamBuffer{in: i}
	b.assureLookahead()
	return b
}

// Deref is `int operator*() const`.
// C: int operator*() const
func (b *Gluco2StreamBuffer) Deref() int {
	if b.pos >= b.size {
		return gluco2EOF
	}
	return int(b.buf[b.pos])
}

// Inc is `void operator++()`.
// C: void operator++()
func (b *Gluco2StreamBuffer) Inc() {
	b.pos++
	b.assureLookahead()
}

// Position is StreamBuffer::position.
// C: int position() const
func (b *Gluco2StreamBuffer) Position() int { return b.pos }

// gluco2IsEofStream is the overload isEof(StreamBuffer&) (static inline).
// C: static inline bool isEof(StreamBuffer& in)
func gluco2IsEofStream(in *Gluco2StreamBuffer) bool { return in.Deref() == gluco2EOF }

// gluco2IsEofChars is the overload isEof(const char*) (static inline).
// C: static inline bool isEof(const char* in)
func gluco2IsEofChars(in *Gluco2CharPtr) bool { return in.Deref() == 0 }

// gluco2IsEof dispatches to the isEof overload of the stream type.
func gluco2IsEof(in Gluco2CharStream) bool {
	switch v := in.(type) {
	case *Gluco2StreamBuffer:
		return gluco2IsEofStream(v)
	case *Gluco2CharPtr:
		return gluco2IsEofChars(v)
	}
	panic("gluco2IsEof: unsupported stream type")
}

// gluco2SkipWhitespace is Gluco2::skipWhitespace (static template).
// C: template<class B> static void skipWhitespace(B& in)
func gluco2SkipWhitespace[B Gluco2CharStream](in B) {
	for (in.Deref() >= 9 && in.Deref() <= 13) || in.Deref() == 32 {
		in.Inc()
	}
}

// gluco2SkipLine is Gluco2::skipLine (static template; uses isEof).
// C: template<class B> static void skipLine(B& in)
func gluco2SkipLine[B Gluco2CharStream](in B) {
	for {
		if gluco2IsEof(in) {
			return
		}
		if in.Deref() == '\n' {
			in.Inc()
			return
		}
		in.Inc()
	}
}

// gluco2ParseInt is Gluco2::parseInt (static template); an unexpected character
// prints a message and exits with status 3, as in C.
// C: template<class B> static int parseInt(B& in)
func gluco2ParseInt[B Gluco2CharStream](in B) int {
	val := 0
	neg := false
	gluco2SkipWhitespace(in)
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

// gluco2Match is Gluco2::match (static template; only instantiated with const char*).
// C: template<class B> static bool match(B& in, const char* str)
func gluco2Match(in *Gluco2CharPtr, str string) bool {
	i := 0
	for ; i < len(str); i++ {
		if in.At(i) != int(str[i]) {
			return false
		}
	}
	in.Advance(i)
	return true
}

// gluco2EagerMatch is Gluco2::eagerMatch (static template).
// C: template<class B> static bool eagerMatch(B& in, const char* str)
func gluco2EagerMatch[B Gluco2CharStream](in B, str string) bool {
	for i := 0; i < len(str); i++ {
		if int(str[i]) != in.Deref() {
			return false
		}
		in.Inc()
	}
	return true
}

// gluco2ParseDouble is parseDouble (static template; only the form X.XXXXXe-XX is accepted).
// C: template<class B> static double parseDouble(B& in)
func gluco2ParseDouble[B Gluco2CharStream](in B) float64 {
	neg := false
	accu := 0.0
	var currentExponent float64
	var exponent int
	gluco2SkipWhitespace(in)
	if in.Deref() == gluco2EOF {
		return 0
	}
	if in.Deref() == '-' {
		neg = true
		in.Inc()
	} else if in.Deref() == '+' {
		in.Inc()
	}
	if in.Deref() < '1' || in.Deref() > '9' {
		fmt.Printf("PARSE ERROR! Unexpected char: %c\n", rune(in.Deref()))
		os.Exit(3)
	}
	accu = float64(in.Deref() - '0')
	in.Inc()
	if in.Deref() != '.' {
		fmt.Printf("PARSE ERROR! Unexpected char: %c\n", rune(in.Deref()))
		os.Exit(3)
	}
	in.Inc() // skip dot
	currentExponent = 0.1
	for in.Deref() >= '0' && in.Deref() <= '9' {
		accu = accu + currentExponent*float64(in.Deref()-'0')
		currentExponent /= 10
		in.Inc()
	}
	if in.Deref() != 'e' {
		fmt.Printf("PARSE ERROR! Unexpected char: %c\n", rune(in.Deref()))
		os.Exit(3)
	}
	in.Inc()                      // skip dot
	exponent = gluco2ParseInt(in) // read exponent
	accu *= math.Pow(10, float64(exponent))
	if neg {
		return -accu
	}
	return accu
}
