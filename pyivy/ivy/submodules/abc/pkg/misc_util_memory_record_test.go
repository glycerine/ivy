package abc

// Standalone tests for src/misc/util/utilMem.c and utilMem.h behavior.

import (
	"errors"
	"slices"
	"testing"
)

func TestMemoryRecorderLifecycleAndMultiset(t *testing.T) {
	recorder, err := NewMemoryRecorder(MemoryRecorderOptions{MaxEvents: 16})
	if err != nil || recorder.IsActive() {
		t.Fatal(recorder, err)
	}
	if id, err := recorder.RecordAllocation(91); err != nil || id != 91 {
		t.Fatal(id, err)
	}
	if allocations, frees, err := recorder.EventCounts(); err != nil || allocations != 0 || frees != 0 {
		t.Fatal(allocations, frees, err)
	}
	if err := recorder.Start(); err != nil || !recorder.IsActive() {
		t.Fatal(err)
	}
	if err := recorder.Start(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	for _, id := range []MemoryRecordID{9, 2, 9, 7, 2} {
		if returned, err := recorder.RecordAllocation(id); err != nil || returned != id {
			t.Fatal(id, returned, err)
		}
	}
	for _, id := range []MemoryRecordID{9, 2, 7} {
		if returned, err := recorder.RecordFree(id); err != nil || returned != id {
			t.Fatal(id, returned, err)
		}
	}
	report, err := recorder.Recycle()
	if err != nil || report.AllocationEvents != 5 || report.FreeEvents != 3 || !slices.Equal(report.Unreleased, []MemoryRecordID{2, 9}) {
		t.Fatal(report, err)
	}
	report.Unreleased[0] = 100
	if allocations, frees, err := recorder.EventCounts(); err != nil || allocations != 0 || frees != 0 {
		t.Fatal(allocations, frees, err)
	}
	if empty, err := recorder.Recycle(); err != nil || empty.AllocationEvents != 0 || empty.FreeEvents != 0 || len(empty.Unreleased) != 0 {
		t.Fatal(empty, err)
	}
	if err := recorder.RequestInterrupt(); err != nil || !recorder.Interrupted() {
		t.Fatal(err)
	}
	if err := recorder.ClearInterrupt(); err != nil || recorder.Interrupted() {
		t.Fatal(err)
	}
	if err := recorder.Quit(); err != nil || recorder.IsActive() {
		t.Fatal(err)
	}
	if err := recorder.Quit(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestMemoryRecorderErrorsAndAtomicity(t *testing.T) {
	if recorder, err := NewMemoryRecorder(MemoryRecorderOptions{MaxEvents: -1}); recorder != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(recorder, err)
	}
	var nilRecorder *MemoryRecorder
	if err := nilRecorder.Start(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err := nilRecorder.RecordAllocation(1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	recorder, _ := NewMemoryRecorder(MemoryRecorderOptions{MaxEvents: 3})
	if err := recorder.Start(); err != nil {
		t.Fatal(err)
	}
	_, _ = recorder.RecordAllocation(3)
	_, _ = recorder.RecordFree(2)
	if _, err := recorder.Recycle(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if allocations, frees, err := recorder.EventCounts(); err != nil || allocations != 1 || frees != 1 {
		t.Fatal(allocations, frees, err)
	}
	if _, err := recorder.RecordAllocation(2); err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.RecordAllocation(4); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	report, err := recorder.Recycle()
	if err != nil || !slices.Equal(report.Unreleased, []MemoryRecordID{3}) {
		t.Fatal(report, err)
	}
}

func FuzzMemoryRecorderMultiset(f *testing.F) {
	f.Add([]byte{0, 1, 0, 2, 1, 1, 2, 0, 3, 2})
	f.Add([]byte{1, 9, 0, 9, 0, 9, 1, 9, 2})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			data = data[:512]
		}
		recorder, err := NewMemoryRecorder(MemoryRecorderOptions{MaxEvents: len(data)/2 + 1})
		if err != nil || recorder.Start() != nil {
			t.Fatal(err)
		}
		allocations := make(map[MemoryRecordID]int)
		allocationEvents, freeEvents := 0, 0
		for index := 0; index+1 < len(data); index += 2 {
			operation, id := data[index]%3, MemoryRecordID(data[index+1]%32)
			switch operation {
			case 0:
				if _, err := recorder.RecordAllocation(id); err != nil {
					t.Fatal(err)
				}
				allocations[id]++
				allocationEvents++
			case 1:
				if allocations[id] == 0 {
					continue
				}
				if _, err := recorder.RecordFree(id); err != nil {
					t.Fatal(err)
				}
				allocations[id]--
				freeEvents++
			case 2:
				if recorder.Interrupted() {
					if err := recorder.ClearInterrupt(); err != nil {
						t.Fatal(err)
					}
				} else if err := recorder.RequestInterrupt(); err != nil {
					t.Fatal(err)
				}
			}
		}
		expected := make([]MemoryRecordID, 0)
		for id, count := range allocations {
			for range count {
				expected = append(expected, id)
			}
		}
		slices.Sort(expected)
		report, err := recorder.Recycle()
		if err != nil || report.AllocationEvents != allocationEvents || report.FreeEvents != freeEvents || !slices.Equal(report.Unreleased, expected) {
			t.Fatal(report, expected, err)
		}
	})
}
