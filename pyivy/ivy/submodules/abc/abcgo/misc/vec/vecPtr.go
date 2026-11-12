package vec

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

type Vec_Ptr_t struct{ data []any }

func Vec_PtrAlloc(n int) *Vec_Ptr_t {
	if n > 0 && n < 8 {
		n = 8
	}
	return Vec_PtrAllocExact(n)
}
func Vec_PtrAllocExact(n int) *Vec_Ptr_t { return &Vec_Ptr_t{data: make([]any, 0, n)} }
func Vec_PtrStart(n int) *Vec_Ptr_t      { p := Vec_PtrAlloc(n); p.resize(n); return p }
func Vec_PtrAllocArray(a []any, n int) *Vec_Ptr_t {
	return &Vec_Ptr_t{data: (a[:n])[:n]}
}
func Vec_PtrAllocArrayCopy(a []any, n int) *Vec_Ptr_t {
	p := Vec_PtrAllocExact(n)
	p.resize(n)
	copy(p.storage(), a[:n])
	return p
}
func Vec_PtrDup(p *Vec_Ptr_t) *Vec_Ptr_t {
	q := Vec_PtrAllocExact(cap(p.data))
	q.resize(len(p.data))
	copy(q.storage(), p.data)
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
func Vec_PtrReleaseArray(p *Vec_Ptr_t) []any       { a := p.data; Vec_PtrZero(p); return a }
func Vec_PtrArray(p *Vec_Ptr_t) []any              { return p.data }
func Vec_PtrSize(p *Vec_Ptr_t) int                 { return len(p.data) }
func Vec_PtrCap(p *Vec_Ptr_t) int                  { return cap(p.data) }
func Vec_PtrEntry(p *Vec_Ptr_t, i int) any         { return p.data[i] }
func Vec_PtrEntryP(p *Vec_Ptr_t, i int) *any       { return &p.data[i] }
func Vec_PtrWriteEntry(p *Vec_Ptr_t, i int, x any) { p.data[i] = x }
func Vec_PtrEntryLast(p *Vec_Ptr_t) any            { return Vec_PtrEntry(p, len(p.data)-1) }
func Vec_PtrGrow(p *Vec_Ptr_t, n int) {
	if n > cap(p.data) {
		data := make([]any, len(p.data), n)
		copy(data, p.data)
		p.data = data
	}
}
func Vec_PtrFill(p *Vec_Ptr_t, n int, x any) {
	Vec_PtrGrow(p, n)
	for i := 0; i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_PtrFillExtra(p *Vec_Ptr_t, n int, x any) {
	if n <= len(p.data) {
		return
	}
	if n > cap(p.data) {
		Vec_PtrGrow(p, max(n, 2*cap(p.data)))
	}
	for i := len(p.data); i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_PtrShrink(p *Vec_Ptr_t, n int) {
	if n < 0 || n > len(p.data) {
		panic("invalid vector shrink")
	}
	clear(p.data[n:])
	p.resize(n)
}
func Vec_PtrClear(p *Vec_Ptr_t) { clear(p.data); p.resize(0) }
func Vec_PtrPush(p *Vec_Ptr_t, x any) {
	if len(p.data) == cap(p.data) {
		Vec_PtrGrow(p, max(16, 2*cap(p.data)))
	}
	p.storage()[len(p.data)] = x
	p.resize(len(p.data) + 1)
}
func Vec_PtrPop(p *Vec_Ptr_t) any {
	x := Vec_PtrEntryLast(p)
	p.data[len(p.data)-1] = nil
	p.resize(len(p.data) - 1)
	return x
}
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
	q := Vec_PtrAlloc(len(p.data))
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
	return float64(cap(p.data)*int(unsafe.Sizeof(uintptr(0))) + int(unsafe.Sizeof(*p)))
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
			p.storage()[i] = even
		} else {
			p.storage()[i] = odd
		}
	}
	p.resize(n)
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
				p.storage()[i] = nil
			}
		}
	}
}
func Vec_PtrFreeFree(p *Vec_Ptr_t) { Vec_PtrFreeData(p); Vec_PtrFree(p) }
func Vec_PtrCopy(dst, src *Vec_Ptr_t) {
	Vec_PtrGrow(dst, len(src.data))
	copy(dst.storage(), Vec_PtrArray(src))
	dst.resize(len(src.data))
}
func Vec_PtrPrintNames(p *Vec_Ptr_t) {
	fmt.Printf("Vector has %d entries: {", len(p.data))
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
	copy(p.storage()[1:len(p.data)], p.storage()[:len(p.data)-1])
	p.storage()[0] = x
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
	for i := len(p.data) - 1; i >= 0; i-- {
		if ptrEqual(p.storage()[i], x) {
			Vec_PtrDrop(p, i)
			return
		}
	}
	panic("entry not found")
}
func Vec_PtrDrop(p *Vec_Ptr_t, i int) {
	_ = Vec_PtrEntry(p, i)
	copy(p.storage()[i:], p.storage()[i+1:len(p.data)])
	p.resize(len(p.data) - 1)
	p.storage()[len(p.data)] = nil
}
func Vec_PtrInsert(p *Vec_Ptr_t, i int, x any) {
	if i < 0 || i >= len(p.data) {
		panic("insert index")
	}
	Vec_PtrPush(p, nil)
	copy(p.storage()[i+1:len(p.data)], p.storage()[i:len(p.data)-1])
	p.storage()[i] = x
}
func Vec_PtrReorder(p *Vec_Ptr_t, n int) {
	if n < 0 || n >= len(p.data) {
		panic("reorder count")
	}
	Vec_PtrGrow(p, len(p.data)+n)
	copy(p.storage()[len(p.data):], p.storage()[:n])
	copy(p.storage(), p.storage()[n:n+len(p.data)])
}
func Vec_PtrReverseOrder(p *Vec_Ptr_t) {
	for i, j := 0, len(p.data)-1; i < j; i, j = i+1, j-1 {
		p.storage()[i], p.storage()[j] = p.storage()[j], p.storage()[i]
	}
}
func Vec_PtrEqual(p, q *Vec_Ptr_t) int {
	if len(p.data) != len(q.data) {
		return 0
	}
	for i, v := range Vec_PtrArray(p) {
		if !ptrEqual(v, q.storage()[i]) {
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
	sort.SliceStable(Vec_PtrArray(p), func(i, j int) bool { return compare(&p.storage()[i], &p.storage()[j]) < 0 })
}
func Vec_PtrUniqify(p *Vec_Ptr_t, compare func(*any, *any) int) {
	if len(p.data) < 2 {
		return
	}
	Vec_PtrSort(p, compare)
	k := 1
	for i := 1; i < len(p.data); i++ {
		if !ptrEqual(p.storage()[i], p.storage()[i-1]) {
			p.storage()[k] = p.storage()[i]
			k++
		}
	}
	clear(p.data[k:])
	p.resize(k)
}
func Vec_PtrUniqify2(p *Vec_Ptr_t, compare func(*any, *any) int, free func(any), counts *Vec_Int_t) {
	if counts != nil {
		Vec_IntFill(counts, min(1, len(p.data)), 1)
	}
	if len(p.data) < 2 {
		return
	}
	Vec_PtrSort(p, compare)
	k := 1
	for i := 1; i < len(p.data); i++ {
		if compare(&p.storage()[i], &p.storage()[k-1]) != 0 {
			p.storage()[k] = p.storage()[i]
			k++
			if counts != nil {
				Vec_IntPush(counts, 1)
			}
		} else {
			if free != nil {
				free(p.storage()[i])
			}
			if counts != nil {
				Vec_IntAddToEntry(counts, len(counts.data)-1, 1)
			}
		}
	}
	clear(p.data[k:])
	p.resize(k)
}
func Vec_PtrAllocSimInfo(entries, words int) *Vec_Ptr_t {
	p := Vec_PtrAllocExact(entries)
	p.resize(entries)
	data := make([]uint32, entries*words)
	for i := range entries {
		p.storage()[i] = data[i*words : (i+1)*words]
	}
	return p
}
func Vec_PtrReadWordsSimInfo(p *Vec_Ptr_t) int {
	if len(p.data) < 2 {
		panic("simulation entries")
	}
	return len(p.storage()[0].([]uint32))
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
	q := Vec_PtrAllocSimInfo(2*len(p.data), Vec_PtrReadWordsSimInfo(p))
	for i, x := range Vec_PtrArray(p) {
		copy(q.storage()[i].([]uint32), x.([]uint32))
	}
	*p = *q
}
func Vec_PtrReallocSimInfo(p *Vec_Ptr_t) {
	q := Vec_PtrAllocSimInfo(len(p.data), 2*Vec_PtrReadWordsSimInfo(p))
	for i, x := range Vec_PtrArray(p) {
		copy(q.storage()[i].([]uint32), x.([]uint32))
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
