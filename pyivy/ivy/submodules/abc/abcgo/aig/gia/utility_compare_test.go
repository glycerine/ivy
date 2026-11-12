package gia

import (
	"reflect"
	"testing"
)

func TestCollectObjectsLimitAndCompare(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	Gia_ManIncrementTravId(p)
	ids := Vec_IntAlloc(3)
	Gia_ManCollectObjs_rec(p, y>>1, ids, 3)
	if got := Vec_IntArray(ids); !reflect.DeepEqual(got, []int32{c >> 1, a >> 1, b >> 1}) {
		t.Fatalf("limited cone %v", got)
	}
	q := Gia_ManDup(p)
	if Gia_ManCompare(p, q) != 1 {
		t.Fatal("identical graphs differ")
	}
	Gia_ObjFlipFaninC0(Gia_ManPo(q, 0))
	if Gia_ManCompare(p, q) != 0 {
		t.Fatal("different output not detected")
	}
}
