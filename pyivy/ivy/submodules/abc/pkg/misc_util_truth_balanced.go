package abc

// Balanced-output wrapper from opt/dau/dauCanon.c: Abc_TtCanonicizeWrap.

type TruthCanonicalAlgorithm uint8

const (
	TruthCanonicalHierarchical TruthCanonicalAlgorithm = iota
	TruthCanonicalAdaptive
	TruthCanonicalCostAware
)

// TruthBalancedOptions selects one concrete canonicalizer, avoiding interfaces.
// Only the selected algorithm's options are used. Limits apply to each polarity
// independently: the wrapper runs at most twice, so total work is bounded by
// twice those limits. The selected Stop callback applies throughout both runs.
type TruthBalancedOptions struct {
	Algorithm    TruthCanonicalAlgorithm
	Hierarchical TruthHierarchicalOptions
	Adaptive     TruthAdaptiveOptions
	CostAware    TruthCostAwareOptions
}

// TruthBalancedResult owns the selected representative. ComplementSelected
// identifies the winning source polarity; strict ties retain the original.
// CacheHit/HitLevel report the selected run's actual lookup status (C's
// hierarchical routine returns zero flags even on a hit). Full/hard flags
// describe the selected run, while work
// counters sum both runs. Source-specific NPN metadata is not retained by the
// cached canonicalizers. FullEnumeration is the native status bit (the
// hierarchical implementation does not set it even in exact mode).
type TruthBalancedResult struct {
	Table                                                     *TruthTable
	Evaluations                                               int
	ComplementSelected, CacheHit, FullEnumeration, Broadened  bool
	HitLevel                                                  int
	Refinements, ExactCandidates, Transitions, Configurations int
}

// CanonicalizeBalanced evaluates both output polarities precisely when half
// the packed minterms are set, using the same cache in original-then-complement
// order. Highest packed words compare first. Unbalanced functions run once.
// Input and representative storage are owned; invalid options, limits or
// cancellation return no partial result. Completed cache stages can remain
// after a failed second run, safely retryable by the underlying cache API.
func (t *TruthTable) CanonicalizeBalanced(cache *TruthHierarchyCache, options TruthBalancedOptions) (TruthBalancedResult, error) {
	fail := func(e error) (TruthBalancedResult, error) { return TruthBalancedResult{}, e }
	if options.Algorithm > TruthCanonicalCostAware {
		return fail(ErrDimension)
	}
	stopped := options.Hierarchical.Stop
	if options.Algorithm == TruthCanonicalAdaptive {
		stopped = options.Adaptive.Stop
	} else if options.Algorithm == TruthCanonicalCostAware {
		stopped = options.CostAware.Stop
	}
	if stopped != nil && stopped() {
		return fail(ErrCancelled)
	}
	populations, e := t.CofactorPopulations()
	if e != nil {
		return fail(e)
	}
	run := func(input *TruthTable) (TruthBalancedResult, error) {
		r := TruthBalancedResult{HitLevel: -1}
		switch options.Algorithm {
		case TruthCanonicalHierarchical:
			v, e := input.CanonicalizeHierarchical(cache, options.Hierarchical)
			if e != nil {
				return r, e
			}
			r.Table, r.CacheHit, r.HitLevel = v.Table, v.CacheHit, v.HitLevel
			r.Refinements, r.ExactCandidates, r.Transitions = v.Attempts, v.ExactCandidates, v.ExactTransitions
		case TruthCanonicalAdaptive:
			v, e := input.CanonicalizeAdaptiveCached(cache, options.Adaptive)
			if e != nil {
				return r, e
			}
			r.Table, r.CacheHit, r.HitLevel = v.Table, v.CacheHit, v.HitLevel
			r.FullEnumeration, r.Configurations = v.FullEnumeration, v.Configurations
		case TruthCanonicalCostAware:
			v, e := input.CanonicalizeCostAwareCached(cache, options.CostAware)
			if e != nil {
				return r, e
			}
			r.Table, r.CacheHit, r.HitLevel = v.Table, v.CacheHit, v.HitLevel
			r.FullEnumeration, r.Broadened, r.Configurations = v.FullEnumeration, v.Broadened, v.Configurations
		}
		return r, nil
	}
	first, e := run(t)
	if e != nil {
		return fail(e)
	}
	first.Evaluations = 1
	if populations.Ones != len(t.words)*32 {
		return first, nil
	}
	second, e := run(t.Not())
	if e != nil {
		return fail(e)
	}
	result := first
	if truthWordsLessReverse(second.Table.words, first.Table.words) {
		result = second
		result.ComplementSelected = true
	}
	result.Evaluations = 2
	firstCounts := [4]int{first.Refinements, first.ExactCandidates, first.Transitions, first.Configurations}
	secondCounts := [4]int{second.Refinements, second.ExactCandidates, second.Transitions, second.Configurations}
	var totals [4]int
	for i := range totals {
		totals[i], e = checkedSize(firstCounts[i], 1, secondCounts[i], int(^uint(0)>>1))
		if e != nil {
			return fail(e)
		}
	}
	result.Refinements, result.ExactCandidates, result.Transitions, result.Configurations = totals[0], totals[1], totals[2], totals[3]
	if stopped != nil && stopped() {
		return fail(ErrCancelled)
	}
	return result, nil
}
