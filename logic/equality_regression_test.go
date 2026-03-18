package logic

// Regression tests for Section 1 equality bugs from AUDIT18MARCH.md.
// These tests ensure that structural equality is used (not pointer equality)
// for Sort comparisons and True/False checks.

import "testing"

// --- §1.2 Regression: SortEqual must match reconstructed BooleanSort ---

// TestSortEqual_ReconstructedBoolean verifies that a freshly constructed
// &BooleanSort{} is recognized as equal to the singleton Boolean.
// Bug: code used == (pointer comparison) which fails for non-singleton instances.
func TestSortEqual_ReconstructedBoolean(t *testing.T) {
	reconstructed := &BooleanSort{}
	if reconstructed == Boolean {
		t.Fatal("test precondition: reconstructed must be a different pointer than singleton")
	}
	if !SortEqual(reconstructed, Boolean) {
		t.Error("SortEqual must recognize reconstructed BooleanSort as equal to singleton Boolean")
	}
}

// TestSortEqual_FunctionSortRangeBoolean verifies that SortEqual correctly
// identifies a FunctionSort's range as Boolean.
func TestSortEqual_FunctionSortRangeBoolean(t *testing.T) {
	S := &UninterpretedSort{Name: "node"}
	fs := &FunctionSort{Sorts: []Sort{S, Boolean}}

	if !SortEqual(fs.Range(), Boolean) {
		t.Error("SortEqual must recognize FunctionSort.Range() as Boolean")
	}
	// Also verify with a freshly constructed BooleanSort.
	fs2 := &FunctionSort{Sorts: []Sort{S, &BooleanSort{}}}
	if !SortEqual(fs2.Range(), Boolean) {
		t.Error("SortEqual must recognize FunctionSort.Range() as Boolean even with fresh BooleanSort")
	}
}

// TestSortEqual_ReconstructedUninterpretedSort verifies structural equality
// for UninterpretedSort at different pointers.
func TestSortEqual_ReconstructedUninterpretedSort(t *testing.T) {
	a := &UninterpretedSort{Name: "mytype"}
	b := &UninterpretedSort{Name: "mytype"}
	if a == b {
		t.Fatal("test precondition: a and b must be different pointers")
	}
	if !SortEqual(a, b) {
		t.Error("SortEqual must match structurally-equal UninterpretedSorts at different pointers")
	}
}

// --- §1.3 Regression: IsTrue/IsFalse must handle non-singleton instances ---

// TestIsTrue_Singleton verifies IsTrue recognizes the True singleton.
func TestIsTrue_Singleton(t *testing.T) {
	if !IsTrue(True) {
		t.Error("IsTrue must recognize the True singleton")
	}
}

// TestIsTrue_NonSingleton verifies IsTrue recognizes a freshly constructed
// empty &And{} that is NOT the True singleton pointer.
// Bug: code used n == lg.True (pointer comparison) which misses non-singletons.
func TestIsTrue_NonSingleton(t *testing.T) {
	freshTrue := &And{} // structurally True, but different pointer
	if freshTrue == True {
		t.Fatal("test precondition: freshTrue must be a different pointer than singleton True")
	}
	if !IsTrue(freshTrue) {
		t.Error("IsTrue must recognize &And{} as True even when not the singleton pointer")
	}
}

// TestIsTrue_NonEmpty verifies IsTrue rejects a non-empty And.
func TestIsTrue_NonEmpty(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)
	nonEmpty := &And{Terms: []Node{eq}}
	if IsTrue(nonEmpty) {
		t.Error("IsTrue must reject non-empty And")
	}
}

// TestIsTrue_RejectsOther verifies IsTrue rejects non-And nodes.
func TestIsTrue_RejectsOther(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	if IsTrue(X) {
		t.Error("IsTrue must reject Variable")
	}
	if IsTrue(False) {
		t.Error("IsTrue must reject False (empty Or)")
	}
}

// TestIsFalse_Singleton verifies IsFalse recognizes the False singleton.
func TestIsFalse_Singleton(t *testing.T) {
	if !IsFalse(False) {
		t.Error("IsFalse must recognize the False singleton")
	}
}

// TestIsFalse_NonSingleton verifies IsFalse recognizes a freshly constructed
// empty &Or{} that is NOT the False singleton pointer.
// Bug: code used n == lg.False (pointer comparison) which misses non-singletons.
func TestIsFalse_NonSingleton(t *testing.T) {
	freshFalse := &Or{} // structurally False, but different pointer
	if freshFalse == False {
		t.Fatal("test precondition: freshFalse must be a different pointer than singleton False")
	}
	if !IsFalse(freshFalse) {
		t.Error("IsFalse must recognize &Or{} as False even when not the singleton pointer")
	}
}

// TestIsFalse_NonEmpty verifies IsFalse rejects a non-empty Or.
func TestIsFalse_NonEmpty(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	eq, _ := NewEq(X, X)
	nonEmpty := &Or{Terms: []Node{eq}}
	if IsFalse(nonEmpty) {
		t.Error("IsFalse must reject non-empty Or")
	}
}

// TestIsFalse_RejectsOther verifies IsFalse rejects non-Or nodes.
func TestIsFalse_RejectsOther(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", S)
	if IsFalse(X) {
		t.Error("IsFalse must reject Variable")
	}
	if IsFalse(True) {
		t.Error("IsFalse must reject True (empty And)")
	}
}
