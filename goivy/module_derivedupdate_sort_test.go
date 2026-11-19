package goivy

import "testing"

// Python DerivedUpdate/NamedUpdate.get_update_axioms (ivy_actions.py):
//
//	if defines not in updated and any(x in self.dependencies for x in updated):
//
// compares Symbols, which are equal only when name and sort both match
// (self.dependencies is used_symbols_ast, keyed like goivy's Key). After
// strip_isolate gives contains the stripped sort key.t -> bool, a
// definition compiled before stripping (depending on contains with sort
// id * key.t -> bool) no longer fires. goivy compared names only, so
// sht.ivy isolate iso_dm added imap.gap to every update of imap.contains.
func updateSortFixture(t *testing.T) (gap, containsOld, containsNew *Const, body Expr) {
	t.Helper()
	id, key, it := isolateMkSort("id"), isolateMkSort("key.t"), isolateMkSort("key.iter.t")
	mk := func(name string, sorts ...Sort) *Const {
		s, err := NewFunctionSort(sorts...)
		if err != nil {
			t.Fatal(err)
		}
		return NewConst(name, s)
	}
	gap = mk("gap", id, it, it, Boolean)
	containsOld = mk("contains", id, key, Boolean) // compiled, unstripped
	containsNew = mk("contains", key, Boolean)     // after strip_isolate
	x, err := NewVariable("X", id)
	if err != nil {
		t.Fatal(err)
	}
	return gap, containsOld, containsNew, MustApply(containsOld, x, NewConst("k", key))
}

func TestDerivedUpdateMatchesSymbolsBySort(t *testing.T) {
	gap, containsOld, containsNew, rhs := updateSortFixture(t)
	it := isolateMkSort("key.iter.t")
	x, err := NewVariable("X", isolateMkSort("id"))
	if err != nil {
		t.Fatal(err)
	}
	du := NewDerivedUpdate(gap, NewDefinition(MustApply(gap, x, NewConst("i", it), NewConst("j", it)), rhs))
	if got, _, _ := du.GetUpdateAxioms([]*Const{containsNew}, nil); len(got) != 1 {
		t.Fatalf("stripped contains (different sort) fired the definition: %v", got)
	}
	// Positive control: the same symbol (name and sort) does fire it.
	if got, _, _ := du.GetUpdateAxioms([]*Const{containsOld}, nil); len(got) != 2 || got[1] != gap {
		t.Fatalf("same-sort contains did not fire the definition: %v", got)
	}
}

func TestNamedUpdateMatchesSymbolsBySort(t *testing.T) {
	gap, containsOld, containsNew, body := updateSortFixture(t)
	nu := NewNamedUpdate(gap, body)
	if got, _, _ := nu.GetUpdateAxioms([]*Const{containsNew}, nil); len(got) != 1 {
		t.Fatalf("stripped contains (different sort) fired the update: %v", got)
	}
	if got, _, _ := nu.GetUpdateAxioms([]*Const{containsOld}, nil); len(got) != 2 || got[1] != gap {
		t.Fatalf("same-sort contains did not fire the update: %v", got)
	}
}
