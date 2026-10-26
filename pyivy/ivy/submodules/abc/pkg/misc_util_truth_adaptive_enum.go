package abc

// Algorithm-0 full adaptive group enumeration from opt/dau/dauCanon.c:
// Abc_NextPermSwapC, TgFirstPermutation/NextPermutation, TgPhaseEnumeration,
// TgSaveBest and TgFullEnumeration (non-SCC branch).

import (
	"math/bits"
	"slices"
)

// truthNextPermutationSwap returns C's next adjacent swap, updating directions
// before the caller swaps both labels and their directions. A negative result
// requests the final first-pair swap that restores the group's initial order.
func truthNextPermutationSwap(labels, directions []int) int {
	k := -1
	for i, value := range labels {
		j := i + directions[i]
		if j >= 0 && j < len(labels) && value > labels[j] && (k < 0 || value > labels[k]) {
			k = i
		}
	}
	if k < 0 {
		k = 0
	}
	for i, value := range labels {
		if value > labels[k] {
			directions[i] = -directions[i]
		}
	}
	j := k + directions[k]
	return min(j, k)
}
func (s *truthAdaptiveState) nextPermutation(directions []int) bool {
	for _, group := range s.groups {
		if group.Size <= 1 {
			continue
		}
		start, end := group.Start, group.Start+group.Size
		index := truthNextPermutationSwap(s.representatives[start:end], directions[start:end])
		swap := start + max(0, index)
		s.swapGroups(swap)
		directions[swap], directions[swap+1] = directions[swap+1], directions[swap]
		if index >= 0 {
			return true
		}
	}
	return false
}
func (s *truthAdaptiveState) flipChainBySource(head int) {
	for v := head; v >= 0; v = s.links[v] {
		if s.weights[v] != 0 {
			s.flipVariable(v)
		}
	}
}

type TruthAdaptiveEnumerationOptions struct {
	HighSymmetry      bool
	MaxConfigurations int // zero: one million; counts every visited phase/configuration
	Stop              func() bool
}

// TruthAdaptiveEnumeration records the minimum encountered by C's algorithm-0
// full search within normalized signature/symmetry groups. Permutation maps
// source inputs to target positions; phases are source-indexed. Configurations
// includes repeated truth values and Permutations includes each group ordering.
// This does not enumerate both output polarities of a balanced function, nor
// claim the same representative as unrestricted brute-force NPN minimization.
type TruthAdaptiveEnumeration struct {
	Table                                 *TruthTable
	Permutation                           []int
	InputPhases                           []bool
	OutputPhase                           bool
	Configurations, Permutations, Changes int
}

// EnumerateAdaptiveGroups normalizes signatures, reduces symmetry and visits
// group permutations in C's Johnson-Trotter order. For each ordering it walks
// first-group phases in weight-sorted Gray order, retaining the first strict
// reverse-word minimum. The phase state carries between permutations as in C.
// This is the full non-SCC branch, not the SCC/adaptive threshold/cache wrapper.
// Owned dynamic state replaces static scratch. Cancellation or exhaustion of
// MaxConfigurations returns an error and no partial result.
func (t *TruthTable) EnumerateAdaptiveGroups(options TruthAdaptiveEnumerationOptions) (TruthAdaptiveEnumeration, error) {
	fail := func(e error) (TruthAdaptiveEnumeration, error) { return TruthAdaptiveEnumeration{}, e }
	if options.MaxConfigurations < 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	reduction, e := t.ReduceInputSymmetry(options.HighSymmetry)
	if e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	work := newTruthAdaptiveState(reduction)
	best := work.clone()
	directions := make([]int, len(work.representatives))
	for i := range directions {
		directions[i] = -1
	}
	limit := options.MaxConfigurations
	if limit == 0 {
		limit = 1 << 20
	}
	result := TruthAdaptiveEnumeration{}
	visit := func() error {
		if stopped() {
			return ErrCancelled
		}
		if result.Configurations >= limit {
			return ErrLimit
		}
		result.Configurations++
		if truthWordsLessReverse(work.table.words, best.table.words) {
			best = work.clone()
			result.Changes++
		}
		return nil
	}
	for {
		result.Permutations++
		if e := visit(); e != nil {
			return fail(e)
		}
		if !work.phased && len(work.groups) > 0 {
			group := work.groups[0]
			phases := slices.Clone(work.representatives[group.Start : group.Start+group.Size])
			// Stable insertion by accumulated symmetry weight, preserving native ties.
			for i := 1; i < len(phases); i++ {
				v := phases[i]
				j := i
				for j > 0 && work.weights[phases[j-1]] > work.weights[v] {
					phases[j] = phases[j-1]
					j--
				}
				phases[j] = v
			}
			for step := uint64(1); step < uint64(1)<<uint(len(phases)); step++ {
				work.flipChainBySource(phases[bits.TrailingZeros64(step)])
				if e := visit(); e != nil {
					return fail(e)
				}
			}
		}
		if !work.nextPermutation(directions) {
			break
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	result.Table = best.table
	result.Permutation = best.position
	result.InputPhases = best.phases
	result.OutputPhase = best.output
	return result, nil
}
