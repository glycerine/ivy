package dch

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestDchProvedImplicationsAndConstantShortcut(t *testing.T) {
	for _, flip := range []bool{false, true} {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		left := g.AND(g.AND(a, b), c)
		right := g.AND(a, g.AND(b, c))
		zero := g.AND(g.AND(a, b), g.AND(a, b.Not()))
		g.AddOutput(left)
		g.AddOutput(right)
		g.AddOutput(zero)
		pars := DefaultParams()
		pars.PolarFlip = flip
		m, err := newManager(g, &pars)
		if err != nil {
			t.Fatal(err)
		}
		m.recycleSolver()
		if _, err := m.satLiteral(left); err != nil {
			t.Fatal(err)
		}
		if _, err := m.satLiteral(right); err != nil {
			t.Fatal(err)
		}
		before := m.solver.ClauseCount()
		result, _, err := m.provePair(left, right)
		if err != nil || result != proofEquivalent {
			t.Fatalf("flip %v: result %v error %v", flip, result, err)
		}
		if m.solver.ClauseCount() < before+2 {
			t.Fatalf("flip %v: proved implications absent: before=%d after=%d", flip, before, m.solver.ClauseCount())
		}
		beforeCalls := m.stats.SATCalls
		result, _, err = m.provePair(g.Const1(), zero)
		if err != nil || result != proofEquivalent || m.stats.SATCalls != beforeCalls+1 {
			t.Fatalf("flip %v: constant result %v calls %d error %v", flip, result, m.stats.SATCalls-beforeCalls, err)
		}
	}
}

func TestDchSweepSubstitutesProvedNodesBeforeLaterCones(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(g.AND(a, b), c)
	y := g.AND(a, g.AND(b, c))
	z1, z2 := g.AND(x, d), g.AND(y, d)
	g.AddOutput(x)
	g.AddOutput(y)
	g.AddOutput(z1)
	g.AddOutput(z2)
	m, err := newManager(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.computeEquivalences(); err != nil {
		t.Fatal(err)
	}
	if m.mapped[x.NodeID()] != m.mapped[y.NodeID()] || m.mapped[z1.NodeID()] != m.mapped[z2.NodeID()] {
		t.Fatalf("proved nodes were not substituted in later cones: %v", m.mapped)
	}
	if m.stats.Proofs < 2 {
		t.Fatalf("expected both equivalences: %+v", m.stats)
	}
	for _, n := range []aig.Edge{x, y, z1, z2} {
		m.satGraph.AddOutput(m.mapped[n.NodeID()])
	}
	for bits := 0; bits < 16; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
		values := simulateAssignment(g, m.order, inputs)
		work, err := m.satGraph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for i, n := range []aig.Edge{x, y, z1, z2} {
			if work[i] != values[n.NodeID()] {
				t.Fatalf("bits %d node %d changed", bits, n.NodeID())
			}
		}
	}
}
