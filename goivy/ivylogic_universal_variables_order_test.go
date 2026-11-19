package goivy

import "testing"

// Python universal_variables (ivy_logic.py) now collects into an
// insertion-ordered dict (it was a set, ordered by PYTHONHASHSEED); goivy
// collected into a Go map, iterated in random order. Both must give the
// first-occurrence order. ForAll stores its variables sorted by name (both
// sides), so the inner quantifier contributes W, X, Z: result X, Y, W, Z.
func TestUniversalVariablesFirstOccurrenceOrder(t *testing.T) {
	s := isolateMkSort("T")
	mk := func(n string) *LogicVariable {
		v, err := NewVariable(n, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	x, y, z, w := mk("X"), mk("Y"), mk("Z"), mk("W")
	inner, err := NewForAll([]*LogicVariable{z, w, x}, True)
	if err != nil {
		t.Fatal(err)
	}
	outer, err := NewForAll([]*LogicVariable{x, y}, True)
	if err != nil {
		t.Fatal(err)
	}
	fmla := &LogicAnd{Terms: []Expr{outer, inner}}
	for i := 0; i < 50; i++ {
		got := UniversalVariables([]Expr{fmla})
		want := []*LogicVariable{x, y, w, z}
		if len(got) != len(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		for j := range want {
			if got[j].Name != want[j].Name {
				t.Fatalf("run %d: got %v, want X Y W Z", i, got)
			}
		}
	}
}
