package pdr

import (
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestPdrTimeLimitChoosesEarliestActiveDeadline(t *testing.T) {
	for _, tc := range []struct{ global, output, want int64 }{
		{0, 0, 0}, {10, 0, 10}, {0, 20, 20}, {10, 20, 10}, {20, 10, 10}, {10, 10, 10}, {-1, 10, -1},
	} {
		p := &Pdr_Man_t{timeToStop: tc.global, timeToStopOne: tc.output}
		if got := Pdr_ManTimeLimit(p); got != tc.want {
			t.Fatalf("global=%d output=%d: got %d, want %d", tc.global, tc.output, got, tc.want)
		}
	}
}

// A two-register delay reaches the bad state in frame two. Stopping at
// frame one must not be confused with a proof or with a counterexample.
func TestPdrFrameLimitReportsUnknown(t *testing.T) {
	for _, solveAll := range []bool{false, true} {
		g := aig.NewGraph()
		first, second := g.NewInput(), g.NewInput()
		g.AddOutput(second)
		g.AddOutput(g.Const1())
		g.AddOutput(first)
		aig.Aig_ManSetRegNum(g, 2)
		opts := DefaultOptions()
		opts.Silent, opts.SolveAll, opts.StoreCounterexamples = true, solveAll, true
		opts.MaxFrames = 1
		result, err := Solve(g, &opts)
		if err != nil || result.Status != Unknown || result.Outputs[0].Status != Unknown || result.Outputs[0].Witness != nil {
			t.Fatalf("solveAll=%v limited result %+v err %v", solveAll, result, err)
		}
		opts.MaxFrames = 5
		result, err = Solve(g, &opts)
		if err != nil || result.Status != Failed || result.Outputs[0].Witness == nil ||
			result.Outputs[0].Witness.IFrame != 2 || saig.Saig_ManVerifyCex(g, result.Outputs[0].Witness) != 1 {
			t.Fatalf("solveAll=%v unrestricted result %+v err %v", solveAll, result, err)
		}
	}
}

func TestPdrSatConflictLimitDoesNotProve(t *testing.T) {
	solver := bsat.NewSolver()
	// Five pigeons cannot occupy four holes without sharing a hole. This
	// needs multiple conflicts but remains inexpensive to solve without a cap.
	for pigeon := int32(0); pigeon < 5; pigeon++ {
		var clause []bsat.Literal
		for hole := int32(0); hole < 4; hole++ {
			clause = append(clause, bsat.Positive(pigeon*4+hole))
		}
		if !solver.AddClause(clause...) {
			t.Fatal("unexpected initial inconsistency")
		}
	}
	for hole := int32(0); hole < 4; hole++ {
		for a := int32(0); a < 5; a++ {
			for b := a + 1; b < 5; b++ {
				if !solver.AddClause(bsat.Negative(a*4+hole), bsat.Negative(b*4+hole)) {
					t.Fatal("unexpected initial inconsistency")
				}
			}
		}
	}
	p := &Pdr_Man_t{pPars: &Pdr_Par_t{}}
	if got := pdrSolve(p, solver, nil, 1); got != bsat.Unknown {
		t.Fatalf("one-conflict query = %v, want unknown", got)
	}
	if got := pdrSolve(p, solver, nil, 0); got != bsat.Unsatisfiable {
		t.Fatalf("unlimited query = %v, want UNSAT", got)
	}
}

func TestPdrSatDeadlineAndLimitRestoration(t *testing.T) {
	for _, mono := range []bool{false, true} {
		p, _ := pdrTestSATManager(t, mono)
		solver := Pdr_ManSolver(p, 0)
		solver.SetLimits(17, 23)
		// An expired per-output budget must take precedence over the later
		// global budget and must not return a spurious SAT model.
		p.timeToStop = util.Abc_Clock() + int64(time.Second)
		p.timeToStopOne = util.Abc_Clock() - 1
		if got := Pdr_ManCheckCube(p, 0, nil, nil, 0, 0, 0); got != -1 {
			t.Fatalf("mono=%v expired output query returned %d", mono, got)
		}
		if conflicts, propagations := solver.Limits(); conflicts != 17 || propagations != 23 {
			t.Fatalf("mono=%v limits after query = %d,%d", mono, conflicts, propagations)
		}
		// The global deadline must also interrupt a query when the per-output
		// budget has not expired.
		p.timeToStop = util.Abc_Clock() - 1
		p.timeToStopOne = util.Abc_Clock() + int64(time.Second)
		if got := Pdr_ManCheckCube(p, 0, nil, nil, 0, 0, 0); got != -1 {
			t.Fatalf("mono=%v expired global query returned %d", mono, got)
		}
		// Once the expired output budget is cleared, the same solver can be
		// used by another output within the remaining global budget.
		p.timeToStop = util.Abc_Clock() + int64(time.Second)
		p.timeToStopOne = 0
		if got := Pdr_ManCheckCube(p, 0, nil, nil, 0, 0, 0); got != 0 {
			t.Fatalf("mono=%v later query returned %d", mono, got)
		}
	}
}

// pdrSat.c restores the value returned by sat_solver_set_runtime_limit,
// which may differ from the PDR manager's global deadline.
func TestPdrSolveRestoresPreviousSolverDeadline(t *testing.T) {
	solver := bsat.NewSolver()
	solver.SetDeadline(time.Now().Add(-time.Second))
	p := &Pdr_Man_t{pPars: &Pdr_Par_t{}}
	if got := pdrSolve(p, solver, nil, 0); got != bsat.Satisfiable {
		t.Fatalf("temporary unlimited query = %v", got)
	}
	if got := solver.Solve(); got != bsat.Unknown {
		t.Fatalf("previous expired deadline was not restored: %v", got)
	}
}

func TestPdrSolveRestoresPreviousClockDomain(t *testing.T) {
	solver := bsat.NewSolver()
	oldClock := func() int64 { return 99 }
	solver.SetRuntimeDeadline(oldClock, 100)
	p := &Pdr_Man_t{pPars: &Pdr_Par_t{}, runtimeClock: func() int64 { return 20 }, timeToStop: 10}
	if got := pdrSolve(p, solver, nil, 0); got != bsat.Unknown {
		t.Fatalf("expired CPU-domain query = %v", got)
	}
	clock, deadline := solver.RuntimeDeadline()
	if clock == nil || clock() != 99 || deadline != 100 {
		t.Fatalf("previous domain was not restored: clock=%v deadline=%d", clock != nil, deadline)
	}
	if got := solver.Solve(); got != bsat.Satisfiable {
		t.Fatalf("restored solver = %v", got)
	}
}

// pdrSat.c treats an inconclusive generalization query as SAT when the
// generalization conflict limit is nonzero, even if the deadline caused it.
func TestPdrGeneralizationExpiredDeadlineMatchesC(t *testing.T) {
	for _, tc := range []struct{ tryConf, genLimit int32 }{{0, 1}, {1, 0}, {1, 1}} {
		p, _ := pdrTestSATManager(t, false)
		p.pPars.nConfGenLimit = tc.genLimit
		p.timeToStop = util.Abc_Clock() - 1
		cube := testPdrSet([]int32{0}, nil)
		got := Pdr_ManCheckCube(p, 0, cube, nil, 0, tc.tryConf, 0)
		want := int32(-1)
		wantSATCalls := int32(0)
		if tc.tryConf != 0 && tc.genLimit != 0 {
			want, wantSATCalls = 0, 1
		}
		if got != want || p.nCalls != 1 || p.nCallsS != wantSATCalls || p.nCallsU != 0 {
			t.Fatalf("tryConf=%d genLimit=%d: result=%d calls=%d sat=%d unsat=%d", tc.tryConf, tc.genLimit, got, p.nCalls, p.nCallsS, p.nCallsU)
		}
	}
}

// pdrSat.c returns immediately on an inconclusive property query, before
// touching the caller's predecessor pointer or SAT result counters.
func TestPdrUnknownPropertyLeavesPredecessorUntouched(t *testing.T) {
	for _, mono := range []bool{false, true} {
		p, _ := pdrTestSATManager(t, mono)
		p.timeToStop = util.Abc_Clock() - 1
		sentinel := testPdrSet([]int32{0}, nil)
		predecessor := sentinel
		status := Pdr_ManCheckCube(p, 0, nil, &predecessor, 0, 0, 0)
		if status != -1 || predecessor != sentinel || p.nCalls != 1 || p.nCallsS != 0 || p.nCallsU != 0 || p.nXsimRuns != 0 {
			t.Fatalf("mono=%v status=%d predecessor preserved=%v calls=(%d,%d,%d) simulations=%d", mono, status, predecessor == sentinel, p.nCalls, p.nCallsS, p.nCallsU, p.nXsimRuns)
		}
	}
}

// C's fTryConf path maps an actual conflict-limit UNKNOWN to SAT, while an
// identical non-generalization query returns UNKNOWN. The fixture needs SAT
// search to refute the guarded pigeonhole formula, so this does not rely on
// a deadline or a trivial root contradiction.
func TestPdrGeneralizationConflictLimitMatchesC(t *testing.T) {
	for _, tryConf := range []int32{0, 1} {
		p, cube, _, _, _ := pdrCtgLateTimeoutFixture(t, []int32{0, 2})
		p.vSolvers[1].SetStopFunc(0, nil)
		p.pPars.nConfGenLimit = 1
		p.timeToStop = 0
		got := Pdr_ManCheckCube(p, 1, cube, nil, 1, tryConf, 1)
		want, satCalls := int32(-1), int32(0)
		if tryConf != 0 {
			want, satCalls = 0, 1
		}
		stats := p.vSolvers[1].Statistics()
		if got != want || p.nCallsS != satCalls || p.nCallsU != 0 || stats.Conflicts <= 1 {
			t.Fatalf("tryConf=%d result=%d satCalls=%d unsatCalls=%d conflicts=%d", tryConf, got, p.nCallsS, p.nCallsU, stats.Conflicts)
		}
	}
}
