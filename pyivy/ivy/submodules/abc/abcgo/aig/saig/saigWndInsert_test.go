package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestWindowExtractInsertSequential(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi.Not(), lo)
	g.AddOutput(gate)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 1)
	window := Saig_ManWindowExtract(g, lo, 2)
	q := Saig_ManWindowInsert(g, lo, 2, window)
	if q == nil || q.InputCount() != 2 || q.OutputCount() != 2 || q.RegisterCount() != 1 {
		t.Fatalf("graph %v", q)
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		a, e1 := g.Evaluate(in)
		b, e2 := q.Evaluate(in)
		if e1 != nil || e2 != nil || a[0] != b[0] || a[1] != b[1] {
			t.Fatalf("%v original=%v rebuilt=%v", in, a, b)
		}
	}
}
func TestWindowInsertReplacement(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	g.AddOutput(gate)
	w := aig.NewGraph()
	x, y := w.NewInput(), w.NewInput()
	w.AddOutput(w.OR(x, y))
	q := Saig_ManWindowInsert(g, gate, 2, w)
	if q == nil {
		t.Fatal("replacement rejected")
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] != (in[0] || in[1]) {
			t.Fatalf("%v -> %v (%v)", in, out, err)
		}
	}
	bad := aig.NewGraph()
	bad.NewInput()
	bad.AddOutput(bad.Const1())
	if Saig_ManWindowInsert(g, gate, 2, bad) != nil {
		t.Fatal("accepted mismatched window")
	}
}
