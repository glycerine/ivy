package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRehashRefinesHashCollisions(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	p := Ssw_ClassesStart(g)
	p.pFuncNodeIsConst = func(_ any, x aig.Edge) int32 {
		if x == d {
			return 1
		}
		return 0
	}
	p.pFuncNodeHash = func(any, aig.Edge) uint32 { return 0 }
	p.pFuncNodesAreEqual = func(_ any, x, y aig.Edge) int32 {
		if (x == c) == (y == c) {
			return 1
		}
		return 0
	}
	nodes := vec.Vec_PtrAlloc(4)
	for _, x := range []aig.Edge{a, b, c, d} {
		vec.Vec_PtrPush(nodes, x)
	}
	if Ssw_ClassesPrepareRehash(p, nodes, 0) != 1 || p.nCands1 != 1 || p.nClasses != 1 || p.nLits != 1 {
		t.Fatal("wrong class stats")
	}
	if aig.Aig_ObjRepr(g, b) != a || aig.Aig_ObjRepr(g, c) != aig.InvalidEdge || aig.Aig_ObjRepr(g, d) != g.Const1() {
		t.Fatal("wrong representatives")
	}
}
