package fraig

import "testing"

func TestSparseReductionProvesContradictoryCone(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 1, FDoSparse: 1, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	left := Fraig_NodeAnd(p, a, b)
	right := Fraig_NodeAnd(p, Fraig_Not(a), c)
	zero := Fraig_NodeAnd(p, left, right)
	if zero != Fraig_Not(p.pConst1) {
		t.Fatalf("sparse contradiction was not reduced to zero: %v", zero)
	}
	if p.nSatCalls == 0 || p.nSatZeros == 0 {
		t.Fatalf("sparse SAT path not used: calls=%d zeros=%d", p.nSatCalls, p.nSatZeros)
	}
}
func TestChoicingRecordsIndependentAssociativeImplementation(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 1, FChoicing: 1, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b, c := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1), Fraig_ManReadIthVar(p, 2)
	left := Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), c)
	right := Fraig_NodeAnd(p, a, Fraig_NodeAnd(p, b, c))
	if right != left {
		t.Fatal("functional representative differs")
	}
	choice := Fraig_Regular(left).pNextE
	if choice == nil || choice.pRepr != Fraig_Regular(left) {
		t.Fatal("independent equivalent implementation not recorded")
	}
	for bits := 0; bits < 8; bits++ {
		model := []int32{int32(bits & 1), int32(bits >> 1 & 1), int32(bits >> 2 & 1)}
		if Fraig_ManSimulateBitNode(p, choice, model) != Fraig_ManSimulateBitNode(p, left, model) {
			t.Fatalf("choice differs for assignment %d", bits)
		}
	}
}

func TestChoicingDoesNotCreateTransitiveFaninCycle(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128, FFuncRed: 1, FChoicing: 1, NBTLimit: -1})
	defer Fraig_ManFree(p)
	a, b := Fraig_ManReadIthVar(p, 0), Fraig_ManReadIthVar(p, 1)
	expanded := Fraig_NodeAnd(p, a, Fraig_NodeOr(p, a, b))
	if expanded != a {
		t.Fatal("absorption should reduce to the input")
	}
	if a.pNextE != nil {
		t.Fatal("choice linking to its own transitive fanin creates a cycle")
	}
}
