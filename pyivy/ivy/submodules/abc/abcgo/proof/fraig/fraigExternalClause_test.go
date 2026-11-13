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
