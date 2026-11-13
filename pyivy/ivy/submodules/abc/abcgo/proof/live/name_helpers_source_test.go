package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceNameHelpers(t *testing.T) {
	old := aig.Aig_ManStart(20)
	pi, lo := aig.Aig_ObjCreateCi(old), aig.Aig_ObjCreateCi(old)
	live, fair := aig.Aig_ObjCreateCo(old, pi), aig.Aig_ObjCreateCo(old, lo)
	aig.Aig_ObjCreateCo(old, lo)
	aig.Aig_ManSetRegNum(old, 1)
	n := &Network{Graph: old, InputNames: []string{"data", "state"}, OutputNames: []string{"assert_fair_x", "assume_fair_y", "next"}}
	dst := aig.Aig_ManStart(20)
	inputs := make([]aig.Edge, 7)
	for i := range inputs {
		inputs[i] = aig.Aig_ObjCreateCi(dst)
	}
	for i := 0; i < 5; i++ {
		aig.Aig_ObjCreateCo(dst, 0)
	}
	aig.Aig_ManSetRegNum(dst, 5)
	aig.Aig_ObjSetCopy(old, pi, inputs[0])
	aig.Aig_ObjSetCopy(old, lo, inputs[2])
	for i, want := range []string{"data", "SAVE_BIERE"} {
		if got := retrieveTruePiName(n, old, dst, inputs[i]); got != want {
			t.Fatalf("PI %d got %q want %q", i, got, want)
		}
	}
	for i, want := range []string{"state", "SAVED_LO", "state__SHADOW", "assert_fair_x__LIVENESS", "assume_fair_y__FAIRNESS"} {
		if got := retrieveLOName(n, old, dst, inputs[i+2], []aig.Edge{live}, []aig.Edge{fair}); got != want {
			t.Fatalf("LO %d got %q want %q", i, got, want)
		}
	}
	if getPoIndex(old, fair) != 1 || getPoIndex(old, pi) != -1 {
		t.Fatal("source PO lookup mismatch")
	}
	updateNewNetworkNameManager(n, dst, []string{"a", "save"}, []string{"r"})
	if len(n.InputNames) != 3 || n.InputNames[2] != "r" {
		t.Fatalf("updated names %v", n.InputNames)
	}
}
