package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceArenaRejectsEmptyBarrierVector(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	for _, e := range []aig.Edge{x, x, x, state, state} {
		g.AddOutput(e)
	}
	aig.Aig_ManSetRegNum(g, 1)
	defer func() {
		r := recover()
		if r != "live: source requires nonempty barriers" {
			t.Fatalf("source requires barriers: got %v", r)
		}
	}()
	var index int32
	createNewAigWith0LivePoWithDSC(g, []aig.Edge{state}, &index, 0, 1, 2, nil)
}
func TestNativeArenaSupportsEmptyBarrierVector(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	for _, e := range []aig.Edge{x, x, x, state, state} {
		g.AddOutput(e)
	}
	aig.Aig_ManSetRegNum(g, 1)
	var index int32
	out := createNewAigWith0LivePoWithDSCMode(g, []aig.Edge{state}, &index, 0, 1, 2, nil, false)
	if got := out.RegisterCount(); got != 3 {
		t.Fatalf("native empty arena registers %d", got)
	}
	if index != 4 {
		t.Fatalf("property index %d", index)
	}
}
func TestSourceArenaRejectsPopulatedOutputVector(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	aig.Aig_ObjSetCopy(g, x, x)
	var next []aig.Edge
	monotone := []aig.Edge{x}
	defer func() {
		r := recover()
		if r != "live: source arena output vector must be empty" {
			t.Fatalf("expected source assertion, got %v", r)
		}
	}()
	createArenaViolation(g, g, x, x, [][]int32{{0}}, []aig.Edge{x}, &next, &monotone)
}
