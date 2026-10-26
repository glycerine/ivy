package abc

// Uncached adaptive dispatch from opt/dau/dauCanon.c:
// Abc_SccEnumCost, Abc_TgEnumerationCost and Abc_TtCanonicizeAda.

type TruthAdaptiveOptions struct {
	Threshold int // C encoding: >=1000 selects SCC; %1000>=100 selects high symmetry;
	// %100 is the enumeration threshold; values above 84 force full search.
	MaxConfigurations int // zero: one million, shared across estimation and search
	Stop              func() bool
}

// TruthAdaptiveResult retains forward transformation metadata. FullEnumeration
// reports the C exact-search status separately from input/output phase flags;
// it means full search in the selected normalized orbit. Algorithm is 0 for
// ordinary group enumeration or 1 for total-SCC-filtered enumeration.
// Configurations counts estimation signatures plus search visits; Permutations
// counts full-search orderings (zero on the simple path).
type TruthAdaptiveResult struct {
	TruthAdaptiveCanonicalization
	EnumerationCost                         int
	FullEnumeration                         bool
	Algorithm, Configurations, Permutations int
}

func truthAdaptiveEnumerationCost(variables, algorithm, phaseBits, permutationCost int) int {
	cost := float64(variables)
	if algorithm == 0 {
		cost += float64(phaseBits)*1.09 + float64(permutationCost)*0.01144
	} else {
		cost += float64(phaseBits)*0.855 + float64(permutationCost)*0.00797
	}
	return int(cost + 0.5)
}

// CanonicalizeAdaptive ports C's uncached threshold dispatcher. It chooses
// native-compatible simple, full-group or total-SCC-filtered search using the
// fitted enumeration cost and encoded threshold. The cache/hierarchical branch
// remains separate. Work limits and cancellation return no partial result.
// Empty-input constants retain a valid output-phase transformation, extending
// C's special case that clears truth without returning the constant's phase.
func (t *TruthTable) CanonicalizeAdaptive(options TruthAdaptiveOptions) (TruthAdaptiveResult, error) {
	fail := func(e error) (TruthAdaptiveResult, error) { return TruthAdaptiveResult{}, e }
	if options.Threshold < 0 || options.MaxConfigurations < 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	high := options.Threshold%1000 >= 100
	reduction, e := t.ReduceInputSymmetry(high)
	if e != nil {
		return fail(e)
	}
	result := TruthAdaptiveResult{}
	if options.Threshold >= 1000 {
		result.Algorithm = 1
	}
	limit := options.MaxConfigurations
	if limit == 0 {
		limit = 1 << 20
	}
	estimatedVisits := 0
	if t.variables == 0 {
		result.Table = reduction.Table
		result.Permutation = []int{}
		result.InputPhases = reduction.InputPhases
		result.OutputPhase = reduction.OutputPhase
		if stopped() {
			return fail(ErrCancelled)
		}
		return result, nil
	}
	result.EnumerationCost, estimatedVisits, e = t.adaptiveCost(reduction, options, limit)
	if e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	remaining := limit - estimatedVisits
	if remaining <= 0 {
		return fail(ErrLimit)
	}
	threshold := options.Threshold % 100
	result.FullEnumeration = threshold > 84 || result.EnumerationCost < threshold
	if !result.FullEnumeration {
		simple, e := t.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{HighSymmetry: high, MaxRefinements: remaining, Stop: options.Stop})
		if e != nil {
			return fail(e)
		}
		result.TruthAdaptiveCanonicalization = simple
		result.Configurations = estimatedVisits + simple.Attempts
	} else {
		var full TruthAdaptiveEnumeration
		if result.Algorithm == 0 {
			full, e = t.EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{HighSymmetry: high, MaxConfigurations: remaining, Stop: options.Stop})
		} else {
			var scc TruthSCCEnumeration
			scc, e = t.EnumerateAdaptiveSCC(TruthSCCEnumerationOptions{PhaseMode: TruthSCCPhaseTotal, HighSymmetry: high, MaxConfigurations: remaining, Stop: options.Stop})
			full = scc.TruthAdaptiveEnumeration
		}
		if e != nil {
			return fail(e)
		}
		result.TruthAdaptiveCanonicalization = TruthAdaptiveCanonicalization{Table: full.Table, Permutation: full.Permutation, InputPhases: full.InputPhases, OutputPhase: full.OutputPhase, Attempts: full.Configurations, Changes: full.Changes}
		result.Configurations = estimatedVisits + full.Configurations
		result.Permutations = full.Permutations
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}

func (t *TruthTable) adaptiveCost(reduction TruthSymmetryReduction, options TruthAdaptiveOptions, limit int) (int, int, error) {
	if len(reduction.Groups) == 0 {
		return 0, 0, nil
	}
	algorithm := 0
	if options.Threshold >= 1000 {
		algorithm = 1
	}
	phaseBits, visits := 0, 0
	if !reduction.PhasesDetermined {
		if algorithm == 0 {
			phaseBits = reduction.Groups[0].Size
		} else {
			filtered, e := t.FilterAdaptivePhases(TruthSCCPhaseOptions{Mode: TruthSCCPhaseTotal, HighSymmetry: options.Threshold%1000 >= 100, MaxConfigurations: limit, Stop: options.Stop})
			if e != nil {
				return 0, 0, e
			}
			phaseBits, visits = filtered.CandidateLog2, filtered.Visited
		}
	}
	permutationCost := truthSCCPermutationCost(reduction.Groups, make([]uint64, len(reduction.Representatives)))
	return truthAdaptiveEnumerationCost(t.variables, algorithm, phaseBits, permutationCost), visits, nil
}
