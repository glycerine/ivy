package abc

// Exact-search permutation schedules from bool/lucky/luckySimple.c:
// setSwapInfoPtr/freeSwapInfoPtr, nextSwap, fillInSwapArray, oneBitPosition,
// fillInFlipArray and setPermInfoPtr/freePermInfoPtr.

import (
	"math/bits"
	"slices"
)

// TruthPermutationSchedule owns a Johnson-Trotter adjacent-swap walk and a
// reflected Gray input-phase walk. Accessors return copies in execution order;
// C stores both arrays backwards and consumes them from the end. The initial
// identity/zero-phase configuration is implicit, so lengths are n!-1 and 2^n-1.
// The zero value represents the valid zero-input schedule.
type TruthPermutationSchedule struct {
	variables    int
	swaps, flips []int
}

// NewTruthPermutationSchedule bounds the combined number of stored swaps and
// flips. Zero maxSteps means one million. Checked counts avoid C's factorial
// overflow and oversized allocations before building either schedule.
func NewTruthPermutationSchedule(variables, maxSteps int) (*TruthPermutationSchedule, error) {
	if variables < 0 || variables > 36 || maxSteps < 0 {
		return nil, ErrDimension
	}
	if maxSteps == 0 {
		maxSteps = 1 << 20
	}
	factorial := uint64(1)
	maxFactorial := uint64(maxSteps) + 1
	for i := 2; i <= variables; i++ {
		if factorial > maxFactorial/uint64(i) {
			return nil, ErrLimit
		}
		factorial *= uint64(i)
	}
	swaps := int(factorial - 1)
	flips := uint64(1)<<uint(variables) - 1
	if flips > uint64(maxSteps-swaps) {
		return nil, ErrLimit
	}
	// Check total integer slice payload as well as logical schedule length.
	if _, e := checkedSize(0, bits.UintSize/8, swaps+int(flips), int(^uint(0)>>1)); e != nil {
		return nil, e
	}
	schedule := &TruthPermutationSchedule{variables: variables, swaps: make([]int, swaps), flips: make([]int, int(flips))}
	labels, directions := make([]int, variables), make([]int, variables)
	for i := range labels {
		labels[i] = i
		directions[i] = -1
	}
	for i := range schedule.swaps {
		index := truthNextPermutationSwap(labels, directions)
		schedule.swaps[i] = index
		labels[index], labels[index+1] = labels[index+1], labels[index]
		directions[index], directions[index+1] = directions[index+1], directions[index]
	}
	for i := range schedule.flips {
		schedule.flips[i] = bits.TrailingZeros64(uint64(i) + 1)
	}
	return schedule, nil
}
func (s *TruthPermutationSchedule) Variables() int {
	if s == nil {
		return 0
	}
	return s.variables
}
func (s *TruthPermutationSchedule) AdjacentSwaps() []int {
	if s == nil {
		return nil
	}
	return slices.Clone(s.swaps)
}
func (s *TruthPermutationSchedule) GrayFlips() []int {
	if s == nil {
		return nil
	}
	return slices.Clone(s.flips)
}
