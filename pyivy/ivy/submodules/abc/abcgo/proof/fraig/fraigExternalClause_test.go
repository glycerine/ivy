package fraig

import "testing"

func TestExternalClauseConstrainsFreshSolver(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	Fraig_ManAddClause(p, []*Fraig_Node_t{Fraig_Not(a), b}, 2)
	if got := Fraig_ManCheckClauseUsingSat(p, Fraig_Not(a), b, -1); got != 1 {
		t.Fatalf("external clause was not enforced: %d", got)
	}
}

func TestExternalClauseEncodesInternalNodes(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	both := Fraig_NodeAnd(p, a, b)
	Fraig_ManAddClause(p, []*Fraig_Node_t{both}, 1)
	if got := Fraig_ManCheckClauseUsingSat(p, a, b, -1); got != 1 {
		t.Fatalf("clause forcing AND did not constrain fanins: %d", got)
	}
}
func TestExternalClauseSimplifiesConstants(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	Fraig_ManAddClause(p, []*Fraig_Node_t{Fraig_Not(p.pConst1), a}, 2)
	if got := Fraig_ManCheckClauseUsingSat(p, a, b, -1); got != 1 {
		t.Fatalf("constant-false clause literal was not removed: %d", got)
	}
}

func TestExternalNonUnitClauseIncludesItsGateCone(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	both := Fraig_NodeAnd(p, a, b)
	Fraig_ManAddClause(p, []*Fraig_Node_t{both, c}, 2)
	if got := Fraig_ManCheckClauseUsingSat(p, a, c, -1); got != 1 {
		t.Fatalf("external (a AND b) OR c must imply a OR c: %d", got)
	}
}
