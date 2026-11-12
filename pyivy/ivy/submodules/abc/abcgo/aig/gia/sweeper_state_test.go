package gia

import (
	"reflect"
	"testing"
)

func TestSweeperProbeLifecycle(t *testing.T) {
	p := Gia_SweeperStart(nil)
	if Gia_SweeperIsRunning(p) != 1 {
		t.Fatal("not running")
	}
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_SweeperProbeCreate(p, a), Gia_SweeperProbeCreate(p, b)
	if Gia_SweeperProbeUpdate(p, x, a^1) != a || Gia_SweeperProbeDelete(p, y) != b {
		t.Fatal("probe update/delete")
	}
	if got := Vec_IntArray(Gia_SweeperCollectValidProbeIds(p)); !reflect.DeepEqual(got, []int32{x}) {
		t.Fatalf("valid probes %v", got)
	}
	Gia_SweeperCondPush(p, x)
	if Gia_SweeperCondPop(p) != x {
		t.Fatal("condition stack")
	}
	Gia_SweeperSetConflictLimit(p, 42)
	Gia_SweeperSetRuntimeLimit(p, 9)
	if s := sweeper(p); s.nConfMax != 42 || s.nTimeOut != 9 || Swp_ManObj2Lit(s, 0) != 2 {
		t.Fatal("sweeper state")
	}
	Gia_SweeperStop(p)
	if Gia_SweeperIsRunning(p) != 0 {
		t.Fatal("still running")
	}
}
