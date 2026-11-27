package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestCounterexampleFromUnrolling(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.XOR(x, lo))
	aig.Aig_ManSetRegNum(g, 1)
	// intCtrex.c: unroll exactly nFrames frames from the zero state.
	cex := Inter_ManGetCounterExample(g, 2, 0)
	if cex == nil || cex.IFrame != 1 || cex.NRegs != 1 || cex.NPis != 1 {
		t.Fatalf("incorrect counterexample metadata: %+v", cex)
	}
	if util.Abc_InfoHasBit(cex.PData, 0) || !util.Abc_InfoHasBit(cex.PData, 1) {
		t.Fatalf("trace must start in zero state and toggle via first input: %+v", cex)
	}
}
