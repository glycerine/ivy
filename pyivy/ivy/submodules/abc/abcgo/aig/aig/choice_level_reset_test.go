package aig

import "testing"

func TestChoiceLevelResetsAllObjectLevels(t *testing.T) {
	p := NewGraph()
	a, b, unused := p.NewInput(), p.NewInput(), p.NewInput()
	gate := p.AND(a, b)
	dangling := p.AND(a, unused)
	out := p.AddOutput(gate)
	p.node(a).Level = 7
	p.node(b).Level = 9
	p.node(unused).Level = 11
	p.node(dangling).Level = 42
	if got := Aig_ManChoiceLevel(p); got != 1 {
		t.Errorf("choice level = %d, want 1", got)
	}
	for _, e := range []Edge{a, b, unused, dangling} {
		if got := p.node(e).Level; got != 0 {
			t.Errorf("node %d level = %d, want 0", e.NodeID(), got)
		}
	}
	if p.node(gate).Level != 1 || p.node(out).Level != 1 {
		t.Errorf("gate/output levels = %d/%d, want 1/1", p.node(gate).Level, p.node(out).Level)
	}
}
