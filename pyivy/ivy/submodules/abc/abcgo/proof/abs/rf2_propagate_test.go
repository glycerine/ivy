package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAlternateRefinementPropagate(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, c, d)
	z := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, z)
	p := Rf2_ManStart(g)
	defer Rf2_ManStop(p, 0)
	p.vMap = vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(p.vMap, x>>1)
	vec.Vec_IntPush(p.vMap, y>>1)
	p.pCex = util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(p.pCex.PData, 0)
	util.Abc_InfoSetBit(p.pCex.PData, 1)
	Rf2_ManCollect(p)
	Rf2_ManAssignJustIds(p)
	if got := vec.Vec_IntArray(Rf2_ManPropagate(p, 32)); !reflect.DeepEqual(got, []int32{3}) {
		t.Fatalf("masks=%v", got)
	}
	if got := vec.Vec_IntArray(Rf2_ManPropagate(p, 1)); len(got) != 0 {
		t.Fatalf("limited masks=%v", got)
	}
}

func TestAlternateRefinementBounds(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, c, d)
	z := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, z)
	p := Rf2_ManStart(g)
	defer Rf2_ManStop(p, 0)
	p.vMap = vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(p.vMap, x>>1)
	vec.Vec_IntPush(p.vMap, y>>1)
	p.pCex = util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(p.pCex.PData, 0)
	util.Abc_InfoSetBit(p.pCex.PData, 1)
	Rf2_ManCollect(p)
	p.nMapWords = 1
	Rf2_ManBounds(p)
	out := gia.Gia_ManPo(g, 0)
	if Rf2_ObjA(p, out)[0] != 0 || Rf2_ObjN(p, out)[0] != -1 {
		t.Fatal("incorrect intersection bounds")
	}
}

func TestAlternateRefinementEntryPoint(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, x)
	p := Rf2_ManStart(g)
	defer Rf2_ManStop(p, 0)
	mapping := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(mapping, x>>1)
	cex := util.Abc_CexAlloc(0, 1, 1)
	util.Abc_InfoSetBit(cex.PData, 0)
	if got := Rf2_ManRefine(p, cex, mapping, 0, 0); got != nil {
		t.Fatalf("experimental entry returned %v", vec.Vec_IntArray(got))
	}
}
