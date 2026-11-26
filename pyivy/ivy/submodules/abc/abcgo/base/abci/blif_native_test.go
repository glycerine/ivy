package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// blifBody drops the comment header, which carries a timestamp.
func blifBody(s string) string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(s, "\n") {
		if !strings.HasPrefix(line, "#") {
			b.WriteString(line)
		}
	}
	return b.String()
}

// Native: abc -c "read mul8.aig; if -K 6; write_blif native_mul8_if6.blif".
// Writing a mapped network converts its local AIGs to BDDs; ABC only lets
// CUDD reorder (and so permute fanins) once the manager grows large, so
// LUT-sized functions keep the mapper's fanin order.
func TestMappedBlifMatchesNative(t *testing.T) {
	out := filepath.Join(t.TempDir(), "mul8_if6.blif")
	runScript(t, "read test_vectors/bench/mul8.aig; if -K 6; write_blif "+out)
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("test_vectors/native_mul8_if6.blif")
	if err != nil {
		t.Fatal(err)
	}
	if g, w := blifBody(string(got)), blifBody(string(want)); g != w {
		gl, wl := strings.Split(g, "\n"), strings.Split(w, "\n")
		for i := range wl {
			if i >= len(gl) || gl[i] != wl[i] {
				t.Fatalf("line %d differs:\n got: %q\nwant: %q", i+1, gl[min(i, len(gl)-1)], wl[i])
			}
		}
		t.Fatal("extra output lines")
	}
}
