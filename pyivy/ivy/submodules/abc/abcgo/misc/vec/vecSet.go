package vec

import (
	"reflect"
	"strings"
)

// Vec_Set_t is a paged arena of 64-bit words. Handles encode page and word
// offset; the first two words of each page hold live and shadow limits.
// Explicit integer/byte packing replaces C's aliasing casts.
type Vec_Set_t struct {
	nPageSize                            int32
	uPageMask                            uint32
	nEntries, iPage, iPageS, nPagesAlloc int32
	pPages                               [][]uint64
}

func Vec_SetHandPage(p *Vec_Set_t, h int32) int32  { return h >> p.nPageSize }
func Vec_SetHandShift(p *Vec_Set_t, h int32) int32 { return int32(uint32(h) & p.uPageMask) }
func Vec_SetWordNum(n int32) int32 {
	if n < 0 {
		panic("vec: negative set size")
	}
	return (n + 1) >> 1
}
func Vec_SetEntry(p *Vec_Set_t, h int32) []uint64 {
	page, offset := Vec_SetHandPage(p, h), Vec_SetHandShift(p, h)
	if h < 0 || page > p.iPage || offset < 2 || offset >= 1<<p.nPageSize {
		panic("vec: invalid set handle")
	}
	return p.pPages[page][offset:]
}
func Vec_SetEntryNum(p *Vec_Set_t) int32            { return p.nEntries }
func Vec_SetWriteEntryNum(p *Vec_Set_t, n int32)    { p.nEntries = n }
func Vec_SetLimit(page []uint64) int32              { return int32(page[0]) }
func Vec_SetLimitS(page []uint64) int32             { return int32(page[1]) }
func Vec_SetIncLimit(page []uint64, n int32) int32  { page[0] += uint64(n); return int32(page[0]) }
func Vec_SetIncLimitS(page []uint64, n int32) int32 { page[1] += uint64(n); return int32(page[1]) }
func Vec_SetWriteLimit(page []uint64, n int32)      { page[0] = uint64(n) }
func Vec_SetWriteLimitS(page []uint64, n int32)     { page[1] = uint64(n) }
func Vec_SetHandCurrent(p *Vec_Set_t) int32 {
	return p.iPage<<p.nPageSize | Vec_SetLimit(p.pPages[p.iPage])
}
func Vec_SetHandCurrentS(p *Vec_Set_t) int32 {
	return p.iPageS<<p.nPageSize | Vec_SetLimitS(p.pPages[p.iPageS])
}
func Vec_SetHandMemory(p *Vec_Set_t, h int32) int32 { return h * 8 }
func Vec_SetMemory(p *Vec_Set_t) int32              { return Vec_SetHandMemory(p, Vec_SetHandCurrent(p)) }
func Vec_SetMemoryS(p *Vec_Set_t) int32             { return Vec_SetHandMemory(p, Vec_SetHandCurrentS(p)) }
func Vec_SetMemoryAll(p *Vec_Set_t) int32           { return (p.iPage + 1) * (1 << (p.nPageSize + 3)) }
func Vec_SetAlloc_(p *Vec_Set_t, bits int32) {
	if bits <= 8 || bits >= 30 {
		panic("vec: invalid set page size")
	}
	*p = Vec_Set_t{nPageSize: bits, uPageMask: uint32(1<<bits) - 1, nPagesAlloc: 256, pPages: make([][]uint64, 256)}
	p.pPages[0] = make([]uint64, 1<<bits)
	Vec_SetRestart(p)
}
func Vec_SetAlloc(bits int32) *Vec_Set_t { p := new(Vec_Set_t); Vec_SetAlloc_(p, bits); return p }
func Vec_SetRestart(p *Vec_Set_t) {
	p.nEntries = 0
	p.iPage = 0
	p.iPageS = 0
	p.pPages[0][0] = 2
	p.pPages[0][1] = ^uint64(0)
}
func Vec_SetFree_(p *Vec_Set_t) {
	if p != nil {
		p.pPages = nil
		p.nPagesAlloc = 0
		p.nEntries = 0
	}
}
func Vec_SetFree(p *Vec_Set_t) { Vec_SetFree_(p) }
func Vec_ReportMemory(p *Vec_Set_t) float64 {
	if p == nil {
		return 0
	}
	size := reflect.TypeOf(Vec_Set_t{}).Size() + uintptr(cap(p.pPages))*reflect.TypeOf([]uint64{}).Size()
	for _, page := range p.pPages {
		size += uintptr(cap(page)) * 8
	}
	return float64(size)
}
func (p *Vec_Set_t) ensurePage(index int32) {
	if index < 0 || int64(index)<<p.nPageSize > int64(^uint32(0)>>1) {
		panic("vec: set handle overflow")
	}
	for index >= int32(len(p.pPages)) {
		p.pPages = append(p.pPages, make([][]uint64, len(p.pPages))...)
	}
	p.nPagesAlloc = int32(len(p.pPages))
	if p.pPages[index] == nil {
		p.pPages[index] = make([]uint64, 1<<p.nPageSize)
	}
}
func Vec_SetAppend(p *Vec_Set_t, array []int32, size int32) int32 {
	words := Vec_SetWordNum(size)
	if words >= 1<<p.nPageSize-2 {
		panic("vec: entry does not fit set page")
	}
	if array != nil && int32(len(array)) < size {
		panic("vec: short set input")
	}
	if Vec_SetLimit(p.pPages[p.iPage])+words >= 1<<p.nPageSize {
		p.iPage++
		p.ensurePage(p.iPage)
		p.pPages[p.iPage][0] = 2
		p.pPages[p.iPage][1] = ^uint64(0)
	}
	offset := Vec_SetLimit(p.pPages[p.iPage])
	handle := p.iPage<<p.nPageSize | offset
	if array != nil {
		dst := p.pPages[p.iPage][offset:]
		for i := int32(0); i < size; i++ {
			shift := uint(32 * (i & 1))
			dst[i>>1] = dst[i>>1]&^(uint64(0xffffffff)<<shift) | uint64(uint32(array[i]))<<shift
		}
	}
	Vec_SetIncLimit(p.pPages[p.iPage], words)
	p.nEntries++
	return handle
}
func Vec_SetAppendS(p *Vec_Set_t, size int32) int32 {
	words := Vec_SetWordNum(size)
	if words >= 1<<p.nPageSize-2 {
		panic("vec: shadow entry does not fit page")
	}
	limit := Vec_SetLimitS(p.pPages[p.iPageS])
	if limit < 2 {
		panic("vec: shadow traversal not initialized")
	}
	if limit+words >= 1<<p.nPageSize {
		p.iPageS++
		if p.iPageS > p.iPage || p.pPages[p.iPageS] == nil {
			panic("vec: shadow exceeds live pages")
		}
		Vec_SetWriteLimitS(p.pPages[p.iPageS], 2)
	}
	handle := Vec_SetHandCurrentS(p)
	Vec_SetIncLimitS(p.pPages[p.iPageS], words)
	return handle
}
func Vec_SetFetchH(p *Vec_Set_t, bytes int32) int32 {
	if bytes < 0 {
		panic("vec: negative allocation")
	}
	return Vec_SetAppend(p, nil, (bytes+3)>>2)
}
func Vec_SetFetch(p *Vec_Set_t, bytes int32) []uint64 {
	h := Vec_SetFetchH(p, bytes)
	return Vec_SetEntry(p, h)[:(bytes+7)>>3]
}
func Vec_SetStrsav(p *Vec_Set_t, name string) string {
	data := Vec_SetFetch(p, int32(len(name)+1))
	for i := 0; i <= len(name); i++ {
		value := byte(0)
		if i < len(name) {
			value = name[i]
		}
		shift := uint(8 * (i & 7))
		data[i>>3] = data[i>>3]&^(uint64(255)<<shift) | uint64(value)<<shift
	}
	return strings.Clone(name)
}
func Vec_SetShrink(p *Vec_Set_t, h int32) {
	if h < 2 || h > Vec_SetHandCurrent(p) || Vec_SetHandShift(p, h) < 2 {
		panic("vec: invalid set shrink")
	}
	p.iPage = Vec_SetHandPage(p, h)
	Vec_SetWriteLimit(p.pPages[p.iPage], Vec_SetHandShift(p, h))
}
func Vec_SetShrinkS(p *Vec_Set_t, h int32) {
	if h < 2 || h > Vec_SetHandCurrent(p) || Vec_SetHandShift(p, h) < 2 {
		panic("vec: invalid shadow shrink")
	}
	p.iPageS = Vec_SetHandPage(p, h)
	Vec_SetWriteLimitS(p.pPages[p.iPageS], Vec_SetHandShift(p, h))
}
func Vec_SetShrinkLimits(p *Vec_Set_t) {
	for i := int32(0); i <= p.iPage; i++ {
		limit := Vec_SetLimitS(p.pPages[i])
		if limit < 2 {
			panic("vec: missing shadow limit")
		}
		Vec_SetWriteLimit(p.pPages[i], limit)
	}
}

// Vec_SetReadInt interprets the explicitly packed 32-bit records.
func Vec_SetReadInt(p *Vec_Set_t, h, index int32) int32 {
	data := Vec_SetEntry(p, h)
	return int32(uint32(data[index>>1] >> uint(32*(index&1))))
}
func Vec_SetWriteInt(p *Vec_Set_t, h, index, value int32) {
	data := Vec_SetEntry(p, h)
	shift := uint(32 * (index & 1))
	data[index>>1] = data[index>>1]&^(uint64(0xffffffff)<<shift) | uint64(uint32(value))<<shift
}
