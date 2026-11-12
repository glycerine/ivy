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

type Vec_Wrd_t struct{ data []uint64 }

func Vec_WrdAlloc(n int) *Vec_Wrd_t {
	if n > 0 && n < 16 {
		n = 16
	}
	return Vec_WrdAllocExact(n)
}
func Vec_WrdAllocExact(n int) *Vec_Wrd_t { return &Vec_Wrd_t{data: make([]uint64, 0, n)} }
func Vec_WrdStart(n int) *Vec_Wrd_t      { p := Vec_WrdAlloc(n); p.resize(n); return p }
func Vec_WrdStartFull(n int) *Vec_Wrd_t {
	p := Vec_WrdStart(n)
	for i := 0; i < n; i++ {
		p.storage()[i] = ^uint64(0)
	}
	return p
}
func Vec_WrdAllocArray(a []uint64, n int) *Vec_Wrd_t {
	return &Vec_Wrd_t{data: (a[:n])[:n]}
}
func Vec_WrdAllocArrayCopy(a []uint64, n int) *Vec_Wrd_t {
	p := Vec_WrdAllocExact(n)
	p.resize(n)
	copy(p.storage(), a[:n])
	return p
}
func Vec_WrdDup(p *Vec_Wrd_t) *Vec_Wrd_t {
	q := Vec_WrdAllocExact(cap(p.data))
	q.resize(len(p.data))
	copy(q.storage(), p.data)
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
func Vec_WrdReleaseArray(p *Vec_Wrd_t) []uint64 { a := p.data; Vec_WrdZero(p); return a }
func Vec_WrdArray(p *Vec_Wrd_t) []uint64        { return p.data }
func Vec_WrdSize(p *Vec_Wrd_t) int              { return len(p.data) }
func Vec_WrdCap(p *Vec_Wrd_t) int               { return cap(p.data) }
func Vec_WrdMemory(p *Vec_Wrd_t) float64 {
	if p == nil {
		return 0
	}
	return float64(8*cap(p.data) + int(unsafe.Sizeof(*p)))
}
func Vec_WrdEntry(p *Vec_Wrd_t, i int) uint64         { return p.data[i] }
func Vec_WrdEntryP(p *Vec_Wrd_t, i int) *uint64       { return &p.data[i] }
func Vec_WrdWriteEntry(p *Vec_Wrd_t, i int, x uint64) { p.data[i] = x }
func Vec_WrdAddToEntry(p *Vec_Wrd_t, i int, x uint64) uint64 {
	p.data[i] += x
	return p.storage()[i]
}
func Vec_WrdEntryLast(p *Vec_Wrd_t) uint64 { return Vec_WrdEntry(p, len(p.data)-1) }
func Vec_WrdGrow(p *Vec_Wrd_t, n int) {
	if n > cap(p.data) {
		data := make([]uint64, len(p.data), n)
		copy(data, p.data)
		p.data = data
	}
}
func Vec_WrdFill(p *Vec_Wrd_t, n int, x uint64) {
	Vec_WrdGrow(p, n)
	for i := 0; i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_WrdFillExtra(p *Vec_Wrd_t, n int, x uint64) {
	if n <= len(p.data) {
		return
	}
	if n > cap(p.data) {
		Vec_WrdGrow(p, max(n, 2*cap(p.data)))
	}
	for i := len(p.data); i < n; i++ {
		p.storage()[i] = x
	}
	p.resize(n)
}
func Vec_WrdShrink(p *Vec_Wrd_t, n int) {
	if n < 0 || n > len(p.data) {
		panic("invalid vector shrink")
	}
	p.resize(n)
}
func Vec_WrdClear(p *Vec_Wrd_t) { p.resize(0) }
func Vec_WrdPush(p *Vec_Wrd_t, x uint64) {
	if len(p.data) == cap(p.data) {
		Vec_WrdGrow(p, max(16, 2*cap(p.data)))
	}
	p.storage()[len(p.data)] = x
	p.resize(len(p.data) + 1)
}
func Vec_WrdPushOrder(p *Vec_Wrd_t, x uint64) {
	Vec_WrdPush(p, x)
	i := len(p.data) - 1
	for i > 0 && p.storage()[i-1] > x {
		p.storage()[i] = p.storage()[i-1]
		i--
	}
	p.storage()[i] = x
}
func Vec_WrdPushUnique(p *Vec_Wrd_t, x uint64) int {
	if Vec_WrdFind(p, x) >= 0 {
		return 1
	}
	Vec_WrdPush(p, x)
	return 0
}
func Vec_WrdPop(p *Vec_Wrd_t) uint64 { x := Vec_WrdEntryLast(p); p.resize(len(p.data) - 1); return x }
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
	copy(p.storage()[i:], p.storage()[i+1:len(p.data)])
	p.resize(len(p.data) - 1)
	return 1
}
func Vec_WrdFindMax(p *Vec_Wrd_t) uint64 {
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
func Vec_WrdFindMin(p *Vec_Wrd_t) uint64 {
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
func Vec_WrdEqual(p, q *Vec_Wrd_t) int {
	if len(p.data) != len(q.data) {
		return 0
	}
	for i, v := range Vec_WrdArray(p) {
		if v != q.storage()[i] {
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
	sort.Slice(p.data, func(i, j int) bool {
		if reverse != 0 {
			return p.storage()[i] > p.storage()[j]
		}
		return p.storage()[i] < p.storage()[j]
	})
}
func Vec_WrdLimit(p *Vec_Wrd_t) []uint64         { return p.storage()[len(p.data):] }
func Vec_WrdGetEntry(p *Vec_Wrd_t, i int) uint64 { Vec_WrdFillExtra(p, i+1, 0); return p.storage()[i] }
func Vec_WrdGetEntryP(p *Vec_Wrd_t, i int) *uint64 {
	Vec_WrdFillExtra(p, i+1, 0)
	return &p.storage()[i]
}
func Vec_WrdSetEntry(p *Vec_Wrd_t, i int, x uint64) { Vec_WrdFillExtra(p, i+1, 0); p.storage()[i] = x }
func Vec_WrdPushFirst(p *Vec_Wrd_t, x uint64)       { Vec_WrdInsert(p, 0, x) }
func Vec_WrdPushUniqueOrder(p *Vec_Wrd_t, x uint64) int {
	if Vec_WrdFind(p, x) >= 0 {
		return 1
	}
	Vec_WrdPushOrder(p, x)
	return 0
}
func Vec_WrdInsert(p *Vec_Wrd_t, i int, x uint64) {
	if i < 0 || i > len(p.data) {
		panic("insert index")
	}
	Vec_WrdPush(p, 0)
	copy(p.storage()[i+1:len(p.data)], p.storage()[i:len(p.data)-1])
	p.storage()[i] = x
}
func Vec_WrdReverseOrder(p *Vec_Wrd_t) {
	for i, j := 0, len(p.data)-1; i < j; i, j = i+1, j-1 {
		p.storage()[i], p.storage()[j] = p.storage()[j], p.storage()[i]
	}
}
func Vec_WrdInvert(p *Vec_Wrd_t, fill uint64) *Vec_Wrd_t {
	q := Vec_WrdAlloc(0)
	if len(p.data) == 0 {
		return q
	}
	Vec_WrdFill(q, int(Vec_WrdFindMax(p))+1, fill)
	for i, x := range Vec_WrdArray(p) {
		if x != fill {
			q.storage()[int(x)] = uint64(i)
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
	sort.Slice(Vec_WrdArray(p), func(i, j int) bool { return uint64(p.storage()[i]) < uint64(p.storage()[j]) })
}
func Vec_WrdAppend(p, q *Vec_Wrd_t) { Vec_WrdPushArray(p, Vec_WrdArray(q), len(q.data)) }
func Vec_WrdStartNatural(n int) *Vec_Wrd_t {
	p := Vec_WrdStart(n)
	for i := range n {
		p.storage()[i] = uint64(i)
	}
	return p
}
func Vec_WrdStartRandom(n int) *Vec_Wrd_t {
	p := Vec_WrdStart(n)
	for i := range n {
		p.storage()[i] = rand.Uint64()
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
				p.storage()[i*words+k] = masks[i]
			} else if k&(1<<(i-6)) != 0 {
				p.storage()[i*words+k] = ^uint64(0)
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
	start := len(p.data)
	if n > cap(p.data)-len(p.data) {
		return nil
	}
	p.resize(len(p.data) + n)
	return p.storage()[start:len(p.data)]
}
func Vec_WrdCountCommon(p, q *Vec_Wrd_t) int {
	if len(p.data) < len(q.data) {
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
	if len(p.data) < 2 {
		return
	}
	Vec_WrdSort(p, 0)
	k := 1
	for i := 1; i < len(p.data); i++ {
		if p.storage()[i] != p.storage()[i-1] {
			p.storage()[k] = p.storage()[i]
			k++
		}
	}
	p.resize(k)
}
func wordRecordKey(a []uint64) string {
	b := make([]byte, 8*len(a))
	for i, x := range a {
		binary.LittleEndian.PutUint64(b[8*i:], x)
	}
	return string(b)
}
func Vec_WrdUniqueCount(p *Vec_Wrd_t, size int, mapping **Vec_Int_t) int {
	if size <= 0 || len(p.data)%size != 0 {
		panic("invalid record size")
	}
	seen := map[string]int32{}
	m := Vec_IntAlloc(len(p.data) / size)
	for i := 0; i < len(p.data)/size; i++ {
		key := wordRecordKey(p.storage()[i*size : (i+1)*size])
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
			for _, v := range p.storage()[i*size : (i+1)*size] {
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
	if n <= 0 || len(p.data)%n != 0 {
		panic("invalid word record size")
	}
	for i := 0; i < len(p.data); i += n {
		Gia_ManSimPatWriteOne(os.Stdout, p.storage()[i:], n)
	}
}
func Vec_WrdDumpHex(path string, p *Vec_Wrd_t, n, verbose int) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Cannot open file %q for writing.\n", path)
		return
	}
	defer f.Close()
	if n <= 0 || len(p.data)%n != 0 {
		panic("invalid word record size")
	}
	w := bufio.NewWriter(f)
	for i := 0; i < len(p.data); i += n {
		Gia_ManSimPatWriteOne(w, p.storage()[i:], n)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}
	if verbose != 0 {
		fmt.Printf("Written %d words of simulation data for %d objects into file %q.\n", n, len(p.data)/n, path)
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
			if n == -1 && len(p.data) > 0 {
				n = len(p.data)
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
		n = len(p.data)
	}
	if n == 0 || len(p.data)%n != 0 {
		panic("inconsistent hex records")
	}
	if words != nil {
		*words = n
	}
	if verbose != 0 {
		fmt.Printf("Read %d words of simulation data for %d objects.\n", n, len(p.data)/n)
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
