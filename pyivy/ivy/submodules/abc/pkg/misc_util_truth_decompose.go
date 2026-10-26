package abc

// Output decomposition from src/misc/util/utilTruth.h:
// Abc_Tt6CheckOutDec and Abc_TtCheckOutDec.

// TruthOutputOperator values match C's output-decomposition codes.
type TruthOutputOperator int

const (
	TruthOutputAnd    TruthOutputOperator = iota // x AND G
	TruthOutputNotAnd                            // NOT x AND G
	TruthOutputNotOr                             // NOT x OR G
	TruthOutputOr                                // x OR G
	TruthOutputXor                               // x XOR G
)

// TruthOutputDecomposition retains the original width. Remainder is independent
// of Variable. Combine that variable with Remainder according to Operator to
// reconstruct the source truth table.
type TruthOutputDecomposition struct {
	Variable  int
	Operator  TruthOutputOperator
	Remainder *TruthTable
}

// DecomposeVariable tests C's ordered cofactor cases: c0=false, c1=false,
// c0=true, c1=true, opposite cofactors. It returns found=false for unused
// variables or irreducible pairs instead of asserting on unused inputs.
// General cofactors extend C's eight-variable/four-word scratch restriction.
// No source storage changes; successful remainders own their storage.
func (t *TruthTable) DecomposeVariable(variable int) (TruthOutputDecomposition, bool, error) {
	fail := func(e error) (TruthOutputDecomposition, bool, error) { return TruthOutputDecomposition{}, false, e }
	zero, e := t.Cofactor(variable, false)
	if e != nil {
		return fail(e)
	}
	one, e := t.Cofactor(variable, true)
	if e != nil {
		return fail(e)
	}
	if zero.Equal(one) {
		return TruthOutputDecomposition{}, false, nil
	}
	if zero.IsConstant(false) {
		return TruthOutputDecomposition{variable, TruthOutputAnd, one}, true, nil
	}
	if one.IsConstant(false) {
		return TruthOutputDecomposition{variable, TruthOutputNotAnd, zero}, true, nil
	}
	if zero.IsConstant(true) {
		return TruthOutputDecomposition{variable, TruthOutputNotOr, one}, true, nil
	}
	if one.IsConstant(true) {
		return TruthOutputDecomposition{variable, TruthOutputOr, zero}, true, nil
	}
	opposite, e := zero.Opposite(one)
	if e != nil {
		return fail(e)
	}
	if opposite {
		return TruthOutputDecomposition{variable, TruthOutputXor, zero}, true, nil
	}
	return TruthOutputDecomposition{}, false, nil
}
