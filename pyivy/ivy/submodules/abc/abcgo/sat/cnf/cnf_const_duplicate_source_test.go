package cnf

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Source Cnf_Derive oracle: constant and complemented PI outputs, duplicate
// outputs, and dangling/redundant nodes with every output kept free.
func TestSourceFreeConstantAndDuplicateOutputs(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	xy := g.AND(x, y)
	g.AND(y, z) // dangling
	out := g.AND(xy, z.Not())
	g.AND(x, x) // structurally redundant
	for _, e := range []aig.Edge{g.Const1(), x, x.Not(), out, out} {
		g.AddOutput(e)
	}
	f := Cnf_Derive(g, 5)
	if f == nil {
		t.Fatal("CNF derivation failed")
	}
	if f.Variables != 11 {
		t.Fatalf("variables = %d, want 11", f.Variables)
	}
	for i, input := range g.Inputs() {
		if got := f.VarNums[input.NodeID()]; got != int32(i+7) {
			t.Fatalf("input %d variable = %d", i, got)
		}
	}
	for i, output := range g.Outputs() {
		if got := f.VarNums[output.NodeID()]; got != int32(i+1) {
			t.Fatalf("output %d variable = %d", i, got)
		}
	}
	want := [][]bsat.Literal{{12, 15, 17, 18}, {13, 14}, {13, 16}, {13, 19}, {20}, {2, 21}, {3, 20}, {4, 15}, {5, 14}, {6, 14}, {7, 15}, {8, 13}, {9, 12}, {10, 13}, {11, 12}}
	if !reflect.DeepEqual(f.Clauses, want) {
		t.Fatalf("clauses = %v, want %v", f.Clauses, want)
	}
}
