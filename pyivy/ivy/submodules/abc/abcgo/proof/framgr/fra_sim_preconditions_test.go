package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSimulationSourceTerminalPreconditions(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	out := g.AddOutput(gate)
	s := Fra_SmlStart(g, 0, 2, 1)
	tests := []struct {
		name string
		run  func()
	}{
		{"simulate complemented node", func() { Fra_SmlNodeSimulate(s, gate.Not(), 0) }},
		{"compare complemented input", func() { Fra_SmlNodesCompareInFrame(s, a.Not(), b, 0, 0) }},
		{"random non-input", func() { Fra_SmlAssignRandom(s, gate) }},
		{"constant non-terminal", func() { Fra_SmlAssignConst(s, gate, 1, 0) }},
		{"copy complemented output", func() { Fra_SmlNodeCopyFanin(s, out.Not(), 0) }},
		{"transfer non-input destination", func() { Fra_SmlNodeTransferNext(s, out, gate, 0) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("expected source assertion")
				}
			}()
			tc.run()
		})
	}
}

func TestSimulationSourceTerminalOperations(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	out := g.AddOutput(a.Not())
	s := Fra_SmlStart(g, 0, 2, 1)
	Fra_SmlAssignConst(s, a, 1, 0)
	Fra_SmlNodeCopyFanin(s, out, 0)
	if got := s.FrameWords(out.NodeID(), 0)[0]; got != 0 {
		t.Fatalf("complemented output = %#x", got)
	}
	Fra_SmlNodeTransferNext(s, out, a, 0)
	if got := s.FrameWords(a.NodeID(), 1)[0]; got != 0 {
		t.Fatalf("transferred input = %#x", got)
	}
	Fra_SmlAssignConst(s, g.Const1(), 1, 1)
	if got := s.FrameWords(g.Const1().NodeID(), 1)[0]; got != ^uint32(0) {
		t.Fatalf("constant word = %#x", got)
	}
	Fra_SmlAssignRandom(s, a)
}
