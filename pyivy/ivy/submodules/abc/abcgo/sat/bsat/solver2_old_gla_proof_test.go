package bsat

import (
	"reflect"
	"testing"
)

func TestOldGLAProofModesAndLifecycle(t *testing.T) {
	for _, full := range []bool{false, true} {
		p := NewLabeledSolver2ProofMode(full)
		if (p.solver.solver.pPrf1 != nil) != full || p.IncrementalProofActive() {
			t.Fatalf("full=%v initial proof state", full)
		}
		if !p.AddClause(1, Negative(1), Positive(2)) {
			t.Fatal("background clause")
		}
		p.Simplify()
		p.Bookmark()
		if !full {
			p.StartIncrementalProof(4, 1)
			p.RegisterIncrementalProofID(2)
		}
		if !p.AddClause(2, Negative(2)) {
			t.Fatal("refinement clause")
		}
		if got := p.Solve(Positive(1)); got != Unsatisfiable {
			t.Fatalf("full=%v result=%v", full, got)
		}
		want := []int32{2}
		if full {
			want = []int32{2, 1}
		}
		if got := p.ProofCore(); !reflect.DeepEqual(got, want) {
			t.Fatalf("full=%v core=%v want=%v", full, got, want)
		}
		p.StopIncrementalProof()
		p.Rollback()
		if p.IncrementalProofActive() || p.Solve(Positive(1)) != Satisfiable {
			t.Fatalf("full=%v: failed to stop or roll back", full)
		}
	}
}

func TestOldGLAIncrementalProofGrowsAcrossWords(t *testing.T) {
	p := NewLabeledSolver2ProofMode(false)
	if !p.AddClause(1, Negative(1), Positive(2)) {
		t.Fatal("background clause")
	}
	p.Simplify()
	p.Bookmark()
	p.StartIncrementalProof(80, 1)
	for id := int32(2); id < 70; id++ {
		p.GrowIncrementalProof(1)
		p.RegisterIncrementalProofID(id)
	}
	if !p.AddClause(69, Negative(2)) || p.Solve(Positive(1)) != Unsatisfiable {
		t.Fatal("expected unsatisfiable refinement")
	}
	if got := p.ProofCore(); !reflect.DeepEqual(got, []int32{69}) {
		t.Fatalf("core after growth = %v", got)
	}
	p.StopIncrementalProof()
	p.Rollback()
}
