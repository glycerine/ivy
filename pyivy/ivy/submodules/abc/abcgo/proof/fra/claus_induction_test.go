package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestClausInductionPass(t *testing.T) {
	for _, toggle := range []bool{false, true} {
		g := aig.NewGraph()
		state := g.NewInput()
		next := state
		if toggle {
			next = next.Not()
		}
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		data := cnf.Cnf_Derive(g, 1)
		lit := 2*data.VarNums[state.NodeID()] + 1 // invariant !state
		candidates := [][]int32{{lit}, {lit}}
		costs := []int32{1, 1}
		removed, err := clausInductivePass(data, 1, nil, candidates, costs, 0)
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if toggle {
			want = 2
		}
		if removed != want {
			t.Fatalf("toggle=%v removed=%d want=%d", toggle, removed, want)
		}
	}
}
