package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaManagerStartInitializesCutAndConstant(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	p := Ga2_ManStart(g, pars)
	defer Rnm_ManStop(p.pRnm, 0)
	if p.nMarked != 1 || p.nTable <= 0 || len(p.pSops) != 65536 || vec.Vec_IntSize(p.vAbs) != 1 || vec.Vec_IntEntry(p.vAbs, 0) != 0 || Ga2_ObjId(p, gia.Gia_ManConst0(g)) != 0 {
		t.Fatal("bad manager initialization")
	}
}

func TestGlaManagerStopRestoresScratchState(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a^1, b)
	gia.Gia_ManAppendCo(g, gate)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	p := Ga2_ManStart(g, pars)
	if gia.Gia_ManMapping(g) == nil || !gia.Gia_ManHasStaticFanout(g) {
		t.Fatal("expected temporary manager state")
	}
	Ga2_ManStop(p)
	if gia.Gia_ManMapping(g) != nil || gia.Gia_ManHasStaticFanout(g) {
		t.Fatal("manager scratch state was not cleaned")
	}
}
