package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestSpuriousActuallyRefines(t *testing.T) {
	g := gia.Gia_ManStart(20)
	x := make([]int32, 5)
	for i := range x {
		x[i] = gia.Gia_ManAppendCi(g)
	}
	l := gia.Gia_ManAppendAnd(g, x[0], x[1])
	l = gia.Gia_ManAppendAnd(g, l, x[2])
	r := gia.Gia_ManAppendAnd(g, x[0]^1, x[3])
	r = gia.Gia_ManAppendAnd(g, r, x[4])
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, l, r))
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.NFramesMax = 1
	p.nFramesNoChange = 10
	Gia_ManPerformGlaOld(g, &p, 0)
	if p.nFramesNoChange != 0 {
		t.Fatalf("no refinement, no-change=%d", p.nFramesNoChange)
	}
}
