package gia

import "testing"

func TestStaticMappingFanouts(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	co := Gia_ManAppendCo(p, y)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id  int32
		ins []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		Vec_IntWriteEntry(p.VMapping, int(lut.id), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(len(lut.ins)))
		for _, id := range lut.ins {
			Vec_IntPush(p.VMapping, id)
		}
		Vec_IntPush(p.VMapping, lut.id)
	}
	p.pLutRefs = []int32{7}
	Gia_ManStaticMappingFanoutStart(p)
	for _, pair := range [][2]int32{{a >> 1, x >> 1}, {b >> 1, x >> 1}, {c >> 1, y >> 1}, {x >> 1, y >> 1}, {y >> 1, co >> 1}} {
		if Gia_ObjFanoutNumId(p, pair[0]) != 1 || Gia_ObjFanoutId(p, pair[0], 0) != pair[1] {
			t.Fatalf("incorrect fanouts for %d", pair[0])
		}
	}
	if p.pLutRefs[0] != 7 {
		t.Fatal("existing LUT references changed")
	}
	Gia_ManStaticFanoutStop(p)
	if p.vFanout != nil || p.vFanoutNums != nil {
		t.Fatal("fanouts retained")
	}
}
