package int2

import (
	"testing"

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

func TestPerformInterpolationReportsUnsupported(t *testing.T) {
	for _, graph := range []*Gia_Man_t{testInt2ToggleMachine(), testInt2InputToBadMachine(), testInt2ThreeStateMachine()} {
		if got := Int2_ManPerformInterpolation(graph, &Int2_ManPars_t{}); got != INT2_TIME_OUT {
			t.Fatalf("incomplete interpolation driver returned %d, want undecided", got)
		}
	}
	if got := Int2_ManPerformInterpolation(nil, nil); got != INT2_TIME_OUT {
		t.Fatalf("nil inputs returned %d, want undecided", got)
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
