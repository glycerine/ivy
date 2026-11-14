package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestFlipFirstPoChangesDriverNotTerminal(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b.Not())
	first := g.AddOutput(root)
	second := g.AddOutput(a.Not())
	before := *g.node(first)
	Aig_ManFlipFirstPo(g)
	if g.outputs[0] != first || g.outputs[1] != second {
		t.Error("flipping a PO changed terminal handles")
	}
	if g.node(first).Fanins[0] != root.Not() {
		t.Error("first PO child was not complemented as in C")
	}
	if g.node(second).Fanins[0] != a.Not() || g.node(root).Refs != 1 {
		t.Error("flip changed another output or reference ownership")
	}
	for mask := 0; mask < 4; mask++ {
		got, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || got[0] != (mask != 1) || got[1] != (mask&1 == 0) {
			t.Fatalf("assignment %d: outputs %v (%v)", mask, got, err)
		}
	}
	Aig_ManFlipFirstPo(g)
	if g.outputs[0] != first || *g.node(first) != before {
		t.Fatal("double flip did not restore the original output")
	}
}

func TestFlipFirstPoAgainstCompiledSource(t *testing.T) {
	var want strings.Builder
	for driver := Edge(0); driver < 8; driver++ {
		g := NewGraph()
		a, b := g.NewInput(), g.NewInput()
		g.AND(a, b)
		out := g.AddOutput(driver)
		fmt.Fprintf(&want, "co %d:", out.NodeID())
		for flip := 0; flip < 2; flip++ {
			Aig_ManFlipFirstPo(g)
			fmt.Fprintf(&want, " %d", g.node(out).Fanins[0])
		}
		want.WriteByte('\n')
	}
	got := runAigSourceFixture(t, "flip_po_compare.c", "aigMan.c")
	if got != want.String() {
		t.Fatalf("source flips:\n%sGo flips:\n%s", got, want.String())
	}
}
