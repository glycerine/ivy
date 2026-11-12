// Package mem implements ABC's byte and handle-based arena allocators using Go slices.
package mem

import "fmt"

type Mem_Fixed_t struct {
	entrySize, chunkSize int
	chunks               [][]byte
	free                 [][]byte
	used, peak           int
}

func Mem_FixedStart(size int32) *Mem_Fixed_t {
	if size <= 0 {
		panic("invalid entry size")
	}
	n := 1024
	if int(size)*n >= 65536 {
		n = 65536 / int(size)
	}
	if n < 8 {
		n = 8
	}
	return Mem_FixedStartWithChunk(size, int32(n))
}

// Mem_FixedStartWithChunk allows clients such as GIA to retain their original
// entry-count chunk policy while sharing the allocator implementation.
func Mem_FixedStartWithChunk(size, entries int32) *Mem_Fixed_t {
	if size <= 0 || entries <= 0 {
		panic("invalid fixed allocation dimensions")
	}
	if int64(size)*int64(entries) > int64(int(^uint(0)>>1)) {
		panic("fixed chunk size overflows int")
	}
	return &Mem_Fixed_t{entrySize: int(size), chunkSize: int(entries)}
}

func Mem_FixedEntryFetch(p *Mem_Fixed_t) []byte {
	if len(p.free) == 0 {
		chunk := make([]byte, p.entrySize*p.chunkSize)
		p.chunks = append(p.chunks, chunk)
		for i := p.chunkSize - 1; i >= 0; i-- {
			p.free = append(p.free, chunk[i*p.entrySize:(i+1)*p.entrySize:(i+1)*p.entrySize])
		}
	}
	i := len(p.free) - 1
	b := p.free[i]
	p.free = p.free[:i]
	p.used++
	if p.used > p.peak {
		p.peak = p.used
	}
	return b
}
func Mem_FixedEntryRecycle(p *Mem_Fixed_t, b []byte) {
	if len(b) != p.entrySize || p.used == 0 {
		panic("invalid recycled entry")
	}
	p.used--
	p.free = append(p.free, b)
}
func Mem_FixedRestart(p *Mem_Fixed_t) {
	if len(p.chunks) == 0 {
		return
	}
	clear(p.chunks[1:])
	p.chunks = p.chunks[:1]
	p.free = nil
	for i := p.chunkSize - 1; i >= 0; i-- {
		p.free = append(p.free, p.chunks[0][i*p.entrySize:(i+1)*p.entrySize:(i+1)*p.entrySize])
	}
	p.used = 0
}
func Mem_FixedReadMemUsage(p *Mem_Fixed_t) int32 {
	return int32(len(p.chunks) * p.chunkSize * p.entrySize)
}
func Mem_FixedReadMaxEntriesUsed(p *Mem_Fixed_t) int32 { return int32(p.peak) }
func Mem_FixedStop(p *Mem_Fixed_t, verbose int32) {
	if p == nil {
		return
	}
	if verbose != 0 {
		fmt.Printf("Fixed memory manager: entries used = %d, peak = %d, memory = %d\n", p.used, p.peak, Mem_FixedReadMemUsage(p))
	}
	*p = Mem_Fixed_t{}
}

type Mem_Flex_t struct {
	chunks                           [][]byte
	offset, chunkSize, used, entries int
}

func Mem_FlexStart() *Mem_Flex_t { return Mem_FlexStartWithChunk(4096) }

// Mem_FlexStartWithChunk selects the initial flexible-arena page size.
func Mem_FlexStartWithChunk(bytes int32) *Mem_Flex_t {
	if bytes <= 0 {
		panic("invalid flexible chunk size")
	}
	return &Mem_Flex_t{chunkSize: int(bytes)}
}
func Mem_FlexEntryFetch(p *Mem_Flex_t, n int32) []byte {
	if n < 0 {
		panic("negative allocation")
	}
	size := int(n)
	if len(p.chunks) == 0 || p.offset+size > len(p.chunks[len(p.chunks)-1]) {
		if size > p.chunkSize {
			p.chunkSize = 2 * size
		}
		p.chunks = append(p.chunks, make([]byte, p.chunkSize))
		p.offset = 0
	}
	b := p.chunks[len(p.chunks)-1][p.offset : p.offset+size : p.offset+size]
	p.offset += size
	p.used += size
	p.entries++
	return b
}
func Mem_FlexRestart(p *Mem_Flex_t) {
	if len(p.chunks) == 0 {
		return
	}
	clear(p.chunks[1:])
	p.chunks = p.chunks[:1]
	p.offset = 0
	p.used = 0
	p.entries = 0
}
func Mem_FlexReadMemUsage(p *Mem_Flex_t) int32 { return int32(p.used) }
func Mem_FlexStop(p *Mem_Flex_t, verbose int32) {
	if p == nil {
		return
	}
	if verbose != 0 {
		fmt.Printf("Flexible memory manager: entries = %d, memory used = %d\n", p.entries, p.used)
	}
	*p = Mem_Flex_t{}
}

type Mem_Step_t struct {
	managers []*Mem_Fixed_t
	large    [][]byte
}

func Mem_StepStart(n int32) *Mem_Step_t { return memStepStart(n, 0) }

// Mem_StepStartWithChunk uses a fixed entry count for every size class.
func Mem_StepStartWithChunk(n, entriesPerChunk int32) *Mem_Step_t {
	if entriesPerChunk <= 0 {
		panic("invalid step chunk size")
	}
	return memStepStart(n, entriesPerChunk)
}
func memStepStart(n, entriesPerChunk int32) *Mem_Step_t {
	if n <= 0 || n > 24 {
		panic("invalid step count")
	}
	p := &Mem_Step_t{}
	for i := int32(0); i < n; i++ {
		var manager *Mem_Fixed_t
		if entriesPerChunk > 0 {
			manager = Mem_FixedStartWithChunk(8<<i, entriesPerChunk)
		} else {
			manager = Mem_FixedStart(8 << i)
		}
		p.managers = append(p.managers, manager)
	}
	return p
}
func (p *Mem_Step_t) manager(n int32) *Mem_Fixed_t {
	for _, m := range p.managers {
		if int(n) <= m.entrySize {
			return m
		}
	}
	return nil
}
func Mem_StepEntryFetch(p *Mem_Step_t, n int32) []byte {
	if n == 0 {
		return nil
	}
	if n < 0 {
		panic("negative allocation")
	}
	if m := p.manager(n); m != nil {
		return Mem_FixedEntryFetch(m)
	}
	b := make([]byte, n)
	p.large = append(p.large, b)
	return b
}
func Mem_StepEntryRecycle(p *Mem_Step_t, b []byte, n int32) {
	if n <= 0 {
		return
	}
	if m := p.manager(n); m != nil {
		Mem_FixedEntryRecycle(m, b)
	}
}
func Mem_StepReadMemUsage(p *Mem_Step_t) int32 {
	var n int32
	for _, m := range p.managers {
		n += Mem_FixedReadMemUsage(m)
	}
	return n
}
func Mem_StepStop(p *Mem_Step_t, v int32) {
	if p == nil {
		return
	}
	for _, m := range p.managers {
		Mem_FixedStop(m, v)
	}
	*p = Mem_Step_t{}
}

// Mem_StepRestart restarts every fixed-size class and drops large allocations.
// Previously fetched entries lose ownership, exactly as with FixedRestart.
func Mem_StepRestart(p *Mem_Step_t) {
	if p == nil {
		return
	}
	for _, manager := range p.managers {
		Mem_FixedRestart(manager)
	}
	p.large = nil
}
