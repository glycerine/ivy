package wln

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestHeaderFaninRefsAndCioMetadata(t *testing.T) {
	p := &Wln_Ntk_t{vFanins: make([]Wln_Vec_t, 3)}
	vec.Vec_IntPush(&p.vTypes, -1)
	vec.Vec_IntPush(&p.vTypes, ABC_OPER_CI)
	vec.Vec_IntPush(&p.vTypes, ABC_OPER_CO)
	vec.Vec_IntPush(&p.vCis, 1)
	vec.Vec_IntPush(&p.vCos, 2)
	vec.Vec_IntFill(&p.vRefs, 3, 0)
	Wln_ObjSetFanin(p, 1, 1, 0)
	Wln_ObjSetFanin(p, 2, 1, 0)
	if Wln_ObjIsPi(p, 1) != 1 || Wln_ObjIsPo(p, 2) != 1 {
		t.Fatal("CI/CO metadata classification")
	}
	if got := Wln_ObjRefsInc(p, 1); got != 0 {
		t.Fatalf("postincrement returned %d", got)
	}
	if got := Wln_ObjRefsDec(p, 1); got != 0 {
		t.Fatalf("predecrement returned %d", got)
	}
	Wln_NtkIncrementTravId(p)
	if Wln_ObjIsTravIdCurrent(p, 1) != 0 {
		t.Fatal("fresh traversal contains object")
	}
	Wln_ObjSetTravIdCurrent(p, 1)
	if Wln_ObjIsTravIdCurrent(p, 1) != 1 {
		t.Fatal("failed to mark traversal")
	}
}
