package rwr

import "testing"

func TestSourceNodeComplementAttributes(t *testing.T) {
	node := &Rwr_Node_t{Id: 17}
	plain := Rwr_NodeRef{node: node}
	if Rwr_IsComplement(plain) != 0 || Rwr_Regular(plain) != node {
		t.Fatal("regular node")
	}
	neg := Rwr_Not(plain)
	if Rwr_IsComplement(neg) != 1 || Rwr_Regular(neg) != node || Rwr_Not(neg) != plain {
		t.Fatal("complemented node")
	}
	if Rwr_NotCond(plain, 0) != plain || Rwr_NotCond(plain, 1) != neg {
		t.Fatal("conditional complement")
	}
}
