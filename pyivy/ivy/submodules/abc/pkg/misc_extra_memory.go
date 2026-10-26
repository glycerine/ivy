package abc

// Checked memory-manager facade for src/misc/extra/extraUtilMemory.c.

import "fmt"

type ExtraMemoryLimits struct {
	MaxBytes, MaxBlocks, MaxOutputBytes int
	Stop                                func() bool
}

type ExtraMemoryBlock = SatMemoryBlock
type ExtraFixedMemoryStats = SatFixedMemoryStats
type ExtraFlexMemoryStats = SatFlexMemoryStats

func normalizeExtraMemoryLimits(limits ExtraMemoryLimits) (ExtraMemoryLimits, error) {
	if limits.MaxBytes < 0 || limits.MaxBlocks < 0 || limits.MaxOutputBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 1 << 30
	}
	if limits.MaxBlocks == 0 {
		limits.MaxBlocks = 16 << 20
	}
	if limits.MaxOutputBytes == 0 {
		limits.MaxOutputBytes = 16 << 20
	}
	return limits, nil
}

func extraMemoryStopped(limits ExtraMemoryLimits) bool {
	return limits.Stop != nil && limits.Stop()
}

func extraSatMemoryLimits(limits ExtraMemoryLimits) SatMemoryLimits {
	return SatMemoryLimits{MaxBytes: limits.MaxBytes, MaxBlocks: limits.MaxBlocks}
}

type ExtraFixedArena struct {
	arena  *SatFixedArena
	limits ExtraMemoryLimits
	closed bool
}

func NewExtraFixedArena(entrySize int, limits ExtraMemoryLimits) (*ExtraFixedArena, error) {
	limits, err := normalizeExtraMemoryLimits(limits)
	if err != nil || entrySize < 1 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if extraMemoryStopped(limits) {
		return nil, ErrCancelled
	}
	arena, err := NewSatFixedArena(entrySize, extraSatMemoryLimits(limits))
	if err != nil {
		return nil, err
	}
	return &ExtraFixedArena{arena: arena, limits: limits}, nil
}

func (arena *ExtraFixedArena) check() error {
	if arena == nil {
		return ErrDimension
	}
	if arena.closed || arena.arena == nil || arena.arena.stopped {
		return ErrInvariant
	}
	return nil
}

func (arena *ExtraFixedArena) Fetch() (*ExtraMemoryBlock, error) {
	if err := arena.check(); err != nil {
		return nil, err
	}
	if extraMemoryStopped(arena.limits) {
		return nil, ErrCancelled
	}
	return arena.arena.Fetch()
}

func (arena *ExtraFixedArena) Recycle(block *ExtraMemoryBlock) error {
	if err := arena.check(); err != nil {
		return err
	}
	if extraMemoryStopped(arena.limits) {
		return ErrCancelled
	}
	return arena.arena.Recycle(block)
}

func (arena *ExtraFixedArena) Restart() error {
	if err := arena.check(); err != nil {
		return err
	}
	if extraMemoryStopped(arena.limits) {
		return ErrCancelled
	}
	return arena.arena.Restart()
}

func (arena *ExtraFixedArena) Stats() (ExtraFixedMemoryStats, error) {
	if err := arena.check(); err != nil {
		return ExtraFixedMemoryStats{}, err
	}
	return arena.arena.Stats(), nil
}

func (arena *ExtraFixedArena) MemoryUsage() (int, error) {
	stats, err := arena.Stats()
	return stats.MemoryAllocated, err
}

func (arena *ExtraFixedArena) MaxEntriesUsed() (int, error) {
	stats, err := arena.Stats()
	return stats.PeakEntries, err
}

func (arena *ExtraFixedArena) FormatStatistics() (string, error) {
	stats, err := arena.Stats()
	if err != nil {
		return "", err
	}
	text := fmt.Sprintf("Fixed memory manager: Entry = %5d. Chunk = %5d. Chunks used = %5d.\n   Entries used = %8d. Entries peak = %8d. Memory used = %8d. Memory alloc = %8d.\n", stats.EntrySize, stats.ChunkSize, stats.Chunks, stats.EntriesUsed, stats.PeakEntries, stats.EntrySize*stats.EntriesUsed, stats.MemoryAllocated)
	if len(text) > arena.limits.MaxOutputBytes {
		return "", ErrLimit
	}
	return text, nil
}

func (arena *ExtraFixedArena) Close() bool {
	if arena == nil || arena.closed {
		return false
	}
	if arena.arena != nil {
		arena.arena.Stop()
	}
	arena.arena, arena.closed = nil, true
	return true
}

type ExtraFlexArena struct {
	arena  *SatFlexArena
	limits ExtraMemoryLimits
	closed bool
}

func NewExtraFlexArena(limits ExtraMemoryLimits) (*ExtraFlexArena, error) {
	limits, err := normalizeExtraMemoryLimits(limits)
	if err != nil {
		return nil, err
	}
	if extraMemoryStopped(limits) {
		return nil, ErrCancelled
	}
	arena, err := NewSatFlexArena(extraSatMemoryLimits(limits))
	if err != nil {
		return nil, err
	}
	arena.chunkSize = 1 << 12
	return &ExtraFlexArena{arena: arena, limits: limits}, nil
}

func (arena *ExtraFlexArena) check() error {
	if arena == nil {
		return ErrDimension
	}
	if arena.closed || arena.arena == nil || arena.arena.stopped {
		return ErrInvariant
	}
	return nil
}

func (arena *ExtraFlexArena) Fetch(bytes int) (*ExtraMemoryBlock, error) {
	if err := arena.check(); err != nil {
		return nil, err
	}
	if extraMemoryStopped(arena.limits) {
		return nil, ErrCancelled
	}
	return arena.arena.Fetch(bytes)
}

func (arena *ExtraFlexArena) Stats() (ExtraFlexMemoryStats, error) {
	if err := arena.check(); err != nil {
		return ExtraFlexMemoryStats{}, err
	}
	return arena.arena.Stats(), nil
}

func (arena *ExtraFlexArena) MemoryUsage() (int, error) {
	stats, err := arena.Stats()
	return stats.MemoryAllocated, err
}

// UsedMemory returns the sum of requested byte counts, matching
// Mem_FlexReadMemUsage in src/misc/mem/mem.c.
func (arena *ExtraFlexArena) UsedMemory() (int, error) {
	stats, err := arena.Stats()
	return stats.MemoryUsed, err
}

// Restart retains only the first chunk and invalidates every prior block.
func (arena *ExtraFlexArena) Restart() error {
	if err := arena.check(); err != nil {
		return err
	}
	if extraMemoryStopped(arena.limits) {
		return ErrCancelled
	}
	if len(arena.arena.chunks) == 0 {
		return nil
	}
	for _, chunk := range arena.arena.chunks {
		for _, block := range chunk.blocks {
			block.live = false
		}
	}
	first := arena.arena.chunks[0]
	first.offset = 0
	first.blocks = nil
	arena.arena.chunks = []satFlexChunk{first}
	arena.arena.entries = 0
	arena.arena.used = 0
	arena.arena.allocated = len(first.data)
	return nil
}

func (arena *ExtraFlexArena) FormatStatistics() (string, error) {
	stats, err := arena.Stats()
	if err != nil {
		return "", err
	}
	text := fmt.Sprintf("Flexible memory manager: Chunk size = %d. Chunks used = %d.\n   Entries used = %d. Memory used = %d. Memory alloc = %d.\n", stats.ChunkSize, stats.Chunks, stats.EntriesUsed, stats.MemoryUsed, stats.MemoryAllocated)
	if len(text) > arena.limits.MaxOutputBytes {
		return "", ErrLimit
	}
	return text, nil
}

func (arena *ExtraFlexArena) Close() bool {
	if arena == nil || arena.closed {
		return false
	}
	if arena.arena != nil {
		arena.arena.Stop()
	}
	arena.arena, arena.closed = nil, true
	return true
}

type ExtraStepMemoryStats struct {
	Steps, MapSize, LargeBlocks                int
	ClassMemoryAllocated, LargeMemoryAllocated int
	Classes                                    []ExtraFixedMemoryStats
}

type extraStepAllocation struct {
	requested, class int
	large            bool
}

type ExtraStepArena struct {
	memories   []*ExtraFixedArena
	large      []*ExtraMemoryBlock
	live       map[*ExtraMemoryBlock]extraStepAllocation
	mapSize    int
	largeBytes int
	limits     ExtraMemoryLimits
	closed     bool
}

func NewExtraStepArena(steps int, limits ExtraMemoryLimits) (*ExtraStepArena, error) {
	limits, err := normalizeExtraMemoryLimits(limits)
	if err != nil || steps < 1 || steps > 26 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if extraMemoryStopped(limits) {
		return nil, ErrCancelled
	}
	arena := &ExtraStepArena{mapSize: 4 << uint(steps), limits: limits, live: make(map[*ExtraMemoryBlock]extraStepAllocation)}
	for class := 0; class < steps; class++ {
		entrySize := 8 << uint(class)
		chunkSize := 1 << 10
		if entrySize*chunkSize >= 1<<16 {
			chunkSize = (1 << 16) / entrySize
		}
		if chunkSize < 8 {
			chunkSize = 8
		}
		localBytes := entrySize * chunkSize
		if localBytes < limits.MaxBytes {
			localBytes = limits.MaxBytes
		}
		localBlocks := chunkSize
		if localBlocks < limits.MaxBlocks {
			localBlocks = limits.MaxBlocks
		}
		inner, memoryErr := NewSatFixedArena(entrySize, SatMemoryLimits{MaxBytes: localBytes, MaxBlocks: localBlocks})
		if memoryErr != nil {
			arena.Close()
			return nil, memoryErr
		}
		memory := &ExtraFixedArena{arena: inner, limits: limits}
		arena.memories = append(arena.memories, memory)
	}
	return arena, nil
}

func (arena *ExtraStepArena) check() error {
	if arena == nil {
		return ErrDimension
	}
	if arena.closed || len(arena.memories) == 0 || arena.live == nil {
		return ErrInvariant
	}
	return nil
}

func (arena *ExtraStepArena) class(bytes int) int {
	class, size := 0, 8
	for bytes > size {
		class, size = class+1, size<<1
	}
	return class
}

func (arena *ExtraStepArena) allocated() int {
	total := arena.largeBytes
	for _, memory := range arena.memories {
		total += memory.arena.allocated
	}
	return total
}

func (arena *ExtraStepArena) Fetch(bytes int) (*ExtraMemoryBlock, error) {
	if err := arena.check(); err != nil {
		return nil, err
	}
	if bytes < 0 {
		return nil, ErrDimension
	}
	if extraMemoryStopped(arena.limits) {
		return nil, ErrCancelled
	}
	if bytes == 0 {
		return nil, nil
	}
	if len(arena.live) >= arena.limits.MaxBlocks {
		return nil, ErrLimit
	}
	if bytes > arena.mapSize {
		if bytes > arena.limits.MaxBytes-arena.allocated() {
			return nil, ErrLimit
		}
		block := &ExtraMemoryBlock{data: make([]byte, bytes), requested: bytes, allocation: bytes, live: true}
		arena.large = append(arena.large, block)
		arena.largeBytes += bytes
		arena.live[block] = extraStepAllocation{requested: bytes, class: -1, large: true}
		return block, nil
	}
	class := arena.class(bytes)
	memory := arena.memories[class]
	if len(memory.arena.free) == 0 {
		growth := memory.arena.entrySize * memory.arena.chunkSize
		if growth > arena.limits.MaxBytes-arena.allocated() {
			return nil, ErrLimit
		}
	}
	block, err := memory.Fetch()
	if err != nil {
		return nil, err
	}
	block.requested = bytes
	arena.live[block] = extraStepAllocation{requested: bytes, class: class}
	return block, nil
}

func (arena *ExtraStepArena) Recycle(block *ExtraMemoryBlock, bytes int) error {
	if err := arena.check(); err != nil {
		return err
	}
	if bytes < 0 {
		return ErrDimension
	}
	if extraMemoryStopped(arena.limits) {
		return ErrCancelled
	}
	if bytes == 0 {
		if block != nil {
			return ErrInvariant
		}
		return nil
	}
	allocation, found := arena.live[block]
	if !found || allocation.requested != bytes {
		return ErrInvariant
	}
	if allocation.large {
		return nil
	}
	if err := arena.memories[allocation.class].Recycle(block); err != nil {
		return err
	}
	delete(arena.live, block)
	return nil
}

func (arena *ExtraStepArena) Stats() (ExtraStepMemoryStats, error) {
	if err := arena.check(); err != nil {
		return ExtraStepMemoryStats{}, err
	}
	result := ExtraStepMemoryStats{Steps: len(arena.memories), MapSize: arena.mapSize, LargeBlocks: len(arena.large), LargeMemoryAllocated: arena.largeBytes, Classes: make([]ExtraFixedMemoryStats, len(arena.memories))}
	for i, memory := range arena.memories {
		stats, err := memory.Stats()
		if err != nil {
			return ExtraStepMemoryStats{}, err
		}
		result.Classes[i] = stats
		result.ClassMemoryAllocated += stats.MemoryAllocated
	}
	return result, nil
}

func (arena *ExtraStepArena) MemoryUsage() (int, error) {
	stats, err := arena.Stats()
	return stats.ClassMemoryAllocated, err
}

func (arena *ExtraStepArena) Close() bool {
	if arena == nil || arena.closed {
		return false
	}
	for _, memory := range arena.memories {
		memory.Close()
	}
	for _, block := range arena.large {
		block.live = false
	}
	arena.memories, arena.large, arena.live = nil, nil, nil
	arena.largeBytes, arena.closed = 0, true
	return true
}
