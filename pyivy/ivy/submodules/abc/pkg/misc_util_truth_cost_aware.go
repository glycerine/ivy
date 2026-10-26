package abc

// Cost-aware group reset/search from opt/dau/dauCanon.c: Abc_TgResetGroup,
// Abc_SccPhaseCost, algorithm-2 Abc_SccEnumCost and uncached Abc_TtCanonicizeCA.

import "slices"

func (s *truthAdaptiveState) resetGroups() {
	slices.Sort(s.representatives)
	s.applyRepresentativeOrder()
	s.phased = false
	s.groups = nil
	if len(s.representatives) > 0 {
		s.groups = []TruthVariableGroup{{0, len(s.representatives)}}
	}
}

type TruthCostAwareOptions struct {
	BroadenSearch     bool // C fCA: reset signature groups when fitted phase cost is lower
	MaxConfigurations int  // zero: one million, covering prepass, filtering and search
	Stop              func() bool
}

// TruthCostAwareResult supplies forward metadata and search counts. Broadened
// corresponds to C's hard-search flag and is separate from phase metadata.
// EnumerationEstimate and PhaseEstimate are the initial fitted costs used in
// the native broadening decision; they are not measured execution times.
type TruthCostAwareResult struct {
	TruthAdaptiveEnumeration
	Broadened                          bool
	EnumerationEstimate, PhaseEstimate float64
}

func truthCostAwareEstimates(variables, representatives, phaseBits, permutationCost int) (float64, float64) {
	enumeration := float64(variables)*0.94 + float64(phaseBits)*0.885 + float64(permutationCost)*0.00855 - 20.59
	phase := float64(variables)*0.997 + float64(representatives)*1.043 - 15.9
	return enumeration, phase
}

// CanonicalizeCostAware ports uncached C cost-aware search: high symmetry,
// simple prepass, signature-based phase filtering and split-SCC enumeration.
// When enabled and the native cost comparison favors it, sorted representative
// labels are reset to one group, and cost-first phase filtering replaces the
// initial signature filter. Linked symmetry/phase weights survive the reset.
// Caches are not used. Limits/cancellation return no partial result. Constants
// retain valid forward phases, extending C's degenerate empty search cases.
func (t *TruthTable) CanonicalizeCostAware(options TruthCostAwareOptions) (TruthCostAwareResult, error) {
	fail := func(e error) (TruthCostAwareResult, error) { return TruthCostAwareResult{}, e }
	if options.MaxConfigurations < 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	if e := t.Check(); e != nil {
		return fail(e)
	}
	if t.variables == 0 {
		initial, e := t.CreateTiedGroups()
		if e != nil {
			return fail(e)
		}
		if stopped() {
			return fail(ErrCancelled)
		}
		return TruthCostAwareResult{TruthAdaptiveEnumeration: TruthAdaptiveEnumeration{Table: initial.Table, Permutation: []int{}, InputPhases: initial.InputPhases, OutputPhase: initial.OutputPhase}}, nil
	}
	limit := options.MaxConfigurations
	if limit == 0 {
		limit = 1 << 20
	}
	// C performs this prepass before restoring the initial state for filtering.
	simple, e := t.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{HighSymmetry: true, MaxRefinements: limit, Stop: options.Stop})
	if e != nil {
		return fail(e)
	}
	spent := simple.Attempts
	remaining := limit - spent
	if remaining <= 0 {
		return fail(ErrLimit)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	reduction, e := t.ReduceInputSymmetry(true)
	if e != nil {
		return fail(e)
	}
	base := newTruthAdaptiveState(reduction)
	result, e := costAwareSearch(base, options, remaining)
	if e != nil {
		return fail(e)
	}
	result.Configurations += spent
	return result, nil
}

func costAwareSearch(base *truthAdaptiveState, options TruthCostAwareOptions, limit int) (TruthCostAwareResult, error) {
	fail := func(e error) (TruthCostAwareResult, error) { return TruthCostAwareResult{}, e }
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	spent, remaining := 0, limit
	filter, e := base.filterSCCPhases(TruthSCCPhaseOptions{Mode: TruthSCCPhaseSignature, MaxConfigurations: remaining, Stop: options.Stop})
	if e != nil {
		return fail(e)
	}
	estimate, phaseEstimate := truthCostAwareEstimates(base.table.variables, len(base.representatives), filter.CandidateLog2, filter.PermutationCost)
	result := TruthCostAwareResult{EnumerationEstimate: estimate, PhaseEstimate: phaseEstimate}
	if options.BroadenSearch && estimate > phaseEstimate {
		result.Broadened = true
		spent += filter.Visited
		remaining = limit - spent
		if remaining <= 0 {
			return fail(ErrLimit)
		}
		base.resetGroups()
		filter, e = base.filterSCCPhases(TruthSCCPhaseOptions{Mode: TruthSCCPhaseCost, MaxConfigurations: remaining, Stop: options.Stop})
		if e != nil {
			return fail(e)
		}
	}
	full, e := enumerateTruthSCCState(base, filter, TruthSCCEnumerationOptions{SplitByCofactors: true, MaxConfigurations: remaining, Stop: options.Stop})
	if e != nil {
		return fail(e)
	}
	result.TruthAdaptiveEnumeration = full.TruthAdaptiveEnumeration
	result.Configurations += spent
	if stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}
