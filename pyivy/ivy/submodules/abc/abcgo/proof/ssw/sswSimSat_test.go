package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestCounterexampleBitResimulationSplitsClass(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := &Ssw_Man_t{pAig: g, pPatWords: []uint32{2}}
	classes := Ssw_ClassesStart(g)
	p.ppClasses = classes
	classes.pManData = g
	classes.pFuncNodeIsConst = Ssw_SmlObjIsConstBit
	classes.pFuncNodesAreEqual = Ssw_SmlObjsAreEqualBit
	Ssw_ObjAddClass(classes, a, []aig.Edge{a, b}, 2)
	aig.Aig_ObjSetRepr(g, b, a)
	Ssw_ManResimulateBit(p, b, a)
	if classes.nClasses != 0 || aig.Aig_ObjRepr(g, b) != aig.InvalidEdge {
		t.Fatal("counterexample failed to split class")
	}
}
