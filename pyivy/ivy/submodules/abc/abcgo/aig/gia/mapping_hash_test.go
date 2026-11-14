package gia

import "testing"

func TestDupHashMappingMergesIdenticalLuts(t *testing.T) {
	p := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x, y := Gia_ManAppendAnd(p, a, b), Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	Gia_ManAppendCo(p, y^1)
	p.SetIdentity("hash mapping", "spec")
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, id := range []int32{x >> 1, y >> 1} {
		Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
		for _, value := range []int32{2, a >> 1, b >> 1, id} {
			Vec_IntPush(p.VMapping, value)
		}
	}
	q := Gia_ManDupHashMapping(p)
	if Gia_ManAndNum(q) != 1 || Gia_ManLutNum(q) != 1 || Gia_ManCoNum(q) != 2 {
		t.Fatalf("ANDs %d, LUTs %d, COs %d", Gia_ManAndNum(q), Gia_ManLutNum(q), Gia_ManCoNum(q))
	}
	if name, spec := q.Identity(); name != "hash mapping" || spec != "spec" {
		t.Fatalf("identity %q %q", name, spec)
	}
	co0, co1 := Gia_ManCo(q, 0), Gia_ManCo(q, 1)
	if Gia_ObjFanin0(co0) != Gia_ObjFanin0(co1) || Gia_ObjFaninC0(co0) == Gia_ObjFaninC0(co1) {
		t.Fatal("output polarity not preserved")
	}
	fanins := Gia_ObjLutFanins(q, Gia_ObjFaninId0p(q, co0))
	if len(fanins) != 2 || fanins[0] != a>>1 || fanins[1] != b>>1 {
		t.Fatalf("mapped leaves %v", fanins)
	}
}
