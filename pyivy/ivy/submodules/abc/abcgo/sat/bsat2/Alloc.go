package bsat2

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

type Ref uint32

const RefUndef Ref = ^Ref(0)

// RegionAllocator keeps stable offset references across backing-slice growth.
// Freed storage is accounted as waste and reclaimed by moving live objects.
type RegionAllocator[T any] struct {
	memory []T
	wasted uint32
}

func NewRegionAllocator[T any](capacity uint32) *RegionAllocator[T] {
	r := new(RegionAllocator[T])
	r.Reserve(capacity)
	return r
}
func (r *RegionAllocator[T]) Size() uint32   { return uint32(len(r.memory)) }
func (r *RegionAllocator[T]) Wasted() uint32 { return r.wasted }
func (r *RegionAllocator[T]) Reserve(minimum uint32) {
	if uint64(minimum) > uint64(int(^uint(0)>>1)) {
		panic("region capacity overflow")
	}
	capacity := uint32(cap(r.memory))
	for capacity < minimum {
		delta := ((capacity >> 1) + (capacity >> 3) + 2) &^ 1
		next := capacity + delta
		if next <= capacity {
			panic("region capacity overflow")
		}
		capacity = next
	}
	if int(capacity) == cap(r.memory) {
		return
	}
	memory := make([]T, len(r.memory), int(capacity))
	copy(memory, r.memory)
	r.memory = memory
}
func (r *RegionAllocator[T]) Alloc(size uint32) Ref {
	if size == 0 || uint64(size)+uint64(len(r.memory)) >= uint64(RefUndef) {
		panic("invalid region allocation")
	}
	start := len(r.memory)
	r.Reserve(uint32(start) + size)
	r.memory = r.memory[:start+int(size)]
	return Ref(start)
}
func (r *RegionAllocator[T]) Free(size uint32) {
	if uint64(r.wasted)+uint64(size) > uint64(len(r.memory)) {
		panic("region waste exceeds allocation")
	}
	r.wasted += size
}
func (r *RegionAllocator[T]) At(ref Ref) *T { return &r.memory[int(ref)] }
func (r *RegionAllocator[T]) Slice(ref Ref, size uint32) []T {
	end := uint64(ref) + uint64(size)
	if end > uint64(len(r.memory)) {
		panic("region slice outside allocation")
	}
	return r.memory[int(ref):int(end)]
}
func (r *RegionAllocator[T]) MoveTo(to *RegionAllocator[T]) {
	if r == to {
		return
	}
	*to = *r
	*r = RegionAllocator[T]{}
}
