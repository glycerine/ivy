package vec

import (
	"fmt"
	"sort"
	"unsafe"
)

type Vec_Wec_t struct{ data []Vec_Int_t }

func Vec_WecAlloc(n int) *Vec_Wec_t {
	if n > 0 && n < 8 {
		n = 8
	}
	return Vec_WecAllocExact(n)
}
func Vec_WecAllocExact(n int) *Vec_Wec_t { return &Vec_Wec_t{data: make([]Vec_Int_t, 0, n)} }
func Vec_WecStart(n int) *Vec_Wec_t      { p := Vec_WecAlloc(n); p.resize(n); return p }
func Vec_WecGrow(p *Vec_Wec_t, n int) {
	if n > cap(p.data) {
		data := make([]Vec_Int_t, len(p.data), n)
		copy(data, p.data)
		p.data = data
	}
}
func Vec_WecInit(p *Vec_Wec_t, n int)                { Vec_WecGrow(p, n); p.resize(n) }
func Vec_WecEntry(p *Vec_Wec_t, i int) *Vec_Int_t    { return &p.data[i] }
func Vec_WecEntryLast(p *Vec_Wec_t) *Vec_Int_t       { return Vec_WecEntry(p, len(p.data)-1) }
func Vec_WecEntryEntry(p *Vec_Wec_t, i, k int) int32 { return Vec_IntEntry(Vec_WecEntry(p, i), k) }
func Vec_WecArray(p *Vec_Wec_t) []Vec_Int_t          { return p.data }
func Vec_WecLevelId(p *Vec_Wec_t, level *Vec_Int_t) int {
	for i := range len(p.data) {
		if &p.storage()[i] == level {
			return i
		}
	}
	panic("level not in vector")
}
func Vec_WecCap(p *Vec_Wec_t) int              { return cap(p.data) }
func Vec_WecSize(p *Vec_Wec_t) int             { return len(p.data) }
func Vec_WecLevelSize(p *Vec_Wec_t, i int) int { return Vec_IntSize(Vec_WecEntry(p, i)) }
func Vec_WecSizeSize(p *Vec_Wec_t) int {
	n := 0
	for i := range len(p.data) {
		n += len(p.storage()[i].data)
	}
	return n
}
func Vec_WecSizeUsed(p *Vec_Wec_t) int { return Vec_WecSizeUsedLimits(p, 0, len(p.data)) }
func Vec_WecSizeUsedLimits(p *Vec_Wec_t, start, stop int) int {
	n := 0
	for _, v := range p.storage()[start:stop] {
		if len(v.data) > 0 {
			n++
		}
	}
	return n
}
func Vec_WecShrink(p *Vec_Wec_t, n int) {
	if n < 0 || n > len(p.data) {
		panic("level shrink")
	}
	for i := n; i < len(p.data); i++ {
		Vec_IntClear(&p.storage()[i])
	}
	p.resize(n)
}
func Vec_WecClear(p *Vec_Wec_t) { Vec_WecShrink(p, 0) }
func Vec_WecPush(p *Vec_Wec_t, level int, x int32) {
	if level >= len(p.data) {
		Vec_WecGrow(p, max(2*len(p.data), level+1))
		p.resize(level + 1)
	}
	Vec_IntPush(Vec_WecEntry(p, level), x)
}
func Vec_WecPushLevel(p *Vec_Wec_t) *Vec_Int_t {
	if len(p.data) == cap(p.data) {
		Vec_WecGrow(p, max(16, 2*cap(p.data)))
	}
	p.resize(len(p.data) + 1)
	return Vec_WecEntryLast(p)
}
func Vec_WecInsertLevel(p *Vec_Wec_t, i int) *Vec_Int_t {
	if i < 0 || i > len(p.data) {
		panic("level insert")
	}
	Vec_WecPushLevel(p)
	copy(p.storage()[i+1:len(p.data)], p.storage()[i:len(p.data)-1])
	p.storage()[i] = Vec_Int_t{}
	return &p.storage()[i]
}
func Vec_WecMemory(p *Vec_Wec_t) float64 {
	if p == nil {
		return 0
	}
	n := cap(p.data) * int(unsafe.Sizeof(Vec_Int_t{}))
	for i := range len(p.data) {
		n += 4 * cap(p.storage()[i].data)
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
	if level >= len(p.data) {
		Vec_WecPush(p, level, x)
	} else {
		Vec_IntPushUnique(Vec_WecEntry(p, level), x)
	}
}
func Vec_WecDup(p *Vec_Wec_t) *Vec_Wec_t {
	q := Vec_WecStart(len(p.data))
	for i := range len(p.data) {
		q.storage()[i] = *Vec_IntDup(&p.storage()[i])
	}
	return q
}
func Vec_WecSortCompare1(p, q *Vec_Int_t) int {
	if len(p.data) < len(q.data) {
		return -1
	}
	if len(p.data) > len(q.data) {
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
		v := compare(&p.storage()[i], &p.storage()[j])
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
	for i := range len(p.data) {
		v := &p.storage()[i]
		if skipSingles != 0 && len(v.data) == 1 {
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
	for i := range len(p.data) {
		v := &p.storage()[i]
		fmt.Printf(" %4d : %2d  {", i, len(v.data))
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
		if len(v.data) >= 2 {
			n++
			*used += len(v.data)
		}
	}
	return n
}
func Vec_WecCollectFirsts(p *Vec_Wec_t) *Vec_Int_t {
	q := Vec_IntAlloc(len(p.data))
	for i := range len(p.data) {
		v := &p.storage()[i]
		if len(v.data) > 0 {
			Vec_IntPush(q, Vec_IntEntry(v, 0))
		}
	}
	return q
}
func Vec_WecConvertToVecPtr(p *Vec_Wec_t) *Vec_Ptr_t {
	q := Vec_PtrAlloc(len(p.data))
	for i := range len(p.data) {
		Vec_PtrPush(q, Vec_IntDup(&p.storage()[i]))
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
	for i := range len(p.data) {
		if len(p.storage()[i].data) > 0 {
			p.storage()[k] = p.storage()[i]
			k++
		}
	}
	clear(p.storage()[k:len(p.data)])
	p.resize(k)
}
