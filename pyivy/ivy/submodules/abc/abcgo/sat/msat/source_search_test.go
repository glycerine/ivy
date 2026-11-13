package msat

import (
	"math"
	"testing"
	"time"
)

func TestSearchDoesNotCheckOuterRuntimeDeadline(t *testing.T) {
	p := New(1)
	p.deadline = time.Now().Add(-time.Second)
	params := &Msat_SearchParams_t{0.95, 0.999}
	if got := Msat_SolverSearch(p, 100, 10, 0, params); got != MSAT_TRUE {
		t.Fatalf("C checks elapsed time between searches, search returned %d", got)
	}
}

func TestSearchRetainsExistingDecisionOrder(t *testing.T) {
	p := New(2)
	Msat_OrderClean(p.pOrder, &Msat_IntVec_t{})
	params := &Msat_SearchParams_t{0.95, 0.999}
	if got := Msat_SolverSearch(p, 100, 10, 0, params); got != MSAT_TRUE {
		t.Fatalf("result = %d", got)
	}
	if p.Stats.nDecisions != 0 {
		t.Fatalf("C search does not repopulate empty order, decisions = %d", p.Stats.nDecisions)
	}
}

func TestSearchRequiresCurrentLevelToMatchRoot(t *testing.T) {
	p := New(1)
	p.nLevelRoot = 1
	defer func() {
		if recover() == nil {
			t.Fatal("C search asserts decision level equals root")
		}
	}()
	Msat_SolverSearch(p, 100, 10, 0, &Msat_SearchParams_t{0.95, 0.999})
}

func TestSearchUsesSuppliedDecayValuesEvenWhenZero(t *testing.T) {
	p := New(1)
	Msat_SolverSearch(p, 100, 10, 0, &Msat_SearchParams_t{})
	if !math.IsInf(p.dVarDecay, 1) || !math.IsInf(p.dClaDecay, 1) {
		t.Fatalf("C reciprocal decay values: %v %v", p.dVarDecay, p.dClaDecay)
	}
}
