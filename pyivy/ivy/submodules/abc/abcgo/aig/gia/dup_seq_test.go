package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestRegisterFlipTransitionCoordinateChange(t *testing.T) {
	src := Gia_ManStart(8)
	input, state := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	next := Gia_ManAppendXorReal(src, input, state)
	Gia_ManAppendCo(src, state)
	Gia_ManAppendCo(src, next)
	Gia_ManSetRegNum(src, 1)
	flipped := Gia_ManDupFlip(src, []int32{1})
	for mask := 0; mask < 4; mask++ {
		pi, newState := mask&1 != 0, mask&2 != 0
		oldOut, err := src.Evaluate([]bool{pi, !newState})
		if err != nil {
			t.Fatal(err)
		}
		newOut, err := flipped.Evaluate([]bool{pi, newState})
		if err != nil {
			t.Fatal(err)
		}
		if oldOut[0] != newOut[0] || oldOut[1] == newOut[1] {
			t.Fatal("flipped transition", mask, oldOut, newOut)
		}
	}
}
func TestCycleReplayAndCycledDuplicate(t *testing.T) {
	p := Gia_ManStart(8)
	pi, state := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	next := Gia_ManAppendXorReal(p, pi, state)
	Gia_ManAppendCo(p, state)
	Gia_ManAppendCo(p, next)
	Gia_ManSetRegNum(p, 1)
	cex := util.Abc_CexAlloc(1, 1, 4)
	cex.IFrame = 3
	// PIs 1,0,1: zero -> one -> one -> zero.
	cex.PData[0] |= 1<<1 | 1<<3
	Gia_ManCleanMark0(p)
	Gia_ManCycle(p, cex, 2)
	if Gia_ObjMark0(Gia_ManRo(p, 0)) != 1 {
		t.Fatal("counterexample replay")
	}
	cycled := Gia_ManDupCycled(p, cex, 2)
	out, err := cycled.Evaluate([]bool{false, false})
	if err != nil || !out[0] {
		t.Fatal("cycled initial state", out, err)
	}
	Gia_ManCheckMark0(p)
}

func TestPackedRegisterFlipAcrossWordBoundary(t *testing.T) {
	p := Gia_ManStart(80)
	states := make([]int32, 35)
	for i := range states {
		states[i] = Gia_ManAppendCi(p)
	}
	Gia_ManAppendCo(p, states[34])
	for _, state := range states {
		Gia_ManAppendCo(p, state)
	}
	Gia_ManSetRegNum(p, 35)
	flipped := Gia_ManDupFlip(p, []int32{0, 4})
	out, err := flipped.Evaluate(make([]bool, 35))
	if err != nil || !out[0] || out[35] {
		t.Fatal("second packed word", out, err)
	}
}
