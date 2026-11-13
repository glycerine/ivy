package fraig

import (
	"github.com/glycerine/abcgo/sat/msat"
	"testing"
)

func TestMuxEncodingTruthTable(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{nPatsRand: 128, nPatsDyna: 128, fFuncRed: 0})
	defer Fraig_ManFree(p)
	c, x, y := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	out := Fraig_NodeMux(p, c, x, y)
	root := Fraig_Regular(out)
	if Fraig_NodeIsMuxType(root) == 0 {
		t.Fatal("mux structure not recognized")
	}
	fraigSatEnsureVariables(p)
	Fraig_OrderVariables(p, c, root)
	for cv := 0; cv < 2; cv++ {
		for xv := 0; xv < 2; xv++ {
			for yv := 0; yv < 2; yv++ {
				for fv := 0; fv < 2; fv++ {
					want := yv
					if cv == 1 {
						want = xv
					}
					expected := fv == want
					// An assumption literal is negated exactly when the variable value is zero.
					nodeValue := fv
					if Fraig_IsComplement(out) {
						nodeValue ^= 1
					}
					result := fraigSatSolve(p, -1, 5, fraigLit(c, cv == 0), fraigLit(x, xv == 0), fraigLit(y, yv == 0), fraigLit(root, nodeValue == 0))
					if (result == int(msat.MSAT_TRUE)) != expected {
						t.Fatalf("mux %d %d %d => %d: status %d", cv, xv, yv, fv, result)
					}
				}
			}
		}
	}
}
func TestSatCounterexampleFeedback(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{nPatsRand: 128, nPatsDyna: 128, fFuncRed: 0})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	both := Fraig_NodeAnd(p, a, b)
	if Fraig_NodeIsEquivalent(p, a, both, -1, 5) != 0 {
		t.Fatal("different functions were reported equal")
	}
	if p.nSatCounter != 1 || p.iWordStart == 0 {
		t.Fatalf("counterexample was not fed back: %d %d", p.nSatCounter, p.iWordStart)
	}
}

func TestFunctionalReductionPreservesTruthTable(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{nPatsRand: 128, nPatsDyna: 128, fFuncRed: 1, fFeedBack: 1, fDoSparse: 1, nBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	left := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), c)
	right := Fraig_NodeAnd(p, a, Fraig_NodeAnd(p, b, c))
	if left != right {
		t.Fatal("functional reduction failed to identify associative AND")
	}
	for av := int32(0); av < 2; av++ {
		for bv := int32(0); bv < 2; bv++ {
			for cv := int32(0); cv < 2; cv++ {
				if got := Fraig_ManSimulateBitNode(p, right, []int32{av, bv, cv}); got != av&bv&cv {
					t.Fatalf("AND %d %d %d = %d", av, bv, cv, got)
				}
			}
		}
	}
}
