package fraig

import (
	"testing"
)

func TestRealPigeonholeConflictLimit(t *testing.T) {
	for _, lim := range []int32{1, -1} {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
		var xs [5][4]*Fraig_Node_t
		for i := 0; i < 5; i++ {
			for j := 0; j < 4; j++ {
				xs[i][j] = Fraig_ManReadIthVar(p, int32(i*4+j))
			}
		}
		root := p.pConst1
		for i := 0; i < 5; i++ {
			row := Fraig_Not(p.pConst1)
			for j := 0; j < 4; j++ {
				row = Fraig_NodeOr(p, row, xs[i][j])
			}
			root = Fraig_NodeAnd(p, root, row)
		}
		for j := 0; j < 4; j++ {
			for i := 0; i < 5; i++ {
				for k := 0; k < i; k++ {
					clause := Fraig_NodeOr(p, Fraig_Not(xs[i][j]), Fraig_Not(xs[k][j]))
					root = Fraig_NodeAnd(p, root, clause)
				}
			}
		}
		r := Fraig_NodeIsEquivalent(p, p.pConst1, Fraig_Regular(root), lim, 0)
		wantResult, wantFail := int32(0), int32(1)
		if lim < 0 {
			wantResult, wantFail = 1, 0
		}
		if r != wantResult || p.nSatFailsReal != wantFail || Fraig_Regular(root).fFailTfo != uint32(wantFail) || p.nSatCalls != 1 {
			t.Fatalf("limit %d: result=%d failures=%d failTfo=%d calls=%d", lim, r, p.nSatFailsReal, Fraig_Regular(root).fFailTfo, p.nSatCalls)
		}
		Fraig_ManFree(p)
	}
}
