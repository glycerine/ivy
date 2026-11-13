package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestFilterUnfold2Candidates(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	c := vec.Vec_VecAlloc(1)
	vec.Vec_VecPush(c, 0, lo)
	vec.Vec_VecPush(c, 0, pi)
	classes, err := filterUnfold2Candidates(g, c, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(classes.typeI) != 1 || classes.typeI[0] != lo || len(classes.typeII) != 0 || vec.Vec_VecSizeSize(c) != 1 {
		t.Fatalf("classes=%v candidates=%d", classes, vec.Vec_VecSizeSize(c))
	}
}

func TestFilterUnfold2TypeII(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(g.AND(x, y))
	g.AddOutput(g.Const1())
	aig.Aig_ManSetRegNum(g, 2)
	c := vec.Vec_VecAlloc(1)
	vec.Vec_VecPush(c, 0, x)
	classes, err := filterUnfold2Candidates(g, c, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(classes.typeI) != 0 || len(classes.typeII) != 1 || classes.typeII[0] != x {
		t.Fatalf("classes=%v", classes)
	}
}
