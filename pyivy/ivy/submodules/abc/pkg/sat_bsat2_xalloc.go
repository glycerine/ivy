package abc

// Port of src/sat/bsat2/XAlloc.h (namespace Minisat).

// MinisatOutOfMemoryException is `class OutOfMemoryException{}`.
type MinisatOutOfMemoryException struct{}

// minisatXrealloc is Minisat::xrealloc (static inline): realloc that throws
// OutOfMemoryException on failure. Go allocation does not fail recoverably, so
// the result is a copy of ptr grown (or shrunk) to size bytes.
// C: static inline void* xrealloc(void *ptr, size_t size)
func minisatXrealloc(ptr []byte, size int) []byte {
	mem := make([]byte, size)
	copy(mem, ptr)
	return mem
}
