package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestStrSimWindowMiterComparesUnmatchedCones(t *testing.T) {
	makeGraph := func(invert bool) *aig.Graph {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		if invert {
			b = b.Not()
		}
		g.AddOutput(g.AND(a, b))
		aig.Aig_ManSetRegNum(g, 0)
		return g
	}
	a, b := makeGraph(false), makeGraph(true)
	m := newStrSimMatcher(a, b)
	w := m.windowMiter()
	if w.InputCount() != 2 || w.OutputCount() != 1 || w.RegisterCount() != 0 {
		t.Fatalf("window inputs=%d outputs=%d regs=%d", w.InputCount(), w.OutputCount(), w.RegisterCount())
	}
	if err := w.Check(); err != nil {
		t.Fatal(err)
	}
	// Cut inputs are allocated in graph traversal order; the miter equals a.
	for bits := 0; bits < 4; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0}
		outputs, err := w.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if outputs[0] != inputs[0] {
			t.Fatalf("input %v output %v", inputs, outputs)
		}
	}
}

func TestStrSimWindowMiterIncludesUnmatchedRegisters(t *testing.T) {
	makeGraph := func(invert bool) *aig.Graph {
		g := aig.NewGraph()
		a, lo := g.NewInput(), g.NewInput()
		g.AddOutput(lo)
		if invert {
			a = a.Not()
		}
		g.AddOutput(a)
		aig.Aig_ManSetRegNum(g, 1)
		return g
	}
	a, b := makeGraph(false), makeGraph(true)
	m := newStrSimMatcher(a, b)
	w := m.windowMiter()
	if w.RegisterCount() != 2 || w.InputCount() != 3 || w.OutputCount() != 3 {
		t.Fatalf("window inputs=%d outputs=%d regs=%d", w.InputCount(), w.OutputCount(), w.RegisterCount())
	}
	if err := w.Check(); err != nil {
		t.Fatal(err)
	}
	out, err := w.Evaluate([]bool{true, false, true})
	if err != nil {
		t.Fatal(err)
	}
	if !out[0] || !out[1] || out[2] {
		t.Fatalf("outputs %v", out)
	}
}
