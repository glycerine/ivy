package core

import (
	"context"
	"testing"
)

// A live manager may be canceled after previous successful queries, including
// a query that crosses the recycling thresholds.
func TestCancellationSurvivesSolverRecycling(t *testing.T) {
	c := newCircuit(2)
	ctx, cancel := context.WithCancel(context.Background())
	m, err := NewManager(c, Options{Context: ctx, MaxSATVariables: 1, RecycleQueries: 1})
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.CheckLiteral(c.inputs[0])
	if err != nil || got.Status != Different {
		t.Fatalf("initial query %+v err=%v", got, err)
	}
	cancel()
	got, err = m.CheckConjunction(c.inputs[0], c.inputs[1])
	if err != nil || got.Status != Undecided || got.Counterexample != nil {
		t.Fatalf("canceled query %+v err=%v", got, err)
	}
	if m.Statistics().Recycles != 1 {
		t.Fatalf("expected actual solver recycle: %+v", m.Statistics())
	}
	got, err = m.CheckPair(c.inputs[0], c.inputs[0])
	if err != nil || got.Status != Undecided || got.Counterexample != nil {
		t.Fatalf("post-recycle query %+v err=%v", got, err)
	}
}
