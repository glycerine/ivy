package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaTranslateIncludesInternalCutNodes(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	c := gia.Gia_ManAppendCi(g)
	inner := gia.Gia_ManAppendAnd(g, a, b)
	root := gia.Gia_ManAppendAnd(g, inner, c)
	gia.Gia_ManAppendCo(g, root)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	p := Ga2_ManStart(g, pars)
	defer Ga2_ManStop(p)
	vec.Vec_IntPush(p.vAbs, root>>1)
	classes := Ga2_ManAbsTranslate(p)
	if vec.Vec_IntEntry(classes, int(root>>1)) != 1 || vec.Vec_IntEntry(classes, int(inner>>1)) != 1 || vec.Vec_IntEntry(classes, int(a>>1)) != 0 {
		t.Fatal("cut expansion incorrect")
	}
}
