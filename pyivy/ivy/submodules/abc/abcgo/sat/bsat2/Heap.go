package bsat2

/******************************************************************************************[Heap.h]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

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

// Heap is an indexed minimum heap. Less may observe mutable activity values;
// Update restores the heap after either an increase or decrease in priority.
type Heap struct {
	heap, indices []int32
	less          func(int32, int32) bool
}

func NewHeap(less func(int32, int32) bool) *Heap {
	if less == nil {
		panic("nil heap comparison")
	}
	return &Heap{less: less}
}
func (h *Heap) Len() int              { return len(h.heap) }
func (h *Heap) Contains(n int32) bool { return n >= 0 && int(n) < len(h.indices) && h.indices[n] >= 0 }
func (h *Heap) At(i int) int32        { return h.heap[i] }
func (h *Heap) up(i int) {
	x := h.heap[i]
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(x, h.heap[parent]) {
			break
		}
		h.heap[i] = h.heap[parent]
		h.indices[h.heap[i]] = int32(i)
		i = parent
	}
	h.heap[i] = x
	h.indices[x] = int32(i)
}
func (h *Heap) down(i int) {
	x := h.heap[i]
	for left := 2*i + 1; left < len(h.heap); left = 2*i + 1 {
		child := left
		if left+1 < len(h.heap) && h.less(h.heap[left+1], h.heap[left]) {
			child++
		}
		if !h.less(h.heap[child], x) {
			break
		}
		h.heap[i] = h.heap[child]
		h.indices[h.heap[i]] = int32(i)
		i = child
	}
	h.heap[i] = x
	h.indices[x] = int32(i)
}
func (h *Heap) Insert(n int32) {
	if n < 0 || h.Contains(n) {
		panic("invalid heap insertion")
	}
	for len(h.indices) <= int(n) {
		h.indices = append(h.indices, -1)
	}
	h.indices[n] = int32(len(h.heap))
	h.heap = append(h.heap, n)
	h.up(len(h.heap) - 1)
}
func (h *Heap) Update(n int32) {
	if !h.Contains(n) {
		h.Insert(n)
		return
	}
	h.up(int(h.indices[n]))
	h.down(int(h.indices[n]))
}
func (h *Heap) Decrease(n int32) {
	if !h.Contains(n) {
		panic("missing heap item")
	}
	h.up(int(h.indices[n]))
}
func (h *Heap) Increase(n int32) {
	if !h.Contains(n) {
		panic("missing heap item")
	}
	h.down(int(h.indices[n]))
}
func (h *Heap) RemoveMin() int32 {
	if len(h.heap) == 0 {
		panic("remove from empty heap")
	}
	x := h.heap[0]
	last := h.heap[len(h.heap)-1]
	h.heap = h.heap[:len(h.heap)-1]
	h.indices[x] = -1
	if len(h.heap) > 0 {
		h.heap[0] = last
		h.indices[last] = 0
		h.down(0)
	}
	return x
}
func (h *Heap) Clear(deallocate bool) {
	for _, x := range h.heap {
		h.indices[x] = -1
	}
	h.heap = h.heap[:0]
	if deallocate {
		h.heap = nil
		h.indices = nil
	}
}
func (h *Heap) Build(elements []int32) {
	h.Clear(false)
	for _, x := range elements {
		if x < 0 {
			panic("negative heap item")
		}
		for len(h.indices) <= int(x) {
			h.indices = append(h.indices, -1)
		}
		if h.indices[x] >= 0 {
			panic("duplicate heap item")
		}
		h.indices[x] = int32(len(h.heap))
		h.heap = append(h.heap, x)
	}
	for i := len(h.heap)/2 - 1; i >= 0; i-- {
		h.down(i)
	}
}
