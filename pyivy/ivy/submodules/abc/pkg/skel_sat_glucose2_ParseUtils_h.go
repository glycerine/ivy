package abc

// Hand-checked skeleton of src/sat/glucose2/ParseUtils.h (namespace Gluco2).

// gluco2BufferSize is `static const int buffer_size = 1048576`.
const gluco2BufferSize = 1048576

// Gluco2CharStream is what the parsing templates (B in skipWhitespace, skipLine,
// parseInt, eagerMatch) require of their input: `*in` (Deref) and `++in` (Inc).
// It is implemented by StreamBuffer and by Gluco2CharPtr.
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

// Deref is `*in`.
// C: *in
func (p *Gluco2CharPtr) Deref() int {
	panic("TODO, stubbed")
}

// Inc is `++in`.
// C: ++in
func (p *Gluco2CharPtr) Inc() {
	panic("TODO, stubbed")
}

// At is `in[i]`.
// C: in[i]
func (p *Gluco2CharPtr) At(i int) int {
	panic("TODO, stubbed")
}

// Advance is `in += i`.
// C: in += i
func (p *Gluco2CharPtr) Advance(i int) {
	panic("TODO, stubbed")
}

// Gluco2StreamBuffer is Gluco2::StreamBuffer: a buffered reader over a gzFile.
type Gluco2StreamBuffer struct {
	in   *GzState // gzFile
	buf  [gluco2BufferSize]uint8
	pos  int
	size int
}

// assureLookahead is StreamBuffer::assureLookahead (private).
// C: void assureLookahead()
func (b *Gluco2StreamBuffer) assureLookahead() {
	panic("TODO, stubbed")
}

// NewGluco2StreamBuffer is the explicit constructor StreamBuffer(gzFile i).
// C: explicit StreamBuffer(gzFile i)
func NewGluco2StreamBuffer(i *GzState) *Gluco2StreamBuffer {
	panic("TODO, stubbed")
}

// Deref is `int operator*() const`.
// C: int operator*() const
func (b *Gluco2StreamBuffer) Deref() int {
	panic("TODO, stubbed")
}

// Inc is `void operator++()`.
// C: void operator++()
func (b *Gluco2StreamBuffer) Inc() {
	panic("TODO, stubbed")
}

// Position is StreamBuffer::position.
// C: int position() const
func (b *Gluco2StreamBuffer) Position() int {
	panic("TODO, stubbed")
}

// gluco2IsEofStream is the overload isEof(StreamBuffer&) (static inline).
// C: static inline bool isEof(StreamBuffer& in)
func gluco2IsEofStream(in *Gluco2StreamBuffer) bool {
	panic("TODO, stubbed")
}

// gluco2IsEofChars is the overload isEof(const char*) (static inline).
// C: static inline bool isEof(const char* in)
func gluco2IsEofChars(in *Gluco2CharPtr) bool {
	panic("TODO, stubbed")
}

// gluco2SkipWhitespace is Gluco2::skipWhitespace (static template).
// C: template<class B> static void skipWhitespace(B& in)
func gluco2SkipWhitespace[B Gluco2CharStream](in B) {
	panic("TODO, stubbed")
}

// gluco2SkipLine is Gluco2::skipLine (static template; uses isEof).
// C: template<class B> static void skipLine(B& in)
func gluco2SkipLine[B Gluco2CharStream](in B) {
	panic("TODO, stubbed")
}

// gluco2ParseInt is Gluco2::parseInt (static template).
// C: template<class B> static int parseInt(B& in)
func gluco2ParseInt[B Gluco2CharStream](in B) int {
	panic("TODO, stubbed")
}

// gluco2Match is Gluco2::match (static template; only instantiated with const char*).
// C: template<class B> static bool match(B& in, const char* str)
func gluco2Match(in *Gluco2CharPtr, str string) bool {
	panic("TODO, stubbed")
}

// gluco2EagerMatch is Gluco2::eagerMatch (static template).
// C: template<class B> static bool eagerMatch(B& in, const char* str)
func gluco2EagerMatch[B Gluco2CharStream](in B, str string) bool {
	panic("TODO, stubbed")
}

// gluco2ParseDouble is parseDouble (static template; only the form X.XXXXXe-XX is accepted).
// C: template<class B> static double parseDouble(B& in)
func gluco2ParseDouble[B Gluco2CharStream](in B) float64 {
	panic("TODO, stubbed")
}
