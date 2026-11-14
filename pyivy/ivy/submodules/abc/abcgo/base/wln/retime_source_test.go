package wln

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestRetimeDefaultDelayClassesAndOutputDrivers(t *testing.T) {
	p := Wln_NtkAlloc("", 12)
	a := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 0, 0)
	b := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 0, 0)
	node := Wln_ObjAlloc(p, ABC_OPER_BIT_AND, 0, 0, 0)
	Wln_ObjAddFanin(p, node, a)
	Wln_ObjAddFanin(p, node, b)
	slice := Wln_ObjAlloc(p, ABC_OPER_SLICE, 0, 0, 0)
	Wln_ObjAddFanin(p, slice, node)
	Wln_ObjAddFanin(p, slice, 0)
	concat := Wln_ObjAlloc(p, ABC_OPER_CONCAT, 0, 0, 0)
	Wln_ObjAddFanin(p, concat, node)
	Wln_ObjAddFanin(p, concat, slice)
	lut := Wln_ObjAlloc(p, ABC_OPER_LUT, 0, 0, 0)
	Wln_ObjAddFanin(p, lut, node)
	co0 := Wln_ObjAlloc(p, ABC_OPER_CO, 0, 0, 0)
	Wln_ObjAddFanin(p, co0, node)
	co1 := Wln_ObjAlloc(p, ABC_OPER_CO, 0, 0, 0)
	Wln_ObjAddFanin(p, co1, lut)

	Wln_NtkRetimeCreateDelayInfo(p)
	for obj, want := range map[int32]int32{node: 1, slice: 1, concat: 1, lut: 10} {
		if got := Wln_ObjInstId(p, obj); got != want {
			t.Errorf("object %d delay %d, want %d", obj, got, want)
		}
	}
	if got := Wln_ObjInstId(p, node); got != 1 || Wln_ObjInstId(p, lut) != 10 {
		t.Fatal("CO drivers should cost one except LUT drivers")
	}
	if got := vec.Vec_IntCountPositive(&p.vInstIds); got != 4 {
		t.Fatalf("assigned delay count %d, want 4", got)
	}
}

func TestRetimePreservesExplicitDelays(t *testing.T) {
	p := Wln_NtkAlloc("", 4)
	ci := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 0, 0)
	node := Wln_ObjAlloc(p, ABC_OPER_BIT_INV, 0, 0, 0)
	Wln_ObjAddFanin(p, node, ci)
	Wln_NtkCleanInstId(p)
	Wln_ObjSetInstId(p, node, 7)
	Wln_NtkRetimeCreateDelayInfo(p)
	if got := Wln_ObjInstId(p, node); got != 7 {
		t.Fatalf("explicit delay changed to %d", got)
	}
}
