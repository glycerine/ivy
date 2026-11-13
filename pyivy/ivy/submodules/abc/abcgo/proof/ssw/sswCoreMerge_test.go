package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestConstraintConeKeepsRepresentativesWhenFullMergeEnabled(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(pi)
	g.AddOutput(lo)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	aig.Aig_ManReprStart(g, int32(len(g.Nodes())))
	aig.Aig_ObjSetRepr(g, pi, g.Const1())
	p := &Ssw_Man_t{pAig: g, pPars: &Ssw_Pars_t{fConstrs: 1, fMergeFull: 1}}
	Ssw_ManUpdateEquivs(p, g, 0)
	if aig.Aig_ObjRepr(g, pi) != g.Const1() || p.nEquivsTotal != 1 || p.nEquivsConstr != 0 {
		t.Fatal("full merge must keep constraint-cone representative")
	}
}
