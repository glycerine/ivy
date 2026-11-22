package goivy

import (
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
