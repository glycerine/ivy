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

func TestLabeledSolver2RootValue(t *testing.T) {
	p := NewLabeledSolver2()
	p.NewVariable()
	p.NewVariable()
	p.NewVariable()
	p.AddClause(1, Positive(0))
	p.AddClause(2, Negative(1))
	if value, ok := p.RootValue(0); !ok || !value {
		t.Fatal("missing true root")
	}
	if value, ok := p.RootValue(1); !ok || value {
		t.Fatal("missing false root")
	}
	if _, ok := p.RootValue(2); ok {
		t.Fatal("unassigned root")
	}
}
