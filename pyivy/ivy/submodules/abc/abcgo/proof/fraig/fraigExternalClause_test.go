package fraig

import "testing"

func TestAddClauseDoesNotEncodeInternalNodes(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	both := Fraig_NodeAnd(p, a, b)
	fraigSatEnsureVariables(p)
	Fraig_ManAddClause(p, []*Fraig_Node_t{both}, 1)
	if both.fClauses != 0 || both.vFanins != nil {
		t.Fatal("adding a clause unexpectedly encoded its internal gate")
	}
}

// ABC fraigSat.c:130-157 checks simulation/constant outputs without consulting
// the SAT solver used by fraigMan.c:521-538 to store external clauses.
// These tests characterize the resulting limitation, not a valid constrained
// counterexample contract.
func TestMiterCheckIgnoresExternalClauses(t *testing.T) {
	t.Run("constant output", func(t *testing.T) {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
		defer Fraig_ManFree(p)
		a := Fraig_ManReadIthVar(p, 0)
		fraigSatEnsureVariables(p)
		Fraig_ManAddClause(p, []*Fraig_Node_t{a}, 1)
		Fraig_ManSetPo(p, p.pConst1)
		if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 1 || p.pModel[0] != 0 {
			t.Fatalf("source behavior changed: status=%d model=%v", got, p.pModel)
		}
		if Fraig_ManSimulateBitNode(p, a, p.pModel) != 0 {
			t.Fatal("fixture must demonstrate a model violating the constraint")
		}
	})
	t.Run("simulation output", func(t *testing.T) {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
		defer Fraig_ManFree(p)
		a := Fraig_ManReadIthVar(p, 0)
		fraigSatEnsureVariables(p)
		Fraig_ManAddClause(p, []*Fraig_Node_t{Fraig_Not(a)}, 1)
		Fraig_ManSetPo(p, a)
		if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 1 || p.pModel[0] != 1 {
			t.Fatalf("source behavior changed: status=%d model=%v", got, p.pModel)
		}
		if Fraig_ManSimulateBitNode(p, Fraig_Not(a), p.pModel) != 0 {
			t.Fatal("fixture must demonstrate a model violating the constraint")
		}
	})
}
