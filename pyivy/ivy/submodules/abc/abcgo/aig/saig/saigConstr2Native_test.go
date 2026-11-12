package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

func makeStableLatchGraph() (*aig.Graph, aig.Edge) {
	g := aig.NewGraph()
	g.NewInput() // primary input
	state := g.NewInput()
	g.AddOutput(state) // property
	g.AddOutput(state) // latch input
	aig.Aig_ManSetRegNum(g, 1)
	return g, state
}

func TestFunctionalConstraintInductionMiter(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 0)
	miter, err := InductionMiter(g, FunctionalCandidates{{x}})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(miter.Outputs()); got != 1 {
		t.Fatalf("outputs = %d, want 1", got)
	}
	cnf, err := encodeAIG(miter)
	if err != nil {
		t.Fatal(err)
	}
	assumption := cnf.lit(aig.Aig_ObjChild0(miter, miter.Outputs()[0]))
	if got := cnf.solver.Solve(assumption); got != bsat.Satisfiable {
		t.Fatalf("transition x(0) && !x(1) status = %v, want satisfiable", got)
	}
	filtered, err := FilterInductiveCandidates(g, FunctionalCandidates{{x}}, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || len(filtered[0]) != 0 {
		t.Fatalf("non-inductive candidate survived filtering: %v", filtered)
	}
}

func TestDetectFunctionalStableStateConstraint(t *testing.T) {
	g, state := makeStableLatchGraph()
	candidates, err := DetectFunctionalConstraints(g, 2, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("candidate levels = %d, want 2", len(candidates))
	}
	for level := range candidates {
		found := false
		for _, candidate := range candidates[level] {
			if candidate == state {
				found = true
			}
		}
		if !found {
			t.Errorf("level %d is missing stable state candidate %v: %v", level, state, candidates[level])
		}
	}
}

func TestUnrollPropertyAdvancesLatchState(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	state := g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	unrolled, err := UnrollProperty(g, 2)
	if err != nil {
		t.Fatal(err)
	}
	// Inputs are created frame-wise as [x0, q0, x1]; the final state is x0.
	for _, tc := range []struct {
		inputs []bool
		want   bool
	}{{[]bool{false, true, true}, false}, {[]bool{true, false, false}, true}} {
		got, err := unrolled.Evaluate(tc.inputs)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("Evaluate(%v) = %v, want [%v]", tc.inputs, got, tc.want)
		}
	}
}

func TestUnfoldFunctionalConstraintAddsPipelineFlop(t *testing.T) {
	g, state := makeStableLatchGraph()
	unfolded, err := UnfoldFunctionalConstraints(g, FunctionalCandidates{nil, {state}})
	if err != nil {
		t.Fatal(err)
	}
	if got := aig.Aig_ManRegNum(unfolded); got != 2 {
		t.Fatalf("registers = %d, want 2", got)
	}
	if got := aig.Aig_ManConstrNum(unfolded); got != 1 {
		t.Fatalf("constraints = %d, want 1", got)
	}
	if got := len(unfolded.Outputs()); got != 4 {
		t.Fatalf("outputs = %d, want original property, new constraint, and two LIs", got)
	}
}

func TestFoldFunctionalConstraintGatesProperties(t *testing.T) {
	g := aig.NewGraph()
	property := g.NewInput()
	constraint := g.NewInput()
	g.AddOutput(property)
	g.AddOutput(constraint)
	aig.Aig_ManSetConstrNum(g, 1)
	folded, err := FoldFunctionalConstraints(g, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(folded.Outputs()); got != 1 {
		t.Fatalf("outputs = %d, want 1", got)
	}
	for _, tc := range []struct {
		inputs []bool
		want   bool
	}{{[]bool{false, false}, false}, {[]bool{true, false}, true}, {[]bool{true, true}, false}, {[]bool{false, true}, false}} {
		got, err := folded.Evaluate(tc.inputs)
		if err != nil {
			t.Fatal(err)
		}
		if got[0] != tc.want {
			t.Errorf("Evaluate(%v) = %v, want %v", tc.inputs, got[0], tc.want)
		}
	}
}
