package live

import (
	"bytes"
	"strings"
	"testing"
)

func TestSourceCommandErrorsPreserveNetwork(t *testing.T) {
	for _, args := range [][]string{{"l2s", "-z"}, {"l2s", "-s", "-l"}, {"l2s", "-h"}, {"l2s", "operand"}} {
		n := liveFixture(false)
		frame := &Frame{Network: n, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		if got := Abc_CommandAbcLivenessToSafety(frame, int32(len(args)), args); got != 1 || frame.Network != n {
			t.Fatalf("args %v: status=%d network changed=%v", args, got, frame.Network != n)
		}
	}
	n := liveFixture(false)
	frame := &Frame{Network: n, In: strings.NewReader("2\n"), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if got := Abc_CommandNChooseK(frame, 1, []string{"nck"}); got != 1 || frame.Network != n {
		t.Fatalf("too-wide combination: status=%d network changed=%v", got, frame.Network != n)
	}
	frame = &Frame{Network: n, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if got := Abc_CommandCS_kLiveness(frame, 1, []string{"klive"}); got != 1 || frame.Network != n {
		t.Fatalf("missing source target: status=%d network changed=%v", got, frame.Network != n)
	}
}

// Extra_UtilGetopt stops at --, a lone -, or the first non-option. The
// source nck command ignores the remaining arguments and still reads stdin.
func TestNChooseKGetoptStopsBeforeRemainingArguments(t *testing.T) {
	for _, args := range [][]string{{"nck", "--", "-h"}, {"nck", "-", "-h"}, {"nck", "operand", "-h"}, {"nck", "-c", "--", "-h"}} {
		n := liveFixture(false)
		var out bytes.Buffer
		frame := &Frame{Network: n, In: strings.NewReader("bad"), Out: &out}
		if got := Abc_CommandNChooseK(frame, int32(len(args)), args); got != 0 || frame.Network != n || !strings.Contains(out.String(), "Failed to read integer!") {
			t.Fatalf("args %v: status=%d output=%q", args, got, out.String())
		}
	}
}

func TestNChooseKPrintsNewlineAfterReadingWidth(t *testing.T) {
	var out bytes.Buffer
	frame := &Frame{Network: liveFixture(false), In: strings.NewReader("1\n"), Out: &out}
	if got := Abc_CommandNChooseK(frame, 1, []string{"nck"}); got != 1 {
		t.Fatalf("status %d", got)
	}
	if out.String() != "Enter parameterizedCombK = \n" {
		t.Fatalf("output %q", out.String())
	}
}

// Once C's getopt encounters a non-option or --, it ignores the rest; a
// previously selected liveness directive remains in effect.
func TestLivenessGetoptStopsAfterValidDirective(t *testing.T) {
	for _, args := range [][]string{{"l2s", "-l", "--", "-h"}, {"l2s", "-l", "operand", "-h"}, {"l2s", "-l", "-", "-h"}} {
		frame := &Frame{Network: liveFixture(false), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		if got := Abc_CommandAbcLivenessToSafety(frame, int32(len(args)), args); got != 0 {
			t.Fatalf("args %v: status %d output %v", args, got, frame.Out)
		}
		if registers(frame.Network.Graph) != 1 || trueInputs(frame.Network.Graph) != 0 {
			t.Fatalf("args %v: wrong directive", args)
		}
	}
}

func TestKLiveGetoptStopsAfterValidDirective(t *testing.T) {
	frame := &Frame{Network: liveFixture(false), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	args := []string{"kcs", "-m", "--", "-h"}
	if got := Abc_CommandCS_kLiveness(frame, int32(len(args)), args); got != 0 {
		t.Fatalf("status %d, output %v", got, frame.Out)
	}
}

// Extra_UtilGetopt reports an unknown short option to stderr before the
// command prints its usage message.
func TestSourceUnknownOptionDiagnostic(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*Frame, int32, []string) int32
	}{{"l2s", Abc_CommandAbcLivenessToSafety}, {"kcs", Abc_CommandCS_kLiveness}, {"nck", Abc_CommandNChooseK}} {
		var out, errout bytes.Buffer
		frame := &Frame{Network: liveFixture(false), Out: &out, Err: &errout}
		args := []string{tc.name, "-z"}
		if status := tc.run(frame, 2, args); status != 1 || errout.String() != tc.name+": unknown option z\n" || !strings.HasPrefix(out.String(), "usage: "+tc.name) {
			t.Fatalf("%s status=%d out=%q err=%q", tc.name, status, out.String(), errout.String())
		}
	}
}

func TestSimulationCommandReportsOnlyFirstUnknownOption(t *testing.T) {
	for _, tc := range []struct{ arg, want string }{{"-s", "l2ssim: unknown option s\n"}, {"-1z", ""}, {"--", ""}} {
		var out, errors bytes.Buffer
		frame := &Frame{Network: liveFixture(false), Out: &out, Err: &errors}
		args := []string{"l2ssim", tc.arg}
		if got := Abc_CommandAbcLivenessToSafetySim(frame, 2, args); got != 0 || errors.String() != tc.want {
			t.Fatalf("%q: status=%d errors=%q", tc.arg, got, errors.String())
		}
	}
}
