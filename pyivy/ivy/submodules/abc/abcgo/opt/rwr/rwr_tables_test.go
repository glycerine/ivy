package rwr

import (
	"testing"

	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra/npn"
)

// The C loader starts with five leaves and interprets each pair as a new
// node referring to earlier entries in the forest.
func TestPrecomputedForestReferencesEarlierNodes(t *testing.T) {
	if len(s_RwtAigSubgraphs)%2 != 0 {
		t.Fatal("forest has an incomplete entry")
	}
	for i := 0; i < len(s_RwtAigSubgraphs)/2; i++ {
		a, b := s_RwtAigSubgraphs[2*i], s_RwtAigSubgraphs[2*i+1]
		if a == 0 && b == 0 {
			if i != len(s_RwtAigSubgraphs)/2-1 {
				t.Fatalf("early terminator at entry %d", i)
			}
			return
		}
		if int(a>>2) >= i+5 || int(b>>1) >= i+5 {
			t.Fatalf("entry %d refers forward: %#x, %#x", i, a, b)
		}
	}
	t.Fatal("missing forest terminator")
}

func TestCanonicalDecompositionsPreserveTruth(t *testing.T) {
	forest := rwrLoadForest(s_RwtAigSubgraphs)
	canonical, _, _, classMap := npn.Extra_Truth4VarNPN()
	classes, _ := rwrClassifyForest(forest, canonical, classMap)
	manager := Rwr_ManStart(0)
	for class, nodes := range classes {
		if len(nodes) != len(manager.vClasses[class]) {
			t.Fatalf("class %d has %d candidates, want %d", class, len(manager.vClasses[class]), len(nodes))
		}
		for i, id := range nodes {
			actual := manager.vClasses[class][i]
			if int(actual.Id) != id || uint16(dec.Dec_GraphDeriveTruth(actual.graph)) != forest[id].truth {
				t.Fatalf("class %d candidate %d disagrees with independently decoded forest", class, i)
			}
			graph := rwrDecomposition(forest, id)
			if got := uint16(dec.Dec_GraphDeriveTruth(graph)); got != forest[id].truth {
				t.Fatalf("node %d: decomposition %#x, want %#x", id, got, forest[id].truth)
			}
		}
	}
}

func TestStartedLibraryCanReconstructFunctions(t *testing.T) {
	p := Rwr_ManStart(0)
	permutations := npn.Permutations()
	count := 0
	for class, nodes := range p.vClasses {
		for _, node := range nodes {
			if node.graph == nil || p.pMapInv[class] != node.uTruth {
				t.Fatalf("missing graph or inverse mapping for node %d", node.Id)
			}
			count++
		}
	}
	if count == 0 {
		t.Fatal("library has no canonical implementations")
	}
	for truth := 0; truth < 1<<16; truth++ {
		got := npn.Transform(p.puCanons[truth], p.pPhases[truth], permutations[p.pPerms[truth]])
		if got != uint16(truth) {
			t.Fatalf("truth %#x reconstructed as %#x", truth, got)
		}
	}
}

func TestCutEvaluationSelectsStrictlyBetterCandidate(t *testing.T) {
	p := Rwr_ManStart(0)
	class := -1
	for i, nodes := range p.vClasses {
		if len(nodes) >= 4 && dec.Dec_GraphLeaveNum(nodes[0].graph) == 4 {
			class = i
			break
		}
	}
	if class < 0 {
		t.Fatal("fixture needs a class with four implementations")
	}
	nodes := p.vClasses[class]
	inputs := []any{1, 2, 3, 4}
	calls := 0
	best, gain := p.rwrEvaluateCut(p.pMapInv[class], inputs, 5, func(g *dec.Dec_Graph_t) int32 {
		for i, want := range inputs {
			if dec.Dec_GraphLeafFunc(g, int32(i)) != want {
				t.Fatalf("leaf %d not assigned", i)
			}
		}
		n := calls
		calls++
		switch n {
		case 0:
			return -1
		case 1:
			return 3
		case 2, 3:
			return 1
		default:
			return 5
		}
	})
	if calls != len(nodes) || int(p.nSubgraphs) != len(nodes) || gain != 4 || best != nodes[2].graph {
		t.Fatalf("calls=%d candidates=%d gain=%d", calls, p.nSubgraphs, gain)
	}
	if nodes[1].nScore != 1 || nodes[1].nGain != 2 || nodes[1].nAdded != 3 || nodes[2].nScore != 1 || nodes[2].nGain != 4 || nodes[2].nAdded != 1 || nodes[3].nScore != 0 {
		t.Fatal("candidate scoring does not follow C's strict-improvement rule")
	}
}

func TestPracticalClassesSentinel(t *testing.T) {
	if len(s_RwrPracticalClasses) != 136 || s_RwrPracticalClasses[135] != 0 {
		t.Fatalf("practical classes length or terminator changed: %d", len(s_RwrPracticalClasses))
	}
	for i := 1; i < len(s_RwrPracticalClasses)-1; i++ {
		if s_RwrPracticalClasses[i] <= s_RwrPracticalClasses[i-1] {
			t.Fatalf("classes out of order at %d", i)
		}
	}
	practical := Rwr_ManGetPractical(&Rwr_Man_t{nFuncs: 1 << 16})
	count := 0
	for _, yes := range practical {
		if yes {
			count++
		}
	}
	if count != 135 {
		t.Fatalf("got %d practical classes, want 135", count)
	}
	for _, truth := range s_RwrPracticalClasses[:135] {
		if !practical[truth] {
			t.Fatalf("missing practical class %#x", truth)
		}
	}
}

func TestLoadedForestTruthsAndVolumes(t *testing.T) {
	forest := rwrLoadForest(s_RwtAigSubgraphs)
	if len(forest) != 5+len(s_RwtAigSubgraphs)/2-1 {
		t.Fatalf("unexpected forest length %d", len(forest))
	}
	for i := 5; i < len(forest); i++ {
		n := forest[i]
		if n.xor && (n.a.compl || n.b.compl) {
			t.Fatalf("XOR %d has complemented input", i)
		}
		seen := map[int]bool{}
		var visit func(int)
		visit = func(id int) {
			if id < 5 || seen[id] {
				return
			}
			seen[id] = true
			visit(forest[id].a.node)
			visit(forest[id].b.node)
		}
		visit(i)
		volume := len(seen)
		for id := range seen {
			if forest[id].xor {
				volume++
			}
		}
		if n.volume != volume {
			t.Fatalf("node %d volume %d, want %d", i, n.volume, volume)
		}
		for input := 0; input < 16; input++ {
			values := make([]bool, i+1)
			for j := 1; j < 5; j++ {
				values[j] = input&(1<<(j-1)) != 0
			}
			for j := 5; j <= i; j++ {
				v := forest[j]
				a, b := values[v.a.node] != v.a.compl, values[v.b.node] != v.b.compl
				if v.xor {
					values[j] = a != b
				} else {
					values[j] = a && b
				}
			}
			if values[i] != (n.truth&(1<<input) != 0) {
				t.Fatalf("node %d input %d: truth %#x", i, input, n.truth)
			}
		}
	}
}

func TestForestClassesContainExactlyCanonicalNodes(t *testing.T) {
	forest := rwrLoadForest(s_RwtAigSubgraphs)
	canonical, _, _, classMap := npn.Extra_Truth4VarNPN()
	classes, inverse := rwrClassifyForest(forest, canonical, classMap)
	seen := make([]bool, len(forest))
	for class, nodes := range classes {
		for _, id := range nodes {
			truth := forest[id].truth
			if seen[id] || canonical[truth] != truth || int(classMap[truth]) != class || inverse[class] != truth {
				t.Fatalf("bad class %d node %d truth %#x", class, id, truth)
			}
			seen[id] = true
		}
	}
	for id := 5; id < len(forest); id++ {
		if want := canonical[forest[id].truth] == forest[id].truth; seen[id] != want {
			t.Fatalf("node %d presence %v want %v", id, seen[id], want)
		}
	}
}
