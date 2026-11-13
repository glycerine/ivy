package fraig

import "testing"

func TestConstrainedMiterReplayWithSeparateComplementedCone(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	gate := Fraig_NodeAnd(p, a, b)
	Fraig_ManAddClause(p, []*Fraig_Node_t{Fraig_Not(gate), c}, 2)
	Fraig_ManSetPo(p, Fraig_Not(c))
	Fraig_ManSetPo(p, Fraig_NodeExor(p, a, b))
	for query := 0; query < 3; query++ {
		if got := Fraig_ManCheckMiter(p); got != 0 || len(p.pModel) != 3 {
			t.Fatalf("query %d: status=%d model=%v", query, got, p.pModel)
		}
		x, y, z := p.pModel[0] != 0, p.pModel[1] != 0, p.pModel[2] != 0
		if x && y && !z {
			t.Fatalf("query %d violates external clause: %v", query, p.pModel)
		}
		if !(!z || x != y) {
			t.Fatalf("query %d does not satisfy any output: %v", query, p.pModel)
		}
	}
}
