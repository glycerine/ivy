package bsat

import (
	"reflect"
	"testing"
)

func TestCDCLFinalAssumptionConflict(t *testing.T) {
	a := newSatAssignments(4)
	m := Sat_MemAlloc(9)
	a.decision(toLit(0))
	a.enqueue(toLit(2), clause_from_lit(toLit(0)))
	a.qhead = a.qtail
	a.decision(toLit(1))
	a.enqueue(toLit(3), clause_from_lit(toLit(1)))
	a.qhead = a.qtail
	h := Sat_MemAppend(m, []lit{lit_neg(toLit(2)), lit_neg(toLit(3))}, 2, 0, 0)
	got := satAnalyzeFinal(&a, m, h, 2, false)
	if !reflect.DeepEqual(got, []lit{lit_neg(toLit(1)), lit_neg(toLit(0))}) || a.tagged.size != 0 {
		t.Fatal(got, a.tags)
	}
	for _, tag := range a.tags {
		if tag != 0 {
			t.Fatal(a.tags)
		}
	}
	if root := satAnalyzeFinal(&a, m, h, 0, false); len(root) != 0 {
		t.Fatal(root)
	}
}
func TestCDCLFinalSkipFirst(t *testing.T) {
	a := newSatAssignments(2)
	m := Sat_MemAlloc(9)
	a.decision(toLit(0))
	a.qhead = a.qtail
	a.decision(toLit(1))
	a.qhead = a.qtail
	h := Sat_MemAppend(m, []lit{lit_neg(toLit(0)), lit_neg(toLit(1))}, 2, 0, 0)
	if got := satAnalyzeFinal(&a, m, h, 2, true); !reflect.DeepEqual(got, []lit{lit_neg(toLit(1))}) {
		t.Fatal(got)
	}
}
