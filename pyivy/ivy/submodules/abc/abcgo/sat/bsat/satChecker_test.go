package bsat

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolutionProof(t *testing.T) {
	proof := "1 1 2 0 0\n2 -1 3 0 0\n3 * 1 2 0\n4 -2 0 0\n5 * 3 4 0\n6 -3 0 0\n7 * 5 6 0"
	s, err := CheckResolutionProof(strings.NewReader(proof))
	if err != nil || s != (ProofSummary{Roots: 4, Steps: 3, Clauses: 7, EmptyDerived: true}) {
		t.Fatal(s, err)
	}
}
func TestResolutionMalformed(t *testing.T) {
	for _, p := range []string{"", "2 1 0 0", "1 1 0", "1 1 0 1", "1 1 1 0 0", "1 -2147483648 0 0", "1 * 1 1 0", "1 1 0 0\n2 2 0 0\n3 * 1 2 0", "1 1 2 0 0\n2 -1 -2 0 0\n3 * 1 2 0", "1 1 0 0\n2 -1 0 0\n3 * 1 2 1"} {
		if _, err := CheckResolutionProof(strings.NewReader(p)); err == nil {
			t.Errorf("accepted %q", p)
		}
	}
}
func TestResolutionUnionOrder(t *testing.T) {
	a, b := []int32{1, 2, 3}, []int32{-1, 3, 4}
	r, err := ResolveProofClauses(a, b)
	if err != nil || !reflect.DeepEqual(r, []int32{2, 3, 4}) {
		t.Fatal(r, err)
	}
	if a[0] != 1 || b[0] != -1 {
		t.Fatal("parents mutated")
	}
}
func TestProofNonemptyConclusion(t *testing.T) {
	s, err := CheckResolutionProof(strings.NewReader("1 1 0 0"))
	if err != nil || s.EmptyDerived {
		t.Fatal(s, err)
	}
}

// A tautological parent must not allow deleting both polarities of the pivot
// and manufacturing an empty clause from a satisfiable root set.
func TestProofRejectsTautologyUnsoundResolution(t *testing.T) {
	proof := "1 1 -1 0 0\n2 1 0 0\n3 * 1 2 0"
	if _, err := CheckResolutionProof(strings.NewReader(proof)); err == nil {
		t.Fatal("accepted unsound proof")
	}
	if _, err := ResolveProofClauses([]int32{1, -1}, []int32{1}); err == nil {
		t.Fatal("accepted tautological parent")
	}
}
