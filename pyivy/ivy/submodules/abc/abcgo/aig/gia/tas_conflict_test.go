package gia

import "testing"

func TestTasWatchedClauseMovesAndPropagates(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 100)
	clause := Tas_ManAllocCls(p, 6)
	copy(clause.literals(), []int32{a, b, c})
	Tas_ManWatchClause(p, clause, a)
	Tas_ManWatchClause(p, clause, b)
	Tas_ManAssign(p, Gia_ObjFromLit(g, a^1), 1, nil, nil)
	if conflict := Tas_ManPropagateWatch(p, 1, a^1); conflict != 0 {
		t.Fatal("unexpected conflict while moving watch")
	}
	if p.pWatches[a^1] != 0 || p.pWatches[c^1] != clause.handle {
		t.Fatal("watch was not moved")
	}
	Tas_ManAssign(p, Gia_ObjFromLit(g, b^1), 1, nil, nil)
	if conflict := Tas_ManPropagateWatch(p, 1, b^1); conflict != 0 {
		t.Fatal("unexpected conflict during unit propagation")
	}
	node := Gia_ManObj(g, c>>1)
	if Tas_VarIsAssigned(node) == 0 || Tas_VarValue(node) != 1 || Tas_VarHasReasonCls(p, node) == 0 || Tas_VarReasonCls(p, node).handle != clause.handle {
		t.Fatal("unit literal or reason not recorded")
	}
	Tas_ManCancelUntil(p, 0)
}

func TestTasCreatesWatchedClauseFromConflict(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 10)
	Tas_ManAssign(p, Gia_ObjFromLit(g, a), 0, nil, nil)
	Tas_ManAssign(p, Gia_ObjFromLit(g, b^1), 0, nil, nil)
	Tas_QuePush(&p.pClauses, Gia_ManObj(g, a>>1))
	Tas_QuePush(&p.pClauses, Gia_ManObj(g, b>>1))
	handle := Tas_QueFinish(&p.pClauses)
	clause := Tas_ManCreateCls(p, handle)
	if clause == nil || clause.literals()[0] != a^1 || clause.literals()[1] != b || p.pWatches[a] != clause.handle || p.pWatches[b^1] != clause.handle {
		t.Fatal("learned clause watches or literals incorrect")
	}
}

func TestTasConflictDerivesDecisionReason(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 10)
	decision, implied := Gia_ManObj(g, a>>1), Gia_ManObj(g, b>>1)
	Tas_ManAssign(p, decision, 1, nil, nil)
	Tas_ManAssign(p, implied, 1, decision, nil)
	handle := Tas_ManAnalyze(p, 1, implied, decision, nil)
	if Tas_ManClauseSize(p, handle) != 1 || p.pClauses.pData[handle] != decision {
		t.Fatal("conflict did not resolve to the decision")
	}
	if decision.fPhase != 0 || implied.fPhase != 0 {
		t.Fatal("conflict analysis left temporary marks")
	}
}
