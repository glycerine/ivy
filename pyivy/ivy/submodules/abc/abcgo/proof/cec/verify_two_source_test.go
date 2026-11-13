package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestVerifyTwoUsesCombinationalRegisterInterface(t *testing.T) {
	left, right := gia.Gia_ManStart(16), gia.Gia_ManStart(16)
	la, lr := gia.Gia_ManAppendCi(left), gia.Gia_ManAppendCi(left)
	ra, rr := gia.Gia_ManAppendCi(right), gia.Gia_ManAppendCi(right)
	gia.Gia_ManAppendCo(left, gia.Gia_ManAppendAnd(left, la, lr))
	gia.Gia_ManAppendCo(left, la)
	gia.Gia_ManAppendCo(right, gia.Gia_ManAppendAnd(right, ra, rr))
	gia.Gia_ManAppendCo(right, ra^1)
	gia.Gia_ManSetRegNum(left, 1)
	gia.Gia_ManSetRegNum(right, 1)
	if status := Cec_ManVerifyTwo(left, right, 0); status != 0 {
		t.Fatalf("status %d", status)
	}
	w := left.CombinationalCounterexample()
	if w == nil || w.NRegs != 0 || w.NPis != 2 || w.IPo != 1 {
		t.Fatalf("witness %+v", w)
	}
	inputs := []bool{w.PData[0]&1 != 0, w.PData[0]&2 != 0}
	a, e1 := left.Evaluate(inputs)
	b, e2 := right.Evaluate(inputs)
	if e1 != nil || e2 != nil || a[1] == b[1] {
		t.Fatalf("invalid witness %v: %v %v errors %v %v", inputs, a, b, e1, e2)
	}
	leftAig, rightAig := gia.Gia_ManToAig(left, 0), gia.Gia_ManToAig(right, 0)
	if status := Cec_ManVerifyTwoAigs(leftAig, rightAig, 0); status != 0 {
		t.Fatalf("AIG status %d", status)
	}
	wa := leftAig.CombinationalCounterexample()
	if wa == nil || wa.NRegs != 1 || wa.IPo != 1 {
		t.Fatalf("AIG witness %+v", wa)
	}
	inputs = []bool{wa.PData[0]&2 != 0, wa.PData[0]&1 != 0}
	aa, e1 := leftAig.Evaluate(inputs)
	bb, e2 := rightAig.Evaluate(inputs)
	if e1 != nil || e2 != nil || aa[1] == bb[1] {
		t.Fatalf("invalid AIG witness %v: %v %v errors %v %v", inputs, aa, bb, e1, e2)
	}
}

func TestVerifyTwoInvalidInputsDoNotPanic(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	if got := Cec_ManVerifyTwo(nil, g, 0); got != -1 {
		t.Fatalf("nil status %d", got)
	}
	h := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCo(h, 0)
	if got := Cec_ManVerifyTwo(g, h, 0); got != -1 {
		t.Fatalf("interface status %d", got)
	}
	aig := gia.Gia_ManToAig(g, 0)
	if got := Cec_ManVerifyTwoAigs(nil, aig, 0); got != -1 {
		t.Fatalf("nil AIG status %d", got)
	}
}

func TestVerifyTwoNativeGatesAndAigBridge(t *testing.T) {
	build := func(flip bool) *gia.Gia_Man_t {
		g := gia.Gia_ManStart(16)
		g.EnableMuxes()
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendXorReal(g, a, b)
		m := gia.Gia_ManAppendMuxReal(g, c, x, a)
		o := gia.Gia_ManAppendBuf(g, m)
		if flip {
			o ^= 1
		}
		gia.Gia_ManAppendCo(g, o)
		return g
	}
	for _, flip := range []bool{false, true} {
		left, right := build(false), build(flip)
		want := int32(1)
		if flip {
			want = 0
		}
		if status := Cec_ManVerifyTwo(left, right, 0); status != want {
			t.Fatalf("flip %v GIA status %d", flip, status)
		}
		if flip {
			w := left.CombinationalCounterexample()
			if w == nil {
				t.Fatal("missing native witness")
			}
			inputs := []bool{w.PData[0]&1 != 0, w.PData[0]&2 != 0, w.PData[0]&4 != 0}
			a, e1 := left.Evaluate(inputs)
			b, e2 := right.Evaluate(inputs)
			if e1 != nil || e2 != nil || a[0] == b[0] {
				t.Fatalf("native witness %v outputs %v %v errors %v %v", inputs, a, b, e1, e2)
			}
		}
		la, ra := gia.Gia_ManToAig(left, 0), gia.Gia_ManToAig(right, 0)
		if status := Cec_ManVerifyTwoAigs(la, ra, 0); status != want {
			t.Fatalf("flip %v AIG status %d", flip, status)
		}
		if flip {
			w := la.CombinationalCounterexample()
			if w == nil {
				t.Fatal("missing AIG witness")
			}
			inputs := []bool{w.PData[0]&1 != 0, w.PData[0]&2 != 0, w.PData[0]&4 != 0}
			a, e1 := la.Evaluate(inputs)
			b, e2 := ra.Evaluate(inputs)
			if e1 != nil || e2 != nil || a[0] == b[0] {
				t.Fatalf("AIG witness %v outputs %v %v errors %v %v", inputs, a, b, e1, e2)
			}
		}
	}
}

// C transfers the miter witness slot even when no counterexample exists.
func TestVerifyTwoReplacesExistingWitnessOnEquivalentResult(t *testing.T) {
	left, right := gia.Gia_ManStart(8), gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(left), gia.Gia_ManAppendCi(right)
	gia.Gia_ManAppendCo(left, a)
	gia.Gia_ManAppendCo(right, b)
	if err := left.SetCombinationalCounterexample([]bool{true}, 0); err != nil {
		t.Fatal(err)
	}
	if got := Cec_ManVerifyTwo(left, right, 0); got != 1 {
		t.Fatalf("status %d", got)
	}
	if left.CombinationalCounterexample() != nil {
		t.Fatal("source replaces old witness with nil")
	}
}

func TestSpecialCasesBufferNodeRequiresFurtherVerification(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendBuf(g, a)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	var p Cec_ParCec_t
	Cec_ManCecSetDefaultParams(&p)
	p.fSilent = 1
	if got := Cec_ManHandleSpecialCases(g, &p); got != -1 {
		t.Fatalf("got %d, want source undecided", got)
	}
}

func TestSpecialCasesEquivalentKeepsCallerState(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, a)
	if err := g.SetCombinationalCounterexample([]bool{true}, 0); err != nil {
		t.Fatal(err)
	}
	before := g.CombinationalCounterexample()
	p := Cec_ParCec_t{iOutFail: 17, fSilent: 1}
	if got := Cec_ManHandleSpecialCases(g, &p); got != 1 || p.iOutFail != 17 || g.CombinationalCounterexample() != before {
		t.Fatalf("status %d output %d witness %p", got, p.iOutFail, g.CombinationalCounterexample())
	}
}
