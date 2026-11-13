package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// ABC fraClau.c checks the property before comparing the iteration bound,
// then declares success unless Iter equals nIters, even for negative limits.
func TestClauIterationLimitSourceBehavior(t *testing.T) {
	for _, tc := range []struct{ limit, want int32 }{{-1, 1}, {0, 0}, {1, 0}} {
		g := aig.NewGraph()
		input, state := g.NewInput(), g.NewInput()
		g.AddOutput(state)
		g.AddOutput(input)
		aig.Aig_ManSetRegNum(g, 1)
		if got := Fra_Clau(g, tc.limit, 0, 0); got != tc.want {
			t.Errorf("limit %d: got %d, want %d", tc.limit, got, tc.want)
		}
	}
}

// On this circuit ABC's first inductive strengthening clause makes the main
// solver inconsistent. ABC increments Iter before comparing it to the bound:
// bound 1 therefore returns 0, whereas bound 2 returns 1. The C fixture in
// test_vectors/clau_final_iteration_oracle.c returns the same pair; instrumented
// C and Go runs confirm the unsatisfiable-strengthening path is reached.
func TestClauFinalIterationUnsatisfiableStrengthening(t *testing.T) {
	for _, tc := range []struct{ limit, want int32 }{{1, 0}, {2, 1}} {
		g := aig.NewGraph()
		e := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		e = append(e, g.AND(e[1].Not(), e[2].Not()))
		e = append(e, g.AND(e[3], e[1].Not()))
		e = append(e, g.AND(e[2].Not(), e[2].Not()))
		e = append(e, g.AND(e[6].Not(), e[0]))
		e = append(e, g.AND(e[1], e[1].Not()))
		e = append(e, g.AND(e[4].Not(), e[1]))
		e = append(e, g.AND(e[1].Not(), e[3]))
		e = append(e, g.AND(e[6].Not(), e[10]))
		g.AddOutput(e[7])
		g.AddOutput(e[2])
		g.AddOutput(e[1])
		aig.Aig_ManSetRegNum(g, 2)
		if got := Fra_Clau(g, tc.limit, 0, 0); got != tc.want {
			t.Errorf("limit %d: got %d, want %d", tc.limit, got, tc.want)
		}
	}
}
