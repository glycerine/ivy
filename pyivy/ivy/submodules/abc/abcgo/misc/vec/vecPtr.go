package vec

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

type Vec_Ptr_t struct {
	nCap, nSize int
	pArray      []any
}

func Vec_PtrAlloc(n int) *Vec_Ptr_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_PtrAllocExact(n)
}
func Vec_PtrAllocExact(n int) *Vec_Ptr_t { return &Vec_Ptr_t{nCap: n, pArray: make([]any, n)} }
func Vec_PtrStart(n int) *Vec_Ptr_t      { p := Vec_PtrAlloc(n); p.nSize = n; return p }
func Vec_PtrAllocArray(a []any, n int) *Vec_Ptr_t {
	return &Vec_Ptr_t{nCap: n, nSize: n, pArray: a[:n]}
}
func Vec_PtrAllocArrayCopy(a []any, n int) *Vec_Ptr_t {
	p := Vec_PtrAllocExact(n)
	p.nSize = n
	copy(p.pArray, a[:n])
	return p
}
func Vec_PtrDup(p *Vec_Ptr_t) *Vec_Ptr_t {
	q := Vec_PtrAllocExact(p.nCap)
	q.nSize = p.nSize
	copy(q.pArray, p.pArray[:p.nSize])
	return q
}
func Vec_PtrDupArray(p *Vec_Ptr_t) *Vec_Ptr_t { q := *p; Vec_PtrZero(p); return &q }
func Vec_PtrZero(p *Vec_Ptr_t)                { *p = Vec_Ptr_t{} }
func Vec_PtrErase(p *Vec_Ptr_t)               { Vec_PtrZero(p) }
func Vec_PtrFree(p *Vec_Ptr_t) {
	if p != nil {
		Vec_PtrZero(p)
	}
}
func Vec_PtrFreeP(p **Vec_Ptr_t)                   { Vec_PtrFree(*p); *p = nil }
func Vec_PtrReleaseArray(p *Vec_Ptr_t) []any       { a := p.pArray; Vec_PtrZero(p); return a }
func Vec_PtrArray(p *Vec_Ptr_t) []any              { return p.pArray[:p.nSize] }
func Vec_PtrSize(p *Vec_Ptr_t) int                 { return p.nSize }
func Vec_PtrCap(p *Vec_Ptr_t) int                  { return p.nCap }
func Vec_PtrEntry(p *Vec_Ptr_t, i int) any         { return p.pArray[:p.nSize][i] }
func Vec_PtrEntryP(p *Vec_Ptr_t, i int) *any       { return &p.pArray[:p.nSize][i] }
func Vec_PtrWriteEntry(p *Vec_Ptr_t, i int, x any) { p.pArray[:p.nSize][i] = x }
func Vec_PtrEntryLast(p *Vec_Ptr_t) any            { return Vec_PtrEntry(p, p.nSize-1) }
func Vec_PtrGrow(p *Vec_Ptr_t, n int) {
	if n > p.nCap {
		p.pArray = append(p.pArray, make([]any, n-p.nCap)...)
		p.nCap = n
	}
}
func Vec_PtrFill(p *Vec_Ptr_t, n int, x any) {
	Vec_PtrGrow(p, n)
	for i := 0; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_PtrFillExtra(p *Vec_Ptr_t, n int, x any) {
	if n <= p.nSize {
		return
	}
	if n > p.nCap {
		Vec_PtrGrow(p, max(n, 2*p.nCap))
	}
	for i := p.nSize; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_PtrShrink(p *Vec_Ptr_t, n int) {
	if n < 0 || n > p.nSize {
		panic("invalid vector shrink")
	}
	p.nSize = n
}
func Vec_PtrClear(p *Vec_Ptr_t) { p.nSize = 0 }
func Vec_PtrPush(p *Vec_Ptr_t, x any) {
	if p.nSize == p.nCap {
		Vec_PtrGrow(p, max(16, 2*p.nCap))
	}
	p.pArray[p.nSize] = x
	p.nSize++
}
func Vec_PtrPop(p *Vec_Ptr_t) any { x := Vec_PtrEntryLast(p); p.nSize--; return x }
func ptrNil(x any) bool {
	if x == nil {
		return true
	}
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	}
	return false
}
func ptrAddress(x any) uintptr {
	if ptrNil(x) {
		return 0
	}
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return v.Pointer()
	case reflect.Uintptr:
		return uintptr(v.Uint())
	case reflect.Int, reflect.Int32, reflect.Int64:
		return uintptr(v.Int())
	}
	panic("pointer ordering requires pointers or uintptr")
}
func ptrEqual(a, b any) bool {
	if ptrNil(a) || ptrNil(b) {
		return ptrNil(a) && ptrNil(b)
	}
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	v := reflect.ValueOf(a)
	if v.Comparable() {
		return a == b
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Func:
		return ptrAddress(a) == ptrAddress(b)
	}
	return false
}
func Vec_PtrDupStr(p *Vec_Ptr_t) *Vec_Ptr_t {
	q := Vec_PtrAlloc(p.nSize)
	for _, x := range Vec_PtrArray(p) {
		if x == nil {
			Vec_PtrPush(q, nil)
		} else {
			Vec_PtrPush(q, strings.Clone(x.(string)))
		}
	}
	return q
}
func Vec_PtrMemory(p *Vec_Ptr_t) float64 {
	if p == nil {
		return 0
	}
	return float64(p.nCap*int(unsafe.Sizeof(uintptr(0))) + int(unsafe.Sizeof(*p)))
}
func Vec_PtrCountZero(p *Vec_Ptr_t) int {
	n := 0
	for _, x := range Vec_PtrArray(p) {
		if ptrNil(x) {
			n++
		}
	}
	return n
}
func Vec_PtrFillTwo(p *Vec_Ptr_t, n int, even, odd any) {
	Vec_PtrGrow(p, n)
	for i := range n {
		if i&1 == 0 {
			p.pArray[i] = even
		} else {
			p.pArray[i] = odd
		}
	}
	p.nSize = n
}
func Vec_PtrGetEntry(p *Vec_Ptr_t, i int) any {
	Vec_PtrFillExtra(p, i+1, nil)
	return Vec_PtrEntry(p, i)
}
func Vec_PtrSetEntry(p *Vec_Ptr_t, i int, x any) {
	Vec_PtrFillExtra(p, i+1, nil)
	Vec_PtrWriteEntry(p, i, x)
}

// GC owns Go allocations; dropping the references releases vector-owned data.
func Vec_PtrFreeData(p *Vec_Ptr_t) {
	if p != nil {
		for i, x := range Vec_PtrArray(p) {
			if x != uintptr(1) && x != uintptr(2) {
				p.pArray[i] = nil
			}
		}
	}
}
func Vec_PtrFreeFree(p *Vec_Ptr_t) { Vec_PtrFreeData(p); Vec_PtrFree(p) }
func Vec_PtrCopy(dst, src *Vec_Ptr_t) {
	Vec_PtrGrow(dst, src.nSize)
	copy(dst.pArray, Vec_PtrArray(src))
	dst.nSize = src.nSize
}
func Vec_PtrPrintNames(p *Vec_Ptr_t) {
	fmt.Printf("Vector has %d entries: {", p.nSize)
	for _, x := range Vec_PtrArray(p) {
		fmt.Printf("%s ", x)
	}
	fmt.Println(" }")
}
func Vec_PtrPushTwo(p *Vec_Ptr_t, a, b any) { Vec_PtrPush(p, a); Vec_PtrPush(p, b) }
func Vec_PtrAppend(p, q *Vec_Ptr_t) {
	for _, x := range Vec_PtrArray(q) {
		Vec_PtrPush(p, x)
	}
}
func Vec_PtrPushFirst(p *Vec_Ptr_t, x any) {
	Vec_PtrPush(p, nil)
	copy(p.pArray[1:p.nSize], p.pArray[:p.nSize-1])
	p.pArray[0] = x
}
func Vec_PtrPushUnique(p *Vec_Ptr_t, x any) int {
	if Vec_PtrFind(p, x) >= 0 {
		return 1
	}
	Vec_PtrPush(p, x)
	return 0
}
func Vec_PtrFind(p *Vec_Ptr_t, x any) int {
	for i, v := range Vec_PtrArray(p) {
		if ptrEqual(v, x) {
			return i
		}
	}
	return -1
}
func Vec_PtrFindStr(p *Vec_Ptr_t, x string) int {
	for i, v := range Vec_PtrArray(p) {
		if s, ok := v.(string); ok && s == x {
			return i
		}
	}
	return -1
}
func Vec_PtrRemove(p *Vec_Ptr_t, x any) {
	for i := p.nSize - 1; i >= 0; i-- {
		if ptrEqual(p.pArray[i], x) {
			Vec_PtrDrop(p, i)
			return
		}
	}
	panic("entry not found")
}
func Vec_PtrDrop(p *Vec_Ptr_t, i int) {
	_ = Vec_PtrEntry(p, i)
	copy(p.pArray[i:], p.pArray[i+1:p.nSize])
	p.nSize--
	p.pArray[p.nSize] = nil
}
func Vec_PtrInsert(p *Vec_Ptr_t, i int, x any) {
	if i < 0 || i >= p.nSize {
		panic("insert index")
	}
	Vec_PtrPush(p, nil)
	copy(p.pArray[i+1:p.nSize], p.pArray[i:p.nSize-1])
	p.pArray[i] = x
}
func Vec_PtrReorder(p *Vec_Ptr_t, n int) {
	if n < 0 || n >= p.nSize {
		panic("reorder count")
	}
	Vec_PtrGrow(p, p.nSize+n)
	copy(p.pArray[p.nSize:], p.pArray[:n])
	copy(p.pArray, p.pArray[n:n+p.nSize])
}
func Vec_PtrReverseOrder(p *Vec_Ptr_t) {
	for i, j := 0, p.nSize-1; i < j; i, j = i+1, j-1 {
		p.pArray[i], p.pArray[j] = p.pArray[j], p.pArray[i]
	}
}
func Vec_PtrEqual(p, q *Vec_Ptr_t) int {
	if p.nSize != q.nSize {
		return 0
	}
	for i, v := range Vec_PtrArray(p) {
		if !ptrEqual(v, q.pArray[i]) {
			return 0
		}
	}
	return 1
}
func Vec_PtrSortComparePtr(a, b *any) int {
	av, bv := ptrAddress(*a), ptrAddress(*b)
	if av < bv {
		return -1
	}
	if av > bv {
		return 1
	}
	return 0
}
func Vec_PtrSort(p *Vec_Ptr_t, compare func(*any, *any) int) {
	if compare == nil {
		compare = Vec_PtrSortComparePtr
	}
	sort.SliceStable(Vec_PtrArray(p), func(i, j int) bool { return compare(&p.pArray[i], &p.pArray[j]) < 0 })
}
func Vec_PtrUniqify(p *Vec_Ptr_t, compare func(*any, *any) int) {
	if p.nSize < 2 {
		return
	}
	Vec_PtrSort(p, compare)
	k := 1
	for i := 1; i < p.nSize; i++ {
		if !ptrEqual(p.pArray[i], p.pArray[i-1]) {
			p.pArray[k] = p.pArray[i]
			k++
		}
	}
	p.nSize = k
}
func Vec_PtrUniqify2(p *Vec_Ptr_t, compare func(*any, *any) int, free func(any), counts *Vec_Int_t) {
	if counts != nil {
		Vec_IntFill(counts, 1, 1)
	}
	if p.nSize < 2 {
		return
	}
	Vec_PtrSort(p, compare)
	k := 1
	for i := 1; i < p.nSize; i++ {
		if compare(&p.pArray[i], &p.pArray[k-1]) != 0 {
			p.pArray[k] = p.pArray[i]
			k++
			if counts != nil {
				Vec_IntPush(counts, 1)
			}
		} else {
			if free != nil {
				free(p.pArray[i])
			}
			if counts != nil {
				Vec_IntAddToEntry(counts, counts.nSize-1, 1)
			}
		}
	}
	p.nSize = k
}
func Vec_PtrAllocSimInfo(entries, words int) *Vec_Ptr_t {
	p := Vec_PtrAllocExact(entries)
	p.nSize = entries
	data := make([]uint32, entries*words)
	for i := range entries {
		p.pArray[i] = data[i*words : (i+1)*words]
	}
	return p
}
func Vec_PtrReadWordsSimInfo(p *Vec_Ptr_t) int {
	if p.nSize < 2 {
		panic("simulation entries")
	}
	return len(p.pArray[0].([]uint32))
}
func Vec_PtrCleanSimInfo(p *Vec_Ptr_t, start, words int) {
	for _, x := range Vec_PtrArray(p) {
		clear(x.([]uint32)[start:words])
	}
}
func Vec_PtrFillSimInfo(p *Vec_Ptr_t, start, words int) {
	for _, x := range Vec_PtrArray(p) {
		a := x.([]uint32)
		for i := start; i < words; i++ {
			a[i] = 0xffffffff
		}
	}
}
func Vec_PtrDoubleSimInfo(p *Vec_Ptr_t) {
	q := Vec_PtrAllocSimInfo(2*p.nSize, Vec_PtrReadWordsSimInfo(p))
	for i, x := range Vec_PtrArray(p) {
		copy(q.pArray[i].([]uint32), x.([]uint32))
	}
	*p = *q
}
func Vec_PtrReallocSimInfo(p *Vec_Ptr_t) {
	q := Vec_PtrAllocSimInfo(p.nSize, 2*Vec_PtrReadWordsSimInfo(p))
	for i, x := range Vec_PtrArray(p) {
		copy(q.pArray[i].([]uint32), x.([]uint32))
	}
	*p = *q
}
func Vec_PtrAllocTruthTables(vars int) *Vec_Ptr_t {
	words := 1
	if vars > 5 {
		words = 1 << (vars - 5)
	}
	p := Vec_PtrAllocSimInfo(vars, words)
	masks := [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	for i, x := range Vec_PtrArray(p) {
		a := x.([]uint32)
		for k := range a {
			if i < 5 {
				a[k] = masks[i]
			} else if k&(1<<(i-5)) != 0 {
				a[k] = 0xffffffff
			}
		}
	}
	return p
}
