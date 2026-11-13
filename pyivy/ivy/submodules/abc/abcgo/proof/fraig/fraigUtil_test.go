package fraig

import "testing"

func TestTraversalAndTfi(t *testing.T) {
	a, b, c := &Fraig_Node_t{Num: 1, NumPi: 0}, &Fraig_Node_t{Num: 2, NumPi: 1}, &Fraig_Node_t{Num: 3, NumPi: 2}
	n := &Fraig_Node_t{Num: 4, NumPi: -1, p1: a, p2: Fraig_Not(b)}
	root := &Fraig_Node_t{Num: 5, NumPi: -1, p1: n, p2: c}
	m := &Fraig_Man_t{vOutputs: &Fraig_NodeVec_t{pArray: []*Fraig_Node_t{Fraig_Not(root)}}}
	dfs := Fraig_Dfs(m, 0).pArray
	if len(dfs) != 5 || dfs[len(dfs)-1] != root {
		t.Fatalf("DFS = %v", dfs)
	}
	if Fraig_CheckTfi(m, a, root) != 1 || Fraig_CheckTfi(m, root, a) != 0 || Fraig_CheckTfi2(m, b, root) != 1 {
		t.Fatal("incorrect fanin result")
	}
	Fraig_ManMarkRealFanouts(m)
	if a.nFanouts != 1 || n.nFanouts != 1 {
		t.Fatal("incorrect fanout count")
	}
}
func TestMuxRecognition(t *testing.T) {
	c, x, y := &Fraig_Node_t{Num: 1, NumPi: 0}, &Fraig_Node_t{Num: 2, NumPi: 1}, &Fraig_Node_t{Num: 3, NumPi: 2}
	a := &Fraig_Node_t{Num: 4, NumPi: -1, p1: c, p2: x}
	b := &Fraig_Node_t{Num: 5, NumPi: -1, p1: Fraig_Not(c), p2: y}
	root := &Fraig_Node_t{Num: 6, NumPi: -1, p1: Fraig_Not(a), p2: Fraig_Not(b)}
	if Fraig_NodeIsMuxType(root) != 1 {
		t.Fatal("MUX not recognized")
	}
	var thenNode, elseNode *Fraig_Node_t
	if control := Fraig_NodeRecognizeMux(root, &thenNode, &elseNode); control != c || thenNode != Fraig_Not(x) || elseNode != Fraig_Not(y) {
		t.Fatal("wrong MUX decomposition")
	}
}
