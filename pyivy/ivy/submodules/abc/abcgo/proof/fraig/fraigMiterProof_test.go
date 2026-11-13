package fraig

import "testing"

func TestProveMiterUsesSATForSparseContradiction(t *testing.T) {
	for _, invert := range []bool{false, true} {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FTryProve: 1, NSeconds: 5, NBTLimit: -1})
		a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
		node := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), Fraig_NodeAnd(p, Fraig_Not(a), c))
		if Fraig_NodeIsConst(node) != 0 {
			t.Fatal("structural-only mode unexpectedly reduced node")
		}
		output := Fraig_NotCond(node, invert)
		Fraig_ManSetPo(p, output)
		Fraig_ManProveMiter(p)
		want := Fraig_NotCond(p.pConst1, !invert)
		if got := p.vOutputs.pArray[0]; got != want {
			t.Fatalf("invert %t: got %v want %v", invert, got, want)
		}
		if p.nSatCalls == 0 {
			t.Fatal("SAT proof was not attempted")
		}
		Fraig_ManFree(p)
	}
}
