package abc

// Hand-checked skeleton of src/sat/bsat2/ParseUtils.h (namespace Minisat).

// minisatBufferSize is `static const int buffer_size = 1048576`.
const minisatBufferSize = 1048576

// MinisatCharStream is what the parsing templates (B in skipWhitespace, skipLine,
// parseInt, eagerMatch) require of their input: `*in` (Deref) and `++in` (Inc).
// It is implemented by StreamBuffer and by MinisatCharPtr.
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

// Deref is `*in`.
// C: *in
func (p *MinisatCharPtr) Deref() int {
	panic("TODO, stubbed")
}

// Inc is `++in`.
// C: ++in
func (p *MinisatCharPtr) Inc() {
	panic("TODO, stubbed")
}

// At is `in[i]`.
// C: in[i]
func (p *MinisatCharPtr) At(i int) int {
	panic("TODO, stubbed")
}

// Advance is `in += i`.
// C: in += i
func (p *MinisatCharPtr) Advance(i int) {
	panic("TODO, stubbed")
}

// MinisatStreamBuffer is Minisat::StreamBuffer: a buffered reader over a gzFile.
type MinisatStreamBuffer struct {
	in   *GzState // gzFile
	buf  [minisatBufferSize]uint8
	pos  int
	size int
}

// assureLookahead is StreamBuffer::assureLookahead (private).
// C: void assureLookahead()
func (b *MinisatStreamBuffer) assureLookahead() {
	panic("TODO, stubbed")
}

// NewMinisatStreamBuffer is the explicit constructor StreamBuffer(gzFile i).
// C: explicit StreamBuffer(gzFile i)
func NewMinisatStreamBuffer(i *GzState) *MinisatStreamBuffer {
	panic("TODO, stubbed")
}

// Deref is `int operator*() const`.
// C: int operator*() const
func (b *MinisatStreamBuffer) Deref() int {
	panic("TODO, stubbed")
}

// Inc is `void operator++()`.
// C: void operator++()
func (b *MinisatStreamBuffer) Inc() {
	panic("TODO, stubbed")
}

// Position is StreamBuffer::position.
// C: int position() const
func (b *MinisatStreamBuffer) Position() int {
	panic("TODO, stubbed")
}

// minisatIsEofStream is the overload isEof(StreamBuffer&) (static inline).
// C: static inline bool isEof(StreamBuffer& in)
func minisatIsEofStream(in *MinisatStreamBuffer) bool {
	panic("TODO, stubbed")
}

// minisatIsEofChars is the overload isEof(const char*) (static inline).
// C: static inline bool isEof(const char* in)
func minisatIsEofChars(in *MinisatCharPtr) bool {
	panic("TODO, stubbed")
}

// minisatSkipWhitespace is Minisat::skipWhitespace (static template).
// C: template<class B> static void skipWhitespace(B& in)
func minisatSkipWhitespace[B MinisatCharStream](in B) {
	panic("TODO, stubbed")
}

// minisatSkipLine is Minisat::skipLine (static template; uses isEof).
// C: template<class B> static void skipLine(B& in)
func minisatSkipLine[B MinisatCharStream](in B) {
	panic("TODO, stubbed")
}

// minisatParseInt is Minisat::parseInt (static template).
// C: template<class B> static int parseInt(B& in)
func minisatParseInt[B MinisatCharStream](in B) int {
	panic("TODO, stubbed")
}

// minisatMatch is Minisat::match (static template; only instantiated with const char*).
// C: template<class B> static bool match(B& in, const char* str)
func minisatMatch(in *MinisatCharPtr, str string) bool {
	panic("TODO, stubbed")
}

// minisatEagerMatch is Minisat::eagerMatch (static template).
// C: template<class B> static bool eagerMatch(B& in, const char* str)
func minisatEagerMatch[B MinisatCharStream](in B, str string) bool {
	panic("TODO, stubbed")
}
