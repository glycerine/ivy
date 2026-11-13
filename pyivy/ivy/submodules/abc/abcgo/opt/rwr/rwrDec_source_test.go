package rwr

import (
	"github.com/glycerine/abcgo/bool/dec"
	"testing"
)

func TestNodePreprocessPreservesComplementedLeaves(t *testing.T) {
	p := &Rwr_Man_t{}
	a := &Rwr_Node_t{Id: 1, fUsed: true}
	b := &Rwr_Node_t{Id: 2, fUsed: true}
	c := &Rwr_Node_t{Id: 3, fUsed: true}
	shared := &Rwr_Node_t{Id: 5, p0: Rwr_NodeRef{node: a}, p1: Rwr_NodeRef{node: b, complemented: true}}
	root := &Rwr_Node_t{Id: 6, uTruth: (0xaaaa & ^uint16(0xcccc)) ^ 0xf0f0, p0: Rwr_NodeRef{node: shared}, p1: Rwr_NodeRef{node: c}, fExor: true}
	graph := Rwr_NodePreprocess(p, root)
	if got := dec.Dec_GraphDeriveTruth(graph) & 0xffff; got != uint32((0xaaaa & ^uint16(0xcccc))^0xf0f0) {
		t.Fatalf("truth %04x", got)
	}
}
