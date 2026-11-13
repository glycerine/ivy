package fraig

import "testing"

func TestConstrainedMiterContinuesPastUnsatisfiableOutput(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	Fraig_ManAddClause(p, []*Fraig_Node_t{Fraig_Not(a)}, 1)
	Fraig_ManSetPo(p, a)
	Fraig_ManSetPo(p, b)
	for i := 0; i < 3; i++ {
		if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 2 || p.pModel[0] != 0 || p.pModel[1] != 1 {
			t.Fatalf("query %d status %d model %v", i, got, p.pModel)
		}
	}
}
