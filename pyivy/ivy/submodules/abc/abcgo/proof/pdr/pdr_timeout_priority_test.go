package pdr

import (
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
)

// Both pdrCore.c and pdrIncr.c check global, gap, then output expiry after
// an unknown query; only the last case continues to another output.
func TestOutputTimeoutPriorityMatchesSource(t *testing.T) {
	const now = int64(10 * time.Second)
	for _, tc := range []struct {
		name                       string
		global, output, lastSolved int64
		gap                        int32
		want                       bool
	}{
		{"no output deadline", 0, 0, 0, 0, false},
		{"output only", 0, now - 1, 0, 0, true},
		{"global also expired", now - 1, now - 1, 0, 0, false},
		{"gap also expired", 0, now - 1, now - int64(2*time.Second), 1, false},
		{"gap at boundary", 0, now - 1, now - int64(time.Second), 1, true},
		{"gap not started", 0, now - 1, 0, 1, true},
		{"output at boundary", 0, now, 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &Pdr_Man_t{timeToStop: tc.global, timeToStopOne: tc.output, pPars: &Pdr_Par_t{nTimeOutGap: tc.gap, timeLastSolved: tc.lastSolved}}
			if got := pdrOnlyOutputTimedOut(p, now); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// The clock advances during the SAT callback so both budgets expire together.
// ABC aborts for the gap timeout without dropping the output.
func TestGapTimeoutPrecedesOutputDropInBothDrivers(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		g := aig.NewGraph()
		x, y, q := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(x, y))
		g.AddOutput(g.AND(x.Not(), y))
		g.AddOutput(q)
		aig.Aig_ManSetRegNum(g, 1)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSolveAll, pars.fSilent, pars.fNotVerbose = 1, 1, 1
		pars.nTimeOut, pars.nTimeOutGap, pars.nTimeOutOne = 5, 1, 1
		calls := 0
		now := int64(time.Second)
		pars.pFuncStop = func(int32) int32 {
			calls++
			now += int64(2 * time.Second)
			return 1
		}
		p := Pdr_ManStart(g, &pars, nil)
		p.runtimeClock = func() int64 { return now }
		var result int32
		if incremental {
			result = IPdr_ManSolveInt(p, 0, 0)
		} else {
			result = Pdr_ManSolveInt(p)
		}
		if result != -1 || calls == 0 || pars.nDropOuts != 0 || pars.iFrame != 0 {
			t.Fatalf("incremental=%v result=%d calls=%d dropped=%d frame=%d", incremental, result, calls, pars.nDropOuts, pars.iFrame)
		}
		Pdr_ManStop(p)
	}
}
