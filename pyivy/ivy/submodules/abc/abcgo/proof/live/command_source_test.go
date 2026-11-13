package live

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestNChooseKReadsWidthFromInput(t *testing.T) {
	n := liveFixture(false)
	var out bytes.Buffer
	frame := &Frame{Network: n, In: strings.NewReader("1\n"), Out: &out}
	if got := Abc_CommandNChooseK(frame, 1, []string{"nck"}); got != 1 {
		t.Fatalf("source success return: %d", got)
	}
	if frame.Network.Graph.RegisterCount() != 2 {
		t.Fatalf("missing generated monitor: %d", frame.Network.Graph.RegisterCount())
	}
	if !strings.Contains(out.String(), "Enter parameterizedCombK") {
		t.Fatalf("missing prompt: %q", out.String())
	}
}

// kliveness.c uses absorberCount < 500; the 500th safety query is never run.
func TestKLiveSourceDefaultAbsorberBound(t *testing.T) {
	n := liveFixture(false)
	n.OutputNames[0] = "csLiveTarget_signal"
	calls := 0
	result := KLive(n, KLiveOptions{Checker: func(_ *aig.Graph, _ int, _ SafetyOptions) SafetyResult {
		calls++
		if calls == 500 {
			return SafetyResult{Status: 1}
		}
		return SafetyResult{Status: -1}
	}})
	if calls != 499 || result.Absorbers != 499 || result.Status != -1 {
		t.Fatalf("calls %d result %+v", calls, result)
	}
}

func TestNChooseKInputFailureLeavesNetwork(t *testing.T) {
	n := liveFixture(false)
	var out bytes.Buffer
	frame := &Frame{Network: n, In: strings.NewReader("not-an-integer"), Out: &out}
	if got := Abc_CommandNChooseK(frame, 1, []string{"nck"}); got != 0 || frame.Network != n || !strings.Contains(out.String(), "Failed to read integer") {
		t.Fatalf("result=%d output=%q", got, out.String())
	}
}

func TestCommandsRejectNonSourceOptions(t *testing.T) {
	frame := &Frame{Network: liveFixture(false), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if got := Abc_CommandAbcLivenessToSafety(frame, 3, []string{"l2s", "-a", "1"}); got != 1 {
		t.Fatalf("non-source -a accepted: %d", got)
	}
	if got := Abc_CommandCS_kLiveness(frame, 3, []string{"klive", "-K", "1"}); got != 1 {
		t.Fatalf("non-source -K accepted: %d", got)
	}
}

func TestCombinedSourceFlagsAndSimulationOptions(t *testing.T) {
	frame := &Frame{Network: liveFixture(false), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if got := Abc_CommandAbcLivenessToSafety(frame, 2, []string{"l2s", "-1s"}); got != 0 {
		t.Fatalf("combined flags rejected: %d", got)
	}
	frame.Network = liveFixture(false)
	if got := Abc_CommandAbcLivenessToSafetySim(frame, 2, []string{"l2ssim", "-s"}); got != 0 || registers(frame.Network.Graph) <= 1 {
		t.Fatalf("simulation did not select full mode for unknown flag: %d", got)
	}
}

func TestL2SRepeatedOptionsFollowSource(t *testing.T) {
	for _, flag := range []string{"-s", "-l"} {
		frame := &Frame{Network: liveFixture(false)}
		if got := Abc_CommandAbcLivenessToSafety(frame, 3, []string{"l2s", flag, flag}); got != 1 {
			t.Fatalf("repeated %s accepted", flag)
		}
	}
	defer func() {
		if recover() == nil {
			t.Error("repeated -1 did not assert")
		}
	}()
	Abc_CommandAbcLivenessToSafety(&Frame{Network: liveFixture(false)}, 3, []string{"l2s", "-1", "-1"})
}

func TestL2SSafetyOnlyAllowsRepeatedOneStep(t *testing.T) {
	for _, args := range [][]string{{"l2s", "-l", "-1", "-1"}, {"l2s", "-1", "-l", "-1"}} {
		frame := &Frame{Network: liveFixture(false)}
		if got := Abc_CommandAbcLivenessToSafety(frame, int32(len(args)), args); got != 0 {
			t.Fatalf("source accepts %v: %d", args, got)
		}
		if registers(frame.Network.Graph) != 1 || trueInputs(frame.Network.Graph) != 0 {
			t.Fatalf("safety-only mode was altered for %v", args)
		}
	}
}
