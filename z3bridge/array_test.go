package z3bridge

import "testing"

func TestArraySelectStore(t *testing.T) {
	ctx := NewContext()

	// Create array sort: Array(Int, Int)
	intSort := ctx.IntSort()
	arrSort := ctx.ArraySort(intSort, intSort)

	// Verify sort kind
	if arrSort.Kind() != SortArray {
		t.Fatalf("expected SortArray, got %d", arrSort.Kind())
	}

	// Verify domain/range introspection
	if arrSort.ArrayDomain().Kind() != SortInt {
		t.Fatalf("expected domain SortInt, got %d", arrSort.ArrayDomain().Kind())
	}
	if arrSort.ArrayRange().Kind() != SortInt {
		t.Fatalf("expected range SortInt, got %d", arrSort.ArrayRange().Kind())
	}

	// Create a symbolic array constant
	a := ctx.Const("a", arrSort)

	// Store 42 at index 7: a' = Store(a, 7, 42)
	idx := ctx.IntVal(7)
	val := ctx.IntVal(42)
	aPrime := ctx.Store(a, idx, val)

	// Select at index 7: Select(a', 7) should equal 42
	sel := ctx.Select(aPrime, idx)

	solver := ctx.NewSolver()
	solver.Assert(ctx.Not(ctx.Eq(sel, val)))
	result := solver.Check()
	if result != Unsat {
		t.Fatalf("expected Unsat (Select(Store(a,7,42),7) == 42), got %s", result)
	}
}

func TestConstArray(t *testing.T) {
	ctx := NewContext()

	intSort := ctx.IntSort()
	val := ctx.IntVal(0)

	// Create constant array: K(Int, 0)
	ca := ctx.ConstArray(intSort, val)

	// Select at any index should return 0
	anyIdx := ctx.IntVal(999)
	sel := ctx.Select(ca, anyIdx)

	solver := ctx.NewSolver()
	solver.Assert(ctx.Not(ctx.Eq(sel, val)))
	result := solver.Check()
	if result != Unsat {
		t.Fatalf("expected Unsat (Select(K(Int,0), 999) == 0), got %s", result)
	}
}

func TestParseArraySortName(t *testing.T) {
	tests := []struct {
		name       string
		wantDom    string
		wantRng    string
		wantOk     bool
	}{
		{"arr[int][bool]", "int", "bool", true},
		{"arr[node][int]", "node", "int", true},
		{"arr[arr[int][bool]][int]", "arr[int][bool]", "int", true},
		{"bv[32]", "", "", false},
		{"arrcst", "", "", false},
		{"arr[", "", "", false},
		{"arr[][]", "", "", false},
	}
	for _, tt := range tests {
		dom, rng, ok := ParseArraySortName(tt.name)
		if ok != tt.wantOk {
			t.Errorf("ParseArraySortName(%q): ok=%v, want %v", tt.name, ok, tt.wantOk)
			continue
		}
		if ok && (dom != tt.wantDom || rng != tt.wantRng) {
			t.Errorf("ParseArraySortName(%q) = (%q, %q), want (%q, %q)", tt.name, dom, rng, tt.wantDom, tt.wantRng)
		}
	}
}

func TestArraySortTranslation(t *testing.T) {
	ctx := NewContext()

	// Manually create an array sort and verify roundtrip
	intSort := ctx.IntSort()
	boolSort := ctx.BoolSort()
	arrSort := ctx.ArraySort(intSort, boolSort)

	if arrSort.Kind() != SortArray {
		t.Fatalf("expected SortArray kind")
	}
	if arrSort.ArrayDomain().Kind() != SortInt {
		t.Fatalf("expected domain Int")
	}
	if arrSort.ArrayRange().Kind() != SortBool {
		t.Fatalf("expected range Bool")
	}
}
