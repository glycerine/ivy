package int2

import (
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
)

func testInt2ToggleMachine() *Gia_Man_t {
	p := gia.Gia_ManStart(16)
	state := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, 0) // always-safe output
	gia.Gia_ManAppendCo(p, state^1)
	gia.Gia_ManSetRegNum(p, 1)
	return p
}

func testInt2ThreeStateMachine() *Gia_Man_t {
	p := gia.Gia_ManStart(32)
	bit0, bit1 := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	next0 := gia.Gia_ManAppendAnd(p, bit0^1, bit1^1)
	next1 := gia.Gia_ManAppendAnd(p, bit0, bit1^1)
	gia.Gia_ManAppendCo(p, 0) // always-safe output
	gia.Gia_ManAppendCo(p, next0)
	gia.Gia_ManAppendCo(p, next1)
	gia.Gia_ManSetRegNum(p, 2)
	return p
}

func testInt2InputToBadMachine() *Gia_Man_t {
	p := gia.Gia_ManStart(16)
	input := gia.Gia_ManAppendCi(p)
	state := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, state) // state one is a counterexample
	gia.Gia_ManAppendCo(p, input)
	gia.Gia_ManSetRegNum(p, 1)
	return p
}

func TestInt2ReachabilityProvesByExhaustiveClosure(t *testing.T) {
	p := testInt2ToggleMachine()
	defer gia.Gia_ManStop(p)
	if got := int2Reachability(p, 0, time.Time{}); got != INT2_COMPUTED {
		t.Fatalf("reachability status = %d, want proven", got)
	}
}

func TestInt2PerformInterpolationRunsPastInitialFrame(t *testing.T) {
	p := testInt2ToggleMachine()
	defer gia.Gia_ManStop(p)
	// Both reachable states are safe; the driver must close the reachable set
	// rather than returning after only checking the all-zero initial state.
	if got := Int2_ManPerformInterpolation(p, &Int2_ManPars_t{}); got != INT2_COMPUTED {
		t.Fatalf("verification status = %d, want proven", got)
	}
}

func TestInt2ReachabilityFindsLaterCounterexample(t *testing.T) {
	p := testInt2InputToBadMachine()
	defer gia.Gia_ManStop(p)
	if got := int2Reachability(p, 0, time.Time{}); got != INT2_FALSE_NEG {
		t.Fatalf("reachability status = %d, want counterexample", got)
	}
}

func TestInt2ReachabilityDepthLimitIsUndecided(t *testing.T) {
	p := testInt2ThreeStateMachine()
	defer gia.Gia_ManStop(p)
	if got := int2Reachability(p, 1, time.Time{}); got != INT2_TIME_OUT {
		t.Fatalf("bounded status = %d, want undecided", got)
	}
	if got := int2Reachability(p, 0, time.Time{}); got != INT2_COMPUTED {
		t.Fatalf("complete status = %d, want proven", got)
	}
}
