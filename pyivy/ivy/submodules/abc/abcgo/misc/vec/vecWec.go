package vec

import (
	"fmt"
	"sort"
	"unsafe"
)

type Vec_Wec_t struct {
	nCap, nSize int
	pArray      []Vec_Int_t
}

func Vec_WecAlloc(n int) *Vec_Wec_t {
	if n > 0 && n < 8 {
		n = 8
	}
	return Vec_WecAllocExact(n)
}
func Vec_WecAllocExact(n int) *Vec_Wec_t { return &Vec_Wec_t{nCap: n, pArray: make([]Vec_Int_t, n)} }
func Vec_WecStart(n int) *Vec_Wec_t      { p := Vec_WecAlloc(n); p.nSize = n; return p }
func Vec_WecGrow(p *Vec_Wec_t, n int) {
	if n > p.nCap {
		p.pArray = append(p.pArray, make([]Vec_Int_t, n-p.nCap)...)
		p.nCap = n
	}
}
func Vec_WecInit(p *Vec_Wec_t, n int)                { Vec_WecGrow(p, n); p.nSize = n }
func Vec_WecEntry(p *Vec_Wec_t, i int) *Vec_Int_t    { return &p.pArray[:p.nSize][i] }
func Vec_WecEntryLast(p *Vec_Wec_t) *Vec_Int_t       { return Vec_WecEntry(p, p.nSize-1) }
func Vec_WecEntryEntry(p *Vec_Wec_t, i, k int) int32 { return Vec_IntEntry(Vec_WecEntry(p, i), k) }
func Vec_WecArray(p *Vec_Wec_t) []Vec_Int_t          { return p.pArray[:p.nSize] }
func Vec_WecLevelId(p *Vec_Wec_t, level *Vec_Int_t) int {
	for i := range p.nSize {
		if &p.pArray[i] == level {
			return i
		}
	}
	panic("level not in vector")
}
func Vec_WecCap(p *Vec_Wec_t) int              { return p.nCap }
func Vec_WecSize(p *Vec_Wec_t) int             { return p.nSize }
func Vec_WecLevelSize(p *Vec_Wec_t, i int) int { return Vec_IntSize(Vec_WecEntry(p, i)) }
func Vec_WecSizeSize(p *Vec_Wec_t) int {
	n := 0
	for i := range p.nSize {
		n += p.pArray[i].nSize
	}
	return n
}
func Vec_WecSizeUsed(p *Vec_Wec_t) int { return Vec_WecSizeUsedLimits(p, 0, p.nSize) }
func Vec_WecSizeUsedLimits(p *Vec_Wec_t, start, stop int) int {
	n := 0
	for _, v := range p.pArray[start:stop] {
		if v.nSize > 0 {
			n++
		}
	}
	return n
}
func Vec_WecShrink(p *Vec_Wec_t, n int) {
	if n < 0 || n > p.nSize {
		panic("level shrink")
	}
	for i := n; i < p.nSize; i++ {
		Vec_IntClear(&p.pArray[i])
	}
	p.nSize = n
}
func Vec_WecClear(p *Vec_Wec_t) { Vec_WecShrink(p, 0) }
func Vec_WecPush(p *Vec_Wec_t, level int, x int32) {
	if level >= p.nSize {
		Vec_WecGrow(p, max(2*p.nSize, level+1))
		p.nSize = level + 1
	}
	Vec_IntPush(Vec_WecEntry(p, level), x)
}
func Vec_WecPushLevel(p *Vec_Wec_t) *Vec_Int_t {
	if p.nSize == p.nCap {
		Vec_WecGrow(p, max(16, 2*p.nCap))
	}
	p.nSize++
	return Vec_WecEntryLast(p)
}
func Vec_WecInsertLevel(p *Vec_Wec_t, i int) *Vec_Int_t {
	if i < 0 || i > p.nSize {
		panic("level insert")
	}
	Vec_WecPushLevel(p)
	copy(p.pArray[i+1:p.nSize], p.pArray[i:p.nSize-1])
	p.pArray[i] = Vec_Int_t{}
	return &p.pArray[i]
}
func Vec_WecMemory(p *Vec_Wec_t) float64 {
	if p == nil {
		return 0
	}
	n := p.nCap * int(unsafe.Sizeof(Vec_Int_t{}))
	for i := range p.nSize {
		n += 4 * p.pArray[i].nCap
	}
	return float64(n)
}
func Vec_WecZero(p *Vec_Wec_t)  { *p = Vec_Wec_t{} }
func Vec_WecErase(p *Vec_Wec_t) { Vec_WecZero(p) }
func Vec_WecFree(p *Vec_Wec_t) {
	if p != nil {
		Vec_WecErase(p)
	}
}
func Vec_WecFreeP(p **Vec_Wec_t) { Vec_WecFree(*p); *p = nil }
func Vec_WecPushUnique(p *Vec_Wec_t, level int, x int32) {
	if level >= p.nSize {
		Vec_WecPush(p, level, x)
	} else {
		Vec_IntPushUnique(Vec_WecEntry(p, level), x)
	}
}
func Vec_WecDup(p *Vec_Wec_t) *Vec_Wec_t {
	q := Vec_WecStart(p.nSize)
	for i := range p.nSize {
		q.pArray[i] = *Vec_IntDup(&p.pArray[i])
	}
	return q
}
func Vec_WecSortCompare1(p, q *Vec_Int_t) int {
	if p.nSize < q.nSize {
		return -1
	}
	if p.nSize > q.nSize {
		return 1
	}
	return 0
}
func Vec_WecSortCompare2(p, q *Vec_Int_t) int { return -Vec_WecSortCompare1(p, q) }
func Vec_WecSortCompare3(p, q *Vec_Int_t) int {
	a, b := Vec_IntEntry(p, 0), Vec_IntEntry(q, 0)
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func Vec_WecSortCompare4(p, q *Vec_Int_t) int { return -Vec_WecSortCompare3(p, q) }
func Vec_WecSortCompare5(p, q *Vec_Int_t) int {
	a, b := Vec_IntEntryLast(p), Vec_IntEntryLast(q)
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func Vec_WecSortCompare6(p, q *Vec_Int_t) int { return -Vec_WecSortCompare5(p, q) }
func wecSort(p *Vec_Wec_t, reverse int, compare func(*Vec_Int_t, *Vec_Int_t) int) {
	sort.SliceStable(Vec_WecArray(p), func(i, j int) bool {
		v := compare(&p.pArray[i], &p.pArray[j])
		if reverse != 0 {
			return v > 0
		}
		return v < 0
	})
}
func Vec_WecSort(p *Vec_Wec_t, reverse int)           { wecSort(p, reverse, Vec_WecSortCompare1) }
func Vec_WecSortByFirstInt(p *Vec_Wec_t, reverse int) { wecSort(p, reverse, Vec_WecSortCompare3) }
func Vec_WecSortByLastInt(p *Vec_Wec_t, reverse int)  { wecSort(p, reverse, Vec_WecSortCompare5) }
func Vec_WecPrint(p *Vec_Wec_t, skipSingles int) {
	for i := range p.nSize {
		v := &p.pArray[i]
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
func Vec_WecPrintLits(p *Vec_Wec_t) {
	for i := range p.nSize {
		v := &p.pArray[i]
		fmt.Printf(" %4d : %2d  {", i, v.nSize)
		for _, lit := range Vec_IntArray(v) {
			sign := '+'
			if lit&1 != 0 {
				sign = '-'
			}
			fmt.Printf(" %c%d", sign, lit>>1)
		}
		fmt.Println(" }")
	}
}
func Vec_WecCreateClasses(mapping *Vec_Int_t) *Vec_Wec_t {
	q := Vec_WecStart(int(Vec_IntFindMax(mapping)) + 1)
	for i, x := range Vec_IntArray(mapping) {
		Vec_WecPush(q, int(x), int32(i))
	}
	return q
}
func Vec_WecCountNonTrivial(p *Vec_Wec_t, used *int) int {
	*used = 0
	n := 0
	for _, v := range Vec_WecArray(p) {
		if v.nSize >= 2 {
			n++
			*used += v.nSize
		}
	}
	return n
}
func Vec_WecCollectFirsts(p *Vec_Wec_t) *Vec_Int_t {
	q := Vec_IntAlloc(p.nSize)
	for i := range p.nSize {
		v := &p.pArray[i]
		if v.nSize > 0 {
			Vec_IntPush(q, Vec_IntEntry(v, 0))
		}
	}
	return q
}
func Vec_WecConvertToVecPtr(p *Vec_Wec_t) *Vec_Ptr_t {
	q := Vec_PtrAlloc(p.nSize)
	for i := range p.nSize {
		Vec_PtrPush(q, Vec_IntDup(&p.pArray[i]))
	}
	return q
}

// Marks live separately from capacity: Go slices must never use tagged lengths.
func Vec_WecIntHasMark(p *Vec_Int_t) int {
	if p.marked {
		return 1
	}
	return 0
}
func Vec_WecIntSetMark(p *Vec_Int_t) { p.marked = true }
func Vec_WecIntXorMark(p *Vec_Int_t) { p.marked = !p.marked }
func Vec_WecMarkLevels(p *Vec_Wec_t, levels *Vec_Int_t) {
	for _, i := range Vec_IntArray(levels) {
		v := Vec_WecEntry(p, int(i))
		if v.marked {
			panic("level already marked")
		}
		v.marked = true
	}
}
func Vec_WecUnmarkLevels(p *Vec_Wec_t, levels *Vec_Int_t) {
	for _, i := range Vec_IntArray(levels) {
		v := Vec_WecEntry(p, int(i))
		if !v.marked {
			panic("level not marked")
		}
		v.marked = false
	}
}
func Vec_WecRemoveEmpty(p *Vec_Wec_t) {
	k := 0
	for i := range p.nSize {
		if p.pArray[i].nSize > 0 {
			p.pArray[k] = p.pArray[i]
			k++
		}
	}
	clear(p.pArray[k:p.nSize])
	p.nSize = k
}
