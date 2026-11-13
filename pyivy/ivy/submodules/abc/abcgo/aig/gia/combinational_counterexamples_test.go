package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCombinationalCounterexamplesIncludeCOsAndOwnData(t *testing.T) {
	g := Gia_ManStart(8)
	pi, ro := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, pi)
	Gia_ManAppendCo(g, ro)
	Gia_ManAppendCo(g, pi)
	Gia_ManSetRegNum(g, 1)
	witness := util.Abc_CexAlloc(0, 2, 1)
	witness.IPo = 2
	witness.PData[0] = 3
	if err := g.SetCombinationalCounterexamples([]*util.Abc_Cex_t{nil, util.AbcCexSentinel, witness}); err != nil {
		t.Fatal(err)
	}
	witness.PData[0] = 0
	got := g.CombinationalCounterexamples()
	if len(got) != 3 || got[0] != nil || got[1] != util.AbcCexSentinel || got[2].PData[0] != 3 {
		t.Fatalf("unexpected results %+v", got)
	}
	got[2].PData[0] = 0
	if g.CombinationalCounterexamples()[2].PData[0] != 3 {
		t.Fatal("returned witness aliases stored data")
	}
	bad := util.Abc_CexAlloc(0, 2, 1)
	bad.IPo = 1
	if err := g.SetCombinationalCounterexamples([]*util.Abc_Cex_t{nil, nil, bad}); err == nil {
		t.Fatal("accepted invalid output index")
	}
	if g.CombinationalCounterexamples()[2].PData[0] != 3 {
		t.Fatal("invalid replacement changed stored results")
	}
}
