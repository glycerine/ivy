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
