package core

import (
	"context"
	"testing"
	"time"
)

func TestCancelledConstantQueriesAreUndecided(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, opt := range []Options{{Context: ctx}, {Deadline: time.Now().Add(-time.Second)}} {
		m, err := NewManager(newCircuit(0), opt)
		if err != nil {
			t.Fatal(err)
		}
		for _, lit := range []int32{0, 1} {
			got, err := m.CheckLiteral(lit)
			if err != nil || got.Status != Undecided || got.Counterexample != nil {
				t.Fatalf("constant %d: %+v err=%v", lit, got, err)
			}
		}
	}
}
