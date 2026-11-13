package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceDisjunctivePdrSelectsCandidatesFromPending(t *testing.T) {
	p := aig.Aig_ManStart(10)
	state := aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, 0) // hint, source index 0
	aig.Aig_ObjCreateCo(p, 0) // pending, source index 1
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ManSetRegNum(p, 1)
	indices := &aigPoIndices{1, 0, 0, -1}
	vectors := &antecedentConsequentVectorsStruct{attrConsequentCandidates: []int32{0}}
	got := findNewDisjunctiveMonotone(p, indices, vectors)
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("source candidates: %v", got)
	}
}

func TestSourceDisjunctiveDiscoveryThroughSecondLevel(t *testing.T) {
	p := aig.Aig_ManStart(10)
	state := aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, 0)
	aig.Aig_ObjCreateCo(p, 0)
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ManSetRegNum(p, 1)
	n := &Network{Graph: p, OutputNames: []string{"hint_up", "pendingSignal", "next"}}
	got := findDisjunctiveMonotoneSignals(n)
	if len(got) != 1 || len(got[0]) != 1 || got[0][0] != 0 {
		t.Fatalf("source level-1/2 discovery: %v", got)
	}
	result := kLive(n, KLiveOptions{Directive: kCS_WITH_DISCOVER_MONOTONE_SIGNALS}, true)
	if result.Status != -1 || len(result.Barriers) != 1 || result.Barriers[0][0] != 0 {
		t.Fatalf("source preprocessing: %+v", result)
	}
}
