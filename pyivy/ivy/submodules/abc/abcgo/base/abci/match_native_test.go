package abci

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// match (Abc_CommandMatch -> Abc_NtkDarMatch -> Saig_StrSimPerformMatching
// in saigStrSim.c) and simsec (Ssw_SecWithSimilarity in sswIslands.c) match
// the objects of two AIGs by random simulation with Aig_ManRandom; the port
// used math/rand, printed other match counts and none of the statistics.

var matchSeconds = regexp.MustCompile(`[0-9]+\.[0-9]+ sec( \( *[0-9.]+ %\))?`)
var matchMegabytes = regexp.MustCompile(`[0-9]+\.[0-9]+ MB`)

func matchNormalize(out string) string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "ABC command line") || strings.TrimSpace(line) == "" {
			continue
		}
		line = matchSeconds.ReplaceAllString(line, "T sec")
		line = matchMegabytes.ReplaceAllString(line, "M MB")
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

var matchDesigns = []string{"test_vectors/seq/s713.blif", "test_vectors/seq/s1494.blif", "test_vectors/native_s27_drw.blif"}

// matchScripts returns the scripts for one design; %T is replaced by a
// scratch directory and %O by the result file.
func matchScripts(design string) []string {
	var scripts []string
	for _, opts := range []string{"", "-v", "-D 1 -v", "-D 2 -v"} {
		scripts = append(scripts,
			"read "+design+"; strash; dc2; write_aiger %T/m2.aig; read "+design+"; strash; match "+opts+" %T/m2.aig; print_stats; write_aiger %O",
			"read "+design+"; strash; dc2; dretime; write_aiger %T/m3.aig; read "+design+"; strash; match "+opts+" %T/m3.aig; print_stats; write_aiger %O",
			"read "+design+"; strash; match "+opts+" "+design+"; print_stats; write_aiger %O",
			"read "+design+"; strash; dc2; match "+opts+"; print_stats; write_aiger %O",
		)
	}
	scripts = append(scripts,
		"read "+design+"; strash; dc2; simsec -m "+design,
		"read "+design+"; strash; dc2; simsec -m -D 1 "+design,
		"read "+design+"; strash; dc2; simsec -m -D 2 -v "+design,
		"read "+design+"; strash; dc2; dretime; simsec -m -D 1 "+design,
		"read "+design+"; strash; dc2; miter "+design+"; simsec",
		"read "+design+"; strash; dc2; miter -m "+design+"; simsec",
		"read "+design+"; strash; dc2; miter -m "+design+"; simsec -D 1",
		"read "+design+"; strash; dc2; miter -m "+design+"; match -m; print_stats; write_aiger %O",
		"read "+design+"; strash; dc2; miter -m "+design+"; match -m -D 1; print_stats; write_aiger %O",
		// a single-output miter cannot be demitered; ABC crashed (CBUGS_FIXED.md #44)
		"read "+design+"; strash; dc2; miter "+design+"; match -m; print_stats",
	)
	return scripts
}

func TestMatchLikeNative(t *testing.T) {
	if testoracle.NativeABCPath() == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	dir := t.TempDir()
	for _, design := range matchDesigns {
		for _, script := range matchScripts(design) {
			s := strings.ReplaceAll(script, "%T", dir)
			cOut, goOut := filepath.Join(dir, "c.aig"), filepath.Join(dir, "go.aig")
			os.Remove(cOut)
			os.Remove(goOut)
			native := matchNormalize(runNativeABC(t, strings.ReplaceAll(s, "%O", cOut)))
			port, _ := runScriptStatus(t, strings.ReplaceAll(s, "%O", goOut))
			port = matchNormalize(port)
			if native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
				continue
			}
			if strings.Contains(script, "%O") {
				a, errA := testoracle.ReadFile(cOut)
				b, errB := testoracle.ReadFile(goOut)
				// the comment section's time stamp aside
				if errA != nil || errB != nil || aigerStamp.ReplaceAllString(string(a), "") != aigerStamp.ReplaceAllString(string(b), "") {
					t.Errorf("%s: result AIGER differs (%v, %v)", script, errA, errB)
				}
			}
		}
	}
}
