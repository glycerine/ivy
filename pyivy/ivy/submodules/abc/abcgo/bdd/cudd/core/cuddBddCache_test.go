package core

import "testing"

func TestBDDComputedHitRestoresChildReferences(t *testing.T) {
	for _, op := range []string{"and", "ite", "add ite", "computed", "constant lookup"} {
		t.Run(op, func(t *testing.T) {
			dd := NewBDDManager(3)
			defer Cudd_Quit(dd)
			x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
			for _, v := range dd.Vars {
				Cudd_Ref(v)
				Cudd_Ref(v)
			}
			var repeat func() *DdNode
			switch op {
			case "and":
				repeat = func() *DdNode { return Cudd_bddAnd(dd, x, Cudd_Not(y)) }
			case "ite":
				repeat = func() *DdNode { return Cudd_bddIte(dd, x, y, z) }
			case "add ite":
				ax, ay, az := Cudd_addIthVar(dd, 0), Cudd_addIthVar(dd, 1), Cudd_addIthVar(dd, 2)
				repeat = func() *DdNode { return Cudd_addIte(dd, ax, ay, az) }
			default:
				family := uint8(0)
				if op == "constant lookup" {
					family = 2
				}
				n := Cudd_bddAnd(dd, x, Cudd_Not(y))
				Cudd_CacheInsert(dd, 7, x, y, nil, n, family)
				repeat = func() *DdNode { return Cudd_CacheLookup(dd, 7, x, y, nil, family) }
			}
			n := repeat()
			Cudd_Ref(n)
			r := Cudd_Regular(n)
			child := Cudd_Regular(r.T)
			before := child.Ref
			Cudd_RecursiveDeref(dd, n)
			if !r.childrenReleased || child.Ref >= before {
				t.Fatal("test setup did not release the child")
			}
			if repeat() != n {
				t.Fatal("lookup lost the canonical result")
			}
			if op == "constant lookup" {
				if !r.childrenReleased {
					t.Fatal("constant-only lookup acquired child references")
				}
			} else if r.childrenReleased || child.Ref != before || r.Ref != 0 {
				t.Fatalf("cache hit left child refs %d, want %d (root refs %d)", child.Ref, before, r.Ref)
			}
		})
	}
}

func TestBDDISOPLimitRestoresPreexistingRefs(t *testing.T) {
	for truth := 1; truth < 255; truth++ {
		dd := NewBDDManager(3)
		f := bddFromTruth(dd, uint8(truth), 3)
		Cudd_Ref(f)
		refs := map[*DdNode]uint32{dd.One: dd.One.Ref}
		for _, n := range dd.unique {
			refs[n] = n.Ref
		}
		if r, _ := CountISOPBounded(dd, f, f, 1, 100); r == nil {
			for n, before := range refs {
				if n.Ref != before {
					t.Fatalf("truth 0x%02x node %d refs %d, want %d after rollback", truth, n.ID, n.Ref, before)
				}
			}
		}
		Cudd_Quit(dd)
	}
}

func TestBDDAndLimitRestoresPreexistingRefs(t *testing.T) {
	dd := NewBDDManager(4)
	defer Cudd_Quit(dd)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2])
	g := Cudd_bddXor(dd, dd.Vars[1], dd.Vars[3])
	Cudd_Ref(f)
	Cudd_Ref(g)
	refs := map[*DdNode]uint32{dd.One: dd.One.Ref}
	for _, n := range dd.unique {
		refs[n] = n.Ref
	}
	for i := 0; i < 3; i++ {
		if n := Cudd_bddAndLimit(dd, f, g, 1); n != nil {
			t.Fatal("conjunction exceeded its node limit")
		}
	}
	for n, before := range refs {
		if n.Ref != before {
			t.Errorf("node %d refs grew from %d to %d across failed attempts", n.ID, before, n.Ref)
		}
	}
}

func TestBDDAndLimitRestoresReleasedIntermediate(t *testing.T) {
	dd := NewBDDManager(4)
	defer Cudd_Quit(dd)
	x, y, z, w := dd.Vars[0], dd.Vars[1], dd.Vars[2], dd.Vars[3]
	f, g := Cudd_bddXor(dd, x, z), Cudd_bddXor(dd, y, w)
	sub := Cudd_bddAnd(dd, Cudd_Not(z), Cudd_Not(w))
	Cudd_Ref(sub)
	Cudd_RecursiveDeref(dd, sub)
	before := Cudd_Regular(w).Ref
	if !Cudd_Regular(sub).childrenReleased || Cudd_bddAndLimit(dd, f, g, 0) != nil {
		t.Fatal("expected released first cofactor and failed allocation")
	}
	if !Cudd_Regular(sub).childrenReleased || w.Ref != before {
		t.Fatal("failed operation retained the revived cofactor's child references")
	}
}

func TestBDDAndSharesCommutativeCacheEntry(t *testing.T) {
	dd := NewBDDManager(4)
	defer Cudd_Quit(dd)
	f := Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2])
	g := Cudd_bddXor(dd, dd.Vars[1], dd.Vars[3])
	for _, root := range []*DdNode{f, g} {
		Cudd_Ref(root)
		Cudd_Ref(root)
	}
	r := Cudd_bddAnd(dd, f, Cudd_Not(g))
	hits, lookups := dd.cacheHits, dd.cacheLookups
	if Cudd_bddAnd(dd, Cudd_Not(g), f) != r || Cudd_bddOr(dd, Cudd_Not(f), g) != Cudd_Not(r) {
		t.Fatal("swapped or complemented conjunction changed the function")
	}
	if dd.cacheHits-hits != 2 || dd.cacheLookups-lookups != 2 {
		t.Fatalf("swapped AND and De Morgan OR needed %d lookups and %d hits, want two each", dd.cacheLookups-lookups, dd.cacheHits-hits)
	}
}
