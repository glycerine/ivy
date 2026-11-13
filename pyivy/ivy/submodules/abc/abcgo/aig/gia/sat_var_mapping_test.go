package gia

import "testing"

func TestRecordSatVariable(t *testing.T) {
	g := Gia_ManStart(4)
	Gia_ManRecordSatVariable(g, 0, 0)
	g.vVar2Obj = Vec_IntAlloc(0)
	Gia_ManRecordSatVariable(g, 0, 0)
	Gia_ManRecordSatVariable(g, 1, 2)
	if Vec_IntSize(g.vVar2Obj) != 2 || Vec_IntEntry(g.vVar2Obj, 1) != 2 {
		t.Fatal("SAT-to-object mapping was not recorded")
	}
}
