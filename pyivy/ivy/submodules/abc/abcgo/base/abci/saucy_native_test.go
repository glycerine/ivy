package abci

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestSaucySymmetriesLikeNative compares saucy3's output (timing lines
// aside) with native ABC's on two small symmetric designs. The port did not
// close the list of nonsingleton cells with the sentinel n, as
// abcSaucy.c's saucy_search does, so walking the list cycled forever.
func TestSaucySymmetriesLikeNative(t *testing.T) {
	for _, name := range []string{"sym1", "sym2"} {
		want, err := os.ReadFile("test_vectors/" + name + "_saucy_native.txt")
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan string, 1)
		go func() {
			out, _ := runScriptStatus(t, "read test_vectors/"+name+".blif; strash; saucy3")
			done <- out
		}()
		var out string
		select {
		case out = <-done:
		case <-time.After(20 * time.Second):
			t.Fatalf("saucy3 on %s did not finish (native: 0.01 s)", name)
		}
		var kept []string
		for _, line := range strings.SplitAfter(out, "\n") {
			if line != "" && line != "\n" && !strings.HasSuffix(line, " sec\n") {
				kept = append(kept, line)
			}
		}
		if got := strings.Join(kept, ""); got != string(want) {
			t.Errorf("saucy3 on %s:\n got:\n%s\nwant:\n%s", name, got, want)
		}
	}
}
