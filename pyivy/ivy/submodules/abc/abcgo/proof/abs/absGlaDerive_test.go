package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaDeriveSelectsMarkedAbstractedObjects(t *testing.T) {
	p := gia.Gia_ManStart(10)
	x := gia.Gia_ManAppendCi(p)
	q := gia.Gia_ManAppendCi(p)
	gate := gia.Gia_ManAppendAnd(p, x, q)
	gia.Gia_ManAppendCo(p, gate)
	gia.Gia_ManAppendCo(p, x)
	gia.Gia_ManSetRegNum(p, 1)
	classes := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p)))
	gia.Gia_ManSetGateClasses(p, classes)
	for _, lit := range []int32{q, gate} {
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(p, lit>>1), 1)
		vec.Vec_IntWriteEntry(classes, int(lit>>1), 1)
	}
	got := Ga2_ManAbsDerive(p)
	if vec.Vec_IntSize(got) != 2 || vec.Vec_IntEntry(got, 0) != q>>1 || vec.Vec_IntEntry(got, 1) != gate>>1 {
		t.Fatal("wrong selected objects")
	}
	gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(p, q>>1), 0)
	got = Ga2_ManAbsDerive(p)
	if vec.Vec_IntSize(got) != 1 || vec.Vec_IntEntry(got, 0) != gate>>1 {
		t.Fatal("unmarked register selected")
	}
}
