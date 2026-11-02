package abc

import "unsafe"

// Port of src/sat/bsat2/Alloc.h (namespace Minisat).

// MinisatRegionAllocatorRef is RegionAllocator<T>::Ref (typedef uint32_t).
type MinisatRegionAllocatorRef = uint32

// RegionAllocator<T> anonymous enums.
const (
	MinisatRegionAllocatorRefUndef = ^uint32(0) // Ref_Undef = UINT32_MAX
	MinisatRegionAllocatorUnitSize = 4          // Unit_Size = sizeof(uint32_t)
)

// MinisatRegionAllocator is Minisat::RegionAllocator<T>. memory has length cap.
type MinisatRegionAllocator[T any] struct {
	memory  []T
	sz      uint32
	cap     uint32
	wasted_ uint32
}

// capacity is RegionAllocator::capacity (private).
// C: void capacity(uint32_t min_cap)
func (a *MinisatRegionAllocator[T]) capacity(minCap uint32) {
	if a.cap >= minCap {
		return
	}
	prevCap := a.cap
	for a.cap < minCap {
		// All constants here are even, so the capacity stays even.
		delta := ((a.cap >> 1) + (a.cap >> 3) + 2) &^ 1
		a.cap += delta
		if a.cap <= prevCap {
			panic(MinisatOutOfMemoryException{})
		}
	}
	if a.cap == 0 {
		panic("MinisatRegionAllocator.capacity: cap == 0")
	}
	nm := make([]T, a.cap)
	copy(nm, a.memory)
	a.memory = nm
}

// NewMinisatRegionAllocator is the explicit constructor (start_cap defaults to 1024*1024).
// C: explicit RegionAllocator(uint32_t start_cap = 1024*1024)
func NewMinisatRegionAllocator[T any](startCap uint32) *MinisatRegionAllocator[T] {
	a := &MinisatRegionAllocator[T]{}
	a.capacity(startCap)
	return a
}

// Destroy is ~RegionAllocator().
// C: ~RegionAllocator()
func (a *MinisatRegionAllocator[T]) Destroy() {
	a.memory = nil
}

// Size is RegionAllocator::size.
// C: uint32_t size() const
func (a *MinisatRegionAllocator[T]) Size() uint32 { return a.sz }

// Wasted is RegionAllocator::wasted.
// C: uint32_t wasted() const
func (a *MinisatRegionAllocator[T]) Wasted() uint32 { return a.wasted_ }

// Alloc is RegionAllocator::alloc.
// C: Ref alloc(int size)
func (a *MinisatRegionAllocator[T]) Alloc(size int) uint32 {
	if size <= 0 {
		panic("MinisatRegionAllocator.Alloc: size <= 0")
	}
	a.capacity(a.sz + uint32(size))
	prevSz := a.sz
	a.sz += uint32(size)
	// Handle overflow:
	if a.sz < prevSz {
		panic(MinisatOutOfMemoryException{})
	}
	return prevSz
}

// Free is RegionAllocator::_free.
// C: void _free(int size)
func (a *MinisatRegionAllocator[T]) Free(size int) { a.wasted_ += uint32(size) }

// At is operator[] (const and non-const overloads).
// C: T& operator[](Ref r)
func (a *MinisatRegionAllocator[T]) At(r uint32) *T {
	if r >= a.sz {
		panic("MinisatRegionAllocator.At: reference out of range")
	}
	return &a.memory[r]
}

// Lea is RegionAllocator::lea (const and non-const overloads).
// C: T* lea(Ref r)
func (a *MinisatRegionAllocator[T]) Lea(r uint32) *T {
	if r >= a.sz {
		panic("MinisatRegionAllocator.Lea: reference out of range")
	}
	return &a.memory[r]
}

// Ael is RegionAllocator::ael: the reference of a pointer into memory (pointer subtraction).
// C: Ref ael(const T* t)
func (a *MinisatRegionAllocator[T]) Ael(t *T) uint32 {
	base := uintptr(unsafe.Pointer(&a.memory[0]))
	p := uintptr(unsafe.Pointer(t))
	var zero T
	if p < base || p >= base+uintptr(a.sz)*unsafe.Sizeof(zero) {
		panic("MinisatRegionAllocator.Ael: pointer outside the region")
	}
	return uint32((p - base) / unsafe.Sizeof(zero))
}

// MoveTo is RegionAllocator::moveTo.
// C: void moveTo(RegionAllocator& to)
func (a *MinisatRegionAllocator[T]) MoveTo(to *MinisatRegionAllocator[T]) {
	to.memory = a.memory
	to.sz = a.sz
	to.cap = a.cap
	to.wasted_ = a.wasted_
	a.memory = nil
	a.sz = 0
	a.cap = 0
	a.wasted_ = 0
}
