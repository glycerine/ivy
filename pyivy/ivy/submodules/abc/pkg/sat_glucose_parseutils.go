package abc

import (
	"fmt"
	"math"
	"os"
)

// Port of src/sat/glucose/ParseUtils.h (namespace Gluco).

// glucoBufferSize is `static const int buffer_size = 1048576`.
const glucoBufferSize = 1048576

// GlucoCharStream is what the parsing templates (B in skipWhitespace, skipLine,
// parseInt, eagerMatch) require of their input: `*in` (Deref) and `++in` (Inc).
// It is implemented by StreamBuffer and by GlucoCharPtr. Deref returns EOF (-1)
// at the end of a stream and 0 at the end of a character string.
type GlucoCharStream interface {
	Deref() int
	Inc()
}

// GlucoCharPtr is a `const char*` cursor into a NUL-terminated string, as used
// with the string instantiations of the templates (match, isEof(const char*)).
type GlucoCharPtr struct {
	s   string
	pos int
}

// NewGlucoCharPtr starts a cursor at the beginning of s.
func NewGlucoCharPtr(s string) *GlucoCharPtr { return &GlucoCharPtr{s: s} }

// Deref is `*in`.
// C: *in
func (p *GlucoCharPtr) Deref() int { return p.At(0) }

// Inc is `++in`.
// C: ++in
func (p *GlucoCharPtr) Inc() { p.pos++ }

// At is `in[i]` (0 past the end: the terminating NUL).
// C: in[i]
func (p *GlucoCharPtr) At(i int) int {
	if p.pos+i >= len(p.s) {
		return 0
	}
	return int(p.s[p.pos+i])
}

// Advance is `in += i`.
// C: in += i
func (p *GlucoCharPtr) Advance(i int) { p.pos += i }

// Rest is the string the cursor points at.
func (p *GlucoCharPtr) Rest() string {
	if p.pos >= len(p.s) {
		return ""
	}
	return p.s[p.pos:]
}

// glucoEOF is the C macro EOF.
const glucoEOF = -1

// GlucoStreamBuffer is Gluco::StreamBuffer: a buffered reader over a gzFile.
type GlucoStreamBuffer struct {
	in   *GZIPReader // gzFile
	buf  [glucoBufferSize]uint8
	pos  int
	size int
}

// assureLookahead is StreamBuffer::assureLookahead (private).
// C: void assureLookahead()
func (b *GlucoStreamBuffer) assureLookahead() {
	if b.pos >= b.size {
		b.pos = 0
		n, _ := b.in.Read(b.buf[:]) // gzread: a negative result (error) is treated as no data
		if n < 0 {
			n = 0
		}
		b.size = n
	}
}

// NewGlucoStreamBuffer is the explicit constructor StreamBuffer(gzFile i).
// C: explicit StreamBuffer(gzFile i)
func NewGlucoStreamBuffer(i *GZIPReader) *GlucoStreamBuffer {
	b := &GlucoStreamBuffer{in: i}
	b.assureLookahead()
	return b
}

// Deref is `int operator*() const`.
// C: int operator*() const
func (b *GlucoStreamBuffer) Deref() int {
	if b.pos >= b.size {
		return glucoEOF
	}
	return int(b.buf[b.pos])
}

// Inc is `void operator++()`.
// C: void operator++()
func (b *GlucoStreamBuffer) Inc() {
	b.pos++
	b.assureLookahead()
}

// Position is StreamBuffer::position.
// C: int position() const
func (b *GlucoStreamBuffer) Position() int { return b.pos }

// glucoIsEofStream is the overload isEof(StreamBuffer&) (static inline).
// C: static inline bool isEof(StreamBuffer& in)
func glucoIsEofStream(in *GlucoStreamBuffer) bool { return in.Deref() == glucoEOF }

// glucoIsEofChars is the overload isEof(const char*) (static inline).
// C: static inline bool isEof(const char* in)
func glucoIsEofChars(in *GlucoCharPtr) bool { return in.Deref() == 0 }

// glucoIsEof dispatches to the isEof overload of the stream type.
func glucoIsEof(in GlucoCharStream) bool {
	switch v := in.(type) {
	case *GlucoStreamBuffer:
		return glucoIsEofStream(v)
	case *GlucoCharPtr:
		return glucoIsEofChars(v)
	}
	panic("glucoIsEof: unsupported stream type")
}

// glucoSkipWhitespace is Gluco::skipWhitespace (static template).
// C: template<class B> static void skipWhitespace(B& in)
func glucoSkipWhitespace[B GlucoCharStream](in B) {
	for (in.Deref() >= 9 && in.Deref() <= 13) || in.Deref() == 32 {
		in.Inc()
	}
}

// glucoSkipLine is Gluco::skipLine (static template; uses isEof).
// C: template<class B> static void skipLine(B& in)
func glucoSkipLine[B GlucoCharStream](in B) {
	for {
		if glucoIsEof(in) {
			return
		}
		if in.Deref() == '\n' {
			in.Inc()
			return
		}
		in.Inc()
	}
}

// glucoParseInt is Gluco::parseInt (static template); an unexpected character
// prints a message and exits with status 3, as in C.
// C: template<class B> static int parseInt(B& in)
func glucoParseInt[B GlucoCharStream](in B) int {
	val := 0
	neg := false
	glucoSkipWhitespace(in)
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

// glucoMatch is Gluco::match (static template; only instantiated with const char*).
// C: template<class B> static bool match(B& in, const char* str)
func glucoMatch(in *GlucoCharPtr, str string) bool {
	i := 0
	for ; i < len(str); i++ {
		if in.At(i) != int(str[i]) {
			return false
		}
	}
	in.Advance(i)
	return true
}

// glucoEagerMatch is Gluco::eagerMatch (static template).
// C: template<class B> static bool eagerMatch(B& in, const char* str)
func glucoEagerMatch[B GlucoCharStream](in B, str string) bool {
	for i := 0; i < len(str); i++ {
		if int(str[i]) != in.Deref() {
			return false
		}
		in.Inc()
	}
	return true
}

// glucoParseDouble is parseDouble (static template; only the form X.XXXXXe-XX is accepted).
// C: template<class B> static double parseDouble(B& in)
func glucoParseDouble[B GlucoCharStream](in B) float64 {
	neg := false
	accu := 0.0
	var currentExponent float64
	var exponent int
	glucoSkipWhitespace(in)
	if in.Deref() == glucoEOF {
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
	in.Inc()                     // skip dot
	exponent = glucoParseInt(in) // read exponent
	accu *= math.Pow(10, float64(exponent))
	if neg {
		return -accu
	}
	return accu
}
