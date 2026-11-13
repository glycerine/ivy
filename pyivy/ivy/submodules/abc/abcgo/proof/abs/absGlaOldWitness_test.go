package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

func TestOldGlaClearsStaleWitnessOnUnresolvedRun(t *testing.T) {
	g := gia.Gia_ManStart(12)
	for i := 0; i < 3; i++ {
		gia.Gia_ManAppendCi(g)
	}
	q := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, q)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 1)
	old := util.Abc_CexAlloc(1, 3, 1)
	if err := g.SetSequentialCounterexample(old); err != nil {
		t.Fatal(err)
	}
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.nFramesMax = 1
	if got := Gia_ManPerformGlaOld(g, &p, 0); got != -1 {
		t.Fatalf("result=%d", got)
	}
	if g.SequentialCounterexample() != nil {
		t.Fatal("stale witness survived unresolved run")
	}
}
