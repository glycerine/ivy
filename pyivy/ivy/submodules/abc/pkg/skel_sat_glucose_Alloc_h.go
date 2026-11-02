package abc

// Hand-checked skeleton of src/sat/glucose/Alloc.h (namespace Gluco).

// GlucoRegionAllocatorRef is RegionAllocator<T>::Ref (typedef uint32_t).
type GlucoRegionAllocatorRef = uint32

// RegionAllocator<T> anonymous enums.
const (
	GlucoRegionAllocatorRefUndef = ^uint32(0) // Ref_Undef = UINT32_MAX
	GlucoRegionAllocatorUnitSize = 4          // Unit_Size = sizeof(uint32_t)
)

// GlucoRegionAllocator is Gluco::RegionAllocator<T>.
type GlucoRegionAllocator[T any] struct {
	memory  []T
	sz      uint32
	cap     uint32
	wasted_ uint32
}

// capacity is RegionAllocator::capacity (private).
// C: void capacity(uint32_t min_cap)
func (a *GlucoRegionAllocator[T]) capacity(minCap uint32) {
	panic("TODO, stubbed")
}

// NewGlucoRegionAllocator is the explicit constructor (start_cap defaults to 1024*1024).
// C: explicit RegionAllocator(uint32_t start_cap = 1024*1024)
func NewGlucoRegionAllocator[T any](startCap uint32) *GlucoRegionAllocator[T] {
	panic("TODO, stubbed")
}

// Destroy is ~RegionAllocator().
// C: ~RegionAllocator()
func (a *GlucoRegionAllocator[T]) Destroy() {
	panic("TODO, stubbed")
}

// Size is RegionAllocator::size.
// C: uint32_t size() const
func (a *GlucoRegionAllocator[T]) Size() uint32 {
	panic("TODO, stubbed")
}

// Wasted is RegionAllocator::wasted.
// C: uint32_t wasted() const
func (a *GlucoRegionAllocator[T]) Wasted() uint32 {
	panic("TODO, stubbed")
}

// Alloc is RegionAllocator::alloc.
// C: Ref alloc(int size)
func (a *GlucoRegionAllocator[T]) Alloc(size int) uint32 {
	panic("TODO, stubbed")
}

// Free_ is RegionAllocator::free_ (renamed from _free in Minisat).
// C: void free_(int size)
func (a *GlucoRegionAllocator[T]) Free_(size int) {
	panic("TODO, stubbed")
}

// Clear is RegionAllocator::clear: forgets all allocations.
// C: void clear()
func (a *GlucoRegionAllocator[T]) Clear() {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](Ref r)
func (a *GlucoRegionAllocator[T]) At(r uint32) *T {
	panic("TODO, stubbed")
}

// Lea is RegionAllocator::lea (const and non-const overloads).
// C: T* lea(Ref r)
func (a *GlucoRegionAllocator[T]) Lea(r uint32) *T {
	panic("TODO, stubbed")
}

// Ael is RegionAllocator::ael.
// C: Ref ael(const T* t)
func (a *GlucoRegionAllocator[T]) Ael(t *T) uint32 {
	panic("TODO, stubbed")
}

// MoveTo is RegionAllocator::moveTo.
// C: void moveTo(RegionAllocator& to)
func (a *GlucoRegionAllocator[T]) MoveTo(to *GlucoRegionAllocator[T]) {
	panic("TODO, stubbed")
}
