package abc

import "testing"

func TestObjectPointerCompare(t *testing.T) {
	n := NewNetwork("sort")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	if Abc_ObjPointerCompare(a, a) != 0 || Abc_ObjPointerCompare(a, b) == 0 || Abc_ObjPointerCompare(a, b) != -Abc_ObjPointerCompare(b, a) {
		t.Fatal("invalid pointer ordering")
	}
	Abc_NtkDetectMatching(n)
}
