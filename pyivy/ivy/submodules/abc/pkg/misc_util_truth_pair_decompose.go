package abc

// Pair-input disjoint-support decomposition from utilTruth.h:Abc_TtCheckDsdAnd.

// TruthPairOperator values match C's executed cofactor conditions. For codes
// zero through three, bit zero negates First and bit one negates Second.
// C's inline comments for codes one/two interchange those input phases.
type TruthPairOperator int

const (
	TruthPairAnd TruthPairOperator = iota
	TruthPairNotFirstAnd
	TruthPairNotSecondAnd
	TruthPairNotBothAnd
	TruthPairXor
)

// TruthPairDecomposition replaces the pair's function with a bit at First.
// Remainder retains original width but is independent of Second. Evaluate
// Remainder after replacing First with Operator(First,Second) to reconstruct
// the source. A successful result may be degenerate, matching C's precedence.
type TruthPairDecomposition struct {
	First, Second int
	Operator      TruthPairOperator
	Remainder     *TruthTable
}

// DecomposeInputPair recognizes AND with input phases or XOR by comparing
// four cofactors in C's order. No match returns found=false. Unlike C's one-word
// kernel, general-width owned cofactors support larger tables. Source stays
// unchanged; invalid/repeated input positions reject before producing a result.
func (t *TruthTable) DecomposeInputPair(first, second int) (TruthPairDecomposition, bool, error) {
	fail := func(e error) (TruthPairDecomposition, bool, error) { return TruthPairDecomposition{}, false, e }
	if e := t.Check(); e != nil {
		return fail(e)
	}
	if first < 0 || second < 0 || first >= t.variables || second >= t.variables || first == second {
		return fail(ErrDimension)
	}
	zero, e := t.Cofactor(first, false)
	if e != nil {
		return fail(e)
	}
	one, e := t.Cofactor(first, true)
	if e != nil {
		return fail(e)
	}
	c00, e := zero.Cofactor(second, false)
	if e != nil {
		return fail(e)
	}
	c01, e := zero.Cofactor(second, true)
	if e != nil {
		return fail(e)
	}
	c10, e := one.Cofactor(second, false)
	if e != nil {
		return fail(e)
	}
	c11, e := one.Cofactor(second, true)
	if e != nil {
		return fail(e)
	}
	kind := TruthPairOperator(-1)
	var low, high *TruthTable
	switch {
	case c00.Equal(c01) && c00.Equal(c10):
		kind, low, high = TruthPairAnd, c00, c11
	case c11.Equal(c00) && c11.Equal(c10):
		kind, low, high = TruthPairNotFirstAnd, c11, c01
	case c11.Equal(c00) && c11.Equal(c01):
		kind, low, high = TruthPairNotSecondAnd, c11, c10
	case c11.Equal(c01) && c11.Equal(c10):
		kind, low, high = TruthPairNotBothAnd, c11, c00
	case c00.Equal(c11) && c01.Equal(c10):
		kind, low, high = TruthPairXor, c11, c10
	default:
		return TruthPairDecomposition{}, false, nil
	}
	control, e := NewVariableTruthTable(t.variables, first, t.limit)
	if e != nil {
		return fail(e)
	}
	remainder, e := MuxTruthTables(control, high, low)
	if e != nil {
		return fail(e)
	}
	return TruthPairDecomposition{first, second, kind, remainder}, true, nil
}
