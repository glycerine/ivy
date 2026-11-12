package extra

import "fmt"

type Extra_MmFixed_t = Extra_MmFixed_t_

type Extra_MmFlex_t = Extra_MmFlex_t_

type Extra_MmStep_t = Extra_MmStep_t_

/**CFile****************************************************************

  FileName    [extraUtilMemory.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Memory managers.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilMemory.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

// Memory entries are mutable byte slices rather than C pointers. The free
// list is held separately, so recycling never encodes Go pointers into bytes.
type Extra_MmFixed_t_ struct {
	nEntrySize, nChunkSize                   int32
	nEntriesAlloc, nEntriesUsed, nEntriesMax int32
	nMemoryAlloc                             int32
	chunks                                   [][]byte
	free                                     [][]byte
}
type Extra_MmFlex_t_ struct {
	nChunkSize, nEntriesUsed, nMemoryUsed, nMemoryAlloc int32
	chunks                                              [][]byte
	current                                             []byte
}
type Extra_MmStep_t_ struct {
	mems     []*Extra_MmFixed_t_
	nMapSize int32
	large    [][]byte
}

func Extra_MmFixedStart(size int32) *Extra_MmFixed_t_ {
	if size <= 0 || size > (1<<31-1)/8 {
		panic("fixed entry size outside supported range")
	}
	count := int32(1024)
	if int64(size)*1024 >= 65536 {
		count = 65536 / size
	}
	if count < 8 {
		count = 8
	}
	return &Extra_MmFixed_t_{nEntrySize: size, nChunkSize: count}
}
func (p *Extra_MmFixed_t_) addChunk() {
	data := make([]byte, int(p.nEntrySize)*int(p.nChunkSize))
	p.chunks = append(p.chunks, data)
	for i := p.nChunkSize; i > 0; i-- {
		start := (i - 1) * p.nEntrySize
		p.free = append(p.free, data[start:start+p.nEntrySize:start+p.nEntrySize])
	}
	p.nEntriesAlloc += p.nChunkSize
	p.nMemoryAlloc += p.nEntrySize * p.nChunkSize
}
func Extra_MmFixedEntryFetch(p *Extra_MmFixed_t_) []byte {
	if len(p.free) == 0 {
		p.addChunk()
	}
	last := len(p.free) - 1
	entry := p.free[last]
	p.free[last] = nil
	p.free = p.free[:last]
	p.nEntriesUsed++
	if p.nEntriesUsed > p.nEntriesMax {
		p.nEntriesMax = p.nEntriesUsed
	}
	return entry
}
func Extra_MmFixedEntryRecycle(p *Extra_MmFixed_t_, entry []byte) {
	if len(entry) != int(p.nEntrySize) || p.nEntriesUsed <= 0 {
		panic("invalid recycled fixed entry")
	}
	p.nEntriesUsed--
	p.free = append(p.free, entry)
}
func Extra_MmFixedRestart(p *Extra_MmFixed_t_) {
	if len(p.chunks) == 0 {
		return
	}
	first := p.chunks[0]
	p.chunks = [][]byte{first}
	p.free = nil
	for i := p.nChunkSize; i > 0; i-- {
		start := (i - 1) * p.nEntrySize
		p.free = append(p.free, first[start:start+p.nEntrySize:start+p.nEntrySize])
	}
	p.nEntriesAlloc = p.nChunkSize
	p.nEntriesUsed = 0
	p.nMemoryAlloc = p.nEntrySize * p.nChunkSize
}
func Extra_MmFixedReadMemUsage(p *Extra_MmFixed_t_) int32       { return p.nMemoryAlloc }
func Extra_MmFixedReadMaxEntriesUsed(p *Extra_MmFixed_t_) int32 { return p.nEntriesMax }
func Extra_MmFixedStop(p *Extra_MmFixed_t_) {
	if p != nil {
		p.chunks = nil
		p.free = nil
	}
}
func Extra_MmFixedPrint(p *Extra_MmFixed_t_) {
	fmt.Printf("Fixed memory manager: Entry = %5d. Chunk = %5d. Chunks used = %5d.\n", p.nEntrySize, p.nChunkSize, len(p.chunks))
	fmt.Printf("   Entries used = %8d. Entries peak = %8d. Memory used = %8d. Memory alloc = %8d.\n", p.nEntriesUsed, p.nEntriesMax, p.nEntrySize*p.nEntriesUsed, p.nMemoryAlloc)
}
func Extra_MmFlexStart() *Extra_MmFlex_t_ { return &Extra_MmFlex_t_{nChunkSize: 4096} }
func Extra_MmFlexEntryFetch(p *Extra_MmFlex_t_, size int32) []byte {
	if size < 0 {
		panic("negative entry size")
	}
	if p.current == nil || int(size) > len(p.current) {
		if size > p.nChunkSize {
			if size > (1<<31-1)/2 {
				panic("flexible chunk size overflow")
			}
			p.nChunkSize = 2 * size
		}
		p.current = make([]byte, p.nChunkSize)
		p.chunks = append(p.chunks, p.current)
		p.nMemoryAlloc += p.nChunkSize
	}
	entry := p.current[:size:size]
	p.current = p.current[size:]
	p.nEntriesUsed++
	p.nMemoryUsed += size
	return entry
}
func Extra_MmFlexReadMemUsage(p *Extra_MmFlex_t_) int32 { return p.nMemoryAlloc }
func Extra_MmFlexStop(p *Extra_MmFlex_t_) {
	if p != nil {
		p.current = nil
		p.chunks = nil
	}
}
func Extra_MmFlexPrint(p *Extra_MmFlex_t_) {
	fmt.Printf("Flexible memory manager: Chunk size = %d. Chunks used = %d.\n", p.nChunkSize, len(p.chunks))
	fmt.Printf("   Entries used = %d. Memory used = %d. Memory alloc = %d.\n", p.nEntriesUsed, p.nMemoryUsed, p.nMemoryAlloc)
}
func Extra_MmStepStart(steps int32) *Extra_MmStep_t_ {
	if steps <= 0 || steps > 26 {
		panic("invalid number of memory steps")
	}
	p := &Extra_MmStep_t_{nMapSize: 4 << uint(steps)}
	for i := int32(0); i < steps; i++ {
		p.mems = append(p.mems, Extra_MmFixedStart(8<<uint(i)))
	}
	return p
}
func (p *Extra_MmStep_t_) manager(size int32) *Extra_MmFixed_t_ {
	for _, m := range p.mems {
		if size <= m.nEntrySize {
			return m
		}
	}
	panic("size outside memory classes")
}
func Extra_MmStepEntryFetch(p *Extra_MmStep_t_, size int32) []byte {
	if size < 0 {
		panic("negative entry size")
	}
	if size == 0 {
		return nil
	}
	if size > p.nMapSize {
		entry := make([]byte, size)
		p.large = append(p.large, entry)
		return entry
	}
	return Extra_MmFixedEntryFetch(p.manager(size))
}
func Extra_MmStepEntryRecycle(p *Extra_MmStep_t_, entry []byte, size int32) {
	if size == 0 || size > p.nMapSize {
		return
	}
	if size < 0 {
		panic("negative entry size")
	}
	Extra_MmFixedEntryRecycle(p.manager(size), entry)
}
func Extra_MmStepReadMemUsage(p *Extra_MmStep_t_) int32 {
	var n int32
	for _, m := range p.mems {
		n += m.nMemoryAlloc
	}
	return n
}
func Extra_MmStepStop(p *Extra_MmStep_t_) {
	if p != nil {
		for _, m := range p.mems {
			Extra_MmFixedStop(m)
		}
		p.mems = nil
		p.large = nil
	}
}
