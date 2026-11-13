package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestWindowExtractMiterFromRepresentatives(t *testing.T) {
	build := func(invert bool) *aig.Graph {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		if invert {
			b = b.Not()
		}
		g.AddOutput(g.AND(a, b))
		return g
	}
	a, b := build(false), build(true)
	for i := 0; i < 3; i++ {
		x := aig.Edge(i) << 1
		y := aig.Edge(i) << 1
		aig.Aig_ObjSetRepr(a, x, y)
		aig.Aig_ObjSetRepr(b, y, x)
	}
	w := Saig_ManWindowExtractMiter(a, b)
	if w.InputCount() != 2 || w.OutputCount() != 1 {
		t.Fatalf("inputs=%d outputs=%d", w.InputCount(), w.OutputCount())
	}
	if err := w.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := w.Evaluate(in)
		if err != nil || out[0] != in[0] {
			t.Fatalf("%v -> %v (%v)", in, out, err)
		}
	}
}
func TestWindowCollectedDiffNodes(t *testing.T) {
	a, b := aig.NewGraph(), aig.NewGraph()
	x, y := a.NewInput(), b.NewInput()
	a.AddOutput(x)
	b.AddOutput(y)
	aig.Aig_ObjSetRepr(a, a.Const1(), b.Const1())
	aig.Aig_ObjSetRepr(b, b.Const1(), a.Const1())
	v := Saig_ManCollectedDiffNodes(a, b)
	if vec.Vec_PtrSize(v) != 1 || vec.Vec_PtrEntry(v, 0).(aig.Edge) != x {
		t.Fatalf("nodes %v", vec.Vec_PtrArray(v))
	}
}
