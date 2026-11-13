package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestRecordOriginalWitnessRejectsAbstractionOnlyDifference(t *testing.T) {
	a, b := gia.Gia_ManStart(8), gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(a), gia.Gia_ManAppendCi(b)
	gia.Gia_ManAppendCo(a, x)
	gia.Gia_ManAppendCo(b, y)
	miter := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCi(miter)
	gia.Gia_ManAppendCo(miter, 1)
	if err := miter.SetCombinationalCounterexample([]bool{true}, 0); err != nil {
		t.Fatal(err)
	}
	if got := acecRecordOriginalWitness(a, b, miter); got != -1 || a.CombinationalCounterexample() != nil {
		t.Fatalf("status %d accepted abstraction-only witness", got)
	}
}

func TestRecordOriginalWitnessUsesOriginalFailingOutput(t *testing.T) {
	a, b := gia.Gia_ManStart(8), gia.Gia_ManStart(8)
	x, y := gia.Gia_ManAppendCi(a), gia.Gia_ManAppendCi(b)
	gia.Gia_ManAppendCo(a, x)
	gia.Gia_ManAppendCo(b, y)
	gia.Gia_ManAppendCo(a, x)
	gia.Gia_ManAppendCo(b, y^1)
	miter := gia.Gia_ManStart(6)
	gia.Gia_ManAppendCi(miter)
	gia.Gia_ManAppendCo(miter, 1)
	if err := miter.SetCombinationalCounterexample([]bool{true}, 0); err != nil {
		t.Fatal(err)
	}
	if got := acecRecordOriginalWitness(a, b, miter); got != 0 {
		t.Fatalf("status %d", got)
	}
	if w := a.CombinationalCounterexample(); w == nil || w.IPo != 1 {
		t.Fatalf("wrong output: %+v", w)
	}
}
