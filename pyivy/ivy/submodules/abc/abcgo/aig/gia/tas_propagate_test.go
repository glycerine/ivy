package gia

import "testing"

func TestTasAndTruePropagatesComplementedFanins(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	and := Gia_ManAppendAnd(g, a^1, b)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 100)
	Tas_ManAssign(p, Gia_ObjFromLit(g, and), 0, nil, nil)
	if conflict := Tas_ManPropagate(p, 0); conflict != 0 {
		t.Fatalf("unexpected conflict %d", conflict)
	}
	if Tas_VarValue(Gia_ManObj(g, a>>1)) != 0 || Tas_VarValue(Gia_ManObj(g, b>>1)) != 1 || Tas_QueIsEmpty(&p.pJust) == 0 {
		t.Fatal("AND=true did not imply fanin values")
	}
	Tas_ManCancelUntil(p, 0)
}

func TestTasAndFalseResolvesJustification(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	and := Gia_ManAppendAnd(g, a, b)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 100)
	Tas_ManAssign(p, Gia_ObjFromLit(g, and^1), 0, nil, nil)
	if Tas_ManPropagate(p, 0) != 0 || Tas_QueIsEmpty(&p.pJust) != 0 {
		t.Fatal("expected unresolved false AND")
	}
	Tas_ManAssign(p, Gia_ObjFromLit(g, a), 1, nil, nil)
	if conflict := Tas_ManPropagate(p, 1); conflict != 0 {
		t.Fatalf("unexpected conflict %d", conflict)
	}
	node := Gia_ManObj(g, b>>1)
	if Tas_VarIsAssigned(node) == 0 || Tas_VarValue(node) != 0 || Tas_QueIsEmpty(&p.pJust) == 0 {
		t.Fatal("false AND did not imply other input is false")
	}
	Tas_ManCancelUntil(p, 0)
}
