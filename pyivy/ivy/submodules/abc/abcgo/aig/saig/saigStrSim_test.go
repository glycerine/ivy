package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestStrSimPerformMatchingAndWindow(t *testing.T) {
	makeGraph := func(invert bool) *aig.Graph {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		if invert {
			b = b.Not()
		}
		g.AddOutput(g.AND(a, b))
		return g
	}
	a, b := makeGraph(false), makeGraph(true)
	var w *aig.Graph
	pairs := Saig_StrSimPerformMatching(a, b, 0, 0, &w)
	if pairs == nil || vec.Vec_IntSize(pairs) != 6 {
		t.Fatalf("pairs %v", pairs)
	}
	if w == nil || w.OutputCount() != 1 || w.InputCount() != 2 {
		t.Fatalf("window %v", w)
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := w.Evaluate(in)
		if err != nil || out[0] != in[0] {
			t.Fatalf("%v: %v %v", in, out, err)
		}
	}
}

func TestStrSimPerformMatchingIdentical(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	var w *aig.Graph
	pairs := Saig_StrSimPerformMatching(g, g, 0, 0, &w)
	if pairs == nil || vec.Vec_IntSize(pairs) != 8 || w == nil || w.OutputCount() != 0 {
		t.Fatalf("pairs=%v window=%v", pairs, w)
	}
}
