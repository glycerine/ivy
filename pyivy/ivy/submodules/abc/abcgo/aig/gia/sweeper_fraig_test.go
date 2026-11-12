package gia

import "testing"

func TestSweeperFraigUpdatesProbe(t *testing.T) {
	p := Gia_SweeperStart(nil)
	defer Gia_SweeperStop(p)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManHashAnd(p, a, b)
	probe, cond := Gia_SweeperProbeCreate(p, x), Gia_SweeperProbeCreate(p, a)
	ids := Vec_IntAlloc(1)
	Vec_IntPush(ids, probe)
	Gia_SweeperCondPush(p, cond)
	if Gia_SweeperFraig(p, ids, "", 1, 1000, 1, 0) != 1 || Gia_SweeperProbeLit(p, probe) != 0 {
		t.Fatal("probe was not reduced under condition")
	}
}
