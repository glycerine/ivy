package goivy

import (
	"testing"

	"github.com/glycerine/ivy/goivy/smt"
)

// --- ArraySort creation and introspection ---

func TestArraySortIntInt(t *testing.T) {
	ctx := smt.NewZ3Context()
	intSort := ctx.IntSort()
	arrSort := ctx.ArraySort(intSort, intSort)

	if arrSort.Kind() != smt.SortArray {
		t.Fatalf("expected SortArray, got %d", arrSort.Kind())
	}
	if arrSort.ArrayDomain().Kind() != smt.SortInt {
		t.Fatalf("expected domain SortInt, got %d", arrSort.ArrayDomain().Kind())
	}
	if arrSort.ArrayRange().Kind() != smt.SortInt {
		t.Fatalf("expected range SortInt, got %d", arrSort.ArrayRange().Kind())
	}
}

func TestArraySortIntBool(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.BoolSort())

	if arrSort.Kind() != smt.SortArray {
		t.Fatalf("expected SortArray kind")
	}
	if arrSort.ArrayDomain().Kind() != smt.SortInt {
		t.Fatalf("expected domain Int")
	}
	if arrSort.ArrayRange().Kind() != smt.SortBool {
		t.Fatalf("expected range Bool")
	}
}

func TestArraySortBvDomain(t *testing.T) {
	ctx := smt.NewZ3Context()
	bv8 := ctx.BvSort(8)
	arrSort := ctx.ArraySort(bv8, ctx.IntSort())

	if arrSort.Kind() != smt.SortArray {
		t.Fatalf("expected SortArray kind")
	}
	if arrSort.ArrayDomain().Kind() != smt.SortBV {
		t.Fatalf("expected domain BV, got %d", arrSort.ArrayDomain().Kind())
	}
	if ctx.BvSortSize(arrSort.ArrayDomain()) != 8 {
		t.Fatalf("expected domain BV width 8")
	}
}

func TestArraySortUninterpreted(t *testing.T) {
	ctx := smt.NewZ3Context()
	nodeSort := ctx.UninterpretedSort("node")
	valSort := ctx.UninterpretedSort("value")
	arrSort := ctx.ArraySort(nodeSort, valSort)

	if arrSort.Kind() != smt.SortArray {
		t.Fatalf("expected SortArray")
	}
	if arrSort.ArrayDomain().Kind() != smt.SortUninterpreted {
		t.Fatalf("expected uninterpreted domain")
	}
	if arrSort.ArrayRange().Kind() != smt.SortUninterpreted {
		t.Fatalf("expected uninterpreted range")
	}
}

func TestArraySortNested(t *testing.T) {
	ctx := smt.NewZ3Context()
	// Array(Int, Array(Int, Bool)) — nested array
	inner := ctx.ArraySort(ctx.IntSort(), ctx.BoolSort())
	outer := ctx.ArraySort(ctx.IntSort(), inner)

	if outer.Kind() != smt.SortArray {
		t.Fatalf("expected SortArray")
	}
	rng := outer.ArrayRange()
	if rng.Kind() != smt.SortArray {
		t.Fatalf("expected nested array range, got %d", rng.Kind())
	}
	if rng.ArrayRange().Kind() != smt.SortBool {
		t.Fatalf("expected inner range Bool")
	}
}

// --- Select ---

func TestSelectFromSymbolicArray(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	// Select from symbolic array is satisfiable (unconstrained)
	sel := ctx.Select(a, ctx.IntVal(0))
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(sel, ctx.IntVal(42)))
	assertSat(t, slv, "Select from symbolic array can equal 42")
}

func TestSelectFromConstArray(t *testing.T) {
	ctx := smt.NewZ3Context()
	ca := ctx.ConstArray(ctx.IntSort(), ctx.IntVal(7))

	// Select at any index from K(Int, 7) must equal 7
	slv := ctx.NewZ3Solver()
	x := ctx.Const("x", ctx.IntSort())
	sel := ctx.Select(ca, x)
	slv.Assert(ctx.Not(ctx.Eq(sel, ctx.IntVal(7))))
	assertUnsat(t, slv, "Select(K(Int,7), x) == 7 for all x")
}

// --- Store ---

func TestStoreSelectSameIndex(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	// Store then select at same index
	idx := ctx.IntVal(7)
	val := ctx.IntVal(42)
	aPrime := ctx.Store(a, idx, val)
	sel := ctx.Select(aPrime, idx)

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(sel, val)))
	assertUnsat(t, slv, "Select(Store(a,7,42),7) == 42")
}

func TestStoreSelectDifferentIndex(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	// Store at index 7, select at index 3 — should equal original a[3]
	aPrime := ctx.Store(a, ctx.IntVal(7), ctx.IntVal(42))
	selOrig := ctx.Select(a, ctx.IntVal(3))
	selNew := ctx.Select(aPrime, ctx.IntVal(3))

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(selOrig, selNew)))
	assertUnsat(t, slv, "Store at 7 doesn't affect index 3")
}

func TestStoreOverwrite(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	// Store 10 at index 0, then store 20 at index 0 — select should give 20
	a1 := ctx.Store(a, ctx.IntVal(0), ctx.IntVal(10))
	a2 := ctx.Store(a1, ctx.IntVal(0), ctx.IntVal(20))
	sel := ctx.Select(a2, ctx.IntVal(0))

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Not(ctx.Eq(sel, ctx.IntVal(20))))
	assertUnsat(t, slv, "Second store overwrites first at same index")
}

func TestStoreMultipleIndices(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	// Store different values at different indices
	a1 := ctx.Store(a, ctx.IntVal(1), ctx.IntVal(10))
	a2 := ctx.Store(a1, ctx.IntVal(2), ctx.IntVal(20))
	a3 := ctx.Store(a2, ctx.IntVal(3), ctx.IntVal(30))

	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.Select(a3, ctx.IntVal(1)), ctx.IntVal(10)))
	slv.Assert(ctx.Eq(ctx.Select(a3, ctx.IntVal(2)), ctx.IntVal(20)))
	slv.Assert(ctx.Eq(ctx.Select(a3, ctx.IntVal(3)), ctx.IntVal(30)))
	assertSat(t, slv, "Three stores at distinct indices")
}

// --- ConstArray ---

func TestConstArrayBoolRange(t *testing.T) {
	ctx := smt.NewZ3Context()
	ca := ctx.ConstArray(ctx.IntSort(), ctx.BoolVal(true))

	// Every index maps to true
	slv := ctx.NewZ3Solver()
	x := ctx.Const("x", ctx.IntSort())
	slv.Assert(ctx.Not(ctx.Select(ca, x)))
	assertUnsat(t, slv, "K(Int, true)[x] is always true")
}

func TestConstArrayThenStore(t *testing.T) {
	ctx := smt.NewZ3Context()
	ca := ctx.ConstArray(ctx.IntSort(), ctx.IntVal(0))

	// Start with all-zero array, store 99 at index 5
	ca2 := ctx.Store(ca, ctx.IntVal(5), ctx.IntVal(99))

	slv := ctx.NewZ3Solver()
	// Index 5 should be 99
	slv.Assert(ctx.Eq(ctx.Select(ca2, ctx.IntVal(5)), ctx.IntVal(99)))
	// Index 3 should still be 0
	slv.Assert(ctx.Eq(ctx.Select(ca2, ctx.IntVal(3)), ctx.IntVal(0)))
	assertSat(t, slv, "ConstArray then Store")
}

// --- Array equality ---

func TestArrayExtensionality(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)
	b := ctx.Const("b", arrSort)

	// If two arrays are equal, selecting at the same index gives the same value
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(a, b))
	x := ctx.Const("x", ctx.IntSort())
	slv.Assert(ctx.Not(ctx.Eq(ctx.Select(a, x), ctx.Select(b, x))))
	assertUnsat(t, slv, "Equal arrays have equal elements (extensionality)")
}

// --- ParseArraySortName ---

func TestParseArraySortName(t *testing.T) {
	tests := []struct {
		name    string
		wantDom string
		wantRng string
		wantOk  bool
	}{
		{"arr[int][bool]", "int", "bool", true},
		{"arr[node][int]", "node", "int", true},
		{"arr[arr[int][bool]][int]", "arr[int][bool]", "int", true},
		{"arr[int][arr[bool][int]]", "int", "arr[bool][int]", true},
		{"arr[bv[32]][int]", "bv[32]", "int", true},
		{"bv[32]", "", "", false},
		{"arrcst", "", "", false},
		{"arr[", "", "", false},
		{"arr[][]", "", "", false},
		{"arr[int]", "", "", false},
		{"", "", "", false},
		{"array(int,bool)", "", "", false},
		{"arr[int][bool]extra", "", "", false},
	}
	for _, tt := range tests {
		dom, rng, ok := ParseArraySortName(tt.name)
		if ok != tt.wantOk {
			t.Errorf("ParseArraySortName(%q): ok=%v, want %v", tt.name, ok, tt.wantOk)
			continue
		}
		if ok && (dom != tt.wantDom || rng != tt.wantRng) {
			t.Errorf("ParseArraySortName(%q) = (%q, %q), want (%q, %q)",
				tt.name, dom, rng, tt.wantDom, tt.wantRng)
		}
	}
}

// --- TranslateSort with array sort names ---
/*
func TestTranslateSortArrayIntBool(t *testing.T) {
	tr := NewTranslator()
	defer tr.Close()

	// Without LookupNative, arr[int][bool] is just an uninterpreted sort name.
	// Python requires sig.interp to interpret sort names as arrays.
	// Install a LookupNative callback that resolves arr[...] patterns.
	tr.LookupNative = func(name string, sort logic.Sort, kind string) any {
		if kind == "sort" {
			if dom, rng, ok := ParseArraySortName(name); ok {
				domSort, err1 := tr.TranslateSort(&logic.UninterpretedSort{Name: dom})
				rngSort, err2 := tr.TranslateSort(&logic.UninterpretedSort{Name: rng})
				if err1 == nil && err2 == nil {
					return tr.Ctx.ArraySort(domSort, rngSort)
				}
			}
		}
		return nil
	}

	arrIvySort := &logic.UninterpretedSort{Name: "arr[int][bool]"}
	z3s, err := tr.TranslateSort(arrIvySort)
	if err != nil {
		t.Fatalf("TranslateSort(arr[int][bool]) failed: %v", err)
	}
	if z3s.Kind() != smt.SortArray {
		t.Fatalf("expected SortArray, got %d", z3s.Kind())
	}
}

func TestTranslateSortArrayCached(t *testing.T) {
	tr := NewTranslator()
	defer tr.Close()

	// Install LookupNative to resolve arr[...] patterns.
	tr.LookupNative = func(name string, sort logic.Sort, kind string) any {
		if kind == "sort" {
			if dom, rng, ok := ParseArraySortName(name); ok {
				domSort, err1 := tr.TranslateSort(&logic.UninterpretedSort{Name: dom})
				rngSort, err2 := tr.TranslateSort(&logic.UninterpretedSort{Name: rng})
				if err1 == nil && err2 == nil {
					return tr.Ctx.ArraySort(domSort, rngSort)
				}
			}
		}
		return nil
	}

	arrIvySort := &logic.UninterpretedSort{Name: "arr[node][value]"}

	z3s1, err := tr.TranslateSort(arrIvySort)
	if err != nil {
		t.Fatal(err)
	}
	z3s2, err := tr.TranslateSort(arrIvySort)
	if err != nil {
		t.Fatal(err)
	}
	if !z3s1.Equal(z3s2) {
		t.Fatal("repeated TranslateSort should return cached (equal) sort")
	}
}
*/

func TestTranslateSortNonArray(t *testing.T) {
	tr := NewSolver(nil, nil).NewTranslator()
	defer tr.Close()

	// Regular uninterpreted sort should still work
	s, err := tr.TranslateSort(&UninterpretedSort{Name: "node"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Kind() != smt.SortUninterpreted {
		t.Fatalf("expected SortUninterpreted, got %d", s.Kind())
	}
}

// --- Symbolic array solving ---

func TestArraySymbolicSolve(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	// Find array a such that a[0]=1, a[1]=2, a[2]=3
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.Select(a, ctx.IntVal(0)), ctx.IntVal(1)))
	slv.Assert(ctx.Eq(ctx.Select(a, ctx.IntVal(1)), ctx.IntVal(2)))
	slv.Assert(ctx.Eq(ctx.Select(a, ctx.IntVal(2)), ctx.IntVal(3)))
	assertSat(t, slv, "array with specific values at 3 indices")
}

func TestArraySymbolicUnsatConflict(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)

	// a[0] = 1 AND a[0] = 2 is contradictory
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(ctx.Select(a, ctx.IntVal(0)), ctx.IntVal(1)))
	slv.Assert(ctx.Eq(ctx.Select(a, ctx.IntVal(0)), ctx.IntVal(2)))
	assertUnsat(t, slv, "conflicting values at same index")
}

// --- Array with quantifiers ---

func TestArrayForAllSelect(t *testing.T) {
	ctx := smt.NewZ3Context()
	arrSort := ctx.ArraySort(ctx.IntSort(), ctx.IntSort())
	a := ctx.Const("a", arrSort)
	ca := ctx.ConstArray(ctx.IntSort(), ctx.IntVal(5))

	// Assert a == K(Int, 5), then forall x: a[x] == 5
	slv := ctx.NewZ3Solver()
	slv.Assert(ctx.Eq(a, ca))
	x := ctx.Const("x", ctx.IntSort())
	body := ctx.Eq(ctx.Select(a, x), ctx.IntVal(5))
	// Negate forall to check validity: NOT(forall x: a[x]==5) should be unsat
	slv.Assert(ctx.Not(ctx.ForAll([]smt.Z3Expr{x}, body)))
	assertUnsat(t, slv, "K(Int,5) satisfies forall x: a[x]==5")
}
