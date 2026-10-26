package abc

// Stand-alone tests for the owned port of src/sat/bsat/satMem.c.

import (
	"errors"
	"testing"
)

func TestSatFixedArenaSizingReuseRestartAndAccounting(t *testing.T) {
	cases := []struct {
		entry, chunk int
	}{{1, 1024}, {63, 1024}, {64, 1024}, {65, 1008}, {8192, 8}, {10000, 8}}
	for _, test := range cases {
		a, err := NewSatFixedArena(test.entry, SatMemoryLimits{})
		if err != nil || a.Stats().ChunkSize != test.chunk {
			t.Fatalf("entry %d: stats=%+v err=%v", test.entry, a.Stats(), err)
		}
	}
	a, err := NewSatFixedArena(65, SatMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	blocks := make([]*SatMemoryBlock, a.chunkSize+1)
	for i := range blocks {
		blocks[i], err = a.Fetch()
		if err != nil || len(blocks[i].Bytes()) != 65 {
			t.Fatal(i, err)
		}
		blocks[i].Bytes()[0] = byte(i)
	}
	if got := a.Stats(); got.Chunks != 2 || got.EntriesAllocated != 2*a.chunkSize || got.EntriesUsed != len(blocks) || got.PeakEntries != len(blocks) || got.MemoryAllocated != 2*65*a.chunkSize || a.MemoryUsage() != got.MemoryAllocated {
		t.Fatal(got)
	}
	old := blocks[17]
	if err = a.Recycle(old); err != nil || old.Bytes() != nil {
		t.Fatal(err)
	}
	reused, err := a.Fetch()
	if err != nil || reused != old {
		t.Fatal("fixed free list is not LIFO", err)
	}
	stale := blocks[len(blocks)-1]
	if err = a.Restart(); err != nil || stale.Bytes() != nil {
		t.Fatal(err)
	}
	if got := a.Stats(); got.Chunks != 1 || got.EntriesAllocated != a.chunkSize || got.EntriesUsed != 0 || got.PeakEntries != len(blocks) || got.MemoryAllocated != 65*a.chunkSize {
		t.Fatal(got)
	}
	first, err := a.Fetch()
	if err != nil || first != blocks[0] {
		t.Fatal("restart order", err)
	}
	a.Stop()
	if first.Bytes() != nil {
		t.Fatal("stop retained borrowed bytes")
	}
	a.Stop()
	if _, err = a.Fetch(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSatFlexArenaBoundariesGrowthAndLifetime(t *testing.T) {
	a, err := NewSatFlexArena(SatMemoryLimits{MaxBytes: 500000, MaxBlocks: 8})
	if err != nil {
		t.Fatal(err)
	}
	sizes := []int{0, 1, (1 << 16) - 1, 1, 70000}
	want := []SatFlexMemoryStats{
		{1 << 16, 1, 1, 0, 1 << 16},
		{1 << 16, 1, 2, 1, 1 << 16},
		{1 << 16, 1, 3, 1 << 16, 1 << 16},
		{1 << 16, 2, 4, (1 << 16) + 1, 2 << 16},
		{140000, 3, 5, (1 << 16) + 70001, (2 << 16) + 140000},
	}
	var blocks []*SatMemoryBlock
	for i, size := range sizes {
		block, fetchErr := a.Fetch(size)
		if fetchErr != nil || len(block.Bytes()) != size || a.Stats() != want[i] {
			t.Fatalf("%d: block=%v stats=%+v want=%+v err=%v", i, block, a.Stats(), want[i], fetchErr)
		}
		blocks = append(blocks, block)
	}
	if a.MemoryUsage() != want[len(want)-1].MemoryAllocated {
		t.Fatal(a.MemoryUsage())
	}
	if _, err = a.Fetch(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	a.Stop()
	for _, block := range blocks {
		if block.Bytes() != nil {
			t.Fatal("stop retained flex block")
		}
	}
}

func TestSatStepArenaClassesLargeRetentionRestartAndUsage(t *testing.T) {
	a, err := NewSatStepArena(3, SatMemoryLimits{MaxBytes: 2 << 20, MaxBlocks: 1 << 16})
	if err != nil {
		t.Fatal(err)
	}
	if a.mapSize != 32 {
		t.Fatal(a.mapSize)
	}
	sizes := []int{0, 1, 4, 5, 8, 9, 16, 17, 32, 33, 100}
	wantClass := []int{-1, 0, 0, 0, 0, 1, 1, 2, 2, 3, 3}
	blocks := make([]*SatMemoryBlock, len(sizes))
	for i, size := range sizes {
		blocks[i], err = a.Fetch(size)
		if err != nil || (size == 0) != (blocks[i] == nil) {
			t.Fatal(i, err)
		}
		class := -1
		if size > a.mapSize {
			class = len(a.memories)
		} else if size > 0 {
			class = a.class(size)
		}
		if class != wantClass[i] || (blocks[i] != nil && len(blocks[i].Bytes()) != size) {
			t.Fatal(i, class, blocks[i])
		}
	}
	stats := a.Stats()
	if stats.Steps != 3 || stats.MapSize != 32 || stats.LargeBlocks != 2 || stats.LargeMemoryAllocated != 133 || stats.ClassMemoryAllocated != (8+16+32)*1024 || a.MemoryUsage() != stats.ClassMemoryAllocated {
		t.Fatal(stats)
	}
	if err = a.Recycle(blocks[1], 1); err != nil {
		t.Fatal(err)
	}
	reused, err := a.Fetch(1)
	if err != nil || reused != blocks[1] {
		t.Fatal("class reuse", err)
	}
	large := blocks[len(blocks)-1]
	if err = a.Recycle(large, 100); err != nil || large.Bytes() == nil {
		t.Fatal("native large recycle must retain until restart", err)
	}
	if err = a.Restart(); err != nil || large.Bytes() != nil || reused.Bytes() != nil {
		t.Fatal("restart lifetime", err)
	}
	stats = a.Stats()
	if stats.LargeBlocks != 0 || stats.LargeMemoryAllocated != 0 || stats.ClassMemoryAllocated != (8+16+32)*1024 {
		t.Fatal(stats)
	}
}

func TestSatMemoryErrorsAndLimits(t *testing.T) {
	if a, err := NewSatFixedArena(0, SatMemoryLimits{}); a != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(a, err)
	}
	if a, err := NewSatFixedArena(16, SatMemoryLimits{MaxBytes: 100}); a != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(a, err)
	}
	if a, err := NewSatFlexArena(SatMemoryLimits{MaxBytes: -1}); a != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(a, err)
	}
	for _, steps := range []int{-1, 0, 27} {
		if a, err := NewSatStepArena(steps, SatMemoryLimits{}); a != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(steps, a, err)
		}
	}
	a, _ := NewSatFixedArena(8, SatMemoryLimits{})
	b, _ := NewSatFixedArena(8, SatMemoryLimits{})
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
	step, _ := NewSatStepArena(2, SatMemoryLimits{})
	small, _ := step.Fetch(7)
	if err := step.Recycle(small, 8); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := step.Recycle(nil, 0); err != nil {
		t.Fatal(err)
	}
	if err := step.Recycle(small, -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	step.Stop()
	if small.Bytes() != nil {
		t.Fatal("stop retained small block")
	}
}

func FuzzSatFixedArenaModel(f *testing.F) {
	f.Add([]byte{0, 1, 0, 2, 3, 1, 0, 4})
	f.Fuzz(func(t *testing.T, operations []byte) {
		a, err := NewSatFixedArena(17, SatMemoryLimits{MaxBytes: 1 << 20, MaxBlocks: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		var live []*SatMemoryBlock
		peak := 0
		for _, operation := range operations {
			if len(live) == 0 || operation&1 == 0 {
				block, fetchErr := a.Fetch()
				if fetchErr != nil {
					t.Fatal(fetchErr)
				}
				block.Bytes()[0] = operation
				live = append(live, block)
				if peak < len(live) {
					peak = len(live)
				}
			} else {
				i := int(operation) % len(live)
				block := live[i]
				if err := a.Recycle(block); err != nil || block.Bytes() != nil {
					t.Fatal(err)
				}
				live[i] = live[len(live)-1]
				live = live[:len(live)-1]
			}
			stats := a.Stats()
			if stats.EntriesUsed != len(live) || stats.PeakEntries != peak || stats.EntriesAllocated != stats.Chunks*stats.ChunkSize || stats.MemoryAllocated != stats.EntriesAllocated*17 {
				t.Fatal(stats, len(live), peak)
			}
		}
	})
}

func FuzzSatFlexStepArenaModel(f *testing.F) {
	f.Add([]byte{0, 1, 7, 15, 31, 63, 127, 255})
	f.Fuzz(func(t *testing.T, operations []byte) {
		flex, _ := NewSatFlexArena(SatMemoryLimits{MaxBytes: 16 << 20, MaxBlocks: 1 << 16})
		step, _ := NewSatStepArena(6, SatMemoryLimits{MaxBytes: 16 << 20, MaxBlocks: 1 << 16})
		used, entries := 0, 0
		for _, operation := range operations {
			size := int(operation) * 17
			block, err := flex.Fetch(size)
			if err != nil || len(block.Bytes()) != size {
				t.Fatal(size, err)
			}
			used, entries = used+size, entries+1
			if stats := flex.Stats(); stats.MemoryUsed != used || stats.EntriesUsed != entries || stats.MemoryUsed > stats.MemoryAllocated {
				t.Fatal(stats, used, entries)
			}
			stepBlock, err := step.Fetch(size)
			if err != nil || (size == 0) != (stepBlock == nil) || (stepBlock != nil && len(stepBlock.Bytes()) != size) {
				t.Fatal(size, err)
			}
			if size > 0 && size <= step.mapSize && operation&1 == 0 {
				if err := step.Recycle(stepBlock, size); err != nil || stepBlock.Bytes() != nil {
					t.Fatal(size, err)
				}
				reused, err := step.Fetch(size)
				if err != nil || reused != stepBlock {
					t.Fatal(size, err)
				}
			}
		}
	})
}
