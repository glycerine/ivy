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

// A consumer can inspect the returned input assignment and replay it against
// its own miter. This uses the public package boundary rather than internals.
func TestPublicMiterCounterexampleReplay(t *testing.T) {
	var params fraig.Fraig_Params_t
	fraig.Fraig_ParamsSetDefault(&params)
	params.NPatsRand = 128
	params.NPatsDyna = 128
	params.NBTLimit = -1
	p := fraig.Fraig_ManCreate(&params)
	defer fraig.Fraig_ManFree(p)
	a, b, c := fraig.Fraig_ManReadIthVar(p, 0), fraig.Fraig_ManReadIthVar(p, 1), fraig.Fraig_ManReadIthVar(p, 2)
	output := fraig.Fraig_NodeAnd(p, fraig.Fraig_NodeExor(p, a, b), c)
	fraig.Fraig_ManSetPo(p, output)
	if got := fraig.Fraig_ManCheckMiter(p); got != 0 {
		t.Fatalf("miter status %d", got)
	}
	model := fraig.Fraig_ManReadModel(p)
	if len(model) != 3 {
		t.Fatalf("model %v", model)
	}
	if (model[0] != model[1]) == false || model[2] == 0 {
		t.Fatalf("model does not satisfy miter: %v", model)
	}
}
