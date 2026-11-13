package gia

import "testing"

func TestAbstractionClassAccessors(t *testing.T) {
	p := Gia_ManStart(10)
	if Gia_ManGateClasses(p) != nil || Gia_ManObjClasses(p) != nil {
		t.Fatal("unexpected initial maps")
	}
	p.SetGateClass(0, 1)
	if Vec_IntEntry(Gia_ManGateClasses(p), 0) != 1 {
		t.Fatal("gate map mismatch")
	}
	obj := Vec_IntStart(2)
	Gia_ManSetObjClasses(p, obj)
	if Gia_ManObjClasses(p) != obj {
		t.Fatal("object map ownership mismatch")
	}
}
