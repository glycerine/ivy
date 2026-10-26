package abc

// Adaptive permutation state and simple enumeration from opt/dau/dauCanon.c:
// Abc_TgImplementPerm, SwapAdjacentSymGroups, SymGroupPerm, PermPhase and
// SimpleEnumeration. The swap-only comparison direction follows C exactly.

import "slices"

type truthAdaptiveState struct {
	table                            *TruthTable
	order, position, representatives []int
	phases                           []bool
	links, weights                   []int
	groups                           []TruthVariableGroup
	phased, output                   bool
}

func newTruthAdaptiveState(r TruthSymmetryReduction) *truthAdaptiveState {
	n := r.Table.variables
	s := &truthAdaptiveState{table: r.Table, order: make([]int, n), position: make([]int, n), representatives: r.Representatives, phases: r.InputPhases, links: r.Links, weights: r.PhaseWeights, groups: r.Groups, phased: r.PhasesDetermined, output: r.OutputPhase}
	for i := range s.order {
		s.order[i] = i
		s.position[i] = i
	}
	s.applyOrder(r.ExpandedOrder)
	return s
}
func (s *truthAdaptiveState) clone() *truthAdaptiveState {
	r := *s
	r.table = s.table.Clone()
	r.order = slices.Clone(s.order)
	r.position = slices.Clone(s.position)
	r.representatives = slices.Clone(s.representatives)
	r.phases = slices.Clone(s.phases)
	// Links, weights and signature groups are immutable during enumeration.
	return &r
}
func (s *truthAdaptiveState) applyOrder(order []int) {
	for target, source := range order {
		old := s.position[source]
		other := s.order[target]
		s.table.swapVariables(target, old)
		s.order[target], s.order[old] = source, other
		s.position[source], s.position[other] = target, old
	}
}
func (s *truthAdaptiveState) swapGroups(index int) {
	s.representatives[index], s.representatives[index+1] = s.representatives[index+1], s.representatives[index]
	s.applyRepresentativeOrder()
}
func (s *truthAdaptiveState) applyRepresentativeOrder() {
	order := make([]int, 0, len(s.order))
	for _, head := range s.representatives {
		for v := head; v >= 0; v = s.links[v] {
			order = append(order, v)
		}
	}
	for v := s.links[len(s.order)]; v >= 0; v = s.links[v] {
		order = append(order, v)
	}
	s.applyOrder(order)
}
func (s *truthAdaptiveState) flipVariable(source int) {
	s.table.flipVariable(s.position[source])
	s.phases[source] = !s.phases[source]
}
func (s *truthAdaptiveState) flipGroup(index int) {
	for v := s.representatives[index]; v >= 0; v = s.links[v] {
		if s.weights[v] != 0 {
			s.flipVariable(v)
		}
	}
}
func (s *truthAdaptiveState) refineGroups(index int, swapOnly bool) (*truthAdaptiveState, uint8) {
	candidate := s.clone()
	if swapOnly {
		candidate.swapGroups(index)
		// C uses source < candidate here, unlike its minimizing phase branch.
		if truthWordsLessReverse(s.table.words, candidate.table.words) {
			return candidate, 4
		}
		return s, 0
	}
	best := s
	config := uint8(0)
	actions := [7]int{0, 1, 0, 2, 1, 0, 1}
	configs := [7]uint8{1, 3, 2, 6, 7, 5, 4}
	for i, action := range actions {
		if action == 2 {
			candidate.swapGroups(index)
		} else {
			candidate.flipGroup(index + action)
		}
		if truthWordsLessReverse(candidate.table.words, best.table.words) {
			best = candidate.clone()
			config = configs[i]
		}
	}
	return best, config
}

type TruthAdaptiveSimpleOptions struct {
	HighSymmetry   bool
	MaxRefinements int // zero: one million; counts pair and individual phase visits
	Stop           func() bool
}

// TruthAdaptiveCanonicalization supplies FORWARD metadata: source-indexed
// input phases, source-to-target permutation, and an output phase.
// Rounds is at most five; Attempts counts pair and phase visits; Changes counts
// accepted configurations. All result storage is owned.
type TruthAdaptiveCanonicalization struct {
	Table                     *TruthTable
	Permutation               []int
	InputPhases               []bool
	OutputPhase               bool
	Rounds, Attempts, Changes int
}

// CanonicalizeAdaptiveSimple ports the adaptive canonicalizer's simple search:
// stable signature groups, symmetry reduction, deferred permutation application,
// then five rounds of group-pair and individual-input phase passes. This is a
// heuristic, not an exact NPN representative. C's swap-only branch accepts a
// larger packed value; this native behavior is preserved, so the whole search
// is not monotone. Cancellation/budgets return no partial result.
func (t *TruthTable) CanonicalizeAdaptiveSimple(options TruthAdaptiveSimpleOptions) (TruthAdaptiveCanonicalization, error) {
	fail := func(e error) (TruthAdaptiveCanonicalization, error) { return TruthAdaptiveCanonicalization{}, e }
	if options.MaxRefinements < 0 {
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
	s := newTruthAdaptiveState(reduction)
	limit := options.MaxRefinements
	if limit == 0 {
		limit = 1 << 20
	}
	result := TruthAdaptiveCanonicalization{}
	visit := func() error {
		if stopped() {
			return ErrCancelled
		}
		if result.Attempts >= limit {
			return ErrLimit
		}
		result.Attempts++
		return nil
	}
	ids := make([]int, len(s.representatives))
	for i, g := range s.groups {
		for j := g.Start; j < g.Start+g.Size; j++ {
			ids[j] = i
		}
	}
	pair := func(index int) (bool, error) {
		if e := visit(); e != nil {
			return false, e
		}
		next, config := s.refineGroups(index, ids[index] > 0 || s.phased)
		s = next
		if config != 0 {
			result.Changes++
			return true, nil
		}
		return false, nil
	}
	phase := func(source int) (bool, error) {
		if e := visit(); e != nil {
			return false, e
		}
		candidate := s.table.Clone()
		candidate.flipVariable(s.position[source])
		if truthWordsLessReverse(candidate.words, s.table.words) {
			s.table = candidate
			s.phases[source] = !s.phases[source]
			result.Changes++
			return true, nil
		}
		return false, nil
	}
	for round := 0; round < 5; round++ {
		result.Rounds = round + 1
		changed := false
		for i := len(s.representatives) - 2; i >= 0; i-- {
			if ids[i] == ids[i+1] {
				moved, e := pair(i)
				if e != nil {
					return fail(e)
				}
				changed = changed || moved
			}
		}
		for i := 1; i < len(s.representatives)-1; i++ {
			if ids[i] == ids[i+1] {
				moved, e := pair(i)
				if e != nil {
					return fail(e)
				}
				changed = changed || moved
			}
		}
		for i := t.variables - 1; i >= 0; i-- {
			if s.weights[i] != 0 {
				moved, e := phase(i)
				if e != nil {
					return fail(e)
				}
				changed = changed || moved
			}
		}
		for i := 1; i < t.variables; i++ {
			if s.weights[i] != 0 {
				moved, e := phase(i)
				if e != nil {
					return fail(e)
				}
				changed = changed || moved
			}
		}
		if !changed {
			break
		}
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	result.Table = s.table
	result.Permutation = s.position
	result.InputPhases = s.phases
	result.OutputPhase = s.output
	return result, nil
}
