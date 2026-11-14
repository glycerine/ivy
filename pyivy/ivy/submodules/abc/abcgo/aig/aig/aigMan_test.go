package aig

import (
	"fmt"
	"testing"
)

func TestManagerStartAcceptsUnknownCapacityHints(t *testing.T) {
	for _, hint := range []int32{0, -1, -1000, -1 << 31} {
		t.Run(fmt.Sprint(hint), func(t *testing.T) {
			defer func() {
				if failure := recover(); failure != nil {
					t.Errorf("source accepts nonpositive soft limits: %v", failure)
				}
			}()
			g := Aig_ManStart(hint)
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(a, b.Not()))
			for mask := 0; mask < 4; mask++ {
				out, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
				if err != nil || out[0] != (mask == 1) {
					t.Fatalf("assignment %d: %v (%v)", mask, out, err)
				}
			}
			if err := g.Check(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagerCleanupKeepsStableEdges(t *testing.T) {
	p := Aig_ManStart(32)
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	live := p.AND(a, b)
	dead := p.AND(a, c)
	p.AddOutput(live)
	if Aig_ManCleanup(p) != 1 || p.valid(dead) || !p.valid(live) {
		t.Fatal("incorrect cleanup")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	if Aig_ManCiCleanup(p) != 1 {
		t.Fatal("unused input retained")
	}
	if p.NodeCount() != 5 || Aig_ManObjNum(p) != 5 || len(p.LiveEdges()) != 5 {
		t.Fatal("deleted objects counted")
	}
	values, err := p.Evaluate([]bool{true, true})
	if err != nil || !values[0] {
		t.Fatal(values, err)
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestManagerCopyAndMiter(t *testing.T) {
	p := Aig_ManStart(0)
	a, b := p.NewInput(), p.NewInput()
	x := p.XOR(a, b)
	q := Aig_ManExtractMiter(p, x, x)
	for mask := 0; mask < 4; mask++ {
		out, err := q.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || out[0] {
			t.Fatal(out, err)
		}
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	q.AddOutput(q.Const0())
	if Aig_ManCoCleanup(q) != 2 {
		t.Fatal("constant zero outputs retained")
	}
	Aig_ManStopP(&q)
	if q != nil {
		t.Fatal("manager pointer not cleared")
	}
}
func TestManagerRegisterCleanup(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(p.Const0())
	p.AddOutput(b)
	Aig_ManSetRegNum(p, 1)
	if Aig_ManCiCleanup(p) != 1 || p.valid(a) {
		t.Fatal("incorrect CI cleanup")
	}
	if Aig_ManCoCleanup(p) != 1 || len(p.outputs) != 1 {
		t.Fatal("incorrect CO cleanup")
	}
	if p.inputs[0] != b {
		t.Fatal("register input was removed")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
