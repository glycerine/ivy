package abc

// Hand-checked skeleton of src/sat/bsat2/Alloc.h (namespace Minisat).

// MinisatRegionAllocatorRef is RegionAllocator<T>::Ref (typedef uint32_t).
type MinisatRegionAllocatorRef = uint32

// RegionAllocator<T> anonymous enums.
const (
	MinisatRegionAllocatorRefUndef = ^uint32(0) // Ref_Undef = UINT32_MAX
	MinisatRegionAllocatorUnitSize = 4          // Unit_Size = sizeof(uint32_t)
)

// MinisatRegionAllocator is Minisat::RegionAllocator<T>.
type MinisatRegionAllocator[T any] struct {
	memory  []T
	sz      uint32
	cap     uint32
	wasted_ uint32
}

// capacity is RegionAllocator::capacity (private).
// C: void capacity(uint32_t min_cap)
func (a *MinisatRegionAllocator[T]) capacity(minCap uint32) {
	panic("TODO, stubbed")
}

// NewMinisatRegionAllocator is the explicit constructor (start_cap defaults to 1024*1024).
// C: explicit RegionAllocator(uint32_t start_cap = 1024*1024)
func NewMinisatRegionAllocator[T any](startCap uint32) *MinisatRegionAllocator[T] {
	panic("TODO, stubbed")
}

// Destroy is ~RegionAllocator().
// C: ~RegionAllocator()
func (a *MinisatRegionAllocator[T]) Destroy() {
	panic("TODO, stubbed")
}

// Size is RegionAllocator::size.
// C: uint32_t size() const
func (a *MinisatRegionAllocator[T]) Size() uint32 {
	panic("TODO, stubbed")
}

// Wasted is RegionAllocator::wasted.
// C: uint32_t wasted() const
func (a *MinisatRegionAllocator[T]) Wasted() uint32 {
	panic("TODO, stubbed")
}

// Alloc is RegionAllocator::alloc.
// C: Ref alloc(int size)
func (a *MinisatRegionAllocator[T]) Alloc(size int) uint32 {
	panic("TODO, stubbed")
}

// Free is RegionAllocator::_free.
// C: void _free(int size)
func (a *MinisatRegionAllocator[T]) Free(size int) {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](Ref r)
func (a *MinisatRegionAllocator[T]) At(r uint32) *T {
	panic("TODO, stubbed")
}

// Lea is RegionAllocator::lea (const and non-const overloads).
// C: T* lea(Ref r)
func (a *MinisatRegionAllocator[T]) Lea(r uint32) *T {
	panic("TODO, stubbed")
}

// Ael is RegionAllocator::ael.
// C: Ref ael(const T* t)
func (a *MinisatRegionAllocator[T]) Ael(t *T) uint32 {
	panic("TODO, stubbed")
}

// MoveTo is RegionAllocator::moveTo.
// C: void moveTo(RegionAllocator& to)
func (a *MinisatRegionAllocator[T]) MoveTo(to *MinisatRegionAllocator[T]) {
	panic("TODO, stubbed")
}
