package fraig

import "testing"

func TestConstrainedMiterWithInputCreatedAfterClause(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a := Fraig_ManReadIthVar(p, 0)
	Fraig_ManAddClause(p, []*Fraig_Node_t{a}, 1)
	b := Fraig_ManReadIthVar(p, 1)
	Fraig_ManSetPo(p, b)
	if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 2 || p.pModel[0] != 1 || p.pModel[1] != 1 {
		t.Fatalf("status %d model %v", got, p.pModel)
	}
}
