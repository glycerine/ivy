package fraig

import "testing"

func TestConstrainedMiterRejectsInvalidSimulationWitness(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a := Fraig_ManReadIthVar(p, 0)
	Fraig_ManAddClause(p, []*Fraig_Node_t{Fraig_Not(a)}, 1)
	Fraig_ManSetPo(p, a)
	if got := Fraig_ManCheckMiter(p); got != 1 {
		t.Fatalf("a with constraint !a must be UNSAT, got %d model %v", got, p.pModel)
	}
}
func TestConstrainedMiterProducesValidWitness(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	Fraig_ManAddClause(p, []*Fraig_Node_t{Fraig_Not(a)}, 1)
	Fraig_ManSetPo(p, b)
	if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 2 || p.pModel[0] != 0 || p.pModel[1] != 1 {
		t.Fatalf("invalid constrained model: status %d model %v", got, p.pModel)
	}
}
func TestConstrainedConstantMiterFindsRequiredInput(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a := Fraig_ManReadIthVar(p, 0)
	Fraig_ManAddClause(p, []*Fraig_Node_t{a}, 1)
	Fraig_ManSetPo(p, p.pConst1)
	if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 1 || p.pModel[0] != 1 {
		t.Fatalf("constant miter witness violates constraint: status %d model %v", got, p.pModel)
	}
}
