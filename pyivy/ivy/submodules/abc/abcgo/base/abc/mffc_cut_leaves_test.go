package abc

import "testing"

func TestMffcLabelWithCutLeavesRestoresReferences(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	defer Abc_NtkDelete(n)
	a, b := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	inner, root := Abc_NtkCreateNode(n), Abc_NtkCreateNode(n)
	Abc_ObjAddFanin(inner, a)
	Abc_ObjAddFanin(inner, b)
	Abc_ObjAddFanin(root, inner)
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(po, root)
	if got := Abc_NodeMffcLabel(root, nil); got != 2 {
		t.Fatalf("full MFFC size = %d, want 2", got)
	}
	if got := Abc_NodeMffcLabelWithLeaves(root, []*Object{inner}, nil); got != 1 {
		t.Fatalf("cut MFFC size = %d, want 1", got)
	}
	if got := Abc_NodeMffcLabel(root, nil); got != 2 {
		t.Fatalf("MFFC references not restored: size = %d", got)
	}
}
