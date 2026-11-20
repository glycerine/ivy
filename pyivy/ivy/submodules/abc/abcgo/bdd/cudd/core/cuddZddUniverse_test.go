package core

import "testing"

func TestZDDUniverseReleasesOnlyTheManagerReference(t *testing.T) {
	dd := Cudd_Init(1, 3, 16, 64, 0)
	defer Cudd_Quit(dd)
	root := dd.Univ[0]
	oneRefs := dd.One.Ref
	check := func(phase string, zlive int, refs uint32) {
		t.Helper()
		if got := Cudd_zddReadNodeCount(dd); got != zlive || dd.One.Ref != refs || Cudd_ReadDead(dd) != 0 {
			t.Errorf("%s ZDD live %d, one refs %d, BDD dead %d, want %d, %d, 0", phase, got, dd.One.Ref, Cudd_ReadDead(dd), zlive, refs)
		}
	}
	check("initial", 5, oneRefs)
	FreeZDDUniverse(dd)
	check("free", 2, oneRefs-2)
	FreeZDDUniverse(dd)
	check("already free", 2, oneRefs-2)
	if InitializeZDDUniverse(dd) != 1 || dd.Univ[0] != root {
		t.Fatal("failed to reclaim the existing universe")
	}
	check("rebuilt", 5, oneRefs)
	Cudd_Ref(root)
	FreeZDDUniverse(dd)
	check("external reference", 5, oneRefs)
	Cudd_RecursiveDeref(dd, root)
	check("last reference", 2, oneRefs-2)
	if InitializeZDDUniverse(dd) != 1 {
		t.Fatal("failed to rebuild the released universe")
	}
	check("rebuilt again", 5, oneRefs)
}

func TestReplacingZDDUniverseReleasesPreviousRoot(t *testing.T) {
	dd := Cudd_Init(0, 1, 16, 64, 0)
	defer Cudd_Quit(dd)
	old := dd.Univ[0]
	if !dd.ensureZddVar(2) {
		t.Fatal("failed to grow universe")
	}
	if dd.Univ[0] == old || !old.childrenReleased {
		t.Fatal("growth kept the replaced, unowned universe live")
	}
	if got := Cudd_zddReadNodeCount(dd); got != 5 {
		t.Errorf("grown universe has %d live nodes, want three decisions and two terminals", got)
	}
}

func TestReadZddOneDuringUniverseRebuild(t *testing.T) {
	dd := Cudd_Init(0, 2, 16, 64, 0)
	defer Cudd_Quit(dd)
	FreeZDDUniverse(dd)
	if got := Cudd_ReadZddOne(dd, 1); got != nil {
		t.Fatal("read returned a level from the freed universe")
	}
	if Cudd_ReadZddOne(dd, 2) != dd.One {
		t.Fatal("constant one is missing beyond the last level")
	}
}

func TestZddProjectionAfterFreeRebuildsItsUniverse(t *testing.T) {
	dd := Cudd_Init(0, 3, 16, 64, 0)
	defer Cudd_Quit(dd)
	v := Cudd_zddIthVar(dd, 0)
	Cudd_Ref(v)
	defer Cudd_RecursiveDeref(dd, v)
	FreeZDDUniverse(dd)
	if got := Cudd_zddIthVar(dd, 0); got != v || len(dd.Univ) != 3 {
		t.Fatal("projection lookup lost its canonical subgraph")
	}
}
