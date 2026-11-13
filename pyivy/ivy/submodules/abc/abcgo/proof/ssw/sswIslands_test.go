package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestIslandMatchingExtendAndComplete(t *testing.T) {
	build := func() (*aig.Graph, aig.Edge, aig.Edge, aig.Edge) {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		x := g.AND(a, b)
		y := g.AND(x, a.Not())
		g.AddOutput(y)
		aig.Aig_ManSetRegNum(g, 0)
		return g, a, x, y
	}
	a, ai, ax, ay := build()
	b, bi, bx, by := build()
	pairs := vec.Vec_IntAlloc(10)
	for _, pair := range [][2]aig.Edge{{a.Const1(), b.Const1()}, {ai, bi}, {a.Inputs()[1], b.Inputs()[1]}, {ax, bx}} {
		vec.Vec_IntPushTwo(pairs, int32(pair[0].NodeID()), int32(pair[1].NodeID()))
	}
	Ssw_MatchingStart(a, b, pairs)
	if Ssw_MatchingCountUnmached(a) != 1 {
		t.Fatal("incorrect unmatched count")
	}
	var frontier vec.Vec_Ptr_t
	Ssw_MatchingExtendOne(a, &frontier)
	if vec.Vec_PtrSize(&frontier) != 2 {
		t.Fatalf("frontier %v", vec.Vec_PtrArray(&frontier))
	}
	Ssw_MatchingExtend(a, b, 1, 0)
	if _, ok := sswIslandMatch(a, ax); ok {
		t.Fatal("matched neighbor was not removed")
	}
	if _, ok := sswIslandMatch(a, ai); !ok {
		t.Fatal("PI was removed")
	}
	Ssw_MatchingComplete(a, b)
	if match, ok := sswIslandMatch(a, ay); !ok || match != by {
		t.Fatalf("completion did not recover gate: %v %v", match, ok)
	}
	Ssw_MatchingComplete(b, a)
	if vec.Vec_IntSize(Ssw_MatchingPairs(a, b)) != 10 {
		t.Fatal("unexpected complete pairs")
	}
}

func TestIslandMatchingRejectsMissingInputs(t *testing.T) {
	a, b := aig.NewGraph(), aig.NewGraph()
	a.NewInput()
	b.NewInput()
	aig.Aig_ManSetRegNum(a, 0)
	aig.Aig_ManSetRegNum(b, 0)
	p := vec.Vec_IntAlloc(2)
	vec.Vec_IntPushTwo(p, 0, 0)
	defer func() {
		if recover() == nil {
			t.Fatal("accepted missing PI match")
		}
	}()
	Ssw_MatchingStart(a, b, p)
}

func TestIslandMiterRetainsMatchingAndBehavior(t *testing.T) {
	build := func() *aig.Graph {
		g := aig.NewGraph()
		pi, reg := g.NewInput(), g.NewInput()
		n := g.AND(pi, reg)
		g.AddOutput(n)
		g.AddOutput(pi)
		aig.Aig_ManSetRegNum(g, 1)
		return g
	}
	a, b := build(), build()
	pairs := Saig_StrSimPerformMatching_hack(a, b)
	Ssw_MatchingStart(a, b, pairs)
	all := Ssw_MatchingPairs(a, b)
	miter := sswIslandMiter(a, b)
	mapped := Ssw_MatchingMiter(miter, a, b, all)
	if vec.Vec_IntSize(mapped) != 4 {
		t.Fatalf("miter pair IDs: %v", vec.Vec_IntArray(mapped))
	}
	for mask := 0; mask < 8; mask++ {
		input, leftReg, rightReg := mask&1 != 0, mask&2 != 0, mask&4 != 0
		outputs, err := miter.Evaluate([]bool{input, leftReg, rightReg})
		if err != nil {
			t.Fatal(err)
		}
		if outputs[0] != ((input && leftReg) != (input && rightReg)) || outputs[1] != input || outputs[2] != input {
			t.Fatalf("miter mask %d outputs %v", mask, outputs)
		}
	}
}
