package abc

// Tests for owned arenas derived from aig/aig/aigMem.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestAigMemoryNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_memory.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Fixed struct {
			Reuse                       []int
			BeforeRestart, AfterRestart AigFixedMemoryStats
			RestartFirst                int
		}
		Flex struct {
			States       [][]int
			AfterRestart []int
		}
		Step struct {
			Classes                                 []int
			ZeroIsNull, Reuse, MapSize, LargeChunks int
			MemoryUsage                             int
			Fixed                                   []AigFixedMemoryStats
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	fixed, err := NewAigFixedArena(16, 64, AigMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	blocks := make([]*AigMemoryBlock, 20)
	for i := range blocks {
		blocks[i], err = fixed.Fetch()
		if err != nil || len(blocks[i].Bytes()) != 16 {
			t.Fatal(i, err)
		}
	}
	if fixed.Recycle(blocks[3]) != nil || fixed.Recycle(blocks[7]) != nil {
		t.Fatal("recycle")
	}
	reused0, _ := fixed.Fetch()
	reused1, _ := fixed.Fetch()
	if (reused0 == blocks[7]) != (fixture.Fixed.Reuse[0] != 0) || (reused1 == blocks[3]) != (fixture.Fixed.Reuse[1] != 0) || fixed.Stats() != fixture.Fixed.BeforeRestart {
		t.Fatal(reused0 == blocks[7], reused1 == blocks[3], fixed.Stats(), fixture.Fixed)
	}
	if fixed.Restart() != nil {
		t.Fatal("restart")
	}
	after, _ := fixed.Fetch()
	if (after == blocks[0]) != (fixture.Fixed.RestartFirst != 0) || fixed.Stats() != fixture.Fixed.AfterRestart {
		t.Fatal(after == blocks[0], fixed.Stats(), fixture.Fixed.AfterRestart)
	}

	flex, err := NewAigFlexArena(AigMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for i, size := range []int{1, 7, 0, 262140, 10, 300000} {
		block, err := flex.Fetch(size)
		stats := flex.Stats()
		got := []int{stats.EntriesUsed, stats.ChunkSize, stats.Chunks, stats.MemoryUsed, stats.MemoryAllocated}
		if err != nil || len(block.Bytes()) != size || !reflect.DeepEqual(got, fixture.Flex.States[i]) {
			t.Fatal(i, got, fixture.Flex.States[i], err)
		}
	}
	if err = flex.Restart(); err != nil {
		t.Fatal(err)
	}
	stats := flex.Stats()
	if got := []int{stats.EntriesUsed, stats.ChunkSize, stats.Chunks, stats.MemoryUsed, stats.MemoryAllocated}; !reflect.DeepEqual(got, fixture.Flex.AfterRestart) {
		t.Fatal(got, fixture.Flex.AfterRestart)
	}

	step, err := NewAigStepArena(3, AigMemoryLimits{})
	if err != nil {
		t.Fatal(err)
	}
	sizes := []int{0, 1, 4, 5, 8, 9, 16, 17, 32, 33, 100}
	stepBlocks := make([]*AigMemoryBlock, len(sizes))
	classes := make([]int, len(sizes))
	for i, size := range sizes {
		stepBlocks[i], err = step.Fetch(size)
		if err != nil {
			t.Fatal(i, err)
		}
		if size == 0 {
			classes[i] = -1
		} else if size > step.mapSize {
			classes[i] = len(step.memories)
		} else {
			classes[i] = step.class(size)
		}
	}
	if err = step.Recycle(stepBlocks[1], sizes[1]); err != nil {
		t.Fatal(err)
	}
	stepReuse, _ := step.Fetch(1)
	stepStats := step.Stats()
	if !reflect.DeepEqual(classes, fixture.Step.Classes) || (stepBlocks[0] == nil) != (fixture.Step.ZeroIsNull != 0) ||
		(stepReuse == stepBlocks[1]) != (fixture.Step.Reuse != 0) || stepStats.MapSize != fixture.Step.MapSize ||
		stepStats.LargeBlocks != fixture.Step.LargeChunks || stepStats.MemoryAllocated != fixture.Step.MemoryUsage ||
		!reflect.DeepEqual(stepStats.Classes, fixture.Step.Fixed) {
		t.Fatal(classes, stepStats, fixture.Step)
	}
	fixed.Stop()
	flex.Stop()
	step.Stop()
}

func TestAigMemoryErrorsLimitsAndLifetime(t *testing.T) {
	for _, args := range []struct{ entry, maximum int }{{0, 1}, {8, -1}} {
		if arena, err := NewAigFixedArena(args.entry, args.maximum, AigMemoryLimits{}); arena != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(args, arena, err)
		}
	}
	if arena, err := NewAigFixedArena(16, 64, AigMemoryLimits{MaxBytes: 127}); arena != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(arena, err)
	}
	a, _ := NewAigFixedArena(8, 64, AigMemoryLimits{})
	b, _ := NewAigFixedArena(8, 64, AigMemoryLimits{})
	block, _ := a.Fetch()
	if err := b.Recycle(block); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err := a.Recycle(block); err != nil || block.Bytes() != nil {
		t.Fatal(err, block.Bytes())
	}
	if err := a.Recycle(block); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	block, _ = a.Fetch()
	a.Restart()
	if block.Bytes() != nil {
		t.Fatal("restart did not invalidate")
	}
	a.Stop()
	if _, err := a.Fetch(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	flex, _ := NewAigFlexArena(AigMemoryLimits{MaxBytes: 300000, MaxBlocks: 2})
	first, err := flex.Fetch(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = flex.Fetch(299999); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	flex.Restart()
	if first.Bytes() != nil {
		t.Fatal("flex restart did not invalidate")
	}
	if _, err = flex.Fetch(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}

	for _, steps := range []int{0, -1, 27} {
		if arena, err := NewAigStepArena(steps, AigMemoryLimits{}); arena != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(steps, arena, err)
		}
	}
	step, _ := NewAigStepArena(2, AigMemoryLimits{})
	small, _ := step.Fetch(7)
	if len(small.Bytes()) != 7 || step.Recycle(small, 8) == nil || step.Recycle(small, 7) != nil || small.Bytes() != nil {
		t.Fatal("step recycle")
	}
	large, _ := step.Fetch(100)
	if err = step.Recycle(large, 100); err != nil || large.Bytes() == nil {
		t.Fatal(err)
	}
	step.Stop()
	if large.Bytes() != nil {
		t.Fatal("step stop did not invalidate")
	}
	if _, err = step.Fetch(1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzAigFixedArena(f *testing.F) {
	f.Add([]byte{0, 1, 2, 0, 3, 1, 0, 2})
	f.Fuzz(func(t *testing.T, data []byte) {
		arena, err := NewAigFixedArena(16, 64, AigMemoryLimits{MaxBytes: 1 << 20, MaxBlocks: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		live := make([]*AigMemoryBlock, 0)
		peak := 0
		for _, operation := range data {
			if operation&3 == 0 || len(live) == 0 {
				block, err := arena.Fetch()
				if err != nil {
					t.Fatal(err)
				}
				block.Bytes()[0] = operation
				live = append(live, block)
				if peak < len(live) {
					peak = len(live)
				}
			} else {
				index := int(operation) % len(live)
				block := live[index]
				if err := arena.Recycle(block); err != nil || block.Bytes() != nil {
					t.Fatal(err)
				}
				live[index] = live[len(live)-1]
				live = live[:len(live)-1]
			}
			stats := arena.Stats()
			if stats.EntriesUsed != len(live) || stats.PeakEntries != peak || stats.EntriesAllocated != stats.Chunks*stats.ChunkSize || stats.MemoryAllocated != stats.EntriesAllocated*16 {
				t.Fatal(stats, len(live), peak)
			}
		}
	})
}

func FuzzAigFlexStepArena(f *testing.F) {
	f.Add([]byte{0, 1, 7, 15, 31, 63, 127, 255})
	f.Fuzz(func(t *testing.T, data []byte) {
		flex, err := NewAigFlexArena(AigMemoryLimits{MaxBytes: 16 << 20, MaxBlocks: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		step, err := NewAigStepArena(6, AigMemoryLimits{MaxBytes: 16 << 20, MaxBlocks: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		var borrowed []*AigMemoryBlock
		for _, operation := range data {
			if operation == 255 {
				if err := flex.Restart(); err != nil {
					t.Fatal(err)
				}
				for _, block := range borrowed {
					if block.Bytes() != nil {
						t.Fatal("stale flex block")
					}
				}
				borrowed = borrowed[:0]
				continue
			}
			size := int(operation) * 17
			block, err := flex.Fetch(size)
			if err != nil || len(block.Bytes()) != size {
				t.Fatal(size, err)
			}
			borrowed = append(borrowed, block)
			stepBlock, err := step.Fetch(size)
			if err != nil || (size == 0) != (stepBlock == nil) || (stepBlock != nil && len(stepBlock.Bytes()) != size) {
				t.Fatal(size, stepBlock, err)
			}
			if size != 0 && size <= step.mapSize && operation&1 == 0 {
				if err := step.Recycle(stepBlock, size); err != nil || stepBlock.Bytes() != nil {
					t.Fatal(size, err)
				}
				reused, err := step.Fetch(size)
				if err != nil || reused != stepBlock {
					t.Fatal(size, reused == stepBlock, err)
				}
			}
			stats := flex.Stats()
			if stats.MemoryUsed < 0 || stats.MemoryUsed > stats.MemoryAllocated || stats.EntriesUsed < len(borrowed) {
				t.Fatal(stats, len(borrowed))
			}
		}
	})
}
