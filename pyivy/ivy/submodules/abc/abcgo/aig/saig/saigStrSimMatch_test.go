package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestStrSimMatchingRound(t *testing.T) {
	makeGraph := func(reverse bool) (*aig.Graph, aig.Edge) {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		if reverse {
			a, b = b, a
		}
		n := g.AND(a, b.Not())
		g.AddOutput(n)
		return g, n
	}
	a, aGate := makeGraph(false)
	b, bGate := makeGraph(false)
	m := newStrSimMatcher(a, b)
	m.round()
	if n := m.detectUnique(); n != 1 {
		t.Fatalf("found %d matches", n)
	}
	if got := m.pair[0][aGate]; got != bGate {
		t.Fatalf("got %v want %v", got, bGate)
	}
	m.round()
	if n := m.detectUnique(); n != 0 {
		t.Fatalf("found %d extra matches", n)
	}
	c, _ := makeGraph(true)
	other := newStrSimMatcher(a, c)
	other.round()
	if n := other.detectUnique(); n != 0 {
		t.Fatalf("matched different functions: %d", n)
	}
}

func TestStrSimMatchingRejectsAmbiguousAndConstantSignatures(t *testing.T) {
	makeGraph := func() (*aig.Graph, aig.Edge, aig.Edge) {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		x, y := g.AND(a, b), g.AND(a, c)
		g.AddOutput(x)
		g.AddOutput(y)
		return g, x, y
	}
	a, ax, ay := makeGraph()
	b, bx, by := makeGraph()
	m := newStrSimMatcher(a, b)
	for i := 0; i < SAIG_WORDS; i++ {
		m.sim[0].words[ax.NodeID()][i], m.sim[0].words[ay.NodeID()][i] = 42, 42
		m.sim[1].words[bx.NodeID()][i], m.sim[1].words[by.NodeID()][i] = 42, 17
	}
	if n := m.detectUnique(); n != 0 {
		t.Fatalf("matched ambiguous signature: %d", n)
	}
	for i := 0; i < SAIG_WORDS; i++ {
		m.sim[0].words[ax.NodeID()][i], m.sim[1].words[bx.NodeID()][i] = 0, 0
	}
	if n := m.detectUnique(); n != 0 {
		t.Fatalf("matched constant signature: %d", n)
	}
}

func TestStrSimPruneInconsistent(t *testing.T) {
	makeGraph := func() (*aig.Graph, aig.Edge, aig.Edge) {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		x, y := g.AND(a, b), g.AND(a, c)
		g.AddOutput(x)
		g.AddOutput(y)
		return g, x, y
	}
	a, ax, ay := makeGraph()
	b, bx, by := makeGraph()
	m := newStrSimMatcher(a, b)
	m.link(ax, bx)
	m.link(ay, by)
	if removed := m.pruneInconsistent(); removed != 0 {
		t.Fatalf("removed %d valid matches", removed)
	}
	m.link(ax, by)
	delete(m.pair[1], bx)
	delete(m.pair[0], ay)
	if removed := m.pruneInconsistent(); removed != 1 {
		t.Fatalf("removed %d inconsistent matches", removed)
	}
}

func TestStrSimExtendPreservesPrimaryInputs(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	y := g.AND(x, c)
	g.AddOutput(y)
	h := aig.NewGraph()
	aa, bb, cc := h.NewInput(), h.NewInput(), h.NewInput()
	xx := h.AND(aa, bb)
	yy := h.AND(xx, cc)
	h.AddOutput(yy)
	m := newStrSimMatcher(g, h)
	m.link(x, xx)
	m.extend(1)
	if _, ok := m.pair[0][x]; ok {
		t.Fatal("matched boundary gate retained")
	}
	if _, ok := m.pair[0][a]; !ok {
		t.Fatal("primary input match removed")
	}
}
