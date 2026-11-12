// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package utils

import "github.com/glycerine/abcgo/sat/satoko/utils/vec"

// Heap orders variable IDs by decreasing activity and keeps inverse positions.
type Heap struct {
	indices []int
	data    []uint32
	weights *vec.Activities
}
type heap_t = Heap

func NewHeap(weights *vec.Activities) *Heap      { return &Heap{weights: weights} }
func heap_alloc(weights *vec.Activities) *heap_t { return NewHeap(weights) }
func (h *Heap) Len() int                         { return len(h.data) }
func (h *Heap) Contains(x uint32) bool           { return int(x) < len(h.indices) && h.indices[x] >= 0 }
func (h *Heap) better(x, y uint32) bool          { return h.weights.At(x) > h.weights.At(y) }
func (h *Heap) up(i int) {
	x := h.data[i]
	for i > 0 {
		p := (i - 1) / 2
		if !h.better(x, h.data[p]) {
			break
		}
		h.data[i] = h.data[p]
		h.indices[h.data[i]] = i
		i = p
	}
	h.data[i] = x
	h.indices[x] = i
}
func (h *Heap) down(i int) {
	x := h.data[i]
	for 2*i+1 < len(h.data) {
		child := 2*i + 1
		if child+1 < len(h.data) && h.better(h.data[child+1], h.data[child]) {
			child++
		}
		if !h.better(h.data[child], x) {
			break
		}
		h.data[i] = h.data[child]
		h.indices[h.data[i]] = i
		i = child
	}
	h.data[i] = x
	h.indices[x] = i
}
func (h *Heap) Insert(x uint32) {
	for len(h.indices) <= int(x) {
		h.indices = append(h.indices, -1)
	}
	if h.Contains(x) {
		panic("duplicate heap entry")
	}
	h.indices[x] = len(h.data)
	h.data = append(h.data, x)
	h.up(len(h.data) - 1)
}
func (h *Heap) Update(x uint32) {
	if !h.Contains(x) {
		h.Insert(x)
		return
	}
	h.up(h.indices[x])
	h.down(h.indices[x])
}
func (h *Heap) Pop() uint32 {
	if len(h.data) == 0 {
		panic("pop empty heap")
	}
	x := h.data[0]
	last := h.data[len(h.data)-1]
	h.data = h.data[:len(h.data)-1]
	h.indices[x] = -1
	if len(h.data) > 0 {
		h.data[0] = last
		h.indices[last] = 0
		h.down(0)
	}
	return x
}
func (h *Heap) Clear() { h.data = nil; h.indices = nil }
func (h *Heap) Build(entries []uint32) {
	h.Clear()
	for _, x := range entries {
		h.Insert(x)
	}
}
func heap_free(h *heap_t)        { h.Clear(); h.weights = nil }
func heap_size(h *heap_t) uint32 { return uint32(h.Len()) }
func heap_in_heap(h *heap_t, x uint32) int32 {
	if h.Contains(x) {
		return 1
	}
	return 0
}
func heap_increase(h *heap_t, x uint32) {
	if !h.Contains(x) {
		panic("absent heap entry")
	}
	h.down(h.indices[x])
}
func heap_decrease(h *heap_t, x uint32) {
	if !h.Contains(x) {
		panic("absent heap entry")
	}
	h.up(h.indices[x])
}
func heap_insert(h *heap_t, x uint32)          { h.Insert(x) }
func heap_update(h *heap_t, x uint32)          { h.Update(x) }
func heap_build(h *heap_t, entries *vec.Uints) { h.Build(entries.Data()) }
func heap_clear(h *heap_t)                     { h.Clear() }
func heap_remove_min(h *heap_t) uint32         { return h.Pop() }
