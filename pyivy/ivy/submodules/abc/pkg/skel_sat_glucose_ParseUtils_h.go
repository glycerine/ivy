package abc

// Hand-checked skeleton of src/sat/glucose/ParseUtils.h (namespace Gluco).

// glucoBufferSize is `static const int buffer_size = 1048576`.
const glucoBufferSize = 1048576

// GlucoCharStream is what the parsing templates (B in skipWhitespace, skipLine,
// parseInt, eagerMatch) require of their input: `*in` (Deref) and `++in` (Inc).
// It is implemented by StreamBuffer and by GlucoCharPtr.
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

// Deref is `*in`.
// C: *in
func (p *GlucoCharPtr) Deref() int {
	panic("TODO, stubbed")
}

// Inc is `++in`.
// C: ++in
func (p *GlucoCharPtr) Inc() {
	panic("TODO, stubbed")
}

// At is `in[i]`.
// C: in[i]
func (p *GlucoCharPtr) At(i int) int {
	panic("TODO, stubbed")
}

// Advance is `in += i`.
// C: in += i
func (p *GlucoCharPtr) Advance(i int) {
	panic("TODO, stubbed")
}

// GlucoStreamBuffer is Gluco::StreamBuffer: a buffered reader over a gzFile.
type GlucoStreamBuffer struct {
	in   *GzState // gzFile
	buf  [glucoBufferSize]uint8
	pos  int
	size int
}

// assureLookahead is StreamBuffer::assureLookahead (private).
// C: void assureLookahead()
func (b *GlucoStreamBuffer) assureLookahead() {
	panic("TODO, stubbed")
}

// NewGlucoStreamBuffer is the explicit constructor StreamBuffer(gzFile i).
// C: explicit StreamBuffer(gzFile i)
func NewGlucoStreamBuffer(i *GzState) *GlucoStreamBuffer {
	panic("TODO, stubbed")
}

// Deref is `int operator*() const`.
// C: int operator*() const
func (b *GlucoStreamBuffer) Deref() int {
	panic("TODO, stubbed")
}

// Inc is `void operator++()`.
// C: void operator++()
func (b *GlucoStreamBuffer) Inc() {
	panic("TODO, stubbed")
}

// Position is StreamBuffer::position.
// C: int position() const
func (b *GlucoStreamBuffer) Position() int {
	panic("TODO, stubbed")
}

// glucoIsEofStream is the overload isEof(StreamBuffer&) (static inline).
// C: static inline bool isEof(StreamBuffer& in)
func glucoIsEofStream(in *GlucoStreamBuffer) bool {
	panic("TODO, stubbed")
}

// glucoIsEofChars is the overload isEof(const char*) (static inline).
// C: static inline bool isEof(const char* in)
func glucoIsEofChars(in *GlucoCharPtr) bool {
	panic("TODO, stubbed")
}

// glucoSkipWhitespace is Gluco::skipWhitespace (static template).
// C: template<class B> static void skipWhitespace(B& in)
func glucoSkipWhitespace[B GlucoCharStream](in B) {
	panic("TODO, stubbed")
}

// glucoSkipLine is Gluco::skipLine (static template; uses isEof).
// C: template<class B> static void skipLine(B& in)
func glucoSkipLine[B GlucoCharStream](in B) {
	panic("TODO, stubbed")
}

// glucoParseInt is Gluco::parseInt (static template).
// C: template<class B> static int parseInt(B& in)
func glucoParseInt[B GlucoCharStream](in B) int {
	panic("TODO, stubbed")
}

// glucoMatch is Gluco::match (static template; only instantiated with const char*).
// C: template<class B> static bool match(B& in, const char* str)
func glucoMatch(in *GlucoCharPtr, str string) bool {
	panic("TODO, stubbed")
}

// glucoEagerMatch is Gluco::eagerMatch (static template).
// C: template<class B> static bool eagerMatch(B& in, const char* str)
func glucoEagerMatch[B GlucoCharStream](in B, str string) bool {
	panic("TODO, stubbed")
}

// glucoParseDouble is parseDouble (static template; only the form X.XXXXXe-XX is accepted).
// C: template<class B> static double parseDouble(B& in)
func glucoParseDouble[B GlucoCharStream](in B) float64 {
	panic("TODO, stubbed")
}
