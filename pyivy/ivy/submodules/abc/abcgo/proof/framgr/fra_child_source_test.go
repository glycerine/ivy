package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// fra.h requires the parent object, unlike its fanin edges, to be regular.
func TestSourceMappedChildRejectsComplementedParent(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	n := g.AND(a, b)
	m := &Fra_Man_t{PManAig: g, NFramesAll: 1, PMemFraig: make([]aig.Edge, len(g.Nodes()))}
	for _, get := range []struct {
		name string
		fn   func(*Fra_Man_t, aig.Edge, int32) aig.Edge
	}{{"fanin0", Fra_ObjChild0Fra}, {"fanin1", Fra_ObjChild1Fra}} {
		t.Run(get.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "fra: complemented parent object" {
					t.Fatalf("panic=%v, want complemented parent assertion", got)
				}
			}()
			get.fn(m, n.Not(), 0)
		})
	}
}
