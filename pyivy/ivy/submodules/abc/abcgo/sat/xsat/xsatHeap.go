// Go port of the indexed integer-activity heap in xsatHeap.h.
package xsat

type xSAT_Heap_t struct {
	weights     *[]uint32
	data, index []int
}

func xSAT_HeapAlloc(weights *[]uint32) *xSAT_Heap_t { return &xSAT_Heap_t{weights: weights} }
func xSAT_HeapFree(h *xSAT_Heap_t)                  { h.data = nil; h.index = nil; h.weights = nil }
func xSAT_HeapSize(h *xSAT_Heap_t) int              { return len(h.data) }
func xSAT_HeapInHeap(h *xSAT_Heap_t, v int) int {
	return boolInt(v >= 0 && v < len(h.index) && h.index[v] >= 0)
}
func Left(i int) int                        { return 2*i + 1 }
func Right(i int) int                       { return 2*i + 2 }
func Parent(i int) int                      { return (i - 1) >> 1 }
func Compare(h *xSAT_Heap_t, a, b int) bool { return (*h.weights)[a] > (*h.weights)[b] }
func xSAT_HeapPercolateUp(h *xSAT_Heap_t, i int) {
	v := h.data[i]
	for i > 0 {
		p := (i - 1) / 2
		if !Compare(h, v, h.data[p]) {
			break
		}
		h.data[i] = h.data[p]
		h.index[h.data[i]] = i
		i = p
	}
	h.data[i] = v
	h.index[v] = i
}
func xSAT_HeapPercolateDown(h *xSAT_Heap_t, i int) {
	v := h.data[i]
	for 2*i+1 < len(h.data) {
		child := 2*i + 1
		if child+1 < len(h.data) && Compare(h, h.data[child+1], h.data[child]) {
			child++
		}
		if !Compare(h, h.data[child], v) {
			break
		}
		h.data[i] = h.data[child]
		h.index[h.data[i]] = i
		i = child
	}
	h.data[i] = v
	h.index[v] = i
}
func xSAT_HeapInsert(h *xSAT_Heap_t, v int) {
	if v < 0 || v >= len(*h.weights) {
		panic("heap weight index out of range")
	}
	for len(h.index) <= v {
		h.index = append(h.index, -1)
	}
	if h.index[v] >= 0 {
		panic("duplicate heap entry")
	}
	h.index[v] = len(h.data)
	h.data = append(h.data, v)
	xSAT_HeapPercolateUp(h, len(h.data)-1)
}
func xSAT_HeapIncrease(h *xSAT_Heap_t, v int) {
	if xSAT_HeapInHeap(h, v) == 0 {
		panic("absent heap entry")
	}
	xSAT_HeapPercolateDown(h, h.index[v])
}
func xSAT_HeapDecrease(h *xSAT_Heap_t, v int) {
	if xSAT_HeapInHeap(h, v) == 0 {
		panic("absent heap entry")
	}
	xSAT_HeapPercolateUp(h, h.index[v])
}
func xSAT_HeapUpdate(h *xSAT_Heap_t, v int) {
	if xSAT_HeapInHeap(h, v) == 0 {
		xSAT_HeapInsert(h, v)
	} else {
		xSAT_HeapPercolateUp(h, h.index[v])
		xSAT_HeapPercolateDown(h, h.index[v])
	}
}
func xSAT_HeapClear(h *xSAT_Heap_t) {
	for _, v := range h.data {
		h.index[v] = -1
	}
	h.data = h.data[:0]
}
func xSAT_HeapBuild(h *xSAT_Heap_t, vars []int) {
	xSAT_HeapClear(h)
	for _, v := range vars {
		xSAT_HeapInsert(h, v)
	}
}
func xSAT_HeapRemoveMin(h *xSAT_Heap_t) int {
	if len(h.data) == 0 {
		panic("pop empty heap")
	}
	v := h.data[0]
	last := h.data[len(h.data)-1]
	h.data = h.data[:len(h.data)-1]
	h.index[v] = -1
	if len(h.data) > 0 {
		h.data[0] = last
		h.index[last] = 0
		xSAT_HeapPercolateDown(h, 0)
	}
	return v
}
