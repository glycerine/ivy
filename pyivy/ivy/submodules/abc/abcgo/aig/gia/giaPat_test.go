package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestGiaSatPatternConeAndCleanup(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	co := Gia_ManAppendCo(g, x)
	root := Gia_ManObj(g, co>>1)
	var visit, cex Vec_Int_t
	Gia_SatCollectCone(g, Gia_ObjFanin0(root), &visit)
	if got := vec.Vec_IntArray(&visit); !reflect.DeepEqual(got, []int32{a >> 1, b >> 1, x >> 1}) {
		t.Fatalf("cone %v", got)
	}
	for _, id := range vec.Vec_IntArray(&visit) {
		Sat_ObjSetXValue(Gia_ManObj(g, id), 0)
	}
	vec.Vec_IntPush(&cex, 0)
	vec.Vec_IntPush(&cex, 2)
	vec.Vec_IntPush(&cex, 4)
	Gia_SatVerifyPattern(g, root, &cex, &visit)
	for _, obj := range g.pObjs {
		if Sat_ObjXValue(obj) != 0 {
			t.Fatalf("unclean mark at %d", obj.id)
		}
	}
}
