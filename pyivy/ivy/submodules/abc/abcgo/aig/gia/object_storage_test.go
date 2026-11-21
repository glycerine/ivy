package gia

import (
	"testing"
	"unsafe"
)

func TestGiaObjectStorageFitsTwoPointersAnd24Bytes(t *testing.T) {
	var obj Gia_Obj_t
	if size, limit := unsafe.Sizeof(obj), 2*unsafe.Sizeof(&obj)+24; size > limit {
		t.Fatalf("object takes %d bytes, want at most %d", size, limit)
	}
}

func TestCompactObjectKeepsWideBackwardEdgesAndScratchValue(t *testing.T) {
	p := Gia_ManStart(2)
	first := Gia_ManAppendCi(p)
	for range 512 {
		Gia_ManAppendCi(p)
	}
	last := Gia_ManCiLit(p, 512)
	for phase := range 4 {
		a, b := first^int32(phase&1), last^int32(phase>>1)
		obj := Gia_ManObj(p, Gia_ManAppendAnd(p, a, b)>>1)
		Gia_ObjSetValue(obj, -1234567)
		Gia_ObjSetMark0(obj, 5)
		Gia_ObjSetMark1(obj, 2)
		Gia_ObjSetPhaseValue(obj, -3)
		if Gia_ObjId(p, obj) <= 512 || Gia_ObjFaninLit0(obj, obj.id) != a || Gia_ObjFaninLit1(obj, obj.id) != b ||
			Gia_ObjValue(obj) != -1234567 || Gia_ObjMark0(obj) != 1 || Gia_ObjMark1(obj) != 0 || Gia_ObjPhase(obj) != 1 || Gia_Regular(Gia_Not(obj)) != obj {
			t.Fatalf("lost object field for phase %d", phase)
		}
	}
}
