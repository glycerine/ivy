package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

func TestLegacyAbstractionDriverPreservesRealCounterexample(t *testing.T) {
	g := aig.NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.Const1())
	aig.Aig_ManSetRegNum(g, 1)
	params := Gia_ParAbs_t{nFramesBmc: 3, nRatio: 0}
	_ = Saig_ManCexAbstractionFlops(g, &params)
	cex := g.SequentialCounterexample()
	if cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
		t.Fatal("missing or invalid remapped counterexample")
	}
}
