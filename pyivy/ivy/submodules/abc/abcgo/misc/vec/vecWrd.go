package vec

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"sort"
	"unsafe"
)

type Vec_Wrd_t struct {
	nCap, nSize int
	pArray      []uint64
}

func Vec_WrdAlloc(n int) *Vec_Wrd_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_WrdAllocExact(n)
}
func Vec_WrdAllocExact(n int) *Vec_Wrd_t { return &Vec_Wrd_t{nCap: n, pArray: make([]uint64, n)} }
func Vec_WrdStart(n int) *Vec_Wrd_t      { p := Vec_WrdAlloc(n); p.nSize = n; return p }
func Vec_WrdStartFull(n int) *Vec_Wrd_t {
	p := Vec_WrdStart(n)
	for i := 0; i < n; i++ {
		p.pArray[i] = ^uint64(0)
	}
	return p
}
func Vec_WrdAllocArray(a []uint64, n int) *Vec_Wrd_t {
	return &Vec_Wrd_t{nCap: n, nSize: n, pArray: a[:n]}
}
func Vec_WrdAllocArrayCopy(a []uint64, n int) *Vec_Wrd_t {
	p := Vec_WrdAllocExact(n)
	p.nSize = n
	copy(p.pArray, a[:n])
	return p
}
func Vec_WrdDup(p *Vec_Wrd_t) *Vec_Wrd_t {
	q := Vec_WrdAllocExact(p.nCap)
	q.nSize = p.nSize
	copy(q.pArray, p.pArray[:p.nSize])
	return q
}
func Vec_WrdDupArray(p *Vec_Wrd_t) *Vec_Wrd_t { q := *p; Vec_WrdZero(p); return &q }
func Vec_WrdErase(p *Vec_Wrd_t)               { Vec_WrdZero(p) }
func Vec_WrdFree(p *Vec_Wrd_t) {
	if p != nil {
		Vec_WrdZero(p)
	}
}
func Vec_WrdFreeP(p **Vec_Wrd_t)                { Vec_WrdFree(*p); *p = nil }
func Vec_WrdReleaseArray(p *Vec_Wrd_t) []uint64 { a := p.pArray; Vec_WrdZero(p); return a }
func Vec_WrdArray(p *Vec_Wrd_t) []uint64        { return p.pArray[:p.nSize] }
func Vec_WrdSize(p *Vec_Wrd_t) int              { return p.nSize }
func Vec_WrdCap(p *Vec_Wrd_t) int               { return p.nCap }
func Vec_WrdMemory(p *Vec_Wrd_t) float64 {
	if p == nil {
		return 0
	}
	return float64(8*p.nCap + int(unsafe.Sizeof(*p)))
}
func Vec_WrdEntry(p *Vec_Wrd_t, i int) uint64         { return p.pArray[:p.nSize][i] }
func Vec_WrdEntryP(p *Vec_Wrd_t, i int) *uint64       { return &p.pArray[:p.nSize][i] }
func Vec_WrdWriteEntry(p *Vec_Wrd_t, i int, x uint64) { p.pArray[:p.nSize][i] = x }
func Vec_WrdAddToEntry(p *Vec_Wrd_t, i int, x uint64) uint64 {
	p.pArray[:p.nSize][i] += x
	return p.pArray[i]
}
func Vec_WrdEntryLast(p *Vec_Wrd_t) uint64 { return Vec_WrdEntry(p, p.nSize-1) }
func Vec_WrdGrow(p *Vec_Wrd_t, n int) {
	if n > p.nCap {
		p.pArray = append(p.pArray, make([]uint64, n-p.nCap)...)
		p.nCap = n
	}
}
func Vec_WrdFill(p *Vec_Wrd_t, n int, x uint64) {
	Vec_WrdGrow(p, n)
	for i := 0; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_WrdFillExtra(p *Vec_Wrd_t, n int, x uint64) {
	if n <= p.nSize {
		return
	}
	if n > p.nCap {
		Vec_WrdGrow(p, max(n, 2*p.nCap))
	}
	for i := p.nSize; i < n; i++ {
		p.pArray[i] = x
	}
	p.nSize = n
}
func Vec_WrdShrink(p *Vec_Wrd_t, n int) {
	if n < 0 || n > p.nSize {
		panic("invalid vector shrink")
	}
	p.nSize = n
}
func Vec_WrdClear(p *Vec_Wrd_t) { p.nSize = 0 }
func Vec_WrdPush(p *Vec_Wrd_t, x uint64) {
	if p.nSize == p.nCap {
		Vec_WrdGrow(p, max(16, 2*p.nCap))
	}
	p.pArray[p.nSize] = x
	p.nSize++
}
func Vec_WrdPushOrder(p *Vec_Wrd_t, x uint64) {
	Vec_WrdPush(p, x)
	i := p.nSize - 1
	for i > 0 && p.pArray[i-1] > x {
		p.pArray[i] = p.pArray[i-1]
		i--
	}
	p.pArray[i] = x
}
func Vec_WrdPushUnique(p *Vec_Wrd_t, x uint64) int {
	if Vec_WrdFind(p, x) >= 0 {
		return 1
	}
	Vec_WrdPush(p, x)
	return 0
}
func Vec_WrdPop(p *Vec_Wrd_t) uint64 { x := Vec_WrdEntryLast(p); p.nSize--; return x }
func Vec_WrdFind(p *Vec_Wrd_t, x uint64) int {
	for i, v := range Vec_WrdArray(p) {
		if v == x {
			return i
		}
	}
	return -1
}
func Vec_WrdRemove(p *Vec_Wrd_t, x uint64) int {
	i := Vec_WrdFind(p, x)
	if i < 0 {
		return 0
	}
	copy(p.pArray[i:], p.pArray[i+1:p.nSize])
	p.nSize--
	return 1
}
func Vec_WrdFindMax(p *Vec_Wrd_t) uint64 {
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
func Vec_WrdFindMin(p *Vec_Wrd_t) uint64 {
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
func Vec_WrdEqual(p, q *Vec_Wrd_t) int {
	if p.nSize != q.nSize {
		return 0
	}
	for i, v := range Vec_WrdArray(p) {
		if v != q.pArray[i] {
			return 0
		}
	}
	return 1
}
func Vec_WrdSortCompare1(a, b *uint64) int {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func Vec_WrdSortCompare2(a, b *uint64) int { return -Vec_WrdSortCompare1(a, b) }
func Vec_WrdSort(p *Vec_Wrd_t, reverse int) {
	sort.Slice(p.pArray[:p.nSize], func(i, j int) bool {
		if reverse != 0 {
			return p.pArray[i] > p.pArray[j]
		}
		return p.pArray[i] < p.pArray[j]
	})
}
func Vec_WrdLimit(p *Vec_Wrd_t) []uint64            { return p.pArray[p.nSize:] }
func Vec_WrdGetEntry(p *Vec_Wrd_t, i int) uint64    { Vec_WrdFillExtra(p, i+1, 0); return p.pArray[i] }
func Vec_WrdGetEntryP(p *Vec_Wrd_t, i int) *uint64  { Vec_WrdFillExtra(p, i+1, 0); return &p.pArray[i] }
func Vec_WrdSetEntry(p *Vec_Wrd_t, i int, x uint64) { Vec_WrdFillExtra(p, i+1, 0); p.pArray[i] = x }
func Vec_WrdPushFirst(p *Vec_Wrd_t, x uint64)       { Vec_WrdInsert(p, 0, x) }
func Vec_WrdPushUniqueOrder(p *Vec_Wrd_t, x uint64) int {
	if Vec_WrdFind(p, x) >= 0 {
		return 1
	}
	Vec_WrdPushOrder(p, x)
	return 0
}
func Vec_WrdInsert(p *Vec_Wrd_t, i int, x uint64) {
	if i < 0 || i > p.nSize {
		panic("insert index")
	}
	Vec_WrdPush(p, 0)
	copy(p.pArray[i+1:p.nSize], p.pArray[i:p.nSize-1])
	p.pArray[i] = x
}
func Vec_WrdReverseOrder(p *Vec_Wrd_t) {
	for i, j := 0, p.nSize-1; i < j; i, j = i+1, j-1 {
		p.pArray[i], p.pArray[j] = p.pArray[j], p.pArray[i]
	}
}
func Vec_WrdInvert(p *Vec_Wrd_t, fill uint64) *Vec_Wrd_t {
	q := Vec_WrdAlloc(0)
	if p.nSize == 0 {
		return q
	}
	Vec_WrdFill(q, int(Vec_WrdFindMax(p))+1, fill)
	for i, x := range Vec_WrdArray(p) {
		if x != fill {
			q.pArray[int(x)] = uint64(i)
		}
	}
	return q
}
func Vec_WrdSum(p *Vec_Wrd_t) uint64 {
	var sum uint64
	for _, x := range Vec_WrdArray(p) {
		sum += x
	}
	return sum
}
func Vec_WrdCountZero(p *Vec_Wrd_t) int { return Vec_WrdCountEntry(p, 0) }
func Vec_WrdSortCompareUnsigned(a, b *uint64) int {
	if *a < *b {
		return -1
	}
	if *a > *b {
		return 1
	}
	return 0
}
func Vec_WrdSortUnsigned(p *Vec_Wrd_t) {
	sort.Slice(Vec_WrdArray(p), func(i, j int) bool { return uint64(p.pArray[i]) < uint64(p.pArray[j]) })
}
func Vec_WrdAppend(p, q *Vec_Wrd_t) { Vec_WrdPushArray(p, Vec_WrdArray(q), q.nSize) }
func Vec_WrdStartNatural(n int) *Vec_Wrd_t {
	p := Vec_WrdStart(n)
	for i := range n {
		p.pArray[i] = uint64(i)
	}
	return p
}
func Vec_WrdStartRandom(n int) *Vec_Wrd_t {
	p := Vec_WrdStart(n)
	for i := range n {
		p.pArray[i] = rand.Uint64()
	}
	return p
}
func Vec_WrdStartTruthTables(vars int) *Vec_Wrd_t {
	words := 1
	if vars > 6 {
		words = 1 << (vars - 6)
	}
	p := Vec_WrdStart(words * vars)
	masks := [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	for i := range vars {
		for k := range words {
			if i < 6 {
				p.pArray[i*words+k] = masks[i]
			} else if k&(1<<(i-6)) != 0 {
				p.pArray[i*words+k] = ^uint64(0)
			}
		}
	}
	return p
}
func Vec_WrdFetch(p *Vec_Wrd_t, n int) []uint64 {
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
	return p.pArray[start:p.nSize]
}
func Vec_WrdCountCommon(p, q *Vec_Wrd_t) int {
	if p.nSize < q.nSize {
		p, q = q, p
	}
	seen := map[uint64]bool{}
	for _, x := range Vec_WrdArray(q) {
		seen[x] = true
	}
	n := 0
	for _, x := range Vec_WrdArray(p) {
		if seen[x] {
			n++
		}
	}
	return n
}
func Vec_WrdUniqify(p *Vec_Wrd_t) {
	if p.nSize < 2 {
		return
	}
	Vec_WrdSort(p, 0)
	k := 1
	for i := 1; i < p.nSize; i++ {
		if p.pArray[i] != p.pArray[i-1] {
			p.pArray[k] = p.pArray[i]
			k++
		}
	}
	p.nSize = k
}
func wordRecordKey(a []uint64) string {
	b := make([]byte, 8*len(a))
	for i, x := range a {
		binary.LittleEndian.PutUint64(b[8*i:], x)
	}
	return string(b)
}
func Vec_WrdUniqueCount(p *Vec_Wrd_t, size int, mapping **Vec_Int_t) int {
	if size <= 0 || p.nSize%size != 0 {
		panic("invalid record size")
	}
	seen := map[string]int32{}
	m := Vec_IntAlloc(p.nSize / size)
	for i := 0; i < p.nSize/size; i++ {
		key := wordRecordKey(p.pArray[i*size : (i+1)*size])
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
func Vec_WrdUniqifyHash(p *Vec_Wrd_t, size int) *Vec_Wrd_t {
	var mapping *Vec_Int_t
	n := Vec_WrdUniqueCount(p, size, &mapping)
	q := Vec_WrdAlloc(n * size)
	for i, x := range Vec_IntArray(mapping) {
		if int(x) == i {
			for _, v := range p.pArray[i*size : (i+1)*size] {
				Vec_WrdPush(q, v)
			}
		}
	}
	return q
}
func Gia_ManSimPatWriteOne(w io.Writer, a []uint64, n int) {
	digits := "0123456789ABCDEF"
	for _, x := range a[:n] {
		for k := 0; k < 16; k++ {
			fmt.Fprintf(w, "%c", digits[(x>>uint(k*4))&15])
		}
	}
	fmt.Fprintln(w)
}
func Vec_WrdPrintHex(p *Vec_Wrd_t, n int) {
	if n <= 0 || p.nSize%n != 0 {
		panic("invalid word record size")
	}
	for i := 0; i < p.nSize; i += n {
		Gia_ManSimPatWriteOne(os.Stdout, p.pArray[i:], n)
	}
}
func Vec_WrdDumpHex(path string, p *Vec_Wrd_t, n, verbose int) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Cannot open file %q for writing.\n", path)
		return
	}
	defer f.Close()
	if n <= 0 || p.nSize%n != 0 {
		panic("invalid word record size")
	}
	w := bufio.NewWriter(f)
	for i := 0; i < p.nSize; i += n {
		Gia_ManSimPatWriteOne(w, p.pArray[i:], n)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	if verbose != 0 {
		fmt.Printf("Written %d words of simulation data for %d objects into file %q.\n", n, p.nSize/n, path)
	}
}
func Vec_WrdReadHexOne(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	}
	panic("invalid hexadecimal digit")
}
func Vec_WrdReadHex(path string, words *int, verbose int) *Vec_Wrd_t {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Cannot open file %q for reading.\n", path)
		return nil
	}
	defer f.Close()
	p := Vec_WrdAlloc(1000)
	r := bufio.NewReader(f)
	n := -1
	chars := 0
	var value uint64
	for {
		c, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if c == '\r' || c == '\t' || c == ' ' {
			continue
		}
		if c == '\n' {
			if chars > 0 {
				Vec_WrdPush(p, value)
				chars = 0
				value = 0
			}
			if n == -1 && p.nSize > 0 {
				n = p.nSize
			}
			continue
		}
		value |= uint64(Vec_WrdReadHexOne(c)) << uint(chars*4)
		chars++
		if chars == 16 {
			Vec_WrdPush(p, value)
			chars = 0
			value = 0
		}
	}
	if chars > 0 {
		Vec_WrdPush(p, value)
	}
	if n < 0 {
		n = p.nSize
	}
	if n == 0 || p.nSize%n != 0 {
		panic("inconsistent hex records")
	}
	if words != nil {
		*words = n
	}
	if verbose != 0 {
		fmt.Printf("Read %d words of simulation data for %d objects.\n", n, p.nSize/n)
	}
	return p
}
func Vec_WrdZero(p *Vec_Wrd_t) { *p = Vec_Wrd_t{} }
func Vec_WrdCountEntry(p *Vec_Wrd_t, x uint64) int {
	n := 0
	for _, v := range Vec_WrdArray(p) {
		if v == x {
			n++
		}
	}
	return n
}
func Vec_WrdPushArray(p *Vec_Wrd_t, a []uint64, n int) {
	for _, x := range a[:n] {
		Vec_WrdPush(p, x)
	}
}
