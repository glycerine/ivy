package fraig

import (
	"testing"
	"unsafe"
)

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

func TestStructuralHashUsesTaggedPointerFormula(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	for _, x := range []*Fraig_Node_t{a, Fraig_Not(a)} {
		for _, y := range []*Fraig_Node_t{b, Fraig_Not(b)} {
			xp, yp := uintptr(unsafe.Pointer(Fraig_Regular(x))), uintptr(unsafe.Pointer(Fraig_Regular(y)))
			if Fraig_IsComplement(x) {
				xp |= 1
			}
			if Fraig_IsComplement(y) {
				yp |= 1
			}
			const bins = 1009
			want := uint32((xp + yp*12582917) % bins)
			if got := fraigStructuralKey(x, y, bins); got != want {
				t.Fatalf("tagged pointer hash = %d, want %d", got, want)
			}
		}
	}
}
