package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldGlaFrameVariables(t *testing.T) {
	p := &Gla_Man_t{nObjs: 3, pObjs: make([]Gla_Obj_t, 3), nSatVars: 1, vAddedNew: vec.Vec_IntAlloc(0)}
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	if Gla_ManObj(p, 0) != nil || Gla_ObjId(p, Gla_ManObj(p, 2)) != 2 {
		t.Fatal("object indexing")
	}
	if Gla_ManCheckVar(p, 1, 4) != 0 {
		t.Fatal("unused variable exists")
	}
	if got := Gla_ManGetVar(p, 1, 4); got != 1 {
		t.Fatalf("first variable=%d", got)
	}
	if got := Gla_ManGetVar(p, 1, 4); got != 1 || Gla_ManCheckVar(p, 1, 4) != 1 {
		t.Fatal("variable not reused")
	}
	if got := Gla_ManGetVar(p, 1, 0); got != 2 {
		t.Fatalf("second variable=%d", got)
	}
	if got := vec.Vec_IntArray(p.vAddedNew); !reflect.DeepEqual(got, []int32{1, 4, 1, 0}) {
		t.Fatalf("new variables=%v", got)
	}
}
func TestOldGlaRefinementFrameStorage(t *testing.T) {
	g := gia.Gia_ManStart(4)
	a := gia.Gia_ManAppendCi(g)
	p := &Gla_Man_t{pGia: g, pvRefis: make([][]Rfn_Obj_t, gia.Gia_ManObjNum(g))}
	obj := gia.Gia_ManObj(g, a>>1)
	Gla_ObjRef(p, obj, 3).Prio = 7
	if Gla_ObjRef(p, obj, 3).Prio != 7 || Gla_ObjRef(p, obj, 2).Prio != 0 {
		t.Fatal("refinement frames not isolated")
	}
	Gla_ObjClearRef(Gla_ObjRef(p, obj, 3))
	if Gla_ObjRef(p, obj, 3).Prio != 0 {
		t.Fatal("refinement clear")
	}
}
