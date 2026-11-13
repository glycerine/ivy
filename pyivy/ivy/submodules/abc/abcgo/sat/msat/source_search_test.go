package msat

import (
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
