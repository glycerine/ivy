package aig

import (
	"github.com/glycerine/abcgo/misc/tim"
	"testing"
)

func TestTraversalSupportAndComposition(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	ab := p.AND(a, b)
	root := p.OR(ab, c)
	p.AddOutput(root)
	if Aig_DagSize(p, root) != 2 || Aig_SupportSize(p, Aig_Regular(root)) != 3 {
		t.Fatal("wrong cone metrics")
	}
	order := Aig_ManDfs(p, 1)
	if len(order) != 2 || order[0] != ab {
		t.Fatal(order)
	}
	preorder := Aig_ManDfsPreorder(p, 1)
	if len(preorder) != 2 || preorder[0] != Aig_Regular(root) {
		t.Fatal(preorder)
	}
	composed := Aig_Compose(p, root, p.Const0(), 2)
	p.AddOutput(composed)
	for mask := 0; mask < 8; mask++ {
		x, y, z := mask&1 != 0, mask&2 != 0, mask&4 != 0
		out, err := p.Evaluate([]bool{x, y, z})
		if err != nil || out[0] != (x && y || z) || out[1] != (x && y) {
			t.Fatal(out, err)
		}
	}
	q := NewGraph()
	transferred := Aig_Transfer(p, q, root, 3)
	q.AddOutput(transferred)
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 8; mask++ {
		out, err := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
		if err != nil || out[0] != (mask&3 == 3 || mask&4 != 0) {
			t.Fatal(out, err)
		}
	}
	var cut []Edge
	Aig_ObjCollectCut(p, root, []Edge{a, b, c}, &cut)
	if len(cut) != 2 {
		t.Fatal(cut)
	}
	Aig_ManCheckMarkA(p)
	var super []Edge
	Aig_ObjCollectSuper(p, ab, &super)
	if len(super) != 2 {
		t.Fatal(super)
	}
	if Aig_ManVerifyTopoOrder(p) != 1 {
		t.Fatal("topology rejected")
	}
}

func TestReverseDfsIncludesDanglingNodes(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	first := p.AND(a, b)
	last := p.AND(first, c)
	dangling := p.AND(a, c)
	p.AddOutput(last)
	order := Aig_ManDfsReverse(p)
	positions := map[Edge]int{}
	for i, e := range order {
		positions[e] = i
	}
	if len(order) != 3 || positions[last] >= positions[first] {
		t.Fatal(order)
	}
	if _, ok := positions[dangling]; !ok {
		t.Fatal("dangling node missing")
	}
}

func TestChoiceLevelUsesMaximumAlternativeDepth(t *testing.T) {
	p := NewGraph()
	a, b, c, d := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
	ab := p.AND(a, b)
	balanced := p.AND(ab, p.AND(c, d))
	chain := p.AND(p.AND(ab, c), d)
	out := p.AddOutput(balanced)
	p.equivs = map[int]Edge{balanced.NodeID(): chain}
	if Aig_ManChoiceLevel(p) != 3 || p.node(balanced).Level != 3 || p.node(out).Level != 3 {
		t.Fatal("choice class underestimated maximum level")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestChoiceLevelsAcrossTimingBox(t *testing.T) {
	p := NewGraph()
	a, b, boxOutput := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(a, b))
	p.AddOutput(p.AND(a, boxOutput))
	manager := tim.Tim_ManStart(3, 2)
	tim.Tim_ManCreateBox(manager, 0, 1, 2, 1, 0, 0)
	p.Timing = manager
	if Aig_ManChoiceLevel(p) != 3 || p.node(boxOutput).Level != 2 {
		t.Fatal("box boundary was ignored")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
