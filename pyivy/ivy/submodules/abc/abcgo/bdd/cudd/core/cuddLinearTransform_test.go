package core

import "testing"

func TestBDDLinearTransformTruthTableAndRootOwnership(t *testing.T) {
	dd := NewBDDManager(4)
	// Exercise nontrivial order and complemented edges.
	if Cudd_ShuffleHeap(dd, []int{1, 0, 2, 3}) == 0 {
		t.Fatal("failed to reorder variables")
	}
	x, y := 1, 3
	f := Cudd_bddIte(dd, dd.Vars[0], Cudd_bddXor(dd, dd.Vars[1], dd.Vars[2]), Cudd_Not(dd.Vars[3]))
	Cudd_Ref(f)
	got := Cudd_bddLinearTransform(dd, f, x, y)
	if got == nil {
		t.Fatal("linear transform returned nil")
	}
	Cudd_Ref(got)
	if Cudd_Regular(f).Ref != 1 || Cudd_Regular(got).Ref != 1 {
		t.Fatalf("root references changed unexpectedly: input=%d result=%d", Cudd_Regular(f).Ref, Cudd_Regular(got).Ref)
	}
	for mask := 0; mask < 1<<len(dd.Vars); mask++ {
		assignment := make([]bool, len(dd.Vars))
		for i := range assignment {
			assignment[i] = mask&(1<<i) != 0
		}
		changed := append([]bool(nil), assignment...)
		changed[x] = assignment[x] == assignment[y]
		want, err := dd.Evaluate(f, changed)
		if err != nil {
			t.Fatal(err)
		}
		value, err := dd.Evaluate(got, assignment)
		if err != nil || value != want {
			t.Fatalf("mask %04b: transformed=%v want=%v (err=%v)", mask, value, want, err)
		}
		original, err := dd.Evaluate(f, assignment)
		if err != nil {
			t.Fatal(err)
		}
		if now, _ := dd.Evaluate(f, assignment); now != original {
			t.Fatal("input root changed")
		}
	}
	Cudd_RecursiveDeref(dd, got)
	Cudd_RecursiveDeref(dd, f)
}

func TestBDDLinearTransformRejectsInvalidVariables(t *testing.T) {
	dd := NewBDDManager(2)
	f := dd.Vars[0]
	for _, pair := range [][2]int{{-1, 0}, {0, -1}, {2, 0}, {0, 2}, {1, 1}} {
		if got := Cudd_bddLinearTransform(dd, f, pair[0], pair[1]); got != nil {
			t.Errorf("transform with indices %v unexpectedly returned a root", pair)
		}
	}
}

func TestBDDLinearTransformInPlacePreservesRootPointers(t *testing.T) {
	dd := NewBDDManager(4)
	f := Cudd_bddIte(dd, dd.Vars[0], Cudd_bddXor(dd, dd.Vars[1], dd.Vars[2]), dd.Vars[3])
	g := Cudd_bddIte(dd, dd.Vars[2], Cudd_Not(dd.Vars[0]), dd.Vars[1])
	Cudd_Ref(f)
	Cudd_Ref(g)
	fr, gr := f, g
	wantF := Cudd_bddLinearTransform(dd, f, 1, 2)
	wantG := Cudd_bddLinearTransform(dd, g, 1, 2)
	wantValues := make([][2]bool, 1<<len(dd.Vars))
	for mask := range wantValues {
		assignment := make([]bool, len(dd.Vars))
		for i := range assignment {
			assignment[i] = mask&(1<<i) != 0
		}
		wantValues[mask][0], _ = dd.Evaluate(wantF, assignment)
		wantValues[mask][1], _ = dd.Evaluate(wantG, assignment)
	}
	if !Cudd_bddLinearTransformInPlace(dd, 1, 2) {
		t.Fatal("manager-wide linear transform failed")
	}
	if fr != f || gr != g || f.Ref != fr.Ref || g.Ref != gr.Ref {
		t.Fatal("manager-wide transform changed a retained root pointer or refcount")
	}
	for mask := 0; mask < 1<<len(dd.Vars); mask++ {
		assignment := make([]bool, len(dd.Vars))
		for i := range assignment {
			assignment[i] = mask&(1<<i) != 0
		}
		for i, root := range []*DdNode{f, g} {
			got, err := dd.Evaluate(root, assignment)
			want := wantValues[mask][i]
			if err != nil || got != want {
				t.Fatalf("mask %04b root %d: transformed=%v want=%v (err=%v)", mask, i, got, want, err)
			}
		}
	}
	if err := Cudd_CheckConsistency(dd); err != nil {
		t.Fatalf("unique table invalid after transform: %v", err)
	}
}

func TestBDDLinearLevelCountsExcludeIsolatedProjections(t *testing.T) {
	dd := NewBDDManager(3)
	if got := Cudd_bddLinearNodeCount(dd); got != 0 {
		t.Fatalf("initial non-isolated BDD node count = %d, want 0", got)
	}
	if got := Cudd_bddLinearPreprocessSize(dd); got != 3 {
		t.Fatalf("preprocess size = %d, want 3 isolated projections", got)
	}
	f := Cudd_bddIte(dd, dd.Vars[0], dd.Vars[1], dd.LogicZero)
	Cudd_Ref(f)
	if got := Cudd_bddLinearLevelNodeCount(dd, 0); got != 1 {
		t.Fatalf("level 0 count = %d, want 1", got)
	}
	if got := Cudd_bddLinearLevelNodeCount(dd, 1); got != 1 {
		t.Fatalf("level 1 count = %d, want 1", got)
	}
	if got := Cudd_bddLinearNodeCount(dd); got != 2 {
		t.Fatalf("non-isolated BDD node count = %d, want 2", got)
	}
}
