package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestOldGlaVtaActualStartup(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, n)
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.nFramesStart = 2
	p.nFramesMax = 3
	p.nTimeOut = 7
	got := Gia_ManPerformGlaOld(g, &p, 1)
	if got != 0 {
		t.Fatalf("result=%d", got)
	}
	if cex := g.SequentialCounterexample(); cex == nil || gia.Gia_ManVerifyCex(g, cex, 0) == 0 {
		t.Fatal("invalid VTA witness")
	}
	if p.nFramesStart != 2 || p.nFramesMax != 3 || p.nTimeOut != 7 {
		t.Fatal("VTA did not restore parameters")
	}
}

func TestOldGlaVtaTransfersUnresolvedAbstraction(t *testing.T) {
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
	p.nFramesStart = 1
	p.nFramesMax = 2
	p.nTimeOut = 7
	if got := Gia_ManPerformGlaOld(g, &p, 1); got != -1 {
		t.Fatalf("result=%d", got)
	}
	if gia.Gia_ManObjClasses(g) != nil || gia.Gia_ManGateClasses(g) == nil {
		t.Fatal("VTA abstraction not transferred")
	}
	if p.nFramesStart != 1 || p.nFramesMax != 2 || p.nTimeOut != 7 {
		t.Fatal("VTA did not restore parameters")
	}
}
