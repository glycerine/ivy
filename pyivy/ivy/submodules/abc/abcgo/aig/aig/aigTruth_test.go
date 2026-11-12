package aig

import "testing"

func TestCutTruthAndChecks(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	e := p.XOR(a, b)
	table, err := Aig_ManCutTruth(p, e, []Edge{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if table[0]&15 != 6 {
		t.Fatalf("xor table %x", table)
	}
	inv, err := Aig_ManCutTruth(p, e.Not(), []Edge{a, b})
	if err != nil || inv[0]&15 != 9 {
		t.Fatal(inv, err)
	}
	if _, err := Aig_ManCutTruth(p, e, []Edge{a}); err == nil {
		t.Fatal("accepted incomplete cut")
	}
	if Aig_ManCheck(p) != 1 {
		t.Fatal("valid manager rejected")
	}
	Aig_ManCheckMarkA(p)
	Aig_ManCheckPhase(p)
	p.nodes[e.NodeID()].Phase = !p.nodes[e.NodeID()].Phase
	defer func() {
		if recover() == nil {
			t.Fatal("bad phase accepted")
		}
	}()
	Aig_ManCheckPhase(p)
}
