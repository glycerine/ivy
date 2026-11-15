package gia

import (
	"math/rand"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestCofactorRankingStablePrefixes(t *testing.T) {
	rng := rand.New(rand.NewSource(137))
	for _, limit := range []int32{0, 1, 2, 7, 64, 257} {
		nodes := Vec_PtrAlloc(int(limit))
		var all []*Cof_Obj_t
		for i := 0; i < 256; i++ {
			node := &Cof_Obj_t{Id: int32(i + 1), refs: int32(rng.Intn(11))}
			all = append(all, node)
			Cof_ManInsertEntry_rec(nodes, node, limit)
			expected := slices.Clone(all)
			sort.SliceStable(expected, func(i, j int) bool { return expected[i].refs > expected[j].refs })
			expected = expected[:min(len(expected), int(limit))]
			got := Vec_PtrArray(nodes)
			if len(got) != len(expected) {
				t.Fatalf("limit%d step%d len%d", limit, i, len(got))
			}
			for j := range expected {
				if got[j] != expected[j] {
					t.Fatalf("limit%d step%d rank%d got%v want%d", limit, i, j, got[j], expected[j].Id)
				}
			}
		}
	}
}

func BenchmarkCofactorRanking(b *testing.B) {
	nodes := make([]Cof_Obj_t, 4096)
	for i := range nodes {
		nodes[i].Id = int32(i + 1)
		nodes[i].refs = int32((i * 37) % 1021)
	}
	ranked := Vec_PtrAlloc(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vec.Vec_PtrClear(ranked)
		for j := range nodes {
			Cof_ManInsertEntry_rec(ranked, &nodes[j], 100)
		}
	}
}

func TestCofactorTraversalWrapAndReferenceRestoration(t *testing.T) {
	g := cofReviewGraph(19)
	defer Gia_ManStop(g)
	p := Cof_ManCreateLogicSimple(g)
	defer Cof_ManStop(p)
	for _, obj := range p.objects {
		obj.Value = 1
	}
	p.nTravIds = 1<<31 - 1
	roots := []*Cof_Obj_t{p.objects[Gia_ManCo(g, 0).id]}
	support := Cof_ManSuppSize(p, roots, 1)
	if p.nTravIds != 1 {
		t.Fatalf("traversal ID did not restart: %d", p.nTravIds)
	}
	if support == 0 || Cof_ManSuppSize(p, roots, 1) != support {
		t.Fatal("stale traversal values contaminated support")
	}
	before := make([]int32, len(p.objects))
	for i, obj := range p.objects {
		before[i] = Cof_ObjFanoutNum(obj)
	}
	for _, obj := range p.objects {
		if Cof_ObjIsNode(obj) == 0 {
			continue
		}
		Cof_ObjMffcSize(obj)
		for i, other := range p.objects {
			if Cof_ObjFanoutNum(other) != before[i] {
				t.Fatalf("root%d damaged references at node%d", obj.Id, i)
			}
		}
	}
}

func TestCofactorEmptyFanioReportsFiniteAverages(t *testing.T) {
	inputOnly := Gia_ManStart(4)
	Gia_ManAppendCo(inputOnly, Gia_ManAppendCi(inputOnly))
	for _, g := range []*Gia_Man_t{Gia_ManStart(1), inputOnly} {
		p := Cof_ManCreateLogicSimple(g)
		text := hashReviewStdout(t, func() { Cof_ManPrintFanio(p) })
		if strings.Contains(text, "NaN") || strings.Contains(text, "Inf") {
			t.Fatalf("non-finite empty report: %s", text)
		}
		if !strings.Contains(text, "MFFCs: Max = 0. Ave =  0.00.") {
			t.Fatalf("empty MFFC statistic: %s", text)
		}
		Cof_ManStop(p)
		Gia_ManStop(g)
	}
}
