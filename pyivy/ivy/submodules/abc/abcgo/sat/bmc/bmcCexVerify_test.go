package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCexVerifyCareAcrossFrames(t *testing.T) {
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
	if Bmc_CexVerify(g, cex, care) != 1 {
		t.Fatal("valid care set rejected")
	}
	care.PData[0] &^= 1
	if Bmc_CexVerify(g, cex, care) != 0 {
		t.Fatal("missing required input accepted")
	}
}

func TestCexDepthEndToEnd(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, pi, lo))
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	cex := util.Abc_CexAlloc(1, 1, 2)
	cex.IFrame = 1
	cex.PData[0] = 6
	result := Bmc_CexDepthTest(g, cex, 1, 0)
	if result == nil || gia.Gia_ManPoNum(result) != 2 || gia.Gia_ManRegNum(result) != 1 {
		t.Fatal("invalid depth result")
	}
}
