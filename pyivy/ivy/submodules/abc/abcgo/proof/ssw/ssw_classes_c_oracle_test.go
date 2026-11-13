package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// The C fixture reports two distinct target candidates and one pair class.
func TestClassPreparationMatchesC(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b.Not())
	g.AddOutput(q)
	aig.Aig_ManSetRegNum(g, 1)
	classes := Ssw_ClassesPrepareTargets(g)
	if Ssw_ClassesCand1Num(classes) != 2 || Ssw_ClassesClassNum(classes) != 0 || Ssw_ObjIsConst1Cand(g, a) == 0 || Ssw_ObjIsConst1Cand(g, b) == 0 || Ssw_ObjIsConst1Cand(g, q) != 0 {
		t.Fatal("target candidates differ from compiled C")
	}
	Ssw_ClassesStop(classes)
	g = aig.NewGraph()
	a, b = g.NewInput(), g.NewInput()
	g.NewInput()
	pairs := vec.Vec_IntAllocArrayCopy([]int32{int32(a.NodeID()), int32(b.NodeID())}, 2)
	classes = Ssw_ClassesPreparePairsSimple(g, pairs)
	defer Ssw_ClassesStop(classes)
	if Ssw_ClassesClassNum(classes) != 1 || aig.Aig_ObjRepr(g, b) != a {
		t.Fatal("pair representative differs from compiled C")
	}
}

func TestClassTargetsDuplicateFaninAssertsLikeC(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	g.AddOutput(a)
	g.AddOutput(a.Not())
	defer func() {
		if recover() == nil {
			t.Fatal("C asserts on duplicate target fanin")
		}
	}()
	Ssw_ClassesPrepareTargets(g)
}
