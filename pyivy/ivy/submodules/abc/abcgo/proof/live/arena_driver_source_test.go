package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceArenaDriverBuildsWindowAndBarrierState(t *testing.T) {
	p := aig.Aig_ManStart(20)
	input, state := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	for _, e := range []aig.Edge{input, 0, 0, 0, input, state} {
		aig.Aig_ObjCreateCo(p, e)
	}
	aig.Aig_ManSetRegNum(p, 1)
	n := &Network{Graph: p, OutputNames: []string{"csLiveTarget_x", "windowBegins_x", "withinWindow_x", "pendingSignal", "csLevel1Stabil_x", "next"}}
	barriers := collectUserGivenDisjunctiveMonotoneSignals(n)
	if len(barriers) != 1 || len(barriers[0]) != 1 || barriers[0][0] != 4 {
		t.Fatalf("user barrier selection %v", barriers)
	}
	index := int32(-1)
	out := generateWorkingAigWithDSC(p, n, &index, barriers)
	if index != 5 || trueOutputs(out) != 6 || registers(out) != 5 {
		t.Fatalf("interface index=%d outputs=%d registers=%d", index, trueOutputs(out), registers(out))
	}
	states := make([]bool, registers(out))
	first, next := evaluateStep(out, states, []bool{true})
	if !first[index] {
		t.Fatalf("initial active target output %v", first)
	}
	second, _ := evaluateStep(out, next, []bool{false})
	if second[index] {
		t.Fatalf("falling barrier under pending must violate monotonicity: %v", second)
	}
}
