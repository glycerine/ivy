package abc

import "unsafe"

// Port of src/sat/glucose2/Alloc.h (namespace Gluco2).

// Gluco2RegionAllocatorRef is RegionAllocator<T>::Ref (typedef uint32_t).
type Gluco2RegionAllocatorRef = uint32

// RegionAllocator<T> anonymous enums.
const (
	Gluco2RegionAllocatorRefUndef = ^uint32(0) // Ref_Undef = UINT32_MAX
	Gluco2RegionAllocatorUnitSize = 4          // Unit_Size = sizeof(uint32_t)
)

// Gluco2RegionAllocator is Gluco2::RegionAllocator<T>. memory has length cap.
type Gluco2RegionAllocator[T any] struct {
	memory  []T
	sz      uint32
	cap     uint32
	wasted_ uint32
}

// capacity is RegionAllocator::capacity (private).
// C: void capacity(uint32_t min_cap)
func (a *Gluco2RegionAllocator[T]) capacity(minCap uint32) {
	if a.cap >= minCap {
		return
	}
	prevCap := a.cap
	for a.cap < minCap {
		// All constants here are even, so the capacity stays even.
		delta := ((a.cap >> 1) + (a.cap >> 3) + 2) &^ 1
		a.cap += delta
		if a.cap <= prevCap {
			panic(Gluco2OutOfMemoryException{})
		}
	}
	if a.cap == 0 {
		panic("Gluco2RegionAllocator.capacity: cap == 0")
	}
	nm := make([]T, a.cap)
	copy(nm, a.memory)
	a.memory = nm
}

// NewGluco2RegionAllocator is the explicit constructor (start_cap defaults to 1024*1024).
// C: explicit RegionAllocator(uint32_t start_cap = 1024*1024)
func NewGluco2RegionAllocator[T any](startCap uint32) *Gluco2RegionAllocator[T] {
	a := &Gluco2RegionAllocator[T]{}
	a.capacity(startCap)
	return a
}

// Destroy is ~RegionAllocator().
// C: ~RegionAllocator()
func (a *Gluco2RegionAllocator[T]) Destroy() {
	a.memory = nil
}

// Size is RegionAllocator::size.
// C: uint32_t size() const
func (a *Gluco2RegionAllocator[T]) Size() uint32 { return a.sz }

// Wasted is RegionAllocator::wasted.
// C: uint32_t wasted() const
func (a *Gluco2RegionAllocator[T]) Wasted() uint32 { return a.wasted_ }

// Alloc is RegionAllocator::alloc.
// C: Ref alloc(int size)
func (a *Gluco2RegionAllocator[T]) Alloc(size int) uint32 {
	if size <= 0 {
		panic("Gluco2RegionAllocator.Alloc: size <= 0")
	}
	a.capacity(a.sz + uint32(size))
	prevSz := a.sz
	a.sz += uint32(size)
	// Handle overflow:
	if a.sz < prevSz {
		panic(Gluco2OutOfMemoryException{})
	}
	return prevSz
}

// Free_ is RegionAllocator::free_ (renamed from _free in Minisat).
// C: void free_(int size)
func (a *Gluco2RegionAllocator[T]) Free_(size int) { a.wasted_ += uint32(size) }

// Clear is RegionAllocator::clear: forgets all allocations.
// C: void clear()
func (a *Gluco2RegionAllocator[T]) Clear() {
	a.sz = 0
	a.wasted_ = 0
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](Ref r)
func (a *Gluco2RegionAllocator[T]) At(r uint32) *T {
	if r >= a.sz {
		panic("Gluco2RegionAllocator.At: reference out of range")
	}
	return &a.memory[r]
}

// Lea is RegionAllocator::lea (const and non-const overloads).
// C: T* lea(Ref r)
func (a *Gluco2RegionAllocator[T]) Lea(r uint32) *T {
	if r >= a.sz {
		panic("Gluco2RegionAllocator.Lea: reference out of range")
	}
	return &a.memory[r]
}

// Ael is RegionAllocator::ael: the reference of a pointer into memory (pointer subtraction).
// C: Ref ael(const T* t)
func (a *Gluco2RegionAllocator[T]) Ael(t *T) uint32 {
	base := uintptr(unsafe.Pointer(&a.memory[0]))
	p := uintptr(unsafe.Pointer(t))
	var zero T
	if p < base || p >= base+uintptr(a.sz)*unsafe.Sizeof(zero) {
		panic("Gluco2RegionAllocator.Ael: pointer outside the region")
	}
	return uint32((p - base) / unsafe.Sizeof(zero))
}

// MoveTo is RegionAllocator::moveTo.
// C: void moveTo(RegionAllocator& to)
func (a *Gluco2RegionAllocator[T]) MoveTo(to *Gluco2RegionAllocator[T]) {
	to.memory = a.memory
	to.sz = a.sz
	to.cap = a.cap
	to.wasted_ = a.wasted_
	a.memory = nil
	a.sz = 0
	a.cap = 0
	a.wasted_ = 0
}
