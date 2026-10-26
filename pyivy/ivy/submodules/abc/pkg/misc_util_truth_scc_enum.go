package abc

// SCC-guided enumeration from opt/dau/dauCanon.c: Abc_TgSplitGroupsByScc,
// Abc_TgPermEnumerationScc and Abc_TgPhaseEnumerationScc.

func (s *truthAdaptiveState) splitGroupsBySCC() error {
	coefficients, e := s.shiftedGroupCoefficients(false)
	if e != nil {
		return e
	}
	groups := make([]TruthVariableGroup, 0, len(s.representatives))
	for _, group := range s.groups {
		start, end := group.Start, group.Start+group.Size
		for i := start + 1; i < end; i++ {
			value, label := coefficients[i], s.representatives[i]
			j := i
			for j > start && coefficients[j-1] > value {
				coefficients[j], s.representatives[j] = coefficients[j-1], s.representatives[j-1]
				j--
			}
			coefficients[j], s.representatives[j] = value, label
		}
		for begin := start; begin < end; {
			next := begin + 1
			for next < end && coefficients[next] == coefficients[begin] {
				next++
			}
			groups = append(groups, TruthVariableGroup{begin, next - begin})
			begin = next
		}
	}
	s.groups = groups
	s.applyRepresentativeOrder()
	return nil
}

type TruthSCCEnumerationOptions struct {
	PhaseMode         TruthSCCPhaseMode
	HighSymmetry      bool
	SplitByCofactors  bool // SCC group splitting used by C algorithm >1
	MaxConfigurations int  // zero: one million, shared by filtering and enumeration
	Stop              func() bool
}

// TruthSCCEnumeration retains forward metadata and counts from the group
// enumerator. Configurations includes PhaseConfigurations plus visited
// permutation candidates. Permutations counts candidates after phase filtering
// and optional SCC group splitting. FilteredPhases counts selected masks (one
// implicit phase when signatures already determine polarity).
type TruthSCCEnumeration struct {
	TruthAdaptiveEnumeration
	PhaseConfigurations, FilteredPhases int
}

// EnumerateAdaptiveSCC filters first-group phases by SCC, then enumerates the
// retained phases' signature-group permutations, optionally splitting groups
// by sorted cofactor coefficients first. Native mask/traversal/tie order is
// preserved. This is SCC-guided group search, not unrestricted NPN minimization
// or the adaptive threshold/cache wrapper. All scratch/results are owned;
// cancellation or shared-budget exhaustion returns no partial result.
func (t *TruthTable) EnumerateAdaptiveSCC(options TruthSCCEnumerationOptions) (TruthSCCEnumeration, error) {
	fail := func(e error) (TruthSCCEnumeration, error) { return TruthSCCEnumeration{}, e }
	filter, e := t.FilterAdaptivePhases(TruthSCCPhaseOptions{Mode: options.PhaseMode, HighSymmetry: options.HighSymmetry, MaxConfigurations: options.MaxConfigurations, Stop: options.Stop})
	if e != nil {
		return fail(e)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	reduction, e := t.ReduceInputSymmetry(options.HighSymmetry)
	if e != nil {
		return fail(e)
	}
	base := newTruthAdaptiveState(reduction)
	return enumerateTruthSCCState(base, filter, options)
}

func enumerateTruthSCCState(base *truthAdaptiveState, filter TruthSCCPhaseFilter, options TruthSCCEnumerationOptions) (TruthSCCEnumeration, error) {
	fail := func(e error) (TruthSCCEnumeration, error) { return TruthSCCEnumeration{}, e }
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	limit := options.MaxConfigurations
	if limit == 0 {
		limit = 1 << 20
	}
	result := TruthSCCEnumeration{PhaseConfigurations: filter.Visited, FilteredPhases: len(filter.Masks)}
	result.Configurations = filter.Visited
	masks := filter.Masks
	if base.phased {
		masks = []uint64{0}
		result.FilteredPhases = 1
	}
	var best *truthAdaptiveState
	for _, mask := range masks {
		work := base.clone()
		for i, head := range filter.PhaseGroups {
			if mask&(uint64(1)<<uint(i)) != 0 {
				work.flipChainBySource(head)
			}
		}
		if options.SplitByCofactors {
			if e := work.splitGroupsBySCC(); e != nil {
				return fail(e)
			}
		}
		directions := make([]int, len(work.representatives))
		for i := range directions {
			directions[i] = -1
		}
		for {
			if stopped() {
				return fail(ErrCancelled)
			}
			if result.Configurations >= limit {
				return fail(ErrLimit)
			}
			result.Configurations++
			result.Permutations++
			// C starts SCC search from all ones: even the first candidate may need
			// replacing the initial best, so no unfiltered truth seeds this minimum.
			if best == nil || truthWordsLessReverse(work.table.words, best.table.words) {
				best = work.clone()
				result.Changes++
			}
			if !work.nextPermutation(directions) {
				break
			}
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
