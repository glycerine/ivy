package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestCounterexampleInnerStates(t *testing.T) {
	g := gia.Gia_ManStart(16)
	pi, lo := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, lo)
	gia.Gia_ManAppendCo(g, pi)
	gia.Gia_ManSetRegNum(g, 1)
	c := util.Abc_CexAlloc(1, 1, 2)
	c.IFrame = 1
	c.PData[0] = 1 << 1
	var marks *util.Abc_Cex_t
	states := Bmc_CexInnerStates(g, c, &marks, 0)
	if states.NRegs != 0 || states.NPis != 2 || states.PData[0] != 9 {
		t.Fatalf("states %+v bits %b", states, states.PData[0])
	}
	if marks.PData[0] != 2 {
		t.Fatalf("implications %b", marks.PData[0])
	}
}
