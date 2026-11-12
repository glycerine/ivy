package gia

import (
	"reflect"
	"testing"
)

func TestStructuralUnateness(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, _ := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b^1))
	Gia_ManAppendCo(p, Gia_ManAppendXor(p, a, b))
	Gia_ManAppendCo(p, 1)
	want := [][]int32{{2, 1, 3}, {0, 0, 3}, {3, 3, 3}}
	v := Gia_ManCheckUnateVec(p, nil, nil)
	for o := range want {
		for i, w := range want[o] {
			if got := Gia_ManCheckUnate(p, int32(i), int32(o)); got != w {
				t.Fatalf("output %d input %d: %d", o, i, got)
			}
		}
	}
	if got := Vec_IntArray(Vec_WecEntry(v, 0)); !reflect.DeepEqual(got, []int32{0, 3}) {
		t.Fatalf("first support %v", got)
	}
	if got := Vec_IntArray(Vec_WecEntry(v, 1)); !reflect.DeepEqual(got, []int32{0, 1, 2, 3}) {
		t.Fatalf("xor support %v", got)
	}
}
