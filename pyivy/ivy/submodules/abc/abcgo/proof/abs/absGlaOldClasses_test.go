package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldGlaCollectPseudoInputs(t *testing.T) {
	p := &Gla_Man_t{nObjs: 5, pObjs: make([]Gla_Obj_t, 5), vAbs: vec.Vec_IntAlloc(2)}
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	p.pObjs[1].fPi = 1
	p.pObjs[2].fAnd = 1
	p.pObjs[3].fAnd = 1
	p.pObjs[3].fAbs = 1
	p.pObjs[3].nFanins = 2
	p.pObjs[3].Fanins = [4]int32{1, 2}
	p.pObjs[4].fAnd = 1
	p.pObjs[4].fAbs = 1
	p.pObjs[4].nFanins = 1
	p.pObjs[4].Fanins[0] = 2
	vec.Vec_IntPush(p.vAbs, 3)
	vec.Vec_IntPush(p.vAbs, 4)
	pis := vec.Vec_IntAlloc(0)
	if got := vec.Vec_IntArray(Gla_ManCollectPPis(p, pis)); !reflect.DeepEqual(got, []int32{2}) {
		t.Fatalf("PPIs=%v", got)
	}
	if got := vec.Vec_IntArray(pis); !reflect.DeepEqual(got, []int32{1}) {
		t.Fatalf("PIs=%v", got)
	}
	if Gia_GlaAbsCount(p, 0, 1) != 2 || Gla_ManCountPPis(p) != 1 {
		t.Fatal("incorrect counts")
	}
}
func TestOldGlaTranslateCutCone(t *testing.T) {
	g := gia.Gia_ManStart(12)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	y := gia.Gia_ManAppendAnd(g, x, c)
	v := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	gia.Gia_ManIncrementTravId(g)
	gia.Gia_ObjSetTravIdCurrent(g, gia.Gia_ManObj(g, a>>1))
	if Gla_ManTranslate_rec(g, gia.Gia_ManObj(g, y>>1), v, 3) != 1 || vec.Vec_IntEntry(v, int(x>>1)) != 3 || vec.Vec_IntEntry(v, int(y>>1)) != 3 {
		t.Fatalf("classes=%v", vec.Vec_IntArray(v))
	}
}
