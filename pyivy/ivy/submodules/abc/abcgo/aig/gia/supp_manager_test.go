package gia

import (
	"reflect"
	"testing"
)

func TestMin2ManagerSupport(t *testing.T) {
	p := Gia_ManStart(16)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	m := Gia_Man2SuppStart(p)
	defer Gia_Man2SuppStop(m)
	m.iLits = [2]int32{x, y}
	if Gia_Min2ManGatherSupp(m) != 1 {
		t.Fatal("expected shared cone")
	}
	if !reflect.DeepEqual(Vec_IntArray(m.vCis[0]), []int32{a >> 1, b >> 1}) || !reflect.DeepEqual(Vec_IntArray(m.vCis[1]), []int32{c >> 1}) {
		t.Fatal("wrong support inputs")
	}
	m.iLits = [2]int32{a, c}
	if Gia_Min2ManGatherSupp(m) != 0 {
		t.Fatal("unexpected overlap")
	}
	if m.nSupps != 2 || m.iPattern != 1 || m.pSat.Options().LearntRatio != 0 {
		t.Fatal("wrong manager state")
	}
}
