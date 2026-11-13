package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestConstraintUnrollingStartsZeroAndTransfersLatch(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	frames := Ssw_FramesWithConstraints(g, 2)
	for _, x := range []bool{false, true} {
		outs, err := frames.Evaluate([]bool{x, false})
		if err != nil || len(outs) != 2 || !outs[0] || outs[1] == x {
			t.Fatalf("input %v outputs %v error %v", x, outs, err)
		}
	}
}
