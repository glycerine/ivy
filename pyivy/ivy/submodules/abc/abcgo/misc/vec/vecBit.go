package vec

import (
	"math/bits"
	"unsafe"
)

// Vec_Bit_t packs bits into 32-bit words, matching ABC's bit numbering.
type Vec_Bit_t struct {
	nCap, nSize int
	pArray      []int32
}

func Vec_BitAlloc(n int) *Vec_Bit_t {
	words := (n + 31) / 32
	return &Vec_Bit_t{nCap: 32 * words, pArray: make([]int32, words)}
}
func Vec_BitStart(n int) *Vec_Bit_t { p := Vec_BitAlloc(n); p.nSize = p.nCap; return p }
func Vec_BitStartFull(n int) *Vec_Bit_t {
	p := Vec_BitStart(n)
	for i := range p.pArray {
		p.pArray[i] = -1
	}
	return p
}
func Vec_BitDup(p *Vec_Bit_t) *Vec_Bit_t {
	q := Vec_BitAlloc(p.nCap)
	q.nSize = p.nSize
	copy(q.pArray, p.pArray)
	return q
}
func Vec_BitFree(p *Vec_Bit_t) {
	if p != nil {
		*p = Vec_Bit_t{}
	}
}
func Vec_BitFreeP(p **Vec_Bit_t)               { Vec_BitFree(*p); *p = nil }
func Vec_BitReleaseArray(p *Vec_Bit_t) []int32 { a := p.pArray; *p = Vec_Bit_t{}; return a }
func Vec_BitArray(p *Vec_Bit_t) []int32        { return p.pArray[:(p.nSize+31)/32] }
func Vec_BitSize(p *Vec_Bit_t) int             { return p.nSize }
func Vec_BitCap(p *Vec_Bit_t) int              { return p.nCap }
func Vec_BitMemory(p *Vec_Bit_t) float64 {
	if p == nil {
		return 0
	}
	return float64(4*len(p.pArray) + int(unsafe.Sizeof(*p)))
}
func Vec_BitEntry(p *Vec_Bit_t, i int) int {
	if i < 0 || i >= p.nSize {
		panic("bit index")
	}
	return int(uint32(p.pArray[i>>5]) >> uint(i&31) & 1)
}
func Vec_BitWriteEntry(p *Vec_Bit_t, i, x int) {
	if i < 0 || i >= p.nSize {
		panic("bit index")
	}
	mask := int32(uint32(1) << uint(i&31))
	switch x {
	case 0:
		p.pArray[i>>5] &= ^mask
	case 1:
		p.pArray[i>>5] |= mask
	default:
		panic("bit value")
	}
}
func Vec_BitAddEntry(p *Vec_Bit_t, i int) int {
	old := Vec_BitEntry(p, i)
	Vec_BitWriteEntry(p, i, 1)
	return old
}
func Vec_BitEntryLast(p *Vec_Bit_t) int { return Vec_BitEntry(p, p.nSize-1) }
func Vec_BitGrow(p *Vec_Bit_t, n int) {
	if n <= p.nCap {
		return
	}
	words := (n + 31) / 32
	p.pArray = append(p.pArray, make([]int32, words-len(p.pArray))...)
	p.nCap = words * 32
}
func Vec_BitFill(p *Vec_Bit_t, n, x int) {
	Vec_BitGrow(p, n)
	p.nSize = (n + 31) / 32 * 32
	var fill int32
	if x == 1 {
		fill = -1
	} else if x != 0 {
		panic("bit value")
	}
	for i := range (n + 31) / 32 {
		p.pArray[i] = fill
	}
}
func Vec_BitFillExtra(p *Vec_Bit_t, n, x int) {
	if n <= p.nSize {
		return
	}
	if x != 0 && x != 1 {
		panic("bit value")
	}
	if n > p.nCap {
		Vec_BitGrow(p, max(n, 2*p.nCap))
	}
	old := p.nSize
	p.nSize = n
	for i := old; i < n; i++ {
		Vec_BitWriteEntry(p, i, x)
	}
}
func Vec_BitGetEntry(p *Vec_Bit_t, i int) int { Vec_BitFillExtra(p, i+1, 0); return Vec_BitEntry(p, i) }
func Vec_BitSetEntry(p *Vec_Bit_t, i, x int)  { Vec_BitFillExtra(p, i+1, 0); Vec_BitWriteEntry(p, i, x) }
func Vec_BitShrink(p *Vec_Bit_t, n int) {
	if n < 0 || n > p.nSize {
		panic("bit shrink")
	}
	p.nSize = n
}
func Vec_BitClear(p *Vec_Bit_t) { p.nSize = 0 }
func Vec_BitPush(p *Vec_Bit_t, x int) {
	if p.nSize == p.nCap {
		Vec_BitGrow(p, max(32, 2*p.nCap))
	}
	i := p.nSize
	p.nSize++
	Vec_BitWriteEntry(p, i, x)
}
func Vec_BitPop(p *Vec_Bit_t) int   { x := Vec_BitEntryLast(p); p.nSize--; return x }
func Vec_BitCountWord(x uint32) int { return bits.OnesCount32(x) }
func Vec_BitCount(p *Vec_Bit_t) int {
	n := 0
	for i, x := range Vec_BitArray(p) {
		if i == p.nSize/32 && p.nSize&31 != 0 {
			x &= int32((uint32(1) << uint(p.nSize&31)) - 1)
		}
		n += bits.OnesCount32(uint32(x))
	}
	return n
}
func Vec_BitReset(p *Vec_Bit_t) { clear(Vec_BitArray(p)) }
