package gia

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestTasManagerAssignmentsReasonsAndModel(t *testing.T) {
	g := Gia_ManStart(10)
	aLit, bLit := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	gateLit := Gia_ManAppendAnd(g, aLit^1, bLit)
	a, b, gate := Gia_ManObj(g, aLit>>1), Gia_ManObj(g, bLit>>1), Gia_ManObj(g, gateLit>>1)
	p := Tas_ManAlloc(g, 10)
	if p.Pars.nBTLimit != 10 || p.Pars.nJustLimit != 2000 || p.pClauses.iHead != 1 {
		t.Fatal("bad SAT defaults")
	}
	if Tas_VarIsJust(gate) != 1 || Tas_VarFanin0Value(gate) != 2 {
		t.Fatal("unassigned fanin state incorrect")
	}
	Tas_VarAssign(a)
	Tas_VarSetValue(a, 1)
	Tas_VarAssign(b)
	Tas_VarSetValue(b, 0)
	if Tas_VarFanin0Value(gate) != 0 || Tas_VarFanin1Value(gate) != 0 || Tas_VarToLit(p, a) != aLit || Tas_LitIsTrue(a, aLit^1) != 0 {
		t.Fatal("assignment or literal value incorrect")
	}
	Tas_QuePush(&p.pProp, a)
	Tas_QuePush(&p.pProp, b)
	Tas_ManSaveModel(p, p.vModel)
	if got := vec.Vec_IntArray(p.vModel); !reflect.DeepEqual(got, []int32{0, 3}) {
		t.Fatalf("model = %v", got)
	}
	clause := p.pStore.allocate(2)
	a.Value = 0
	vec.Vec_IntPush(p.vLevReas, 4)
	vec.Vec_IntPush(p.vLevReas, 0)
	vec.Vec_IntPush(p.vLevReas, clause.handle)
	if Tas_VarDecLevel(p, a) != 4 || Tas_VarHasReasonCls(p, a) != 1 || Tas_ClsHandle(p, Tas_VarReasonCls(p, a)) != clause.handle {
		t.Fatal("clause reason mismatch")
	}
	Tas_VarUnassign(a)
	Tas_VarUnassign(b)
	if a.Value != ^uint32(0) || Tas_VarIsAssigned(a) != 0 {
		t.Fatal("unassignment failed")
	}
	Tas_ManStop(p)
}
