package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// show built its .dot/.ps file names and the dot command line with sprintf
// into char[200]/[1000] (abcShow.c, aigShow.c, giaShow.c, ivyShow.c): a design
// path over 200 bytes aborted native ABC (CBUGS_FIXED.md #54). The name (the
// path with '/' turned into '_') is now bounded; a file name this long cannot
// be created, which both report.
func TestShowLongPathLikeNative(t *testing.T) {
	dir := filepath.Join(t.TempDir(), strings.Repeat("p", 120), strings.Repeat("q", 120))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("test_vectors/native_mul4_gen.aig")
	if err != nil {
		t.Fatal(err)
	}
	design := filepath.Join(dir, "mul4.aig")
	if err := os.WriteFile(design, data, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, prep := range []string{"strash", "strash; &get"} {
		script := "read " + design + "; " + prep + "; show"
		native := printSymmNormalize(runNativeABC(t, script))
		out, _ := runScriptStatus(t, script)
		if port := printSymmNormalize(out); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
