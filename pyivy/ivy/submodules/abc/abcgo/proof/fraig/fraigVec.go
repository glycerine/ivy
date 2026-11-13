package fraig

// Node-vector operations adapted from fraigVec.c by Alan Mishchenko.
import (
	"sort"
	"unsafe"
)

func Fraig_NodeVecAlloc(capacity int32) *Fraig_NodeVec_t {
	if capacity < 0 {
		panic("fraig: negative vector capacity")
	}
	if capacity > 0 && capacity < 8 {
		capacity = 8
	}
	return &Fraig_NodeVec_t{pArray: make([]*Fraig_Node_t, 0, capacity)}
}
func Fraig_NodeVecFree(p *Fraig_NodeVec_t) {
	if p != nil {
		p.pArray = nil
	}
}
func Fraig_NodeVecDup(p *Fraig_NodeVec_t) *Fraig_NodeVec_t {
	q := &Fraig_NodeVec_t{pArray: make([]*Fraig_Node_t, len(p.pArray), cap(p.pArray))}
	copy(q.pArray, p.pArray)
	return q
}
func Fraig_NodeVecReadArray(p *Fraig_NodeVec_t) []*Fraig_Node_t { return p.pArray }
func Fraig_NodeVecReadSize(p *Fraig_NodeVec_t) int32            { return int32(len(p.pArray)) }
func Fraig_NodeVecGrow(p *Fraig_NodeVec_t, capacity int32) {
	if capacity < 0 {
		panic("fraig: negative vector capacity")
	}
	if int(capacity) <= cap(p.pArray) {
		return
	}
	entries := make([]*Fraig_Node_t, len(p.pArray), capacity)
	copy(entries, p.pArray)
	p.pArray = entries
}
func Fraig_NodeVecShrink(p *Fraig_NodeVec_t, size int32) {
	if size < 0 || int(size) > len(p.pArray) {
		panic("fraig: invalid vector size")
	}
	p.pArray = p.pArray[:size]
}
func Fraig_NodeVecClear(p *Fraig_NodeVec_t)                    { Fraig_NodeVecShrink(p, 0) }
func Fraig_NodeVecPush(p *Fraig_NodeVec_t, node *Fraig_Node_t) { p.pArray = append(p.pArray, node) }
func Fraig_NodeVecPushUnique(p *Fraig_NodeVec_t, node *Fraig_Node_t) int32 {
	for _, old := range p.pArray {
		if old == node {
			return 1
		}
	}
	Fraig_NodeVecPush(p, node)
	return 0
}

// C orders the tagged pointer values, including the complement bit.
func fraigEdgeLess(a, b *Fraig_Node_t) bool {
	x, y := uintptr(unsafe.Pointer(Fraig_Regular(a))), uintptr(unsafe.Pointer(Fraig_Regular(b)))
	if Fraig_IsComplement(a) {
		x |= 1
	}
	if Fraig_IsComplement(b) {
		y |= 1
	}
	return x < y
}
func fraigInsertOrdered(p *Fraig_NodeVec_t, node *Fraig_Node_t, less func(*Fraig_Node_t, *Fraig_Node_t) bool) {
	Fraig_NodeVecPush(p, node)
	for i := len(p.pArray) - 1; i > 0 && less(p.pArray[i], p.pArray[i-1]); i-- {
		p.pArray[i], p.pArray[i-1] = p.pArray[i-1], p.pArray[i]
	}
}
func Fraig_NodeVecPushOrder(p *Fraig_NodeVec_t, node *Fraig_Node_t) {
	fraigInsertOrdered(p, node, fraigEdgeLess)
}
func Fraig_NodeVecPushUniqueOrder(p *Fraig_NodeVec_t, node *Fraig_Node_t) int32 {
	for _, old := range p.pArray {
		if old == node {
			return 1
		}
	}
	Fraig_NodeVecPushOrder(p, node)
	return 0
}
func Fraig_NodeVecPushOrderByLevel(p *Fraig_NodeVec_t, node *Fraig_Node_t) {
	fraigInsertOrdered(p, node, func(a, b *Fraig_Node_t) bool { return Fraig_Regular(a).Level > Fraig_Regular(b).Level })
}
func Fraig_NodeVecPushUniqueOrderByLevel(p *Fraig_NodeVec_t, node *Fraig_Node_t) int32 {
	for _, old := range p.pArray {
		if old == node {
			return 1
		}
	}
	Fraig_NodeVecPushOrderByLevel(p, node)
	return 0
}
func Fraig_NodeVecPop(p *Fraig_NodeVec_t) *Fraig_Node_t {
	i := len(p.pArray) - 1
	if i < 0 {
		panic("fraig: pop from empty vector")
	}
	node := p.pArray[i]
	p.pArray = p.pArray[:i]
	return node
}
func Fraig_NodeVecRemove(p *Fraig_NodeVec_t, node *Fraig_Node_t) {
	for i, old := range p.pArray {
		if old == node {
			copy(p.pArray[i:], p.pArray[i+1:])
			Fraig_NodeVecShrink(p, int32(len(p.pArray)-1))
			return
		}
	}
	panic("fraig: node not found in vector")
}
func Fraig_NodeVecWriteEntry(p *Fraig_NodeVec_t, i int32, node *Fraig_Node_t) { p.pArray[i] = node }
func Fraig_NodeVecReadEntry(p *Fraig_NodeVec_t, i int32) *Fraig_Node_t        { return p.pArray[i] }
func fraigCompare(a, b int32) int32 {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func Fraig_NodeVecCompareLevelsIncreasing(a, b **Fraig_Node_t) int32 {
	return fraigCompare(Fraig_Regular(*a).Level, Fraig_Regular(*b).Level)
}
func Fraig_NodeVecCompareLevelsDecreasing(a, b **Fraig_Node_t) int32 {
	return -Fraig_NodeVecCompareLevelsIncreasing(a, b)
}
func Fraig_NodeVecCompareNumbers(a, b **Fraig_Node_t) int32 {
	return fraigCompare(Fraig_Regular(*a).Num, Fraig_Regular(*b).Num)
}
func Fraig_NodeVecCompareRefCounts(a, b **Fraig_Node_t) int32 {
	x, y := Fraig_Regular(*a), Fraig_Regular(*b)
	if c := fraigCompare(x.nRefs, y.nRefs); c != 0 {
		return c
	}
	return fraigCompare(x.Level, y.Level)
}
func Fraig_NodeVecSortByLevel(p *Fraig_NodeVec_t, increasing int32) {
	sort.Slice(p.pArray, func(i, j int) bool {
		a, b := Fraig_Regular(p.pArray[i]).Level, Fraig_Regular(p.pArray[j]).Level
		if increasing != 0 {
			return a < b
		}
		return a > b
	})
}
func Fraig_NodeVecSortByNumber(p *Fraig_NodeVec_t) {
	sort.Slice(p.pArray, func(i, j int) bool { return Fraig_Regular(p.pArray[i]).Num < Fraig_Regular(p.pArray[j]).Num })
}
func Fraig_NodeVecSortByRefCount(p *Fraig_NodeVec_t) {
	sort.Slice(p.pArray, func(i, j int) bool {
		a, b := Fraig_Regular(p.pArray[i]), Fraig_Regular(p.pArray[j])
		if a.nRefs != b.nRefs {
			return a.nRefs < b.nRefs
		}
		return a.Level < b.Level
	})
}
