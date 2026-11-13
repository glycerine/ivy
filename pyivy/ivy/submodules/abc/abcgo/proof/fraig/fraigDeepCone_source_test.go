package fraig

import "testing"

// The source recursive cone walk stops at shared fanins and emits the root
// last. A deep chain checks traversal without depending on random simulation.
func TestDeepFanoutFreeConeTraversal(t *testing.T) {
	a, b := &Fraig_Node_t{NumPi: 0}, &Fraig_Node_t{NumPi: 1}
	root := a
	const depth = 4096
	for i := 0; i < depth; i++ {
		child := root
		if i%2 != 0 {
			child = Fraig_Not(child)
		}
		root = &Fraig_Node_t{NumPi: -1, nRefs: 1, p1: child, p2: b}
	}
	leaves, inside := Fraig_NodeVecAlloc(2), Fraig_NodeVecAlloc(depth)
	Fraig_DetectFanoutFreeCone_rec(root, leaves, inside, 1)
	if len(inside.pArray) != depth || inside.pArray[depth-1] != root {
		t.Fatalf("inside size %d or root order incorrect", len(inside.pArray))
	}
	if len(leaves.pArray) != 2 || leaves.pArray[0] != a || leaves.pArray[1] != b {
		t.Fatalf("unexpected cone leaves %v", leaves.pArray)
	}
}
