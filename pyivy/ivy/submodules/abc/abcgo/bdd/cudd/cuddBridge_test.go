package cudd

import "testing"

func TestBridgePublicForwarders(t *testing.T) {
	dd := NewBDDManager(1)
	f := Cudd_BddToAdd(dd, Cudd_bddIthVar(dd, 0))
	if f == nil || Cudd_addBddPattern(dd, f) != dd.Vars[0] || Cudd_addBddThreshold(dd, f, 1) != dd.Vars[0] || Cudd_addBddStrictThreshold(dd, f, 0) != dd.Vars[0] || Cudd_addBddInterval(dd, f, 1, 1) != dd.Vars[0] || Cudd_addBddIthBit(dd, f, 0) != dd.Vars[0] {
		t.Fatal("root bridge forwarders")
	}
	dest := NewBDDManager(1)
	if Cudd_bddTransfer(dd, dest, dd.Vars[0]) != dest.Vars[0] {
		t.Fatal("transfer forwarder")
	}
}
