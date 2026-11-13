package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec2AndCec3SourceZeroPhasePrecheck(t *testing.T) {
	for _, engine := range []int{2, 3} {
		g := gia.Gia_ManStart(6)
		x := gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, x^1)
		var out *Gia_Man_t
		var status int32
		if engine == 2 {
			var p Cec2_Par_t
			Cec2_SetDefaultParams(&p)
			p.fIsMiter = 1
			status = Cec2_ManPerformSweeping(g, &p, &out)
		} else {
			var p Cec3_Par_t
			Cec3_SetDefaultParams(&p)
			p.fIsMiter = 1
			status = Cec3_ManPerformSweeping(g, &p, &out)
		}
		cex := g.SequentialCounterexample()
		if status != 0 || out != nil || cex == nil || cex.IPo != 0 || cex.PData[0] != 0 {
			t.Fatalf("engine %d status %d graph %v witness %+v", engine, status, out, cex)
		}
	}
}
