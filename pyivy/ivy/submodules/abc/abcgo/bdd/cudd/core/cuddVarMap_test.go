package core

import "testing"

func mappingCubes() (dd *DdManager, sources, targets []*DdNode) {
	const width = 8
	dd = Cudd_Init(2*width, 0, 64, 4096, 0)
	for mask := range 1 << width {
		x, y := dd.One, dd.One
		for i := width - 1; i >= 0; i-- {
			neg := mask&(1<<i) == 0
			x = Cudd_bddAnd(dd, Cudd_NotCond(dd.Vars[i+width], neg), x)
			y = Cudd_bddAnd(dd, Cudd_NotCond(dd.Vars[i], neg), y)
		}
		Cudd_Ref(x)
		Cudd_Ref(y)
		sources, targets = append(sources, x), append(targets, y)
	}
	return
}

func BenchmarkMapStateCubes(b *testing.B) {
	dd, sources, targets := mappingCubes()
	defer Cudd_Quit(dd)
	if Cudd_SetVarMap(dd, dd.Vars[:8], dd.Vars[8:]) != 1 {
		b.Fatal("registration")
	}
	b.ReportAllocs()
	for b.Loop() {
		for i, x := range sources {
			if Cudd_bddVarMap(dd, x) != targets[i] {
				b.Fatal("state mapping", i)
			}
		}
	}
}

func TestRegisteredMappingLifecycle(t *testing.T) {
	dd := NewBDDManager(3)
	defer Cudd_Quit(dd)
	if Cudd_bddVarMap(dd, dd.One) != nil {
		t.Fatal("expected no mapping before registration")
	}
	a, b, c := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	Cudd_CacheInsert(dd, 219, a, b, nil, c, 0)
	if Cudd_SetVarMap(dd, []*DdNode{a}, []*DdNode{b}) != 1 || Cudd_CacheLookup(dd, 219, a, b, nil, 0) != c {
		t.Fatal("initial registration discarded an unrelated cached result")
	}
	f := Cudd_bddAnd(dd, a, Cudd_Not(c))
	Cudd_Ref(f)
	want := Cudd_bddAnd(dd, b, Cudd_Not(c))
	Cudd_Ref(want)
	if Cudd_bddVarMap(dd, f) != want || Cudd_bddVarMap(dd, Cudd_Not(f)) != Cudd_Not(want) {
		t.Fatal("mapping lost complement phase")
	}
	if Cudd_ShuffleHeap(dd, []int{2, 0, 1}) == 0 || Cudd_bddVarMap(dd, f) != want {
		t.Fatal("reordering changed the registered index mapping")
	}
	d := Cudd_bddNewVarAtLevel(dd, 0)
	if Cudd_bddVarMap(dd, d) != d || Cudd_bddVarMap(dd, a) != b {
		t.Fatal("new variable did not extend the mapping by identity")
	}
	foreign := NewBDDManager(1)
	defer Cudd_Quit(foreign)
	if Cudd_SetVarMap(dd, []*DdNode{a}, nil) != 0 ||
		Cudd_SetVarMap(dd, []*DdNode{foreign.Vars[0]}, []*DdNode{a}) != 0 ||
		Cudd_bddVarMap(dd, foreign.One) != nil || Cudd_bddVarMap(dd, f) != want {
		t.Fatal("invalid registration or foreign input altered the mapping")
	}
	Cudd_CacheInsert(dd, 219, a, b, nil, c, 0)
	if Cudd_SetVarMap(dd, nil, nil) != 1 || Cudd_CacheLookup(dd, 219, a, b, nil, 0) != nil || Cudd_bddVarMap(dd, f) != f {
		t.Fatal("replacing a map did not flush stale results and reset to identity")
	}
	if !Cudd_DestroyBDDSubtables(dd, 1) || Cudd_bddVarMap(dd, f) != nil {
		t.Fatal("successful subtable destruction retained the map")
	}
}

func TestOverlappingRegisteredMappingAllTruthTables(t *testing.T) {
	dd := NewBDDManager(3)
	defer Cudd_Quit(dd)
	// Native registration assigns each pair in order, even when they overlap.
	if Cudd_SetVarMap(dd, dd.Vars[:2], dd.Vars[1:]) != 1 {
		t.Fatal("registration")
	}
	perm := []int{1, 2, 1}
	for i := range 256 {
		f := bddFromTruth(dd, uint8(i), 3)
		got := Cudd_bddVarMap(dd, f)
		if got != Cudd_bddPermute(dd, f, perm) {
			t.Fatalf("overlapping pairs changed truth %02x", i)
		}
	}
}
