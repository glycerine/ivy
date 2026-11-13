package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
	"time"
)

// ABC's correspondence driver does not consult TimeLimit; this documents the
// source limitation rather than claiming the option cancels correspondence.
func TestCorrespondenceRetainsSourceTimeLimitBehavior(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(pi, lo))
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := new(Ssw_Pars_t)
	Ssw_ManSetDefaultParams(p)
	p.TimeLimit = 1
	p.nStepsMax = 0
	calls := 0
	p.pFunc = func(any) {
		calls++
		if calls == 1 {
			time.Sleep(1100 * time.Millisecond)
		}
	}
	if got := Ssw_SignalCorrespondence(g, p); got == nil {
		t.Fatal("ABC ignores TimeLimit in correspondence")
	}
	if calls != 2 {
		t.Fatalf("source invokes callback twice after BMC, got %d", calls)
	}
}
