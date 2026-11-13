package live

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// ABC's proof-oriented one-step constructor returns the input pointer before
// building a monitor on a combinational circuit. This was also verified with
// a compiled C fixture using one input and one output.
func TestOneStepCombinationalGraphReturnsOriginal(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	n, err := NewNetwork(g, []string{"x"}, []string{"out"})
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []int32{IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE, FULL_BIERE_ONE_LOOP_MODE} {
		got, err := LivenessToSafety(n, TransformOptions{Mode: mode})
		if err != nil || got != n || got.Graph != g {
			t.Fatalf("mode %d did not return original: %v %v", mode, got, err)
		}
		if got := LivenessToSafetyTransformationOneStepLoop(mode, n, g, nil, nil, nil, nil); got != g {
			t.Fatalf("mode %d source entry did not return original graph", mode)
		}
	}
	if got := LivenessToSafetyTransformationOneStepLoop(IGNORE_LIVENESS_KEEP_SAFETY_MODE, n, g, nil, nil, nil, nil); got != g {
		t.Fatal("source checks for registers before considering mode")
	}
}
