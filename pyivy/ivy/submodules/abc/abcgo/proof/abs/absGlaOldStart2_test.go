package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldGlaStart2BuildsCutFanins(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	vec.Vec_IntWriteEntry(classes, 0, 1)
	vec.Vec_IntWriteEntry(classes, int(gate>>1), 1)
	gia.Gia_ManSetGateClasses(g, classes)
	var pars Abs_Par_t
	Abs_ParSetDefaults(&pars)
	p := Gla_ManStart2(g, &pars)
	if p.pObjRoot == nil || p.pObjRoot.nFanins != 1 || Gia_GlaAbsCount(p, 0, 1) != 1 {
		t.Fatal("invalid mapped manager")
	}
	Gia_GlaAddTimeFrame(p, 0)
	if Gla_ManGetOutLit(p, 0) < 0 {
		t.Fatal("missing SAT output literal")
	}
}
