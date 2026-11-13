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
