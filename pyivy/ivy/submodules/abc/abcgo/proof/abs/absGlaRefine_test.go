package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaRefineAddsLayerOfPseudoInputs(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	c := gia.Gia_ManAppendCi(g)
	inner := gia.Gia_ManAppendAnd(g, a, b)
	root := gia.Gia_ManAppendAnd(g, inner, c)
	gia.Gia_ManAppendCo(g, root)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.FUseSimple = 1
	pars.FAddLayer = 1
	p := Ga2_ManStart(g, pars)
	defer Ga2_ManStop(p)
	Ga2_ManSetupNode(p, gia.Gia_ManObj(g, root>>1), 1)
	Ga2_ManSetupNode(p, gia.Gia_ManObj(g, inner>>1), 0)
	Ga2_ManSetupNode(p, gia.Gia_ManObj(g, c>>1), 0)
	got := Ga2_ManRefine(p)
	if vec.Vec_IntSize(got) != 1 || vec.Vec_IntEntry(got, 0) != inner>>1 || p.nObjAdded != 1 {
		t.Fatal("wrong selected PPI layer")
	}
}
