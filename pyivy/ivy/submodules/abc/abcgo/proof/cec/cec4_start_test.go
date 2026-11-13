package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4StartNewCopiesInputMappingAndRegisters(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManSetRegNum(g, 1)
	out := Cec4_ManStartNew(g)
	if gia.Gia_ManCiNum(out) != 2 || gia.Gia_ManRegNum(out) != 1 {
		t.Fatal("input or register count lost")
	}
	for i := int32(0); i < 2; i++ {
		if got := gia.Gia_ObjValue(gia.Gia_ManCi(g, i)); got != gia.Gia_ObjId(out, gia.Gia_ManCi(out, i))<<1 {
			t.Fatalf("input %d copy %d", i, got)
		}
	}
}
