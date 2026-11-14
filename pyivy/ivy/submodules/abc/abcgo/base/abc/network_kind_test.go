package abc

import "testing"

func TestChangeObjectKindKeepsKindIndexesConsistent(t *testing.T) {
	n := NewNetwork("kind change")
	box := n.NewObject(ObjectBlackbox)
	if err := n.ChangeObjectKind(box, ObjectWhitebox); err != nil {
		t.Fatal(err)
	}
	if box.Kind != ObjectWhitebox || len(n.ObjectsOfKind(ObjectBlackbox)) != 0 || len(n.ObjectsOfKind(ObjectWhitebox)) != 1 || n.ObjectsOfKind(ObjectWhitebox)[0] != box {
		t.Fatal("kind indexes did not follow object kind change")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
	second := n.NewObject(ObjectBlackbox)
	Abc_ObjBlackboxToWhitebox(second)
	if second.Kind != ObjectWhitebox || len(n.ObjectsOfKind(ObjectWhitebox)) != 2 {
		t.Fatal("ABC conversion did not update indexes")
	}
}
