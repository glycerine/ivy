package live

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestLTLRespectsFairnessAssumptions(t *testing.T) {
	n := liveFixture(false)
	for _, tc := range []struct {
		name     string
		fairness []aig.Edge
		want     int32
	}{
		{"no fairness", nil, 0},
		{"fair loop", []aig.Edge{0}, 0},
		{"unfair loop", []aig.Edge{1}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := LivenessToSafety(n, TransformOptions{Mode: IGNORE_SAFETY_KEEP_LIVENESS_MODE, LTL: []string{":G :F assert_fair_q"}, Fairness: tc.fairness})
			if err != nil {
				t.Fatal(err)
			}
			if got := CheckSafety(result.Graph, 0, SafetyOptions{}); got.Status != tc.want {
				t.Fatalf("status %d, want %d: %+v", got.Status, tc.want, got)
			}
		})
	}
}

func TestLTLWrapperPreservesFairness(t *testing.T) {
	n := liveFixture(false)
	n.LTLProperties = []string{":G :F assert_fair_q"}
	g := LivenessToSafetyTransformationWithLTL(IGNORE_SAFETY_KEEP_LIVENESS_MODE, n, n.Graph, nil, []aig.Edge{1}, nil, nil)
	if got := CheckSafety(g, 0, SafetyOptions{}); got.Status != 1 {
		t.Fatalf("wrapper ignored false fairness: %+v", got)
	}
}

func TestLTLFairnessIsConjunctiveAndLoopLocal(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.AddOutput(aig.Edge(1))
	n, err := NewNetwork(g, []string{"x"}, []string{"x", "never"})
	if err != nil {
		t.Fatal(err)
	}
	m, err := LivenessToSafety(n, TransformOptions{
		Mode: IGNORE_SAFETY_KEEP_LIVENESS_MODE,
		LTL:  []string{":G :F never"}, Fairness: []aig.Edge{x, x.Not()},
	})
	if err != nil {
		t.Fatal(err)
	}
	state := make([]bool, registers(m.Graph))
	step := func(x, save bool) bool {
		values := append([]bool{x, save}, state...)
		out, err := m.Graph.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		copy(state, out[trueOutputs(m.Graph):])
		return out[0]
	}
	// A true x before SAVE is outside the loop and cannot satisfy fairness.
	if step(true, false) || step(false, true) || step(false, false) || step(false, false) {
		t.Fatal("a pre-SAVE event satisfied loop fairness")
	}
	// Both x and !x must occur inside the loop before it may be reported.
	if step(true, false) {
		t.Fatal("fairness used the current rather than recorded event")
	}
	if !step(false, false) {
		t.Fatal("both loop fairness events should permit the violation")
	}
}
