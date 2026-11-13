package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestOldGlaAddLayerFinishesWhenNoPpisRemain(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	node := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, node)
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.nFramesMax = 1
	p.fAddLayer = 1
	if got := Gia_ManPerformGlaOld(g, &p, 0); got != 0 {
		t.Fatalf("result=%d", got)
	}
	if cex := g.SequentialCounterexample(); cex == nil || gia.Gia_ManVerifyCex(g, cex, 0) == 0 {
		t.Fatal("missing valid witness")
	}
}
