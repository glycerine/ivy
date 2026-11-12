package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestGiaGateAbstractionCollection(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	classes := vec.Vec_IntStart(int(g.nObjs))
	vec.Vec_IntWriteEntry(classes, int(x>>1), 1)
	var pis, ppis, flops, nodes *Vec_Int_t
	Gia_ManGlaCollect(g, classes, &pis, &ppis, &flops, &nodes)
	if !reflect.DeepEqual(vec.Vec_IntArray(pis), []int32{a >> 1, b >> 1}) || !reflect.DeepEqual(vec.Vec_IntArray(nodes), []int32{x >> 1}) || vec.Vec_IntSize(ppis) != 0 || vec.Vec_IntSize(flops) != 0 {
		t.Fatal("bad abstraction")
	}
}
