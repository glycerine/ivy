package abc

// Cofactor refinement and default heuristic NPN pipelines from
// opt/dau/dauCanon.c:Abc_Tt6CofactorPermNaive, CofactorPermNaive/Config/Perm,
// Abc_TtCanonicize and Abc_TtCanonicizePerm (active naive candidate path).

// TruthPairRefinement Configuration uses bit 0/1 for source-position input
// phases and bit 2 for swapping the adjacent pair AFTER those phases.
// Zero means unchanged. Naive refinement retains the first visited minimum.
type TruthPairRefinement struct {
	Table         *TruthTable
	Configuration uint8
}

func truthWordsLessReverse(a, b []uint64) bool {
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return false
}
func (t *TruthTable) refineAdjacent(first int, swapOnly bool) TruthPairRefinement {
	best := t.Clone()
	if swapOnly {
		best.swapVariables(first, first+1)
		if truthWordsLessReverse(best.words, t.words) {
			return TruthPairRefinement{best, 4}
		}
		copy(best.words, t.words)
		return TruthPairRefinement{best, 0}
	}
	candidate := t.Clone()
	config := uint8(0)
	// C visits input-phase/swap configurations in a Gray walk, not numeric order.
	actions := [7]int{0, 1, 0, 2, 1, 0, 1}
	configs := [7]uint8{1, 3, 2, 6, 7, 5, 4}
	for i, action := range actions {
		if action == 2 {
			candidate.swapVariables(first, first+1)
		} else {
			candidate.flipVariable(first + action)
		}
		if truthWordsLessReverse(candidate.words, best.words) {
			copy(best.words, candidate.words)
			config = configs[i]
		}
	}
	return TruthPairRefinement{best, config}
}

// RefineAdjacentCofactors minimizes the high-word-first packed value over the
// pair's phases and swap (or swap alone). Source storage stays unchanged.
func (t *TruthTable) RefineAdjacentCofactors(first int, swapOnly bool) (TruthPairRefinement, error) {
	if e := t.Check(); e != nil {
		return TruthPairRefinement{}, e
	}
	if first < 0 || first >= t.variables-1 {
		return TruthPairRefinement{}, ErrDimension
	}
	return t.refineAdjacent(first, swapOnly), nil
}

type TruthCanonicalOptions struct {
	OnlyPermutation bool
	MaxRefinements  int // zero: one million; counts visited tied adjacent pairs
	Stop            func() bool
}

// TruthCanonicalization retains SemiCanonicalize's forward metadata contract.
// Rounds counts C's outer iterations (at most five), Attempts the visited tied
// pairs, and Changes the nonidentity improving configurations.
type TruthCanonicalization struct {
	TruthSemiCanonicalization
	Rounds, Attempts, Changes int
}

// CanonicalizeNPN ports C's default heuristic: population semicanonicalization,
// then up to five descending/ascending adjacent-pair refinement rounds within
// equal-signature groups. Balanced groups may flip inputs; other groups only
// swap. OnlyPermutation disables all phase changes. This is C's heuristic,
// not an exhaustive/exact NPN canonical form. Budgets/cancellation return no
// partial result. Owned scratch replaces C's fixed static arrays and buffers.
func (t *TruthTable) CanonicalizeNPN(options TruthCanonicalOptions) (TruthCanonicalization, error) {
	fail := func(e error) (TruthCanonicalization, error) { return TruthCanonicalization{}, e }
	if options.MaxRefinements < 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	semi, e := t.SemiCanonicalize(options.OnlyPermutation)
	if e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return t.refineCanonicalSignatures(semi, options)
}

func (t *TruthTable) refineCanonicalSignatures(semi TruthSemiCanonicalization, options TruthCanonicalOptions) (TruthCanonicalization, error) {
	fail := func(e error) (TruthCanonicalization, error) { return TruthCanonicalization{}, e }
	if options.MaxRefinements < 0 {
		return fail(ErrDimension)
	}
	limit := options.MaxRefinements
	if limit == 0 {
		limit = 1 << 20
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	result := TruthCanonicalization{TruthSemiCanonicalization: semi}
	labels := make([]int, t.variables)
	for source, target := range semi.Permutation {
		labels[target] = source
	}
	refine := func(first int) (bool, error) {
		if stopped() {
			return false, ErrCancelled
		}
		if result.Attempts >= limit {
			return false, ErrLimit
		}
		result.Attempts++
		swapOnly := options.OnlyPermutation || semi.ZeroCofactorCounts[first] != semi.PackedOnes/2
		pair := result.Table.refineAdjacent(first, swapOnly)
		result.Table = pair.Table
		config := pair.Configuration
		if config == 0 {
			return false, nil
		}
		result.Changes++
		if config&1 != 0 {
			source := labels[first]
			result.InputPhases[source] = !result.InputPhases[source]
		}
		if config&2 != 0 {
			source := labels[first+1]
			result.InputPhases[source] = !result.InputPhases[source]
		}
		if config&4 != 0 {
			labels[first], labels[first+1] = labels[first+1], labels[first]
		}
		return true, nil
	}
	for round := 0; round < 5; round++ {
		result.Rounds = round + 1
		changed := false
		for first := t.variables - 2; first >= 0; first-- {
			if semi.ZeroCofactorCounts[first] != semi.ZeroCofactorCounts[first+1] {
				continue
			}
			moved, err := refine(first)
			if err != nil {
				return fail(err)
			}
			changed = changed || moved
		}
		if !changed {
			break
		}
		changed = false
		for first := 1; first < t.variables-1; first++ {
			if semi.ZeroCofactorCounts[first] != semi.ZeroCofactorCounts[first+1] {
				continue
			}
			moved, err := refine(first)
			if err != nil {
				return fail(err)
			}
			changed = changed || moved
		}
		if !changed {
			break
		}
	}
	for target, source := range labels {
		result.Permutation[source] = target
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}
