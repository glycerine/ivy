package dch

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestNonTFOSimulationRefinesOnlyQueriedClass(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	pars := DefaultParams()
	pars.SimulateTFO = false
	m, err := newManager(g, &pars)
	if err != nil {
		t.Fatal(err)
	}
	m.classes = &Classes{groups: [][]aig.Edge{{a, b}, {c, d}}}
	m.classes.index()
	// This assignment distinguishes both classes, but the C non-TFO path
	// explicitly resimulates and refines only the queried representative class.
	m.refineCounterexample([]bool{true, false, true, false}, b, a)
	if got := m.classes.representative(b); got != aig.InvalidEdge {
		t.Fatalf("queried class not split: %v", got)
	}
	if got := m.classes.representative(d); got != c {
		t.Fatalf("unrelated class changed: representative %v, want %v", got, c)
	}
}

func TestNonTFOConstantCandidatesUseSourceIDWindow(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 502)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	pars := DefaultParams()
	pars.SimulateTFO = false
	m, err := newManager(g, &pars)
	if err != nil {
		t.Fatal(err)
	}
	// IDs 2..501 are the 500-entry window; 1 and 502 are outside it.
	m.classes = &Classes{groups: [][]aig.Edge{{g.Const1(), inputs[0], inputs[1], inputs[500], inputs[501]}}}
	m.classes.index()
	values := make([]bool, len(inputs))
	for i := range values {
		values[i] = true
	}
	m.refineCounterexample(values, inputs[1], g.Const1())
	if m.classes.representative(inputs[0]) != g.Const1() || m.classes.representative(inputs[501]) != g.Const1() {
		t.Fatal("constant candidate outside the window was refined")
	}
	if m.classes.representative(inputs[1]) != aig.InvalidEdge || m.classes.representative(inputs[500]) != inputs[1] {
		t.Fatalf("window candidates not grouped: %v", m.classes.groups)
	}
}
