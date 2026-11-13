package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestConstrainedSimulationRefinesClasses(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	classes := Ssw_ClassesStart(g)
	classes.pManData = g
	classes.pFuncNodeIsConst = Ssw_SmlObjIsConstBit
	classes.pFuncNodesAreEqual = Ssw_SmlObjsAreEqualBit
	Ssw_ObjAddClass(classes, pi, []aig.Edge{pi, lo}, 2)
	aig.Aig_ObjSetRepr(g, lo, pi)
	init := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(init, 0)
	vec.Vec_IntPush(init, 1)
	p := &Ssw_Man_t{pAig: g, ppClasses: classes, vInits: init}
	Ssw_ManRefineByConstrSim(p)
	if classes.nClasses != 0 || aig.Aig_ObjRepr(g, lo) != aig.InvalidEdge {
		t.Fatal("class not refined")
	}
}
