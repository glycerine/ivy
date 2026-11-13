package framgr

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/fracomb"
)

// This deterministic 3-SAT miter exceeds the initial 600-conflict limit, then
// is found satisfiable by the source CEC sequence using the real FRAIG sweep.
func TestCECDriverFallbackSATWitness(t *testing.T) {
	r := rand.New(rand.NewSource(623))
	g := aig.NewGraph()
	vars := make([]aig.Edge, 100)
	for i := range vars {
		vars[i] = g.NewInput()
		r.Intn(2) // Preserve the seed sequence of the generated fixture.
	}
	formula := g.Const1()
	for i := 0; i < 420; i++ {
		clause := g.Const0()
		for j := 0; j < 3; j++ {
			v := vars[r.Intn(len(vars))]
			if r.Intn(2) == 0 {
				v = v.Not()
			}
			clause = g.OR(clause, v)
		}
		formula = g.AND(formula, clause)
	}
	g.AddOutput(formula)
	original := g
	if status := fracomb.Fra_FraigSat(g, 600, 0, 0, 0, 0, 1, 0, 0, 0); status != -1 {
		t.Fatalf("fixture did not exhaust initial SAT limit: %d", status)
	}
	calls := 0
	nodeLimit, miterLimit := int32(2), int32(300)
	status := fracomb.Fra_FraigCec(&g, 100000, 0, func(g *aig.Graph, n, m int32) *aig.Graph {
		if n != nodeLimit || m != miterLimit {
			t.Fatalf("sweep %d limits = (%d, %d), want (%d, %d)", calls, n, m, nodeLimit, miterLimit)
		}
		calls++
		nodeLimit *= 8
		miterLimit *= 2
		return SweepForCEC(g, n, m)
	})
	if status != 0 || calls < 2 {
		t.Fatalf("status=%d sweeps=%d", status, calls)
	}
	model, ok := g.Data.([]int32)
	// FRAIG simulation stores the source PO object ID after the PI values,
	// as ABC's Fra_SmlCheckOutputSavePattern does.
	if !ok || len(model) != len(vars)+1 {
		t.Fatalf("invalid witness %v", g.Data)
	}
	if want := int32(g.Outputs()[0].NodeID()); model[len(vars)] != want {
		t.Fatalf("source PO ID = %d, want %d", model[len(vars)], want)
	}
	inputs := make([]bool, len(vars))
	for i, bit := range model[:len(vars)] {
		inputs[i] = bit != 0
	}
	outputs, err := g.Evaluate(inputs)
	if err != nil || len(outputs) != 1 || !outputs[0] {
		t.Fatalf("invalid witness outputs=%v error=%v", outputs, err)
	}
	outputs, err = original.Evaluate(inputs)
	if err != nil || len(outputs) != 1 || !outputs[0] {
		t.Fatalf("witness does not satisfy original miter: outputs=%v error=%v", outputs, err)
	}
}
