package util

import "testing"

func TestMemoryRecorderLifecycle(t *testing.T) {
	Util_MemRecStart()
	defer Util_MemRecQuit()
	if !Util_MemRecIsSet() {
		t.Fatal("not started")
	}
	for _, id := range []AbcPointerValue{5, 2, 9} {
		Util_MemRecAlloc(id)
	}
	Util_MemRecFree(2)
	Util_MemRecRecycle()
	if len(s_vAllocs.pArray) != 0 || len(s_vFrees.pArray) != 0 {
		t.Fatal("not recycled")
	}
}
func TestMemoryRecorderDifference(t *testing.T) {
	a, b := Vec_MemAlloc(0), Vec_MemAlloc(0)
	for _, id := range []AbcPointerValue{1, 2, 3, 5} {
		Vec_MemPush(a, id)
	}
	Vec_MemPush(b, 2)
	Vec_MemPush(b, 5)
	r := Vec_MemTwoMerge(a, b)
	if len(r.pArray) != 2 || r.pArray[0] != 1 || r.pArray[1] != 3 {
		t.Fatal(r.pArray)
	}
}
