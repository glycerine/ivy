package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestUnfoldConstraints2(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	q := unfoldConstraints2(g, unfold2Classes{typeI: []aig.Edge{pi.Not()}, typeII: []aig.Edge{lo}})
	if q.InputCount() != 3 || q.OutputCount() != 5 || q.RegisterCount() != 2 || aig.Aig_ManConstrNum(q) != 2 {
		t.Fatalf("inputs %d outputs %d regs %d constraints %d", q.InputCount(), q.OutputCount(), q.RegisterCount(), aig.Aig_ManConstrNum(q))
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{in[1], !in[0], in[2] && in[1], in[0], in[1]}
		for i := range want {
			if out[i] != want[i] {
				t.Fatalf("%v -> %v want %v", in, out, want)
			}
		}
	}
}
func TestFoldConstraints2FreezesState(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	g.AddOutput(pi.Not())
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	q := Saig_ManDupFoldConstrsFunc2(g, 0, 0, 0)
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	if q.InputCount() != 3 || q.OutputCount() != 3 || q.RegisterCount() != 2 {
		t.Fatalf("inputs %d outputs %d regs %d", q.InputCount(), q.OutputCount(), q.RegisterCount())
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		out, err := q.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		fail := in[0] || in[2]
		next := !in[0]
		if fail {
			next = in[1]
		}
		want := []bool{in[1] && !fail, next, fail}
		for i := range want {
			if out[i] != want[i] {
				t.Fatalf("%v -> %v want %v", in, out, want)
			}
		}
	}
}
func TestUnfoldConstraints2Discovery(t *testing.T) {
	g := aig.NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo.Not())
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	var typeII int32
	q := Saig_ManDupUnfoldConstrsFunc2(g, 1, 0, 0, 0, 0, &typeII)
	if q == nil || aig.Aig_ManConstrNum(q) != 1 || typeII != 0 {
		t.Fatalf("graph %v count %d", q, typeII)
	}
}
