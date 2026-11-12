package bsat

import (
	"reflect"
	"testing"
)

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
	if !reflect.DeepEqual(restored, []int32{4, 3}) || a.qtail != 3 || a.trail_lim.size != 1 {
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
	if got := a.rollback(bound); !reflect.DeepEqual(got, []int32{1}) || a.assigns[1] != 3 || a.reasons[1] != 0 {
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
