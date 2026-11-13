package fraig

import "testing"

func TestConstantMiterResults(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	Fraig_NodeVecPush(p.vOutputs, Fraig_Not(p.pConst1))
	if got := Fraig_ManCheckMiter(p); got != 1 {
		t.Fatal("false output", got)
	}
	p.vOutputs.pArray[0] = p.pConst1
	if got := Fraig_ManCheckMiter(p); got != 0 || p.pModel == nil {
		t.Fatal("true output", got, p.pModel)
	}
}
