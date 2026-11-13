package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
	"testing"
)

func TestSourceAbstractionShadowChoices(t *testing.T) {
	p := aig.Aig_ManStart(20)
	in, s0, s1 := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, s0)
	aig.Aig_ObjCreateCo(p, s1)
	aig.Aig_ObjCreateCo(p, in)
	aig.Aig_ObjCreateCo(p, s0)
	aig.Aig_ManSetRegNum(p, 2)
	n := &Network{Graph: p}
	live, fair := []aig.Edge{p.Outputs()[0]}, []aig.Edge{p.Outputs()[1]}
	full := LivenessToSafetyTransformation(FULL_BIERE_MODE, n, p, live, fair, nil, nil)
	both := LivenessToSafetyTransformationAbs(FULL_BIERE_MODE, n, p, []int32{0, 1}, live, fair, nil, nil)
	none := LivenessToSafetyTransformationAbs(FULL_BIERE_MODE, n, p, []int32{}, live, fair, nil, nil)
	if registers(full) != registers(both) || registers(full)-registers(none) != 2 {
		t.Fatalf("full=%d both=%d none=%d registers", registers(full), registers(both), registers(none))
	}
	a, err := full.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	b, err := both.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 256; bits++ {
		if !slices.Equal(evalAiger(a, bits, 4), evalAiger(b, bits, 4)) {
			t.Fatalf("full abstraction changed trace %d", bits)
		}
	}
}

// prepareFlopVector only builds 0..length-1 in C; it does not inspect the
// AIG's register count. Validation, if needed, happens in later consumers.
func TestSourcePrepareFlopVectorDoesNotRestrictLengthToRegisters(t *testing.T) {
	p := aig.NewGraph()
	if got := prepareFlopVector(p, 3); !slices.Equal(got, []int32{0, 1, 2}) {
		t.Fatalf("vector %v", got)
	}
}

// In liveness.c an abstraction shadows only the selected registers. With a
// toggling unshadowed latch, the abstract checker can therefore report a loop
// one cycle before the full-state construction, even under satisfied fairness.
func TestSourceAbstractionShadowsOnlySelectedRegister(t *testing.T) {
	g := aig.NewGraph()
	s0, s1 := g.NewInput(), g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const1())
	g.AddOutput(s0)
	g.AddOutput(s1.Not())
	aig.Aig_ManSetRegNum(g, 2)
	n := &Network{Graph: g}
	live, fair := []aig.Edge{g.Outputs()[0]}, []aig.Edge{g.Outputs()[1]}
	full := LivenessToSafetyTransformation(IGNORE_SAFETY_KEEP_LIVENESS_MODE, n, g, live, fair, nil, nil)
	partial := LivenessToSafetyTransformationAbs(IGNORE_SAFETY_KEEP_LIVENESS_MODE, n, g, []int32{0}, live, fair, nil, nil)
	fullState, partState := make([]bool, registers(full)), make([]bool, registers(partial))
	fullOut, fullState := evaluateStep(full, fullState, []bool{true})
	partOut, partState := evaluateStep(partial, partState, []bool{true})
	if fullOut[0] || partOut[0] {
		t.Fatal("loop asserted before state was saved")
	}
	fullOut, _ = evaluateStep(full, fullState, []bool{false})
	partOut, _ = evaluateStep(partial, partState, []bool{false})
	if fullOut[0] || !partOut[0] {
		t.Fatalf("full=%v partial=%v", fullOut, partOut)
	}
}
