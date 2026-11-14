package gia

import "testing"

func TestObjectReadAccessorsMatchStoredBits(t *testing.T) {
	p := Gia_ManStart(1)
	obj := Gia_ManObj(p, 0)

	// These are the same stored bits exposed by Gia_ObjPhase and the C
	// Gia_Obj_t fMark0/fMark1 fields. Read accessors must not normalize values.
	obj.fPhase, obj.fMark0, obj.fMark1 = 1, 1, 0
	if got := Gia_ObjPhase(obj); got != 1 {
		t.Fatalf("phase = %d, want 1", got)
	}
	if got := Gia_ObjMark0(obj); got != 1 {
		t.Fatalf("mark0 = %d, want 1", got)
	}
	if got := Gia_ObjMark1(obj); got != 0 {
		t.Fatalf("mark1 = %d, want 0", got)
	}
	obj.fMark0, obj.fMark1 = 0, 1
	if Gia_ObjMark0(obj) != 0 || Gia_ObjMark1(obj) != 1 {
		t.Fatal("mark accessors did not return the current stored bits")
	}
}
