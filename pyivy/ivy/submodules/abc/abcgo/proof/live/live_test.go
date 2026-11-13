package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"slices"
	"testing"
	"time"
)

func liveFixture(oscillate bool) *Network {
	p := aig.Aig_ManStart(20)
	state := aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ObjCreateCo(p, 0)
	next := state
	if oscillate {
		next = next.Not()
	}
	aig.Aig_ObjCreateCo(p, next)
	aig.Aig_ManSetRegNum(p, 1)
	n, err := NewNetwork(p, []string{"state"}, []string{"assert_fair_q", "assume_fair_true", "next"})
	if err != nil {
		panic(err)
	}
	return n
}
func TestSafetyReachabilityAndLimits(t *testing.T) {
	oscillator := liveFixture(true)
	result := CheckSafety(oscillator.Graph, 0, SafetyOptions{})
	if result.Status != 0 || len(result.Inputs) != 2 || len(result.States) != 2 || !result.States[1][0] {
		t.Fatalf("bad oscillator witness %+v", result)
	}
	if got := CheckSafety(liveFixture(false).Graph, 0, SafetyOptions{}); got.Status != 1 {
		t.Fatalf("zero latch should be safe %+v", got)
	}
	if got := CheckSafety(oscillator.Graph, 0, SafetyOptions{MaxStates: 1}); got.Status != -1 || got.Error == nil {
		t.Fatal("unsupported state limit must return unknown with an error")
	}
}
func TestBiereLoopAndFairness(t *testing.T) {
	for _, restart := range []bool{false, true} {
		for _, oscillate := range []bool{false, true} {
			n := liveFixture(oscillate)
			transformed, err := LivenessToSafety(n, TransformOptions{Mode: 2, Liveness: populateLivenessVector(n, n.Graph), Fairness: populateFairnessVector(n, n.Graph), Restartable: restart})
			if err != nil {
				t.Fatal(err)
			}
			result := CheckSafety(transformed.Graph, 0, SafetyOptions{})
			want := int32(0)
			if oscillate {
				want = 1
			}
			if result.Status != want {
				t.Fatalf("restart%v oscillate%v status%d want%d", restart, oscillate, result.Status, want)
			}
			if registers(transformed.Graph) != 5 || trueInputs(transformed.Graph) != 1 {
				t.Fatal("monitor interface")
			}
		}
	}
	n := liveFixture(false)
	result, err := LivenessToSafety(n, TransformOptions{Mode: 2, Liveness: []aig.Edge{n.Graph.Outputs()[0]}, Fairness: []aig.Edge{1}})
	if err != nil {
		t.Fatal(err)
	}
	if CheckSafety(result.Graph, 0, SafetyOptions{}).Status != 1 {
		t.Fatal("false fairness must disable unfair loop")
	}
}
func TestLTLPipeline(t *testing.T) {
	n := liveFixture(true)
	n.OutputNames[1] = "nq"
	// Replace the fair-true output by the complementary state signal.
	aig.Aig_ObjPatchFanin0(n.Graph, n.Graph.Outputs()[1], n.Graph.Inputs()[0].Not())
	formula := "& :G :F assert_fair_q :G :F nq"
	node, err := ParseFormula(formula)
	if err != nil || node.String() != formula || isWellFormed(node) != 1 {
		t.Fatalf("parse %v %v", node, err)
	}
	result, err := LivenessToSafety(n, TransformOptions{Mode: 2, LTL: []string{formula}})
	if err != nil {
		t.Fatal(err)
	}
	if CheckSafety(result.Graph, 0, SafetyOptions{}).Status != 1 {
		t.Fatal("oscillator satisfies both recurrent phases")
	}
	for _, bad := range []string{"", ":G", ":Q x", "& x", "x y"} {
		if _, err := ParseFormula(bad); err == nil {
			t.Fatal("malformed formula accepted", bad)
		}
	}
	if isWellFormed(parseFormulaCreateAST(":G assert_fair_q")) != 0 {
		t.Fatal("unsupported temporal fragment accepted")
	}
}
func TestCombinationOrderAndBoundary(t *testing.T) {
	if countCombination(0, 0) != 1 || countCombination(5, 0) != 1 || countCombination(5, 2) != 10 {
		t.Fatal("combination count")
	}
	want := [][]int{{0, 1}, {0, 2}, {1, 2}, {0, 3}, {1, 3}, {2, 3}}
	got := Combinations(4, 2)
	if len(got) != len(want) {
		t.Fatal(got)
	}
	for i := range want {
		if !slices.Equal(got[i], want[i]) {
			t.Fatalf("colex%d %v", i, got[i])
		}
	}
}

func TestL2SCommandFlagsAndAbstraction(t *testing.T) {
	frame := &Frame{Network: liveFixture(true)}
	if Abc_CommandAbcLivenessToSafety(frame, 2, []string{"l2s", "-s"}) != 0 || trueOutputs(frame.Network.Graph) != 1 || trueInputs(frame.Network.Graph) != 1 {
		t.Fatal("-s must ignore safety and preserve liveness")
	}
	frame = &Frame{Network: liveFixture(true)}
	if Abc_CommandAbcLivenessToSafety(frame, 2, []string{"l2s", "-l"}) != 0 || registers(frame.Network.Graph) != 1 || trueInputs(frame.Network.Graph) != 0 {
		t.Fatal("-l must ignore liveness")
	}
	frame = &Frame{Network: liveFixture(true)}
	if Abc_CommandAbcLivenessToSafetyAbstraction(frame, 2, []string{"l3s", "-s"}) != 0 || registers(frame.Network.Graph) != 4 {
		t.Fatal("default abstraction should shadow half the original latches")
	}
}

func monotoneFixture() *Network {
	p := aig.Aig_ManStart(30)
	input, state := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	aig.Aig_ObjCreateCo(p, 0)
	aig.Aig_ObjCreateCo(p, state)
	aig.Aig_ObjCreateCo(p, 1)
	aig.Aig_ObjCreateCo(p, state.Not())
	aig.Aig_ObjCreateCo(p, 0)
	aig.Aig_ObjCreateCo(p, aig.Aig_Or(p, input, state))
	aig.Aig_ManSetRegNum(p, 1)
	n, _ := NewNetwork(p, []string{"input", "state"}, []string{"pendingSignal", "hint_up", "unrelated", "hint_down", "csSafetyInvar_true", "next"})
	return n
}
func TestMonotoneAndDisjunctiveDiscovery(t *testing.T) {
	n := monotoneFixture()
	if got := FindMonotone(n, SafetyOptions{}, nil); !slices.Equal(got, []int32{1}) {
		t.Fatalf("monotone candidates %v", got)
	}
	got := FindDisjunctiveMonotone(n, SafetyOptions{}, nil)
	if len(got) != 2 || !slices.Equal(got[0], []int32{1}) || !slices.Equal(got[1], []int32{1, 3}) {
		t.Fatalf("disjunction levels %v", got)
	}
	unknown := func(*aig.Graph, int, SafetyOptions) SafetyResult { return SafetyResult{Status: -1} }
	if found := FindMonotone(n, SafetyOptions{}, unknown); len(found) != 0 {
		t.Fatal("unknown result classified as proof")
	}
}
func TestKLiveAbsorberProofBoundary(t *testing.T) {
	for _, oscillate := range []bool{false, true} {
		n := liveFixture(oscillate)
		n.OutputNames[0] = "csLiveTarget_signal"
		if !oscillate {
			aig.Aig_ObjPatchFanin0(n.Graph, n.Graph.Outputs()[0], 0)
		}
		result := KLive(n, KLiveOptions{MaxAbsorbers: 2})
		if !oscillate && result.Status != 1 {
			t.Fatalf("constant true target should prove %+v", result)
		}
		if oscillate && (result.Status != -1 || result.Absorbers != 2 || result.LastSafety.Status != 0) {
			t.Fatalf("finite absorber failure is not a liveness disproof %+v", result)
		}
	}
}

func TestFullOneStepRetainsSafetyAssertions(t *testing.T) {
	n := liveFixture(false)
	result, err := LivenessToSafety(n, TransformOptions{Mode: 4, Liveness: populateLivenessVector(n, n.Graph), SafetyAssertions: []aig.Edge{1}})
	if err != nil {
		t.Fatal(err)
	}
	if trueOutputs(result.Graph) != 2 {
		t.Fatal("full one-step mode discarded safety output")
	}
	outputs, _ := evaluateStep(result.Graph, make([]bool, registers(result.Graph)), []bool{false})
	if !outputs[0] || outputs[1] {
		t.Fatal("safety violation should remain visible independently of SAVE")
	}
}

func TestInvalidLTLReportsError(t *testing.T) {
	n := liveFixture(true)
	if _, err := LivenessToSafety(n, TransformOptions{Mode: 2, LTL: []string{":G :F missing_signal"}}); err == nil {
		t.Fatal("unknown LTL signal accepted")
	}
}
func TestMonotoneOnlyIsNotLivenessProof(t *testing.T) {
	result := KLive(monotoneFixture(), KLiveOptions{Directive: kCS_WITH_DISCOVER_MONOTONE_SIGNALS})
	if result.Status == 1 {
		t.Fatal("preprocessing-only run reported a liveness proof")
	}
}

func TestSimulationCommandHasOnlyLivenessOutput(t *testing.T) {
	frame := &Frame{Network: liveFixture(true)}
	if Abc_CommandAbcLivenessToSafetySim(frame, 1, []string{"l2ssim"}) != 0 || trueOutputs(frame.Network.Graph) != 1 {
		t.Fatal("simulation command added a safety output")
	}
}

func TestSafetyNameAliasesAndDeadline(t *testing.T) {
	n := liveFixture(true)
	n.OutputNames[0], n.OutputNames[1] = "Assert_ready", "Assume_environment"
	if len(populateSafetyAssertionVector(n, n.Graph)) != 1 || len(populateSafetyAssumptionVector(n, n.Graph)) != 1 {
		t.Fatal("source safety name aliases were lost")
	}
	n.OutputNames[0] = "csLiveTarget_signal"
	calls := 0
	checker := func(*aig.Graph, int, SafetyOptions) SafetyResult { calls++; return SafetyResult{Status: 0} }
	result := KLive(n, KLiveOptions{Checker: checker, Safety: SafetyOptions{Deadline: time.Now().Add(-time.Second)}})
	if result.Status != -1 || calls != 0 {
		t.Fatal("expired deadline invoked a proof engine")
	}
}
