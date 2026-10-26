package abc

// Mux-input DSD from src/misc/util/utilTruth.h:Abc_TtCheckDsdMux.

// TruthMuxInputDecomposition describes source = G(control ? high : low^phase).
// Remainder represents G with the mux value placed in Control's variable slot;
// WhenFalse/WhenTrue are source variable positions, both unused by Remainder.
// Numeric variable positions avoid confusing position zero with literal false.
type TruthMuxInputDecomposition struct {
	Control, WhenFalse, WhenTrue int
	ComplementFalse              bool
	Remainder                    *TruthTable
}

// DecomposeMuxInput requires exactly one input exclusive to each control
// cofactor, as in C. Matching data cofactors produce a positive low input;
// swapped data cofactors produce a complemented low input. The latter corrects
// C's complemented-cofactor test, which can return an invalid decomposition
// and miss valid phased muxes. Owned general cofactors extend C's one-word
// kernel. No match exposes no partial result; source storage stays unchanged.
func (t *TruthTable) DecomposeMuxInput(control int) (TruthMuxInputDecomposition, bool, error) {
	fail := func(e error) (TruthMuxInputDecomposition, bool, error) { return TruthMuxInputDecomposition{}, false, e }
	low, e := t.Cofactor(control, false)
	if e != nil {
		return fail(e)
	}
	high, e := t.Cofactor(control, true)
	if e != nil {
		return fail(e)
	}
	lowVariable, highVariable := -1, -1
	for variable := 0; variable < t.variables; variable++ {
		if variable == control {
			continue
		}
		inLow, e := low.HasVariable(variable)
		if e != nil {
			return fail(e)
		}
		inHigh, e := high.HasVariable(variable)
		if e != nil {
			return fail(e)
		}
		if inLow && !inHigh {
			if lowVariable >= 0 {
				return TruthMuxInputDecomposition{}, false, nil
			}
			lowVariable = variable
		}
		if inHigh && !inLow {
			if highVariable >= 0 {
				return TruthMuxInputDecomposition{}, false, nil
			}
			highVariable = variable
		}
	}
	if lowVariable < 0 || highVariable < 0 {
		return TruthMuxInputDecomposition{}, false, nil
	}
	c00, e := low.Cofactor(lowVariable, false)
	if e != nil {
		return fail(e)
	}
	c01, e := low.Cofactor(lowVariable, true)
	if e != nil {
		return fail(e)
	}
	c10, e := high.Cofactor(highVariable, false)
	if e != nil {
		return fail(e)
	}
	c11, e := high.Cofactor(highVariable, true)
	if e != nil {
		return fail(e)
	}
	positive := c00.Equal(c10) && c01.Equal(c11)
	negative := c00.Equal(c11) && c01.Equal(c10)
	if !positive && !negative {
		return TruthMuxInputDecomposition{}, false, nil
	}
	complement := !positive
	variable, e := NewVariableTruthTable(t.variables, control, t.limit)
	if e != nil {
		return fail(e)
	}
	remainder, e := MuxTruthTables(variable, c11, c10)
	if e != nil {
		return fail(e)
	}
	return TruthMuxInputDecomposition{control, lowVariable, highVariable, complement, remainder}, true, nil
}
