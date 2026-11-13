package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSimulationQueriesRespectSourcePrefixRules(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	s := Fra_SmlStart(g, 1, 2, 1)
	copy(s.NodeWords(a.NodeID()), []uint32{0xf, 0, 0})
	copy(s.NodeWords(b.NodeID()), []uint32{0, 0, 0})
	if got := Fra_SmlNodeCountOnes(s, a); got != 4 {
		t.Errorf("count includes prefix: got %d", got)
	}
	if got := Fra_SmlNodeIsZero(s, a); got != 1 {
		t.Errorf("zero excludes prefix: got %d", got)
	}
	if got := Fra_SmlNodeIsConst(s, a); got != 1 {
		t.Errorf("constant excludes prefix: got %d", got)
	}
	if got := Fra_SmlNodesAreEqual(s, a, b); got != 1 {
		t.Errorf("equality excludes prefix: got %d", got)
	}
	if got := Fra_SmlNodesCompareInFrame(s, a, b, 0, 0); got != 0 {
		t.Errorf("frame comparison includes frame zero: got %d", got)
	}
	if got := Fra_SmlNodesCompareInFrame(s, a, b, 1, 2); got != 1 {
		t.Errorf("different zero frames compare equal: got %d", got)
	}
	s.NodeWords(b.NodeID())[2] = 1
	if got := Fra_SmlNodesAreEqual(s, a, b); got != 0 {
		t.Errorf("equality missed non-prefix bit: got %d", got)
	}
	if got := Fra_SmlNodeNotEquWeight(s, int32(a.NodeID()), int32(b.NodeID())); got != 1 {
		t.Errorf("difference weight excludes prefix: got %d", got)
	}
}
