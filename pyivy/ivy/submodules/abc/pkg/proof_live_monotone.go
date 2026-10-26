package abc

// Portable helpers from src/proof/live/monotone.c and
// src/proof/live/disjunctiveMonotone.c.

import (
	"math/bits"
	"strings"
)

type LiveMonotoneOptions struct {
	MaxItems int
	MaxWork  int
	MaxBytes int
	Stop     func() bool
}

type LiveMonotoneOutputs struct {
	Pending   int
	Hints     []int
	HintBegin int
	HintEnd   int
}

type liveMonotoneBudget struct {
	options LiveMonotoneOptions
	work    int
}

var liveCompiledSourceInventory = [...]string{
	"liveness.c",
	"liveness_sim.c",
	"ltl_parser.c",
	"kliveness.c",
	"monotone.c",
	"disjunctiveMonotone.c",
	"arenaViolation.c",
	"kLiveConstraints.c",
	"combination.c",
}

func LiveCompiledSourceInventory() []string {
	result := make([]string, len(liveCompiledSourceInventory))
	copy(result, liveCompiledSourceInventory[:])
	return result
}

func newLiveMonotoneBudget(options LiveMonotoneOptions) (*liveMonotoneBudget, error) {
	if options.MaxItems < 0 || options.MaxWork < 0 || options.MaxBytes < 0 {
		return nil, ErrDimension
	}
	if options.MaxItems == 0 {
		options.MaxItems = 1 << 16
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 8 << 20
	}
	return &liveMonotoneBudget{options: options}, nil
}

func (b *liveMonotoneBudget) take(n int) error {
	if n < 0 || b.work > int(^uint(0)>>1)-n {
		return ErrLimit
	}
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	b.work += n
	if b.work > b.options.MaxWork {
		return ErrLimit
	}
	return nil
}

func (b *liveMonotoneBudget) resultInts(n int) error {
	if n < 0 || n > b.options.MaxItems {
		return ErrLimit
	}
	intBytes := bits.UintSize / 8
	if n > b.options.MaxBytes/intBytes {
		return ErrLimit
	}
	return nil
}

// LiveFindMonotoneOutputs implements findPendingSignal and findHintOutputs.
// Native substring matching and source output order are preserved.
func LiveFindMonotoneOutputs(names []string, options LiveMonotoneOptions) (LiveMonotoneOutputs, error) {
	result := LiveMonotoneOutputs{Pending: -1, HintBegin: -1, HintEnd: -1}
	b, err := newLiveMonotoneBudget(options)
	if err != nil {
		return LiveMonotoneOutputs{}, err
	}
	if len(names) > b.options.MaxItems {
		return LiveMonotoneOutputs{}, ErrLimit
	}
	hintCount := 0
	for _, name := range names {
		if err = b.take(1); err != nil {
			return LiveMonotoneOutputs{}, err
		}
		if strings.Contains(name, "hint_") {
			hintCount++
		}
	}
	if err = b.resultInts(hintCount); err != nil {
		return LiveMonotoneOutputs{}, err
	}
	result.Hints = make([]int, 0, hintCount)
	for index, name := range names {
		if err = b.take(1); err != nil {
			return LiveMonotoneOutputs{}, err
		}
		if result.Pending < 0 && strings.Contains(name, "pendingSignal") {
			result.Pending = index
		}
		if strings.Contains(name, "hint_") {
			result.Hints = append(result.Hints, index)
		}
	}
	if len(result.Hints) != 0 {
		result.HintBegin = result.Hints[0]
		result.HintEnd = result.Hints[len(result.Hints)-1]
	}
	return result, nil
}

func liveValidateIndices(values []int, b *liveMonotoneBudget) error {
	if len(values) > b.options.MaxItems {
		return ErrLimit
	}
	for _, value := range values {
		if err := b.take(1); err != nil {
			return err
		}
		if value < 0 {
			return ErrDimension
		}
	}
	return nil
}

func liveContainsIndex(values []int, target int, b *liveMonotoneBudget) (bool, error) {
	for _, value := range values {
		if err := b.take(1); err != nil {
			return false, err
		}
		if value == target {
			return true, nil
		}
	}
	return false, nil
}

// LiveVectorDifference implements vectorDifference while returning owned
// storage. Source order and duplicate entries from the first vector survive.
func LiveVectorDifference(first, second []int, options LiveMonotoneOptions) ([]int, error) {
	b, err := newLiveMonotoneBudget(options)
	if err != nil {
		return nil, err
	}
	if err = liveValidateIndices(first, b); err != nil {
		return nil, err
	}
	if err = liveValidateIndices(second, b); err != nil {
		return nil, err
	}
	if err = b.resultInts(len(first)); err != nil {
		return nil, err
	}
	result := make([]int, 0, len(first))
	for _, value := range first {
		found, findErr := liveContainsIndex(second, value, b)
		if findErr != nil {
			return nil, findErr
		}
		if !found {
			result = append(result, value)
		}
	}
	return result, nil
}

// LiveRemainingMonotoneCandidates implements findRemainingMonotoneCandidates.
func LiveRemainingMonotoneCandidates(known, hints []int, options LiveMonotoneOptions) ([]int, error) {
	return LiveVectorDifference(hints, known, options)
}

// LiveConsequentCandidates implements updateAnteConseVectors.
func LiveConsequentCandidates(antecedents, candidates []int, options LiveMonotoneOptions) ([]int, error) {
	return LiveVectorDifference(candidates, antecedents, options)
}

// LivePushUnique implements Vec_IntPushUniqueLocal on owned Go storage.
func LivePushUnique(values []int, entry int, options LiveMonotoneOptions) ([]int, bool, error) {
	b, err := newLiveMonotoneBudget(options)
	if err != nil {
		return nil, false, err
	}
	if entry < 0 {
		return nil, false, ErrDimension
	}
	if err = liveValidateIndices(values, b); err != nil {
		return nil, false, err
	}
	found, err := liveContainsIndex(values, entry, b)
	if err != nil {
		return nil, false, err
	}
	size := len(values)
	if !found {
		if size == int(^uint(0)>>1) {
			return nil, false, ErrLimit
		}
		size++
	}
	if err = b.resultInts(size); err != nil {
		return nil, false, err
	}
	result := make([]int, len(values), size)
	copy(result, values)
	if !found {
		result = append(result, entry)
	}
	return result, found, nil
}
