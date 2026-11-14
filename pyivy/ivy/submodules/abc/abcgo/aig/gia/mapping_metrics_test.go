package gia

import "testing"

func TestMappingMetrics(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	co := Gia_ManAppendCo(p, y)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, lut := range []struct {
		id int32
		in []int32
	}{{x >> 1, []int32{a >> 1, b >> 1}}, {y >> 1, []int32{x >> 1, c >> 1}}} {
		Vec_IntWriteEntry(p.VMapping, int(lut.id), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(len(lut.in)))
		for _, id := range lut.in {
			Vec_IntPush(p.VMapping, id)
		}
		Vec_IntPush(p.VMapping, lut.id)
	}
	var levels []int32
	if Gia_ManLutNum(p) != 2 || Gia_ManLutFaninCount(p) != 4 || Gia_ManLutLevel(p, &levels) != 2 || levels[co>>1] != 2 {
		t.Fatalf("bad mapping metrics: %v", levels)
	}
	Gia_ManSetRefsMapped(p)
	Gia_ManSetLutRefs(p)
	if p.Refs[x>>1] != 1 || p.Refs[y>>1] != 1 || p.pLutRefs[x>>1] != 1 || p.pLutRefs[y>>1] != 1 {
		t.Fatal("bad mapped refs")
	}
	if Gia_ManComputeOverlap(p) != 0 || Gia_ManComputeOverlap2(p) != 0 {
		t.Fatal("bad overlap")
	}
}
