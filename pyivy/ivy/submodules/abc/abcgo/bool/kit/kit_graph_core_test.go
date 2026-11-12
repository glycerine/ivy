package kit

import "testing"

func TestGraphBooleanOperations(t *testing.T) {
	for kind := int32(0); kind < 2; kind++ {
		for operation := 0; operation < 4; operation++ {
			g := Kit_GraphCreate(3)
			a, b, c := Kit_EdgeCreate(0, 0), Kit_EdgeCreate(1, 0), Kit_EdgeCreate(2, 0)
			var root Kit_Edge_t
			switch operation {
			case 0:
				root = Kit_GraphAddNodeAnd(g, a, b)
			case 1:
				root = Kit_GraphAddNodeOr(g, a, b)
			case 2:
				root = Kit_GraphAddNodeXor(g, a, b, kind)
			case 3:
				root = Kit_GraphAddNodeMux(g, c, a, b, kind)
			}
			Kit_GraphSetRoot(g, root)
			truth := Kit_GraphToTruth(g)
			for x := 0; x < 32; x++ {
				a, b, c := x&1 != 0, x&2 != 0, x&4 != 0
				var want bool
				switch operation {
				case 0:
					want = a && b
				case 1:
					want = a || b
				case 2:
					want = a != b
				case 3:
					want = (c && a) || (!c && b)
				}
				if ((truth>>uint(x))&1 != 0) != want {
					t.Fatalf("operation=%d kind=%d assignment=%d truth=%x", operation, kind, x, truth)
				}
			}
			if int32(g.pNodes[root.Node].Level) != Kit_GraphLevelNum_rec(g, g.pNodes[root.Node]) {
				t.Fatal("cached level mismatch")
			}
			old := truth
			Kit_GraphComplement(g)
			if Kit_GraphToTruth(g) != ^old {
				t.Fatal("complement mismatch")
			}
		}
	}
}
func TestGraphConstantsAndStableNodes(t *testing.T) {
	if Kit_GraphToTruth(Kit_GraphCreateConst0()) != 0 || Kit_GraphToTruth(Kit_GraphCreateConst1()) != ^uint32(0) {
		t.Fatal("constant truth")
	}
	g := Kit_GraphCreateLeaf(0, 2, 0)
	first := Kit_GraphNode(g, 0)
	root := Kit_EdgeCreate(0, 0)
	for i := 0; i < 1000; i++ {
		root = Kit_GraphAddNodeAnd(g, root, Kit_EdgeCreate(1, 0))
	}
	if first != Kit_GraphNode(g, 0) || Kit_GraphNodeInt(g, first) != 0 {
		t.Fatal("node pointers moved during append")
	}
	g.eRoot = root
	if Kit_GraphToTruth(g) != 0x88888888 {
		t.Fatal("long AND chain truth")
	}
	if Kit_GraphLeafDepth_rec(g, g.pNodes[root.Node], first) != 1000 {
		t.Fatal("leaf depth")
	}
	for i := uint32(0); i < 100; i++ {
		if got := Kit_EdgeToInt_(Kit_IntToEdge_(i)); got != i {
			t.Fatal(i, got)
		}
	}
}
