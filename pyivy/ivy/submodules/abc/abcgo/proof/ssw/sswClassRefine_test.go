package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestRecursiveClassRefinement(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d, e := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	p := Ssw_ClassesStart(g)
	group := map[aig.Edge]int{a: 0, b: 1, c: 0, d: 2, e: 1}
	p.pFuncNodesAreEqual = func(_ any, x, y aig.Edge) int32 {
		if group[x] == group[y] {
			return 1
		}
		return 0
	}
	Ssw_ObjAddClass(p, a, []aig.Edge{a, b, c, d, e}, 5)
	for _, obj := range []aig.Edge{b, c, d, e} {
		aig.Aig_ObjSetRepr(g, obj, a)
	}
	if got := Ssw_ClassesRefine(p, 1); got != 2 {
		t.Fatalf("splits %d", got)
	}
	if p.nClasses != 2 || p.nLits != 2 {
		t.Fatalf("class stats %d %d", p.nClasses, p.nLits)
	}
	if aig.Aig_ObjRepr(g, c) != a || aig.Aig_ObjRepr(g, e) != b || aig.Aig_ObjRepr(g, d) != aig.InvalidEdge {
		t.Fatal("wrong representatives")
	}
}

func TestRefineConstantCandidates(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	p := Ssw_ClassesStart(g)
	for _, x := range []aig.Edge{a, b, c} {
		aig.Aig_ObjSetRepr(g, x, g.Const1())
	}
	p.nCands1 = 3
	p.pFuncNodeIsConst = func(_ any, x aig.Edge) int32 {
		if x == a {
			return 1
		}
		return 0
	}
	if Ssw_ClassesRefineConst1(p, 0) != 1 || p.nCands1 != 1 || p.nClasses != 1 || aig.Aig_ObjRepr(g, c) != b || aig.Aig_ObjRepr(g, b) != aig.InvalidEdge {
		t.Fatal("constant candidates incorrectly refined")
	}
}

func TestConstantOnlyRefinementMaintainsCandidateCount(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := Ssw_ClassesStart(g)
	p.fConstCorr = 1
	p.nCands1 = 2
	aig.Aig_ObjSetRepr(g, a, g.Const1())
	aig.Aig_ObjSetRepr(g, b, g.Const1())
	p.pFuncNodeIsConst = func(_ any, obj aig.Edge) int32 {
		if obj == a {
			return 1
		}
		return 0
	}
	if Ssw_ClassesRefineConst1(p, 0) != 1 {
		t.Fatal("expected refinement")
	}
	if p.nCands1 != 1 || aig.Aig_ObjRepr(g, b) != aig.InvalidEdge {
		t.Fatalf("wrong candidate count %d", p.nCands1)
	}
	Ssw_ClassesCheck(p)
}
