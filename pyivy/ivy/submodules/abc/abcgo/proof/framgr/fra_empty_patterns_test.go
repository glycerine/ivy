package framgr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Compiled C Fra_SmlSimulateCombGiven returns a zero-word simulation for an
// empty file when the graph has at least one input.
func TestEmptyPatternFileReturnsSimulation(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	out := g.AddOutput(x)
	name := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(name, nil, 0600); err != nil {
		t.Fatal(err)
	}
	s := Fra_SmlSimulateCombGiven(g, name, 1, 0)
	if s == nil {
		t.Fatal("empty pattern file returned nil")
	}
	if s.WordsPerFrame != 0 || s.WordsTotal != 0 || s.NonconstantOutput || len(s.NodeWords(out.NodeID())) != 0 {
		t.Fatalf("unexpected simulation: %+v", s)
	}
}

// C reserves pattern bit zero for the base assignment and stops flipping
// inputs once the sole simulation word has no remaining pattern bits.
func TestDistanceOneFitsWordBoundary(t *testing.T) {
	g := aig.NewGraph()
	for i := 0; i < 35; i++ {
		g.NewInput()
	}
	s := Fra_SmlStart(g, 0, 1, 1)
	Fra_SmlAssignDist1(s, []uint32{0, 0})
	for i, input := range g.Inputs() {
		var want uint32
		if i < 31 {
			want = uint32(1) << uint(i+1)
		}
		if got := s.NodeWords(input.NodeID())[0]; got != want {
			t.Fatalf("input %d: %#x, want %#x", i, got, want)
		}
	}
}
