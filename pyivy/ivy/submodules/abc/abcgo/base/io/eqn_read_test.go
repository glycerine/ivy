package io

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

// TestEqnReadsOneAigNodePerEquation: Io_ReadEqnNetwork builds an AIG netlist
// with one node per equation (its fanins the formula's distinct names in
// order of appearance), not a web of small SOP nodes.
func TestEqnReadsOneAigNodePerEquation(t *testing.T) {
	n, err := ParseEqn("t.eqn", "INORDER = a b c; OUTORDER = y z;\ny = a*b + !c;\nz = c*(a+b) + a;\n", true)
	if err != nil {
		t.Fatal(err)
	}
	if abc.Abc_NtkIsAigNetlist(n) == 0 {
		t.Fatalf("type %d function %d, want an AIG netlist", n.Type, n.Function)
	}
	nodes := n.ObjectsOfKind(abc.ObjectNode)
	if len(nodes) != 2 {
		t.Fatalf("%d nodes, want 2", len(nodes))
	}
	var fanins []string
	for _, e := range nodes[1].Fanins() {
		fanins = append(fanins, abc.Abc_ObjName(e.Object))
	}
	if len(fanins) != 3 || fanins[0] != "c" || fanins[1] != "a" || fanins[2] != "b" {
		t.Fatalf("z fanins %v, want [c a b]", fanins)
	}
	one, err := ParseEqn("u.eqn", "INORDER = a b c; OUTORDER = y; y = a*b + !c;", true)
	if err != nil {
		t.Fatal(err)
	}
	checkBooleanNetwork(t, one, func(v []bool) bool { return (v[0] && v[1]) || !v[2] })
}
