package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/cut"
)

func TestAbcNtkStartCutManForRewrite(t *testing.T) {
	network := abc.NewNetwork("rewrite-cuts")
	manager := Abc_NtkStartCutManForRewrite(network)
	if manager == nil {
		t.Fatal("cut manager is nil")
	}
	defer cut.Cut_ManStop(manager)
	params := cut.Cut_ManReadParams(manager)
	if params.NVarsMax != 4 || params.NKeepMax != 250 || params.NIdsMax != abc.Abc_NtkObjNumMax(network) || params.FTruth == 0 || params.FFilter == 0 {
		t.Fatalf("unexpected rewrite cut parameters: %+v", params)
	}
}

func TestAbcRwrExpWithCutRecRecognizesComplementedLeaf(t *testing.T) {
	network := abc.NewNetwork("rewrite-complemented-leaf")
	leaf := network.NewObject(abc.ObjectPI)
	leaves := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(leaves, leaf)
	defer vec.Vec_PtrFree(leaves)

	Abc_RwrExpWithCut_rec(leaf.Signal().Not(), leaves, 1)
	if !leaf.MarkA {
		t.Fatal("complemented leaf was not marked as part of the cut")
	}
}
