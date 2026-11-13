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

func TestSourceKLiveBarrierDirectivesRequireNonemptyBarriers(t *testing.T) {
	p := aig.Aig_ManStart(10)
	state := aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, 0)
	aig.Aig_ObjCreateCo(p, 0)
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ManSetRegNum(p, 1)
	n := &Network{Graph: p, OutputNames: []string{"csLiveTarget_x", "csSafetyInvar_x", "next"}}
	defer func() {
		r := recover()
		if r != "live: source requires nonempty barriers" {
			t.Fatalf("expected source nonempty-barrier assertion, got %v", r)
		}
	}()
	kLive(n, KLiveOptions{Directive: kCS_WITH_SAFETY_AND_USER_GIVEN_DCS_INVARIANTS, MaxAbsorbers: 1}, true)
}

func TestSourceKLiveConstraintRequiresInvariant(t *testing.T) {
	n := liveFixture(false)
	n.OutputNames[0] = "csLiveTarget_x"
	defer func() {
		r := recover()
		if r != "live: source requires safety invariant" {
			t.Fatalf("expected source invariant assertion, got %v", r)
		}
	}()
	kLive(n, KLiveOptions{Directive: kCS_WITH_SAFETY_INVARIANTS, MaxAbsorbers: 1}, true)
}
