package live

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// C liveness.c uses pNtk for names and pAig for graph objects; the two
// arguments are not required to point at the same graph representation.
func TestSourceWrappersUseExplicitAIG(t *testing.T) {
	named := aig.NewGraph()
	named.NewInput()
	named.NewInput()
	named.AddOutput(named.Const1())
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	output := g.AddOutput(gate)
	n := &Network{Graph: named, OutputNames: []string{"assert_safety_x"}}
	selected := populateSafetyAssertionVector(n, g)
	if len(selected) != 1 || selected[0] != output {
		t.Fatalf("selected %v, want explicit AIG output %v", selected, output)
	}
	got := LivenessToSafetyTransformation(IGNORE_LIVENESS_KEEP_SAFETY_MODE, n, g, nil, nil, selected, nil)
	want, err := LivenessToSafety(&Network{Graph: g, OutputNames: n.OutputNames}, TransformOptions{Mode: IGNORE_LIVENESS_KEEP_SAFETY_MODE, SafetyAssertions: selected})
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		x, err := got.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		y, err := want.Graph.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if len(x) != len(y) || x[0] != y[0] {
			t.Fatalf("input=%v got=%v want=%v", in, x, y)
		}
	}
	if n.Graph != named {
		t.Fatal("wrapper changed caller's network")
	}
}

func TestSourceSignalSelectorsUseExplicitAIG(t *testing.T) {
	named := aig.NewGraph()
	a := named.NewInput()
	for i := 0; i < 4; i++ {
		named.AddOutput(a)
	}
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	for i := 0; i < 4; i++ {
		g.AddOutput(gate)
	}
	n := &Network{Graph: named, OutputNames: []string{"assert_fair_x", "assume_fair_x", "Assert_x", "Assume_x"}}
	for i, selectSignal := range []func(*Network, *aig.Graph) []aig.Edge{populateLivenessVector, populateFairnessVector, populateSafetyAssertionVector, populateSafetyAssumptionVector} {
		got := selectSignal(n, g)
		if len(got) != 1 || got[0] != g.Outputs()[i] {
			t.Fatalf("selector %d got %v", i, got)
		}
	}
}

// arenaViolation.c searches only as many PO names as the supplied AIG has.
func TestArenaSelectorsUseExplicitAIGOutputCount(t *testing.T) {
	named := aig.NewGraph()
	named.AddOutput(named.NewInput())
	named.AddOutput(named.Const1())
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	n := &Network{Graph: named, OutputNames: []string{"other", "windowBegins_withinWindow_pendingSignal"}}
	if collectWindowBeginSignalWithDSC(n, g) != -1 || collectWithinWindowSignalWithDSC(n, g) != -1 || collectPendingSignalWithDSC(n, g) != -1 {
		t.Fatal("arena selectors examined an output absent from the supplied AIG")
	}
}
