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
