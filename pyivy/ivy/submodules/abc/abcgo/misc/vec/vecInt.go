package vec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"unsafe"
)

// Vec_Int_t stores its live entries in one slice; capacity is derived.
type Vec_Int_t struct {
	marked bool
	data   []int32
}

func Vec_IntAlloc(n int) *Vec_Int_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_IntAllocExact(n)
}
func Vec_IntAllocExact(n int) *Vec_Int_t { return &Vec_Int_t{data: make([]int32, 0, n)} }
func Vec_IntStart(n int) *Vec_Int_t      { p := Vec_IntAlloc(n); p.resize(n); return p }

// StartFull matches C's bytewise 0xff initialization, setting every integer to -1.
func Vec_IntStartFull(n int) *Vec_Int_t {
	p := Vec_IntStart(n)
	for i := 0; i < n; i++ {
		p.storage()[i] = -1
	}
	return p
}
func Vec_IntAllocArray(a []int32, n int) *Vec_Int_t {
	return &Vec_Int_t{data: (a[:n])[:n]}
}
func Vec_IntAllocArrayCopy(a []int32, n int) *Vec_Int_t {
	p := Vec_IntAllocExact(n)
	p.resize(n)
	copy(p.storage(), a[:n])
	return p
}
func Vec_IntDup(p *Vec_Int_t) *Vec_Int_t {
	q := Vec_IntAllocExact(cap(p.data))
	q.resize(len(p.data))
	copy(q.storage(), p.data)
	return q
}
func Vec_IntDupArray(p *Vec_Int_t) *Vec_Int_t { q := *p; Vec_IntZero(p); return &q }
func Vec_IntZero(p *Vec_Int_t)                { *p = Vec_Int_t{} }
func Vec_IntErase(p *Vec_Int_t)               { Vec_IntZero(p) }
func Vec_IntFree(p *Vec_Int_t) {
	if p != nil {
		Vec_IntZero(p)
	}
}
func Vec_IntFreeP(p **Vec_Int_t)               { Vec_IntFree(*p); *p = nil }
func Vec_IntReleaseArray(p *Vec_Int_t) []int32 { a := p.data; Vec_IntZero(p); return a }
func Vec_IntArray(p *Vec_Int_t) []int32        { return p.data }
func Vec_IntArrayP(p *Vec_Int_t) *[]int32      { return &p.data }
func Vec_IntSize(p *Vec_Int_t) int             { return len(p.data) }
func Vec_IntCap(p *Vec_Int_t) int              { return cap(p.data) }
func Vec_IntMemory(p *Vec_Int_t) float64 {
	if p == nil {
		return 0
	}
	return float64(4*cap(p.data) + int(unsafe.Sizeof(*p)))
}
func Vec_IntEntry(p *Vec_Int_t, i int) int32         { return p.data[i] }
func Vec_IntEntryP(p *Vec_Int_t, i int) *int32       { return &p.data[i] }
func Vec_IntWriteEntry(p *Vec_Int_t, i int, x int32) { p.data[i] = x }
func Vec_IntAddToEntry(p *Vec_Int_t, i int, x int32) int32 {
	p.data[i] += x
	return p.storage()[i]
}
func Vec_IntUpdateEntry(p *Vec_Int_t, i int, x int32) {
	if Vec_IntEntry(p, i) < x {
		Vec_IntWriteEntry(p, i, x)
	}
}
func Vec_IntEntryLast(p *Vec_Int_t) int32 { return Vec_IntEntry(p, len(p.data)-1) }
func Vec_IntGrow(p *Vec_Int_t, n int) {
	if n > cap(p.data) {
		data := make([]int32, len(p.data), n)
		copy(data, p.data)
		p.data = data
	}
}
func Vec_IntFill(p *Vec_Int_t, n int, x int32) {
	Vec_IntGrow(p, n)
	for i := 0; i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_IntFillExtra(p *Vec_Int_t, n int, x int32) {
	if n <= len(p.data) {
		return
	}
	if n > cap(p.data) {
		Vec_IntGrow(p, max(n, 2*cap(p.data)))
	}
	for i := len(p.data); i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_IntShrink(p *Vec_Int_t, n int) {
	if n < 0 || n > len(p.data) {
		panic("invalid vector shrink")
	}
	p.resize(n)
}
func Vec_IntClear(p *Vec_Int_t) { p.resize(0) }
func Vec_IntPush(p *Vec_Int_t, x int32) {
	if len(p.data) == cap(p.data) {
		Vec_IntGrow(p, max(16, 2*cap(p.data)))
	}
	p.storage()[len(p.data)] = x
	p.resize(len(p.data) + 1)
}
func Vec_IntPushOrder(p *Vec_Int_t, x int32) {
	Vec_IntPush(p, x)
	i := len(p.data) - 1
	for i > 0 && p.storage()[i-1] > x {
		p.storage()[i] = p.storage()[i-1]
		i--
	}
	p.storage()[i] = x
}
func Vec_IntPushUnique(p *Vec_Int_t, x int32) int {
	if Vec_IntFind(p, x) >= 0 {
		return 1
	}
	Vec_IntPush(p, x)
	return 0
}
func Vec_IntPop(p *Vec_Int_t) int32 { x := Vec_IntEntryLast(p); p.resize(len(p.data) - 1); return x }
func Vec_IntFind(p *Vec_Int_t, x int32) int {
	for i, v := range Vec_IntArray(p) {
		if v == x {
			return i
		}
	}
	return -1
}
func Vec_IntRemove(p *Vec_Int_t, x int32) int {
	i := Vec_IntFind(p, x)
	if i < 0 {
		return 0
	}
	copy(p.storage()[i:], p.storage()[i+1:len(p.data)])
	p.resize(len(p.data) - 1)
	return 1
}
func Vec_IntFindMax(p *Vec_Int_t) int32 {
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
func Vec_IntFindMin(p *Vec_Int_t) int32 {
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
func Vec_IntEqual(p, q *Vec_Int_t) int {
	if len(p.data) != len(q.data) {
		return 0
	}
	for i, v := range Vec_IntArray(p) {
		if v != q.storage()[i] {
			return 0
		}
	}
	return 1
}
func Vec_IntPrint(p *Vec_Int_t) {
	fmt.Printf("Vector has %d entries: {", len(p.data))
	for _, x := range Vec_IntArray(p) {
		fmt.Printf(" %d", x)
	}
	fmt.Println(" }")
}
func Vec_IntSortCompare1(a, b *int32) int {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func Vec_IntSortCompare2(a, b *int32) int { return -Vec_IntSortCompare1(a, b) }
func Vec_IntSort(p *Vec_Int_t, reverse int) {
	sort.Slice(p.data, func(i, j int) bool {
		if reverse != 0 {
			return p.storage()[i] > p.storage()[j]
		}
		return p.storage()[i] < p.storage()[j]
	})
}

func Vec_IntStartRange(first int32, n int) *Vec_Int_t {
	p := Vec_IntStart(n)
	for i := range n {
		p.storage()[i] = first + int32(i)
	}
	return p
}
func Vec_IntStartNatural(n int) *Vec_Int_t { return Vec_IntStartRange(0, n) }
func Vec_IntReleaseNewArray(p *Vec_Int_t) []int32 {
	a := make([]int32, len(p.data)+1)
	a[0] = int32(len(a))
	copy(a[1:], Vec_IntArray(p))
	return a
}
func Vec_IntLimit(p *Vec_Int_t) []int32 { return p.storage()[len(p.data):] }
func Vec_IntDowndateEntry(p *Vec_Int_t, i int, x int32) {
	if Vec_IntEntry(p, i) > x {
		Vec_IntWriteEntry(p, i, x)
	}
}
func Vec_IntGrowResize(p *Vec_Int_t, n int) { Vec_IntGrow(p, n); p.resize(n) }
func Vec_IntFillTwo(p *Vec_Int_t, n int, even, odd int32) {
	Vec_IntGrowResize(p, n)
	for i := range n {
		if i&1 == 0 {
			p.storage()[i] = even
		} else {
			p.storage()[i] = odd
		}
	}
}
func Vec_IntFillNatural(p *Vec_Int_t, n int) {
	Vec_IntGrowResize(p, n)
	for i := range n {
		p.storage()[i] = int32(i)
	}
}
func Vec_IntGetEntry(p *Vec_Int_t, i int) int32 { Vec_IntFillExtra(p, i+1, 0); return p.storage()[i] }
func Vec_IntGetEntryFull(p *Vec_Int_t, i int) int32 {
	Vec_IntFillExtra(p, i+1, -1)
	return p.storage()[i]
}
func Vec_IntGetEntryP(p *Vec_Int_t, i int) *int32 {
	Vec_IntFillExtra(p, i+1, 0)
	return &p.storage()[i]
}
func Vec_IntSetEntry(p *Vec_Int_t, i int, x int32) { Vec_IntFillExtra(p, i+1, 0); p.storage()[i] = x }
func Vec_IntSetEntryFull(p *Vec_Int_t, i int, x int32) {
	Vec_IntFillExtra(p, i+1, -1)
	p.storage()[i] = x
}
func Vec_IntPushTwo(p *Vec_Int_t, a, b int32)      { Vec_IntPush(p, a); Vec_IntPush(p, b) }
func Vec_IntPushThree(p *Vec_Int_t, a, b, c int32) { Vec_IntPushTwo(p, a, b); Vec_IntPush(p, c) }
func Vec_IntPushFour(p *Vec_Int_t, a, b, c, d int32) {
	Vec_IntPushTwo(p, a, b)
	Vec_IntPushTwo(p, c, d)
}
func Vec_IntPushArray(p *Vec_Int_t, a []int32, n int) {
	for _, x := range a[:n] {
		Vec_IntPush(p, x)
	}
}
func Vec_IntPushFirst(p *Vec_Int_t, x int32) { Vec_IntInsert(p, 0, x) }
func Vec_IntPushOrderReverse(p *Vec_Int_t, x int32) {
	Vec_IntPush(p, x)
	i := len(p.data) - 1
	for i > 0 && p.storage()[i-1] < x {
		p.storage()[i] = p.storage()[i-1]
		i--
	}
	p.storage()[i] = x
}
func Vec_IntPushOrderCost(p *Vec_Int_t, x int32, cost *Vec_Int_t) {
	Vec_IntPush(p, x)
	i := len(p.data) - 1
	for i > 0 && Vec_IntEntry(cost, int(p.storage()[i-1])) > Vec_IntEntry(cost, int(x)) {
		p.storage()[i] = p.storage()[i-1]
		i--
	}
	p.storage()[i] = x
}
func Vec_IntPushUniqueOrder(p *Vec_Int_t, x int32) int {
	if Vec_IntFind(p, x) >= 0 {
		return 1
	}
	Vec_IntPushOrder(p, x)
	return 0
}
func Vec_IntPushUniqueOrderCost(p *Vec_Int_t, x int32, cost *Vec_Int_t) int {
	if Vec_IntFind(p, x) >= 0 {
		return 1
	}
	Vec_IntPushOrderCost(p, x, cost)
	return 0
}

// Fetch reserves existing storage. A failed reservation is atomic: it returns
// nil without corrupting the vector's logical length.
func Vec_IntFetch(p *Vec_Int_t, n int) []uint32 {
	if n == 0 {
		return nil
	}
	if n < 0 {
		panic("negative fetch")
	}
	start := len(p.data)
	if n > cap(p.data)-len(p.data) {
		return nil
	}
	p.resize(len(p.data) + n)
	return unsafe.Slice((*uint32)(unsafe.Pointer(&p.storage()[start])), n)
}
func Vec_IntRemove1(p *Vec_Int_t, x int32) int {
	for i := 1; i < len(p.data); i++ {
		if p.storage()[i] == x {
			Vec_IntDrop(p, i)
			return 1
		}
	}
	return 0
}
func Vec_IntDrop(p *Vec_Int_t, i int) {
	_ = p.data[i]
	copy(p.storage()[i:], p.storage()[i+1:len(p.data)])
	p.resize(len(p.data) - 1)
}
func Vec_IntInsert(p *Vec_Int_t, i int, x int32) {
	if i < 0 || i > len(p.data) {
		panic("insert index")
	}
	Vec_IntPush(p, 0)
	copy(p.storage()[i+1:len(p.data)], p.storage()[i:len(p.data)-1])
	p.storage()[i] = x
}
func Vec_IntReverseOrder(p *Vec_Int_t) {
	for i, j := 0, len(p.data)-1; i < j; i, j = i+1, j-1 {
		p.storage()[i], p.storage()[j] = p.storage()[j], p.storage()[i]
	}
}
func Vec_IntRemoveOdd(p *Vec_Int_t) {
	if len(p.data)&1 != 0 {
		panic("odd pair count")
	}
	p.resize(len(p.data) / (2))
	for i := range len(p.data) {
		p.storage()[i] = p.storage()[2*i]
	}
}
func Vec_IntRemoveEven(p *Vec_Int_t) {
	if len(p.data)&1 != 0 {
		panic("odd pair count")
	}
	p.resize(len(p.data) / (2))
	for i := range len(p.data) {
		p.storage()[i] = p.storage()[2*i+1]
	}
}
func Vec_IntInvert(p *Vec_Int_t, fill int32) *Vec_Int_t {
	q := Vec_IntAlloc(0)
	if len(p.data) == 0 {
		return q
	}
	Vec_IntFill(q, int(Vec_IntFindMax(p))+1, fill)
	for i, x := range Vec_IntArray(p) {
		if x != fill {
			q.storage()[int(x)] = int32(i)
		}
	}
	return q
}
func Vec_IntCondense(p *Vec_Int_t, fill int32) *Vec_Int_t {
	q := Vec_IntAlloc(len(p.data))
	for _, x := range Vec_IntArray(p) {
		if x != fill {
			Vec_IntPush(q, x)
		}
	}
	return q
}
func Vec_IntSum(p *Vec_Int_t) int32 {
	var sum int32
	for _, x := range Vec_IntArray(p) {
		sum += x
	}
	return sum
}
func Vec_IntCountEntry(p *Vec_Int_t, x int32) int {
	n := 0
	for _, v := range Vec_IntArray(p) {
		if v == x {
			n++
		}
	}
	return n
}
func Vec_IntCountLarger(p *Vec_Int_t, x int32) int {
	n := 0
	for _, v := range Vec_IntArray(p) {
		if v > x {
			n++
		}
	}
	return n
}
func Vec_IntCountSmaller(p *Vec_Int_t, x int32) int {
	n := 0
	for _, v := range Vec_IntArray(p) {
		if v < x {
			n++
		}
	}
	return n
}
func Vec_IntCountPositive(p *Vec_Int_t) int { return Vec_IntCountLarger(p, 0) }
func Vec_IntCountZero(p *Vec_Int_t) int     { return Vec_IntCountEntry(p, 0) }
func Vec_IntAddPositive(p *Vec_Int_t) int32 {
	var sum int32
	for _, x := range Vec_IntArray(p) {
		if x > 0 {
			sum += x
		}
	}
	return sum
}
func Vec_IntCountCommon(p, q *Vec_Int_t) int {
	present := map[int32]bool{}
	if len(p.data) < len(q.data) {
		p, q = q, p
	}
	for _, x := range Vec_IntArray(q) {
		present[x] = true
	}
	n := 0
	for _, x := range Vec_IntArray(p) {
		if present[x] {
			n++
		}
	}
	return n
}
func Vec_IntSortMulti(p *Vec_Int_t, n, reverse int) {
	if n <= 0 || len(p.data)%n != 0 {
		panic("invalid record size")
	}
	records := make([][]int32, len(p.data)/n)
	for i := range records {
		records[i] = append([]int32(nil), p.storage()[n*i:n*(i+1)]...)
	}
	sort.SliceStable(records, func(i, j int) bool {
		if reverse != 0 {
			return records[i][0] > records[j][0]
		}
		return records[i][0] < records[j][0]
	})
	for i, r := range records {
		copy(p.storage()[n*i:], r)
	}
}
func Vec_IntIsSorted(p *Vec_Int_t, reverse int) int {
	for i := 1; i < len(p.data); i++ {
		if reverse != 0 && p.storage()[i-1] < p.storage()[i] || reverse == 0 && p.storage()[i-1] > p.storage()[i] {
			return 0
		}
	}
	return 1
}
func Vec_IntUniqify(p *Vec_Int_t) int {
	if len(p.data) < 2 {
		return 0
	}
	Vec_IntSort(p, 0)
	k := 1
	for i := 1; i < len(p.data); i++ {
		if p.storage()[i] != p.storage()[i-1] {
			p.storage()[k] = p.storage()[i]
			k++
		}
	}
	removed := len(p.data) - k
	p.resize(k)
	return removed
}
func Vec_IntCountDuplicates(p *Vec_Int_t) int { q := Vec_IntDup(p); return Vec_IntUniqify(q) }
func Vec_IntCheckUniqueSmall(p *Vec_Int_t) int {
	if Vec_IntCountDuplicates(p) != 0 {
		return 0
	}
	return 1
}
func Vec_IntCountUnique(p *Vec_Int_t) int { return len(p.data) - Vec_IntCountDuplicates(p) }
func Vec_IntUniqifyPairs(p *Vec_Int_t) int {
	if len(p.data)%2 != 0 {
		panic("odd pairs")
	}
	if len(p.data) < 4 {
		return 0
	}
	Vec_IntSortMulti(p, 2, 0)
	k := 1
	for i := 1; i < len(p.data)/2; i++ {
		if p.storage()[2*i] != p.storage()[2*i-2] || p.storage()[2*i+1] != p.storage()[2*i-1] {
			copy(p.storage()[2*k:], p.storage()[2*i:2*i+2])
			k++
		}
	}
	removed := len(p.data)/2 - k
	p.resize(2 * k)
	return removed
}
func Vec_IntUniqueHashKey2(a []byte, n int) uint32 {
	primes := [4]uint32{12582917, 25165843, 50331653, 100663319}
	var key uint32
	for i, x := range a[:n] {
		key += uint32(x) * primes[i&3]
	}
	return key
}
func Vec_IntUniqueHashKeyDebug(a []byte, n, mask int) uint32 {
	primes := [4]uint32{12582917, 25165843, 50331653, 100663319}
	var key uint32
	for i, x := range a[:n] {
		key += uint32(x) * primes[i&3]
		fmt.Printf("%d : %3d  %12d %12d \n", i, x, key, key&uint32(mask))
	}
	return key
}
func Vec_IntUniqueHashKey(a []byte, n int) uint32 {
	p := [16]uint32{0x984b6ad9, 0x18a6eed3, 0x950353e2, 0x6222f6eb, 0xdfbedd47, 0xef0f9023, 0xac932a26, 0x590eaf55, 0x97d0a034, 0xdc36cd2e, 0x22736b37, 0xdc9066b0, 0x2eb2f98b, 0x5d9c7baf, 0x85747c9e, 0x8aca1055}
	q := [16]uint32{0x8d8a5ebe, 0x1e6a15dc, 0x197d49db, 0x5bab9c89, 0x4b55dea7, 0x55dede49, 0x9a6a8080, 0xe5e51035, 0xe148d658, 0x8a17eb3b, 0xe22e4b38, 0xe5be2a9a, 0xbe938cbb, 0x3b981069, 0x7f9c0c8e, 0xf756df10}
	var key uint32
	for i, x := range a[:n] {
		key += q[(2*i)&15]*p[x&15] + q[(2*i+1)&15]*p[x>>4]
	}
	return key
}
func Vec_IntUniqueProfile(data *Vec_Int_t, table, next []int32, mask, size int) {
	for _, head := range table[:mask+1] {
		n := 0
		for head != -1 {
			n++
			head = next[head]
		}
		if n >= 7 {
			fmt.Println(n)
		}
	}
	fmt.Println()
}
func Vec_IntUniqueLookup(data *Vec_Int_t, i, size int, next []int32, start *int32) *int32 {
	for *start != -1 {
		equal := true
		for j := range size {
			if data.storage()[i*size+j] != data.storage()[int(*start)*size+j] {
				equal = false
				break
			}
		}
		if equal {
			return start
		}
		start = &next[*start]
	}
	return start
}
func intRecordKey(a []int32) string {
	b := make([]byte, 4*len(a))
	for i, x := range a {
		binary.LittleEndian.PutUint32(b[4*i:], uint32(x))
	}
	return string(b)
}
func Vec_IntUniqueCount(data *Vec_Int_t, size int, mapping **Vec_Int_t) int {
	if size <= 0 || len(data.data)%size != 0 {
		panic("invalid record size")
	}
	seen := map[string]int32{}
	m := Vec_IntAlloc(len(data.data) / size)
	for i := 0; i < len(data.data)/size; i++ {
		key := intRecordKey(data.storage()[i*size : (i+1)*size])
		representative, found := seen[key]
		if !found {
			representative = int32(i)
			seen[key] = representative
		}
		Vec_IntPush(m, representative)
	}
	if mapping != nil {
		*mapping = m
	}
	return len(seen)
}
func Vec_IntUniqifyHash(data *Vec_Int_t, size int) *Vec_Int_t {
	var m *Vec_Int_t
	n := Vec_IntUniqueCount(data, size, &m)
	q := Vec_IntAlloc(n * size)
	for i, x := range Vec_IntArray(m) {
		if x == int32(i) {
			Vec_IntPushArray(q, data.storage()[i*size:], size)
		}
	}
	return q
}
func Vec_IntSortCompareUnsigned(a, b *uint32) int {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func Vec_IntSortUnsigned(p *Vec_Int_t) {
	sort.Slice(Vec_IntArray(p), func(i, j int) bool { return uint32(p.storage()[i]) < uint32(p.storage()[j]) })
}
func Vec_IntTwoCountCommon(a, b *Vec_Int_t) int {
	q := Vec_IntAlloc(0)
	return Vec_IntTwoFindCommon(a, b, q)
}
func Vec_IntTwoFindCommon(a, b, q *Vec_Int_t) int        { return intFindCommon(a, b, q, false) }
func Vec_IntTwoFindCommonReverse(a, b, q *Vec_Int_t) int { return intFindCommon(a, b, q, true) }
func intFindCommon(a, b, q *Vec_Int_t, reverse bool) int {
	Vec_IntClear(q)
	i, j := 0, 0
	for i < len(a.data) && j < len(b.data) {
		av, bv := a.storage()[i], b.storage()[j]
		if av == bv {
			Vec_IntPush(q, av)
			i++
			j++
		} else if !reverse && av < bv || reverse && av > bv {
			i++
		} else {
			j++
		}
	}
	return len(q.data)
}
func Vec_IntTwoSplit(a, b, common, an, bn *Vec_Int_t) {
	i, j := 0, 0
	for i < len(a.data) && j < len(b.data) {
		av, bv := a.storage()[i], b.storage()[j]
		if av == bv {
			Vec_IntPush(common, av)
			i++
			j++
		} else if av < bv {
			Vec_IntPush(an, av)
			i++
		} else {
			Vec_IntPush(bn, bv)
			j++
		}
	}
	Vec_IntPushArray(an, a.storage()[i:], len(a.data)-i)
	Vec_IntPushArray(bn, b.storage()[j:], len(b.data)-j)
}
func Vec_IntTwoRemoveCommon(a, b, common *Vec_Int_t) int {
	an, bn := Vec_IntAlloc(0), Vec_IntAlloc(0)
	Vec_IntClear(common)
	Vec_IntTwoSplit(a, b, common, an, bn)
	Vec_IntFill(a, len(an.data), 0)
	copy(a.storage(), Vec_IntArray(an))
	Vec_IntFill(b, len(bn.data), 0)
	copy(b.storage(), Vec_IntArray(bn))
	return len(common.data)
}
func Vec_IntTwoRemove(a, b *Vec_Int_t) int {
	common, an, bn := Vec_IntAlloc(0), Vec_IntAlloc(0), Vec_IntAlloc(0)
	Vec_IntTwoSplit(a, b, common, an, bn)
	Vec_IntFill(a, len(an.data), 0)
	copy(a.storage(), Vec_IntArray(an))
	return len(a.data)
}
func Vec_IntTwoMerge2Int(a, b, q *Vec_Int_t) {
	i, j, k := 0, 0, 0
	for i < len(a.data) && j < len(b.data) {
		av, bv := a.storage()[i], b.storage()[j]
		if av == bv {
			q.storage()[k] = av
			i++
			j++
		} else if av < bv {
			q.storage()[k] = av
			i++
		} else {
			q.storage()[k] = bv
			j++
		}
		k++
	}
	k += copy(q.storage()[k:], a.storage()[i:len(a.data)])
	k += copy(q.storage()[k:], b.storage()[j:len(b.data)])
	q.resize(k)
}
func Vec_IntTwoMerge(a, b *Vec_Int_t) *Vec_Int_t {
	q := Vec_IntAlloc(len(a.data) + len(b.data))
	Vec_IntTwoMerge2Int(a, b, q)
	return q
}
func Vec_IntTwoMerge2(a, b, q *Vec_Int_t) {
	Vec_IntGrow(q, len(a.data)+len(b.data))
	Vec_IntTwoMerge2Int(a, b, q)
}
func Vec_IntSelectSort(a []int32, n int) {
	sort.Slice(a[:n], func(i, j int) bool { return a[i] < a[j] })
}
func Vec_IntSelectSortReverse(a []int32, n int) {
	sort.Slice(a[:n], func(i, j int) bool { return a[i] > a[j] })
}
func Vec_IntSelectSortCost(a []int32, n int, cost *Vec_Int_t) {
	sort.SliceStable(a[:n], func(i, j int) bool { return cost.storage()[a[i]] < cost.storage()[a[j]] })
}
func Vec_IntSelectSortCostReverse(a []int32, n int, cost *Vec_Int_t) {
	sort.SliceStable(a[:n], func(i, j int) bool { return cost.storage()[a[i]] > cost.storage()[a[j]] })
}
func Vec_IntSelectSortCost2(a []int32, n int, cost []int32)        { intSortCost2(a, n, cost, false) }
func Vec_IntSelectSortCost2Reverse(a []int32, n int, cost []int32) { intSortCost2(a, n, cost, true) }
func intSortCost2(a []int32, n int, cost []int32, reverse bool) {
	for i := 0; i < n-1; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if !reverse && cost[j] < cost[best] || reverse && cost[j] > cost[best] {
				best = j
			}
		}
		a[i], a[best] = a[best], a[i]
		cost[i], cost[best] = cost[best], cost[i]
	}
}
func Vec_IntPrintBinary(p *Vec_Int_t) {
	for _, x := range Vec_IntArray(p) {
		if x == 0 {
			fmt.Print(0)
		} else {
			fmt.Print(1)
		}
	}
}
func Vec_IntCompareVec(p, q *Vec_Int_t) int {
	if p == nil {
		if q == nil {
			return 0
		}
		return -1
	}
	if q == nil {
		return 1
	}
	if len(p.data) != len(q.data) {
		return len(p.data) - len(q.data)
	}
	return bytes.Compare([]byte(intRecordKey(Vec_IntArray(p))), []byte(intRecordKey(Vec_IntArray(q))))
}
func Vec_IntAppend(p, q *Vec_Int_t) { Vec_IntPushArray(p, Vec_IntArray(q), len(q.data)) }
func Vec_IntAppendSkip(p, q *Vec_Int_t, skip int) {
	for i, x := range Vec_IntArray(q) {
		if i != skip {
			Vec_IntPush(p, x)
		}
	}
}
func Vec_IntAppendMinus(p, q *Vec_Int_t, minus int) {
	Vec_IntClear(p)
	for _, x := range Vec_IntArray(q) {
		if minus != 0 {
			x = -x
		}
		Vec_IntPush(p, x)
	}
}
func Vec_IntRemapArray(mapping, old, new *Vec_Int_t, n int) {
	if len(old.data) == 0 {
		return
	}
	Vec_IntFill(new, n, 0)
	for i, j := range Vec_IntArray(mapping) {
		if j > 0 && int(j) < n && i < len(old.data) && old.storage()[i] != 0 {
			new.storage()[j] = old.storage()[i]
		}
	}
}
