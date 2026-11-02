package abc

// Port of src/sat/glucose/XAlloc.h (namespace Gluco).

// GlucoOutOfMemoryException is `class OutOfMemoryException{}`.
type GlucoOutOfMemoryException struct{}

// glucoXrealloc is Gluco::xrealloc (static inline): realloc that throws
// OutOfMemoryException on failure. Go allocation does not fail recoverably, so
// the result is a copy of ptr grown (or shrunk) to size bytes.
// C: static inline void* xrealloc(void *ptr, size_t size)
func glucoXrealloc(ptr []byte, size int) []byte {
	mem := make([]byte, size)
	copy(mem, ptr)
	return mem
}
