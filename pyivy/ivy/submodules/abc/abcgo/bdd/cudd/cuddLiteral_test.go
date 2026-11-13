package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"testing"
)

// Enumerate every pair of four-variable literal sets, including absent,
// positive, and negative occurrences, against a direct set intersection.
func TestLiteralSetIntersectionAllFourVariableCubes(t *testing.T) {
	dd := NewBDDManager(4)
	cube := func(code int) (*DdNode, [4]int) {
		result := dd.One
		var phases [4]int
		for i := range phases {
			phases[i] = code % 3
			code /= 3
			if phases[i] == 0 {
				continue
			}
			lit := dd.Vars[i]
			if phases[i] == 2 {
				lit = Cudd_Not(lit)
			}
			result = core.Cudd_bddAnd(dd, result, lit)
			if result == nil {
				t.Fatal("cube construction failed")
			}
		}
		return result, phases
	}
	for a := 0; a < 81; a++ {
		f, af := cube(a)
		for b := 0; b < 81; b++ {
			g, bf := cube(b)
			expected := dd.One
			for i := range af {
				if af[i] == 0 || af[i] != bf[i] {
					continue
				}
				lit := dd.Vars[i]
				if af[i] == 2 {
					lit = Cudd_Not(lit)
				}
				expected = core.Cudd_bddAnd(dd, expected, lit)
			}
			if got := Cudd_bddLiteralSetIntersection(dd, f, g); got != expected {
				t.Fatalf("cubes %d,%d: got %p want %p", a, b, got, expected)
			}
		}
	}
}
