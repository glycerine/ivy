package abc

import "testing"

func TestLatchInitializationAPIs(t *testing.T) {
	n := NewNetwork("latches")
	latch := n.NewObject(ObjectLatch)
	if Abc_LatchIsInitNone(latch) != 1 {
		t.Fatal("fresh latch is not uninitialized")
	}
	setters := []func(*Object){Abc_LatchSetInitNone, Abc_LatchSetInit0, Abc_LatchSetInit1, Abc_LatchSetInitDc}
	predicates := []func(*Object) int32{Abc_LatchIsInitNone, Abc_LatchIsInit0, Abc_LatchIsInit1, Abc_LatchIsInitDc}
	for i, set := range setters {
		set(latch)
		for j, predicate := range predicates {
			if (predicate(latch) != 0) != (i == j) {
				t.Fatal(i, j, latch.Data)
			}
		}
	}
	if Abc_NtkHasOnlyLatchBoxes(n) != 1 || Abc_NtkIsComb(n) != 0 {
		t.Fatal("latch network classification")
	}
	n.NewObject(ObjectBlackbox)
	if Abc_NtkHasOnlyLatchBoxes(n) != 0 {
		t.Fatal("mixed boxes accepted")
	}
}
