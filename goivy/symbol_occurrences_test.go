package goivy

import "testing"

func TestSymbolOccurrenceOrderAndStop(t *testing.T) {
	a, b, c := NewConst("a", Boolean), NewConst("b", Boolean), NewConst("c", Boolean)
	p := &Apply{Func: a, Terms: []Expr{b, c}}
	shared := &Eq{T1: p, T2: p}
	for limit := 1; limit <= 6; limit++ {
		var got []Expr
		for x := range SymbolsIluAst(shared) {
			got = append(got, x)
			if len(got) == limit {
				break
			}
		}
		want := []Expr{a, b, c, a, b, c}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("symbol %d %v, want %v", i, got[i], want[i])
			}
		}
		if len(got) != limit {
			t.Fatal("early stop", limit, len(got))
		}
	}
	captured := &Eq{T1: a, T2: b}
	var got []Expr
	for x := range SymbolsIluAst(captured) {
		got = append(got, x)
		captured.T2 = c
	}
	if len(got) != 2 || got[0] != a || got[1] != b {
		t.Fatal("children changed during traversal", got)
	}
}
