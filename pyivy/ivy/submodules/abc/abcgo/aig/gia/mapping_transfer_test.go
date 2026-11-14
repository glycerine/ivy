package gia

import (
	"reflect"
	"testing"
)

func TestMappingAndPackingTransfer(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	Gia_ManAppendCo(p, y)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id     int32
		fanins []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		Vec_IntWriteEntry(p.VMapping, int(lut.id), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(len(lut.fanins)))
		for _, id := range lut.fanins {
			Vec_IntPush(p.VMapping, id)
		}
		Vec_IntPush(p.VMapping, lut.id)
	}
	p.VPacking = Vec_IntAlloc(4)
	for _, v := range []int32{1, 2, x >> 1, y >> 1} {
		Vec_IntPush(p.VPacking, v)
	}
	q := Gia_ManDup(p)
	Gia_ManTransferMapping(q, p)
	Gia_ManTransferPacking(q, p)
	if !reflect.DeepEqual(Vec_IntArray(q.VMapping), Vec_IntArray(p.VMapping)) || !reflect.DeepEqual(Vec_IntArray(q.VPacking), Vec_IntArray(p.VPacking)) {
		t.Fatalf("mapping or packing mismatch")
	}
	if Gia_ManLutNum(q) != 2 || Gia_ManLutLevel(q, nil) != 2 {
		t.Fatal("bad transferred metrics")
	}
}
