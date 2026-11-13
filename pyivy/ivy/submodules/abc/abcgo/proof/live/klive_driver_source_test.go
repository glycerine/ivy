package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceKLiveDriverUsesMutatingPDRPath(t *testing.T) {
	p := aig.Aig_ManStart(10)
	state := aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, 0) // target initially true
	aig.Aig_ObjCreateCo(p, 1) // invariant false
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ManSetRegNum(p, 1)
	n := &Network{Graph: p, OutputNames: []string{"csLiveTarget_x", "csSafetyInvar_x", "next"}}
	result := kLive(n, KLiveOptions{Directive: kCS_WITH_SAFETY_INVARIANTS, MaxAbsorbers: 1}, true)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.Graph == nil {
		t.Fatal("source driver did not construct graph")
	}
	if got := outputDriver(result.Graph, 3); got != aig.Edge(1) {
		t.Fatalf("source PDR must patch absorber target to false; got %v", got)
	}
}
