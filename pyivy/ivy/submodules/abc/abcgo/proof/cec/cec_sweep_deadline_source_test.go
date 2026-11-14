package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
	"time"
)

// C checks TimeLimit only after the SAT query and completed reduction.
func TestSweepExpiredClockStillRunsFirstIteration(t *testing.T) {
	g, a, b := sweepFixture()
	var pars Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&pars)
	pars.NWords = 1
	pars.NRounds = 1
	pars.NItersMax = 1
	pars.fUseOrigIds = 1
	pars.TimeLimit = 1
	base := time.Unix(100, 0)
	calls := 0
	clock := func() time.Time {
		calls++
		if calls == 1 {
			return base
		}
		return base.Add(2 * time.Second)
	}
	out, err := cecOriginalSweepMappedClock(g, g, nil, nil, &pars, clock)
	if err != nil || out == nil {
		t.Fatalf("out=%v err=%v", out, err)
	}
	pairs := vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))
	found := false
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i] == a && pairs[i+1] == b || pairs[i] == b && pairs[i+1] == a {
			found = true
		}
	}
	if !found || calls != 2 {
		t.Fatalf("first iteration absent: pairs=%v clock calls=%d", pairs, calls)
	}
}
