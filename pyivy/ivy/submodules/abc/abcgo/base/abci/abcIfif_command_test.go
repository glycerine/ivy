package abci

import (
	"math"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/util"
)

func captureIfifPrint(t *testing.T, run func() int32) (int32, string) {
	t.Helper()
	previous := util.PrintBridge
	var output strings.Builder
	util.PrintBridge = func(text string) { output.WriteString(text) }
	defer func() { util.PrintBridge = previous }()
	return run(), output.String()
}

func TestAbcCommandIfifHelpAndMissingArgument(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	status, output := captureIfifPrint(t, func() int32 {
		return Abc_CommandIfif(frame, 2, []string{"ifif", "-h"})
	})
	if status != 1 || !strings.Contains(output, "usage: ifif [-DNcvwh]") || !strings.Contains(output, "default = 0.50") || !strings.Contains(output, "default = cluster") {
		t.Fatalf("help status=%d output=%q", status, output)
	}

	status, output = captureIfifPrint(t, func() int32 {
		return Abc_CommandIfif(frame, 2, []string{"ifif", "-D"})
	})
	if status != 1 || !strings.Contains(output, "switch \"-D\" should be followed by a floating point number") || !strings.Contains(output, "usage: ifif") {
		t.Fatalf("missing-value status=%d output=%q", status, output)
	}
}

func TestAbcCommandIfifNetworkAndLibraryErrors(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	status, output := captureIfifPrint(t, func() int32 {
		return Abc_CommandIfif(frame, 1, []string{"ifif"})
	})
	if status != 1 || output != "Empty network.\n" {
		t.Fatalf("empty network status=%d output=%q", status, output)
	}

	previousLib := ifpkg.Abc_FrameReadLibLut()
	ifpkg.Abc_FrameSetLibLut(nil)
	defer ifpkg.Abc_FrameSetLibLut(previousLib)
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	frame.SetCurrentNetwork(network)
	status, output = captureIfifPrint(t, func() int32 {
		return Abc_CommandIfif(frame, 1, []string{"ifif"})
	})
	if status != 1 || output != "LUT library is not given.\n" {
		t.Fatalf("missing library status=%d output=%q", status, output)
	}
}

func TestAbcAtofUsesNumericPrefix(t *testing.T) {
	for input, want := range map[string]float32{"  +0.25tail": .25, "-2e1suffix": -20, "not-a-number": 0} {
		if got := abcAtof(input); got != want {
			t.Errorf("abcAtof(%q) = %g, want %g", input, got, want)
		}
	}
	if got := abcAtof("-infinitytail"); !math.IsInf(float64(got), -1) {
		t.Errorf("abcAtof(-infinitytail) = %g, want negative infinity", got)
	}
	if got := abcAtof("nan(payload)"); !math.IsNaN(float64(got)) {
		t.Errorf("abcAtof(nan(payload)) = %g, want NaN", got)
	}
}
