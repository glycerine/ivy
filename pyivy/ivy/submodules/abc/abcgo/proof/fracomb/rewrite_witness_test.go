package fracomb

import (
	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
	"testing"
)

// ABC's CEC fallback rewrites the miter before its later SAT queries. The
// returned PI-order model must remain a witness on the original graph.
func TestRewrittenMiterSATWitnessReplaysOriginalInputs(t *testing.T) {
	for _, solver := range []int32{0} {
		for _, flip := range []int32{0, 1} {
			original := aig.NewGraph()
			original.NewInput()
			a, b, c := original.NewInput(), original.NewInput(), original.NewInput()
			original.AddOutput(original.AND(original.AND(a, b.Not()), c))
			original.AddOutput(original.AND(a.Not(), c))
			rewritten, err := dar.Dar_ManRwsat(original, true, false)
			if err != nil {
				t.Fatal(err)
			}
			if status := Fra_FraigSat(rewritten, 10000, 0, 0, 0, 0, flip, 0, solver, 0); status != 0 {
				t.Fatalf("solver=%d flip=%d status=%d", solver, flip, status)
			}
			model, ok := rewritten.Data.([]int32)
			if !ok || len(model) != 4 {
				t.Fatalf("bad model %v", rewritten.Data)
			}
			inputs := make([]bool, len(model))
			for i := range inputs {
				inputs[i] = model[i] != 0
			}
			got, err := original.Evaluate(inputs)
			if err != nil || (!got[0] && !got[1]) {
				t.Fatalf("solver=%d flip=%d input=%v original=%v err=%v", solver, flip, inputs, got, err)
			}
		}
	}
}
