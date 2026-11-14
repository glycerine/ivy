package mpm

import "testing"

// The C manager uses sentinel-terminated pages and recovers the manager from
// a page prefix. Exercise both sides of a page boundary and pointer stability.
func TestMigPagesPreserveObjectsAndManager(t *testing.T) {
	p := &Mig_Man_t{}
	zero := Mig_ManAppendObj(p)
	first := Mig_ManAppendCi(p)
	firstObj := Mig_ManObj(p, first>>1)
	for p.nObjs <= 1<<MIG_BASE {
		Mig_ManAppendCi(p)
	}
	last := Mig_ManObj(p, 1<<MIG_BASE)
	if Mig_ObjMan(zero) != p || Mig_ObjMan(last) != p || Mig_ManObj(p, first>>1) != firstObj {
		t.Fatal("page boundary changed object ownership or identity")
	}
	if Mig_ObjId(last) != 1<<MIG_BASE || Mig_ObjIsCi(last) == 0 {
		t.Fatalf("last object id=%d ci=%d", Mig_ObjId(last), Mig_ObjIsCi(last))
	}
	if Mig_ObjIsNone(&Mig_ManPage(p, 0)[MIG_MASK+1]) == 0 {
		t.Fatal("first page sentinel is not marked empty")
	}
}

func TestMigAppendOrdersAndPolarity(t *testing.T) {
	p := &Mig_Man_t{}
	Mig_ManAppendObj(p)
	a, b, c := Mig_ManAppendCi(p), Mig_ManAppendCi(p), Mig_ManAppendCi(p)
	and := Mig_ManObj(p, Mig_ManAppendAnd(p, b^1, a)>>1)
	xor := Mig_ManObj(p, Mig_ManAppendXor(p, a, b)>>1)
	mux := Mig_ManObj(p, Mig_ManAppendMux(p, b, a, c)>>1)
	if Mig_ObjIsAnd(and) == 0 || Mig_ObjFaninLit(and, 0) != a || Mig_ObjFaninLit(and, 1) != b^1 {
		t.Fatal("AND fanin order or polarity differs from C")
	}
	if Mig_ObjIsXor(xor) == 0 || Mig_ObjFaninLit(xor, 0) != b || Mig_ObjFaninLit(xor, 1) != a {
		t.Fatal("XOR fanin order differs from C")
	}
	if Mig_ObjFaninLit(mux, 0) != a || Mig_ObjFaninLit(mux, 1) != b || Mig_ObjFaninLit(mux, 2) != c^1 {
		t.Fatal("MUX fanin order or swapped control polarity differs from C")
	}
}

func TestMigSharedSupportAndReferenceCounts(t *testing.T) {
	p := Mig_ManStart()
	defer Mig_ManStop(p)
	a, b, c := Mig_ManAppendCi(p), Mig_ManAppendCi(p), Mig_ManAppendCi(p)
	left := Mig_ManAppendAnd(p, a, b)
	right := Mig_ManAppendAnd(p, a, c)
	root := Mig_ManAppendAnd(p, left, right)
	Mig_ManAppendCo(p, root)
	if got := Mig_ManSuppSizeOne(Mig_ManObj(p, root>>1)); got != 3 {
		t.Fatalf("shared support size = %d, want 3", got)
	}
	Mig_ManSetRefs(p)
	for _, tc := range []struct{ lit, want int32 }{{a, 2}, {b, 1}, {c, 1}, {left, 1}, {right, 1}, {root, 1}} {
		if got := Mig_ObjRefNum(Mig_ManObj(p, tc.lit>>1)); got != tc.want {
			t.Errorf("literal %d refs=%d, want %d", tc.lit, got, tc.want)
		}
	}
}

func TestMigReferenceIteratorStopsAtConstantFanin(t *testing.T) {
	p := Mig_ManStart()
	defer Mig_ManStop(p)
	a := Mig_ManAppendCi(p)
	Mig_ManAppendAnd(p, 0, a)
	Mig_ManSetRefs(p)
	if got := Mig_ObjRefNum(Mig_ManObj(p, a>>1)); got != 0 {
		t.Fatalf("C fanin iterator stops at zero, got %d references", got)
	}
}
