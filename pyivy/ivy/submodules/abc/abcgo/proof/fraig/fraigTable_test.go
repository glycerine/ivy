package fraig

import "testing"

func TestSimulationCompareAndFirstDifference(t *testing.T) {
	a := &Fraig_Node_t{puSimR: []uint32{0, 4}, uHashR: 7}
	b := &Fraig_Node_t{puSimR: []uint32{0, 12}, uHashR: 8}
	if got := Fraig_FindFirstDiff(a, b, 0, 2, 1); got != 35 {
		t.Fatalf("difference = %d", got)
	}
	if Fraig_CompareSimInfo(a, b, 2, 1) != 0 {
		t.Fatal("different simulations equal")
	}
	mask := []uint32{0, ^uint32(0)}
	if Fraig_CompareSimInfoUnderMask(a, b, 2, 1, mask) != 0 {
		t.Fatal("masked difference missed")
	}
	Fraig_CollectXors(a, b, 2, 1, mask)
	if mask[0] != 0 || mask[1] != 8 {
		t.Fatalf("xor = %v", mask)
	}
	b.puSimR[1] = 4
	b.uHashR = 7
	if Fraig_CompareSimInfo(a, b, 2, 1) != 1 {
		t.Fatal("equal simulations differ")
	}
	b.fInv = 1
	if got := Fraig_FindFirstDiff(a, b, 1, 2, 1); got != -1 {
		t.Fatalf("equal phases differ at %d", got)
	}
}

func TestFunctionalTableLookupAndRehash(t *testing.T) {
	m := &Fraig_Man_t{pTableF0: Fraig_HashTableCreate(2), iWordStart: 1}
	a := &Fraig_Node_t{puSimD: []uint32{2}, uHashD: 2}
	b := &Fraig_Node_t{puSimD: []uint32{3}, uHashD: 3}
	if Fraig_HashTableLookupF0(m, a) != nil || Fraig_HashTableLookupF0(m, b) != nil {
		t.Fatal("distinct simulations collide")
	}
	c := &Fraig_Node_t{puSimD: []uint32{2}, uHashD: 2}
	if Fraig_HashTableLookupF0(m, c) != a {
		t.Fatal("matching simulation not found")
	}
	b.uHashD = 2
	if Fraig_TableRehashF0(m, 1) != 1 || a.pNextD != b && b.pNextD != a {
		t.Fatal("rehash did not group matching hashes")
	}
}
