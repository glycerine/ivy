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

func TestInt2ReachabilityNativeXor(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	state := gia.Gia_ManAppendCi(p)
	x := gia.Gia_ManAppendXorReal(p, a, b)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, x, a^1))
	gia.Gia_ManAppendCo(p, state)
	gia.Gia_ManSetRegNum(p, 1)
	if got := int2Reachability(p, 0, time.Time{}); got != INT2_FALSE_NEG {
		t.Fatalf("native XOR reachability = %d, want counterexample", got)
	}
}

func TestInt2ReachabilityIgnoresForbiddenInputs(t *testing.T) {
	p := gia.Gia_ManStart(16)
	x := gia.Gia_ManAppendCi(p)
	state := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManAppendCo(p, x) // constraint violation when x is true
	gia.Gia_ManAppendCo(p, state)
	gia.Gia_ManSetRegNum(p, 1)
	gia.Gia_ManSetConstrNum(p, 1)
	if got := int2Reachability(p, 0, time.Time{}); got != INT2_COMPUTED {
		t.Fatalf("constrained reachability = %d, want proof", got)
	}
}

func TestInt2EvalFrameNativeMuxAndBuffer(t *testing.T) {
	p := gia.Gia_ManStart(32)
	p.EnableMuxes()
	s, a, b := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	state := gia.Gia_ManAppendCi(p)
	mux := gia.Gia_ManAppendMuxReal(p, s, a^1, b)
	buf := gia.Gia_ManAppendBuf(p, mux)
	gia.Gia_ManAppendCo(p, buf)
	gia.Gia_ManAppendCo(p, state^1)
	gia.Gia_ManSetRegNum(p, 1)
	for mask := 0; mask < 16; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		st := []bool{mask&8 != 0}
		want, err := p.Evaluate(append(append([]bool(nil), inputs...), st...))
		if err != nil {
			t.Fatal(err)
		}
		next := make([]bool, 1)
		bad, legal := int2EvalFrame(p, inputs, st, make([]bool, gia.Gia_ManObjNum(p)), next)
		if !legal || bad != want[0] || (!bad && next[0] != want[1]) {
			t.Fatalf("mask %d: bad=%v legal=%v next=%v, want %v", mask, bad, legal, next, want)
		}
	}
}
