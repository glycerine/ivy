package cudd

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"testing"
)

func TestCofMintermSignatures(t *testing.T) {
	dd := NewBDDManager(4)
	f := core.Cudd_bddIte(dd, dd.Vars[0], core.Cudd_bddXor(dd, dd.Vars[1], dd.Vars[3]), Cudd_Not(dd.Vars[2]))
	for _, node := range []*DdNode{f, Cudd_Not(f), dd.Vars[2]} {
		got := Cudd_CofMinterm(dd, node)
		counts := [5]int{}
		for bits := 0; bits < 16; bits++ {
			eval := func(a int) bool {
				n := node
				for !core.Cudd_IsConstant(n) {
					r := core.Cudd_Regular(n)
					child := r.E
					if a&(1<<r.Index) != 0 {
						child = r.T
					}
					n = core.Cudd_NotCond(child, core.Cudd_IsComplement(n))
				}
				return n != dd.LogicZero
			}
			if eval(bits) {
				counts[4]++
			}
			for i := 0; i < 4; i++ {
				if eval(bits | 1<<i) {
					counts[i]++
				}
			}
		}
		for i, n := range counts {
			if got[i] != float64(n)/16 {
				t.Fatalf("node %p signature %d: got %v want %v", node, i, got[i], float64(n)/16)
			}
		}
	}
}
