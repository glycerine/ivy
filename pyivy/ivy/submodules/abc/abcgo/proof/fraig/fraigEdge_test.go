package fraig

import "testing"

func TestComplementedEdgesAreCanonical(t *testing.T) {
	node := &Fraig_Node_t{Num: 7, Level: 3}
	negative := Fraig_Not(node)
	if !Fraig_IsComplement(negative) || Fraig_IsComplement(node) || Fraig_Regular(negative) != node {
		t.Fatal("incorrect edge polarity or regular node")
	}
	if Fraig_Not(negative) != node || Fraig_Not(node) != negative || Fraig_NotCond(node, false) != node || Fraig_NotCond(node, true) != negative {
		t.Fatal("complement handles are not canonical")
	}
	if Fraig_Regular(negative).Num != 7 || Fraig_Regular(negative).Level != 3 {
		t.Fatal("complemented edge lost node data")
	}
}
