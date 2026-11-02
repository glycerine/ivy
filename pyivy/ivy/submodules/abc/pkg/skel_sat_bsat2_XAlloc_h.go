package abc

// Hand-checked skeleton of src/sat/bsat2/XAlloc.h (namespace Minisat).

// MinisatOutOfMemoryException is `class OutOfMemoryException{}`.
type MinisatOutOfMemoryException struct{}

// minisatXrealloc is Minisat::xrealloc (static inline); it throws OutOfMemoryException on failure.
// C: static inline void* xrealloc(void *ptr, size_t size)
func minisatXrealloc(ptr []byte, size int) []byte {
	panic("TODO, stubbed")
}
