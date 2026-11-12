package vec

import (
	"fmt"
	"sort"
)

// Vec_Vec_t stores separately allocated pointer or integer levels. Each vector
// is homogeneous; Go type assertions replace C's interchangeable layouts.
type Vec_Vec_t = Vec_Ptr_t

func Vec_VecAlloc(n int) *Vec_Vec_t { return Vec_PtrAlloc(n) }
func Vec_VecStart(n int) *Vec_Vec_t { p := Vec_VecAlloc(n); Vec_VecExpand(p, n-1); return p }
func Vec_VecExpand(p *Vec_Vec_t, level int) {
	if level < p.nSize {
		return
	}
	Vec_PtrGrow(p, level+1)
	for i := p.nSize; i <= level; i++ {
		p.pArray[i] = Vec_PtrAlloc(0)
	}
	p.nSize = level + 1
}
func Vec_VecExpandInt(p *Vec_Vec_t, level int) {
	if level < p.nSize {
		return
	}
	Vec_PtrGrow(p, level+1)
	for i := p.nSize; i <= level; i++ {
		p.pArray[i] = Vec_IntAlloc(0)
	}
	p.nSize = level + 1
}
func Vec_VecSize(p *Vec_Vec_t) int { return p.nSize }
func Vec_VecCap(p *Vec_Vec_t) int  { return p.nCap }
func Vec_VecLevelSize(p *Vec_Vec_t, i int) int {
	switch v := Vec_PtrEntry(p, i).(type) {
	case *Vec_Int_t:
		return v.nSize
	case *Vec_Ptr_t:
		return v.nSize
	}
	panic("invalid level type")
}
func Vec_VecEntry(p *Vec_Vec_t, i int) *Vec_Ptr_t    { return Vec_PtrEntry(p, i).(*Vec_Ptr_t) }
func Vec_VecEntryInt(p *Vec_Vec_t, i int) *Vec_Int_t { return Vec_PtrEntry(p, i).(*Vec_Int_t) }
func Vec_VecMemory(p *Vec_Vec_t) float64 {
	if p == nil {
		return 0
	}
	n := Vec_PtrMemory(p)
	for _, x := range Vec_PtrArray(p) {
		if x != nil {
			n += Vec_PtrMemory(x.(*Vec_Ptr_t))
		}
	}
	return n
}
func Vec_VecMemoryInt(p *Vec_Vec_t) float64 {
	if p == nil {
		return 0
	}
	n := Vec_PtrMemory(p)
	for _, x := range Vec_PtrArray(p) {
		if x != nil {
			n += Vec_IntMemory(x.(*Vec_Int_t))
		}
	}
	return n
}
func Vec_VecEntryEntry(p *Vec_Vec_t, i, k int) any { return Vec_PtrEntry(Vec_VecEntry(p, i), k) }
func Vec_VecEntryEntryInt(p *Vec_Vec_t, i, k int) int32 {
	return Vec_IntEntry(Vec_VecEntryInt(p, i), k)
}
func Vec_VecFree(p *Vec_Vec_t) {
	if p == nil {
		return
	}
	Vec_VecErase(p)
}
func Vec_VecErase(p *Vec_Vec_t) {
	for _, x := range Vec_PtrArray(p) {
		switch v := x.(type) {
		case *Vec_Ptr_t:
			Vec_PtrFree(v)
		case *Vec_Int_t:
			Vec_IntFree(v)
		}
	}
	Vec_PtrErase(p)
}
func Vec_VecFreeP(p **Vec_Vec_t) { Vec_VecFree(*p); *p = nil }
func Vec_VecDup(p *Vec_Vec_t) *Vec_Vec_t {
	q := Vec_VecAlloc(p.nSize)
	for _, x := range Vec_PtrArray(p) {
		Vec_PtrPush(q, Vec_PtrDup(x.(*Vec_Ptr_t)))
	}
	return q
}
func Vec_VecDupInt(p *Vec_Vec_t) *Vec_Vec_t {
	q := Vec_VecAlloc(p.nSize)
	for _, x := range Vec_PtrArray(p) {
		Vec_PtrPush(q, Vec_IntDup(x.(*Vec_Int_t)))
	}
	return q
}
func Vec_VecSizeSize(p *Vec_Vec_t) int {
	n := 0
	for i := range p.nSize {
		n += Vec_VecLevelSize(p, i)
	}
	return n
}
func Vec_VecClear(p *Vec_Vec_t) {
	for _, x := range Vec_PtrArray(p) {
		switch v := x.(type) {
		case *Vec_Ptr_t:
			Vec_PtrClear(v)
		case *Vec_Int_t:
			Vec_IntClear(v)
		}
	}
}
func Vec_VecPush(p *Vec_Vec_t, level int, x any) {
	Vec_VecExpand(p, level)
	Vec_PtrPush(Vec_VecEntry(p, level), x)
}
func Vec_VecPushInt(p *Vec_Vec_t, level int, x int32) {
	Vec_VecExpandInt(p, level)
	Vec_IntPush(Vec_VecEntryInt(p, level), x)
}
func Vec_VecPushUnique(p *Vec_Vec_t, level int, x any) {
	Vec_VecExpand(p, level)
	Vec_PtrPushUnique(Vec_VecEntry(p, level), x)
}
func Vec_VecPushUniqueInt(p *Vec_Vec_t, level int, x int32) {
	Vec_VecExpandInt(p, level)
	Vec_IntPushUnique(Vec_VecEntryInt(p, level), x)
}
func Vec_VecSortCompare1(a, b **Vec_Ptr_t) int { return (*a).nSize - (*b).nSize }
func Vec_VecSortCompare2(a, b **Vec_Ptr_t) int { return -Vec_VecSortCompare1(a, b) }
func Vec_VecSort(p *Vec_Vec_t, reverse int) {
	sort.SliceStable(Vec_PtrArray(p), func(i, j int) bool {
		a, b := Vec_VecLevelSize(p, i), Vec_VecLevelSize(p, j)
		if reverse != 0 {
			return a > b
		}
		return a < b
	})
}
func Vec_VecSortCompare3(a, b **Vec_Int_t) int {
	av, bv := Vec_IntEntry(*a, 0), Vec_IntEntry(*b, 0)
	if av < bv {
		return -1
	}
	if av > bv {
		return 1
	}
	return 0
}
func Vec_VecSortCompare4(a, b **Vec_Int_t) int { return -Vec_VecSortCompare3(a, b) }
func Vec_VecSortByFirstInt(p *Vec_Vec_t, reverse int) {
	sort.SliceStable(Vec_PtrArray(p), func(i, j int) bool {
		a, b := Vec_VecEntryEntryInt(p, i, 0), Vec_VecEntryEntryInt(p, j, 0)
		if reverse != 0 {
			return a > b
		}
		return a < b
	})
}
func Vec_VecPrintInt(p *Vec_Vec_t, skipSingles int) {
	for i := range p.nSize {
		v := Vec_VecEntryInt(p, i)
		if skipSingles != 0 && v.nSize == 1 {
			continue
		}
		fmt.Printf(" %4d : {", i)
		for _, x := range Vec_IntArray(v) {
			fmt.Printf(" %d", x)
		}
		fmt.Println(" }")
	}
}
