package abci

import (
	"context"
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	base_main "github.com/glycerine/abcgo/base/main"
)

// The liveness commands (src/proof/live, ported in proof/live): l2s,
// l2ssim, l3s, kcs and nck print what C prints and install the same
// network (print_stats, write_aiger and write_blif bytes), on designs with
// outputs named as liveness.c, kliveness.c, kLiveConstraints.c,
// arenaViolation.c and disjunctiveMonotone.c look for
// (test_vectors/live), on s27/s713/mul4 (no properties), and in the cases
// that crashed C (CBUGS_FIXED.md #47, #52, #53, #55-#57).

var livePercent = regexp.MustCompile(`\( *[0-9.]+ %\)`)
var livePreTime = regexp.MustCompile(`pre-processing time = [0-9.]+`)
var liveAigerDate = regexp.MustCompile(`written by ABC on [^\n]*`)

func liveNormalize(out string) string {
	out = livePercent.ReplaceAllString(out, "(P %)")
	out = livePreTime.ReplaceAllString(out, "pre-processing time = T")
	return printSymmNormalize(out)
}

// runNativeABCInput is runNativeABC with the given standard input (nck
// reads its parameter with scanf).
func runNativeABCInput(t *testing.T, script, input string) string {
	t.Helper()
	bin := testoracle.NativeABCPath()
	if bin == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	// C hung on nck with K above the register count (CBUGS_FIXED.md #57)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := testoracle.CommandContext(t, ctx, bin, "-c", script)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native abc failed on %q: %v\n%s", script, err, out)
	}
	return ansiEscape.ReplaceAllString(string(out), "")
}

// readLiveFile returns a written file with its time stamp masked ("" when
// the command stopped the script before the write).
func readLiveFile(t *testing.T, path string) string {
	data, err := testoracle.ReadFile(path)
	if err != nil {
		return ""
	}
	return liveAigerDate.ReplaceAllString(string(data), "written")
}

// liveCompare runs script, then print_stats, write_aiger and write_blif,
// in native ABC and in the port, and compares everything.
func liveCompare(t *testing.T, script, input string) {
	t.Helper()
	dir := t.TempDir()
	tail := func(tag string) string {
		return "; print_stats; write_aiger " + filepath.Join(dir, tag+".aig") + "; write_blif " + filepath.Join(dir, tag+".blif")
	}
	native := liveNormalize(runNativeABCInput(t, script+tail("c"), input))
	out, _ := runScriptStatusWith(t, script+tail("g"), func(f *base_main.Frame) {
		f.In = strings.NewReader(input)
		f.BatchMode = true // as abc -c (PDR then prints only closed progress lines)
	})
	if port := liveNormalize(out); native != port {
		t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		return
	}
	for _, ext := range []string{".aig", ".blif"} {
		c, g := readLiveFile(t, filepath.Join(dir, "c"+ext)), readLiveFile(t, filepath.Join(dir, "g"+ext))
		if c != g {
			t.Errorf("%s: %s differs\nnative:\n%q\nport:\n%q", script, ext, c, g)
		}
	}
}

var liveDesigns = []string{
	"read test_vectors/live/l2s_basic.blif; strash",
	"read test_vectors/live/l2s_basic.blif",
	"read test_vectors/live/l2s_multi.blif; strash",
	"read test_vectors/live/l2s_multi.blif",
	"read test_vectors/live/l2s_comb.blif; strash",
	"read test_vectors/seq/s27.blif; strash",
	"read test_vectors/seq/s713.blif; strash",
	"read test_vectors/seq/s713.blif",
	"read test_vectors/native_mul4_gen.aig; strash",
}

func TestLivenessToSafetyLikeNative(t *testing.T) {
	for _, design := range liveDesigns {
		for _, cmd := range []string{
			"l2s", "l2s -1", "l2s -s", "l2s -l", "l2s -1 -s", "l2s -1 -l",
			"l2s -s -1", "l2s -l -1", "l2s -1 -1", "l2s -s -l", "l2s -h",
			"l2ssim", "l2ssim -1",
		} {
			liveCompare(t, design+"; "+cmd, "")
		}
	}
	// l2s on its own result (its outputs are renamed assert_safety_*),
	// and l2ssim reading the option state of the previous command
	liveCompare(t, "read test_vectors/live/l2s_multi.blif; strash; l2s; l2s -l", "")
	liveCompare(t, "read test_vectors/live/l2s_basic.blif; strash; l2s -l; l2ssim -1", "")
	// a one-step l2s on a combinational network returns its AIG and names
	// the PIs from the previous transformation (globals in liveness.c)
	liveCompare(t, "read test_vectors/live/l2s_basic.blif; strash; l2s -l; read test_vectors/live/l2s_comb.blif; strash; l2s -1", "")
}

// l3s without -1 or -l needs LTL formulas, which no reader stores (C's
// BLIF-MV reader crashes on .ltlformula); C then aborts the session
// (exit(0)), and crashed when the network did not come from "read".
func TestLivenessToSafetyLTLLikeNative(t *testing.T) {
	for _, design := range liveDesigns {
		for _, cmd := range []string{"l3s -l", "l3s -1", "l3s -1 -s", "l3s -1 -l", "l3s", "l3s -s", "l3s -h"} {
			liveCompare(t, design+"; "+cmd, "")
		}
	}
	liveCompare(t, "read test_vectors/live/l2s_basic.blif; strash; l3s -f", "")
	liveCompare(t, "read_aiger test_vectors/native_mul4_gen.aig; l3s", "")
	liveCompare(t, "read_aiger test_vectors/native_mul4_gen.aig; l3s -s", "")
}

func TestKLivenessLikeNative(t *testing.T) {
	designs := []string{
		"read test_vectors/live/kcs_count.blif; strash",
		"read test_vectors/live/kcs_count.blif",
		"read test_vectors/live/kcs_mono.blif; strash",
		"read test_vectors/live/kcs_mono.blif",
		"read test_vectors/live/kcs_nomono.blif; strash",
		"read test_vectors/live/kcs_plain.blif; strash",
		"read test_vectors/seq/s27.blif; strash",
		"read test_vectors/native_mul4_gen.aig; strash",
	}
	for _, design := range designs {
		for _, cmd := range []string{"kcs", "kcs -c", "kcs -m", "kcs -C", "kcs -g", "kcs -h"} {
			liveCompare(t, design+"; "+cmd, "")
		}
	}
}

func TestNChooseKLikeNative(t *testing.T) {
	for _, design := range []string{
		"read test_vectors/live/kcs_plain.blif; strash",
		"read test_vectors/live/l2s_multi.blif",
		"read test_vectors/native_mul4_gen.aig; strash",
	} {
		for _, input := range []string{"0", "1", "2", "3", "-1", "x", ""} {
			liveCompare(t, design+"; nck", input)
		}
	}
	liveCompare(t, "read test_vectors/live/kcs_plain.blif; strash; nck -h", "")
}

func TestLivenessNoNetworkLikeNative(t *testing.T) {
	for _, cmd := range []string{"l2s", "l2ssim", "l3s", "kcs", "nck"} {
		native := liveNormalize(runNativeABC(t, cmd))
		out, _ := runScriptStatus(t, cmd)
		if port := liveNormalize(out); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", cmd, native, port)
		}
	}
}
