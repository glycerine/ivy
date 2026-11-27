package abci

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var giaBmcDesigns = []string{"seq/s713.blif", "seq/s1494.blif", "native_s27_drw.blif"}

var giaBmcOptions = []string{"", "-v", "-F 10 -v", "-S 3 -v", "-A 5 -v", "-A 0 -F 6 -v", "-T 5 -v", "-K 4 -v", "-c -v", "-s -v -F 6", "-w -F 6", "-v -w"}

// &bmc is a line-by-line port of bmcBmcAnd.c (Gia_ManBmcPerform): ternary
// simulation, the cone unrolling, the Mf CNF, the per-frame SAT calls and
// the -v statistics equal C's; write_cex writes the same file.
func TestAbc9BmcLikeNative(t *testing.T) {
	dir := t.TempDir()
	nativeFile, portFile := filepath.Join(dir, "native.cex"), filepath.Join(dir, "port.cex")
	var scripts []string
	for _, file := range giaBmcDesigns {
		for _, opts := range giaBmcOptions {
			scripts = append(scripts, "read test_vectors/"+file+"; strash; &get; &bmc "+opts+"; print_status")
		}
	}
	for _, file := range []string{"seq/s713.blif", "seq/s1494.blif"} {
		scripts = append(scripts, "read test_vectors/"+file+"; strash; dc2; miter test_vectors/"+file+"; &get; &bmc -F 20 -v; print_status")
	}
	for _, script := range scripts {
		os.Remove(nativeFile)
		os.Remove(portFile)
		native := bmc3Exact(runNativeABC(t, script+"; write_cex "+nativeFile))
		port, _ := runScriptStatus(t, script+"; write_cex "+portFile)
		if port = bmc3Exact(port); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
		n, err1 := os.ReadFile(nativeFile)
		g, err2 := os.ReadFile(portFile)
		if (err1 == nil) != (err2 == nil) || string(n) != string(g) {
			t.Errorf("%s: write_cex differs (%v, %v)\nnative:\n%s\nport:\n%s", script, err1, err2, n, g)
		}
	}
}

// &bmc -d dumps the unrolled frames into frames.aig in the current
// directory; the Go file equals C's.
func TestAbc9BmcDumpFramesLikeNative(t *testing.T) {
	defer os.Remove("frames.aig")
	for _, file := range giaBmcDesigns {
		script := "read test_vectors/" + file + "; strash; &get; &bmc -d -F 6; print_status"
		os.Remove("frames.aig")
		native := bmc3Exact(runNativeABC(t, script))
		n, err1 := os.ReadFile("frames.aig")
		os.Remove("frames.aig")
		port, _ := runScriptStatus(t, script)
		g, err2 := os.ReadFile("frames.aig")
		if port = bmc3Exact(port); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
		if err1 != nil || err2 != nil || string(n) != string(g) {
			t.Errorf("%s: frames.aig differs (%v, %v)", script, err1, err2)
		}
	}
}

// runNativeABCTimeout runs native abc with a time limit: the &bmc bugs
// below made it loop forever while allocating memory.
func runNativeABCTimeout(t *testing.T, script string, limit time.Duration) (string, bool) {
	t.Helper()
	bin := os.Getenv("ABC_BINARY")
	if bin == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	out, _ := exec.CommandContext(ctx, bin, "-c", script).CombinedOutput()
	return ansiEscape.ReplaceAllString(string(out), ""), ctx.Err() == nil
}

// Each script's &bmc result, as bmc3 and "&bmc -F 1" confirm.
var giaBmcBugScripts = []struct{ script, want string }{
	// CBUGS_FIXED.md #39: Gia_ManBmcUnroll stopped its backward pass at the
	// first frame (from the end) whose cones need no earlier state and
	// dropped every earlier frame, x-valued outputs included. The frame-0
	// failure was missed (Status = -1, Frames = 49); on s713 the reported
	// CEX was in frame 1 although output 8 fails in frame 0 (and output 0,
	// constant 1 in frame 0, is #40).
	{"read test_vectors/seq/bmc_and_early.blif; strash; &get; &bmc; print_status", "Output 0 of miter \"bmc_and_early\" was asserted in frame 0."},
	{"read test_vectors/seq/s713.blif; strash; &get; &bmc; print_status", "Status = 0  Frames = 0   CEX: Po =   0  Frame =   0"},
	// CBUGS_FIXED.md #40: an output that ternary simulation proves 1 was
	// unrolled as constant 0, and a constant-1 output was skipped as
	// "trivially SAT"; output 0 (1 in frame 0) was never reported.
	{"read test_vectors/seq/bmc_and_const1.blif; strash; &get; &bmc; print_status", "Output 0 of miter \"bmc_and_const1\" was asserted in frame 0."},
	{"read test_vectors/seq/bmc_and_const1.blif; strash; &get; &bmc -F 1; print_status", "Status = 0  Frames = 0   CEX: Po =   0  Frame =   0"},
	// CBUGS_FIXED.md #41: Bmc_MnaTernary stops nFramesAdd frames after the
	// first x-valued output, so it never stopped (allocating a state per
	// frame) when no output is ever x-valued, or with -A 0 and no -F.
	{"read test_vectors/seq/bmc_and_zero.blif; strash; &get; &bmc -v; print_status", "Status = -1  Frames = 0"},
	{"read test_vectors/seq/s27.blif; strash; &get; &bmc -A 0 -v; print_status", "Status = "},
}

func TestNativeAbc9BmcFindsEveryFailure(t *testing.T) {
	for _, c := range giaBmcBugScripts {
		out, done := runNativeABCTimeout(t, c.script, 10*time.Second)
		if !done || !strings.Contains(out, c.want) {
			t.Errorf("%s (finished: %v): want %q in\n%s", c.script, done, c.want, out)
		}
	}
}

func TestAbc9BmcBugCasesLikeNative(t *testing.T) {
	for _, c := range giaBmcBugScripts {
		native, done := runNativeABCTimeout(t, c.script, 10*time.Second)
		if !done {
			t.Fatalf("%s: native abc did not finish", c.script)
		}
		port, _ := runScriptStatus(t, c.script)
		if n, g := bmc3Exact(native), bmc3Exact(port); n != g {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", c.script, n, g)
		}
	}
}
