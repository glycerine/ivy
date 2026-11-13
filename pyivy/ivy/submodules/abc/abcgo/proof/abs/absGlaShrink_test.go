package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaShrinkDropsAddedObjectsAndLiterals(t *testing.T) {
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
	Ga2_ManSetupNode(p, obj, 1)
	Ga2_ManSetupNode(p, gia.Gia_ManObj(g, a>>1), 0)
	Ga2_ManSetupNode(p, gia.Gia_ManObj(g, b>>1), 0)
	p.LimAbs = 1
	frame := vec.Vec_IntStartFull(vec.Vec_IntSize(p.vValues))
	p.vId2Lit = append(p.vId2Lit, frame)
	vec.Vec_IntWriteEntry(frame, 1, 10)
	vec.Vec_IntWriteEntry(frame, 2, 8)
	p.nSatVars = 7
	Ga2_ManShrinkAbs(p, 2, 3, 3)
	if vec.Vec_IntSize(p.vValues) != 3 || Ga2_ObjId(p, gia.Gia_ManObj(g, b>>1)) != -1 || vec.Vec_IntEntry(frame, 1) != -1 || vec.Vec_IntEntry(frame, 2) != -1 || p.nSatVars != 3 {
		t.Fatal("bad rollback")
	}
	Ga2_ManShrinkAbs(p, 1, 1, 1)
	if vec.Vec_IntSize(frame) != 0 || Ga2_ObjId(p, obj) != -1 || len(p.vCnfs) != 2 {
		t.Fatal("constant-only rollback")
	}
}
