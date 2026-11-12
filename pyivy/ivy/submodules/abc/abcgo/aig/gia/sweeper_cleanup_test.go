package gia

import "testing"

func TestSweeperCleanupPreservesProbesAndConditions(t *testing.T) {
	p := Gia_SweeperStart(nil)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManHashAnd(p, a, b)
	Gia_ManHashAnd(p, a, b^1)
	px := Gia_SweeperProbeCreate(p, x)
	dead := Gia_SweeperProbeCreate(p, b)
	Gia_SweeperProbeDelete(p, dead)
	pa := Gia_SweeperProbeCreate(p, a)
	Gia_SweeperCondPush(p, pa)
	zero := Gia_SweeperProbeCreate(p, 0)
	if Gia_SweeperCheckEquiv(p, px, zero) != 1 {
		t.Fatal("expected conditional proof")
	}
	q := Gia_SweeperCleanup(p, "")
	defer Gia_SweeperStop(q)
	if Gia_SweeperIsRunning(q) != 1 || Gia_ManAndNum(q) != 1 || Gia_SweeperCondPop(q) != pa {
		t.Fatal("cleanup lost state or retained dead logic")
	}
	if got := Vec_IntArray(Gia_SweeperCollectValidProbeIds(q)); len(got) != 3 || got[0] != px || got[1] != pa || got[2] != zero {
		t.Fatalf("probe IDs %v", got)
	}
	if Gia_SweeperCheckEquiv(q, px, zero) != 0 {
		t.Fatal("cleanup retained conditional proof")
	}
	cex := Gia_SweeperGetCex(q)
	if Vec_IntEntry(cex, 0) != 1 || Vec_IntEntry(cex, 1) != 1 {
		t.Fatal("wrong cleanup counterexample")
	}
}
