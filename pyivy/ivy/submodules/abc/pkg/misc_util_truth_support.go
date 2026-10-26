package abc

// Truth support compaction/stretching from src/misc/util/utilTruth.h:
// Abc_TtStretch6, Shrink, MinimumBase, MinBase and Expand.

import "slices"

// TruthSupportReduction maps compact variable positions to source variables.
// Moved matches MinimumBase's return: true if support was not already a low
// prefix. Width may shrink even when Moved is false (including constants).
type TruthSupportReduction struct {
	Table           *TruthTable
	SourceVariables []int
	Moved           bool
}

// Stretch adds unused high variables by repeating the existing words. It
// preserves ABC's stretched small-table representation and allocation limit.
func (t *TruthTable) Stretch(variables int) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if variables < t.variables {
		return nil, ErrDimension
	}
	result, e := NewConstantTruthTable(variables, false, t.limit)
	if e != nil {
		return nil, e
	}
	for i := range result.words {
		result.words[i] = t.words[i%len(t.words)]
	}
	return result, nil
}

// ProjectVariables retains inputs in caller-specified order and removes the
// rest. Every removed input must be functionally unused; otherwise it returns
// ErrUnsupported. Unlike unchecked C Shrink, no active input is discarded.
// Result variable i denotes original variable keep[i]; all storage is owned.
func (t *TruthTable) ProjectVariables(keep []int) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if len(keep) > t.variables {
		return nil, ErrDimension
	}
	included := make([]bool, t.variables)
	order := make([]int, t.variables)
	for target, source := range keep {
		if source < 0 || source >= t.variables || included[source] {
			return nil, ErrDimension
		}
		included[source] = true
		order[source] = target
	}
	next := len(keep)
	for variable, yes := range included {
		if yes {
			continue
		}
		used, e := t.HasVariable(variable)
		if e != nil {
			return nil, e
		}
		if used {
			return nil, ErrUnsupported
		}
		order[variable] = next
		next++
	}
	result, e := t.PermuteVariables(order)
	if e != nil {
		return nil, e
	}
	count, e := TruthTableWordCount(len(keep))
	if e != nil {
		return nil, e
	}
	result.variables = len(keep)
	result.words = slices.Clone(result.words[:count])
	if e = result.Check(); e != nil {
		return nil, e
	}
	return result, nil
}
func (t *TruthTable) CompactSupport() (TruthSupportReduction, error) {
	support, e := t.Support()
	if e != nil {
		return TruthSupportReduction{}, e
	}
	moved := false
	for i, variable := range support {
		if variable != i {
			moved = true
		}
	}
	result, e := t.ProjectVariables(support)
	if e != nil {
		return TruthSupportReduction{}, e
	}
	return TruthSupportReduction{result, support, moved}, nil
}

// ExpandVariables embeds source variables at distinct target positions. The
// sourceToTarget array has one entry per source variable, matching Expand's
// cut embedding without requiring sorted labels. Extra target variables are
// unused. Source, caller mapping and storage limits remain unchanged.
func (t *TruthTable) ExpandVariables(variables int, sourceToTarget []int) (*TruthTable, error) {
	if e := t.Check(); e != nil {
		return nil, e
	}
	if variables < t.variables || len(sourceToTarget) != t.variables {
		return nil, ErrDimension
	}
	if _, e := TruthTableWordCount(variables); e != nil {
		return nil, e
	}
	used := make([]bool, variables)
	order := make([]int, variables)
	for source, target := range sourceToTarget {
		if target < 0 || target >= variables || used[target] {
			return nil, ErrDimension
		}
		used[target] = true
		order[source] = target
	}
	next := t.variables
	for target, yes := range used {
		if !yes {
			order[next] = target
			next++
		}
	}
	stretched, e := t.Stretch(variables)
	if e != nil {
		return nil, e
	}
	return stretched.PermuteVariables(order)
}
