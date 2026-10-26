package abc

// Port of src/aig/aig/aigMem.c using checked owned byte arenas.

type AigMemoryLimits struct {
	MaxBytes, MaxBlocks int // zero: 1 GiB and 16M blocks
}

func checkAigMemoryLimits(limits AigMemoryLimits) (AigMemoryLimits, error) {
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

// AigMemoryBlock is a checked borrowed arena allocation. Bytes returns nil
// after recycle or restart. The block itself must not be copied.
type AigMemoryBlock struct {
	data                         []byte
	fixed                        *AigFixedArena
	step                         *AigStepArena
	requested, allocation, index int
	generation                   uint64
	live                         bool
}

func (b *AigMemoryBlock) Bytes() []byte {
	if b == nil || !b.live || (b.fixed != nil && b.generation != b.fixed.generation) {
		return nil
	}
	return b.data[:b.requested:b.requested]
}

type AigFixedMemoryStats struct {
	EntrySize, ChunkSize, Chunks, EntriesAllocated int
	EntriesUsed, PeakEntries, MemoryAllocated      int
}

type aigFixedChunk struct {
	data   []byte
	blocks []*AigMemoryBlock
}

type AigFixedArena struct {
	entrySize, chunkSize, used, peak, allocated int
	limits                                      AigMemoryLimits
	chunks                                      []aigFixedChunk
	free                                        []*AigMemoryBlock
	generation                                  uint64
	stopped                                     bool
}

func NewAigFixedArena(entrySize, entriesMaximum int, limits AigMemoryLimits) (*AigFixedArena, error) {
	if entrySize < 1 || entriesMaximum < 0 {
		return nil, ErrDimension
	}
	limits, err := checkAigMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	chunkSize := entriesMaximum / 8
	if chunkSize < 8 {
		chunkSize = 8
	}
	if _, err = checkedSize(0, entrySize, chunkSize, limits.MaxBytes); err != nil {
		return nil, err
	}
	return &AigFixedArena{entrySize: entrySize, chunkSize: chunkSize, limits: limits, generation: 1}, nil
}

func (a *AigFixedArena) grow() error {
	if a == nil || a.stopped {
		return ErrInvariant
	}
	bytes, err := checkedSize(0, a.entrySize, a.chunkSize, a.limits.MaxBytes-a.allocated)
	if err != nil || a.chunkSize > a.limits.MaxBlocks-a.Stats().EntriesAllocated {
		return ErrLimit
	}
	chunk := aigFixedChunk{data: make([]byte, bytes), blocks: make([]*AigMemoryBlock, a.chunkSize)}
	for i := 0; i < a.chunkSize; i++ {
		block := &AigMemoryBlock{data: chunk.data[i*a.entrySize : (i+1)*a.entrySize], fixed: a, requested: a.entrySize, allocation: a.entrySize, index: i, generation: a.generation}
		chunk.blocks[i] = block
	}
	a.chunks = append(a.chunks, chunk)
	for i := len(chunk.blocks) - 1; i >= 0; i-- {
		a.free = append(a.free, chunk.blocks[i])
	}
	a.allocated += bytes
	return nil
}

func (a *AigFixedArena) Fetch() (*AigMemoryBlock, error) {
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

func (a *AigFixedArena) Recycle(block *AigMemoryBlock) error {
	if a == nil || a.stopped || block == nil || block.fixed != a || !block.live || block.generation != a.generation {
		return ErrInvariant
	}
	block.live = false
	a.used--
	a.free = append(a.free, block)
	return nil
}

func (a *AigFixedArena) Restart() error {
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
	a.allocated = a.entrySize * a.chunkSize
	a.used = 0
	a.free = a.free[:0]
	for i := len(first.blocks) - 1; i >= 0; i-- {
		block := first.blocks[i]
		block.generation = a.generation
		a.free = append(a.free, block)
	}
	return nil
}

func (a *AigFixedArena) Stats() AigFixedMemoryStats {
	if a == nil {
		return AigFixedMemoryStats{}
	}
	return AigFixedMemoryStats{a.entrySize, a.chunkSize, len(a.chunks), len(a.chunks) * a.chunkSize, a.used, a.peak, a.allocated}
}

func (a *AigFixedArena) Stop() {
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

type AigFlexMemoryStats struct {
	ChunkSize, Chunks, EntriesUsed, MemoryUsed, MemoryAllocated int
}

type aigFlexChunk struct {
	data   []byte
	offset int
	blocks []*AigMemoryBlock
}

type AigFlexArena struct {
	chunkSize, entries, used, allocated int
	limits                              AigMemoryLimits
	chunks                              []aigFlexChunk
	stopped                             bool
}

func NewAigFlexArena(limits AigMemoryLimits) (*AigFlexArena, error) {
	limits, err := checkAigMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	return &AigFlexArena{chunkSize: 1 << 18, limits: limits}, nil
}

func (a *AigFlexArena) Fetch(bytes int) (*AigMemoryBlock, error) {
	if a == nil || a.stopped || bytes < 0 {
		return nil, ErrDimension
	}
	if a.entries == a.limits.MaxBlocks || bytes > a.limits.MaxBytes-a.allocated {
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
		a.chunks = append(a.chunks, aigFlexChunk{data: make([]byte, a.chunkSize)})
		a.allocated += a.chunkSize
	}
	chunk := &a.chunks[len(a.chunks)-1]
	start := chunk.offset
	chunk.offset += bytes
	block := &AigMemoryBlock{data: chunk.data[start:chunk.offset], requested: bytes, allocation: bytes, index: start, live: true}
	chunk.blocks = append(chunk.blocks, block)
	a.entries++
	a.used += bytes
	return block, nil
}

func (a *AigFlexArena) Restart() error {
	if a == nil || a.stopped {
		return ErrInvariant
	}
	if len(a.chunks) == 0 {
		return nil
	}
	for _, chunk := range a.chunks {
		for _, block := range chunk.blocks {
			block.live = false
		}
	}
	first := a.chunks[0]
	first.offset, first.blocks = 0, nil
	a.chunks = []aigFlexChunk{first}
	// Preserve native's nominal accounting after a large allocation, while
	// Fetch still checks the retained first slice's real capacity.
	a.allocated = a.chunkSize
	a.entries, a.used = 0, 0
	return nil
}

func (a *AigFlexArena) Stats() AigFlexMemoryStats {
	if a == nil {
		return AigFlexMemoryStats{}
	}
	return AigFlexMemoryStats{a.chunkSize, len(a.chunks), a.entries, a.used, a.allocated}
}

func (a *AigFlexArena) Stop() {
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

type AigStepMemoryStats struct {
	Steps, MapSize, LargeBlocks, MemoryAllocated int
	Classes                                      []AigFixedMemoryStats
}

type AigStepArena struct {
	memories            []*AigFixedArena
	large               []*AigMemoryBlock
	mapSize, largeBytes int
	limits              AigMemoryLimits
	stopped             bool
}

func NewAigStepArena(steps int, limits AigMemoryLimits) (*AigStepArena, error) {
	if steps < 1 || steps > 26 {
		return nil, ErrDimension
	}
	limits, err := checkAigMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	a := &AigStepArena{mapSize: 4 << steps, limits: limits}
	for i := 0; i < steps; i++ {
		memory, err := NewAigFixedArena(8<<i, 1<<13, limits)
		if err != nil {
			return nil, err
		}
		a.memories = append(a.memories, memory)
	}
	return a, nil
}

func (a *AigStepArena) class(bytes int) int {
	class, size := 0, 8
	for bytes > size {
		class, size = class+1, size<<1
	}
	return class
}

func (a *AigStepArena) Fetch(bytes int) (*AigMemoryBlock, error) {
	if a == nil || a.stopped || bytes < 0 {
		return nil, ErrDimension
	}
	if bytes == 0 {
		return nil, nil
	}
	if bytes > a.mapSize {
		stats := a.Stats()
		if len(a.large) == a.limits.MaxBlocks || bytes > a.limits.MaxBytes-stats.MemoryAllocated-a.largeBytes {
			return nil, ErrLimit
		}
		block := &AigMemoryBlock{data: make([]byte, bytes), step: a, requested: bytes, allocation: bytes, live: true}
		a.large = append(a.large, block)
		a.largeBytes += bytes
		return block, nil
	}
	class := a.class(bytes)
	memory := a.memories[class]
	if len(memory.free) == 0 {
		growth := memory.entrySize * memory.chunkSize
		if growth > a.limits.MaxBytes-a.Stats().MemoryAllocated-a.largeBytes {
			return nil, ErrLimit
		}
	}
	block, err := memory.Fetch()
	if err != nil {
		return nil, err
	}
	block.step, block.requested = a, bytes
	return block, nil
}

func (a *AigStepArena) Recycle(block *AigMemoryBlock, bytes int) error {
	if a == nil || a.stopped || bytes < 0 {
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

func (a *AigStepArena) Stats() AigStepMemoryStats {
	if a == nil {
		return AigStepMemoryStats{}
	}
	total := 0
	for _, memory := range a.memories {
		total += memory.Stats().MemoryAllocated
	}
	classes := make([]AigFixedMemoryStats, len(a.memories))
	for i, memory := range a.memories {
		classes[i] = memory.Stats()
	}
	return AigStepMemoryStats{len(a.memories), a.mapSize, len(a.large), total, classes}
}

func (a *AigStepArena) Stop() {
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
