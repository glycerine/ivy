package gia

import "testing"

func TestSwapCombinationalCounterexamples(t *testing.T) {
	a, b := Gia_ManStart(4), Gia_ManStart(4)
	Gia_ManAppendCi(a)
	Gia_ManAppendCi(b)
	if err := a.SetCombinationalCounterexample([]bool{true}, -1); err != nil {
		t.Fatal(err)
	}
	if err := b.SetCombinationalCounterexample([]bool{false}, -1); err != nil {
		t.Fatal(err)
	}
	first, second := a.CombinationalCounterexample(), b.CombinationalCounterexample()
	SwapCombinationalCounterexamples(a, b)
	if a.CombinationalCounterexample() != second || b.CombinationalCounterexample() != first {
		t.Fatal("counterexamples were not swapped")
	}
}
