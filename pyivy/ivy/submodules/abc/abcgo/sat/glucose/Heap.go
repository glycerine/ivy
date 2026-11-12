package glucose

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

// Heap is an indexed integer heap. The comparator may read mutable priorities;
// callers notify the heap with Update, Decrease, or Increase after changing one.
type Heap struct {
	less      func(int, int) bool
	nodes     []int
	positions []int
}

func NewHeap(less func(int, int) bool) *Heap { return &Heap{less: less} }
func (h *Heap) Size() int                    { return len(h.nodes) }
func (h *Heap) Empty() bool                  { return len(h.nodes) == 0 }
func (h *Heap) InHeap(n int) bool            { return n >= 0 && n < len(h.positions) && h.positions[n] >= 0 }
func (h *Heap) At(i int) int                 { return h.nodes[i] }
func (h *Heap) up(i int) {
	x := h.nodes[i]
	for i > 0 {
		p := (i - 1) / 2
		if !h.less(x, h.nodes[p]) {
			break
		}
		h.nodes[i] = h.nodes[p]
		h.positions[h.nodes[i]] = i
		i = p
	}
	h.nodes[i] = x
	h.positions[x] = i
}
func (h *Heap) down(i int) {
	x := h.nodes[i]
	for 2*i+1 < len(h.nodes) {
		c := 2*i + 1
		if c+1 < len(h.nodes) && h.less(h.nodes[c+1], h.nodes[c]) {
			c++
		}
		if !h.less(h.nodes[c], x) {
			break
		}
		h.nodes[i] = h.nodes[c]
		h.positions[h.nodes[i]] = i
		i = c
	}
	h.nodes[i] = x
	h.positions[x] = i
}
func (h *Heap) Decrease(n int) {
	if !h.InHeap(n) {
		panic("heap member missing")
	}
	h.up(h.positions[n])
}
func (h *Heap) Increase(n int) {
	if !h.InHeap(n) {
		panic("heap member missing")
	}
	h.down(h.positions[n])
}
func (h *Heap) Update(n int) {
	if !h.InHeap(n) {
		h.Insert(n)
		return
	}
	h.up(h.positions[n])
	h.down(h.positions[n])
}
func (h *Heap) Insert(n int) {
	if n < 0 {
		panic("negative heap key")
	}
	for len(h.positions) <= n {
		h.positions = append(h.positions, -1)
	}
	if h.InHeap(n) {
		panic("duplicate heap key")
	}
	h.positions[n] = len(h.nodes)
	h.nodes = append(h.nodes, n)
	h.up(len(h.nodes) - 1)
}
func (h *Heap) RemoveMin() int {
	x := h.nodes[0]
	last := h.nodes[len(h.nodes)-1]
	h.nodes = h.nodes[:len(h.nodes)-1]
	h.positions[x] = -1
	if len(h.nodes) > 0 {
		h.nodes[0] = last
		h.positions[last] = 0
		if len(h.nodes) > 1 {
			h.down(0)
		}
	}
	return x
}
func (h *Heap) Build(ns []int) {
	h.Clear()
	for _, n := range ns {
		if n < 0 {
			panic("negative heap key")
		}
		for len(h.positions) <= n {
			h.positions = append(h.positions, -1)
		}
		if h.InHeap(n) {
			panic("duplicate heap key")
		}
		h.positions[n] = len(h.nodes)
		h.nodes = append(h.nodes, n)
	}
	for i := len(h.nodes)/2 - 1; i >= 0; i-- {
		h.down(i)
	}
}
func (h *Heap) Clear() {
	for _, n := range h.nodes {
		h.positions[n] = -1
	}
	h.nodes = h.nodes[:0]
}
