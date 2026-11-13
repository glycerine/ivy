package fraig_test

import (
	"github.com/glycerine/abcgo/proof/fraig"
	"testing"
)

func TestParametersFromAnotherPackage(t *testing.T) {
	var params fraig.Fraig_Params_t
	fraig.Fraig_ParamsSetDefault(&params)
	if params.NPatsRand != 2048 || params.FFuncRed != 1 {
		t.Fatalf("incorrect defaults: %+v", params)
	}
	params.NPatsRand = 256
	params.NPatsDyna = 512
	params.FFuncRed = 1
	params.FDoSparse = 1
	params.NBTLimit = -1
	p := fraig.Fraig_ManCreate(&params)
	defer fraig.Fraig_ManFree(p)
	if fraig.Fraig_ManReadPatternNumRandom(p) != 256 || fraig.Fraig_ManReadDoSparse(p) != 1 {
		t.Fatal("configured parameters were not applied")
	}
	fraig.Fraig_ParamsSetDefaultFull(&params)
	if params.NBTLimit != -1 || params.FDoSparse != 1 || params.FTryProve != 0 {
		t.Fatalf("incorrect full defaults: %+v", params)
	}
}
