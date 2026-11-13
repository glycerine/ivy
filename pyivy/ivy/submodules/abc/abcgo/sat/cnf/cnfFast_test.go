package cnf

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCutDeriveTruthRepeatsAcrossWord(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	root := g.AND(x, y)
	leaves, nodes := vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(leaves, x)
	vec.Vec_PtrPush(leaves, y)
	vec.Vec_PtrPush(nodes, root)
	if got, want := Cnf_CutDeriveTruth(g, leaves, nodes), uint64(0x8888888888888888); got != want {
		t.Fatalf("truth = %#016x, want %#016x", got, want)
	}
}
