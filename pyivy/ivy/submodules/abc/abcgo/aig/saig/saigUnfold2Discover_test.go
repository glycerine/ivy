package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDiscoverUnfold2Candidates(t *testing.T) {
	for _, invert := range []bool{false, true} {
		g := aig.NewGraph()
		lo := g.NewInput()
		prop := lo
		if invert {
			prop = lo.Not()
		}
		g.AddOutput(prop)
		g.AddOutput(lo)
		aig.Aig_ManSetRegNum(g, 1)
		c, classes, err := discoverUnfold2Candidates(g, 1, 0, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		want := lo
		if !invert {
			want = lo.Not()
		}
		if vec.Vec_VecSizeSize(c) != 1 || len(classes.typeI) != 1 || classes.typeI[0] != want {
			t.Fatalf("invert=%v candidates=%v classes=%v", invert, c, classes)
		}
	}
}

func TestDiscoverUnfold2TypeII(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(x.Not())
	g.AddOutput(g.AND(x, y))
	g.AddOutput(g.Const1())
	aig.Aig_ManSetRegNum(g, 2)
	_, classes, err := discoverUnfold2Candidates(g, 1, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(classes.typeII) != 1 || classes.typeII[0] != x {
		t.Fatalf("classes %v", classes)
	}
}
