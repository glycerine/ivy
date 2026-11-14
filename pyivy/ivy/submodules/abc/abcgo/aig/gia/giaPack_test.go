package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestGiaLutCollectAndPacking(t *testing.T) {
	g := Gia_ManStart(10)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	out := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, out)
	var pars Jf_Par_t
	ofDefaultParameters(&pars)
	pars.NLutSize = 2
	ofMap(g, &pars)
	want := []int32{x >> 1, out >> 1}
	if got := vec.Vec_IntArray(Gia_ManLutCollect(g)); !reflect.DeepEqual(got, want) {
		t.Fatalf("distance order %v", got)
	}
	if got := vec.Vec_IntArray(Gia_ManLutCollect2(g)); !reflect.DeepEqual(got, want) {
		t.Fatalf("breadth order %v", got)
	}
	Gia_ManLutPacking(g, 2, 10, 2, 0)
	if got := vec.Vec_IntArray(g.vPacking); !reflect.DeepEqual(got, []int32{1, 2, out >> 1, x >> 1}) {
		t.Fatalf("packing %v", got)
	}
}

func TestGiaLutPackingZeroInput(t *testing.T) {
	g := Gia_ManStart(6)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	out := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, out)
	g.vMapping = vec.Vec_IntStart(int(g.nObjs))
	vec.Vec_IntWriteEntry(g.vMapping, int(out>>1), int32(vec.Vec_IntSize(g.vMapping)))
	vec.Vec_IntPush(g.vMapping, 0)
	vec.Vec_IntPush(g.vMapping, out>>1)
	Gia_ManLutPacking(g, 2, 10, 2, 0)
	if got := vec.Vec_IntArray(g.vPacking); !reflect.DeepEqual(got, []int32{1, 1, out >> 1}) {
		t.Fatalf("packing %v", got)
	}
}
