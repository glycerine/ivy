package abc

// Cached hierarchical canonicalization from opt/dau/dauCanon.c:
// Abc_TtCanonicizeHie, including optional exact adjacent-group minimization.

type TruthHierarchicalOptions struct {
	Exact                                           bool // enable C fExact group minimization after heuristic refinement
	MaxScheduleSteps, MaxCandidates, MaxTransitions int  // exact-stage bounds; zero: one million each
	MaxRefinements                                  int  // zero: one million; counts adjacent refinement visits
	Stop                                            func() bool
}

// TruthHierarchicalResult owns the representative returned by the staged
// heuristic/cache. CacheHit reports an intermediate-stage representative hit;
// HitLevel is its zero-based level or -1. Transformation metadata is omitted:
// representative cache entries do not carry the source-specific NPN mapping.
// Refinement counters are zero for hits before the refinement stage.
type TruthHierarchicalResult struct {
	Table                             *TruthTable
	CacheHit                          bool
	HitLevel                          int
	Rounds, Attempts, Changes         int
	ExactSearch                       bool // exact stage ran; cache hits do not imply stored provenance
	ExactCandidates, ExactTransitions int
}

// CanonicalizeHierarchical ports C's hierarchical pipeline with cache
// lookups after output polarity, input phases, permutation and refinement,
// followed by final insertion. A nil/closed/zero-level cache disables lookups.
// Five levels retain all stages; smaller caches preserve C's earlier terminal
// stage behavior. Cache width must match truth. Pending stages survive failed
// traversals safely and can be retried; errors return no partial representative.
// Exact enables minimization within the native signature groups and phase
// policy; it does not claim the unrestricted NPN representative. Cache policy
// belongs to the caller: hits do not record which search mode produced a class.
func (t *TruthTable) CanonicalizeHierarchical(cache *TruthHierarchyCache, options TruthHierarchicalOptions) (TruthHierarchicalResult, error) {
	fail := func(e error) (TruthHierarchicalResult, error) { return TruthHierarchicalResult{}, e }
	if options.MaxRefinements < 0 || options.MaxScheduleSteps < 0 || options.MaxCandidates < 0 || options.MaxTransitions < 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	populations, e := t.CofactorPopulations()
	if e != nil {
		return fail(e)
	}
	if cache != nil && !cache.closed {
		if cache.variables != t.variables {
			return fail(ErrDimension)
		}
		if e = cache.Begin(); e != nil {
			return fail(e)
		}
	}
	result := TruthHierarchicalResult{HitLevel: -1}
	table := t.Clone()
	if t.variables == 0 {
		clear(table.words)
		result.Table = table
		return result, nil
	}
	ones, zero := populations.Ones, populations.Zero
	output := ones > len(t.words)*32
	if output {
		for i, w := range table.words {
			table.words[i] = ^w
		}
		ones = len(t.words)*64 - ones
		for i, c := range zero {
			zero[i] = len(t.words)*32 - c
		}
	}
	lookup := func(level int) (bool, error) {
		if stopped() {
			return false, ErrCancelled
		}
		found, e := cache.RetrieveOrInsert(level, table)
		if e != nil {
			return false, e
		}
		if found.Status == TruthHierarchyHit {
			result.Table = found.Representative
			result.CacheHit = true
			result.HitLevel = level
			return true, nil
		}
		return false, nil
	}
	if hit, e := lookup(0); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	phases := make([]bool, t.variables)
	labels := make([]int, t.variables)
	for i, c := range zero {
		labels[i] = i
		if c < ones-c {
			table.flipVariable(i)
			phases[i] = true
			zero[i] = ones - c
		}
	}
	if hit, e := lookup(1); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	// C's hierarchical pass uses the same nonstable selection sort as SemiCanonicalize.
	for i := 0; i < t.variables-1; i++ {
		best := i + 1
		for j := i + 2; j < t.variables; j++ {
			if zero[best] > zero[j] {
				best = j
			}
		}
		if zero[i] <= zero[best] {
			continue
		}
		zero[i], zero[best] = zero[best], zero[i]
		labels[i], labels[best] = labels[best], labels[i]
		table.swapVariables(i, best)
	}
	if hit, e := lookup(2); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	order := make([]int, t.variables)
	for target, source := range labels {
		order[source] = target
	}
	semi := TruthSemiCanonicalization{Table: table, Permutation: order, InputPhases: phases, OutputPhase: output, PackedOnes: ones, ZeroCofactorCounts: zero}
	refined, e := t.refineCanonicalSignatures(semi, TruthCanonicalOptions{MaxRefinements: options.MaxRefinements, Stop: options.Stop})
	if e != nil {
		return fail(e)
	}
	table = refined.Table
	result.Rounds, result.Attempts, result.Changes = refined.Rounds, refined.Attempts, refined.Changes
	if hit, e := lookup(3); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	if options.Exact {
		groups := []int{}
		for begin := 0; begin < t.variables; {
			end := begin + 1
			for end < t.variables && zero[end] == zero[begin] {
				end++
			}
			groups = append(groups, end-begin)
			begin = end
		}
		flipOutput := ones == len(t.words)*32
		// For one input C pStore[1] is the stored total, not a second signature.
		second := ones
		if t.variables > 1 {
			second = zero[1]
		}
		flipSingletons := flipOutput || zero[0] != second && zero[0] == ones-zero[0]
		exact, e := table.MinimizeGroups(groups, flipOutput, flipSingletons, TruthExactOptions{MaxScheduleSteps: options.MaxScheduleSteps, MaxCandidates: options.MaxCandidates, MaxTransitions: options.MaxTransitions, Stop: options.Stop})
		if e != nil {
			return fail(e)
		}
		table = exact.Table
		result.ExactSearch = true
		result.ExactCandidates = exact.Candidates
		result.ExactTransitions = exact.Transitions
	}
	if _, e := lookup(4); e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	result.Table = table
	return result, nil
}
