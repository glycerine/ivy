package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestSecTransferRequiresOriginalReplay(t *testing.T) {
	original := aig.NewGraph()
	x := original.NewInput()
	original.NewInput()
	original.AddOutput(x)
	original.AddOutput(original.Const0())
	aig.Aig_ManSetRegNum(original, 1)
	reduced := aig.Aig_ManDupSimple(original)
	cex := util.Abc_CexAlloc(1, 1, 1)
	if err := reduced.SetSequentialCounterexample(cex); err != nil {
		t.Fatal(err)
	}
	if secTransferWitness(original, reduced) {
		t.Fatal("accepted nonfailing witness")
	}
	util.Abc_InfoSetBit(cex.PData, 1)
	if err := reduced.SetSequentialCounterexample(cex); err != nil {
		t.Fatal(err)
	}
	if !secTransferWitness(original, reduced) || original.SequentialCounterexample() == nil {
		t.Fatal("failed to transfer valid witness")
	}
	for _, edge := range original.LiveEdges() {
		if aig.Aig_ObjIsMarkB(original, edge) {
			t.Fatalf("witness replay left mark on %v", edge)
		}
	}

}
