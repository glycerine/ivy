package hop

import (
	"github.com/glycerine/abcgo/bool/dec"
	"testing"
)

func TestDecompositionNetworkAdapter(t *testing.T) {
	p := Hop_ManStart()
	defer Hop_ManStop(p)
	var _ dec.VariableNetwork = p
	root := dec.Dec_GraphFactorSop(p, "10 1\n01 1\n").(*Hop_Obj_t)
	if got := Hop_ManComputeTruth6(p, root, 2); got != 0x6666666666666666 {
		t.Fatalf("factored SOP truth %x", got)
	}
	if Hop_ManCheck(p) != 1 {
		t.Fatal("factoring produced inconsistent graph")
	}
}
