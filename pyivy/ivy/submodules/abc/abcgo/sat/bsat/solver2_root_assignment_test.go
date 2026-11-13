package bsat

import "testing"

func TestLabeledSolver2RootAssigned(t *testing.T) {
	p := NewLabeledSolver2()
	p.NewVariable()
	p.NewVariable()
	if p.RootAssigned(0) || p.RootAssigned(1) {
		t.Fatal("fresh variables assigned")
	}
	if !p.AddClause(1, Positive(0)) || !p.RootAssigned(0) || p.RootAssigned(1) {
		t.Fatal("unit assignment missing")
	}
	if p.Solve(Negative(1)) != Satisfiable || p.RootAssigned(1) {
		t.Fatal("temporary assumption retained")
	}
}
