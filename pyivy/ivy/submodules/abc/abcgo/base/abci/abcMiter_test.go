package abci

import "testing"

func TestVecObjPtrArrayPreservesEntryOrder(t *testing.T) {
	first := &Abc_Obj_t{}
	second := &Abc_Obj_t{}
	v := Vec_PtrAlloc(2)
	defer Vec_PtrFree(v)
	Vec_PtrPush(v, first)
	Vec_PtrPush(v, second)

	got := vecObjPtrArray(v)
	if len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("vecObjPtrArray() = %#v, want [%p %p]", got, first, second)
	}
}
