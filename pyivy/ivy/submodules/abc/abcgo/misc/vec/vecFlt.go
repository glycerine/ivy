package vec

import (
	"fmt"
	"math"
	"sort"
	"unsafe"
)

// Vec_Flt_t retains ABC's logical size and allocated capacity separately.
// pArray contains capacity elements; Array exposes only the live prefix.
type Vec_Flt_t struct {
	nCap, nSize int
	pArray      []float32
}

func Vec_FltAlloc(n int) *Vec_Flt_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_FltAllocExact(n)
}
func Vec_FltAllocExact(n int) *Vec_Flt_t { return &Vec_Flt_t{nCap: n, pArray: make([]float32, n)} }
func Vec_FltStart(n int) *Vec_Flt_t      { p := Vec_FltAlloc(n); p.nSize = n; return p }

// StartFull matches C's bytewise 0xff initialization, including its NaN payload.
func Vec_FltStartFull(n int) *Vec_Flt_t {
	p := Vec_FltStart(n)
	for i := 0; i < n; i++ {
		p.pArray[i] = math.Float32frombits(0xffffffff)
	}
	return p
}
func Vec_FltAllocArray(a []float32, n int) *Vec_Flt_t {
	return &Vec_Flt_t{nCap: n, nSize: n, pArray: a[:n]}
}
func Vec_FltAllocArrayCopy(a []float32, n int) *Vec_Flt_t {
	p := Vec_FltAllocExact(n)
	p.nSize = n
	copy(p.pArray, a[:n])
	return p
}
func Vec_FltDup(p *Vec_Flt_t) *Vec_Flt_t {
	q := Vec_FltAllocExact(p.nCap)
	q.nSize = p.nSize
	copy(q.pArray, p.pArray[:p.nSize])
	return q
}
func Vec_FltDupArray(p *Vec_Flt_t) *Vec_Flt_t { q := *p; Vec_FltZero(p); return &q }
func Vec_FltZero(p *Vec_Flt_t)                { *p = Vec_Flt_t{} }
func Vec_FltErase(p *Vec_Flt_t)               { Vec_FltZero(p) }
func Vec_FltFree(p *Vec_Flt_t) {
	if p != nil {
		Vec_FltZero(p)
	}
}
func Vec_FltFreeP(p **Vec_Flt_t)                 { Vec_FltFree(*p); *p = nil }
func Vec_FltReleaseArray(p *Vec_Flt_t) []float32 { a := p.pArray; Vec_FltZero(p); return a }
func Vec_FltArray(p *Vec_Flt_t) []float32        { return p.pArray[:p.nSize] }
func Vec_FltArrayP(p *Vec_Flt_t) *[]float32      { return &p.pArray }
func Vec_FltSize(p *Vec_Flt_t) int               { return p.nSize }
func Vec_FltCap(p *Vec_Flt_t) int                { return p.nCap }
func Vec_FltMemory(p *Vec_Flt_t) float64 {
	if p == nil {
		return 0
	}
	return float64(4*p.nCap + int(unsafe.Sizeof(*p)))
}
func Vec_FltEntry(p *Vec_Flt_t, i int) float32         { return p.pArray[:p.nSize][i] }
func Vec_FltEntryP(p *Vec_Flt_t, i int) *float32       { return &p.pArray[:p.nSize][i] }
func Vec_FltWriteEntry(p *Vec_Flt_t, i int, x float32) { p.pArray[:p.nSize][i] = x }
func Vec_FltAddToEntry(p *Vec_Flt_t, i int, x float32) { p.pArray[:p.nSize][i] += x }
func Vec_FltUpdateEntry(p *Vec_Flt_t, i int, x float32) {
	if Vec_FltEntry(p, i) < x {
		Vec_FltWriteEntry(p, i, x)
	}
}
func Vec_FltEntryLast(p *Vec_Flt_t) float32 { return Vec_FltEntry(p, p.nSize-1) }
func Vec_FltGrow(p *Vec_Flt_t, n int) {
	if n > p.nCap {
		p.pArray = append(p.pArray, make([]float32, n-p.nCap)...)
		p.nCap = n
	}
}
func Vec_FltFill(p *Vec_Flt_t, n int, x float32) {
	Vec_FltGrow(p, n)
	for i := 0; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_FltFillExtra(p *Vec_Flt_t, n int, x float32) {
	if n <= p.nSize {
		return
	}
	if n > p.nCap {
		Vec_FltGrow(p, max(n, 2*p.nCap))
	}
	for i := p.nSize; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_FltShrink(p *Vec_Flt_t, n int) {
	if n < 0 || n > p.nSize {
		panic("invalid vector shrink")
	}
	p.nSize = n
}
func Vec_FltClear(p *Vec_Flt_t) { p.nSize = 0 }
func Vec_FltPush(p *Vec_Flt_t, x float32) {
	if p.nSize == p.nCap {
		Vec_FltGrow(p, max(16, 2*p.nCap))
	}
	p.pArray[p.nSize] = x
	p.nSize++
}
func Vec_FltPushOrder(p *Vec_Flt_t, x float32) {
	Vec_FltPush(p, x)
	i := p.nSize - 1
	for i > 0 && p.pArray[i-1] > x {
		p.pArray[i] = p.pArray[i-1]
		i--
	}
	p.pArray[i] = x
}
func Vec_FltPushUnique(p *Vec_Flt_t, x float32) int {
	if Vec_FltFind(p, x) >= 0 {
		return 1
	}
	Vec_FltPush(p, x)
	return 0
}
func Vec_FltPop(p *Vec_Flt_t) float32 { x := Vec_FltEntryLast(p); p.nSize--; return x }
func Vec_FltFind(p *Vec_Flt_t, x float32) int {
	for i, v := range Vec_FltArray(p) {
		if v == x {
			return i
		}
	}
	return -1
}
func Vec_FltRemove(p *Vec_Flt_t, x float32) int {
	i := Vec_FltFind(p, x)
	if i < 0 {
		return 0
	}
	copy(p.pArray[i:], p.pArray[i+1:p.nSize])
	p.nSize--
	return 1
}
func Vec_FltFindMax(p *Vec_Flt_t) float32 {
	if p.nSize == 0 {
		return 0
	}
	v := p.pArray[0]
	for _, x := range p.pArray[1:p.nSize] {
		if v < x {
			v = x
		}
	}
	return v
}
func Vec_FltFindMin(p *Vec_Flt_t) float32 {
	if p.nSize == 0 {
		return 0
	}
	v := p.pArray[0]
	for _, x := range p.pArray[1:p.nSize] {
		if v > x {
			v = x
		}
	}
	return v
}
func Vec_FltEqual(p, q *Vec_Flt_t) int {
	if p.nSize != q.nSize {
		return 0
	}
	for i, v := range Vec_FltArray(p) {
		if v != q.pArray[i] {
			return 0
		}
	}
	return 1
}
func Vec_FltPrint(p *Vec_Flt_t) {
	fmt.Printf("Vector has %d entries: {", p.nSize)
	for _, x := range Vec_FltArray(p) {
		fmt.Printf(" %f", x)
	}
	fmt.Println(" }")
}
func Vec_FltSortCompare1(a, b *float32) int {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func Vec_FltSortCompare2(a, b *float32) int { return -Vec_FltSortCompare1(a, b) }
func Vec_FltSort(p *Vec_Flt_t, reverse int) {
	sort.Slice(p.pArray[:p.nSize], func(i, j int) bool {
		if reverse != 0 {
			return p.pArray[i] > p.pArray[j]
		}
		return p.pArray[i] < p.pArray[j]
	})
}
