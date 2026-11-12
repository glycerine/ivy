package glucose

/*****************************************************************************************[Alloc.h]
Copyright (c) 2008-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

const RegionRefUndef uint32 = ^uint32(0)

// RegionAllocator keeps integer references stable across slice reallocations.
// Borrowed Data slices are valid until the next allocation, just as in C++.
type RegionAllocator[T any] struct {
	memory []T
	wasted uint32
}

func NewRegionAllocator[T any](capacity int) *RegionAllocator[T] {
	if capacity < 0 {
		panic("negative region capacity")
	}
	return &RegionAllocator[T]{memory: make([]T, 0, capacity)}
}
func (a *RegionAllocator[T]) Size() uint32   { return uint32(len(a.memory)) }
func (a *RegionAllocator[T]) Wasted() uint32 { return a.wasted }
func (a *RegionAllocator[T]) Alloc(n int) uint32 {
	if n <= 0 || uint64(len(a.memory))+uint64(n) >= uint64(RegionRefUndef) {
		panic(OutOfMemoryException{})
	}
	start := uint32(len(a.memory))
	a.memory = append(a.memory, make([]T, n)...)
	return start
}
func (a *RegionAllocator[T]) Free(n uint32) {
	if uint64(a.wasted)+uint64(n) > uint64(len(a.memory)) {
		panic("region free exceeds allocation")
	}
	a.wasted += n
}
func (a *RegionAllocator[T]) At(ref uint32) *T { return &a.memory[ref] }
func (a *RegionAllocator[T]) Data(ref uint32) []T {
	if ref >= uint32(len(a.memory)) {
		panic("region reference out of range")
	}
	return a.memory[ref:]
}
func (a *RegionAllocator[T]) Clear() { clear(a.memory); a.memory = a.memory[:0]; a.wasted = 0 }
func (a *RegionAllocator[T]) MoveTo(to *RegionAllocator[T]) {
	if a == to {
		return
	}
	*to = *a
	*a = RegionAllocator[T]{}
}

// AEL is the safe inverse of At. It searches by pointer equality, including the
// final allocated element that the original strict-end assertion rejected.
func (a *RegionAllocator[T]) AEL(element *T) uint32 {
	for i := range a.memory {
		if &a.memory[i] == element {
			return uint32(i)
		}
	}
	panic("pointer is outside region")
}
