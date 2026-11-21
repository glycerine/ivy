package kit

import "testing"

func TestLiteralTruthPreservesStoredLeafFunctions(t *testing.T) {
	g := Kit_GraphCreateLeaf(2, 4, 1)
	for i, leaf := range g.pNodes {
		leaf.iFunc = int32(10 + i)
	}
	if word := Kit_GraphToTruth(g); word != 0x0f0f0f0f {
		t.Fatalf("negative literal truth %08x", word)
	}
	for i, leaf := range g.pNodes {
		if leaf.iFunc != int32(10+i) {
			t.Errorf("literal truth replaced function %d with %08x", i, uint32(leaf.iFunc))
		}
	}
}

func TestInternalTruthLeavesEveryIntermediateWord(t *testing.T) {
	g := Kit_GraphCreate(3)
	a, b, c := Kit_EdgeCreate(0, 0), Kit_EdgeCreate(1, 1), Kit_EdgeCreate(2, 0)
	pair := Kit_GraphAddNodeAnd(g, a, b)
	root := Kit_GraphAddNodeOr(g, pair, c)
	Kit_GraphSetRoot(g, root)
	if got := Kit_GraphToTruth(g); got != 0xf2f2f2f2 {
		t.Fatalf("root truth %08x", got)
	}
	for i, want := range []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0x22222222, 0x0d0d0d0d} {
		if got := uint32(g.pNodes[i].iFunc); got != want {
			t.Errorf("node %d: %08x, want %08x", i, got, want)
		}
	}
}
