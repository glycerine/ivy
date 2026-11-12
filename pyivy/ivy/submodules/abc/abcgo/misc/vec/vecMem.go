package vec

import (
	"fmt"
	"io"
	"os"
	"slices"
	"unsafe"
)

// Vec_Mem_t keeps fixed-sized word records in pages so record slices remain
// stable when more pages are allocated. Hash buckets contain record indices.
type Vec_Mem_t struct {
	nEntrySize, nEntries, LogPageSze int
	ppPages                          [][]uint64
	vTable, vNexts                   *Vec_Int_t
}

func Vec_MemAlloc_(p *Vec_Mem_t, size, logPage int) {
	if size <= 0 || logPage < 0 || logPage > 30 {
		panic("invalid paged vector geometry")
	}
	*p = Vec_Mem_t{nEntrySize: size, LogPageSze: logPage}
}
func Vec_MemAlloc(size, logPage int) *Vec_Mem_t {
	p := new(Vec_Mem_t)
	Vec_MemAlloc_(p, size, logPage)
	return p
}
func Vec_MemFree(p *Vec_Mem_t) {
	if p != nil {
		*p = Vec_Mem_t{}
	}
}
func Vec_MemFreeP(p **Vec_Mem_t) { Vec_MemFree(*p); *p = nil }
func Vec_MemDup(p *Vec_Mem_t) *Vec_Mem_t {
	q := Vec_MemAlloc(p.nEntrySize, p.LogPageSze)
	for i := range p.nEntries {
		Vec_MemPush(q, Vec_MemReadEntry(p, i))
	}
	if p.vTable != nil {
		Vec_MemHashAlloc(q, len(p.vTable.data))
	}
	return q
}

// Fill/Clean provide the zero-filled record operations left unimplemented in
// the original header. Existing records retain stable page storage.
func Vec_MemFill(p *Vec_Mem_t, n int) {
	if n < 0 {
		panic("negative entry count")
	}
	if n > 0 {
		Vec_MemGetEntry(p, n-1)
	}
	if n < p.nEntries {
		Vec_MemShrink(p, n)
	}
	Vec_MemClean(p, n)
}
func Vec_MemClean(p *Vec_Mem_t, n int) {
	if n < 0 || n > p.nEntries {
		panic("clean range")
	}
	for i := range n {
		clear(Vec_MemReadEntry(p, i))
	}
	if p.vTable != nil {
		vecMemRehash(p, len(p.vTable.data))
	}
}
func Vec_MemEntrySize(p *Vec_Mem_t) int { return p.nEntrySize }
func Vec_MemEntryNum(p *Vec_Mem_t) int  { return p.nEntries }
func Vec_MemPageSize(p *Vec_Mem_t) int  { return p.LogPageSze }
func Vec_MemPageNum(p *Vec_Mem_t) int   { return len(p.ppPages) }
func Vec_MemMemory(p *Vec_Mem_t) float64 {
	if p == nil {
		return 0
	}
	return float64(8*p.nEntrySize*(1<<p.LogPageSze)*(len(p.ppPages)) + cap(p.ppPages)*int(unsafe.Sizeof(uintptr(0))) + int(unsafe.Sizeof(*p)))
}

// Vec_MemReadEntry returns a borrowed record. If its words are edited directly
// while hashing is enabled, call Vec_MemHashResize before the next lookup.
func Vec_MemReadEntry(p *Vec_Mem_t, i int) []uint64 {
	if i < 0 || i >= p.nEntries {
		panic("memory vector index")
	}
	start := p.nEntrySize * (i & ((1 << p.LogPageSze) - 1))
	return p.ppPages[i>>p.LogPageSze][start : start+p.nEntrySize]
}
func Vec_MemReadEntryLast(p *Vec_Mem_t) []uint64 { return Vec_MemReadEntry(p, p.nEntries-1) }
func Vec_MemWriteEntry(p *Vec_Mem_t, i int, data []uint64) {
	copy(Vec_MemReadEntry(p, i), data[:p.nEntrySize])
	if p.vTable != nil {
		vecMemRehash(p, len(p.vTable.data))
	}
}
func Vec_MemGetEntry(p *Vec_Mem_t, i int) []uint64 {
	if i < 0 {
		panic("negative entry index")
	}
	if i >= p.nEntries {
		page := i >> p.LogPageSze
		for len(p.ppPages) <= page {
			p.ppPages = append(p.ppPages, make([]uint64, p.nEntrySize*(1<<p.LogPageSze)))
		}
		p.nEntries = i + 1
	}
	return Vec_MemReadEntry(p, i)
}
func Vec_MemSetEntry(p *Vec_Mem_t, i int, data []uint64) {
	existing := i < p.nEntries
	copy(Vec_MemGetEntry(p, i), data[:p.nEntrySize])
	if p.vTable != nil && (existing || len(p.vNexts.data) != p.nEntries) {
		vecMemRehash(p, len(p.vTable.data))
	}
}
func Vec_MemPush(p *Vec_Mem_t, data []uint64) { Vec_MemSetEntry(p, p.nEntries, data) }
func Vec_MemShrink(p *Vec_Mem_t, n int) {
	if n < 0 || n > p.nEntries {
		panic("memory vector shrink")
	}
	pages := (n + (1 << p.LogPageSze) - 1) >> p.LogPageSze
	clear(p.ppPages[pages:])
	p.ppPages = p.ppPages[:pages]
	p.nEntries = n
	if p.vNexts != nil {
		Vec_MemHashResize(p)
	}
}
func Vec_MemDumpDigit(w io.Writer, digit int) {
	if digit < 0 || digit > 15 {
		panic("hexadecimal digit")
	}
	fmt.Fprintf(w, "%c", "0123456789ABCDEF"[digit])
}
func Vec_MemDump(w io.Writer, p *Vec_Mem_t) {
	if w == os.Stdout {
		fmt.Printf("Memory vector has %d entries: \n", p.nEntries)
	}
	for i := range p.nEntries {
		a := Vec_MemReadEntry(p, i)
		for j := len(a) - 1; j >= 0; j-- {
			fmt.Fprintf(w, "%016X", a[j])
		}
		fmt.Fprintln(w)
	}
}
func vecMemPrime(n int) int {
	n = max(n, 2)
	for {
		prime := true
		for d := 2; d <= n/d; d++ {
			if n%d == 0 {
				prime = false
				break
			}
		}
		if prime {
			return n
		}
		n++
	}
}
func Vec_MemHashAlloc(p *Vec_Mem_t, n int) {
	if p.vTable != nil || p.vNexts != nil {
		panic("hash already allocated")
	}
	p.vTable = Vec_IntStartFull(vecMemPrime(n))
	p.vNexts = Vec_IntAlloc(n)
	if p.nEntries > 0 {
		vecMemRehash(p, len(p.vTable.data))
	}
}
func Vec_MemHashFree(p *Vec_Mem_t) {
	if p != nil {
		Vec_IntFreeP(&p.vTable)
		Vec_IntFreeP(&p.vNexts)
	}
}
func Vec_MemHashKey(p *Vec_Mem_t, data []uint64) uint32 {
	primes := [8]uint32{1699, 4177, 5147, 5647, 6343, 7103, 7873, 8147}
	var key uint32
	for i, x := range data[:p.nEntrySize] {
		key += uint32(x)*primes[(2*i)&7] + uint32(x>>32)*primes[(2*i+1)&7]
	}
	return uint32(uint64(key) % uint64(len(p.vTable.data)))
}
func Vec_MemHashLookup(p *Vec_Mem_t, data []uint64) *int32 {
	spot := Vec_IntEntryP(p.vTable, int(Vec_MemHashKey(p, data)))
	for *spot != -1 {
		if slices.Equal(Vec_MemReadEntry(p, int(*spot)), data[:p.nEntrySize]) {
			return spot
		}
		spot = Vec_IntEntryP(p.vNexts, int(*spot))
	}
	return spot
}
func Vec_MemHashProfile(p *Vec_Mem_t) {
	for _, head := range Vec_IntArray(p.vTable) {
		n := 0
		for head != -1 {
			n++
			head = Vec_IntEntry(p.vNexts, int(head))
		}
		fmt.Printf("%d ", n)
	}
	fmt.Println()
}

// Rehash mutable records without changing their IDs. Explicit writes may
// create equal records; lookup deterministically returns the lowest ID.
func vecMemRehash(p *Vec_Mem_t, buckets int) {
	Vec_IntFill(p.vTable, vecMemPrime(buckets), -1)
	Vec_IntFill(p.vNexts, p.nEntries, -1)
	for i := p.nEntries - 1; i >= 0; i-- {
		key := int(Vec_MemHashKey(p, Vec_MemReadEntry(p, i)))
		p.vNexts.data[i] = p.vTable.data[key]
		p.vTable.data[key] = int32(i)
	}
}
func Vec_MemHashResize(p *Vec_Mem_t) { vecMemRehash(p, 2*len(p.vTable.data)) }

func Vec_MemHashInsert(p *Vec_Mem_t, data []uint64) int {
	if p.vTable == nil {
		panic("hash not allocated")
	}
	if p.nEntries > len(p.vTable.data) {
		Vec_MemHashResize(p)
	}
	spot := Vec_MemHashLookup(p, data)
	if *spot != -1 {
		return int(*spot)
	}
	id := len(p.vNexts.data)
	*spot = int32(id)
	Vec_IntPush(p.vNexts, -1)
	Vec_MemPush(p, data)
	return id
}
func Vec_MemAllocForTTSimple(vars int) *Vec_Mem_t {
	words := 1
	if vars > 6 {
		words = 1 << (vars - 6)
	}
	p := Vec_MemAlloc(words, 12)
	Vec_MemHashAlloc(p, 10000)
	return p
}
func Vec_MemAllocForTT(vars, complement int) *Vec_Mem_t {
	p := Vec_MemAllocForTTSimple(vars)
	data := make([]uint64, p.nEntrySize)
	if Vec_MemHashInsert(p, data) != 0 {
		panic("constant truth ID")
	}
	for i := range data {
		data[i] = 0xaaaaaaaaaaaaaaaa
		if complement != 0 {
			data[i] = 0x5555555555555555
		}
	}
	if Vec_MemHashInsert(p, data) != 1 {
		panic("variable truth ID")
	}
	return p
}
func Vec_MemAddMuxTT(p *Vec_Mem_t, vars int) {
	words := 1
	if vars > 6 {
		words = 1 << (vars - 6)
	}
	data := make([]uint64, words)
	for i := range data {
		data[i] = 0xcacacacacacacaca
	}
	if Vec_MemHashInsert(p, data) != 2 {
		panic("mux truth ID")
	}
}
func Vec_MemDumpTruthTables(p *Vec_Mem_t, name string, lutSize int) {
	var w io.Writer = os.Stdout
	path := "stdout"
	if name != "" {
		path = fmt.Sprintf("tt_%s_%02d.txt", name, lutSize)
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		w = f
	}
	Vec_MemDump(w, p)
	fmt.Printf("Dumped %d %d-var truth tables into file %q (%.2f MB).\n", p.nEntries, lutSize, path, float64(8*p.nEntries*p.nEntrySize)/(1<<20))
}
