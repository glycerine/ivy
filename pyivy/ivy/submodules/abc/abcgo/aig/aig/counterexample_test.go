package aig

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestGraphCounterexampleOwnership(t *testing.T) {
	p := NewGraph()
	a := p.NewInput()
	p.NewInput()
	p.AddOutput(a)
	p.nRegs = 1
	input := []bool{true, false}
	if err := p.SetCombinationalCounterexample(input, 0); err != nil {
		t.Fatal(err)
	}
	input[0] = false
	m := p.CombinationalCounterexample()
	if m.PData[0] != 2 || m.NRegs != 1 || m.NPis != 1 {
		t.Fatal("incorrect state/PI packing", m)
	}
	if err := p.SetCombinationalCounterexample(nil, 0); err == nil || p.CombinationalCounterexample() != m {
		t.Fatal("invalid call replaced model")
	}
	if p.TakeCombinationalCounterexample() != m || p.CombinationalCounterexample() != nil {
		t.Fatal("ownership transfer")
	}
	p.SetCombinationalCounterexample([]bool{false, true}, -1)
	if p.CombinationalCounterexample().PData[0] != 1 {
		t.Fatal("register bit")
	}
	p.ClearCombinationalCounterexample()
	if p.CombinationalCounterexample() != nil {
		t.Fatal("clear")
	}
}

func TestInductionCounterexampleTake(t *testing.T) {
	p := NewGraph()
	p.NewInput()
	p.NewInput()
	p.NewInput()
	p.nRegs = 2
	witness := &util.Abc_Cex_t{NRegs: 1, NPis: 1, NBits: 2, PData: []uint32{3}}
	if err := p.SetInductionCounterexample(witness); err != nil {
		t.Fatal(err)
	}
	if got := p.TakeInductionCounterexample(); got != witness {
		t.Fatalf("TakeInductionCounterexample() = %p, want %p", got, witness)
	}
	if p.InductionCounterexample() != nil || p.TakeInductionCounterexample() != nil {
		t.Fatal("take did not clear the saved witness")
	}
	if err := p.SetInductionCounterexample(witness); err != nil {
		t.Fatal(err)
	}
	p.ClearInductionCounterexample()
	if p.InductionCounterexample() != nil {
		t.Fatal("clear did not discard the saved witness")
	}
}
