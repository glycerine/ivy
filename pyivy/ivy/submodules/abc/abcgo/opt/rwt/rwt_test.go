package rwt

import (
	"testing"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestLoadedRewriteForestAndDecompositions(t *testing.T) {
	p := Rwt_ManStart(0)
	if p == nil {
		t.Fatal("Rwt_ManStart returned nil")
	}
	defer Rwt_ManStop(p)
	if len(p.vForest) <= 5 || len(p.vClasses) != 222 || len(p.nodeGraphs) == 0 {
		t.Fatalf("incomplete rewrite library: forest=%d classes=%d graphs=%d", len(p.vForest), len(p.vClasses), len(p.nodeGraphs))
	}
	for node, graph := range p.nodeGraphs {
		if graph == nil {
			t.Fatalf("nil decomposition for truth %04x", node.uTruth)
		}
		if got := uint16(dec.Dec_GraphDeriveTruth(graph)); got != uint16(node.uTruth) {
			t.Fatalf("decomposition truth mismatch: graph=%04x node=%04x", got, node.uTruth)
		}
	}
}

func TestComplementedEdgeHandles(t *testing.T) {
	base := &Rwt_Node_t{}
	neg := &Rwt_Node_t{regular: base, fComplement: true, neg: base}
	base.neg = neg
	if Rwt_IsComplement(base) != 0 || Rwt_IsComplement(neg) != 1 {
		t.Fatal("complement flag is incorrect")
	}
	if Rwt_Regular(base) != base || Rwt_Regular(neg) != base {
		t.Fatal("regular edge does not identify its base node")
	}
	if Rwt_Not(base) != neg || Rwt_Not(neg) != base || Rwt_NotCond(base, 0) != base || Rwt_NotCond(base, 1) != neg {
		t.Fatal("edge complement did not toggle")
	}
}

func TestIvyIntegrationHooks(t *testing.T) {
	p := Rwt_ManStart(0)
	defer Rwt_ManStop(p)
	if Rwt_ManReadCurrentLeaves(p) == nil {
		t.Fatal("RWT did not expose current leaves")
	}
	var node *Rwt_Node_t
	for class := range p.vClasses {
		if nodes := Rwt_ManClassSubgraphs(p, int32(class)); len(nodes) != 0 {
			node = nodes[0]
			break
		}
	}
	if node == nil {
		t.Fatal("RWT has no loaded class members")
	}
	if Rwt_ManNodeDecomposition(p, node) == nil {
		t.Fatal("RWT node has no decomposition graph")
	}
	graph := dec.Dec_GraphCreateConst0()
	Rwt_ManStoreBest(p, graph, true, []any{"fanin"}, uint16(0xaaaa), 2, false)
	if Rwt_ManReadDecs(p) != graph || Rwt_ManReadCompl(p) != 1 || p.nNodesGained != 2 || p.nNodesRewritten != 1 || vec.Vec_PtrSize(Rwt_ManReadLeaves(p)) != 1 {
		t.Fatal("best-cut integration state was not recorded")
	}
}

func TestCanonicalLookupForFourInputVariable(t *testing.T) {
	p := Rwt_ManStart(1)
	defer Rwt_ManStop(p)
	if p.pPractical[0] != 1 || p.pPractical[1] != 1 || p.pPractical[3] != 1 {
		t.Fatalf("unexpected practical-class seeds: 0=%d 1=%d 3=%d", p.pPractical[0], p.pPractical[1], p.pPractical[3])
	}
	class, phase, perm := Rwt_ManCanonInfo(p, 0xaaaa)
	if len(perm) != 4 || phase != p.pPhases[0xaaaa] || p.pMap[0xaaaa] != class {
		t.Fatalf("bad NPN tables for four-input variable: class=%d perm=%v canon=%04x", class, perm, p.puCanons[0xaaaa])
	}
}
