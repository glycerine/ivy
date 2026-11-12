package mapper

import "sort"

// Map_NodeVec_t is the mapper's compact, growable node list. The translated
// algorithms still read nSize directly, so it is kept in sync with pArray.
func Map_NodeVecAlloc(capacity int32) *Map_NodeVec_t {
	if capacity > 0 && capacity < 16 {
		capacity = 16
	}
	if capacity < 0 {
		panic("mapper: negative node vector capacity")
	}
	return &Map_NodeVecStruct_t_{pArray: make([]*Map_Node_t, 0, capacity), nCap: capacity}
}

func Map_NodeVecFree(v *Map_NodeVec_t) {
	if v == nil {
		return
	}
	v.pArray = nil
	v.nSize, v.nCap = 0, 0
}

func Map_NodeVecDup(v *Map_NodeVec_t) *Map_NodeVec_t {
	if v == nil {
		return nil
	}
	out := Map_NodeVecAlloc(v.nSize)
	out.pArray = append(out.pArray, v.pArray[:v.nSize]...)
	out.nSize = int32(len(out.pArray))
	return out
}

func Map_NodeVecReadArray(v *Map_NodeVec_t) []*Map_Node_t {
	if v == nil {
		return nil
	}
	return v.pArray[:v.nSize]
}

func Map_NodeVecReadSize(v *Map_NodeVec_t) int32 {
	if v == nil {
		return 0
	}
	return v.nSize
}

func Map_NodeVecGrow(v *Map_NodeVec_t, minimum int32) {
	if v == nil || minimum <= v.nCap {
		return
	}
	if minimum < 0 {
		panic("mapper: negative node vector capacity")
	}
	items := make([]*Map_Node_t, v.nSize, minimum)
	copy(items, v.pArray[:v.nSize])
	v.pArray, v.nCap = items, minimum
}

func Map_NodeVecShrink(v *Map_NodeVec_t, size int32) {
	if v == nil || size < 0 || size > v.nSize {
		panic("mapper: invalid node vector shrink")
	}
	clear(v.pArray[size:v.nSize])
	v.pArray = v.pArray[:size]
	v.nSize = size
}

func Map_NodeVecClear(v *Map_NodeVec_t) {
	if v == nil {
		return
	}
	clear(v.pArray[:v.nSize])
	v.pArray = v.pArray[:0]
	v.nSize = 0
}

func Map_NodeVecPush(v *Map_NodeVec_t, node *Map_Node_t) {
	if v == nil {
		panic("mapper: push to nil node vector")
	}
	if v.nSize == v.nCap {
		capacity := int32(16)
		if v.nCap >= 16 {
			capacity = 2 * v.nCap
		}
		Map_NodeVecGrow(v, capacity)
	}
	v.pArray = append(v.pArray[:v.nSize], node)
	v.nSize++
	if int32(cap(v.pArray)) > v.nCap {
		v.nCap = int32(cap(v.pArray))
	}
}

func Map_NodeVecPushUnique(v *Map_NodeVec_t, node *Map_Node_t) int32 {
	for _, existing := range Map_NodeVecReadArray(v) {
		if existing == node {
			return 1
		}
	}
	Map_NodeVecPush(v, node)
	return 0
}

func Map_NodeVecPop(v *Map_NodeVec_t) *Map_Node_t {
	if v == nil || v.nSize == 0 {
		panic("mapper: pop from empty node vector")
	}
	last := v.nSize - 1
	node := v.pArray[last]
	v.pArray[last] = nil
	v.pArray = v.pArray[:last]
	v.nSize = last
	return node
}

func Map_NodeVecRemove(v *Map_NodeVec_t, node *Map_Node_t) {
	for i, candidate := range Map_NodeVecReadArray(v) {
		if candidate == node {
			copy(v.pArray[i:], v.pArray[i+1:v.nSize])
			v.nSize--
			v.pArray[v.nSize] = nil
			v.pArray = v.pArray[:v.nSize]
			return
		}
	}
	panic("mapper: removing node not present in vector")
}

func Map_NodeVecWriteEntry(v *Map_NodeVec_t, index int32, node *Map_Node_t) {
	if v == nil || index < 0 || index >= v.nSize {
		panic("mapper: node vector index out of range")
	}
	v.pArray[index] = node
}

func Map_NodeVecReadEntry(v *Map_NodeVec_t, index int32) *Map_Node_t {
	if v == nil || index < 0 || index >= v.nSize {
		panic("mapper: node vector index out of range")
	}
	return v.pArray[index]
}

func Map_NodeVecSortByLevel(v *Map_NodeVec_t) {
	if v == nil {
		return
	}
	sort.SliceStable(v.pArray[:v.nSize], func(i, j int) bool {
		a, b := Map_Regular(v.pArray[i]), Map_Regular(v.pArray[j])
		if a.Level != b.Level {
			return a.Level < b.Level
		}
		return a.Num < b.Num
	})
}
