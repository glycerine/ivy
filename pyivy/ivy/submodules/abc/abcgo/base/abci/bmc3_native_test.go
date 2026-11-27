package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	bmc3Seconds = regexp.MustCompile(`[0-9]+\.[0-9]+ sec`)
	bmc3Memory  = regexp.MustCompile(`\s*[0-9.]+ MB`)
	bmc3Percent = regexp.MustCompile(`\(\s*[0-9.]+ %\)`)
)

// bmc3Exact masks only run times, memory (sat_solver_memory is an estimate
// of Go's own layout) and percentages of run time.
func bmc3Exact(out string) string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "ABC command line") || strings.TrimSpace(line) == "" {
			continue
		}
		line = bmc3Seconds.ReplaceAllString(line, "N sec")
		line = bmc3Memory.ReplaceAllString(line, " N MB")
		line = bmc3Percent.ReplaceAllString(line, "(P %)")
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return strings.Join(lines, "\n")
}

// bmc3 is a line-by-line port of bmcBmc3.c: the verbose statistics, the SAT
// models, the solve-all order and the frame counts equal C's.
func TestBmc3LikeNative(t *testing.T) {
	var scripts []string
	for _, file := range []string{"seq/s713.blif", "seq/s1494.blif", "native_s27_drw.blif"} {
		for _, opts := range []string{"", "-v", "-F 10 -v", "-a -F 10", "-a -v -F 10", "-S 3 -v", "-C 100 -v", "-J 2 -v -F 10", "-a -x -v -F 8", "-u -v -F 8", "-a -H 1000 -v -F 8", "-r -P 100 -Q 50 -R 50 -v -F 12", "-s -v -F 8"} {
			scripts = append(scripts, "read test_vectors/"+file+"; strash; bmc3 "+opts+"; print_status")
		}
	}
	// unfalsified miters: multi-frame SAT work, conflicts and learned clauses
	scripts = append(scripts,
		"read test_vectors/seq/s713.blif; strash; dc2; miter test_vectors/seq/s713.blif; bmc3 -F 12 -v; print_status",
		"read test_vectors/seq/s713.blif; strash; dc2; miter test_vectors/seq/s713.blif; bmc3 -J 2 -C 5 -v -F 20; print_status",
		"read test_vectors/seq/s1494.blif; strash; dc2; miter test_vectors/seq/s1494.blif; bmc3 -F 20 -v; print_status",
		"read test_vectors/seq/s1494.blif; strash; dc2; miter test_vectors/seq/s1494.blif; bmc3 -a -v -F 20; print_status",
		"read test_vectors/seq/s1494.blif; strash; dc2; miter test_vectors/seq/s1494.blif; bmc3 -C 10 -v -F 20; print_status",
	)
	for _, script := range scripts {
		native := bmc3Exact(runNativeABC(t, script))
		port, _ := runScriptStatus(t, script)
		if port = bmc3Exact(port); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}

// write_cex writes the same files; after "bmc3 -a" without -x the CEX
// vector holds the "solved, not stored" sentinel, which native write_cex
// dereferenced (CBUGS_FIXED.md #36).
func TestBmc3WriteCexLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, file := range []string{"seq/s713.blif", "seq/s1494.blif", "native_s27_drw.blif"} {
		for _, opts := range []string{"-F 10", "-S 3", "-a -x -F 10", "-a -F 6"} {
			script := "read test_vectors/" + file + "; strash; bmc3 " + opts + "; write_cex "
			nativeFile, portFile := filepath.Join(dir, "native.cex"), filepath.Join(dir, "port.cex")
			runNativeABC(t, script+nativeFile)
			runScriptStatus(t, script+portFile)
			native, err1 := os.ReadFile(nativeFile)
			port, err2 := os.ReadFile(portFile)
			if err1 != nil || err2 != nil || string(native) != string(port) {
				t.Errorf("%s: files differ (%v, %v)\nnative:\n%s\nport:\n%s", script, err1, err2, native, port)
			}
		}
	}
}

// With -a, an output that ternary simulation proves constant 1 is SAT
// without calling the solver, and C then read the model of that missing
// solve: glucose (-g) dereferenced its empty model, sat_solver read
// uninitialized or stale values (CBUGS_FIXED.md #37).
func TestBmc3SolveAllConstantOutputLikeNative(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/seq/s713.blif; strash; bmc3 -a -g -F 8; print_status",
		"read test_vectors/seq/s713.blif; strash; bmc3 -a -x -g -F 8; print_status",
		"read test_vectors/seq/s1494.blif; strash; bmc3 -a -g -F 8; print_status",
		"read test_vectors/seq/s713.blif; strash; bmc3 -a -s -v -F 8; print_status",
	} {
		native := bmc3Exact(runNativeABC(t, script))
		port, _ := runScriptStatus(t, script)
		if port = bmc3Exact(port); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}

var bmc3LogTimes = regexp.MustCompile(`Time\(ms\) +-?[0-9]+ +-?[0-9]+`)

// "bmc3 -W file" logs each output's solve; the second column reads the
// per-output budget p->pTime4Outs, which exists only with -a -H, and native
// dereferenced it as NULL otherwise (CBUGS_FIXED.md #38).
func TestBmc3LogFileLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, opts := range []string{"-F 3", "-a -F 3", "-a -H 500 -F 3"} {
		script := "read test_vectors/seq/s1494.blif; strash; dc2; miter test_vectors/seq/s1494.blif; bmc3 " + opts + " -W "
		nativeFile, portFile := filepath.Join(dir, "native.log"), filepath.Join(dir, "port.log")
		runNativeABC(t, script+nativeFile)
		runScriptStatus(t, script+portFile)
		native, err1 := os.ReadFile(nativeFile)
		port, err2 := os.ReadFile(portFile)
		n := bmc3LogTimes.ReplaceAllString(string(native), "Time(ms) T T")
		g := bmc3LogTimes.ReplaceAllString(string(port), "Time(ms) T T")
		if err1 != nil || err2 != nil || n != g || n == "" {
			t.Errorf("%s: logs differ (%v, %v)\nnative:\n%s\nport:\n%s", script, err1, err2, native, port)
		}
	}
}
