package abc

// Port of src/sat/bsat/satMem.c and satMem.h using checked owned byte arenas.

type SatMemoryLimits struct {
	MaxBytes, MaxBlocks int // zero selects conservative package defaults
}

func checkSatMemoryLimits(limits SatMemoryLimits) (SatMemoryLimits, error) {
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

// SatMemoryBlock is an owned arena allocation. Bytes returns nil after the
// allocation is recycled or its arena is restarted or stopped.
type SatMemoryBlock struct {
	data                         []byte
	fixed                        *SatFixedArena
	step                         *SatStepArena
	requested, allocation, index int
	generation                   uint64
	live                         bool
}

func (b *SatMemoryBlock) Bytes() []byte {
	if b == nil || !b.live || (b.fixed != nil && b.generation != b.fixed.generation) {
		return nil
	}
	return b.data[:b.requested:b.requested]
}

type SatFixedMemoryStats struct {
	EntrySize, ChunkSize, Chunks, EntriesAllocated int
	EntriesUsed, PeakEntries, MemoryAllocated      int
}

type satFixedChunk struct {
	data   []byte
	blocks []*SatMemoryBlock
}

type SatFixedArena struct {
	entrySize, chunkSize, used, peak, allocated int
	limits                                      SatMemoryLimits
	chunks                                      []satFixedChunk
	free                                        []*SatMemoryBlock
	generation                                  uint64
	stopped                                     bool
}

func NewSatFixedArena(entrySize int, limits SatMemoryLimits) (*SatFixedArena, error) {
	if entrySize < 1 {
		return nil, ErrDimension
	}
	limits, err := checkSatMemoryLimits(limits)
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
	return &SatFixedArena{entrySize: entrySize, chunkSize: chunkSize, limits: limits, generation: 1}, nil
}

func (a *SatFixedArena) grow() error {
	if a == nil || a.stopped {
		return ErrInvariant
	}
	bytes, err := checkedSize(0, a.entrySize, a.chunkSize, a.limits.MaxBytes-a.allocated)
	if err != nil || a.chunkSize > a.limits.MaxBlocks-len(a.chunks)*a.chunkSize {
		return ErrLimit
	}
	chunk := satFixedChunk{data: make([]byte, bytes), blocks: make([]*SatMemoryBlock, a.chunkSize)}
	for i := range chunk.blocks {
		chunk.blocks[i] = &SatMemoryBlock{data: chunk.data[i*a.entrySize : (i+1)*a.entrySize], fixed: a, requested: a.entrySize, allocation: a.entrySize, index: i, generation: a.generation}
	}
	a.chunks = append(a.chunks, chunk)
	for i := len(chunk.blocks) - 1; i >= 0; i-- {
		a.free = append(a.free, chunk.blocks[i])
	}
	a.allocated += bytes
	return nil
}

func (a *SatFixedArena) Fetch() (*SatMemoryBlock, error) {
	if a == nil || a.stopped {
		return nil, ErrInvariant
	}
	if len(a.free) == 0 {
		if err := a.grow(); err != nil {
			return nil, err
		}
	}
	block := a.free[len(a.free)-1]
	a.free = a.free[:len(a.free)-1]
	block.live, block.generation = true, a.generation
	a.used++
	if a.peak < a.used {
		a.peak = a.used
	}
	return block, nil
}

func (a *SatFixedArena) Recycle(block *SatMemoryBlock) error {
	if a == nil || a.stopped || block == nil || block.fixed != a || !block.live || block.generation != a.generation {
		return ErrInvariant
	}
	block.live = false
	a.used--
	a.free = append(a.free, block)
	return nil
}

func (a *SatFixedArena) Restart() error {
	if a == nil || a.stopped {
		return ErrInvariant
	}
	if len(a.chunks) == 0 {
		return nil
	}
	a.generation++
	first := a.chunks[0]
	for _, chunk := range a.chunks {
		for _, block := range chunk.blocks {
			block.live = false
		}
	}
	a.chunks = a.chunks[:1]
	a.allocated, a.used = a.entrySize*a.chunkSize, 0
	a.free = a.free[:0]
	for i := len(first.blocks) - 1; i >= 0; i-- {
		first.blocks[i].generation = a.generation
		a.free = append(a.free, first.blocks[i])
	}
	return nil
}

func (a *SatFixedArena) MemoryUsage() int {
	if a == nil {
		return 0
	}
	return a.allocated
}

func (a *SatFixedArena) Stats() SatFixedMemoryStats {
	if a == nil {
		return SatFixedMemoryStats{}
	}
	return SatFixedMemoryStats{a.entrySize, a.chunkSize, len(a.chunks), len(a.chunks) * a.chunkSize, a.used, a.peak, a.allocated}
}

func (a *SatFixedArena) Stop() {
	if a == nil || a.stopped {
		return
	}
	for _, chunk := range a.chunks {
		for _, block := range chunk.blocks {
			block.live = false
		}
	}
	a.chunks, a.free, a.allocated, a.used, a.stopped = nil, nil, 0, 0, true
}

type SatFlexMemoryStats struct {
	ChunkSize, Chunks, EntriesUsed, MemoryUsed, MemoryAllocated int
}

type satFlexChunk struct {
	data   []byte
	offset int
	blocks []*SatMemoryBlock
}

type SatFlexArena struct {
	chunkSize, entries, used, allocated int
	limits                              SatMemoryLimits
	chunks                              []satFlexChunk
	stopped                             bool
}

func NewSatFlexArena(limits SatMemoryLimits) (*SatFlexArena, error) {
	limits, err := checkSatMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	return &SatFlexArena{chunkSize: 1 << 16, limits: limits}, nil
}

func (a *SatFlexArena) Fetch(bytes int) (*SatMemoryBlock, error) {
	if a == nil || a.stopped {
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
		a.chunks = append(a.chunks, satFlexChunk{data: make([]byte, a.chunkSize)})
		a.allocated += a.chunkSize
	}
	chunk := &a.chunks[len(a.chunks)-1]
	start := chunk.offset
	chunk.offset += bytes
	block := &SatMemoryBlock{data: chunk.data[start:chunk.offset], requested: bytes, allocation: bytes, index: start, live: true}
	chunk.blocks = append(chunk.blocks, block)
	a.entries++
	a.used += bytes
	return block, nil
}

func (a *SatFlexArena) MemoryUsage() int {
	if a == nil {
		return 0
	}
	return a.allocated
}

func (a *SatFlexArena) Stats() SatFlexMemoryStats {
	if a == nil {
		return SatFlexMemoryStats{}
	}
	return SatFlexMemoryStats{a.chunkSize, len(a.chunks), a.entries, a.used, a.allocated}
}

func (a *SatFlexArena) Stop() {
	if a == nil || a.stopped {
		return
	}
	for _, chunk := range a.chunks {
		for _, block := range chunk.blocks {
			block.live = false
		}
	}
	a.chunks, a.entries, a.used, a.allocated, a.stopped = nil, 0, 0, 0, true
}

type SatStepMemoryStats struct {
	Steps, MapSize, LargeBlocks, ClassMemoryAllocated int
	LargeMemoryAllocated                              int
	Classes                                           []SatFixedMemoryStats
}

type SatStepArena struct {
	memories            []*SatFixedArena
	large               []*SatMemoryBlock
	mapSize, largeBytes int
	limits              SatMemoryLimits
	stopped             bool
}

func NewSatStepArena(steps int, limits SatMemoryLimits) (*SatStepArena, error) {
	if steps < 1 || steps > 26 {
		return nil, ErrDimension
	}
	limits, err := checkSatMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	a := &SatStepArena{mapSize: 4 << steps, limits: limits}
	for i := 0; i < steps; i++ {
		memory, makeErr := NewSatFixedArena(8<<i, limits)
		if makeErr != nil {
			return nil, makeErr
		}
		a.memories = append(a.memories, memory)
	}
	return a, nil
}

func (a *SatStepArena) class(bytes int) int {
	class, size := 0, 8
	for bytes > size {
		class, size = class+1, size<<1
	}
	return class
}

func (a *SatStepArena) Fetch(bytes int) (*SatMemoryBlock, error) {
	if a == nil || a.stopped {
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
		block := &SatMemoryBlock{data: make([]byte, bytes), step: a, requested: bytes, allocation: bytes, live: true}
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

func (a *SatStepArena) Recycle(block *SatMemoryBlock, bytes int) error {
	if a == nil || a.stopped {
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
	if bytes > a.mapSize {
		return nil
	}
	return a.memories[a.class(bytes)].Recycle(block)
}

func (a *SatStepArena) Restart() error {
	if a == nil || a.stopped {
		return ErrInvariant
	}
	for _, block := range a.large {
		block.live = false
	}
	a.large, a.largeBytes = nil, 0
	for _, memory := range a.memories {
		if err := memory.Restart(); err != nil {
			return err
		}
	}
	return nil
}

// MemoryUsage matches Sat_MmStepReadMemUsage and excludes direct oversized
// allocations retained by the step arena.
func (a *SatStepArena) MemoryUsage() int {
	if a == nil {
		return 0
	}
	total := 0
	for _, memory := range a.memories {
		total += memory.MemoryUsage()
	}
	return total
}

func (a *SatStepArena) Stats() SatStepMemoryStats {
	if a == nil {
		return SatStepMemoryStats{}
	}
	classes := make([]SatFixedMemoryStats, len(a.memories))
	for i, memory := range a.memories {
		classes[i] = memory.Stats()
	}
	return SatStepMemoryStats{len(a.memories), a.mapSize, len(a.large), a.MemoryUsage(), a.largeBytes, classes}
}

func (a *SatStepArena) Stop() {
	if a == nil || a.stopped {
		return
	}
	for _, memory := range a.memories {
		memory.Stop()
	}
	for _, block := range a.large {
		block.live = false
	}
	a.memories, a.large, a.largeBytes, a.stopped = nil, nil, 0, true
}
