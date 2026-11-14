package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestOldGlaSelectChoosesLowerPriorityZero(t *testing.T) {
	for _, prop := range []int32{0, 1} {
		g := gia.Gia_ManStart(8)
		x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		and := gia.Gia_ManAppendAnd(g, x, y)
		gia.Gia_ManAppendCo(g, and)
		p := &Gla_Man_t{pGia: g, pPars: &Abs_Par_t{FPropFanout: prop}, pvRefis: make([][]Rfn_Obj_t, gia.Gia_ManObjNum(g)), vObjCounts: vec.Vec_IntStart(1)}
		xo, yo, root := gia.Gia_ManObj(g, x>>1), gia.Gia_ManObj(g, y>>1), gia.Gia_ManObj(g, and>>1)
		*Gla_ObjRef(p, xo, 0) = Rfn_Obj_t{fPPi: 1, Prio: 1, Sign: 1}
		*Gla_ObjRef(p, yo, 0) = Rfn_Obj_t{fPPi: 1, Prio: 2, Sign: 1}
		*Gla_ObjRef(p, root, 0) = Rfn_Obj_t{Prio: 1, Sign: 1}
		if prop != 0 {
			gia.Gia_ManStaticFanoutStart(g)
		}
		selected := vec.Vec_IntAlloc(0)
		Gla_ManRefSelect_rec(p, root, 0, selected, 1)
		if got := vec.Vec_IntArray(selected); !reflect.DeepEqual(got, []int32{x >> 1}) {
			t.Fatalf("prop=%d selected %v", prop, got)
		}
	}
}

func TestOldGlaFanoutPropagationMarksControllingPPi(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	start := gia.Gia_ManAppendAnd(g, x, y)
	ppi := gia.Gia_ManAppendAnd(g, start, z)
	gia.Gia_ManAppendCo(g, ppi)
	gia.Gia_ManStaticFanoutStart(g)
	defer gia.Gia_ManStaticFanoutStop(g)

	p := &Gla_Man_t{
		pGia:       g,
		pPars:      &Abs_Par_t{FPropFanout: 1},
		pvRefis:    make([][]Rfn_Obj_t, gia.Gia_ManObjNum(g)),
		vObjCounts: vec.Vec_IntStart(1),
	}
	startObj := gia.Gia_ManObj(g, start>>1)
	ppiObj := gia.Gia_ManObj(g, ppi>>1)
	*Gla_ObjRef(p, startObj, 0) = Rfn_Obj_t{Value: 0, Sign: 1}
	*Gla_ObjRef(p, ppiObj, 0) = Rfn_Obj_t{fPPi: 1, Prio: 2, Sign: 1}

	selected := vec.Vec_IntAlloc(0)
	Gia_ManRefSetAndPropFanout_rec(p, startObj, 0, selected, 1)
	if got, want := vec.Vec_IntArray(selected), []int32{ppi >> 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("selected pseudo-inputs %v, want %v", got, want)
	}
	if Gla_ObjRef(p, startObj, 0).fVisit == 0 || Gla_ObjRef(p, ppiObj, 0).fVisit == 0 {
		t.Fatal("fanout propagation did not mark its controlling cone")
	}
}
