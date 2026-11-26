package abci

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var ifVerboseTimes = regexp.MustCompile(`(T|Total time) = *[0-9.]* sec`)

// TestIfVerboseLikeNative: "if -v" prints ABC's statistics lines (memory
// from ABC's own record sizes, then one line per pass), times aside.
func TestIfVerboseLikeNative(t *testing.T) {
	want, err := os.ReadFile("test_vectors/mul8_if_v_native.txt")
	if err != nil {
		t.Fatal(err)
	}
	out := runScript(t, "read test_vectors/bench/mul8.aig; strash; if -K 6 -v")
	var kept []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimRight(line, " "); line != "" {
			kept = append(kept, ifVerboseTimes.ReplaceAllString(line, "$1 = sec"))
		}
	}
	got := strings.Join(kept, "\n") + "\n"
	var wantLines []string
	for _, line := range strings.Split(strings.TrimRight(string(want), "\n"), "\n") {
		wantLines = append(wantLines, strings.TrimRight(line, " "))
	}
	if w := strings.Join(wantLines, "\n") + "\n"; got != w {
		t.Fatalf("if -v:\n got:\n%s\nwant:\n%s", got, w)
	}
}
