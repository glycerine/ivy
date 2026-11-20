package goivy

import "testing"

// Python collect_relevant_destructors: every sort but RangeSort/TopSort has
// .rng (a constant sort's rng is itself), so a constant of struct sort q
// pulls in q's destructors, depth first, appended in discovery order
// (test/annot1.ivy with keep_destructors=true).
func TestAddRelevantDestructorsOrder(t *testing.T) {
	m := mkModule()
	q := isolateMkSort("q")
	r := isolateMkSort("r")
	qb, _ := NewFunctionSort(q, r)
	qa, _ := NewFunctionSort(q, Boolean)
	rz, _ := NewFunctionSort(r, Boolean)
	dqb := NewConst("q.b", qb)
	dqa := NewConst("q.a", qa)
	drz := NewConst("r.z", rz)
	m.SortDestructors.Set("q", []*Const{dqb, dqa})
	m.SortDestructors.Set("r", []*Const{drz})

	self := NewConst("self", q)
	all := NewInsMap[NodeKey, Expr]()
	all.Set(ConstSymKey(self), self)
	addRelevantDestructors(m, all)

	var got []string
	for _, s := range all.All() {
		got = append(got, s.(*Const).Name)
	}
	want := []string{"self", "q.b", "r.z", "q.a"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if c, _ := all.Get(ConstSymKey(dqb)).(*Const); c != dqb {
		t.Errorf("q.b should be the destructor symbol itself, got %v", all.Get(ConstSymKey(dqb)))
	}
}
