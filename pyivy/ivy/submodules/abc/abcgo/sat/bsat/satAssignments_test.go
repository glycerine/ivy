package bsat

import (
	"reflect"
	"testing"
)

func TestBacktrackRestoresHeapWithoutAllocation(t *testing.T) {
	for _, topLevel := range []bool{false, true} {
		s := sat_solver_new()
		sat_solver_setnvars(s, 4)
		allocs := testing.AllocsPerRun(10, func() {
			if topLevel {
				sat_solver_enqueue(s, toLit(0), 0)
			} else {
				sat_solver_decision(s, toLit(0))
			}
			sat_solver_enqueue(s, lit_neg(toLit(1)), 0)
			sat_solver_enqueue(s, toLit(2), 0)
			s.qhead = s.qtail
			sat_solver_enqueue(s, toLit(3), 0)
			// The first three variables have been propagated and taken
			// out of the heap. The pending fourth variable is still in it.
			s.order.size = 1
			s.order.ptr[0] = 3
			copy(s.orderpos, []int32{-1, -1, -1, 0})
			if topLevel {
				sat_solver_canceluntil_rollback(s, 0)
			} else {
				sat_solver_canceluntil(s, 0)
			}
		})
		if got := veci_begin(&s.order); !reflect.DeepEqual(got, []int32{3, 2, 1, 0}) {
			t.Fatalf("topLevel %t restored heap %v", topLevel, got)
		}
		if allocs != 0 {
			t.Errorf("topLevel %t allocated %.0f objects per backtrack", topLevel, allocs)
		}
	}
}

func TestCDCLTrailDecisionAndBacktrack(t *testing.T) {
	a := newSatAssignments(5)
	if !a.enqueue(toLit(0), 0) {
		t.Fatal("root")
	}
	a.qhead = a.qtail
	if !a.decision(toLit(1)) {
		t.Fatal("decision")
	}
	if !a.enqueue(lit_neg(toLit(2)), 22) {
		t.Fatal("implication")
	}
	a.qhead = a.qtail
	if !a.decision(toLit(3)) {
		t.Fatal("decision2")
	}
	if !a.enqueue(toLit(4), 44) {
		t.Fatal("implication2")
	}
	a.qhead = a.qtail
	restored := a.cancelUntil(1)
	if !reflect.DeepEqual(restored, []lit{toLit(3), toLit(4)}) || a.qtail != 3 || a.trail_lim.size != 1 {
		t.Fatal(restored, a)
	}
	if a.assigns[0] != 0 || a.assigns[1] != 0 || a.assigns[2] != 1 || a.assigns[3] != 3 || a.reasons[4] != 0 {
		t.Fatal(a)
	}
	a.cancelUntil(0)
	if a.qtail != 1 || a.assigns[0] != 0 || a.polarity[2] != 0 {
		t.Fatal(a)
	}
}
func TestCDCLConflictingEnqueueAndRollback(t *testing.T) {
	a := newSatAssignments(3)
	if !a.enqueue(toLit(0), 7) || !a.enqueue(toLit(0), 8) || a.enqueue(lit_neg(toLit(0)), 9) || a.qtail != 1 || a.reasons[0] != 7 {
		t.Fatal(a)
	}
	a.qhead = a.qtail
	bound := a.qtail
	a.enqueue(toLit(1), 2)
	a.qhead = a.qtail
	if got := a.rollback(bound); !reflect.DeepEqual(got, []lit{toLit(1)}) || a.assigns[1] != 3 || a.reasons[1] != 0 {
		t.Fatal(got, a)
	}
}
func TestCDCLUnpropagatedHeapRestore(t *testing.T) {
	a := newSatAssignments(2)
	a.decision(toLit(0))
	a.enqueue(toLit(1), 2)
	if got := a.cancelUntil(0); len(got) != 0 || a.qtail != 0 || a.assigns[0] != 3 || a.assigns[1] != 3 {
		t.Fatal(got, a)
	}
}

func TestFailedAssumptionBeforeRootPropagation(t *testing.T) {
	s := NewSolver()
	s.SetNumVars(1)
	s.AddClause(Positive(0))
	if s.qhead != 0 || s.qtail != 1 {
		t.Fatal("expected an unpropagated root assignment")
	}
	if s.Solve(Negative(0)) != Unsatisfiable || s.qhead != 1 || s.qtail != 1 {
		t.Fatal("failed assumption did not restore the root trail")
	}
	if s.Solve() != Satisfiable || s.assigns[0] != var1 {
		t.Fatal("backtracking lost the root unit")
	}
}
func TestCDCLPhaseSavingEarlierLevels(t *testing.T) {
	a := newSatAssignments(3)
	a.decision(toLit(0))
	a.enqueue(lit_neg(toLit(1)), 2)
	a.qhead = a.qtail
	a.decision(toLit(2))
	a.qhead = a.qtail
	a.cancelUntil(0)
	if a.polarity[0] != 1 || a.polarity[1] != 0 {
		t.Fatal("phase saving", a.polarity)
	}
}
func TestCDCLTrailGuards(t *testing.T) {
	for _, f := range []func(){func() { a := newSatAssignments(1); a.enqueue(toLit(0), 0); a.decision(toLit(0)) }, func() { a := newSatAssignments(1); a.rollback(1) }, func() { a := newSatAssignments(1); a.enqueue(toLit(1), 0) }, func() { a := newSatAssignments(1); a.cancelUntil(-1) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("expected invariant failure")
				}
			}()
			f()
		}()
	}
}
