package vec

import "testing"

func TestAttributeLifecycle(t *testing.T) {
	starts, frees, managers := 0, 0, 0
	manager := new(int)
	p := Vec_AttAlloc(1, manager, func(any) { managers++ }, func(any) any { starts++; return new(int) }, func(any, any) { frees++ })
	first := Vec_AttEntry(p, 20)
	if first == nil || Vec_AttEntry(p, 20) != first || starts != 1 || len(p.pArrayPtr) < 21 {
		t.Fatal("lazy initialization or growth")
	}
	Vec_AttFreeEntry(p, 20)
	if frees != 1 {
		t.Fatal(frees)
	}
	Vec_AttEntry(p, 0)
	Vec_AttClear(p)
	if frees != 2 || p.pArrayPtr[0] != nil {
		t.Fatal("clear")
	}
	if Vec_AttFree(p, 1) != nil || managers != 1 {
		t.Fatal("manager ownership")
	}
}
func TestAttributeUnmanaged(t *testing.T) {
	p := Vec_AttAlloc(0, "manager", nil, nil, nil)
	Vec_AttWriteEntry(p, 100, "value")
	if Vec_AttEntry(p, 100) != "value" || Vec_AttEntry(p, 101) != nil {
		t.Fatal("write/growth")
	}
	if Vec_AttFree(p, 0) != "manager" {
		t.Fatal("ownership return")
	}
	if Vec_AttFree(nil, 1) != nil {
		t.Fatal("nil free")
	}
}
