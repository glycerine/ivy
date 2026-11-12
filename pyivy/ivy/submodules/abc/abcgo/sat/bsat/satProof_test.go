package bsat

import (
	"reflect"
	"slices"
	"testing"
)

func TestResolutionGraphCoreAndTraversal(t *testing.T) {
	g := NewProofGraph()
	a := g.Append([]int32{1, 9})
	b := g.Append([]int32{a << 2, 13})
	root := g.Append([]int32{a << 2, b<<2 | 2, 9})
	want := []int32{3, 2, 0}
	if got := Proof_DeriveCore(g, root); !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	if got := Proof_DeriveCore(g, root); !reflect.DeepEqual(got, want) {
		t.Fatal("repeated collection", got)
	}
	used := Proof_CollectUsedRec(g, []int32{root})
	Proof_CleanCollected(g, used)
	slices.Sort(used)
	iter := Proof_CollectUsedIter(g, []int32{root}, 1)
	if !reflect.DeepEqual(used, iter) {
		t.Fatal(used, iter)
	}
	Proof_CleanCollected(g, iter)
	Sat_ProofCheck0(g)
}
func TestResolutionGraphCompactionAndPivot(t *testing.T) {
	g := NewProofGraph()
	dead := g.Append([]int32{5})
	a := g.Append([]int32{1, 9})
	pivot := g.next
	root := g.Append([]int32{a<<2 | 2, 13})
	g.Append([]int32{dead << 2, 17})
	roots := []int32{-1, root}
	newPivot := Sat_ProofReduce(g, &roots, pivot)
	if len(g.nodes) != 2 || roots[1] != newPivot || !reflect.DeepEqual(Proof_DeriveCore(g, roots[1]), []int32{3, 2, 0}) {
		t.Fatal(roots, newPivot, g)
	}
	if g.nodes[2].pEnts[0] != 1 || g.nodes[newPivot].pEnts[0]&2 == 0 {
		t.Fatal("tag lost")
	}
	Sat_ProofCheck0(g)
}
func TestResolutionGraphRejectsForwardReferences(t *testing.T) {
	g := NewProofGraph()
	defer func() {
		if recover() == nil {
			t.Fatal("accepted forward proof reference")
		}
	}()
	g.Append([]int32{40 << 2})
}
