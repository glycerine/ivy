package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestSaigManSupportFindsUniqueLatchInputs(t *testing.T) {
	p := aig.NewGraph()
	pi := p.NewInput()
	lo0 := p.NewInput()
	lo1 := p.NewInput()
	poValue := p.AND(lo0, lo1)
	li0Value := p.AND(lo0, pi)
	li1Value := pi
	po := p.AddOutput(poValue)
	li0 := p.AddOutput(li0Value)
	li1 := p.AddOutput(li1Value)
	aig.Aig_ManSetRegNum(p, 2)

	got := Saig_ManSupport(p, []aig.Edge{po, po})
	want := []aig.Edge{li0, li1}
	if len(got) != len(want) {
		t.Fatalf("support length = %d, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("support[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestSaigManSupportIgnoresPrimaryInputsAndConstants(t *testing.T) {
	p := aig.NewGraph()
	pi := p.NewInput()
	lo := p.NewInput()
	po := p.AddOutput(p.OR(pi, p.Const0()))
	li := p.AddOutput(lo)
	aig.Aig_ManSetRegNum(p, 1)

	if got := Saig_ManSupport(p, []aig.Edge{po}); len(got) != 0 {
		t.Fatalf("primary-only cone has state support: %v", got)
	}
	if got := Saig_ManSupport(p, []aig.Edge{li}); len(got) != 1 || got[0] != li {
		t.Fatalf("latch self-input support = %v, want [%v]", got, li)
	}
}
