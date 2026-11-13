package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestOldGlaHonorsFrameLimit(t *testing.T) {
	g := gia.Gia_ManStart(12)
	for i := 0; i < 3; i++ {
		gia.Gia_ManAppendCi(g)
	}
	q := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, q)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 1)
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.nFramesMax = 2
	got := Gia_ManPerformGlaOld(g, &p, 0)
	if got != -1 || p.iFrame != 2 || g.SequentialCounterexample() != nil {
		t.Fatalf("result=%d frame=%d cex=%v", got, p.iFrame, g.SequentialCounterexample())
	}
}
