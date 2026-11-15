package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestAreaCollectorPreservesUnrelatedMark0(t *testing.T) {
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	co := Gia_ManAppendCo(g, root)
	for _, lit := range []int32{a, root, co} {
		Gia_ManObj(g, lit>>1).fMark0 = 1
	}
	manager := Dam_ManAlloc(g)
	Dam_ManCollectSets(manager)
	for _, lit := range []int32{a, root, co} {
		if Gia_ManObj(g, lit>>1).fMark0 != 1 {
			t.Fatalf("area collection erased caller's mark0 on node%d", lit>>1)
		}
	}
}

func TestAreaCollectorClearsOnlyItsVisitedMuxMarks(t *testing.T) {
	g := Gia_ManStart(1)
	a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendAnd(g, a, b)
	mux := Gia_ManAppendMuxReal(g, c, x, d)
	unrelated := Gia_ManAppendMuxReal(g, b, a, d)
	co := Gia_ManAppendCo(g, mux)
	for _, lit := range []int32{a, x, unrelated, co} {
		Gia_ManObj(g, lit>>1).fMark0 = 1
	}
	manager := Dam_ManAlloc(g)
	for repeat := 0; repeat < 2; repeat++ {
		Dam_ManCollectSets(manager)
		if len(manager.vVisit) != 1 || manager.vVisit[0] != mux>>1 {
			t.Fatalf("visited MUX list %v", manager.vVisit)
		}
		if Gia_ManObj(g, mux>>1).fMark0 != 0 {
			t.Fatal("temporary MUX mark leaked")
		}
		for _, lit := range []int32{a, x, unrelated, co} {
			if Gia_ManObj(g, lit>>1).fMark0 != 1 {
				t.Fatalf("caller mark erased on node%d", lit>>1)
			}
		}
	}
}

func BenchmarkAreaCollectionAndExtraction(b *testing.B) {
	b.ReportAllocs()
	for run := 0; run < b.N; run++ {
		g := areaReviewGraph(5, 0)
		p := Dam_ManAlloc(g)
		p.nLevelMax = Gia_ManSetLevels(g, nil)
		p.vNodLevR = append([]int32(nil), Vec_IntArray(Gia_ManReverseLevel(g))...)
		Dam_ManCreatePairs(p, 0)
		if vec.Vec_QueTopPriority(p.vQue) >= 2 {
			Dam_ManUpdate(p, int32(vec.Vec_QuePop(p.vQue)))
		}
	}
}
