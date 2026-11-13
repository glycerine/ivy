package gia

import "testing"

func TestResetDoms(t *testing.T) {
	p := Gia_ManStart(4)
	Gia_ManAppendCi(p)
	Gia_ManResetDoms(p)
	for i := int32(0); i < Gia_ManObjNum(p); i++ {
		if got := Gia_ObjDom(p, Gia_ManObj(p, i)); got != -1 {
			t.Fatalf("dom %d = %d", i, got)
		}
	}
	Gia_ObjSetDom(p, Gia_ManObj(p, 1), 0)
	Gia_ManResetDoms(p)
	if Gia_ObjDom(p, Gia_ManObj(p, 1)) != -1 {
		t.Fatal("reset retained previous dominator")
	}
}
