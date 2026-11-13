package cnf

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Cnf_Derive in ABC still assigns every PI and emits the constant-one unit
// even when the graph has no outputs or reachable logic.
func TestSourceOutputlessGraph(t *testing.T) {
	for _, nInputs := range []int{0, 4} {
		g := aig.NewGraph()
		inputs := make([]aig.Edge, nInputs)
		for i := range inputs {
			inputs[i] = g.NewInput()
		}
		f := Cnf_Derive(g, 0)
		if f == nil {
			t.Fatal("Cnf_Derive returned nil")
		}
		if f.Variables != nInputs+2 {
			t.Fatalf("%d inputs: variables=%d", nInputs, f.Variables)
		}
		for i, input := range inputs {
			if got := f.VarNums[input.NodeID()]; got != int32(i+1) {
				t.Fatalf("input %d variable=%d", i, got)
			}
		}
		want := [][]bsat.Literal{{bsat.Literal(2 * (nInputs + 1))}}
		if !reflect.DeepEqual(f.Clauses, want) {
			t.Fatalf("clauses=%v, want %v", f.Clauses, want)
		}
	}
}
