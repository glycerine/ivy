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
		for _, n := range dd.unique.entries {
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
	for _, n := range dd.unique.entries {
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

func TestAndCacheGatingCountsProjectionPins(t *testing.T) {
	for _, external := range []bool{false, true} {
		dd := NewBDDManager(2)
		x, y := dd.Vars[0], dd.Vars[1]
		if external {
			Cudd_Ref(x)
			Cudd_Ref(y)
		}
		for i := range 3 {
			lookups, hits := dd.cacheLookups, dd.cacheHits
			root := Cudd_bddAnd(dd, x, y)
			Cudd_Ref(root)
			var wantLookups, wantHits uint64
			if external {
				wantLookups = 1
				if i != 0 {
					wantHits = 1
				}
			}
			if dd.cacheLookups-lookups != wantLookups || dd.cacheHits-hits != wantHits {
				t.Errorf("external=%v call %d got lookups/hits %d/%d, want %d/%d", external, i, dd.cacheLookups-lookups, dd.cacheHits-hits, wantLookups, wantHits)
			}
			Cudd_RecursiveDeref(dd, root)
			x, y = y, x
		}
		Cudd_Quit(dd)
	}
}

func TestComputedRecordsAcrossIndexGrowthFlushAndResize(t *testing.T) {
	dd := Cudd_Init(0, 0, 8, 1024, 0)
	defer Cudd_Quit(dd)
	Cudd_SetMaxCache(dd, 1)
	operations := make([]uintptr, 0, 160)
	distinct := map[uint32]bool{}
	for op := uintptr(1); len(operations) < cap(operations); op++ {
		k := computedKey{f: dd.One, g: dd.Zero, op: op}
		slot := k.hash() >> dd.computed.shift
		if distinct[slot] {
			continue
		}
		distinct[slot] = true
		operations = append(operations, op)
		Cudd_CacheInsert(dd, op, dd.One, dd.Zero, nil, dd.PlusInfinity, 0)
		if len(operations) == 4 && dd.computed.index != nil {
			t.Fatal("four records allocated a full index")
		}
	}
	if dd.computed.index == nil || dd.computed.sparse != nil || Cudd_CacheSize(dd) != len(operations) {
		t.Fatal("index growth lost occupied slots")
	}
	for _, op := range operations {
		if Cudd_CacheLookup(dd, op, dd.One, dd.Zero, nil, 0) != dd.PlusInfinity {
			t.Fatalf("lookup failed after index growth for operation %d", op)
		}
	}
	Cudd_CacheFlush(dd)
	if Cudd_ReadCacheUsedSlots(dd) != float64(len(operations))/1024 || Cudd_CacheSize(dd) != 0 {
		t.Fatal("flush lost occupied slot counts")
	}
	for e := range dd.computed.records {
		if *e != (computedEntry{}) {
			t.Fatal("flush retained a diagram reference")
		}
	}
	for _, op := range operations[:35] {
		Cudd_CacheInsert(dd, op, dd.One, dd.Zero, nil, dd.MinusInfinity, 0)
	}
	Cudd_ResizeCache(dd)
	if Cudd_ReadCacheSlots(dd) != 2048 || Cudd_ReadCacheUsedSlots(dd) != 35.0/2048 || Cudd_CacheSize(dd) != 35 {
		t.Fatal("resize retained invalidated records or lost surviving records")
	}
	for i, op := range operations {
		got := Cudd_CacheLookup(dd, op, dd.One, dd.Zero, nil, 0)
		if i < 35 && got != dd.MinusInfinity || i >= 35 && got != nil {
			t.Fatalf("wrong result for operation %d after resizing", op)
		}
	}
}
