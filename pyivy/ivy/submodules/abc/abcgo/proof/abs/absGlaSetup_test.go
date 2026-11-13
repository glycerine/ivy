package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaSetupNodeBuildsCutCovers(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	p := Ga2_ManStart(g, pars)
	defer Ga2_ManStop(p)
	obj := gia.Gia_ManObj(g, gate>>1)
	Ga2_ManSetupNode(p, obj, 0)
	if !Ga2_ObjIsLeaf(p, obj) || vec.Vec_IntSize(p.vAbs) != 1 {
		t.Fatal("leaf setup")
	}
	Ga2_ManSetupNode(p, obj, 1)
	if !Ga2_ObjIsAbs(p, obj) || vec.Vec_IntSize(p.vAbs) != 2 || vec.Vec_IntSize(Ga2_ObjCnf0(p, obj)) != 1 || vec.Vec_IntSize(Ga2_ObjCnf1(p, obj)) != 2 {
		t.Fatal("bad AND covers")
	}
}
