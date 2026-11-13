package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func constraintMonitorFixture() *Network {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.Const1())  // pending
	g.AddOutput(state)       // hint
	g.AddOutput(state.Not()) // target
	g.AddOutput(g.Const1())  // invariant
	g.AddOutput(x)           // constraint violation
	g.AddOutput(x)           // next state
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	n, err := NewNetwork(g, nil, []string{"pendingSignal", "hint_0", "csLiveTarget_", "csSafetyInvar_", "hint_constraint", "state_next"})
	if err != nil {
		panic(err)
	}
	return n
}
func checkMonitorConstraint(t *testing.T, g *aig.Graph) {
	t.Helper()
	if got := aig.Aig_ManConstrNum(g); got != 1 {
		t.Fatalf("constraint count %d", got)
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	for _, x := range []bool{false, true} {
		values := make([]bool, g.InputCount())
		values[0] = x
		out, err := g.Evaluate(values)
		if err != nil || out[trueOutputs(g)-1] != x {
			t.Fatalf("constraint x=%v output %v err=%v", x, out, err)
		}
	}
}
func TestConstraintMetadataThroughMonitorBuilders(t *testing.T) {
	n := constraintMonitorFixture()
	g := n.Graph
	indexes := &aigPoIndices{attrPendingSignalIndex: 0, attrSafetyInvarIndex: 3}
	var start int32
	mono := createMonotoneTester(g, indexes, &monotoneVectorsStruct{attrCandMonotone: []int32{1}, attrHintMonotone: []int32{1}}, &start)
	if start != 4 {
		t.Fatalf("monotone start %d", start)
	}
	checkMonitorConstraint(t, mono)
	checkMonitorConstraint(t, negateOutputs(mono, int(start)))
	disj := createDisjunctiveMonotoneTester(g, indexes, &antecedentConsequentVectorsStruct{attrConsequentCandidates: []int32{1}}, &start)
	if start != 4 {
		t.Fatalf("disjunctive start %d", start)
	}
	checkMonitorConstraint(t, disj)
	checkMonitorConstraint(t, negateOutputs(disj, int(start)))
	checkMonitorConstraint(t, disjunctiveTester(g, []aig.Edge{g.Outputs()[1]}, 1))
	zero := int32(-1)
	working := createNewAigWith0LivePo(g, []aig.Edge{outputDriver(g, 2)}, &zero)
	if zero != 4 {
		t.Fatalf("zero-live index %d", zero)
	}
	checkMonitorConstraint(t, working)
	target := int32(-1)
	first := introduceAbsorberLogic(working, &zero, &target, 0)
	checkMonitorConstraint(t, first)
	second := introduceAbsorberLogic(first, &zero, &target, 1)
	checkMonitorConstraint(t, second)
	checkMonitorConstraint(t, kLiveProofModel(second, 0, target, -1))
	arena := createNewAigWith0LivePoWithDSC(g, []aig.Edge{outputDriver(g, 2)}, nil, 0, 0, 0, [][]int32{{1}})
	checkMonitorConstraint(t, arena)
	if hints := findHintOutputs(n); len(hints) != 1 || hints[0] != 1 {
		t.Fatalf("constraint selected as hint: %v", hints)
	}
}

func TestKLiveUsesNativeConstraints(t *testing.T) {
	constrained := constraintMonitorFixture()
	result := KLive(constrained, KLiveOptions{Directive: 0, MaxAbsorbers: 3})
	if result.Status != 1 {
		t.Fatalf("constrained stabilization status %d error %v", result.Status, result.Error)
	}
	// Dropping the constraint permits x=1 forever, making the target false forever.
	unconstrained := constraintMonitorFixture()
	aig.Aig_ManSetConstrNum(unconstrained.Graph, 0)
	result = KLive(unconstrained, KLiveOptions{Directive: 0, MaxAbsorbers: 3})
	if result.Status == 1 {
		t.Fatal("unconstrained recurrence was incorrectly proved")
	}
}
