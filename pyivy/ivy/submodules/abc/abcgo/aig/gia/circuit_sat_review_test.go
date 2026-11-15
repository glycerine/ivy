package gia

import "testing"

func TestCircuitSATRejectsInvalidIdleQueues(t *testing.T) {
	for field := 0; field < 6; field++ {
		g := Gia_ManStart(16)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		root := Gia_ManAppendAnd(g, a, b)
		p := Cbs_ManAlloc(g)
		switch field {
		case 0:
			p.pProp.iHead = 1
		case 1:
			p.pProp.iTail = 1
		case 2:
			p.pJust.iHead = 1
		case 3:
			p.pJust.iTail = 1
		case 4:
			p.pClauses.iHead = 2
		case 5:
			p.pClauses.iTail = 2
		}
		// Keep a valid trail entry when checking an active propagation queue.
		if field == 1 {
			p.pProp.pData = append(p.pProp.pData, g.pObjs[a>>1])
			Cbs_VarAssign(g.pObjs[a>>1])
			Cbs_VarSetValue(g.pObjs[a>>1], 1)
		}
		cexReviewMustPanic(t, func() { Cbs_ManSolve(p, Gia_ObjFromLit(g, root)) })
		Cbs_ManStop(p)
	}
}
func TestCircuitSATRejectsForeignAndOutputRoots(t *testing.T) {
	for pair := 0; pair < 2; pair++ {
		for kind := 0; kind < 3; kind++ {
			if pair == 1 && kind == 0 {
				continue
			} // A nil second root is a valid single-root all-object query.
			g := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			root := Gia_ManAppendAnd(g, a, b)
			co := Gia_ManAppendCo(g, root)
			other := Gia_ManStart(8)
			foreign := Gia_ManAppendCi(other)
			bad := []*Gia_Obj_t{nil, Gia_ObjFromLit(other, foreign), Gia_ObjFromLit(g, co)}[kind]
			p := Cbs_ManAlloc(g)
			cexReviewMustPanic(t, func() {
				if pair == 0 {
					Cbs_ManSolve(p, bad)
				} else {
					Cbs_ManSolve2(p, Gia_ObjFromLit(g, root), bad)
				}
			})
			Cbs_ManStop(p)
		}
	}
}

func BenchmarkCircuitSATConflictScratch(b *testing.B) {
	for mode, name := range []string{"single", "paired"} {
		b.Run(name, func(b *testing.B) {
			g, lits := cbs0ReviewGraph(8)
			p := Cbs_ManAlloc(g)
			defer Cbs_ManStop(p)
			solve := func() int32 {
				if mode == 0 {
					return Cbs_ManSolve(p, Gia_ObjFromLit(g, lits[19]))
				}
				return Cbs_ManSolve2(p, Gia_ObjFromLit(g, lits[19]), Gia_ObjFromLit(g, lits[0]))
			}
			if solve() != 1 {
				b.Fatal("hard conflict fixture is not UNSAT")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if solve() != 1 {
					b.Fatal("conflict result changed")
				}
			}
		})
	}
}

func TestCircuitSATClauseReportForms(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	p := Cbs_ManAlloc(g)
	defer Cbs_ManStop(p)
	Cbs_ManAssign(p, Gia_ObjFromLit(g, a), 0, nil, nil)
	Cbs_ManAssign(p, Gia_ObjFromLit(g, b^1), 1, nil, nil)
	Cbs_QuePush(&p.pClauses, Gia_ObjFromLit(g, a))
	Cbs_QuePush(&p.pClauses, Gia_ObjFromLit(g, b))
	clause := Cbs_QueFinish(&p.pClauses)
	if Cbs_ManClauseSize(p, clause) != 2 || Cbs_ClauseDecLevel(p, clause) != 0 {
		t.Fatal("clause size/decision level")
	}
	text := hashReviewStdout(t, func() { Cbs_ManPrintClause(p, 2, clause); Cbs_ManPrintClauseNew(p, 2, clause) })
	if text != "Level  2 : 1=1(0) 2=0(1) \nLevel  2 : +1 -2 \n" {
		t.Fatalf("source clause report forms differ: %q", text)
	}
}
