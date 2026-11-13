package msat

import "testing"

func TestRemoveLearnedClearsAllReasonsWithoutCancelingDecisions(t *testing.T) {
	p := New(2)
	if !p.AddClause([]int{0, 2}) {
		t.Fatal("clause")
	}
	if Msat_SolverAssume(p, 1) == 0 || Msat_SolverPropagate(p) != nil {
		t.Fatal("assumption")
	}
	if p.pReasons[1] == nil {
		t.Fatal("expected implication reason")
	}
	Msat_SolverRemoveLearned(p)
	if p.pReasons[1] != nil {
		t.Fatal("C clears all reasons")
	}
	if len(p.vTrailLim.data) != 1 || p.pAssigns[0] != 1 {
		t.Fatal("C does not cancel decisions")
	}
}

func TestClauseIsLockedUsesFirstLiteralReason(t *testing.T) {
	p := New(2)
	var c *Msat_Clause_t
	if Msat_ClauseCreate(p, &Msat_IntVec_t{data: []int{0, 2}}, 0, &c) == 0 {
		t.Fatal("clause")
	}
	p.pReasons[1] = c
	p.pAssigns[1] = 2
	if Msat_ClauseIsLocked(p, c) != 0 {
		t.Fatal("C checks only the first literal's reason")
	}
	p.pReasons[0] = c
	p.pAssigns[0] = -1
	if Msat_ClauseIsLocked(p, c) == 0 {
		t.Fatal("C does not additionally check assignment")
	}
}
