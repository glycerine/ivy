package rwr

import "testing"

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
