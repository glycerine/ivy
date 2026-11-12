package gia

import (
	"reflect"
	"testing"
)

func TestCrossCutAndConeCollection(t *testing.T) {
	p := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	out := Gia_ManAppendCo(p, y)
	if got := Vec_IntArray(Gia_ManBfsForCrossCut(p)); !reflect.DeepEqual(got, []int32{x >> 1, y >> 1}) {
		t.Fatalf("level order %v", got)
	}
	nodes := Vec_IntAlloc(10)
	Gia_ManPrintCollect2_rec(p, Gia_ManObj(p, out>>1), nodes)
	if got := Vec_IntArray(nodes); !reflect.DeepEqual(got, []int32{c >> 1, a >> 1, b >> 1, x >> 1, y >> 1, out >> 1}) {
		t.Fatalf("cone %v", got)
	}
}
