package rwr

import (
	"testing"

	"github.com/glycerine/abcgo/bool/dec"
)

type rewriteTestNetwork struct {
	marks []int
	calls int
}

func (n *rewriteTestNetwork) Lookup(id int32) (any, bool)     { return id, id >= 0 }
func (n *rewriteTestNetwork) Complement(v any) any            { return ^v.(int32) }
func (n *rewriteTestNetwork) FanoutCount(any) int             { return 2 }
func (n *rewriteTestNetwork) MarkBoundary(_ []any, delta int) { n.marks = append(n.marks, delta) }
func (n *rewriteTestNetwork) LabelMFFC(any) int32             { return 3 }
func (n *rewriteTestNetwork) CountGraph(_ any, _ *dec.Dec_Graph_t, saved, level int32) int32 {
	if saved != 3 || level != 12 {
		panic("incorrect MFFC or required level")
	}
	n.calls++
	return 1
}

func TestRewriteCutsChoosesLibraryImplementation(t *testing.T) {
	p := rwrStartLibrary()
	n := &rewriteTestNetwork{}
	truth := uint16(0x6996)
	if len(p.classes[p.classMap[truth]]) == 0 {
		t.Fatal("missing XOR4 library class")
	}
	cuts := []rwrCut{{truth, []int32{-1, 1, 2, 3}}, {truth, []int32{0, 1, 2, 3}}}
	choice := p.rwrRewriteCuts(n, nil, cuts, 12, false)
	if choice.graph == nil || choice.gain != 2 || n.calls == 0 || len(n.marks) != 2 || n.marks[0] != 1 || n.marks[1] != -1 {
		t.Fatalf("choice=%+v calls=%d marks=%v", choice, n.calls, n.marks)
	}
	if p.nCutsBad != 1 || p.nCutsGood != 1 || p.nNodesConsidered != 1 || p.nNodesRewritten != 1 || p.nNodesGained != 2 || p.classScores[p.classMap[truth]] != 1 {
		t.Fatal("rewrite statistics do not match source")
	}
	for i, leaf := range choice.fanins {
		if dec.Dec_GraphLeafFunc(choice.graph, int32(i)) != leaf {
			t.Fatalf("best graph lost leaf %d", i)
		}
	}
}
