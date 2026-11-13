package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestEssentialBitsAcrossFrames(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, pi, lo))
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	cex := util.Abc_CexAlloc(1, 1, 2)
	cex.IFrame = 1
	cex.PData[0] = 6
	var implied *util.Abc_Cex_t
	states := Bmc_CexInnerStates(g, cex, &implied, 0)
	care := Bmc_CexCareBits(g, states, implied, nil, 1, 0)
	essential := Bmc_CexEssentialBits(g, states, care, 0)
	if essential.PData[0] != 13 {
		t.Fatalf("essential bits %04b", essential.PData[0])
	}
}
func TestEssentialBitsWithAlternativeReason(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a^1, b^1)^1)
	states := util.Abc_CexAlloc(0, 2, 1)
	states.PData[0] = 3
	care := util.Abc_CexAlloc(0, 2, 1)
	care.PData[0] = 3
	essential := Bmc_CexEssentialBits(g, states, care, 0)
	if essential.PData[0] != 0 {
		t.Fatalf("essential bits %b", essential.PData[0])
	}
}
