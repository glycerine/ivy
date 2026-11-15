package gia

import "testing"

func TestReplaceCutRetainsUncutOutputLogic(t *testing.T) {
	src := Gia_ManStart(1)
	a, b, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, b)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, b))
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	out := Gia_ManDupReplaceCut(src)
	if Gia_ManCiNum(out) != 3 || Gia_ManAndNum(out) != 1 {
		t.Fatal("uncut PI-only output became an independent free input")
	}
	for mask := 0; mask < 8; mask++ {
		a, b, r := mask&1 != 0, mask&2 != 0, mask&4 != 0
		checkQuantificationOutputs(t, out, []bool{a, b, r}, []bool{a, b, a && b, r})
	}
}

func TestReplaceCutPreservesConstraintRoles(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	if out := Gia_ManDupReplaceCut(src); out.nConstrs != 1 {
		t.Fatal("cut replacement reclassified an unchanged constraint output")
	}
}

func TestReplaceCutPreservesSimpleMode(t *testing.T) {
	src := Gia_ManStart(1)
	src.fGiaSimple = 1
	r := Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, Gia_ManAppendAnd(src, r, r^1))
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	out := Gia_ManDupReplaceCut(src)
	if out.fGiaSimple != 1 {
		t.Fatal("cut replacement lost simple construction mode")
	}
	for _, r := range []bool{false, true} {
		checkQuantificationOutputs(t, out, []bool{r}, []bool{false, r})
	}
}

func BenchmarkHighLightFlopLogic(b *testing.B) {
	src := Gia_ManStart(1)
	a, c, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	root := Gia_ManAppendAnd(src, Gia_ManAppendAnd(src, a, c), r)
	for i := 0; i < 256; i++ {
		root = Gia_ManAppendAnd(src, root, r^int32(i&1))
	}
	Gia_ManAppendCo(src, root)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Gia_ManHighLightFlopLogic(src)
	}
}
