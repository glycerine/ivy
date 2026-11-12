package aig

import "testing"

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
