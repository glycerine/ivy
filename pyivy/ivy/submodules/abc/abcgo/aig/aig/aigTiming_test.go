package aig

import "testing"

func TestRequiredAndReverseLevels(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	first := p.AND(a, b)
	last := p.AND(first, c)
	p.AddOutput(last)
	Aig_ManFanoutStart(p)
	defer Aig_ManFanoutStop(p)
	Aig_ManStartReverseLevels(p, 2)
	if Aig_ObjReverseLevel(p, last) != 1 || Aig_ObjReverseLevel(p, first) != 2 || Aig_ObjRequiredLevel(p, first) != 3 {
		t.Fatal(p.reverseLevels)
	}
	if err := Aig_ManVerifyReverseLevel(p); err != nil {
		t.Fatal(err)
	}
	Aig_ObjSetReverseLevel(p, first, 7)
	if Aig_ManVerifyReverseLevel(p) == nil {
		t.Fatal("corruption accepted")
	}
	Aig_ManUpdateReverseLevel(p, first)
	if err := Aig_ManVerifyReverseLevel(p); err != nil {
		t.Fatal(err)
	}
	Aig_ManStopReverseLevels(p)
}
func TestArrivalLevelsSurviveRecomputation(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.node(a).Level = 4
	n := p.AND(a, b)
	p.AddOutput(n)
	Aig_ManFanoutStart(p)
	defer Aig_ManFanoutStop(p)
	Aig_ManUpdateLevel(p, n)
	if p.node(a).Level != 4 || p.node(n).Level != 5 {
		t.Fatal("input arrival level lost")
	}
	if err := Aig_ManVerifyLevel(p); err != nil {
		t.Fatal(err)
	}
	p.node(n).Level = 2
	if Aig_ManVerifyLevel(p) == nil {
		t.Fatal("level corruption accepted")
	}
}
