package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestRetimeNodeBackward(t *testing.T) {
	for mask := 0; mask < 8; mask++ {
		g := aig.NewGraph()
		a, b, lo := g.NewInput(), g.NewInput(), g.NewInput()
		x, y := a, b
		if mask&1 != 0 {
			x = x.Not()
		}
		if mask&2 != 0 {
			y = y.Not()
		}
		node := g.AND(x, y)
		if mask&4 != 0 {
			node = node.Not()
		}
		g.AddOutput(lo)
		g.AddOutput(node)
		aig.Aig_ManSetRegNum(g, 1)
		result := Saig_ManRetimeNodeBwd(g, lo)
		g.AddOutput(result)
		if g.RegisterCount() != 3 {
			t.Fatalf("mask=%d: got %d registers", mask, g.RegisterCount())
		}
		for values := 0; values < 16; values++ {
			in := []bool{values&1 != 0, values&2 != 0, false, values&4 != 0, values&8 != 0}
			out, err := g.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			c0, c1 := (mask&1 != 0) != (mask&4 != 0), (mask&2 != 0) != (mask&4 != 0)
			if out[2] != (in[0] != c0) || out[3] != (in[1] != c1) || out[4] != ((in[3] != c0) && (in[4] != c1)) {
				t.Fatalf("mask=%d inputs=%v outputs=%v", mask, in, out)
			}
		}
	}
}

func TestRetimeNodeBackwardRejectsNonGate(t *testing.T) {
	g := aig.NewGraph()
	in, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(in)
	aig.Aig_ManSetRegNum(g, 1)
	if got := Saig_ManRetimeNodeBwd(g, lo); got != aig.InvalidEdge {
		t.Fatalf("got %v", got)
	}
	if g.RegisterCount() != 1 {
		t.Fatal("register count changed")
	}
}
