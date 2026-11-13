package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestConstraintConeRemovesOnlyIntersectingRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	prop := g.AND(pi, lo)
	g.AddOutput(prop)
	g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	aig.Aig_ManReprStart(g, int32(len(g.Nodes())))
	aig.Aig_ObjSetRepr(g, pi, g.Const1())
	aig.Aig_ObjSetRepr(g, prop, g.Const1())
	p := &Ssw_Man_t{pAig: g, pPars: &Ssw_Pars_t{fConstrs: 1}}
	Ssw_ManUpdateEquivs(p, g, 0)
	if aig.Aig_ObjRepr(g, pi) != aig.InvalidEdge || aig.Aig_ObjRepr(g, prop) != g.Const1() {
		t.Fatal("incorrect constraint cone representative removal")
	}
	if p.nEquivsTotal != 2 || p.nEquivsConstr != 1 {
		t.Fatalf("incorrect stats %d %d", p.nEquivsTotal, p.nEquivsConstr)
	}
}
