package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestFlipConePdrMutatesSourceForInvariant(t *testing.T) {
	p := aig.Aig_ManStart(10)
	state := aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ObjCreateCo(p, state.Not())
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ManSetRegNum(p, 1)
	flipConePdr(p, kCS_WITH_SAFETY_INVARIANTS, 0, 1, 0)
	if got := outputDriver(p, 0); got != aig.Edge(1) {
		t.Fatalf("source output must be patched to state AND !state, got %v", got)
	}
}
