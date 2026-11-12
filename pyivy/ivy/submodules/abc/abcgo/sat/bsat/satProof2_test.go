package bsat

import (
	"reflect"
	"testing"
)

func proofClause(id, learned int32) *clause { return &clause{lrn: learned, lits: []lit{id}} }
func TestProofDependencyChainsGrowCompact(t *testing.T) {
	p := Prf_ManAlloc()
	Prf_ManRestart(p, nil, 10, 130)
	Prf_ManChainStart(p, proofClause(5, 0))
	Prf_ManChainResolve(p, proofClause(64, 0))
	Prf_ManChainStop(p)
	Prf_ManChainStart(p, proofClause(10, 1))
	Prf_ManChainResolve(p, proofClause(129, 0))
	Prf_ManChainStop(p)
	if got := Prf_ManUnsatCore(p); !reflect.DeepEqual(got, []int32{129, 64, 5}) {
		t.Fatal(got)
	}
	Prf_ManGrow(p, 512)
	if p.nWords != 8 || Prf_ManSize(p) != 2 {
		t.Fatal("growth")
	}
	Prf_ManChainStart(p, proofClause(11, 1))
	Prf_ManChainStop(p)
	Prf_ManAddSaved(p, 12, 20)
	Prf_ManAddSaved(p, 10, 21)
	Prf_ManCompact(p, 22)
	if p.iFirst != 20 || Prf_ManSize(p) != 2 || !reflect.DeepEqual(Prf_ManUnsatCore(p), []int32{64, 5}) {
		t.Fatal("compaction", Prf_ManUnsatCore(p))
	}
	Prf_ManShrink(p, 21)
	if !reflect.DeepEqual(Prf_ManUnsatCore(p), []int32{129, 64, 5}) {
		t.Fatal("shrink")
	}
	Prf_ManStop(p)
	if p.vInfo != nil || p.vId2Pr != nil {
		t.Fatal("ownership")
	}
}
func TestProofProblemIDMapping(t *testing.T) {
	p := Prf_ManAlloc()
	mapping := []int32{2, -1, 0, 2}
	Prf_ManRestart(p, &mapping, 5, 3)
	Prf_ManChainStart(p, proofClause(0, 0))
	Prf_ManChainResolve(p, proofClause(1, 0))
	Prf_ManChainResolve(p, proofClause(2, 0))
	Prf_ManChainStop(p)
	if got := Prf_ManUnsatCore(p); !reflect.DeepEqual(got, []int32{3, 2, 0}) {
		t.Fatal(got)
	}
}
func TestProofBitWordBoundaries(t *testing.T) {
	for _, test := range []struct{ bits, words int32 }{{0, 0}, {1, 1}, {63, 1}, {64, 1}, {65, 2}, {129, 3}} {
		if got := Prf_BitWordNum(test.bits); got != test.words {
			t.Fatal(test, got)
		}
	}
}
func TestProofMappingTracksExternalGrowth(t *testing.T) {
	p := Prf_ManAlloc()
	mapping := make([]int32, 1, 1)
	Prf_ManRestart(p, &mapping, 0, 130)
	Prf_ManChainStart(p, proofClause(0, 0))
	Prf_ManChainStop(p)
	mapping = append(mapping, 64)
	Prf_ManChainStart(p, proofClause(1, 0))
	Prf_ManChainStop(p)
	if got := Prf_ManUnsatCore(p); !reflect.DeepEqual(got, []int32{1}) {
		t.Fatal("mapping reallocation lost", got)
	}
}
