package fraig

import (
	"github.com/glycerine/abcgo/sat/msat"
	"testing"
)

// The four clauses in fraigSat.c encode the regular EXOR/NEXOR root.
func TestSourceXorClausesAllAssignments(t *testing.T) {
	for phase := 0; phase < 4; phase++ {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
		a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
		x, y := a, b
		if phase&1 != 0 {
			x = Fraig_Not(x)
		}
		if phase&2 != 0 {
			y = Fraig_Not(y)
		}
		node := Fraig_Regular(Fraig_NodeExor(p, x, y))
		fraigSatEnsureVariables(p)
		Fraig_SupergateAddClausesExor(p, node)
		for _, n := range []*Fraig_Node_t{a, b, node} {
			msat.Msat_IntVecPush(p.vVarsInt, int(n.Num))
		}
		for bits := 0; bits < 8; bits++ {
			av, bv, out := bits&1 != 0, bits&2 != 0, bits&4 != 0
			got := fraigSatSolve(p, -1, 5, fraigLit(a, !av), fraigLit(b, !bv), fraigLit(node, !out))
			want := (av != bv) == (Fraig_NodeIsExor(node) != 0)
			if (got == int(msat.MSAT_TRUE)) != (out == want) {
				t.Fatalf("phase=%d bits=%d result=%d expected root=%v", phase, bits, got, want)
			}
		}
		Fraig_ManFree(p)
	}
}
