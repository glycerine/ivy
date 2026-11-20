package goivy

import (
	"strings"
	"testing"
)

// Python HavocAction.action_update builds
//
//	And(*([Or(Not(new_n(vs)),n(vs),eq) for eq in eqs] +
//	      [Or(new_n(vs),Not(n(vs)),eq) for eq in eqs]))
//
// so all the ~new|old clauses precede all the new|~old clauses
// (test/learning_switch1.ivy: pending(p,sw0,sw1) := *).
func TestHavocRelationClauseOrder(t *testing.T) {
	s := proofMkSort("s")
	rSort, _ := NewFunctionSort(s, s, Boolean)
	r := NewConst("r", rSort)
	a := NewConst("a", s)
	b := NewConst("b", s)
	lhs, _ := NewApply(r, a, b)
	u := NewHavocAction(lhs).ActionUpdate(testCtx())
	if len(u.TR.Fmlas) != 4 {
		t.Fatalf("want 4 clauses, got %d: %v", len(u.TR.Fmlas), u.TR.Fmlas)
	}
	for i, f := range u.TR.Fmlas {
		s := f.String()
		wantNegNew := i < 2
		if gotNegNew := strings.HasPrefix(s, "~new_r("); gotNegNew != wantNegNew {
			t.Errorf("clause %d = %s; want ~new_r first: %v", i, s, wantNegNew)
		}
		wantEq := []string{"X0 = a", "X1 = b"}[i%2]
		if !strings.Contains(s, wantEq) {
			t.Errorf("clause %d = %s; want %s", i, s, wantEq)
		}
	}
}
