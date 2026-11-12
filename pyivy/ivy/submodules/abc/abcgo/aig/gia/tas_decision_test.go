package gia

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestTasDecisionHeuristics(t *testing.T) {
	g := Gia_ManStart(10)
	a := Gia_ManAppendCi(g)
	b := Gia_ManAppendCi(g)
	c := Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, a, c)
	Gia_ManCreateRefs(g)
	p := Tas_ManAlloc(g, 100)
	first, second := Gia_ManObj(g, x>>1), Gia_ManObj(g, y>>1)
	Tas_QuePush(&p.pJust, first)
	Tas_QuePush(&p.pJust, second)
	p.pActivity[a>>1], p.pActivity[b>>1], p.pActivity[c>>1] = 1, 2, 2
	p.pActivity[first.id], p.pActivity[second.id] = 3, 1
	if Tas_ManFindActive(p).id != c>>1 || Tas_ManDecideHighestFanin(p).id != c>>1 || Tas_ManDecideHighest(p) != second || Tas_ManDecideLowest(p) != first || Tas_ManDecideHighestA(p) != first || Tas_ManDecideMaxFF(p) != first {
		t.Fatal("incorrect SAT decision heuristic")
	}
}

func TestTasAssignmentAndBacktrack(t *testing.T) {
	g := Gia_ManStart(10)
	aLit, bLit := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	gateLit := Gia_ManAppendAnd(g, aLit, bLit)
	Gia_ManFillValue(g)
	a, b, gate := Gia_ManObj(g, aLit>>1), Gia_ManObj(g, bLit>>1), Gia_ManObj(g, gateLit>>1)
	p := Tas_ManAlloc(g, 100)
	Tas_ManAssign(p, a, 0, nil, nil)
	Tas_ManAssign(p, Gia_Not(b), 1, nil, nil)
	Tas_ManAssign(p, gate, 1, a, b)
	if Tas_VarReason0(p, gate) != a || Tas_VarReason1(p, gate) != b || Tas_VarDecLevel(p, gate) != 1 {
		t.Fatal("assignment reason mismatch")
	}
	Tas_ManCancelUntil(p, 1)
	if Tas_VarIsAssigned(a) != 1 || Tas_VarIsAssigned(b) != 0 || Tas_VarIsAssigned(gate) != 0 || vec.Vec_IntSize(p.vLevReas) != 3 {
		t.Fatal("backtracking did not restore the trail")
	}
	clause := p.pStore.allocate(2)
	Tas_ManAssign(p, Gia_Not(b), 2, nil, clause)
	if Tas_VarHasReasonCls(p, b) == 0 || Tas_VarReasonCls(p, b).handle != clause.handle {
		t.Fatal("clause reason mismatch")
	}
	Tas_ManCancelUntil(p, 0)
}
