package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestClassManagerPreservesRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	aig.Aig_ManReprStart(g, 2)
	aig.Aig_ObjSetRepr(g, x, g.Const1())
	p := Ssw_ClassesStart(g)
	if Ssw_ClassesReadAig(p) != g || aig.Aig_ObjRepr(g, x) != g.Const1() {
		t.Fatal("existing representatives not preserved")
	}
	vec.Vec_PtrPush(Ssw_ClassesGetRefined(p), x)
	Ssw_ClassesClearRefined(p)
	if vec.Vec_PtrSize(Ssw_ClassesGetRefined(p)) != 0 {
		t.Fatal("refined set was not cleared")
	}
	Ssw_ClassesStop(p)
	if aig.Aig_ObjRepr(g, x) != g.Const1() {
		t.Fatal("stop changed graph representatives")
	}
}
func TestClassManagerStartsRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	p := Ssw_ClassesStart(g)
	if !aig.Aig_ManHasReprs(g) || Ssw_ClassesClassNum(p) != 0 || Ssw_ClassesCand1Num(p) != 0 || Ssw_ClassesLitNum(p) != 0 {
		t.Fatal("incorrect initial state")
	}
}
