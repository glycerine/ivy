package minisatp

import (
	"reflect"
	"testing"
)

func TestProofRecordsRootsAndResolutionInInsertionOrder(t *testing.T) {
	var p proof
	a := p.addRoot([]int32{4, 0, 4})
	b := p.addRoot([]int32{1})
	p.beginChain(a)
	p.resolve(b, 0)
	c := p.endChain()
	if a != 0 || b != 1 || c != 2 {
		t.Fatalf("IDs = %d, %d, %d", a, b, c)
	}
	if !reflect.DeepEqual(p.events[0].root, []int32{0, 4}) {
		t.Fatalf("root = %v", p.events[0].root)
	}
	want := Chain{Clauses: []ClauseID{a, b}, Pivots: []int32{0}}
	if !reflect.DeepEqual(p.events[2].chain, want) {
		t.Fatalf("chain = %+v, want %+v", p.events[2].chain, want)
	}
	p.beginChain(c)
	if got := p.endChain(); got != c {
		t.Fatalf("single-clause chain = %d, want %d", got, c)
	}
	p.deleted(a)
	if p.next != 3 || !p.events[3].deleted {
		t.Fatalf("deletion changed clause numbering: %+v", p)
	}
}

func TestRootClauseNormalizationAndConflictProof(t *testing.T) {
	s := newSolver(2)
	if !s.addClause([]int32{0, 0}) || !s.addClause([]int32{2, 3}) {
		t.Fatal("consistent clauses rejected")
	}
	if len(s.proof.events) != 1 {
		t.Fatalf("tautology should not be logged: %+v", s.proof.events)
	}
	if s.addClause([]int32{1}) {
		t.Fatal("contradictory unit accepted")
	}
	if len(s.proof.events) != 3 {
		t.Fatalf("expected two roots and a resolution: %+v", s.proof.events)
	}
	got := s.proof.events[2].chain
	want := Chain{Clauses: []ClauseID{1, 0}, Pivots: []int32{0}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("conflict chain = %+v, want %+v", got, want)
	}
}

func TestRootPropagationProducesResolutionRefutation(t *testing.T) {
	s := newSolver(2)
	for _, c := range [][]int32{{1, 2}, {3}, {0}} {
		if !s.addClause(c) {
			t.Fatalf("unexpected immediate conflict adding %v", c)
		}
	}
	if s.propagate() == nil || s.ok {
		t.Fatal("expected root propagation conflict")
	}
	derived := s.proof.events[len(s.proof.events)-1].chain
	if !reflect.DeepEqual(derived.Pivots, []int32{0}) {
		t.Fatalf("final resolution = %+v", derived)
	}
}

func TestConflictAnalysisRecordsFirstUIPResolution(t *testing.T) {
	s := newSolver(2)
	if !s.addClause([]int32{0, 2}) || !s.addClause([]int32{0, 3}) {
		t.Fatal("unexpected clause conflict")
	}
	s.limits = append(s.limits, len(s.trail))
	s.enqueue(1, nil) // x=false implies y and not y.
	c := s.propagate()
	if c == nil {
		t.Fatal("expected conflict")
	}
	learnt, level := s.analyze(c)
	if !reflect.DeepEqual(learnt, []int32{0}) || level != 0 {
		t.Fatalf("learnt %v at level %d", learnt, level)
	}
	chain := s.proof.events[len(s.proof.events)-1].chain
	if !reflect.DeepEqual(chain.Pivots, []int32{1}) {
		t.Fatalf("unexpected resolution chain %+v", chain)
	}
}
