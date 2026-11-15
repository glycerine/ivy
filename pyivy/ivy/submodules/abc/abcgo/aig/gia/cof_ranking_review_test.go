package gia

import (
	"math/rand"
	"slices"
	"sort"
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
