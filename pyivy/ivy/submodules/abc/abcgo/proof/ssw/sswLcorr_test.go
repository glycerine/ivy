package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestLatchSweepRefinesDifferentTransitions(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	l0, l1 := g.NewInput(), g.NewInput()
	g.AddOutput(l0)
	g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 2)
	pars := &Ssw_Pars_t{NFramesK: 1, NBTLimit: 1000}
	p := Ssw_ManCreate(g, pars)
	p.pMSat = Ssw_SatStart(0)
	p.ppClasses = Ssw_ClassesStart(g)
	p.pSml = Ssw_SmlStart(g, 0, 2, 1)
	p.ppClasses.pManData = p.pSml
	p.ppClasses.pFuncNodeIsConst = func(data any, obj aig.Edge) int32 { return Ssw_SmlObjIsConstWord(data.(*Ssw_Sml_t), obj) }
	p.ppClasses.pFuncNodesAreEqual = func(data any, x, y aig.Edge) int32 { return Ssw_SmlObjsAreEqualWord(data.(*Ssw_Sml_t), x, y) }
	Ssw_ObjAddClass(p.ppClasses, l0, []aig.Edge{l0, l1}, 2)
	aig.Aig_ObjSetRepr(g, l1, l0)
	if Ssw_ManSweepLatch(p) != 1 || aig.Aig_ObjRepr(g, l1) != aig.InvalidEdge {
		t.Fatal("different transitions should refine latch class")
	}
}

func TestLatchSweepKeepsIdenticalTransitions(t *testing.T) {
	g := aig.NewGraph()
	pi := g.NewInput()
	l0, l1 := g.NewInput(), g.NewInput()
	g.AddOutput(l0)
	g.AddOutput(pi)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 2)
	p := Ssw_ManCreate(g, &Ssw_Pars_t{NFramesK: 1})
	p.pMSat = Ssw_SatStart(0)
	p.ppClasses = Ssw_ClassesStart(g)
	p.pSml = Ssw_SmlStart(g, 0, 2, 1)
	Ssw_ObjAddClass(p.ppClasses, l0, []aig.Edge{l0, l1}, 2)
	aig.Aig_ObjSetRepr(g, l1, l0)
	if Ssw_ManSweepLatch(p) != 0 || aig.Aig_ObjRepr(g, l1) != l0 {
		t.Fatal("identical transitions should remain candidates")
	}
}
