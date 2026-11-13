package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
	"testing"
)

func TestCareBitsTraceAcrossRegister(t *testing.T) {
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
	if care.PData[0] != 13 {
		t.Fatalf("care bits %04b", care.PData[0])
	}
}
func TestCareBitsAllVersusOne(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b)^1)
	states := util.Abc_CexAlloc(0, 2, 1)
	all := Bmc_CexCareBits(g, states, nil, nil, 1, 0)
	one := Bmc_CexCareBits(g, states, nil, nil, 0, 0)
	if all.PData[0] != 3 || bits.OnesCount32(one.PData[0]) != 1 {
		t.Fatalf("all=%02b one=%02b", all.PData[0], one.PData[0])
	}
}
