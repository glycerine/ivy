package fraig_test

import (
	"github.com/glycerine/abcgo/proof/fraig"
	"testing"
)

func TestPublicGraphConstructionAndMiter(t *testing.T) {
	var params fraig.Fraig_Params_t
	fraig.Fraig_ParamsSetDefault(&params)
	params.FFuncRed = 1
	params.FDoSparse = 1
	params.NPatsRand = 128
	params.NPatsDyna = 128
	params.NBTLimit = -1
	p := fraig.Fraig_ManCreate(&params)
	defer fraig.Fraig_ManFree(p)
	a, b, c := fraig.Fraig_ManReadIthVar(p, 0), fraig.Fraig_ManReadIthVar(p, 1), fraig.Fraig_ManReadIthVar(p, 2)
	left := fraig.Fraig_NodeAnd(p, fraig.Fraig_NodeAnd(p, a, b), c)
	right := fraig.Fraig_NodeAnd(p, a, fraig.Fraig_NodeAnd(p, b, c))
	if left != right {
		t.Fatal("functional reduction failed to merge equal graphs")
	}
	difference := fraig.Fraig_NodeExor(p, left, right)
	fraig.Fraig_ManSetPo(p, difference)
	if result := fraig.Fraig_ManCheckMiter(p); result != 1 {
		t.Fatalf("equal graphs have nonzero miter: %d", result)
	}
}
