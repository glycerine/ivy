package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestDuplicateDualRailUnknownOutput(t *testing.T) {
	src := aig.NewGraph()
	x := src.NewInput()
	src.AddOutput(x)
	dual, err := DuplicateDualRail(src, nil, 1, false, false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(dual.Inputs()) != 2 || len(dual.Outputs()) != 1 {
		t.Fatalf("dual interface = %d inputs, %d outputs", len(dual.Inputs()), len(dual.Outputs()))
	}
	for _, tc := range []struct {
		inputs []bool
		want   bool
	}{{[]bool{false, false}, true}, {[]bool{false, true}, false}, {[]bool{true, false}, false}, {[]bool{true, true}, true}} {
		got, err := dual.Evaluate(tc.inputs)
		if err != nil {
			t.Fatal(err)
		}
		if got[0] != tc.want {
			t.Errorf("Evaluate(%v) = %v, want %v", tc.inputs, got[0], tc.want)
		}
	}
}

func TestBlockPrimaryOutputsAddsRegisters(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	Saig_ManBlockPo(g, 2)
	if got := aig.Aig_ManRegNum(g); got != 2 {
		t.Fatalf("registers = %d, want 2", got)
	}
	if got := len(g.Outputs()); got != 3 {
		t.Fatalf("outputs = %d, want property plus two LIs", got)
	}
}
