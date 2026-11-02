package abc

// Hand-checked skeleton of src/sat/glucose2/XAlloc.h (namespace Gluco2).

// Gluco2OutOfMemoryException is `class OutOfMemoryException{}`.
type Gluco2OutOfMemoryException struct{}

// gluco2Xrealloc is Gluco2::xrealloc (static inline); it throws OutOfMemoryException on failure.
// C: static inline void* xrealloc(void *ptr, size_t size)
func gluco2Xrealloc(ptr []byte, size int) []byte {
	panic("TODO, stubbed")
}
