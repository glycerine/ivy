package ivylogic

// Regression tests for Section 1 equality bugs from AUDIT18MARCH.md.
// These tests ensure map keys and helpers use structural equality.

import (
	"testing"

	lg "github.com/glycerine/ivy/goivy/logic"
)

// --- §1.1 Regression: GetSortRefinement must use structural keys ---

// TestGetSortRefinement_StructuralKeys verifies that GetSortRefinement
// returns a map keyed by lg.SortKey (string) so that lookups with
// structurally-equal sorts at different pointers succeed.
//
// Bug: the old code used map[lg.Sort]lg.Sort where lg.Sort is an interface,
// so Go maps used pointer equality for lookups. Two structurally-equal sorts
// at different pointers would silently miss.
func TestGetSortRefinement_StructuralKeys(t *testing.T) {
	sig := NewSig()

	// Add sort "mytype" that maps (via Interp) to another uninterpreted sort.
	// This makes "mytype" non-canonical, so GetSortRefinement will include it.
	sortA := &lg.UninterpretedSort{Name: "mytype"}
	canonical := &lg.UninterpretedSort{Name: "canonical"}
	if err := sig.AddSort(sortA); err != nil {
		t.Fatal(err)
	}
	if err := sig.AddSort(canonical); err != nil {
		t.Fatal(err)
	}
	// Interp maps sort name → another uninterpreted sort → makes it non-canonical.
	sig.Interp["mytype"] = canonical

	refinement := GetSortRefinement(sig)

	// Now look up using a DIFFERENT pointer with the same structural identity.
	differentPointer := &lg.UninterpretedSort{Name: "mytype"}
	if differentPointer == sortA {
		t.Fatal("test precondition: differentPointer must be a different pointer")
	}

	lookupKey := lg.SortKey(differentPointer)
	result, ok := refinement[lookupKey]
	if !ok {
		t.Fatal("GetSortRefinement lookup with SortKey of a different-pointer sort must find the entry; " +
			"old map[lg.Sort] keys used pointer equality and would miss here")
	}
	if !lg.SortEqual(result, canonical) {
		t.Errorf("expected canonical sort %v, got %v", canonical, result)
	}
}

// TestGetSortRefinement_ReturnType verifies the return type is
// map[lg.NodeKey]lg.Sort, not map[lg.Sort]lg.Sort.
func TestGetSortRefinement_ReturnType(t *testing.T) {
	sig := NewSig()
	result := GetSortRefinement(sig)
	// This is a compile-time check: if the return type were map[lg.Sort]lg.Sort,
	// this assignment would fail to compile.
	var _ map[lg.NodeKey]lg.Sort = result
	_ = result
}

// --- §1.3 Regression: ivylogic.IvyIsTrue/IsFalse delegate to lg.IsTrue/lg.IsFalse ---

// TestIsTrueIsFalse_NonSingleton verifies that ivylogic.IvyIsTrue and
// ivylogic.IvyIsFalse handle non-singleton empty And/Or nodes.
//
// Bug: code used n == lg.True (pointer comparison) which would miss
// freshly constructed &And{} / &Or{} nodes.
func TestIsTrueIsFalse_NonSingleton(t *testing.T) {
	freshTrue := &lg.And{} // different pointer than lg.True
	freshFalse := &lg.Or{} // different pointer than lg.False

	if !IvyIsTrue(freshTrue) {
		t.Error("ivylogic.IvyIsTrue must recognize non-singleton &And{}")
	}
	if !IvyIsFalse(freshFalse) {
		t.Error("ivylogic.IvyIsFalse must recognize non-singleton &Or{}")
	}

	// Singletons should also work.
	if !IvyIsTrue(lg.True) {
		t.Error("ivylogic.IvyIsTrue must recognize lg.True singleton")
	}
	if !IvyIsFalse(lg.False) {
		t.Error("ivylogic.IvyIsFalse must recognize lg.False singleton")
	}

	// Cross-check: True is not False, False is not True.
	if IvyIsTrue(lg.False) {
		t.Error("IsTrue must reject False")
	}
	if IvyIsFalse(lg.True) {
		t.Error("IsFalse must reject True")
	}
}
