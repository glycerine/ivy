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

func TestSecTransfersCombinationalMixedModel(t *testing.T) {
	original := aig.NewGraph()
	a, b := original.NewInput(), original.NewInput()
	original.AddOutput(original.XOR(a, b))
	reduced := aig.Aig_ManDupSimple(original)
	reduced.Data = []int32{1, 0}
	if !secTransferWitness(original, reduced) {
		t.Fatal("mixed combinational witness not transferred")
	}
	cex := original.SequentialCounterexample()
	if cex == nil || !util.Abc_InfoHasBit(cex.PData, 0) || util.Abc_InfoHasBit(cex.PData, 1) {
		t.Fatalf("wrong witness: %+v", cex)
	}
	reduced.Data = []int32{1, 1}
	original.ClearSequentialCounterexample()
	if secTransferWitness(original, reduced) {
		t.Fatal("accepted nonfailing combinational model")
	}
}
