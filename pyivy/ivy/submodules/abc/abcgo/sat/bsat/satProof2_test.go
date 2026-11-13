package bsat

import (
	"os"
	"os/exec"
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
	Prf_ManAddSaved(p, 10, 20)
	Prf_ManAddSaved(p, 12, 21)
	Prf_ManCompact(p, 22)
	if p.iFirst != 20 || Prf_ManSize(p) != 2 || !reflect.DeepEqual(Prf_ManUnsatCore(p), []int32{129, 64, 5}) {
		t.Fatal("compaction", Prf_ManUnsatCore(p))
	}
	Prf_ManShrink(p, 21)
	if !reflect.DeepEqual(Prf_ManUnsatCore(p), []int32{64, 5}) {
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

// satProof2.h reserves 1000 extra rows on growth and compacts in place.
func TestProofGrowthAndCompactionCapacity(t *testing.T) {
	if oracle := os.Getenv("ABC_PROOF_GROWTH_ORACLE"); oracle != "" {
		out, err := exec.Command(oracle).CombinedOutput()
		const want = "grow words=2 rows=1500 cap=5000\ncompact words=2 rows=750 cap=5000 first=1 last=1\n"
		if err != nil || string(out) != want {
			t.Fatalf("C oracle: %v, %q", err, out)
		}
	}
	p := Prf_ManAlloc()
	Prf_ManRestart(p, nil, 0, 1)
	for i := int32(0); i < 1500; i++ {
		Prf_ManChainStart(p, proofClause(0, 0))
		Prf_ManChainStop(p)
	}
	Prf_ManGrow(p, 64)
	if p.nWords != 2 || cap(p.vInfo) != (1500+1000)*2 {
		t.Fatalf("growth: words=%d capacity=%d", p.nWords, cap(p.vInfo))
	}
	for i := int32(0); i < 1500; i += 2 {
		Prf_ManAddSaved(p, i, i/2)
	}
	Prf_ManCompact(p, 750)
	if Prf_ManSize(p) != 750 || cap(p.vInfo) != (1500+1000)*2 {
		t.Fatalf("compaction: rows=%d capacity=%d", Prf_ManSize(p), cap(p.vInfo))
	}
	for i := int32(0); i < 750; i++ {
		if Prf_ManClauseInfo(p, i)[0] != 1 {
			t.Fatalf("row %d lost its root dependency", i)
		}
	}
}
