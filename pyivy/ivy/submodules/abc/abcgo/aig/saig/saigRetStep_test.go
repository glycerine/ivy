package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestRetimeStepsBackwardRedirectsLatchOutput(t *testing.T) {
	g := aig.NewGraph()
	a, b, lo := g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	g.AddOutput(lo.Not())
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 1)
	if n := Saig_ManRetimeSteps(g, 2, 0, 0); n != 1 {
		t.Fatalf("got %d steps, want 1", n)
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	if g.RegisterCount() != 3 {
		t.Fatalf("got %d registers", g.RegisterCount())
	}
	for bits := 0; bits < 16; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, false, bits&4 != 0, bits&8 != 0}
		outputs, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if outputs[0] != !(inputs[3] && inputs[4]) || outputs[2] != inputs[0] || outputs[3] != inputs[1] {
			t.Fatalf("inputs=%v outputs=%v", inputs, outputs)
		}
	}
}

func TestRetimeStepsForward(t *testing.T) {
	g := aig.NewGraph()
	a, b, lo0, lo1 := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(lo0, lo1)
	g.AddOutput(gate)
	g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 2)
	if n := Saig_ManRetimeSteps(g, 1, 1, 0); n != 1 {
		t.Fatalf("got %d steps", n)
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, false, false, bits&4 != 0}
		out, err := g.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != in[4] || out[3] != (in[0] && in[1]) {
			t.Fatalf("inputs=%v outputs=%v", in, out)
		}
	}
}
