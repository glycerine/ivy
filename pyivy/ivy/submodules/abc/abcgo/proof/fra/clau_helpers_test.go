package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
	"reflect"
	"testing"
)

func TestClauMappingAndReduction(t *testing.T) {
	from := vec.Vec_IntAllocArrayCopy([]int32{1, 3}, 2)
	to := vec.Vec_IntAllocArrayCopy([]int32{5, 2}, 2)
	mapping := clauMapping(from, to, 4)
	in := vec.Vec_IntAllocArrayCopy([]int32{2, 7}, 2)
	if got := vec.Vec_IntArray(clauRemap(mapping, in, true)); !reflect.DeepEqual(got, []int32{11, 4}) {
		t.Fatal(got)
	}
	main := vec.Vec_IntAllocArrayCopy([]int32{2, 7}, 2)
	other := vec.Vec_IntAllocArrayCopy([]int32{0, 2, 6}, 3)
	clauReduce(main, other)
	if got := vec.Vec_IntArray(main); !reflect.DeepEqual(got, []int32{2}) {
		t.Fatal(got)
	}
}
func TestClauSaveVariables(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	data := cnf.Cnf_Derive(g, 1)
	if got := vec.Vec_IntArray(clauLatchVars(g, data, true)); len(got) != 1 || got[0] != data.VarNums[lo.NodeID()] {
		t.Fatal(got)
	}
	if got := vec.Vec_IntArray(clauLatchVars(g, data, false)); len(got) != 1 || got[0] != data.VarNums[g.Outputs()[1].NodeID()] {
		t.Fatal(got)
	}
	if got := vec.Vec_IntArray(clauInputVars(g, data, 1)); len(got) != 1 || got[0] != data.VarNums[lo.NodeID()] {
		t.Fatal(got)
	}
	if got := vec.Vec_IntArray(clauOutputVars(g, data)); len(got) != 2 || got[0] != data.VarNums[g.Outputs()[0].NodeID()] {
		t.Fatal(got)
	}
}
