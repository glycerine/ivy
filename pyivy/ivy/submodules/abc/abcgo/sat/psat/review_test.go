package psat

import (
	"bytes"
	"testing"
)

func TestProofReaderRejectsTrailingData(t *testing.T) {
	s := New(true)
	s.AddClause([]int{0})
	s.AddClause([]int{1})
	s.Solve(nil, 0)
	var b bytes.Buffer
	if err := s.Proof().Write(&b); err != nil {
		t.Fatal(err)
	}
	b.WriteString("{}\n")
	if _, err := ReadProof(&b); err == nil {
		t.Fatal("accepted a second object after the proof")
	}
}
func TestFailedRequestPreservesProofAndRootIdentity(t *testing.T) {
	s := New(true)
	raw := []int{2, 0, 2}
	s.AddClause(raw)
	raw[0] = 9
	s.AddClause([]int{1})
	s.AddClause([]int{3})
	if s.Solve(nil, 0) != Unsatisfiable {
		t.Fatal("formula")
	}
	before := s.Proof()
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected invalid assumption")
			}
		}()
		s.Solve([]int{8}, 0)
	}()
	if !s.ProofReady() || s.Proof().Final != before.Final {
		t.Fatal("invalid assumption destroyed prior certificate")
	}
	if s.Proof().Roots[0][0] != 0 || s.Proof().Roots[0][1] != 2 {
		t.Fatal("input ownership/normalization")
	}
	var lits []int
	n := M114p_SolverGetFirstRoot(s, &lits)
	if n != 2 {
		t.Fatal(n, lits)
	}
	lits[0] = 99
	if s.Proof().Roots[0][0] != 0 {
		t.Fatal("legacy iterator aliases producer")
	}
}
func TestEmptyRootNativeProof(t *testing.T) {
	s := New(true)
	if s.AddClause(nil) {
		t.Fatal("empty clause accepted as consistent")
	}
	if s.Solve(nil, 0) != Unsatisfiable || !s.ProofReady() {
		t.Fatal("missing empty-root proof")
	}
	if err := s.Proof().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestProofValidationAnchorsRequestedFormula(t *testing.T) {
	s := New(true)
	roots := [][]int{{0}, {1}}
	for _, root := range roots {
		s.AddClause(root)
	}
	s.Solve(nil, 0)
	proof := s.Proof()
	if err := proof.ValidateAgainst(roots, nil); err != nil {
		t.Fatal(err)
	}
	if proof.ValidateAgainst([][]int{{2}, {3}}, nil) == nil {
		t.Fatal("accepted proof of a different formula")
	}
	if proof.ValidateAgainst(roots, []int{0}) == nil {
		t.Fatal("accepted mismatched assumptions")
	}
}
