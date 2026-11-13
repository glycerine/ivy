package fraig

import "testing"

func TestFanoutAddRemoveAndTransfer(t *testing.T) {
	a, b, dst := &Fraig_Node_t{}, &Fraig_Node_t{}, &Fraig_Node_t{}
	first := &Fraig_Node_t{p1: a, p2: b}
	second := &Fraig_Node_t{p1: b, p2: Fraig_Not(a)}
	third := &Fraig_Node_t{p1: a, p2: b}
	for _, n := range []*Fraig_Node_t{first, second, third} {
		Fraig_NodeAddFaninFanout(a, n)
	}
	if n := Fraig_NodeGetFanoutNum(a); n != 3 {
		t.Fatalf("fanouts = %d", n)
	}
	Fraig_NodeRemoveFaninFanout(a, second)
	if n := Fraig_NodeGetFanoutNum(a); n != 2 {
		t.Fatalf("after removal = %d", n)
	}
	Fraig_NodeAddFaninFanout(a, second)
	Fraig_NodeTransferFanout(a, dst)
	if Fraig_NodeGetFanoutNum(a) != 0 || Fraig_NodeGetFanoutNum(dst) != 3 {
		t.Fatal("transfer lost fanouts")
	}
	if Fraig_Regular(second.p2) != dst || !Fraig_IsComplement(second.p2) {
		t.Fatal("transfer lost complemented phase")
	}
}
