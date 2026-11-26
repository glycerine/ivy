package aig

import "testing"

// The native ABC binary is built with GCC, which evaluates the two Aig_And
// arguments of Aig_Or( p, Aig_And(p, p0, Aig_Not(p1)), Aig_And(p, Aig_Not(p0), p1) )
// in Aig_Exor right to left. Native ABC therefore creates AND(!p0, p1) before
// AND(p0, !p1); node IDs (and with them CNF variable order and SAT search)
// depend on this. Observed with a driver linked against libabc.a:
//
//	a=1 b=2 x=5'   3: 1' 2   4: 1 2'
func TestAigExorCreatesNodesInNativeOrder(t *testing.T) {
	p := Aig_ManStart(10)
	a := Aig_ObjCreateCi(p)
	b := Aig_ObjCreateCi(p)
	x := Aig_Exor(p, a, b)
	n3, n4 := p.Node(Aig_ManObj(p, 3)), p.Node(Aig_ManObj(p, 4))
	if n3.Fanins != [2]Edge{a.Not(), b} {
		t.Fatalf("node 3 fanins = %v, want [!a b]", n3.Fanins)
	}
	if n4.Fanins != [2]Edge{a, b.Not()} {
		t.Fatalf("node 4 fanins = %v, want [a !b]", n4.Fanins)
	}
	if x.NodeID() != 5 || !x.Complemented() {
		t.Fatalf("x = %d (compl %v), want 5'", x.NodeID(), x.Complemented())
	}
}
