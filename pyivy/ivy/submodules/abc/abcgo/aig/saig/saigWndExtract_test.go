package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestWindowExtractSequential(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	gate := g.AND(pi.Not(), lo)
	g.AddOutput(gate)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 1)
	q := Saig_ManWindowExtract(g, lo, 2)
	if q == nil || q.InputCount() != 2 || q.OutputCount() != 2 || q.RegisterCount() != 1 {
		t.Fatalf("window %v", q)
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := q.Evaluate(in)
		want := !in[0] && in[1]
		if err != nil || out[0] != want || out[1] != want {
			t.Fatalf("%v -> %v (%v)", in, out, err)
		}
	}
	if got := Saig_ManFindPivot(g); got != lo {
		t.Fatalf("pivot %v", got)
	}
}
func TestWindowExtractCombinational(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 0)
	q := Saig_ManWindowExtract(g, gate, 2)
	if q.InputCount() != 2 || q.OutputCount() != 1 || q.RegisterCount() != 0 {
		t.Fatal("incorrect interface")
	}
	out, err := q.Evaluate([]bool{true, true})
	if err != nil || !out[0] {
		t.Fatalf("%v %v", out, err)
	}
}
