package core

import "testing"

func TestVectorComposeSimultaneousAndOwnership(t *testing.T) {
	dd := NewBDDManager(3)
	f := Cudd_bddAnd(dd, dd.Vars[0], Cudd_Not(dd.Vars[1]))
	vector := []*DdNode{dd.Vars[1], dd.Vars[0], Cudd_bddXor(dd, dd.Vars[0], dd.Vars[2])}
	got := Cudd_bddVectorCompose(dd, f, vector)
	want := Cudd_bddAnd(dd, dd.Vars[1], Cudd_Not(dd.Vars[0]))
	if got != want {
		t.Fatal("composition was not simultaneous")
	}
	other := NewBDDManager(3)
	if Cudd_bddVectorCompose(dd, other.Vars[0], vector) != nil || Cudd_bddVectorCompose(dd, f, vector[:2]) != nil {
		t.Fatal("invalid composition accepted")
	}
	vector[1] = other.Vars[0]
	if Cudd_bddVectorCompose(dd, f, vector) != nil {
		t.Fatal("foreign substitution accepted")
	}
}
func TestVectorComposeAllTruthTables(t *testing.T) {
	dd := NewBDDManager(3)
	vector := []*DdNode{Cudd_bddXor(dd, dd.Vars[1], dd.Vars[2]), Cudd_Not(dd.Vars[0]), Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[2])}
	for table := 0; table < 256; table++ {
		f := bddFromTruth(dd, uint8(table), 3)
		result := Cudd_bddVectorCompose(dd, f, vector)
		for a := 0; a < 8; a++ {
			values := []bool{a&4 != 0, a&2 != 0, a&1 != 0}
			mapped := 0
			for i, g := range vector {
				value, err := dd.Evaluate(g, values)
				if err != nil {
					t.Fatal(err)
				}
				if value {
					mapped |= 1 << uint(i)
				}
			}
			got, err := dd.Evaluate(result, values)
			if err != nil || got != (table>>uint(mapped)&1 != 0) {
				t.Fatal("vector truth", table, a)
			}
		}
	}
}
