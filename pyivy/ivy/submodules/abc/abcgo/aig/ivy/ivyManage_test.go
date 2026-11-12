package ivy

import "testing"

func TestManagerDuplicatePreservesXor(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	x := Ivy_CanonExor(p, a, b)
	Ivy_ObjCreatePo(p, x)
	Ivy_ObjCreatePo(p, Ivy_Not(x))
	copy := Ivy_ManDup(p)
	for bits := 0; bits < 4; bits++ {
		want, got := ivyBalanceEvaluate(p, bits), ivyBalanceEvaluate(copy, bits)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("duplicate changed XOR at %d", bits)
			}
		}
	}
}
func TestSequentialCleanupRemovesUnreachableFeedback(t *testing.T) {
	p := Ivy_ManStart()
	input := Ivy_ObjCreatePi(p)
	Ivy_ObjCreatePo(p, input)
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	p.vPis = p.vPis[:1]
	p.nObjs[IVY_PI] -= 2
	p.nObjs[IVY_LATCH] += 2
	a.Type, b.Type = IVY_LATCH, IVY_LATCH
	a.Init, b.Init = IVY_INIT_0, IVY_INIT_1
	Ivy_ObjConnect(p, a, b, nil)
	Ivy_ObjConnect(p, b, a, nil)
	Ivy_ManStartFanout(p)
	if Ivy_ManCheck(p) == 0 {
		t.Fatal("valid latch feedback graph rejected")
	}
	if Ivy_ManLatchIsSelfFeed(a) != 1 {
		t.Fatal("feedback cycle missed")
	}
	duplicate := Ivy_ManDup(p)
	if Ivy_ManLatchNum(duplicate) != 2 || Ivy_ManCheck(duplicate) == 0 {
		t.Fatal("feedback graph duplication failed")
	}
	if removed := Ivy_ManCleanupSeq(p); removed != 2 {
		t.Fatalf("removed %d, want two latches", removed)
	}
	if Ivy_ManCheck(p) == 0 || Ivy_ManLatchNum(p) != 0 {
		t.Fatal("feedback cleanup left invalid graph")
	}
}
func TestSelfFeedTerminatesOnDifferentLatchCycle(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	a.Type, b.Type, c.Type = IVY_LATCH, IVY_LATCH, IVY_LATCH
	a.pFanin0 = b
	b.pFanin0 = b
	c.pFanin0 = a
	if Ivy_ManLatchIsSelfFeed(c) != 0 {
		t.Fatal("unrelated cycle classified as self-feed")
	}
}
