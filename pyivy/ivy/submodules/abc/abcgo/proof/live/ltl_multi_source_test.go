package live

import (
	"testing"
)

func TestSourceLTLReusesGFMonitorAcrossOutputs(t *testing.T) {
	n := liveFixture(false)
	formula := ":G :F assert_fair_q"
	n.LTLProperties = []string{formula}
	single := LivenessToSafetyTransformationWithLTL(IGNORE_SAFETY_KEEP_LIVENESS_MODE, n, n.Graph, nil, nil, nil, nil)
	n.LTLProperties = []string{formula, formula}
	twice := LivenessToSafetyTransformationWithLTL(IGNORE_SAFETY_KEEP_LIVENESS_MODE, n, n.Graph, nil, nil, nil, nil)
	if registers(twice) != registers(single) {
		t.Fatalf("same C GF signal should reuse monitor: one=%d two=%d", registers(single), registers(twice))
	}
	if trueOutputs(twice) != 2 {
		t.Fatalf("expected two LTL outputs, got %d", trueOutputs(twice))
	}
	for bits := 0; bits < 16; bits++ {
		state := make([]bool, registers(twice))
		for frame := 0; frame < 4; frame++ {
			outs, next := evaluateStep(twice, state, []bool{bits&(1<<frame) != 0})
			if outs[0] != outs[1] {
				t.Fatalf("pattern %d frame %d outputs %v", bits, frame, outs)
			}
			state = next
		}
	}
}
