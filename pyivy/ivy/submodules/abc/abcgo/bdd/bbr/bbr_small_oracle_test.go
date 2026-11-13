package bbr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

func bbrTruth(g *aig.Graph, x, lo aig.Edge, table int) aig.Edge {
	result := g.Const0()
	for state := 0; state < 2; state++ {
		for input := 0; input < 2; input++ {
			index := 2*state + input
			if table&(1<<index) == 0 {
				continue
			}
			a, b := x, lo
			if input == 0 {
				a = a.Not()
			}
			if state == 0 {
				b = b.Not()
			}
			result = g.OR(result, g.AND(a, b))
		}
	}
	return result
}
func bbrFails(next, property int) bool {
	reachable := [2]bool{true, false}
	for {
		previous := reachable
		for state := 0; state < 2; state++ {
			if !reachable[state] {
				continue
			}
			for input := 0; input < 2; input++ {
				index := 2*state + input
				if property&(1<<index) != 0 {
					return true
				}
				dst := (next >> index) & 1
				reachable[dst] = true
			}
		}
		if reachable == previous {
			return false
		}
	}
}
func TestBddReachabilityOneRegisterOracle(t *testing.T) {
	for next := 0; next < 16; next++ {
		for property := 0; property < 16; property++ {
			g := aig.NewGraph()
			x, lo := g.NewInput(), g.NewInput()
			g.AddOutput(bbrTruth(g, x, lo, property))
			g.AddOutput(bbrTruth(g, x, lo, next))
			aig.Aig_ManSetRegNum(g, 1)
			var options Saig_ParBbr_t
			Bbr_ManSetDefaultParams(&options)
			options.FSilent = 1
			got := Aig_ManVerifyUsingBdds(g, &options)
			wantFail := bbrFails(next, property)
			if (got == 0) != wantFail || (!wantFail && got != 1) {
				t.Fatalf("next=%04b property=%04b: status %d", next, property, got)
			}
			if wantFail {
				cex := g.SequentialCounterexample()
				if cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
					t.Fatalf("next=%04b property=%04b: invalid cex", next, property)
				}
			}
		}
	}
}
