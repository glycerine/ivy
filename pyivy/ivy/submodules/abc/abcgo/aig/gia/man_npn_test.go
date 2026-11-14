package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaNpnClassReport(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	mapping := vec.Vec_IntStart(int(g.nObjs))
	vec.Vec_IntWriteEntry(mapping, int(x>>1), int32(vec.Vec_IntSize(mapping)))
	for _, v := range []int32{2, a >> 1, b >> 1, x >> 1} {
		vec.Vec_IntPush(mapping, v)
	}
	g.VMapping = mapping
	Gia_ManPrintNpnClasses(g)
	if g.Refs != nil {
		t.Fatal("mapped references not released")
	}
}
