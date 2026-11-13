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
	for _, nodes := range classes {
		for _, id := range nodes {
			graph := rwrDecomposition(forest, id)
			if got := uint16(dec.Dec_GraphDeriveTruth(graph)); got != forest[id].truth {
				t.Fatalf("node %d: decomposition %#x, want %#x", id, got, forest[id].truth)
			}
		}
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
