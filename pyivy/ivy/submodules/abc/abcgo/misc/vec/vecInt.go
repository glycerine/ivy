package vec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"unsafe"
)

// Vec_Int_t retains ABC's logical size and allocated capacity separately.
// pArray contains capacity elements; Array exposes only the live prefix.
type Vec_Int_t struct {
	marked      bool
	nCap, nSize int
	pArray      []int32
}

func Vec_IntAlloc(n int) *Vec_Int_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_IntAllocExact(n)
}
func Vec_IntAllocExact(n int) *Vec_Int_t { return &Vec_Int_t{nCap: n, pArray: make([]int32, n)} }
func Vec_IntStart(n int) *Vec_Int_t      { p := Vec_IntAlloc(n); p.nSize = n; return p }

// StartFull matches C's bytewise 0xff initialization, setting every integer to -1.
func Vec_IntStartFull(n int) *Vec_Int_t {
	p := Vec_IntStart(n)
	for i := 0; i < n; i++ {
		p.pArray[i] = -1
	}
	return p
}
func Vec_IntAllocArray(a []int32, n int) *Vec_Int_t {
	return &Vec_Int_t{nCap: n, nSize: n, pArray: a[:n]}
}
func Vec_IntAllocArrayCopy(a []int32, n int) *Vec_Int_t {
	p := Vec_IntAllocExact(n)
	p.nSize = n
	copy(p.pArray, a[:n])
	return p
}
func Vec_IntDup(p *Vec_Int_t) *Vec_Int_t {
	q := Vec_IntAllocExact(p.nCap)
	q.nSize = p.nSize
	copy(q.pArray, p.pArray[:p.nSize])
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
func Vec_IntReleaseArray(p *Vec_Int_t) []int32 { a := p.pArray; Vec_IntZero(p); return a }
func Vec_IntArray(p *Vec_Int_t) []int32        { return p.pArray[:p.nSize] }
func Vec_IntArrayP(p *Vec_Int_t) *[]int32      { return &p.pArray }
func Vec_IntSize(p *Vec_Int_t) int             { return p.nSize }
func Vec_IntCap(p *Vec_Int_t) int              { return p.nCap }
func Vec_IntMemory(p *Vec_Int_t) float64 {
	if p == nil {
		return 0
	}
	return float64(4*p.nCap + int(unsafe.Sizeof(*p)))
}
func Vec_IntEntry(p *Vec_Int_t, i int) int32         { return p.pArray[:p.nSize][i] }
func Vec_IntEntryP(p *Vec_Int_t, i int) *int32       { return &p.pArray[:p.nSize][i] }
func Vec_IntWriteEntry(p *Vec_Int_t, i int, x int32) { p.pArray[:p.nSize][i] = x }
func Vec_IntAddToEntry(p *Vec_Int_t, i int, x int32) int32 {
	p.pArray[:p.nSize][i] += x
	return p.pArray[i]
}
func Vec_IntUpdateEntry(p *Vec_Int_t, i int, x int32) {
	if Vec_IntEntry(p, i) < x {
		Vec_IntWriteEntry(p, i, x)
	}
}
func Vec_IntEntryLast(p *Vec_Int_t) int32 { return Vec_IntEntry(p, p.nSize-1) }
func Vec_IntGrow(p *Vec_Int_t, n int) {
	if n > p.nCap {
		p.pArray = append(p.pArray, make([]int32, n-p.nCap)...)
		p.nCap = n
	}
}
func Vec_IntFill(p *Vec_Int_t, n int, x int32) {
	Vec_IntGrow(p, n)
	for i := 0; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_IntFillExtra(p *Vec_Int_t, n int, x int32) {
	if n <= p.nSize {
		return
	}
	if n > p.nCap {
		Vec_IntGrow(p, max(n, 2*p.nCap))
	}
	for i := p.nSize; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_IntShrink(p *Vec_Int_t, n int) {
	if n < 0 || n > p.nSize {
		panic("invalid vector shrink")
	}
	p.nSize = n
}
func Vec_IntClear(p *Vec_Int_t) { p.nSize = 0 }
func Vec_IntPush(p *Vec_Int_t, x int32) {
	if p.nSize == p.nCap {
		Vec_IntGrow(p, max(16, 2*p.nCap))
	}
	p.pArray[p.nSize] = x
	p.nSize++
}
func Vec_IntPushOrder(p *Vec_Int_t, x int32) {
	Vec_IntPush(p, x)
	i := p.nSize - 1
	for i > 0 && p.pArray[i-1] > x {
		p.pArray[i] = p.pArray[i-1]
		i--
	}
	p.pArray[i] = x
}
func Vec_IntPushUnique(p *Vec_Int_t, x int32) int {
	if Vec_IntFind(p, x) >= 0 {
		return 1
	}
	Vec_IntPush(p, x)
	return 0
}
func Vec_IntPop(p *Vec_Int_t) int32 { x := Vec_IntEntryLast(p); p.nSize--; return x }
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
	copy(p.pArray[i:], p.pArray[i+1:p.nSize])
	p.nSize--
	return 1
}
func Vec_IntFindMax(p *Vec_Int_t) int32 {
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
func Vec_IntFindMin(p *Vec_Int_t) int32 {
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
func Vec_IntEqual(p, q *Vec_Int_t) int {
	if p.nSize != q.nSize {
		return 0
	}
	for i, v := range Vec_IntArray(p) {
		if v != q.pArray[i] {
			return 0
		}
	}
	return 1
}
func Vec_IntPrint(p *Vec_Int_t) {
	fmt.Printf("Vector has %d entries: {", p.nSize)
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
	sort.Slice(p.pArray[:p.nSize], func(i, j int) bool {
		if reverse != 0 {
			return p.pArray[i] > p.pArray[j]
		}
		return p.pArray[i] < p.pArray[j]
	})
}

func Vec_IntStartRange(first int32, n int) *Vec_Int_t {
	p := Vec_IntStart(n)
	for i := range n {
		p.pArray[i] = first + int32(i)
	}
	return p
}
func Vec_IntStartNatural(n int) *Vec_Int_t { return Vec_IntStartRange(0, n) }
func Vec_IntReleaseNewArray(p *Vec_Int_t) []int32 {
	a := make([]int32, p.nSize+1)
	a[0] = int32(len(a))
	copy(a[1:], Vec_IntArray(p))
	return a
}
func Vec_IntLimit(p *Vec_Int_t) []int32 { return p.pArray[p.nSize:] }
func Vec_IntDowndateEntry(p *Vec_Int_t, i int, x int32) {
	if Vec_IntEntry(p, i) > x {
		Vec_IntWriteEntry(p, i, x)
	}
}
func Vec_IntGrowResize(p *Vec_Int_t, n int) { Vec_IntGrow(p, n); p.nSize = n }
func Vec_IntFillTwo(p *Vec_Int_t, n int, even, odd int32) {
	Vec_IntGrowResize(p, n)
	for i := range n {
		if i&1 == 0 {
			p.pArray[i] = even
		} else {
			p.pArray[i] = odd
		}
	}
}
func Vec_IntFillNatural(p *Vec_Int_t, n int) {
	Vec_IntGrowResize(p, n)
	for i := range n {
		p.pArray[i] = int32(i)
	}
}
func Vec_IntGetEntry(p *Vec_Int_t, i int) int32        { Vec_IntFillExtra(p, i+1, 0); return p.pArray[i] }
func Vec_IntGetEntryFull(p *Vec_Int_t, i int) int32    { Vec_IntFillExtra(p, i+1, -1); return p.pArray[i] }
func Vec_IntGetEntryP(p *Vec_Int_t, i int) *int32      { Vec_IntFillExtra(p, i+1, 0); return &p.pArray[i] }
func Vec_IntSetEntry(p *Vec_Int_t, i int, x int32)     { Vec_IntFillExtra(p, i+1, 0); p.pArray[i] = x }
func Vec_IntSetEntryFull(p *Vec_Int_t, i int, x int32) { Vec_IntFillExtra(p, i+1, -1); p.pArray[i] = x }
func Vec_IntPushTwo(p *Vec_Int_t, a, b int32)          { Vec_IntPush(p, a); Vec_IntPush(p, b) }
func Vec_IntPushThree(p *Vec_Int_t, a, b, c int32)     { Vec_IntPushTwo(p, a, b); Vec_IntPush(p, c) }
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
	i := p.nSize - 1
	for i > 0 && p.pArray[i-1] < x {
		p.pArray[i] = p.pArray[i-1]
		i--
	}
	p.pArray[i] = x
}
func Vec_IntPushOrderCost(p *Vec_Int_t, x int32, cost *Vec_Int_t) {
	Vec_IntPush(p, x)
	i := p.nSize - 1
	for i > 0 && Vec_IntEntry(cost, int(p.pArray[i-1])) > Vec_IntEntry(cost, int(x)) {
		p.pArray[i] = p.pArray[i-1]
		i--
	}
	p.pArray[i] = x
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

// Fetch preserves the C API's reservation semantics: size advances even when
// the reservation exceeds existing capacity and nil is returned.
func Vec_IntFetch(p *Vec_Int_t, n int) []uint32 {
	if n == 0 {
		return nil
	}
	if n < 0 {
		panic("negative fetch")
	}
	start := p.nSize
	p.nSize += n
	if p.nSize > p.nCap {
		return nil
	}
	return unsafe.Slice((*uint32)(unsafe.Pointer(&p.pArray[start])), n)
}
func Vec_IntRemove1(p *Vec_Int_t, x int32) int {
	for i := 1; i < p.nSize; i++ {
		if p.pArray[i] == x {
			Vec_IntDrop(p, i)
			return 1
		}
	}
	return 0
}
func Vec_IntDrop(p *Vec_Int_t, i int) {
	_ = p.pArray[:p.nSize][i]
	copy(p.pArray[i:], p.pArray[i+1:p.nSize])
	p.nSize--
}
func Vec_IntInsert(p *Vec_Int_t, i int, x int32) {
	if i < 0 || i > p.nSize {
		panic("insert index")
	}
	Vec_IntPush(p, 0)
	copy(p.pArray[i+1:p.nSize], p.pArray[i:p.nSize-1])
	p.pArray[i] = x
}
func Vec_IntReverseOrder(p *Vec_Int_t) {
	for i, j := 0, p.nSize-1; i < j; i, j = i+1, j-1 {
		p.pArray[i], p.pArray[j] = p.pArray[j], p.pArray[i]
	}
}
func Vec_IntRemoveOdd(p *Vec_Int_t) {
	if p.nSize&1 != 0 {
		panic("odd pair count")
	}
	p.nSize /= 2
	for i := range p.nSize {
		p.pArray[i] = p.pArray[2*i]
	}
}
func Vec_IntRemoveEven(p *Vec_Int_t) {
	if p.nSize&1 != 0 {
		panic("odd pair count")
	}
	p.nSize /= 2
	for i := range p.nSize {
		p.pArray[i] = p.pArray[2*i+1]
	}
}
func Vec_IntInvert(p *Vec_Int_t, fill int32) *Vec_Int_t {
	q := Vec_IntAlloc(0)
	if p.nSize == 0 {
		return q
	}
	Vec_IntFill(q, int(Vec_IntFindMax(p))+1, fill)
	for i, x := range Vec_IntArray(p) {
		if x != fill {
			q.pArray[int(x)] = int32(i)
		}
	}
	return q
}
func Vec_IntCondense(p *Vec_Int_t, fill int32) *Vec_Int_t {
	q := Vec_IntAlloc(p.nSize)
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
	if p.nSize < q.nSize {
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
	if n <= 0 || p.nSize%n != 0 {
		panic("invalid record size")
	}
	records := make([][]int32, p.nSize/n)
	for i := range records {
		records[i] = append([]int32(nil), p.pArray[n*i:n*(i+1)]...)
	}
	sort.SliceStable(records, func(i, j int) bool {
		if reverse != 0 {
			return records[i][0] > records[j][0]
		}
		return records[i][0] < records[j][0]
	})
	for i, r := range records {
		copy(p.pArray[n*i:], r)
	}
}
func Vec_IntIsSorted(p *Vec_Int_t, reverse int) int {
	for i := 1; i < p.nSize; i++ {
		if reverse != 0 && p.pArray[i-1] < p.pArray[i] || reverse == 0 && p.pArray[i-1] > p.pArray[i] {
			return 0
		}
	}
	return 1
}
func Vec_IntUniqify(p *Vec_Int_t) int {
	if p.nSize < 2 {
		return 0
	}
	Vec_IntSort(p, 0)
	k := 1
	for i := 1; i < p.nSize; i++ {
		if p.pArray[i] != p.pArray[i-1] {
			p.pArray[k] = p.pArray[i]
			k++
		}
	}
	removed := p.nSize - k
	p.nSize = k
	return removed
}
func Vec_IntCountDuplicates(p *Vec_Int_t) int { q := Vec_IntDup(p); return Vec_IntUniqify(q) }
func Vec_IntCheckUniqueSmall(p *Vec_Int_t) int {
	if Vec_IntCountDuplicates(p) != 0 {
		return 0
	}
	return 1
}
func Vec_IntCountUnique(p *Vec_Int_t) int { return p.nSize - Vec_IntCountDuplicates(p) }
func Vec_IntUniqifyPairs(p *Vec_Int_t) int {
	if p.nSize%2 != 0 {
		panic("odd pairs")
	}
	if p.nSize < 4 {
		return 0
	}
	Vec_IntSortMulti(p, 2, 0)
	k := 1
	for i := 1; i < p.nSize/2; i++ {
		if p.pArray[2*i] != p.pArray[2*i-2] || p.pArray[2*i+1] != p.pArray[2*i-1] {
			copy(p.pArray[2*k:], p.pArray[2*i:2*i+2])
			k++
		}
	}
	removed := p.nSize/2 - k
	p.nSize = 2 * k
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
			if data.pArray[i*size+j] != data.pArray[int(*start)*size+j] {
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
	if size <= 0 || data.nSize%size != 0 {
		panic("invalid record size")
	}
	seen := map[string]int32{}
	m := Vec_IntAlloc(data.nSize / size)
	for i := 0; i < data.nSize/size; i++ {
		key := intRecordKey(data.pArray[i*size : (i+1)*size])
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
			Vec_IntPushArray(q, data.pArray[i*size:], size)
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
	sort.Slice(Vec_IntArray(p), func(i, j int) bool { return uint32(p.pArray[i]) < uint32(p.pArray[j]) })
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
	for i < a.nSize && j < b.nSize {
		av, bv := a.pArray[i], b.pArray[j]
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
	return q.nSize
}
func Vec_IntTwoSplit(a, b, common, an, bn *Vec_Int_t) {
	i, j := 0, 0
	for i < a.nSize && j < b.nSize {
		av, bv := a.pArray[i], b.pArray[j]
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
	Vec_IntPushArray(an, a.pArray[i:], a.nSize-i)
	Vec_IntPushArray(bn, b.pArray[j:], b.nSize-j)
}
func Vec_IntTwoRemoveCommon(a, b, common *Vec_Int_t) int {
	an, bn := Vec_IntAlloc(0), Vec_IntAlloc(0)
	Vec_IntClear(common)
	Vec_IntTwoSplit(a, b, common, an, bn)
	Vec_IntFill(a, an.nSize, 0)
	copy(a.pArray, Vec_IntArray(an))
	Vec_IntFill(b, bn.nSize, 0)
	copy(b.pArray, Vec_IntArray(bn))
	return common.nSize
}
func Vec_IntTwoRemove(a, b *Vec_Int_t) int {
	common, an, bn := Vec_IntAlloc(0), Vec_IntAlloc(0), Vec_IntAlloc(0)
	Vec_IntTwoSplit(a, b, common, an, bn)
	Vec_IntFill(a, an.nSize, 0)
	copy(a.pArray, Vec_IntArray(an))
	return a.nSize
}
func Vec_IntTwoMerge2Int(a, b, q *Vec_Int_t) {
	i, j, k := 0, 0, 0
	for i < a.nSize && j < b.nSize {
		av, bv := a.pArray[i], b.pArray[j]
		if av == bv {
			q.pArray[k] = av
			i++
			j++
		} else if av < bv {
			q.pArray[k] = av
			i++
		} else {
			q.pArray[k] = bv
			j++
		}
		k++
	}
	k += copy(q.pArray[k:], a.pArray[i:a.nSize])
	k += copy(q.pArray[k:], b.pArray[j:b.nSize])
	q.nSize = k
}
func Vec_IntTwoMerge(a, b *Vec_Int_t) *Vec_Int_t {
	q := Vec_IntAlloc(a.nSize + b.nSize)
	Vec_IntTwoMerge2Int(a, b, q)
	return q
}
func Vec_IntTwoMerge2(a, b, q *Vec_Int_t) {
	Vec_IntGrow(q, a.nSize+b.nSize)
	Vec_IntTwoMerge2Int(a, b, q)
}
func Vec_IntSelectSort(a []int32, n int) {
	sort.Slice(a[:n], func(i, j int) bool { return a[i] < a[j] })
}
func Vec_IntSelectSortReverse(a []int32, n int) {
	sort.Slice(a[:n], func(i, j int) bool { return a[i] > a[j] })
}
func Vec_IntSelectSortCost(a []int32, n int, cost *Vec_Int_t) {
	sort.SliceStable(a[:n], func(i, j int) bool { return cost.pArray[a[i]] < cost.pArray[a[j]] })
}
func Vec_IntSelectSortCostReverse(a []int32, n int, cost *Vec_Int_t) {
	sort.SliceStable(a[:n], func(i, j int) bool { return cost.pArray[a[i]] > cost.pArray[a[j]] })
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
	if p.nSize != q.nSize {
		return p.nSize - q.nSize
	}
	return bytes.Compare([]byte(intRecordKey(Vec_IntArray(p))), []byte(intRecordKey(Vec_IntArray(q))))
}
func Vec_IntAppend(p, q *Vec_Int_t) { Vec_IntPushArray(p, Vec_IntArray(q), q.nSize) }
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
	if old.nSize == 0 {
		return
	}
	Vec_IntFill(new, n, 0)
	for i, j := range Vec_IntArray(mapping) {
		if j > 0 && int(j) < n && i < old.nSize && old.pArray[i] != 0 {
			new.pArray[j] = old.pArray[i]
		}
	}
}
