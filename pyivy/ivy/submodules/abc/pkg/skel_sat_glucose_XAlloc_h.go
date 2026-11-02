package abc

// Hand-checked skeleton of src/sat/glucose/XAlloc.h (namespace Gluco).

// GlucoOutOfMemoryException is `class OutOfMemoryException{}`.
type GlucoOutOfMemoryException struct{}

// glucoXrealloc is Gluco::xrealloc (static inline); it throws OutOfMemoryException on failure.
// C: static inline void* xrealloc(void *ptr, size_t size)
func glucoXrealloc(ptr []byte, size int) []byte {
	panic("TODO, stubbed")
}
