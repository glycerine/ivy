package abci

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// End-to-end check of verbose output: each command runs on i10 with its
// statistics printing turned on, and the output, with run times and blank
// lines removed, must equal native ABC's (test_vectors/verbose/*.txt).
// ABC_BINARY runs the native binary live; with ABC_UPDATE_GOLDENS=1 the
// goldens are rewritten from it.
var nativeVerboseScripts = []string{
	"balance -v",
	"rewrite -v",
	"refactor -v",
	"resub -v",
	"dc2 -v",
	"&get; &dc2 -v",
	"&get; &if -K 6 -v",
	"if -K 6 -v",
	"logic; sweep -v",
	"renode; sop; fx -v",
	"renode -v",
	"&get; &b -v",
	"&get; &ps; &st; &ps; &synch2; &ps; &if -K 6; &ps",
}

var verboseTimes = regexp.MustCompile(`[0-9.]* sec`)

func normalizeVerbose(out string) string {
	var kept []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(verboseTimes.ReplaceAllString(line, " sec"), " ")
		if line == "" || strings.Contains(line, "ABC command line") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n") + "\n"
}

func verboseGoldenName(script string) string {
	name := strings.NewReplacer("&", "g_", " ", "_", ";", "", "-", "").Replace(script)
	return filepath.Join("test_vectors", "verbose", name+".txt")
}

func TestVerboseOutputMatchesNative(t *testing.T) {
	if testing.Short() {
		t.Skip("verbose output sweep")
	}
	live := os.Getenv("ABC_BINARY") != ""
	for _, script := range nativeVerboseScripts {
		full := fmt.Sprintf("read ../../../i10.aig; strash; %s", script)
		file := verboseGoldenName(script)
		if live {
			native := normalizeVerbose(runNativeABC(t, full))
			if os.Getenv("ABC_UPDATE_GOLDENS") == "1" {
				os.MkdirAll(filepath.Dir(file), 0o755)
				if err := os.WriteFile(file, []byte(native), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
		want, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("%s: %v", script, err)
			continue
		}
		if got := normalizeVerbose(runScript(t, full)); got != string(want) {
			t.Errorf("%s:\n go:\n%s\n native:\n%s", script, got, want)
		}
	}
}
