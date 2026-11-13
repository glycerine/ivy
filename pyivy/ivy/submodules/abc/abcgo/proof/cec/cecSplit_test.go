package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func splitTestGraph(output int32) *gia.Gia_Man_t {
	p := gia.Gia_ManStart(16)
	x, y := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	and := gia.Gia_ManAppendAnd(p, x, y)
	if output == 0 {
		and = 0
	}
	gia.Gia_ManAppendCo(p, and)
	return p
}

func TestSplitSolveAndMultiOutput(t *testing.T) {
	for _, tc := range []struct{ output, want int32 }{{0, 1}, {1, 0}} {
		p := splitTestGraph(tc.output)
		got := Cec_GiaSplitTest(p, 3, 0, 0, 2, 0, 0, 1)
		if got != tc.want {
			t.Fatalf("output %d: got %d want %d", tc.output, got, tc.want)
		}
		if got == 0 {
			cex := p.CombinationalCounterexample()
			if cex == nil || cex.IPo != 0 || cex.PData[0]&3 != 3 {
				t.Fatalf("bad SAT counterexample: %+v", cex)
			}
		}
	}
	p := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, 0)
	gia.Gia_ManAppendCo(p, x)
	if got := Cec_GiaSplitTest(p, 1, 0, 0, 1, 0, 0, 1); got != 0 {
		t.Fatalf("multi-output status %d", got)
	}
	if cex := p.CombinationalCounterexample(); cex == nil || cex.IPo != 1 || cex.PData[0]&1 == 0 {
		t.Fatalf("bad output witness %+v", cex)
	}
}

func TestSplitBranchesPreserveAssignments(t *testing.T) {
	for _, workers := range []int32{1, 3} {
		p := splitTestGraph(1)
		solve := func(part splitPart, timeout int32) splitResult {
			if len(part.fixes) == 0 {
				return splitResult{status: -1}
			}
			return splitSolve(part, timeout)
		}
		status, model := splitRun(p, workers, 0, 0, 1, 0, 0, 1, solve)
		if status != 0 || len(model) != 2 || !model[0] || !model[1] {
			t.Fatalf("workers %d: status %d model %v", workers, status, model)
		}
	}
}

func TestSplitIterationLimitDoesNotClaimUnsat(t *testing.T) {
	p := splitTestGraph(1)
	alwaysUnknown := func(splitPart, int32) splitResult { return splitResult{status: -1} }
	status, _ := splitRun(p, 2, 0, 1, 1, 0, 0, 1, alwaysUnknown)
	if status != -1 {
		t.Fatalf("expected unknown at iteration limit, got %d", status)
	}
}
