package aig

import "testing"

func TestPatchOutputPreservesUnrelatedCachedLevels(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	other := g.AddOutput(g.AND(a, b))
	patched := g.AddOutput(b)
	g.node(other).Level = 99
	before := *g.node(other)
	Aig_ManFanoutStart(g)
	Aig_ObjPatchFanin0(g, patched, a.Not())
	if *g.node(other) != before {
		t.Fatalf("patching another CO changed this CO: got %+v, want %+v", *g.node(other), before)
	}
	if g.node(patched).Level != 0 || !g.node(patched).Phase {
		t.Fatalf("patched CO cache = %+v", *g.node(patched))
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}
