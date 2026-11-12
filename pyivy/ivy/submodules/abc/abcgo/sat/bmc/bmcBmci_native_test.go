package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func stateRegister(nextIsState bool) *aig.Graph {
	g := aig.NewGraph()
	ro := g.NewInput()
	g.AddOutput(ro) // property output
	next := g.Const0()
	if nextIsState {
		next = ro
	}
	g.AddOutput(next)
	aig.Aig_ManSetRegNum(g, 1)
	return g
}

func TestCompareInitialStatesAIG(t *testing.T) {
	identity := stateRegister(true)
	if got := CompareInitialStatesAIG(identity, []bool{false}, []bool{true}, 1, 0); got != 0 {
		t.Fatalf("identity transition status = %d, want distinguishable (0)", got)
	}

	sharedInput := aig.NewGraph()
	pi, ro := sharedInput.NewInput(), sharedInput.NewInput()
	sharedInput.AddOutput(ro)
	sharedInput.AddOutput(pi)
	aig.Aig_ManSetRegNum(sharedInput, 1)
	if got := CompareInitialStatesAIG(sharedInput, []bool{false}, []bool{true}, 2, 0); got != 1 {
		t.Fatalf("shared-input transition status = %d, want equivalent (1)", got)
	}

	reset := stateRegister(false)
	if got := CompareInitialStatesAIG(reset, []bool{false}, []bool{true}, 1, 0); got != 1 {
		t.Fatalf("reset transition status = %d, want equivalent (1)", got)
	}
	if got := CompareInitialStatesAIG(identity, []bool{false}, []bool{true}, 0, 0); got != 1 {
		t.Fatalf("zero-frame status = %d, want equivalent (1)", got)
	}
	if got := compareInitialStateLiteralsAIG(identity, []int32{2}, []int32{0}, 1, 0); got != 0 {
		t.Fatalf("symbolic initialization status = %d, want distinguishable (0)", got)
	}
	if got := CompareInitialStatesAIG(identity, nil, []bool{true}, 1, 0); got != -1 {
		t.Fatalf("invalid initialization status = %d, want -1", got)
	}
}
