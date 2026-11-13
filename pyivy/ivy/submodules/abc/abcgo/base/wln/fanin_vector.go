package wln

// Wln_Vec_t stores the first two fanins inline, as in wln.h. A third
// fanin moves the vector to a separately allocated, growing array.
type Wln_Vec_t struct {
	nCap, nSize int32
	inline      [2]int32
	heap        []int32
}

// appendFanin implements the storage transitions in Wln_ObjAddFanin.
func (v *Wln_Vec_t) appendFanin(id int32) {
	switch {
	case v.nSize < 2:
		v.inline[v.nSize] = id
		v.nSize++
	case v.nSize == 2:
		v.heap = make([]int32, 4)
		copy(v.heap, v.inline[:])
		v.heap[2] = id
		v.nSize = 3
		v.nCap = 4
	default:
		if v.nSize == v.nCap {
			next := make([]int32, 2*v.nSize)
			copy(next, v.heap)
			v.heap = next
			v.nCap = 2 * v.nSize
		}
		v.heap[v.nSize] = id
		v.nSize++
	}
}

func (v *Wln_Vec_t) fanins() []int32 {
	if v.nSize > 2 {
		return v.heap
	}
	return v.inline[:]
}
