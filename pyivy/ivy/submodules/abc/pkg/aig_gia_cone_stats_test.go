package abc

import "testing"

func TestGiaConeStatisticsBoundariesAndLimits(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	gate, _ := g.AppendAnd(a, b)
	co, err := g.AppendCO(gate)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := g.ConeStatistics([]GiaID{gate.ID(), co.ID(), a.ID(), 0}, 0)
	if err != nil || stats.Inputs != 2 || stats.Gates != 1 {
		t.Fatal(stats, err)
	}
	for _, roots := range [][]GiaID{nil, {0}} {
		stats, err := g.ConeStatistics(roots, 0)
		if err != nil || stats != (GiaConeStatistics{}) {
			t.Fatal(stats, err)
		}
	}
	if _, err := g.ConeStatistics([]GiaID{GiaID(g.ObjectCount())}, 0); err == nil {
		t.Fatal("invalid root")
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.ConeStatistics([]GiaID{gate.ID()}, cap); err == nil {
			t.Fatal("invalid cap")
		}
	}
}
