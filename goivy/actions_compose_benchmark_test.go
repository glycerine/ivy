package goivy

import (
	"slices"
	"strconv"
	"testing"
)

func TestComposeUpdatesAvoidsIntermediateNameCollision(t *testing.T) {
	for _, site := range []string{"transition", "first precondition", "second precondition"} {
		t.Run(site, func(t *testing.T) {
			first := mkTestUpdate([]string{"x"}, mkEq("new_x", "a"), False)
			second := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), False)
			occupied := FormulaToClauses(mkEq("__m_x", "b"), nil)
			switch site {
			case "transition":
				first.TR = AndClausesTyped(first.TR, occupied)
			case "first precondition":
				first.Pre = occupied
			case "second precondition":
				second.Pre = occupied
			}
			result := ComposeUpdates(first, TrueClauses(nil), second)
			if !UsedSymbolNamesClauses(result.TR)["__m_x_a"] {
				t.Errorf("intermediate state name collided with a symbol in %s: %v", site, result.TR)
			}
		})
	}
}

func BenchmarkComposeUpdatesReadOnlySymbols(b *testing.B) {
	for _, tc := range []struct{ name, second string }{{"Disjoint", "y"}, {"Overlapping", "x"}} {
		b.Run(tc.name, func(b *testing.B) {
			first := mkTestUpdate([]string{"x"}, mkEq("new_x", "x"), False)
			for i := 0; i < 128; i++ {
				first.TR.Fmlas = append(first.TR.Fmlas, mkEq("read"+strconv.Itoa(i), "a"))
			}
			second := mkTestUpdate([]string{tc.second}, mkEq(ActionNewName(tc.second), tc.second), False)
			axioms := TrueClauses(nil)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if result := ComposeUpdates(first, axioms, second); len(result.Modified) == 0 {
					b.Fatal("empty set of modified symbols")
				}
			}
		})
	}
}

func TestComposeIntermediateNamesMatchPublicRenamer(t *testing.T) {
	for _, tc := range []struct {
		name          string
		mid, reserved []string
	}{
		{"unused", []string{"x"}, []string{"x", "__m_other"}},
		{"suffix gap", []string{"x"}, []string{"__m_x", "__m_x_a", "__m_x_c", "__m_x_z", "__m_x_a"}},
		{"empty spelling", []string{""}, []string{"__m__a", "__m__b"}},
		{"intersecting bases", []string{"x", "x_a", "x"}, []string{"__m_x", "__m_x_b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var groups []*Clauses
			for _, name := range tc.reserved {
				groups = append(groups, FormulaToClauses(mkEq(name, "read"), nil))
			}
			var symbols []*Const
			reference := NewUniqueRenamer("__m_", tc.reserved)
			var expected []string
			for _, name := range tc.mid {
				symbols = append(symbols, &Const{Name: name})
				expected = append(expected, reference.Rename(name))
			}
			got := composeIntermediateNames(symbols, groups...)
			if !slices.Equal(got, expected) {
				t.Fatalf("intermediate names %v, want %v", got, expected)
			}
		})
	}
}
