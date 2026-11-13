package fraig

import "testing"

func TestConstantConeTraversal(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{nPatsRand: 128, nPatsDyna: 128})
	defer Fraig_ManFree(p)
	p.nTravIds += 2
	if got := Fraig_MarkTfi_rec(p, p.pConst1); got != 0 {
		t.Fatal(got)
	}
	p.nTravIds += 2
	if got := Fraig_MarkTfi2_rec(p, p.pConst1); got != 0 {
		t.Fatal(got)
	}
	p.nTravIds += 2
	if got := Fraig_MarkTfi3_rec(p, p.pConst1); got != 1 {
		t.Fatal(got)
	}
	a := Fraig_ManReadIthVar(p, 0)
	if Fraig_NodeIsInSupergate(p.pConst1, a) != 0 {
		t.Fatal("constant has no fanins")
	}
	if v := Fraig_CollectSupergate(p.pConst1, 0); len(v.pArray) != 1 || v.pArray[0] != p.pConst1 {
		t.Fatal("constant supergate", v.pArray)
	}
}
