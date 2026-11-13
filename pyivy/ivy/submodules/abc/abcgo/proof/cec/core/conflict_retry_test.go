package core

import "testing"

// An exhausted query must remain undecided, not add an unproved implication,
// and a later unlimited retry on the same incremental solver can finish.
func TestConflictExhaustionAndUnlimitedRetry(t *testing.T) {
	c := newCircuit(8)
	left, right := c.inputs[0], c.inputs[7]
	for i := 1; i < 8; i++ {
		left = c.add(Xor, left, c.inputs[i], 0)
	}
	for i := 6; i >= 0; i-- {
		right = c.add(Xor, right, c.inputs[i], 0)
	}
	m, err := NewManager(c, Options{ConflictLimit: 1})
	if err != nil {
		t.Fatal(err)
	}
	first, err := m.CheckPair(left, right)
	if err != nil || first.Status != Undecided || first.Counterexample != nil {
		t.Fatalf("budgeted query %+v err=%v", first, err)
	}
	if m.Statistics().Undecided == 0 {
		t.Fatal("query did not reach the solver resource limit")
	}
	if err = m.SetConflictLimit(0); err != nil {
		t.Fatal(err)
	}
	retry, err := m.CheckPair(left, right)
	if err != nil || retry.Status != Equivalent {
		t.Fatalf("retry %+v err=%v", retry, err)
	}
}

// A conjunction interrupted by the conflict budget must not learn its
// negation. Retrying must still find a witness for a satisfiable query and
// prove the complementary, unsatisfiable query on the same solver.
func TestConjunctionConflictExhaustionAndRetry(t *testing.T) {
	c := newCircuit(8)
	left, right := c.inputs[0], c.inputs[7]
	for i := 1; i < 8; i++ {
		left = c.add(Xor, left, c.inputs[i], 0)
	}
	for i := 6; i >= 0; i-- {
		right = c.add(Xor, right, c.inputs[i], 0)
	}
	m, err := NewManager(c, Options{ConflictLimit: 1})
	if err != nil {
		t.Fatal(err)
	}
	first, err := m.CheckConjunction(left, right^1)
	if err != nil || first.Status != Undecided || first.Counterexample != nil {
		t.Fatalf("budgeted conjunction %+v err=%v", first, err)
	}
	if err := m.SetConflictLimit(0); err != nil {
		t.Fatal(err)
	}
	unsat, err := m.CheckConjunction(left, right^1)
	if err != nil || unsat.Status != Equivalent {
		t.Fatalf("retried unsatisfiable conjunction %+v err=%v", unsat, err)
	}
	sat, err := m.CheckConjunction(left, right)
	if err != nil || sat.Status != Different || sat.Counterexample == nil {
		t.Fatalf("satisfiable conjunction %+v err=%v", sat, err)
	}
	for _, lit := range []int32{left, right} {
		value, err := m.Model().Value(sat.Counterexample.Inputs, lit)
		if err != nil || !value {
			t.Fatalf("witness does not satisfy literal %d: %v err=%v", lit, value, err)
		}
	}
}
