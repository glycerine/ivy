package abc

// Port of src/sat/glucose2/XAlloc.h (namespace Gluco2).

// Gluco2OutOfMemoryException is `class OutOfMemoryException{}`.
type Gluco2OutOfMemoryException struct{}

// gluco2Xrealloc is Gluco2::xrealloc (static inline): realloc that throws
// OutOfMemoryException on failure. Go allocation does not fail recoverably, so
// the result is a copy of ptr grown (or shrunk) to size bytes.
// C: static inline void* xrealloc(void *ptr, size_t size)
func gluco2Xrealloc(ptr []byte, size int) []byte {
	mem := make([]byte, size)
	copy(mem, ptr)
	return mem
}
