package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// ABC's empty-AIG correspondence path unconditionally starts representative
// storage; Aig_ManReprStart asserts that it is not already active.
func TestEmptyCorrespondenceRejectsExistingRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManReprStart(g, aig.Aig_ManObjNumMax(g))
	defer func() {
		if recover() == nil {
			t.Fatal("expected the source representative-storage precondition to fail")
		}
	}()
	Ssw_SignalCorrespondence(g, nil)
}
