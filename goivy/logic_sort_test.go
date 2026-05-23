package goivy

import (
	"errors"
	"testing"
)

func TestLogicUninterpretedSort(t *testing.T) {
	s1 := &UninterpretedSort{Name: "S"}
	s2 := &UninterpretedSort{Name: "S"}
	s3 := &UninterpretedSort{Name: "T"}

	if s1.String() != "S" {
		t.Errorf("String() = %q, want %q", s1.String(), "S")
	}
	if !SortEqual(s1, s2) {
		t.Error("Equal sorts should be equal")
	}
	if SortEqual(s1, s3) {
		t.Error("Different sorts should not be equal")
	}
	if SortEqual(s1, Boolean) {
		t.Error("UninterpretedSort should not equal BooleanSort")
	}
}

func TestLogicBooleanSort(t *testing.T) {
	b := Boolean
	if b.String() != "Boolean" {
		t.Errorf("String() = %q, want %q", b.String(), "Boolean")
	}
	b2 := &BooleanSort{}
	if !SortEqual(b, b2) {
		t.Error("BooleanSorts should be equal")
	}
}

func TestLogicFunctionSortBasic(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs, err := NewFunctionSort(S, S, Boolean)
	if err != nil {
		t.Fatal(err)
	}
	if fs.Arity() != 2 {
		t.Errorf("Arity() = %d, want 2", fs.Arity())
	}
	if fs.Range().String() != "Boolean" {
		t.Errorf("Range() = %s, want Boolean", fs.Range())
	}
	domain := fs.Domain()
	if len(domain) != 2 || domain[0].String() != "S" || domain[1].String() != "S" {
		t.Errorf("Domain() = %v, want [S, S]", domain)
	}
	if fs.String() != "S * S -> Boolean" {
		t.Errorf("String() = %q, want %q", fs.String(), "S * S -> Boolean")
	}
}

func TestLogicFunctionSortZeroArgs(t *testing.T) {
	_, err := NewFunctionSort()
	if err == nil {
		t.Error("Expected error for zero args")
	}
	var ivyErr *IvyError
	if !errors.As(err, &ivyErr) {
		t.Error("Expected IvyError")
	}
}

func TestLogicFunctionSortHigherOrder(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	inner, _ := NewFunctionSort(S, Boolean)
	_, err := NewFunctionSort(inner, Boolean)
	if err == nil {
		t.Error("Expected error for higher-order args")
	}
}

func TestLogicFunctionSortEquality(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs1, _ := NewFunctionSort(S, S, Boolean)
	fs2, _ := NewFunctionSort(S, S, Boolean)
	T := &UninterpretedSort{Name: "T"}
	fs3, _ := NewFunctionSort(T, S, Boolean)

	if !SortEqual(fs1, fs2) {
		t.Error("Same FunctionSorts should be equal")
	}
	if SortEqual(fs1, fs3) {
		t.Error("Different FunctionSorts should not be equal")
	}
}

func TestLogicEnumeratedSort(t *testing.T) {
	es := &LogicEnumeratedSort{Name: "Color", Extension: []string{"red", "green", "blue"}}
	if es.Card() != 3 {
		t.Errorf("Card() = %d, want 3", es.Card())
	}
	if es.String() != "{red,green,blue}" {
		t.Errorf("String() = %q, want %q", es.String(), "{red,green,blue}")
	}
	es2 := &LogicEnumeratedSort{Name: "Color", Extension: []string{"red", "green", "blue"}}
	if !SortEqual(es, es2) {
		t.Error("Same EnumeratedSorts should be equal")
	}
}

func TestLogicRangeSort(t *testing.T) {
	rs := &RangeSort{Name: "idx", Lb: NumeralBound{Value: "0"}, Ub: NumeralBound{Value: "10"}}
	if rs.String() != "{0 .. 10}" {
		t.Errorf("String() = %q, want %q", rs.String(), "{0 .. 10}")
	}
	rs2 := &RangeSort{Name: "idx", Lb: NumeralBound{Value: "0"}, Ub: NumeralBound{Value: "10"}}
	if !SortEqual(rs, rs2) {
		t.Error("Same RangeSorts should be equal")
	}
}

// TestLogicRangeSortSortedNumeralBound covers the post-sort_infer case where
// each NumeralBound carries the inferred Sort. This mirrors Python's
// compile_bound, which returns lg.Const("0", client_sort); lg.Const.__str__
// (via pretty_fmla → constUgly with show_numeral_sorts=True) prints the
// numeral as "0:client". RangeSort.String() and Sexp() must include that
// annotation so that:
//
//   - compiler.CompileTheory ENTER traces match Python's
//     "theoryname={0:client .. 2:client}" (ivy_compiler.py:2382-2387).
//   - The Module's sig.interp canon embeds "lb:0:client ub:2:client"
//     (matches Python's _range_sort_sexp in logic_sexp.py:32-34).
//
// LbString/UbString must still return the bare value because z3bridge and
// mc_helpers parse it as an integer.
func TestLogicRangeSortSortedNumeralBound(t *testing.T) {
	clientSort := &UninterpretedSort{Name: "client"}
	rs := &RangeSort{
		Name: "client",
		Lb:   NumeralBound{Value: "0", Sort: clientSort},
		Ub:   NumeralBound{Value: "2", Sort: clientSort},
	}
	if got, want := rs.String(), "{0:client .. 2:client}"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if got, want := string(rs.Sexp()), "(RangeSort name:client lb:0:client ub:2:client)"; got != want {
		t.Errorf("Sexp() = %q, want %q", got, want)
	}
	if got, want := rs.LbString(), "0"; got != want {
		t.Errorf("LbString() = %q, want %q (must remain bare for integer-parsing consumers)", got, want)
	}
	if got, want := rs.UbString(), "2"; got != want {
		t.Errorf("UbString() = %q, want %q (must remain bare for integer-parsing consumers)", got, want)
	}
}

func TestLogicTopSort(t *testing.T) {
	ts := TopS.(*TopSort)
	if ts.String() != "TopSort" {
		t.Errorf("String() = %q, want %q", ts.String(), "TopSort")
	}
	if ts.IsSortVariable() {
		t.Error("Default TopSort should not be a sort variable")
	}
	sv := &TopSort{Name: "Alpha"}
	if !sv.IsSortVariable() {
		t.Error("Named TopSort should be a sort variable")
	}
	ts2 := &TopSort{Name: "TopSort"}
	if !SortEqual(TopS, ts2) {
		t.Error("TopSorts with same name should be equal")
	}
}

func TestLogicFirstOrderSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	fs, _ := NewFunctionSort(S, Boolean)

	if !FirstOrderSort(S) {
		t.Error("UninterpretedSort should be first-order")
	}
	if !FirstOrderSort(Boolean) {
		t.Error("BooleanSort should be first-order")
	}
	if FirstOrderSort(fs) {
		t.Error("FunctionSort should not be first-order")
	}
	if !FirstOrderSort(TopS) {
		t.Error("TopSort should be first-order")
	}
}

func TestLogicContainsTopSort(t *testing.T) {
	S := &UninterpretedSort{Name: "S"}
	X, _ := NewVariable("X", TopS)
	Z, _ := NewVariable("Z", S)

	if !ContainsTopSort(X) {
		t.Error("Variable with TopSort should contain TopSort")
	}
	if ContainsTopSort(Z) {
		t.Error("Variable with S should not contain TopSort")
	}

	// f: TopS * TopS -> Boolean
	f := NewConst("f", logicMustFuncSort(t, TopS, TopS, Boolean))
	if !ContainsTopSort(f) {
		t.Error("Symbol with TopSort in sort should contain TopSort")
	}

	// g: S * S -> Boolean
	g := NewConst("g", logicMustFuncSort(t, S, S, Boolean))
	if ContainsTopSort(g) {
		t.Error("Symbol without TopSort should not contain TopSort")
	}

	// h: TopS
	h := NewConst("h", TopS)
	if !ContainsTopSort(h) {
		t.Error("Symbol with TopSort sort should contain TopSort")
	}
}

func TestLogicIsBooleanOrTop(t *testing.T) {
	if !IsBooleanOrTop(Boolean) {
		t.Error("Boolean should be BooleanOrTop")
	}
	if !IsBooleanOrTop(TopS) {
		t.Error("TopS should be BooleanOrTop")
	}
	S := &UninterpretedSort{Name: "S"}
	if IsBooleanOrTop(S) {
		t.Error("UninterpretedSort should not be BooleanOrTop")
	}
}

func FuzzUninterpretedSort(f *testing.F) {
	f.Add("foo")
	f.Add("")
	f.Add("S")
	f.Fuzz(func(t *testing.T, name string) {
		s := &UninterpretedSort{Name: name}
		// Reflexivity
		if !SortEqual(s, s) {
			t.Error("Sort should equal itself")
		}
		s2 := &UninterpretedSort{Name: name}
		if !SortEqual(s, s2) {
			t.Error("Sorts with same name should be equal")
		}
		_ = s.String()
	})
}

func FuzzNewFunctionSort(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(5)
	f.Add(10)
	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 || n > 20 {
			return
		}
		sorts := make([]Sort, n)
		for i := range sorts {
			sorts[i] = &UninterpretedSort{Name: "S"}
		}
		fs, err := NewFunctionSort(sorts...)
		if n == 0 {
			if err == nil {
				t.Error("Expected error for 0 sorts")
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if fs.Arity() != n-1 {
				t.Errorf("Arity = %d, want %d", fs.Arity(), n-1)
			}
		}
	})
}

// dup of logicMustFuncSort
// func mustFS(t *testing.T, sorts ...Sort) *LogicFunctionSort {
// 	t.Helper()
// 	fs, err := NewFunctionSort(sorts...)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	return fs
// }
