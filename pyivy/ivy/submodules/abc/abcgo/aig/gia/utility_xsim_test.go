package gia

import (
	"reflect"
	"testing"
)

func TestPoUnknownFirstFrame(t *testing.T) {
	p := Gia_ManStart(8)
	a, r := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, a)
	Gia_ManAppendCo(p, r)
	Gia_ManAppendCo(p, 0)
	Gia_ManAppendCo(p, a)
	Gia_ManSetRegNum(p, 1)
	if got := Vec_IntArray(Gia_ManPoXSim(p, 4, 0)); !reflect.DeepEqual(got, []int32{0, 1, 4}) {
		t.Fatalf("first unknown frames %v", got)
	}
}
