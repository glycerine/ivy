package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestFrameMappingPreservesComplementAndMissing(t *testing.T) {
	original := aig.NewGraph()
	a, b := original.NewInput(), original.NewInput()
	node := original.AND(a.Not(), b)
	buffer := aig.Aig_ObjCreate(original, &aig.Node{
		Kind:   aig.Buffer,
		Fanins: [2]aig.Edge{a.Not(), aig.InvalidEdge},
		Phase:  true,
	})
	target := aig.NewGraph()
	mappedA := target.NewInput()
	mappedB := target.NewInput()
	m := &Fra_Man_t{PManAig: original, NFramesAll: 2, PMemFraig: make([]aig.Edge, len(original.Nodes())*2)}
	for i := range m.PMemFraig {
		m.PMemFraig[i] = aig.InvalidEdge
	}
	Fra_ObjSetFraig(m, a, 1, mappedA)
	Fra_ObjSetFraig(m, b, 1, mappedB)
	if got := Fra_ObjChild0Fra(m, node, 1); got != mappedA.Not() {
		t.Fatalf("complemented fanin = %v, want %v", got, mappedA.Not())
	}
	if got := Fra_ObjChild1Fra(m, node, 1); got != mappedB {
		t.Fatalf("fanin = %v, want %v", got, mappedB)
	}
	if got := Fra_ObjChild0Fra(m, node, 0); got != aig.InvalidEdge {
		t.Fatalf("unmapped complemented fanin = %v", got)
	}
	if got := Fra_ObjChild1Fra(m, buffer, 0); got != aig.InvalidEdge {
		t.Fatalf("missing child mapping = %v, want InvalidEdge", got)
	}
}
