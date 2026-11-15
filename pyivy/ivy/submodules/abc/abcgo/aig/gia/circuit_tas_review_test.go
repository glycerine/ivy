package gia

import (
	"fmt"
	"slices"
	"testing"
)

func TestTasRejectsTypedAndChangedGraphs(t *testing.T) {
	for _, kind := range []string{"xor-constructor", "mux-constructor", "xor-query", "grown-query"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(16)
			a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			Gia_ManFillValue(g)
			if kind == "xor-constructor" || kind == "mux-constructor" {
				if kind == "xor-constructor" {
					Gia_ManAppendXorReal(g, a, b)
				} else {
					g.EnableMuxes()
					Gia_ManAppendMuxReal(g, a, b, c)
				}
				cexReviewMustPanic(t, func() { Tas_ManAlloc(g, 100) })
				return
			}
			p := Tas_ManAlloc(g, 100)
			defer Tas_ManStop(p)
			if kind == "xor-query" {
				Gia_ManAppendXorReal(g, a, b)
			} else {
				Gia_ManAppendCi(g)
			}
			cexReviewMustPanic(t, func() { Tas_ManSolve(p, Gia_ObjFromLit(g, a), nil) })
			if p.pProp.iTail != 0 {
				t.Fatal("invalid graph query assigned root")
			}
		})
	}
}

func TestTasRejectedAssumptionsPreserveState(t *testing.T) {
	for _, kind := range []string{"output", "foreign", "nil"} {
		t.Run(kind, func(t *testing.T) {
			g := Gia_ManStart(16)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			root := Gia_ManAppendAnd(g, a, b)
			co := Gia_ManAppendCo(g, root)
			Gia_ManFillValue(g)
			p := Tas_ManAlloc(g, 100)
			defer Tas_ManStop(p)
			bad := Gia_ObjFromLit(g, co)
			if kind == "nil" {
				bad = nil
			}
			if kind == "foreign" {
				other := Gia_ManStart(8)
				bad = Gia_ObjFromLit(other, Gia_ManAppendCi(other))
			}
			Vec_IntPush(p.vModel, 1234)
			cexReviewMustPanic(t, func() { Tas_ManSolveArray(p, []*Gia_Obj_t{Gia_ObjFromLit(g, root), bad}) })
			if !slices.Equal(Vec_IntArray(p.vModel), []int32{1234}) || p.pProp.iTail != 0 {
				t.Fatal("rejected assumptions changed solver")
			}
			for _, obj := range g.pObjs {
				if obj.fMark0 != 0 || obj.Value != ^uint32(0) {
					t.Fatal("rejected assumptions changed graph")
				}
			}
		})
	}
}

func TestTasStopCancelsActiveAssignments(t *testing.T) {
	g := Gia_ManStart(16)
	a := Gia_ManAppendCi(g)
	Gia_ManFillValue(g)
	p := Tas_ManAlloc(g, 100)
	Tas_ManAssign(p, Gia_ObjFromLit(g, a), 0, nil, nil)
	Tas_ManStop(p)
	if obj := g.pObjs[a>>1]; obj.fMark0 != 0 || obj.fMark1 != 0 || obj.Value != ^uint32(0) {
		t.Fatal("active Stop retained graph assignment")
	}
	Tas_ManStop(p)
	Tas_ManStop(nil)
}

func BenchmarkTasConflictScratch(b *testing.B) {
	for _, pair := range []bool{false, true} {
		b.Run(fmt.Sprintf("pair%v", pair), func(b *testing.B) {
			g, lits := cbs0ReviewGraph(8)
			Gia_ManFillValue(g)
			Gia_ManCleanPhase(g)
			p := Tas_ManAlloc(g, 1000)
			defer Tas_ManStop(p)
			var second *Gia_Obj_t
			if pair {
				second = Gia_ObjFromLit(g, lits[0])
			}
			root := Gia_ObjFromLit(g, lits[19])
			if Tas_ManSolve(p, root, second) != 1 {
				b.Fatal("hard fixture is not UNSAT")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if Tas_ManSolve(p, root, second) != 1 {
					b.Fatal("conflict verdict changed")
				}
			}
		})
	}
}
