package abc

// Go-only allocator-model, lifecycle, and fuzz tests for extraUtilMemory.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestExtraFixedArenaGeometryReuseRestartAndFormat(t *testing.T) {
	for _, row := range []struct{ entry, chunk int }{{4, 1024}, {64, 1024}, {65, 1008}, {8193, 8}} {
		arena, err := NewExtraFixedArena(row.entry, ExtraMemoryLimits{})
		if err != nil {
			t.Fatal(row, err)
		}
		stats, _ := arena.Stats()
		if stats.EntrySize != row.entry || stats.ChunkSize != row.chunk || stats.Chunks != 0 {
			t.Fatal(row, stats)
		}
	}
	arena, err := NewExtraFixedArena(65, ExtraMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	blocks := make([]*ExtraMemoryBlock, 1009)
	for i := range blocks {
		blocks[i], err = arena.Fetch()
		if err != nil || len(blocks[i].Bytes()) != 65 {
			t.Fatal(i, err)
		}
		blocks[i].Bytes()[0] = byte(i)
	}
	stats, _ := arena.Stats()
	if stats != (ExtraFixedMemoryStats{EntrySize: 65, ChunkSize: 1008, Chunks: 2, EntriesAllocated: 2016, EntriesUsed: 1009, PeakEntries: 1009, MemoryAllocated: 131040}) {
		t.Fatal(stats)
	}
	if err = arena.Recycle(blocks[7]); err != nil {
		t.Fatal(err)
	}
	reused, err := arena.Fetch()
	if err != nil || reused != blocks[7] {
		t.Fatal("free-list reuse", err)
	}
	first := blocks[0]
	if err = arena.Restart(); err != nil || first.Bytes() != nil || reused.Bytes() != nil {
		t.Fatal("restart invalidation", err)
	}
	after, err := arena.Fetch()
	if err != nil || after != first {
		t.Fatal("restart first block", err)
	}
	stats, _ = arena.Stats()
	if stats.Chunks != 1 || stats.EntriesAllocated != 1008 || stats.EntriesUsed != 1 || stats.PeakEntries != 1009 || stats.MemoryAllocated != 65520 {
		t.Fatal(stats)
	}
	usage, _ := arena.MemoryUsage()
	peak, _ := arena.MaxEntriesUsed()
	if usage != 65520 || peak != 1009 {
		t.Fatal(usage, peak)
	}
	text, err := arena.FormatStatistics()
	want := "Fixed memory manager: Entry =    65. Chunk =  1008. Chunks used =     1.\n" +
		"   Entries used =        1. Entries peak =     1009. Memory used =       65. Memory alloc =    65520.\n"
	if err != nil || text != want {
		t.Fatal(text, err)
	}
	if !arena.Close() || arena.Close() || after.Bytes() != nil {
		t.Fatal("close lifecycle")
	}
	if _, err = arena.Fetch(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestExtraFixedArenaErrorsAndCancellation(t *testing.T) {
	if arena, err := NewExtraFixedArena(0, ExtraMemoryLimits{}); arena != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(arena, err)
	}
	if arena, err := NewExtraFixedArena(65, ExtraMemoryLimits{MaxBytes: 65519}); arena != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(arena, err)
	}
	if arena, err := NewExtraFixedArena(8, ExtraMemoryLimits{Stop: func() bool { return true }}); arena != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(arena, err)
	}
	a, _ := NewExtraFixedArena(8, ExtraMemoryLimits{})
	b, _ := NewExtraFixedArena(8, ExtraMemoryLimits{})
	block, _ := a.Fetch()
	if err := b.Recycle(block); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := a.Recycle(block); err != nil {
		t.Fatal(err)
	}
	if err := a.Recycle(block); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	calls := 0
	cancelled, _ := NewExtraFixedArena(8, ExtraMemoryLimits{Stop: func() bool { calls++; return calls > 1 }})
	if _, err := cancelled.Fetch(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	short, _ := NewExtraFixedArena(8, ExtraMemoryLimits{MaxOutputBytes: 1})
	if _, err := short.FormatStatistics(); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func TestExtraFlexArenaChunksResizeAccountingAndFormat(t *testing.T) {
	arena, err := NewExtraFlexArena(ExtraMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	first, _ := arena.Fetch(100)
	second, _ := arena.Fetch(3996)
	third, _ := arena.Fetch(1)
	if len(first.Bytes()) != 100 || len(second.Bytes()) != 3996 || len(third.Bytes()) != 1 {
		t.Fatal("block sizes")
	}
	stats, _ := arena.Stats()
	if stats != (ExtraFlexMemoryStats{ChunkSize: 4096, Chunks: 2, EntriesUsed: 3, MemoryUsed: 4097, MemoryAllocated: 8192}) {
		t.Fatal(stats)
	}
	text, err := arena.FormatStatistics()
	want := "Flexible memory manager: Chunk size = 4096. Chunks used = 2.\n" +
		"   Entries used = 3. Memory used = 4097. Memory alloc = 8192.\n"
	if err != nil || text != want {
		t.Fatal(text, err)
	}
	used, err := arena.UsedMemory()
	if err != nil || used != 4097 {
		t.Fatal(used, err)
	}
	if err = arena.Restart(); err != nil || first.Bytes() != nil || second.Bytes() != nil || third.Bytes() != nil {
		t.Fatal("flex restart invalidation", err)
	}
	stats, _ = arena.Stats()
	if stats.Chunks != 1 || stats.EntriesUsed != 0 || stats.MemoryUsed != 0 || stats.MemoryAllocated != 4096 {
		t.Fatal("flex restart stats", stats)
	}
	afterRestart, err := arena.Fetch(17)
	if err != nil || len(afterRestart.Bytes()) != 17 {
		t.Fatal("flex restart fetch", err)
	}
	large, err := NewExtraFlexArena(ExtraMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	block, err := large.Fetch(5000)
	stats, _ = large.Stats()
	if err != nil || len(block.Bytes()) != 5000 || stats.ChunkSize != 10000 || stats.MemoryAllocated != 10000 {
		t.Fatal(stats, err)
	}
	zero, err := large.Fetch(0)
	if err != nil || zero == nil || len(zero.Bytes()) != 0 {
		t.Fatal(zero, err)
	}
	if !large.Close() || block.Bytes() != nil || zero.Bytes() != nil {
		t.Fatal("flex close")
	}
	if _, err = large.Fetch(1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	limited, _ := NewExtraFlexArena(ExtraMemoryLimits{MaxBytes: 4096, MaxBlocks: 1})
	if _, err = limited.Fetch(4096); err != nil {
		t.Fatal(err)
	}
	if _, err = limited.Fetch(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if arena, err := NewExtraFlexArena(ExtraMemoryLimits{MaxBytes: -1}); arena != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(arena, err)
	}
	arena.Close()
}

func TestExtraStepArenaClassesLargeRetentionAndUsage(t *testing.T) {
	arena, err := NewExtraStepArena(3, ExtraMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	sizes := []int{1, 4, 5, 8, 9, 16, 17, 32, 33}
	blocks := make([]*ExtraMemoryBlock, len(sizes))
	for i, size := range sizes {
		blocks[i], err = arena.Fetch(size)
		if err != nil || len(blocks[i].Bytes()) != size {
			t.Fatal(i, size, err)
		}
	}
	zero, err := arena.Fetch(0)
	if err != nil || zero != nil {
		t.Fatal(zero, err)
	}
	stats, err := arena.Stats()
	if err != nil || stats.Steps != 3 || stats.MapSize != 32 || stats.LargeBlocks != 1 || stats.LargeMemoryAllocated != 33 || stats.ClassMemoryAllocated != 57344 || len(stats.Classes) != 3 {
		t.Fatal(stats, err)
	}
	if got := []int{stats.Classes[0].EntrySize, stats.Classes[1].EntrySize, stats.Classes[2].EntrySize}; !slices.Equal(got, []int{8, 16, 32}) {
		t.Fatal(got)
	}
	usage, _ := arena.MemoryUsage()
	if usage != 57344 {
		t.Fatal(usage)
	}
	if err = arena.Recycle(blocks[4], 9); err != nil || blocks[4].Bytes() != nil {
		t.Fatal(err)
	}
	reused, err := arena.Fetch(10)
	if err != nil || reused != blocks[4] || len(reused.Bytes()) != 10 {
		t.Fatal("class reuse", err)
	}
	if err = arena.Recycle(blocks[8], 33); err != nil || len(blocks[8].Bytes()) != 33 {
		t.Fatal("large allocations remain owned", err)
	}
	if err = arena.Recycle(blocks[8], 33); err != nil {
		t.Fatal("large recycle is source-compatible no-op", err)
	}
	if err = arena.Recycle(reused, 9); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = arena.Recycle(nil, 0); err != nil {
		t.Fatal(err)
	}
	other, _ := NewExtraStepArena(2, ExtraMemoryLimits{})
	if err = other.Recycle(blocks[0], 1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if !arena.Close() || arena.Close() {
		t.Fatal("step close")
	}
	for _, block := range blocks {
		if block.Bytes() != nil {
			t.Fatal("live block after close")
		}
	}
	if _, err = arena.Fetch(1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	other.Close()
}

func TestExtraStepArenaLimitsAndCancellation(t *testing.T) {
	for _, steps := range []int{0, 27} {
		if arena, err := NewExtraStepArena(steps, ExtraMemoryLimits{}); arena != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(steps, arena, err)
		}
	}
	wide, err := NewExtraStepArena(26, ExtraMemoryLimits{})
	if err != nil || !wide.Close() {
		t.Fatal("deferred high-class allocation", err)
	}
	arena, err := NewExtraStepArena(2, ExtraMemoryLimits{MaxBytes: 16384, MaxBlocks: 8})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := arena.Fetch(1); err != nil {
		t.Fatal(err)
	}
	if _, err := arena.Fetch(9000); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := arena.Fetch(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	calls := 0
	cancelled, _ := NewExtraStepArena(2, ExtraMemoryLimits{Stop: func() bool { calls++; return calls > 1 }})
	if _, err := cancelled.Fetch(1); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func FuzzExtraMemory(f *testing.F) {
	for _, seed := range []struct {
		operations uint8
		sizes      uint64
	}{{1, 1}, {17, 0x2110090805040100}, {255, ^uint64(0)}} {
		f.Add(seed.operations, seed.sizes)
	}
	f.Fuzz(func(t *testing.T, operations uint8, sizes uint64) {
		fixed, err := NewExtraFixedArena(17, ExtraMemoryLimits{MaxBytes: 1 << 20, MaxBlocks: 1 << 16})
		flex, err2 := NewExtraFlexArena(ExtraMemoryLimits{MaxBytes: 1 << 20, MaxBlocks: 1 << 16})
		step, err3 := NewExtraStepArena(6, ExtraMemoryLimits{MaxBytes: 8 << 20, MaxBlocks: 1 << 16})
		if err != nil || err2 != nil || err3 != nil {
			t.Fatal(err, err2, err3)
		}
		live := []*ExtraMemoryBlock{}
		peak := 0
		flexUsed, flexEntries := 0, 0
		for operation := 0; operation < int(operations%64)+1; operation++ {
			size := int(sizes>>uint((operation&7)*8)&255) + 1
			if operation%3 == 0 && len(live) > 0 {
				index := int((sizes >> uint((operation&7)*8)) % uint64(len(live)))
				block := live[index]
				if err := fixed.Recycle(block); err != nil || block.Bytes() != nil {
					t.Fatal("fixed recycle", err)
				}
				live[index] = live[len(live)-1]
				live = live[:len(live)-1]
			} else {
				block, fetchErr := fixed.Fetch()
				if fetchErr != nil || len(block.Bytes()) != 17 {
					t.Fatal("fixed fetch", fetchErr)
				}
				live = append(live, block)
				if peak < len(live) {
					peak = len(live)
				}
			}
			flexBlock, flexErr := flex.Fetch(size)
			stepBlock, stepErr := step.Fetch(size)
			if flexErr != nil || stepErr != nil || len(flexBlock.Bytes()) != size || len(stepBlock.Bytes()) != size {
				t.Fatal(size, flexErr, stepErr)
			}
			flexUsed += size
			flexEntries++
			for i := range flexBlock.Bytes() {
				flexBlock.Bytes()[i] = byte(operation)
				stepBlock.Bytes()[i] = byte(operation ^ i)
			}
			if size <= step.mapSize {
				if recycleErr := step.Recycle(stepBlock, size); recycleErr != nil || stepBlock.Bytes() != nil {
					t.Fatal("step recycle", recycleErr)
				}
			}
		}
		fixedStats, _ := fixed.Stats()
		flexStats, _ := flex.Stats()
		if fixedStats.EntriesUsed != len(live) || fixedStats.PeakEntries != peak || flexStats.EntriesUsed != flexEntries || flexStats.MemoryUsed != flexUsed {
			t.Fatal(fixedStats, len(live), peak, flexStats, flexEntries, flexUsed)
		}
		fixed.Close()
		flex.Close()
		step.Close()
		for _, block := range live {
			if block.Bytes() != nil {
				t.Fatal("fixed close invalidation")
			}
		}
	})
}
