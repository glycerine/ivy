package vec

import (
	"fmt"
	"math"
	"sort"
	"unsafe"
)

// Vec_Flt_t stores its live entries in one slice; capacity is derived.
type Vec_Flt_t struct{ data []float32 }

func Vec_FltAlloc(n int) *Vec_Flt_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_FltAllocExact(n)
}
func Vec_FltAllocExact(n int) *Vec_Flt_t { return &Vec_Flt_t{data: make([]float32, 0, n)} }
func Vec_FltStart(n int) *Vec_Flt_t      { p := Vec_FltAlloc(n); p.resize(n); return p }

// StartFull matches C's bytewise 0xff initialization, including its NaN payload.
func Vec_FltStartFull(n int) *Vec_Flt_t {
	p := Vec_FltStart(n)
	for i := 0; i < n; i++ {
		p.storage()[i] = math.Float32frombits(0xffffffff)
	}
	return p
}
func Vec_FltAllocArray(a []float32, n int) *Vec_Flt_t {
	return &Vec_Flt_t{data: (a[:n])[:n]}
}
func Vec_FltAllocArrayCopy(a []float32, n int) *Vec_Flt_t {
	p := Vec_FltAllocExact(n)
	p.resize(n)
	copy(p.storage(), a[:n])
	return p
}
func Vec_FltDup(p *Vec_Flt_t) *Vec_Flt_t {
	q := Vec_FltAllocExact(cap(p.data))
	q.resize(len(p.data))
	copy(q.storage(), p.data)
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
func Vec_FltReleaseArray(p *Vec_Flt_t) []float32 { a := p.data; Vec_FltZero(p); return a }
func Vec_FltArray(p *Vec_Flt_t) []float32        { return p.data }
func Vec_FltArrayP(p *Vec_Flt_t) *[]float32      { return &p.data }
func Vec_FltSize(p *Vec_Flt_t) int               { return len(p.data) }
func Vec_FltCap(p *Vec_Flt_t) int                { return cap(p.data) }
func Vec_FltMemory(p *Vec_Flt_t) float64 {
	if p == nil {
		return 0
	}
	return float64(4*cap(p.data) + int(unsafe.Sizeof(*p)))
}
func Vec_FltEntry(p *Vec_Flt_t, i int) float32         { return p.data[i] }
func Vec_FltEntryP(p *Vec_Flt_t, i int) *float32       { return &p.data[i] }
func Vec_FltWriteEntry(p *Vec_Flt_t, i int, x float32) { p.data[i] = x }
func Vec_FltAddToEntry(p *Vec_Flt_t, i int, x float32) { p.data[i] += x }
func Vec_FltUpdateEntry(p *Vec_Flt_t, i int, x float32) {
	if Vec_FltEntry(p, i) < x {
		Vec_FltWriteEntry(p, i, x)
	}
}
func Vec_FltEntryLast(p *Vec_Flt_t) float32 { return Vec_FltEntry(p, len(p.data)-1) }
func Vec_FltGrow(p *Vec_Flt_t, n int) {
	if n > cap(p.data) {
		data := make([]float32, len(p.data), n)
		copy(data, p.data)
		p.data = data
	}
}
func Vec_FltFill(p *Vec_Flt_t, n int, x float32) {
	Vec_FltGrow(p, n)
	for i := 0; i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_FltFillExtra(p *Vec_Flt_t, n int, x float32) {
	if n <= len(p.data) {
		return
	}
	if n > cap(p.data) {
		Vec_FltGrow(p, max(n, 2*cap(p.data)))
	}
	for i := len(p.data); i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_FltShrink(p *Vec_Flt_t, n int) {
	if n < 0 || n > len(p.data) {
		panic("invalid vector shrink")
	}
	p.resize(n)
}
func Vec_FltClear(p *Vec_Flt_t) { p.resize(0) }
func Vec_FltPush(p *Vec_Flt_t, x float32) {
	if len(p.data) == cap(p.data) {
		Vec_FltGrow(p, max(16, 2*cap(p.data)))
	}
	p.storage()[len(p.data)] = x
	p.resize(len(p.data) + 1)
}
func Vec_FltPushOrder(p *Vec_Flt_t, x float32) {
	Vec_FltPush(p, x)
	i := len(p.data) - 1
	for i > 0 && p.storage()[i-1] > x {
		p.storage()[i] = p.storage()[i-1]
		i--
	}
	p.storage()[i] = x
}
func Vec_FltPushUnique(p *Vec_Flt_t, x float32) int {
	if Vec_FltFind(p, x) >= 0 {
		return 1
	}
	Vec_FltPush(p, x)
	return 0
}
func Vec_FltPop(p *Vec_Flt_t) float32 { x := Vec_FltEntryLast(p); p.resize(len(p.data) - 1); return x }
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
	copy(p.storage()[i:], p.storage()[i+1:len(p.data)])
	p.resize(len(p.data) - 1)
	return 1
}
func Vec_FltFindMax(p *Vec_Flt_t) float32 {
	if len(p.data) == 0 {
		return 0
	}
	v := p.storage()[0]
	for _, x := range p.storage()[1:len(p.data)] {
		if v < x {
			v = x
		}
	}
	return v
}
func Vec_FltFindMin(p *Vec_Flt_t) float32 {
	if len(p.data) == 0 {
		return 0
	}
	v := p.storage()[0]
	for _, x := range p.storage()[1:len(p.data)] {
		if v > x {
			v = x
		}
	}
	return v
}
func Vec_FltEqual(p, q *Vec_Flt_t) int {
	if len(p.data) != len(q.data) {
		return 0
	}
	for i, v := range Vec_FltArray(p) {
		if v != q.storage()[i] {
			return 0
		}
	}
	return 1
}
func Vec_FltPrint(p *Vec_Flt_t) {
	fmt.Printf("Vector has %d entries: {", len(p.data))
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
	sort.Slice(p.data, func(i, j int) bool {
		if reverse != 0 {
			return p.storage()[i] > p.storage()[j]
		}
		return p.storage()[i] < p.storage()[j]
	})
}
