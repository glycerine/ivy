package gia

import "testing"

func TestSweeperConditionsDoNotLeakProofs(t *testing.T) {
	p := Gia_SweeperStart(nil)
	defer Gia_SweeperStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManHashAnd(p, a, b)
	pa, pna, px, pzero := Gia_SweeperProbeCreate(p, a), Gia_SweeperProbeCreate(p, a^1), Gia_SweeperProbeCreate(p, x), Gia_SweeperProbeCreate(p, 0)
	Gia_SweeperCondPush(p, pa)
	if Gia_SweeperCheckEquiv(p, px, pzero) != 1 {
		t.Fatal("a & b should be zero under a=0")
	}
	Gia_SweeperCondPop(p)
	if Gia_SweeperCheckEquiv(p, px, pzero) != 0 {
		t.Fatal("conditional proof leaked after pop")
	}
	cex := Gia_SweeperGetCex(p)
	if cex == nil || Vec_IntEntry(cex, 0) != 1 || Vec_IntEntry(cex, 1) != 1 {
		t.Fatalf("counterexample %v", Vec_IntArray(cex))
	}
	Gia_SweeperCondPush(p, pa)
	Gia_SweeperCondPush(p, pna)
	if Gia_SweeperCondCheckUnsat(p) != 1 || Gia_SweeperCheckEquiv(p, pa, pna) != 1 {
		t.Fatal("contradictory conditions not detected")
	}
	Gia_SweeperCondPop(p)
	Gia_SweeperCondPop(p)
	if Gia_SweeperCondCheckUnsat(p) != 0 {
		t.Fatal("empty condition set should be satisfiable")
	}
}
