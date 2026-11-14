package aigpartsat

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func TestPartitionSATUsesMappedCnfVariableLayout(t *testing.T) {
	g := aig.NewGraph()
	var pio []int32
	root := g.Const1()
	for i := 0; i < 6; i++ {
		input := g.NewInput()
		pio = append(pio, int32(input.NodeID()))
		root = g.AND(root, input)
	}
	g.AddOutput(root)
	pio = append(pio, int32(root.NodeID()))
	aig.Aig_ManPartResetNodePolarity(g)
	data := cnf.Cnf_Derive(g, 1)
	if data == nil {
		t.Fatal("mapped CNF derivation failed")
	}
	defer cnf.Cnf_DataFree(data)
	solver := bsat.NewSolver()
	defer solver.Close()
	variables := NewNodeVariables(int(aig.Aig_ManObjNumMax(g)))
	if Aig_ManAddNewCnfToSolver(solver, g, variables, pio, nil, 0) != 0 {
		t.Fatal("unexpected inconsistent CNF")
	}
	if int(solver.NumVars()) != data.Variables {
		t.Errorf("partition used %d variables, mapped CNF uses %d", int(solver.NumVars()), data.Variables)
	}
	terminals := append(g.Inputs(), g.Outputs()...)
	for i, e := range terminals {
		if got, want := variables.IDs[pio[i]], data.VarNums[e.NodeID()]; got != want {
			t.Errorf("terminal %d variable %d, mapped CNF variable %d", i, got, want)
		}
	}
}

func TestPartitionSATSourceEntryUsesDarSynthesis(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	if status := Aig_ManPartitionedSat(g, 0, 10, 0, 10000, 0, 1, 0); status != 0 {
		t.Fatalf("source synthesis path returned %d, want SAT", status)
	}
}
