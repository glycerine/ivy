package fxu

import (
	"fmt"
	"io"
)

type heapItem interface {
	comparable
	fxWeight() int32
	fxIndex() int
	fxSetIndex(int)
}
type priorityHeap[T heapItem] struct{ items []T }

func (h *priorityHeap[T]) swap(a, b int) {
	h.items[a], h.items[b] = h.items[b], h.items[a]
	h.items[a].fxSetIndex(a)
	h.items[b].fxSetIndex(b)
}
func (h *priorityHeap[T]) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.items[i].fxWeight() <= h.items[parent].fxWeight() {
			break
		}
		h.swap(i, parent)
		i = parent
	}
}
func (h *priorityHeap[T]) down(i int) {
	for 2*i+1 < len(h.items) {
		child := 2*i + 1
		if child+1 < len(h.items) && h.items[child+1].fxWeight() > h.items[child].fxWeight() {
			child++
		}
		if h.items[i].fxWeight() >= h.items[child].fxWeight() {
			break
		}
		h.swap(i, child)
		i = child
	}
}
func (h *priorityHeap[T]) insert(x T) {
	if i := x.fxIndex(); i >= 0 && i < len(h.items) && h.items[i] == x {
		panic("FXU duplicate heap item")
	}
	x.fxSetIndex(len(h.items))
	h.items = append(h.items, x)
	h.up(len(h.items) - 1)
}
func (h *priorityHeap[T]) update(x T) {
	i := x.fxIndex()
	if i < 0 || i >= len(h.items) || h.items[i] != x {
		panic("FXU missing heap item")
	}
	if i > 0 && x.fxWeight() > h.items[(i-1)/2].fxWeight() {
		h.up(i)
	} else {
		h.down(i)
	}
}
func (h *priorityHeap[T]) remove(x T) {
	i := x.fxIndex()
	if i < 0 || i >= len(h.items) || h.items[i] != x {
		return
	}
	last := len(h.items) - 1
	h.swap(i, last)
	var zero T
	h.items[last] = zero
	h.items = h.items[:last]
	x.fxSetIndex(-1)
	if i < len(h.items) {
		h.update(h.items[i])
	}
}
func (h *priorityHeap[T]) top() T {
	if len(h.items) == 0 {
		var zero T
		return zero
	}
	return h.items[0]
}
func (h *priorityHeap[T]) pop() T {
	x := h.top()
	if len(h.items) > 0 {
		h.remove(x)
	}
	return x
}
func (h *priorityHeap[T]) clear() {
	for _, x := range h.items {
		x.fxSetIndex(-1)
	}
	clear(h.items)
	h.items = h.items[:0]
}
func (h *priorityHeap[T]) check() {
	for i, x := range h.items {
		if x.fxIndex() != i {
			panic("FXU heap index mismatch")
		}
		if i > 0 && h.items[(i-1)/2].fxWeight() < x.fxWeight() {
			panic("FXU heap order mismatch")
		}
	}
}
func (h *priorityHeap[T]) print(w io.Writer) error {
	for i, x := range h.items {
		if _, err := fmt.Fprintf(w, "%d: %d\n", i, x.fxWeight()); err != nil {
			return err
		}
	}
	return nil
}
func (s *Fxu_Single) fxWeight() int32  { return s.Weight }
func (s *Fxu_Single) fxIndex() int     { return s.heapIndex }
func (s *Fxu_Single) fxSetIndex(i int) { s.heapIndex = i }
func (d *Fxu_Double) fxWeight() int32  { return d.Weight }
func (d *Fxu_Double) fxIndex() int     { return d.heapIndex }
func (d *Fxu_Double) fxSetIndex(i int) { d.heapIndex = i }
