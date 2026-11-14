package abc

import "testing"

func TestAbcNtkResetNamesDropsOldAssignments(t *testing.T) {
	n := NewNetwork("reset")
	pi := Abc_NtkCreatePi(n)
	po := Abc_NtkCreatePo(n)
	Abc_ObjAssignName(pi, "stale_pi", "")
	Abc_ObjAssignName(po, "stale_po", "")

	Abc_NtkResetNames(n)
	if got := Abc_ObjName(pi); got == "stale_pi" {
		t.Fatalf("PI name survived manager reset: %q", got)
	}
	if got := Abc_ObjName(po); got == "stale_po" {
		t.Fatalf("PO name survived manager reset: %q", got)
	}
	if got := Abc_ObjAssignName(pi, "fresh_pi", ""); got != "fresh_pi" {
		t.Fatalf("new name after reset = %q", got)
	}
}
