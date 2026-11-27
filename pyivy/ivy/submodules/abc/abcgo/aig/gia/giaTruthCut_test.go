package gia

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ObjComputeTruthTableCut (giaTruth.c) computes the tables of a cut's
// internal nodes in vTtMemory and resets only the entries it numbered. The
// port checks it against the reference evaluator on AND/XOR/MUX graphs, with
// cuts of CIs and of internal nodes, over repeated calls (a stale number
// left by one call would corrupt the next).
func TestComputeTruthTableCutMatchesEvaluator(t *testing.T) {
	g := Gia_ManStart(64)
	g.EnableMuxes()
	const nCis = 8
	var lits []int32
	for i := 0; i < nCis; i++ {
		lits = append(lits, Gia_ManAppendCi(g))
	}
	seed := uint32(12345)
	next := func(n int) int32 {
		seed = seed*1103515245 + 12345
		return int32((seed >> 16) % uint32(n))
	}
	for k := 0; k < 60; k++ {
		a, b, c := lits[next(len(lits))]^next(2), lits[next(len(lits))]^next(2), lits[next(len(lits))]^next(2)
		if a>>1 == b>>1 || a>>1 == c>>1 || b>>1 == c>>1 {
			continue
		}
		var l int32
		switch k % 3 {
		case 0:
			l = Gia_ManAppendAnd(g, a, b)
		case 1:
			l = Gia_ManAppendXorReal(g, a, b)
		default:
			l = Gia_ManAppendMuxReal(g, c, b, a)
		}
		lits = append(lits, l)
	}
	Gia_ManAppendCo(g, lits[len(lits)-1])
	Gia_ObjComputeTruthTableStart(g, nCis)
	defer Gia_ObjComputeTruthTableStop(g)
	cis := slices.Clone(giaTruthCis(g))
	checked := 0
	for round := 0; round < 2; round++ {
		for _, lit := range lits[nCis:] {
			root := g.pObjs[lit>>1]
			// cut of CIs
			want := giaNewTruthEval(g, nCis, cis).node(root.id)
			leaves := vec.Vec_IntAllocArrayCopy(cis, len(cis))
			got := Gia_ObjComputeTruthTableCut(g, root, leaves)
			if !slices.Equal(got[:len(want)], want) {
				t.Fatalf("round %d node %d (CI cut): %x, want %x", round, root.id, got, want)
			}
			// cut of the root's fanins (internal leaves)
			fanins := giaObjectFanins(g, root)
			want = giaNewTruthEval(g, nCis, fanins).node(root.id)
			got = Gia_ObjComputeTruthTableCut(g, root, vec.Vec_IntAllocArrayCopy(fanins, len(fanins)))
			if !slices.Equal(got[:len(want)], want) {
				t.Fatalf("round %d node %d (fanin cut): %x, want %x", round, root.id, got, want)
			}
			checked++
		}
	}
	for id, n := range vec.Vec_IntArray(g.vTtNums) {
		if n != giaTruthUnnumbered {
			t.Fatalf("object %d left numbered %d", id, n)
		}
	}
	if checked < 60 {
		t.Fatalf("only %d nodes checked", checked)
	}
}
