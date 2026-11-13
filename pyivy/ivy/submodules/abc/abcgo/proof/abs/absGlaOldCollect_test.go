package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestOldGlaCollectBoundaryAndInternalOrder(t *testing.T) {
	g := gia.Gia_ManStart(10)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	and := gia.Gia_ManAppendAnd(g, x, y)
	output := gia.Gia_ManAppendCo(g, and)
	p := &Gla_Man_t{pGia: g, pObjs: make([]Gla_Obj_t, 4), nObjs: 4, vAbs: vec.Vec_IntAllocArrayCopy([]int32{3}, 1)}
	p.pObjs[1] = Gla_Obj_t{id: 1, iGiaObj: x >> 1, fPi: 1}
	p.pObjs[2] = Gla_Obj_t{id: 2, iGiaObj: y >> 1, fPi: 1}
	p.pObjs[3] = Gla_Obj_t{id: 3, iGiaObj: and >> 1, fAnd: 1, fAbs: 1, nFanins: 2, Fanins: [4]int32{1, 2}}
	pis, ppis, cos, nodes := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	Gla_ManCollect(p, pis, ppis, cos, nodes)
	if got := vec.Vec_IntArray(pis); !reflect.DeepEqual(got, []int32{x >> 1, y >> 1}) {
		t.Fatalf("PIs %v", got)
	}
	if vec.Vec_IntSize(ppis) != 0 || !reflect.DeepEqual(vec.Vec_IntArray(cos), []int32{output >> 1}) || !reflect.DeepEqual(vec.Vec_IntArray(nodes), []int32{and >> 1}) {
		t.Fatalf("PPIs %v COs %v nodes %v", vec.Vec_IntArray(ppis), vec.Vec_IntArray(cos), vec.Vec_IntArray(nodes))
	}
}
