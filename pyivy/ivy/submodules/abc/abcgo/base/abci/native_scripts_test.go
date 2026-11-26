package abci

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// End-to-end differential test: each script runs in-process on each
// benchmark and its final print_stats line must equal native ABC's, recorded
// in test_vectors/native_script_stats.txt.
//
// With ABC_BINARY set to a native abc built from ../src, the test also runs
// the native binary live; with ABC_UPDATE_GOLDENS=1 as well, it rewrites the
// golden file from the native results.

var nativeScriptBenches = []string{
	"../../../i10.aig",
	"test_vectors/bench/add32.aig",
	"test_vectors/bench/mul8.aig",
	"test_vectors/bench/mul12.aig",
	"test_vectors/bench/sort16.aig",
}

var nativeScripts = []string{
	"strash",
	"balance",
	"balance; rewrite; refactor; resub; rewrite -z",
	"rewrite -z; refactor -z",
	"drw",
	"drf",
	"drf -z -l",
	"dc2",
	"&get; &dc2; &put",
	"&get; &b; &put",
	"&get; &st; &put",
	"fraig",
	"ifraig",
	"renode",
	"renode; sop; fx",
	"multi",
	"if -K 6",
	"if -K 4 -a",
	"&get; &if -K 6; &put",
	"dch",
	"&get; &dch; &put",
	"&get; &st; &synch2; &put",
	"&get; &syn2; &put",
	"&get; &syn3; &put",
	"&get; &syn4; &put",
	"if -K 6; mfs",
	"if -K 6; mfs -r",
	"if -K 6; lutpack",
	"if -K 4; lutpack; lutpack",
	"&get; &if -K 6; &mfs; &put",
	"&get; &sweep; &put",
	"&get; &fraig; &put",
	"iresyn",
	"strash; supercl",
	"&get; &if -K 6 -n; &put",
}

const nativeStatsFile = "test_vectors/native_script_stats.txt"

// statsLine returns the last print_stats line without the network name.
func statsLine(out string) string {
	var last string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "i/o =") {
			last = line
		}
	}
	if _, rest, ok := strings.Cut(last, ":"); ok {
		return strings.TrimSpace(rest)
	}
	return ""
}

func nativeScriptKey(bench, script string) string {
	return filepath.Base(bench) + "|" + script
}

func readNativeStats(t *testing.T, file string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	stats := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok && !strings.HasPrefix(line, "#") {
			stats[k] = v
		}
	}
	return stats
}

func TestScriptsMatchNativeABC(t *testing.T) {
	checkScriptsMatchNative(t, nativeScriptBenches, nativeScripts, nativeStatsFile)
}

// Sequential (ISCAS'89) designs, read from BLIF.
var nativeSeqBenches = []string{
	"test_vectors/seq/s27.blif",
	"test_vectors/seq/s713.blif",
	"test_vectors/seq/s1494.blif",
}

var nativeSeqScripts = []string{
	"strash",
	"strash; scorr",
	"strash; lcorr",
	"strash; ssweep",
	"strash; scleanup",
	"strash; dretime",
	"retime",
	"strash; retime -M 4",
	"strash; zero",
	"strash; fraig -r",
	"strash; dc2; scorr; dretime",
	"strash; &get; &scorr; &put",
	"strash; &get; &scorr -c; &put",
	"strash; &get; &dc2; &put",
	"strash; if -K 4; mfs",
	"strash; &get; &sweep; &put",
	"strash; &get; &lcorr; &put",
	"strash; &get; &fraig; &put",
}

const nativeSeqStatsFile = "test_vectors/native_seq_stats.txt"

func TestSequentialScriptsMatchNativeABC(t *testing.T) {
	checkScriptsMatchNative(t, nativeSeqBenches, nativeSeqScripts, nativeSeqStatsFile)
}

func checkScriptsMatchNative(t *testing.T, benches, scripts []string, statsFile string) {
	if testing.Short() {
		t.Skip("differential script sweep")
	}
	golden := readNativeStats(t, statsFile)
	live := os.Getenv("ABC_BINARY") != ""
	updated := map[string]string{}
	for _, bench := range benches {
		for _, script := range scripts {
			key := nativeScriptKey(bench, script)
			full := fmt.Sprintf("read %s; %s; print_stats", bench, script)
			want, ok := golden[key]
			if live {
				native := statsLine(runNativeABC(t, full))
				if ok && native != want {
					t.Errorf("%s: native ABC now gives %q, golden %q", key, native, want)
				}
				want, ok = native, true
				updated[key] = native
			}
			if !ok {
				t.Errorf("%s: no golden result", key)
				continue
			}
			if got := statsLine(runScript(t, full)); got != want {
				t.Errorf("%s:\n go:     %s\n native: %s", key, got, want)
			}
		}
	}
	if live && os.Getenv("ABC_UPDATE_GOLDENS") == "1" {
		keys := make([]string, 0, len(updated))
		for k := range updated {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		var b strings.Builder
		b.WriteString("# bench|script=final print_stats line from native ABC (see native_scripts_test.go)\n")
		for _, k := range keys {
			fmt.Fprintf(&b, "%s=%s\n", k, updated[k])
		}
		if err := os.WriteFile(statsFile, []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
