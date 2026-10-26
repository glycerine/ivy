package abc

// Cached cost-aware normalization/search from opt/dau/dauCanon.c:
// Abc_TtCanonicizeCA and its negative hierarchy stages.

// TruthCachedCostAwareResult owns a representative without source NPN mapping.
// EstimatesKnown is false on hits before phase filtering. FullEnumeration and
// Broadened report the current call's native flags, not cached provenance.
// HitLevel is a physical zero-based cache stage or -1. Configurations counts
// the simple prepass, phase filtering and permutation search actually visited.
type TruthCachedCostAwareResult struct {
	Table                                      *TruthTable
	CacheHit                                   bool
	HitLevel                                   int
	EstimatesKnown, FullEnumeration, Broadened bool
	EnumerationEstimate, PhaseEstimate         float64
	Configurations                             int
}

// CanonicalizeCostAwareCached visits C's raw, normalized-phase, reduced,
// simple-refined and final cache stages. Intermediate hits return immediately;
// misses restore the reduced state after the simple prepass and run native
// cost-aware search. Unlike adaptive dispatch, a one-level cache stores final
// representatives. Nil/closed caches use uncached search. The cache width must
// match, and it is single-traversal. Limits/cancellation return no partial
// representative; abandoned stages can retry on the next invocation.
func (t *TruthTable) CanonicalizeCostAwareCached(cache *TruthHierarchyCache, options TruthCostAwareOptions) (TruthCachedCostAwareResult, error) {
	fail := func(e error) (TruthCachedCostAwareResult, error) { return TruthCachedCostAwareResult{}, e }
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
	if cache != nil && !cache.closed && cache.variables != t.variables {
		return fail(ErrDimension)
	}
	if cache == nil || cache.closed {
		plain, e := t.CanonicalizeCostAware(options)
		if e != nil {
			return fail(e)
		}
		return TruthCachedCostAwareResult{Table: plain.Table, HitLevel: -1, EstimatesKnown: t.variables > 0, FullEnumeration: t.variables > 0, Broadened: plain.Broadened, EnumerationEstimate: plain.EnumerationEstimate, PhaseEstimate: plain.PhaseEstimate, Configurations: plain.Configurations}, nil
	}
	if e := cache.Begin(); e != nil {
		return fail(e)
	}
	result := TruthCachedCostAwareResult{HitLevel: -1}
	if t.variables == 0 {
		result.Table = t.Clone()
		clear(result.Table.words)
		return result, nil
	}
	table := t.Clone()
	lookup := func(stage int) (bool, error) {
		if stopped() {
			return false, ErrCancelled
		}
		r, e := cache.RetrieveOrInsert(stage, table)
		if e != nil {
			return false, e
		}
		if r.Status == TruthHierarchyHit {
			result.Table = r.Representative
			result.CacheHit = true
			result.HitLevel = stage + len(cache.levels)
			return true, nil
		}
		return false, nil
	}
	if hit, e := lookup(-5); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	initial, e := t.CreateTiedGroups()
	if e != nil {
		return fail(e)
	}
	table = initial.Table
	if hit, e := lookup(-4); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	reduction, e := t.ReduceInputSymmetry(true)
	if e != nil {
		return fail(e)
	}
	base := newTruthAdaptiveState(reduction)
	table = base.table
	if hit, e := lookup(-3); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	limit := options.MaxConfigurations
	if limit == 0 {
		limit = 1 << 20
	}
	simple, e := t.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{HighSymmetry: true, MaxRefinements: limit, Stop: options.Stop})
	if e != nil {
		return fail(e)
	}
	table = simple.Table
	result.Configurations = simple.Attempts
	if hit, e := lookup(-2); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	remaining := limit - result.Configurations
	if remaining <= 0 {
		return fail(ErrLimit)
	}
	full, e := costAwareSearch(base, options, remaining)
	if e != nil {
		return fail(e)
	}
	table = full.Table
	result.Configurations += full.Configurations
	result.EstimatesKnown = true
	result.EnumerationEstimate = full.EnumerationEstimate
	result.PhaseEstimate = full.PhaseEstimate
	result.FullEnumeration = true
	result.Broadened = full.Broadened
	if _, e := lookup(-1); e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	result.Table = table
	return result, nil
}
