package abc

// Cached adaptive stages from opt/dau/dauCanon.c: Abc_TtCanonicizeAda,
// using its -5/-4/-3/-2/-1 hierarchy aliases and simple-before-full branch.

// TruthCachedAdaptiveResult owns a class representative, without a source NPN
// mapping that the cache does not retain. CostKnown is false on hits before
// cost estimation. FullEnumeration reflects the current call's native status
// once cost is known, not cached representative provenance. HitLevel is a
// physical zero-based cache level or -1. Configurations includes estimation,
// simple prepass and full-search visits actually performed by this call.
type TruthCachedAdaptiveResult struct {
	Table                           *TruthTable
	CacheHit                        bool
	HitLevel                        int
	CostKnown, FullEnumeration      bool
	EnumerationCost, Configurations int
}

// CanonicalizeAdaptiveCached ports adaptive canonicalization's cached stages.
// Raw truth, phase normalization, applied symmetry reduction, simple refinement
// and final representative use C's negative level aliases. Cache hits return
// early; deeper caches run a simple prepass even when full search is selected.
// A one-level cache follows C's uncached branch and receives no entries.
// Nil/closed caches use uncached dispatch. Widths must match live caches.
// Limits/cancellation yield no partial representative; interrupted stages retry
// safely on the next call. The cache is single-traversal, like its base API.
func (t *TruthTable) CanonicalizeAdaptiveCached(cache *TruthHierarchyCache, options TruthAdaptiveOptions) (TruthCachedAdaptiveResult, error) {
	fail := func(e error) (TruthCachedAdaptiveResult, error) { return TruthCachedAdaptiveResult{}, e }
	if options.Threshold < 0 || options.MaxConfigurations < 0 {
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
	if cache == nil || cache.closed || len(cache.levels) == 1 {
		plain, e := t.CanonicalizeAdaptive(options)
		if e != nil {
			return fail(e)
		}
		return TruthCachedAdaptiveResult{Table: plain.Table, HitLevel: -1, CostKnown: true, FullEnumeration: plain.FullEnumeration, EnumerationCost: plain.EnumerationCost, Configurations: plain.Configurations}, nil
	}
	if e := cache.Begin(); e != nil {
		return fail(e)
	}
	result := TruthCachedAdaptiveResult{HitLevel: -1}
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
	high := options.Threshold%1000 >= 100
	reduction, e := t.ReduceInputSymmetry(high)
	if e != nil {
		return fail(e)
	}
	table = newTruthAdaptiveState(reduction).table
	limit := options.MaxConfigurations
	if limit == 0 {
		limit = 1 << 20
	}
	result.EnumerationCost, result.Configurations, e = t.adaptiveCost(reduction, options, limit)
	if e != nil {
		return fail(e)
	}
	result.CostKnown = true
	threshold := options.Threshold % 100
	result.FullEnumeration = threshold > 84 || result.EnumerationCost < threshold
	if hit, e := lookup(-3); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	remaining := limit - result.Configurations
	if remaining <= 0 {
		return fail(ErrLimit)
	}
	simple, e := t.CanonicalizeAdaptiveSimple(TruthAdaptiveSimpleOptions{HighSymmetry: high, MaxRefinements: remaining, Stop: options.Stop})
	if e != nil {
		return fail(e)
	}
	table = simple.Table
	result.Configurations += simple.Attempts
	if hit, e := lookup(-2); e != nil {
		return fail(e)
	} else if hit {
		return result, nil
	}
	if result.FullEnumeration {
		remaining = limit - result.Configurations
		if remaining <= 0 {
			return fail(ErrLimit)
		}
		if options.Threshold < 1000 {
			full, e := t.EnumerateAdaptiveGroups(TruthAdaptiveEnumerationOptions{HighSymmetry: high, MaxConfigurations: remaining, Stop: options.Stop})
			if e != nil {
				return fail(e)
			}
			table = full.Table
			result.Configurations += full.Configurations
		} else {
			full, e := t.EnumerateAdaptiveSCC(TruthSCCEnumerationOptions{PhaseMode: TruthSCCPhaseTotal, HighSymmetry: high, MaxConfigurations: remaining, Stop: options.Stop})
			if e != nil {
				return fail(e)
			}
			table = full.Table
			result.Configurations += full.Configurations
		}
	}
	if _, e := lookup(-1); e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	result.Table = table
	return result, nil
}
