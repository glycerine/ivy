package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceMonotoneEntryDiscardsProvedResults(t *testing.T) {
	p := aig.Aig_ManStart(10)
	state := aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, 0)
	aig.Aig_ObjCreateCo(p, 0)
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ManSetRegNum(p, 1)
	n, err := NewNetwork(p, []string{"state"}, []string{"pendingSignal", "hint_up", "next"})
	if err != nil {
		t.Fatal(err)
	}
	proved := FindMonotone(n, SafetyOptions{}, CheckSafety)
	if len(proved) == 0 {
		t.Fatal("fixture must have a proved hint")
	}
	if got := findMonotoneSignals(n); got != nil {
		t.Fatalf("C entry returns NULL after proof, got %v", got)
	}
}
