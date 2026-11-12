package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestLutDelayTraceCriticalEdges(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	p.vMapping = vec.Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id   int32
		fans []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		vec.Vec_IntWriteEntry(p.vMapping, int(lut.id), int32(vec.Vec_IntSize(p.vMapping)))
		vec.Vec_IntPush(p.vMapping, int32(len(lut.fans)))
		for _, fan := range lut.fans {
			vec.Vec_IntPush(p.vMapping, fan)
		}
		vec.Vec_IntPush(p.vMapping, lut.id)
	}
	if delay := Gia_ManDelayTraceLut(p); delay != 2 {
		t.Fatalf("delay %v", delay)
	}
	defer Gia_ManTimeStop(p)
	if Gia_ObjTimeSlack(p, c>>1) != 1 || Gia_LutDelayTraceTCEdges(p, y>>1, 0.1) != 1 {
		t.Fatal("wrong slack or critical edge")
	}
}
