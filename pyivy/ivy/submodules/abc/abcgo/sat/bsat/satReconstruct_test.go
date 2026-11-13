package bsat

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestResolutionReconstructionProducesCheckableProof(t *testing.T) {
	s := sat_solver_new()
	sat_solver_store_alloc(s)
	for _, c := range [][]lit{{0, 2}, {1, 2}, {0, 3}, {1, 3}} {
		sat_solver_addclause(s, c)
	}
	sat_solver_store_mark_roots(s)
	if sat_solver_solve(s, nil, 0, 0, 0, 0) != l_False {
		t.Fatal("UNSAT")
	}
	p := solverClauseStore(s)
	var proof bytes.Buffer
	r, err := reconstructResolution(p, nil, &proof, 0)
	if err != nil {
		t.Fatal(err)
	}
	summary, err := CheckResolutionProof(strings.NewReader(proof.String()))
	if err != nil || !summary.EmptyDerived || summary.Roots != 4 || r.labels[p.pEmpty.Id] <= 0 {
		t.Fatal(summary, err, proof.String())
	}
	if len(r.dependencies[p.pEmpty.Id]) == 0 {
		t.Fatal("missing antecedents")
	}
}
func TestResolutionReconstructionRootUnitConflict(t *testing.T) {
	p := Sto_ManAlloc()
	Sto_ManAddClause(p, []lit{0})
	Sto_ManMarkClausesA(p)
	Sto_ManAddClause(p, []lit{1})
	Sto_ManMarkRoots(p)
	Sto_ManAddClause(p, nil)
	var proof bytes.Buffer
	if _, err := reconstructResolution(p, nil, &proof, 0); err != nil {
		t.Fatal(err)
	}
	if summary, err := CheckResolutionProof(&proof); err != nil || !summary.EmptyDerived {
		t.Fatal(summary, err)
	}
}
func TestResolutionReconstructionRejectsInvalidProof(t *testing.T) {
	p := Sto_ManAlloc()
	Sto_ManAddClause(p, []lit{0, 2})
	Sto_ManMarkRoots(p)
	Sto_ManAddClause(p, []lit{4})
	Sto_ManAddClause(p, nil)
	if _, err := reconstructResolution(p, nil, nil, 0); err == nil {
		t.Fatal("accepted non-RUP clause")
	}
	if _, err := reconstructResolution(p, nil, nil, time.Now().Add(-time.Second).UnixNano()); err == nil {
		t.Fatal("deadline ignored")
	}
}

func TestResolutionReconstructionEmptyRoot(t *testing.T) {
	p := Sto_ManAlloc()
	if Sto_ManAddClause(p, nil) != 1 {
		t.Fatal("add empty root")
	}
	Sto_ManMarkRoots(p)
	var proof bytes.Buffer
	if _, err := reconstructResolution(p, nil, &proof, 0); err != nil {
		t.Fatal(err)
	}
	summary, err := CheckResolutionProof(&proof)
	if err != nil || !summary.EmptyDerived || summary.Roots != 1 {
		t.Fatalf("summary=%+v err=%v", summary, err)
	}
}
