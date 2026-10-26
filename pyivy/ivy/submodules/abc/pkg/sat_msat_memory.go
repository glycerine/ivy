package abc

// Checked owned byte arenas from src/sat/msat/msatMem.c.

import "fmt"

type MsatMemoryLimits struct {
	MaxBytes, MaxBlocks int
}

func checkMsatMemoryLimits(limits MsatMemoryLimits) (MsatMemoryLimits, error) {
	if limits.MaxBytes < 0 || limits.MaxBlocks < 0 {
		return limits, ErrDimension
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1 << 30
	}
	if limits.MaxBlocks == 0 {
		limits.MaxBlocks = 16 << 20
	}
	return limits, nil
}

type MsatMemoryBlock struct {
	data                         []byte
	fixed                        *MsatFixedArena
	step                         *MsatStepArena
	requested, allocation, index int
	generation                   uint64
	live                         bool
}

func (b *MsatMemoryBlock) Bytes() []byte {
	if b == nil || !b.live || b.fixed != nil && b.generation != b.fixed.generation {
		return nil
	}
	return b.data[:b.requested:b.requested]
}

type MsatFixedMemoryStats struct {
	EntrySize, ChunkSize, Chunks, EntriesAllocated int
	EntriesUsed, PeakEntries, MemoryAllocated      int
}

type msatFixedChunk struct {
	data   []byte
	blocks []*MsatMemoryBlock
}

type MsatFixedArena struct {
	entrySize, chunkSize, used, peak, allocated int
	limits                                      MsatMemoryLimits
	chunks                                      []msatFixedChunk
	free                                        []*MsatMemoryBlock
	generation                                  uint64
	closed                                      bool
}

func NewMsatFixedArena(entrySize int, limits MsatMemoryLimits) (*MsatFixedArena, error) {
	if entrySize < 4 {
		return nil, ErrDimension
	}
	limits, err := checkMsatMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	chunkSize := 1 << 10
	if entrySize*chunkSize >= 1<<16 {
		chunkSize = (1 << 16) / entrySize
	}
	if chunkSize < 8 {
		chunkSize = 8
	}
	if _, err = checkedSize(0, entrySize, chunkSize, limits.MaxBytes); err != nil || chunkSize > limits.MaxBlocks {
		return nil, ErrLimit
	}
	return &MsatFixedArena{entrySize: entrySize, chunkSize: chunkSize, limits: limits, generation: 1}, nil
}

func (a *MsatFixedArena) checkLive() error {
	if a == nil {
		return ErrDimension
	}
	if a.closed || a.entrySize < 4 || a.chunkSize < 8 || a.used < 0 || a.allocated < 0 {
		return ErrInvariant
	}
	return nil
}

func (a *MsatFixedArena) grow() error {
	if err := a.checkLive(); err != nil {
		return err
	}
	bytes, err := checkedSize(0, a.entrySize, a.chunkSize, a.limits.MaxBytes-a.allocated)
	if err != nil || a.chunkSize > a.limits.MaxBlocks-len(a.chunks)*a.chunkSize {
		return ErrLimit
	}
	chunk := msatFixedChunk{data: make([]byte, bytes), blocks: make([]*MsatMemoryBlock, a.chunkSize)}
	for index := range chunk.blocks {
		chunk.blocks[index] = &MsatMemoryBlock{
			data:       chunk.data[index*a.entrySize : (index+1)*a.entrySize],
			fixed:      a,
			requested:  a.entrySize,
			allocation: a.entrySize,
			index:      index,
			generation: a.generation,
		}
	}
	a.chunks = append(a.chunks, chunk)
	for index := len(chunk.blocks) - 1; index >= 0; index-- {
		a.free = append(a.free, chunk.blocks[index])
	}
	a.allocated += bytes
	return nil
}

func (a *MsatFixedArena) Fetch() (*MsatMemoryBlock, error) {
	if err := a.checkLive(); err != nil {
		return nil, err
	}
	if len(a.free) == 0 {
		if err := a.grow(); err != nil {
			return nil, err
		}
	}
	block := a.free[len(a.free)-1]
	a.free = a.free[:len(a.free)-1]
	block.live, block.generation, block.requested = true, a.generation, a.entrySize
	a.used++
	if a.peak < a.used {
		a.peak = a.used
	}
	return block, nil
}

func (a *MsatFixedArena) Recycle(block *MsatMemoryBlock) error {
	if err := a.checkLive(); err != nil {
		return err
	}
	if block == nil || block.fixed != a || !block.live || block.generation != a.generation {
		return ErrInvariant
	}
	block.live = false
	a.used--
	a.free = append(a.free, block)
	return nil
}

func (a *MsatFixedArena) Restart() error {
	if err := a.checkLive(); err != nil {
		return err
	}
	if len(a.chunks) == 0 {
		return nil
	}
	a.generation++
	first := a.chunks[0]
	for chunkIndex, chunk := range a.chunks {
		for _, block := range chunk.blocks {
			block.live = false
			if chunkIndex != 0 {
				block.data = nil
			}
		}
	}
	a.chunks = a.chunks[:1]
	a.allocated, a.used = a.entrySize*a.chunkSize, 0
	a.free = a.free[:0]
	for index := len(first.blocks) - 1; index >= 0; index-- {
		first.blocks[index].generation = a.generation
		a.free = append(a.free, first.blocks[index])
	}
	return nil
}

func (a *MsatFixedArena) MemoryUsage() int {
	if a == nil || a.closed {
		return 0
	}
	return a.allocated
}

func (a *MsatFixedArena) Stats() MsatFixedMemoryStats {
	if a == nil || a.closed {
		return MsatFixedMemoryStats{}
	}
	return MsatFixedMemoryStats{a.entrySize, a.chunkSize, len(a.chunks), len(a.chunks) * a.chunkSize, a.used, a.peak, a.allocated}
}

func (a *MsatFixedArena) Diagnostic() string {
	stats := a.Stats()
	return fmt.Sprintf("Fixed memory manager: Entry = %5d. Chunk = %5d. Chunks used = %5d.\n   Entries used = %8d. Entries peak = %8d. Memory used = %8d. Memory alloc = %8d.\n",
		stats.EntrySize, stats.ChunkSize, stats.Chunks, stats.EntriesUsed, stats.PeakEntries, stats.EntrySize*stats.EntriesUsed, stats.MemoryAllocated)
}

func (a *MsatFixedArena) Close() {
	if a == nil || a.closed {
		return
	}
	for _, chunk := range a.chunks {
		for _, block := range chunk.blocks {
			block.live = false
			block.data = nil
		}
	}
	a.chunks, a.free, a.allocated, a.used, a.closed = nil, nil, 0, 0, true
}

type MsatFlexMemoryStats struct {
	ChunkSize, Chunks, EntriesUsed, MemoryUsed, MemoryAllocated int
}

type msatFlexChunk struct {
	data   []byte
	offset int
	blocks []*MsatMemoryBlock
}

type MsatFlexArena struct {
	chunkSize, entries, used, allocated int
	limits                              MsatMemoryLimits
	chunks                              []msatFlexChunk
	closed                              bool
}

func NewMsatFlexArena(limits MsatMemoryLimits) (*MsatFlexArena, error) {
	limits, err := checkMsatMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	return &MsatFlexArena{chunkSize: 1 << 12, limits: limits}, nil
}

func (a *MsatFlexArena) Fetch(bytes int) (*MsatMemoryBlock, error) {
	if a == nil {
		return nil, ErrDimension
	}
	if a.closed {
		return nil, ErrInvariant
	}
	if bytes < 0 {
		return nil, ErrDimension
	}
	if a.entries == a.limits.MaxBlocks {
		return nil, ErrLimit
	}
	if len(a.chunks) == 0 || bytes > len(a.chunks[len(a.chunks)-1].data)-a.chunks[len(a.chunks)-1].offset {
		if bytes > a.chunkSize {
			if bytes > a.limits.MaxBytes/2 {
				return nil, ErrLimit
			}
			a.chunkSize = 2 * bytes
		}
		if a.chunkSize > a.limits.MaxBytes-a.allocated {
			return nil, ErrLimit
		}
		a.chunks = append(a.chunks, msatFlexChunk{data: make([]byte, a.chunkSize)})
		a.allocated += a.chunkSize
	}
	chunk := &a.chunks[len(a.chunks)-1]
	start := chunk.offset
	chunk.offset += bytes
	block := &MsatMemoryBlock{data: chunk.data[start:chunk.offset], requested: bytes, allocation: bytes, index: start, live: true}
	chunk.blocks = append(chunk.blocks, block)
	a.entries++
	a.used += bytes
	return block, nil
}

func (a *MsatFlexArena) MemoryUsage() int {
	if a == nil || a.closed {
		return 0
	}
	return a.allocated
}

func (a *MsatFlexArena) Stats() MsatFlexMemoryStats {
	if a == nil || a.closed {
		return MsatFlexMemoryStats{}
	}
	return MsatFlexMemoryStats{a.chunkSize, len(a.chunks), a.entries, a.used, a.allocated}
}

func (a *MsatFlexArena) Diagnostic() string {
	stats := a.Stats()
	return fmt.Sprintf("Flexible memory manager: Chunk size = %d. Chunks used = %d.\n   Entries used = %d. Memory used = %d. Memory alloc = %d.\n",
		stats.ChunkSize, stats.Chunks, stats.EntriesUsed, stats.MemoryUsed, stats.MemoryAllocated)
}

func (a *MsatFlexArena) Close() {
	if a == nil || a.closed {
		return
	}
	for _, chunk := range a.chunks {
		for _, block := range chunk.blocks {
			block.live = false
			block.data = nil
		}
	}
	a.chunks, a.entries, a.used, a.allocated, a.closed = nil, 0, 0, 0, true
}

type MsatStepMemoryStats struct {
	Steps, MapSize, LargeBlocks, ClassMemoryAllocated int
	LargeMemoryAllocated                              int
	Classes                                           []MsatFixedMemoryStats
}

type MsatStepArena struct {
	memories            []*MsatFixedArena
	large               []*MsatMemoryBlock
	mapSize, largeBytes int
	limits              MsatMemoryLimits
	closed              bool
}

func NewMsatStepArena(steps int, limits MsatMemoryLimits) (*MsatStepArena, error) {
	if steps < 1 || steps > 26 {
		return nil, ErrDimension
	}
	limits, err := checkMsatMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	a := &MsatStepArena{mapSize: 4 << steps, limits: limits}
	for index := 0; index < steps; index++ {
		memory, makeErr := NewMsatFixedArena(8<<index, limits)
		if makeErr != nil {
			return nil, makeErr
		}
		a.memories = append(a.memories, memory)
	}
	return a, nil
}

func (a *MsatStepArena) class(bytes int) int {
	class, size := 0, 8
	for bytes > size {
		class, size = class+1, size<<1
	}
	return class
}

func (a *MsatStepArena) Fetch(bytes int) (*MsatMemoryBlock, error) {
	if a == nil {
		return nil, ErrDimension
	}
	if a.closed {
		return nil, ErrInvariant
	}
	if bytes < 0 {
		return nil, ErrDimension
	}
	if bytes == 0 {
		return nil, nil
	}
	if bytes > a.mapSize {
		if len(a.large) == a.limits.MaxBlocks || bytes > a.limits.MaxBytes-a.MemoryUsage()-a.largeBytes {
			return nil, ErrLimit
		}
		block := &MsatMemoryBlock{data: make([]byte, bytes), step: a, requested: bytes, allocation: bytes, live: true}
		a.large = append(a.large, block)
		a.largeBytes += bytes
		return block, nil
	}
	memory := a.memories[a.class(bytes)]
	if len(memory.free) == 0 && memory.entrySize*memory.chunkSize > a.limits.MaxBytes-a.MemoryUsage()-a.largeBytes {
		return nil, ErrLimit
	}
	block, err := memory.Fetch()
	if err != nil {
		return nil, err
	}
	block.step, block.requested = a, bytes
	return block, nil
}

func (a *MsatStepArena) Recycle(block *MsatMemoryBlock, bytes int) error {
	if a == nil {
		return ErrDimension
	}
	if a.closed {
		return ErrInvariant
	}
	if bytes < 0 {
		return ErrDimension
	}
	if bytes == 0 {
		if block != nil {
			return ErrInvariant
		}
		return nil
	}
	if block == nil || block.step != a || !block.live || block.requested != bytes {
		return ErrInvariant
	}
	if bytes <= a.mapSize {
		return a.memories[a.class(bytes)].Recycle(block)
	}
	position := -1
	for index, candidate := range a.large {
		if candidate == block {
			position = index
			break
		}
	}
	if position < 0 {
		return ErrInvariant
	}
	block.live = false
	block.data = nil
	a.largeBytes -= block.allocation
	copy(a.large[position:], a.large[position+1:])
	a.large[len(a.large)-1] = nil
	a.large = a.large[:len(a.large)-1]
	return nil
}

// MemoryUsage matches Msat_MmStepReadMemUsage and excludes direct oversized
// allocations, which the C function does not include either.
func (a *MsatStepArena) MemoryUsage() int {
	if a == nil || a.closed {
		return 0
	}
	total := 0
	for _, memory := range a.memories {
		total += memory.MemoryUsage()
	}
	return total
}

func (a *MsatStepArena) Stats() MsatStepMemoryStats {
	if a == nil || a.closed {
		return MsatStepMemoryStats{}
	}
	classes := make([]MsatFixedMemoryStats, len(a.memories))
	for index, memory := range a.memories {
		classes[index] = memory.Stats()
	}
	return MsatStepMemoryStats{len(a.memories), a.mapSize, len(a.large), a.MemoryUsage(), a.largeBytes, classes}
}

func (a *MsatStepArena) Close() {
	if a == nil || a.closed {
		return
	}
	for _, memory := range a.memories {
		memory.Close()
	}
	for _, block := range a.large {
		block.live = false
		block.data = nil
	}
	a.memories, a.large, a.largeBytes, a.closed = nil, nil, 0, true
}
