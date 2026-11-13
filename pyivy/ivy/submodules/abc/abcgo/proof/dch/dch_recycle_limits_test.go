package dch

import (
	"testing"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// dchSat.c recycles only if nSatVarMax is nonzero AND nSatVars exceeds
// it AND nCallsSince exceeds nCallsRecycle; both comparisons are strict.
func TestSourceSolverRecycleThresholdBoundaries(t *testing.T) {
	newProof := func(t *testing.T) (*Manager, aig.Edge, aig.Edge) {
		t.Helper()
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		x, y := g.AND(g.AND(a, b), c), g.AND(a, g.AND(b, c))
		m, err := newManager(g, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, _, err := m.provePair(x, y)
		if err != nil || result != proofEquivalent {
			t.Fatalf("initial proof %v: %v", result, err)
		}
		return m, x, y
	}
	check := func(t *testing.T, m *Manager, x, y aig.Edge, want int) {
		t.Helper()
		result, _, err := m.provePair(x, y)
		if err != nil || result != proofEquivalent {
			t.Fatalf("proof %v: %v", result, err)
		}
		if m.stats.Recycles != want {
			t.Fatalf("recycles = %d, want %d", m.stats.Recycles, want)
		}
	}
	t.Run("strict thresholds", func(t *testing.T) {
		m, x, y := newProof(t)
		m.params.SatVarMax, m.params.CallsRecycle = int(m.nextVar), 0
		check(t, m, x, y, 1) // equality with the variable limit is insufficient
		m.params.SatVarMax, m.params.CallsRecycle = int(m.nextVar)-1, 2
		check(t, m, x, y, 1) // equality with the call limit is insufficient
		check(t, m, x, y, 2) // both limits now strictly exceeded
	})
	t.Run("zero disables limit", func(t *testing.T) {
		m, x, y := newProof(t)
		m.params.SatVarMax, m.params.CallsRecycle = 0, 0
		for i := 0; i < 5; i++ {
			check(t, m, x, y, 1)
		}
	})
	t.Run("negative source thresholds", func(t *testing.T) {
		m, x, y := newProof(t)
		m.params.SatVarMax, m.params.CallsRecycle = -1, -1
		check(t, m, x, y, 2)
		check(t, m, x, y, 3)
	})
	t.Run("signed thresholds accepted at startup", func(t *testing.T) {
		p := DefaultParams()
		p.SatVarMax, p.CallsRecycle = -1, -1
		if _, err := newManager(aig.NewGraph(), &p); err != nil {
			t.Fatalf("source accepts signed recycling thresholds: %v", err)
		}
	})

}
