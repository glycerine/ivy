package fraig

import "testing"

// The C cone walk includes the root even when it is shared, but stops at
// shared internal fanins; complemented edges are regularized before testing.
func TestFanoutFreeConeStopsAtSharedFanin(t *testing.T) {
	x := &Fraig_Node_t{Num: 1, NumPi: 0}
	y := &Fraig_Node_t{Num: 2, NumPi: 1}
	shared := &Fraig_Node_t{Num: 3, NumPi: -1, nRefs: 2, p1: x, p2: y}
	root := &Fraig_Node_t{Num: 4, NumPi: -1, nRefs: 3, p1: Fraig_Not(shared), p2: x}
	fanins, inside := Fraig_NodeVecAlloc(4), Fraig_NodeVecAlloc(4)
	Fraig_DetectFanoutFreeCone_rec(root, fanins, inside, 1)
	if len(fanins.pArray) != 2 || fanins.pArray[0] != shared || fanins.pArray[1] != x {
		t.Fatalf("fanins=%v", fanins.pArray)
	}
	if len(inside.pArray) != 1 || inside.pArray[0] != root {
		t.Fatalf("inside=%v", inside.pArray)
	}
}
