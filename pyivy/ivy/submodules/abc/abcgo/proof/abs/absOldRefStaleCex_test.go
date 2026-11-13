package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestLegacyAbstractionClearsStaleCounterexample(t *testing.T) {
	g := aig.NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	stale := util.Abc_CexAlloc(1, 0, 1)
	util.Abc_InfoSetBit(stale.PData, 0)
	if err := g.SetSequentialCounterexample(stale); err != nil {
		t.Fatal(err)
	}
	params := Gia_ParAbs_t{nFramesBmc: 2}
	_ = Saig_ManCexAbstractionFlops(g, &params)
	if g.SequentialCounterexample() != nil {
		t.Fatal("retained stale counterexample")
	}
}
