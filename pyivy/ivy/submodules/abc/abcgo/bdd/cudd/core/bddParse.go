package core

import (
	"fmt"
	"io"

	"github.com/glycerine/abcgo/misc/parse"
)

// BDDOperations connects the backend-independent ABC formula parser to this
// manager's canonical BDD operations. Bracket ranks shift all supported indices
// simultaneously; negative target indices are rejected.
func BDDOperations(dd *DdManager) parse.Operations[*DdNode] {
	return parse.Operations[*DdNode]{
		Constant: func(value bool) *DdNode {
			if value {
				return dd.One
			}
			return dd.LogicZero
		},
		Variable: func(index int) *DdNode { return Cudd_bddIthVar(dd, index) },
		Not:      Cudd_Not,
		And:      func(a, b *DdNode) *DdNode { return Cudd_bddAnd(dd, a, b) },
		Or:       func(a, b *DdNode) *DdNode { return Cudd_bddOr(dd, a, b) },
		Xor:      func(a, b *DdNode) *DdNode { return Cudd_bddXor(dd, a, b) },
		Shift: func(f *DdNode, shift int) (*DdNode, error) {
			if !dd.valid(f) {
				return nil, fmt.Errorf("cudd: invalid formula edge")
			}
			permutation := make([]int, len(dd.Vars))
			for i := range permutation {
				permutation[i] = i
			}
			for _, i := range dd.Support(f) {
				target := i + shift
				if target < 0 {
					return nil, fmt.Errorf("cudd: rank shift moves variable %d below zero", i)
				}
				permutation[i] = target
			}
			result := Cudd_bddPermute(dd, f, permutation)
			if result == nil {
				return nil, fmt.Errorf("cudd: invalid rank permutation")
			}
			return result, nil
		},
	}
}

// ParseFormula retains the original parser's zero-rank variable inventory and
// diagnostic-stream convention while using standard-library-backed Go nodes.
func ParseFormula(output io.Writer, formula string, nVars, nRanks int, names []string, dd *DdManager, vars []*DdNode) *DdNode {
	if dd == nil || dd.closed || nVars < 0 || nRanks < 0 || nVars > 0 && nRanks > len(dd.Vars)/nVars-1 {
		if output != nil {
			fmt.Fprintln(output, "Parse_FormulaParser(): invalid BDD variable/rank inventory")
		}
		return nil
	}
	return parse.Parse_FormulaParser(output, formula, nVars, nRanks, names, BDDOperations(dd), vars)
}
