package abc

// Go-only handle and state-model tests for src/misc/mem/mem2.h.

import (
	"errors"
	"slices"
	"testing"
)

func TestMmrFlexPagesHandlesReleaseAndOwnership(t *testing.T) {
	arena, err := NewMmrFlexArena(4, MmrLimits{})
	if err != nil {
		t.Fatal(err)
	}
	first, err := arena.Fetch(3)
	if err != nil || first != 1 {
		t.Fatal(first, err)
	}
	second, err := arena.Fetch(10)
	if err != nil || second != 4 {
		t.Fatal(second, err)
	}
	third, err := arena.Fetch(1)
	if err != nil || third != 17 {
		t.Fatal(third, err)
	}
	if value, readErr := arena.Read(first, 0); readErr != nil || value != ^uint64(0) {
		t.Fatal(value, readErr)
	}
	if err = arena.Write(first, 2, 0x1234); err != nil {
		t.Fatal(err)
	}
	snapshot, err := arena.Snapshot(first)
	if err != nil || !slices.Equal(snapshot, []uint64{^uint64(0), ^uint64(0), 0x1234}) {
		t.Fatal(snapshot, err)
	}
	snapshot[2] = 0
	if value, _ := arena.Read(first, 2); value != 0x1234 {
		t.Fatal("borrowed snapshot", value)
	}
	stats, err := arena.Stats()
	if err != nil || stats != (MmrFlexStats{PageBase: 4, PageWords: 16, Pages: 2, LivePages: 2, Entries: 3, PeakEntries: 3, NextOffset: 2}) {
		t.Fatal(stats, err)
	}
	released, err := arena.Release(third)
	if err != nil || !released {
		t.Fatal(released, err)
	}
	if _, err = arena.Read(first, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if value, readErr := arena.Read(third, 0); readErr != nil || value != ^uint64(0) {
		t.Fatal(value, readErr)
	}
	if released, err = arena.Release(third); err != nil || released {
		t.Fatal(released, err)
	}
	memory, _ := arena.MemoryWords()
	if memory != 16 {
		t.Fatal(memory)
	}
	if err = arena.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = arena.Read(third, 0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestMmrFixedHandleSequenceReuseFillAndStats(t *testing.T) {
	arena, err := NewMmrFixedArena(4, 3, MmrLimits{})
	if err != nil {
		t.Fatal(err)
	}
	handles := make([]int, 5)
	for index := range handles {
		handles[index], err = arena.Fetch()
		if err != nil {
			t.Fatal(index, err)
		}
	}
	if !slices.Equal(handles, []int{1, 4, 7, 10, 17}) {
		t.Fatal(handles)
	}
	for index := 0; index < 3; index++ {
		value, readErr := arena.Read(handles[2], index)
		if readErr != nil || value != ^uint64(0) {
			t.Fatal(index, value, readErr)
		}
	}
	if err = arena.Write(handles[2], 1, 99); err != nil {
		t.Fatal(err)
	}
	if err = arena.Recycle(handles[2]); err != nil {
		t.Fatal(err)
	}
	if _, err = arena.Read(handles[2], 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	reused, err := arena.Fetch()
	if err != nil || reused != handles[2] {
		t.Fatal(reused, err)
	}
	for index := 0; index < 3; index++ {
		value, _ := arena.Read(reused, index)
		if value != ^uint64(0) {
			t.Fatal("recycle fill", index, value)
		}
	}
	stats, err := arena.Stats()
	if err != nil || stats != (MmrFixedStats{PageBase: 4, PageWords: 16, EntryWords: 3, Pages: 2, Entries: 5, PeakEntries: 5, FreeEntries: 3}) {
		t.Fatal(stats, err)
	}
	memory, _ := arena.MemoryWords()
	if memory != 32 {
		t.Fatal(memory)
	}
	if err = arena.Recycle(reused); err != nil {
		t.Fatal(err)
	}
	if err = arena.Recycle(reused); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = arena.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMmrStepPackedHandlesRoutingStatisticsAndLimits(t *testing.T) {
	arena, err := NewMmrStepArena(4, 3, MmrLimits{})
	if err != nil {
		t.Fatal(err)
	}
	sizes := []int{1, 3, 7, 3}
	handles := make([]int, len(sizes))
	for index, size := range sizes {
		handles[index], err = arena.Fetch(size)
		if err != nil || handles[index]&7 != size {
			t.Fatal(index, size, handles[index], err)
		}
		for word := 0; word < size; word++ {
			if err = arena.Write(handles[index], word, uint64(index*10+word)); err != nil {
				t.Fatal(err)
			}
		}
	}
	for index, handle := range handles {
		snapshot, snapshotErr := arena.Snapshot(handle)
		if snapshotErr != nil || len(snapshot) != sizes[index] {
			t.Fatal(index, snapshot, snapshotErr)
		}
		for word, value := range snapshot {
			if value != uint64(index*10+word) {
				t.Fatal(index, word, value)
			}
		}
	}
	stats, err := arena.Stats()
	if err != nil || stats.WordBits != 3 || stats.Mask != 7 || stats.Entries != 4 || stats.PeakEntries != 4 || stats.TotalEntries != 4 || stats.MemoryWords != 48 {
		t.Fatal(stats, err)
	}
	if err = arena.Recycle(handles[1]); err != nil {
		t.Fatal(err)
	}
	reused, err := arena.Fetch(3)
	if err != nil || reused != handles[1] {
		t.Fatal(reused, handles[1], err)
	}
	for word := 0; word < 3; word++ {
		value, _ := arena.Read(reused, word)
		if value != ^uint64(0) {
			t.Fatal(word, value)
		}
	}
	stats, _ = arena.Stats()
	if stats.Entries != 4 || stats.PeakEntries != 4 || stats.TotalEntries != 5 {
		t.Fatal(stats)
	}
	if err = arena.Recycle(handles[1]); err != nil { // the numeric source handle aliases its reused allocation
		t.Fatal(err)
	}
	if err = arena.Recycle(reused); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = arena.Close(); err != nil {
		t.Fatal(err)
	}

	limited, err := NewMmrStepArena(4, 2, MmrLimits{MaxWords: 16, MaxEntries: 10})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = limited.Fetch(1); err != nil {
		t.Fatal(err)
	}
	if _, err = limited.Fetch(2); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	limited.Close()
}

func TestMmrErrorsCancellationAndExtraFlexRestart(t *testing.T) {
	for _, create := range []func() error{
		func() error { _, err := NewMmrFlexArena(1, MmrLimits{}); return err },
		func() error { _, err := NewMmrFixedArena(4, 16, MmrLimits{}); return err },
		func() error { _, err := NewMmrFixedArena(4, 15, MmrLimits{}); return err },
		func() error { _, err := NewMmrStepArena(4, 0, MmrLimits{}); return err },
		func() error { _, err := NewMmrFixedArena(4, 1, MmrLimits{MaxWords: -1}); return err },
	} {
		if err := create(); !errors.Is(err, ErrDimension) {
			t.Fatal(err)
		}
	}
	cancelled, err := NewMmrFixedArena(4, 1, MmrLimits{Stop: func() bool { return false }})
	if err != nil {
		t.Fatal(err)
	}
	cancelled.limits.Stop = func() bool { return true }
	if _, err = cancelled.Fetch(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	limited, err := NewMmrFlexArena(4, MmrLimits{MaxWords: 16, MaxEntries: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = limited.Fetch(1); err != nil {
		t.Fatal(err)
	}
	if _, err = limited.Fetch(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzMmrFixedAgainstHandleModel(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6})
	f.Add([]byte{255, 255, 0, 17})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 128 {
			operations = operations[:128]
		}
		arena, err := NewMmrFixedArena(6, 3, MmrLimits{MaxWords: 1 << 16, MaxEntries: 256})
		if err != nil {
			t.Fatal(err)
		}
		live := []int{}
		values := map[int][3]uint64{}
		for index, operation := range operations {
			if operation&1 == 0 || len(live) == 0 {
				handle, fetchErr := arena.Fetch()
				if fetchErr != nil {
					t.Fatal(fetchErr)
				}
				live = append(live, handle)
				value := [3]uint64{uint64(index), uint64(operation), uint64(index) ^ uint64(operation)}
				values[handle] = value
				for word := range value {
					if writeErr := arena.Write(handle, word, value[word]); writeErr != nil {
						t.Fatal(writeErr)
					}
				}
			} else {
				position := int(operation>>1) % len(live)
				handle := live[position]
				if recycleErr := arena.Recycle(handle); recycleErr != nil {
					t.Fatal(recycleErr)
				}
				delete(values, handle)
				live = append(live[:position], live[position+1:]...)
			}
			for _, handle := range live {
				snapshot, snapshotErr := arena.Snapshot(handle)
				want := values[handle]
				if snapshotErr != nil || !slices.Equal(snapshot, want[:]) {
					t.Fatal(index, handle, snapshot, want, snapshotErr)
				}
			}
		}
		stats, err := arena.Stats()
		if err != nil || stats.Entries != len(live) || stats.PeakEntries < len(live) {
			t.Fatal(stats, len(live), err)
		}
	})
}
