package abc

// Go-only tests for src/sat/msat/msatMem.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestMsatFixedArenaSizingReuseRestartAccountingAndDiagnostic(t *testing.T) {
	for _, test := range []struct{ entry, chunk int }{{4, 1024}, {63, 1024}, {64, 1024}, {65, 1008}, {8192, 8}, {10000, 8}} {
		arena, err := NewMsatFixedArena(test.entry, MsatMemoryLimits{})
		if err != nil || arena.Stats().ChunkSize != test.chunk {
			t.Fatalf("entry %d: stats=%+v err=%v", test.entry, arena.Stats(), err)
		}
	}
	arena, err := NewMsatFixedArena(65, MsatMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	blocks := make([]*MsatMemoryBlock, arena.chunkSize+1)
	for index := range blocks {
		blocks[index], err = arena.Fetch()
		if err != nil || len(blocks[index].Bytes()) != 65 {
			t.Fatal(index, err)
		}
		blocks[index].Bytes()[0] = byte(index)
	}
	stats := arena.Stats()
	if stats.Chunks != 2 || stats.EntriesAllocated != 2*arena.chunkSize || stats.EntriesUsed != len(blocks) || stats.PeakEntries != len(blocks) || stats.MemoryAllocated != 2*65*arena.chunkSize || arena.MemoryUsage() != stats.MemoryAllocated {
		t.Fatal(stats)
	}
	if got := arena.Diagnostic(); got != "Fixed memory manager: Entry =    65. Chunk =  1008. Chunks used =     2.\n   Entries used =     1009. Entries peak =     1009. Memory used =    65585. Memory alloc =   131040.\n" {
		t.Fatal(got)
	}
	recycled := blocks[17]
	if err = arena.Recycle(recycled); err != nil || recycled.Bytes() != nil {
		t.Fatal(err)
	}
	reused, err := arena.Fetch()
	if err != nil || reused != recycled {
		t.Fatal("free list must be LIFO", err)
	}
	stale := blocks[len(blocks)-1]
	if err = arena.Restart(); err != nil || stale.Bytes() != nil {
		t.Fatal(err)
	}
	stats = arena.Stats()
	if stats.Chunks != 1 || stats.EntriesAllocated != arena.chunkSize || stats.EntriesUsed != 0 || stats.PeakEntries != len(blocks) || stats.MemoryAllocated != 65*arena.chunkSize {
		t.Fatal(stats)
	}
	first, err := arena.Fetch()
	if err != nil || first != blocks[0] {
		t.Fatal("restart order", err)
	}
	arena.Close()
	arena.Close()
	if first.Bytes() != nil {
		t.Fatal("close retained block")
	}
	if _, err = arena.Fetch(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestMsatFlexArenaFourKiBChunksGrowthAndLifetime(t *testing.T) {
	arena, err := NewMsatFlexArena(MsatMemoryLimits{MaxBytes: 40000, MaxBlocks: 8})
	if err != nil {
		t.Fatal(err)
	}
	sizes := []int{0, 1, 4095, 1, 5000}
	want := []MsatFlexMemoryStats{
		{4096, 1, 1, 0, 4096},
		{4096, 1, 2, 1, 4096},
		{4096, 1, 3, 4096, 4096},
		{4096, 2, 4, 4097, 8192},
		{10000, 3, 5, 9097, 18192},
	}
	var blocks []*MsatMemoryBlock
	for index, size := range sizes {
		block, fetchErr := arena.Fetch(size)
		if fetchErr != nil || len(block.Bytes()) != size || arena.Stats() != want[index] {
			t.Fatalf("%d: stats=%+v want=%+v err=%v", index, arena.Stats(), want[index], fetchErr)
		}
		blocks = append(blocks, block)
	}
	if got := arena.Diagnostic(); got != "Flexible memory manager: Chunk size = 10000. Chunks used = 3.\n   Entries used = 5. Memory used = 9097. Memory alloc = 18192.\n" {
		t.Fatal(got)
	}
	if _, err = arena.Fetch(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	arena.Close()
	for _, block := range blocks {
		if block.Bytes() != nil {
			t.Fatal("close retained flex block")
		}
	}
}

func TestMsatStepArenaClassesImmediateLargeRecycleAndUsage(t *testing.T) {
	arena, err := NewMsatStepArena(3, MsatMemoryLimits{MaxBytes: 2 << 20, MaxBlocks: 1 << 16})
	if err != nil {
		t.Fatal(err)
	}
	sizes := []int{0, 1, 4, 5, 8, 9, 16, 17, 32, 33, 100}
	wantClass := []int{-1, 0, 0, 0, 0, 1, 1, 2, 2, 3, 3}
	blocks := make([]*MsatMemoryBlock, len(sizes))
	for index, size := range sizes {
		blocks[index], err = arena.Fetch(size)
		if err != nil || (size == 0) != (blocks[index] == nil) {
			t.Fatal(index, err)
		}
		class := -1
		if size > arena.mapSize {
			class = len(arena.memories)
		} else if size > 0 {
			class = arena.class(size)
		}
		if class != wantClass[index] || blocks[index] != nil && len(blocks[index].Bytes()) != size {
			t.Fatal(index, class, blocks[index])
		}
	}
	stats := arena.Stats()
	if stats.Steps != 3 || stats.MapSize != 32 || stats.LargeBlocks != 2 || stats.LargeMemoryAllocated != 133 || stats.ClassMemoryAllocated != (8+16+32)*1024 || arena.MemoryUsage() != stats.ClassMemoryAllocated {
		t.Fatal(stats)
	}
	if err = arena.Recycle(blocks[1], 1); err != nil {
		t.Fatal(err)
	}
	reused, err := arena.Fetch(1)
	if err != nil || reused != blocks[1] {
		t.Fatal("class reuse", err)
	}
	large := blocks[len(blocks)-1]
	if err = arena.Recycle(large, 100); err != nil || large.Bytes() != nil {
		t.Fatal("large blocks must be invalid immediately", err)
	}
	stats = arena.Stats()
	if stats.LargeBlocks != 1 || stats.LargeMemoryAllocated != 33 {
		t.Fatal(stats)
	}
	if err = arena.Recycle(blocks[len(blocks)-2], 33); err != nil {
		t.Fatal(err)
	}
	stats = arena.Stats()
	if stats.LargeBlocks != 0 || stats.LargeMemoryAllocated != 0 {
		t.Fatal(stats)
	}
	arena.Close()
	if reused.Bytes() != nil {
		t.Fatal("close retained class block")
	}
}

func TestMsatMemoryErrorsLimitsOwnershipAndAtomicity(t *testing.T) {
	for _, entry := range []int{-1, 0, 3} {
		if arena, err := NewMsatFixedArena(entry, MsatMemoryLimits{}); arena != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(entry, arena, err)
		}
	}
	if arena, err := NewMsatFixedArena(16, MsatMemoryLimits{MaxBytes: 100}); arena != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(arena, err)
	}
	if arena, err := NewMsatFlexArena(MsatMemoryLimits{MaxBytes: -1}); arena != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(arena, err)
	}
	for _, steps := range []int{-1, 0, 27} {
		if arena, err := NewMsatStepArena(steps, MsatMemoryLimits{}); arena != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(steps, arena, err)
		}
	}
	first, _ := NewMsatFixedArena(8, MsatMemoryLimits{})
	second, _ := NewMsatFixedArena(8, MsatMemoryLimits{})
	block, _ := first.Fetch()
	if err := second.Recycle(block); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := first.Recycle(block); err != nil {
		t.Fatal(err)
	}
	if err := first.Recycle(block); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	step, _ := NewMsatStepArena(2, MsatMemoryLimits{})
	small, _ := step.Fetch(7)
	before := step.Stats()
	if err := step.Recycle(small, 8); !errors.Is(err, ErrInvariant) || step.Stats().ClassMemoryAllocated != before.ClassMemoryAllocated {
		t.Fatal(err, step.Stats(), before)
	}
	if err := step.Recycle(nil, 0); err != nil {
		t.Fatal(err)
	}
	if err := step.Recycle(small, -1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := (*MsatFixedArena)(nil).Fetch(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzMsatFixedArenaModel(f *testing.F) {
	f.Add([]byte{0, 1, 0, 2, 3, 1, 0, 4})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 5000 {
			operations = operations[:5000]
		}
		arena, err := NewMsatFixedArena(17, MsatMemoryLimits{MaxBytes: 2 << 20, MaxBlocks: 1 << 17})
		if err != nil {
			t.Fatal(err)
		}
		var live []*MsatMemoryBlock
		peak := 0
		chunks := 0
		for _, operation := range operations {
			switch operation % 5 {
			case 0, 1, 2:
				block, fetchErr := arena.Fetch()
				if fetchErr != nil || block == nil || len(block.Bytes()) != 17 {
					t.Fatal(fetchErr)
				}
				block.Bytes()[0] = operation
				live = append(live, block)
				if peak < len(live) {
					peak = len(live)
				}
				if len(live) > chunks*arena.chunkSize {
					chunks++
				}
			case 3:
				if len(live) != 0 {
					position := int(operation) % len(live)
					block := live[position]
					if err = arena.Recycle(block); err != nil || block.Bytes() != nil {
						t.Fatal(err)
					}
					live = slices.Delete(live, position, position+1)
				}
			case 4:
				if err = arena.Restart(); err != nil {
					t.Fatal(err)
				}
				for _, block := range live {
					if block.Bytes() != nil {
						t.Fatal("restart retained block")
					}
				}
				live = nil
				if chunks > 0 {
					chunks = 1
				}
			}
			stats := arena.Stats()
			if stats.EntriesUsed != len(live) || stats.PeakEntries != peak || stats.Chunks != chunks || stats.MemoryAllocated != chunks*17*arena.chunkSize {
				t.Fatal(operation, stats, len(live), peak, chunks)
			}
		}
	})
}

func FuzzMsatStepArenaClassModel(f *testing.F) {
	f.Add([]byte{0, 1, 4, 5, 8, 9, 16, 17, 32, 33, 255})
	f.Fuzz(func(t *testing.T, sizes []byte) {
		if len(sizes) > 1000 {
			sizes = sizes[:1000]
		}
		arena, err := NewMsatStepArena(4, MsatMemoryLimits{MaxBytes: 16 << 20, MaxBlocks: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		for _, raw := range sizes {
			size := int(raw)
			block, fetchErr := arena.Fetch(size)
			if fetchErr != nil {
				t.Fatal(fetchErr)
			}
			if size == 0 {
				if block != nil {
					t.Fatal(block)
				}
				continue
			}
			if len(block.Bytes()) != size {
				t.Fatal(size, len(block.Bytes()))
			}
			if size <= arena.mapSize {
				class := arena.class(size)
				allocation := 8 << class
				if block.allocation != allocation || allocation < size || class > 0 && allocation/2 >= size {
					t.Fatal(size, class, block.allocation)
				}
			} else if block.allocation != size {
				t.Fatal(size, block.allocation)
			}
			if err = arena.Recycle(block, size); err != nil || block.Bytes() != nil {
				t.Fatal(size, err)
			}
		}
		if stats := arena.Stats(); stats.LargeBlocks != 0 || stats.LargeMemoryAllocated != 0 {
			t.Fatal(stats)
		}
	})
}

func FuzzMsatFlexArenaModel(f *testing.F) {
	f.Add([]byte{0, 0, 1, 0, 255, 15, 1, 0, 0, 32})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 2000 {
			data = data[:2000]
		}
		arena, err := NewMsatFlexArena(MsatMemoryLimits{MaxBytes: 128 << 20, MaxBlocks: 1000})
		if err != nil {
			t.Fatal(err)
		}
		chunkSize, chunks, offset, allocated, entries, used := 4096, 0, 0, 0, 0, 0
		for index := 0; index+1 < len(data); index += 2 {
			size := int(data[index]) | int(data[index+1])<<8
			if chunks == 0 || size > chunkSize-offset {
				if size > chunkSize {
					chunkSize = 2 * size
				}
				chunks++
				offset = 0
				allocated += chunkSize
			}
			offset += size
			entries++
			used += size
			block, fetchErr := arena.Fetch(size)
			if fetchErr != nil || len(block.Bytes()) != size {
				t.Fatal(size, fetchErr)
			}
			want := MsatFlexMemoryStats{chunkSize, chunks, entries, used, allocated}
			if got := arena.Stats(); got != want {
				t.Fatal(size, got, want)
			}
		}
	})
}
