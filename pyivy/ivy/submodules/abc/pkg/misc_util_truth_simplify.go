package abc

// Literal-bound truth simplification from src/misc/util/utilTruth.h:Abc_TtSimplify.

// TruthInputSimplification describes a function over the surviving literal
// values. Bindings uses ABC/GIA literal encoding: 0/1 are constants, other IDs
// are symbolic input labels and the low bit complements that label. IDs do not
// refer to a GiaManager here. SourceVariables gives representative source slots.
// Moved matches MinimumBase's movement flag, not whether the function changed.
type TruthInputSimplification struct {
	Table           *TruthTable
	Bindings        []GiaLit
	SourceVariables []int
	Moved           bool
}

// SimplifyInputs substitutes constants and merges equal/complementary symbolic
// inputs, then compacts support. First occurrences are representatives; their
// original literal phase survives in Bindings. Result input i is the VALUE of
// Bindings[i], including its phase. Thus evaluating all source bindings under
// any shared symbolic assignment equals evaluating the result's bindings.
// Tables/argument slices remain unchanged. Unlike C's six-variable alias path,
// owned general cofactors handle every supported width without fixed scratch.
func (t *TruthTable) SimplifyInputs(bindings []GiaLit) (TruthInputSimplification, error) {
	fail := func(e error) (TruthInputSimplification, error) { return TruthInputSimplification{}, e }
	if e := t.Check(); e != nil {
		return fail(e)
	}
	if len(bindings) != t.variables {
		return fail(ErrDimension)
	}
	for _, l := range bindings {
		if !l.Valid() {
			return fail(ErrInvalidLiteral)
		}
	}
	result := t.Clone()
	for variable, l := range bindings {
		if l.ID() != 0 {
			continue
		}
		var e error
		result, e = result.Cofactor(variable, l.Complemented())
		if e != nil {
			return fail(e)
		}
	}
	representatives := make(map[GiaID]int)
	for variable, l := range bindings {
		if l.ID() == 0 {
			continue
		}
		representative, exists := representatives[l.ID()]
		if !exists {
			representatives[l.ID()] = variable
			continue
		}
		opposite := bindings[representative].Complemented() != l.Complemented()
		low, e := result.Cofactor(representative, false)
		if e != nil {
			return fail(e)
		}
		low, e = low.Cofactor(variable, opposite)
		if e != nil {
			return fail(e)
		}
		high, e := result.Cofactor(representative, true)
		if e != nil {
			return fail(e)
		}
		high, e = high.Cofactor(variable, !opposite)
		if e != nil {
			return fail(e)
		}
		control, e := NewVariableTruthTable(t.variables, representative, t.limit)
		if e != nil {
			return fail(e)
		}
		result, e = MuxTruthTables(control, high, low)
		if e != nil {
			return fail(e)
		}
	}
	compact, e := result.CompactSupport()
	if e != nil {
		return fail(e)
	}
	surviving := make([]GiaLit, len(compact.SourceVariables))
	for i, source := range compact.SourceVariables {
		surviving[i] = bindings[source]
		if surviving[i].ID() == 0 {
			return fail(ErrInvariant)
		}
	}
	return TruthInputSimplification{compact.Table, surviving, compact.SourceVariables, compact.Moved}, nil
}
