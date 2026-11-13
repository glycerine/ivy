package live

import (
	"bytes"
	"strings"
	"testing"
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
	if got := Abc_CommandAbcLivenessToSafetySim(frame, 2, []string{"l2ssim", "-s"}); got != 1 {
		t.Fatalf("simulation accepted unsupported flag: %d", got)
	}
}
