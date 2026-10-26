package abc

// Allocation/free recording from src/misc/util/utilMem.c and utilMem.h.

import (
	"fmt"
	"slices"
)

// MemoryRecordID is a stable caller-assigned identity. It replaces comparison
// and sorting of unrelated C pointers, which is neither needed nor safe in Go.
type MemoryRecordID uint64

type MemoryRecorderOptions struct {
	// MaxEvents bounds the combined allocation and free event count in one
	// recording window. Zero selects one million events.
	MaxEvents int
}

type MemoryRecycleReport struct {
	AllocationEvents int
	FreeEvents       int
	// Unreleased is sorted and owns its storage. Repeated IDs represent
	// repeated allocation events that were not paired with a free event.
	Unreleased []MemoryRecordID
}

// MemoryRecorder is an explicit owned replacement for utilMem.c's writable
// process globals. The zero value is not initialized; use NewMemoryRecorder.
type MemoryRecorder struct {
	maxEvents   int
	active      bool
	interrupted bool
	allocations []MemoryRecordID
	frees       []MemoryRecordID
}

func NewMemoryRecorder(options MemoryRecorderOptions) (*MemoryRecorder, error) {
	if options.MaxEvents < 0 {
		return nil, ErrDimension
	}
	if options.MaxEvents == 0 {
		options.MaxEvents = 1_000_000
	}
	return &MemoryRecorder{maxEvents: options.MaxEvents}, nil
}

func (recorder *MemoryRecorder) IsActive() bool {
	return recorder != nil && recorder.active
}

func (recorder *MemoryRecorder) Start() error {
	if recorder == nil || recorder.maxEvents <= 0 {
		return fmt.Errorf("%w: uninitialized memory recorder", ErrInvariant)
	}
	if recorder.active {
		return fmt.Errorf("%w: memory recorder is already active", ErrInvariant)
	}
	recorder.allocations = recorder.allocations[:0]
	recorder.frees = recorder.frees[:0]
	recorder.active = true
	return nil
}

func (recorder *MemoryRecorder) room() error {
	if recorder == nil || recorder.maxEvents <= 0 {
		return fmt.Errorf("%w: uninitialized memory recorder", ErrInvariant)
	}
	if !recorder.active {
		return nil
	}
	if len(recorder.allocations)+len(recorder.frees) >= recorder.maxEvents {
		return ErrLimit
	}
	return nil
}

// RecordAllocation retains id only while recording is active and returns id,
// mirroring Util_MemRecAlloc's expression-friendly result.
func (recorder *MemoryRecorder) RecordAllocation(id MemoryRecordID) (MemoryRecordID, error) {
	if err := recorder.room(); err != nil {
		return id, err
	}
	if recorder.active {
		recorder.allocations = append(recorder.allocations, id)
	}
	return id, nil
}

// RecordFree retains id only while recording is active and returns id,
// mirroring Util_MemRecFree's expression-friendly result.
func (recorder *MemoryRecorder) RecordFree(id MemoryRecordID) (MemoryRecordID, error) {
	if err := recorder.room(); err != nil {
		return id, err
	}
	if recorder.active {
		recorder.frees = append(recorder.frees, id)
	}
	return id, nil
}

func (recorder *MemoryRecorder) EventCounts() (allocations, frees int, err error) {
	if recorder == nil || recorder.maxEvents <= 0 {
		return 0, 0, fmt.Errorf("%w: uninitialized memory recorder", ErrInvariant)
	}
	return len(recorder.allocations), len(recorder.frees), nil
}

// Recycle pairs the sorted allocation/free multisets and starts a fresh
// recording window. C attempted to free the unmatched pointers; Go ownership
// makes that invalid, so callers receive stable IDs for explicit cleanup or
// diagnostics. A free without a matching allocation is rejected atomically.
func (recorder *MemoryRecorder) Recycle() (MemoryRecycleReport, error) {
	if recorder == nil || recorder.maxEvents <= 0 || !recorder.active {
		return MemoryRecycleReport{}, fmt.Errorf("%w: memory recorder is not active", ErrInvariant)
	}
	allocations := slices.Clone(recorder.allocations)
	frees := slices.Clone(recorder.frees)
	slices.Sort(allocations)
	slices.Sort(frees)
	report := MemoryRecycleReport{
		AllocationEvents: len(allocations),
		FreeEvents:       len(frees),
		Unreleased:       make([]MemoryRecordID, 0, len(allocations)),
	}
	allocationIndex, freeIndex := 0, 0
	for allocationIndex < len(allocations) && freeIndex < len(frees) {
		switch {
		case allocations[allocationIndex] == frees[freeIndex]:
			allocationIndex++
			freeIndex++
		case allocations[allocationIndex] < frees[freeIndex]:
			report.Unreleased = append(report.Unreleased, allocations[allocationIndex])
			allocationIndex++
		default:
			return MemoryRecycleReport{}, fmt.Errorf("%w: free event %d has no allocation", ErrInvariant, frees[freeIndex])
		}
	}
	if freeIndex != len(frees) {
		return MemoryRecycleReport{}, fmt.Errorf("%w: free event %d has no allocation", ErrInvariant, frees[freeIndex])
	}
	report.Unreleased = append(report.Unreleased, allocations[allocationIndex:]...)
	recorder.allocations = recorder.allocations[:0]
	recorder.frees = recorder.frees[:0]
	return report, nil
}

func (recorder *MemoryRecorder) Quit() error {
	if recorder == nil || recorder.maxEvents <= 0 || !recorder.active {
		return fmt.Errorf("%w: memory recorder is not active", ErrInvariant)
	}
	recorder.allocations = nil
	recorder.frees = nil
	recorder.active = false
	return nil
}

func (recorder *MemoryRecorder) RequestInterrupt() error {
	if recorder == nil || recorder.maxEvents <= 0 {
		return fmt.Errorf("%w: uninitialized memory recorder", ErrInvariant)
	}
	recorder.interrupted = true
	return nil
}

func (recorder *MemoryRecorder) ClearInterrupt() error {
	if recorder == nil || recorder.maxEvents <= 0 {
		return fmt.Errorf("%w: uninitialized memory recorder", ErrInvariant)
	}
	recorder.interrupted = false
	return nil
}

func (recorder *MemoryRecorder) Interrupted() bool {
	return recorder != nil && recorder.interrupted
}
