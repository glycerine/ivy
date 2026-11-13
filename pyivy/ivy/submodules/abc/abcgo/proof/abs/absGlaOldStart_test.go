package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldGlaStartBuildsMappedManager(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(gate>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	p := Gla_ManStart(g, &pars)
	if p.pObjRoot == nil || p.pObjRoot.fPo == 0 || Gia_GlaAbsCount(p, 0, 1) != 1 || p.pCnf == nil {
		t.Fatal("invalid mapped manager")
	}
	Gia_GlaAddTimeFrame(p, 0)
	if Gla_ManGetOutLit(p, 0) < 0 {
		t.Fatal("missing output literal")
	}
	Rnm_ManStop(p.pRnm, 0)
}
