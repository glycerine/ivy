package gia

import (
	"slices"
	"testing"
)

func TestCircuitSAT2MiterConstantsPrecedeBudget(t *testing.T) {
	g := Gia_ManStart(16)
	a := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, 1)
	Gia_ManAppendCo(g, 0)
	Gia_ManAppendCo(g, a)
	for _, obj := range g.pObjs {
		obj.fPhase = 1
	}
	var statuses *Vec_Str_t
	store := Cbs2_ManSolveMiterNc(g, -1, &statuses, 0)
	if !slices.Equal(Vec_StrArray(statuses), []byte{0, 1, 255}) || !slices.Equal(Vec_IntArray(store), []int32{0, 0, 2, -1}) {
		t.Fatalf("constant properties must bypass solver budget: statuses%v store%v", Vec_StrArray(statuses), Vec_IntArray(store))
	}
	// Unlike CBS, source CBS2 deliberately leaves phase initialization disabled.
	for _, obj := range g.pObjs {
		if obj.fPhase != 1 {
			t.Fatal("CBS2 changed phases")
		}
	}
}

func TestCircuitSAT2RejectsInvalidIdleQueues(t *testing.T) {
	for field := 0; field < 6; field++ {
		t.Run(string(rune('a'+field)), func(t *testing.T) {
			g := Gia_ManStart(16)
			a := Gia_ManAppendCi(g)
			p := Cbs2_ManAlloc(g)
			defer Cbs2_ManStop(p)
			switch field {
			case 0:
				p.pProp.iHead = 1
			case 1:
				Cbs2_ManAssign(p, a, 0, 0, 0)
			case 2:
				p.pJust.iHead = 1
			case 3:
				p.pJust.iTail = 1
				p.pJust.pData = append(p.pJust.pData, a>>1)
			case 4:
				p.pClauses.iHead = 2
			case 5:
				p.pClauses.iTail = 2
			}
			cexReviewMustPanic(t, func() { Cbs2_ManSolve(p, a) })
		})
	}
}

func TestCircuitSAT2RejectedPairDoesNotModifySolver(t *testing.T) {
	for _, kind := range []string{"negative", "outside", "output", "grown", "typed"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			root := Gia_ManAppendAnd(g, a, b)
			co := Gia_ManAppendCo(g, root)
			p := Cbs2_ManAlloc(g)
			defer Cbs2_ManStop(p)
			bad := int32(-1)
			switch kind {
			case "outside":
				bad = 2 * g.nObjs
			case "output":
				bad = co
			case "grown":
				bad = Gia_ManAppendCi(g)
			case "typed":
				bad = Gia_ManAppendXorReal(g, a, b)
			}
			Vec_IntPush(p.vModel, 1234)
			cexReviewMustPanic(t, func() { Cbs2_ManSolve2(p, root, bad) })
			if p.pProp.iTail != 0 || !slices.Equal(Vec_IntArray(p.vModel), []int32{1234}) {
				t.Fatal("invalid pair modified solver")
			}
			for _, value := range p.vAssign {
				if value < 2 {
					t.Fatal("invalid pair assigned a root")
				}
			}
		})
	}
}

func TestCircuitSAT2ClauseReportForms(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	p := Cbs2_ManAlloc(g)
	defer Cbs2_ManStop(p)
	Cbs2_ManAssign(p, a, 0, 0, 0)
	Cbs2_ManAssign(p, b^1, 1, 0, 0)
	Cbs2_QuePush(&p.pClauses, 0)
	Cbs2_QuePush(&p.pClauses, a>>1)
	Cbs2_QuePush(&p.pClauses, b>>1)
	cube := Cbs2_QueFinish(p, 1)
	// The adjacent learned clause contains complemented assignment literals.
	clause := cube + 3
	got := hashReviewStdout(t, func() { Cbs2_ManPrintCube(p, 2, cube); Cbs2_ManPrintClause(p, 2, clause) })
	if got != "Level  2 : +1 -2 \nLevel  2 : -1 +2 \n" {
		t.Fatalf("cube/clause reports differ: %q", got)
	}
}

func BenchmarkCircuitSAT2ConflictScratch(b *testing.B) {
	for mode, name := range []string{"top-down", "fanout"} {
		b.Run(name, func(b *testing.B) {
			g, lits := cbs0ReviewGraph(8)
			p := Cbs2_ManAlloc(g)
			defer Cbs2_ManStop(p)
			p.Pars.fUseFanout = int32(mode)
			if Cbs2_ManSolve(p, lits[19]) != 1 {
				b.Fatal("hard fixture is not UNSAT")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if Cbs2_ManSolve(p, lits[19]) != 1 {
					b.Fatal("conflict verdict changed")
				}
			}
		})
	}
}
