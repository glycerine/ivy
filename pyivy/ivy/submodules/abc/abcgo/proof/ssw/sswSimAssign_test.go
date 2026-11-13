package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSimulationInputAssignment(t *testing.T) {
	g := aig.NewGraph()
	in := g.NewInput()
	p := Ssw_SmlStart(g, 0, 2, 2)
	Ssw_SmlObjAssignConst(p, in, 1, 1)
	Ssw_SmlObjAssignConstWord(p, in, 0, 1, 1)
	Ssw_SmlObjSetWord(p, in, 42, 0, 0)
	words := Ssw_SmlSimInfo(p, in)
	if words[0] != 42 || words[1] != 0 || words[2] != ^uint32(0) || words[3] != 0 {
		t.Fatalf("got %x", words)
	}
	Ssw_SmlAssignRandom(p, in)
	if words[0]&1 != 0 || words[2]&1 != 0 {
		t.Fatal("first pattern in each frame is not zero")
	}
}

func TestRandomFrameAssignmentMatchesSourceDrawsAndIsolation(t *testing.T) {
	g := aig.NewGraph()
	in := g.NewInput()
	p := Ssw_SmlStart(g, 0, 3, 2)
	defer Ssw_SmlStop(p)
	words := Ssw_SmlSimInfo(p, in)
	for i := range words {
		words[i] = 0xdeadbeef
	}
	aig.Aig_ManRandom(1)
	want := [2]uint32{Ssw_ObjRandomSim(), Ssw_ObjRandomSim()}
	aig.Aig_ManRandom(1)
	Ssw_SmlAssignRandomFrame(p, in, 1)
	for i, got := range words {
		expected := uint32(0xdeadbeef)
		if i >= 2 && i < 4 {
			expected = want[i-2]
		}
		if got != expected {
			t.Fatalf("word %d = %08x, want %08x", i, got, expected)
		}
	}
}
