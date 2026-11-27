package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// pdr (Pdr_ManSolve) with -v prints its parameters, one progress line per
// frame (Pdr_ManPrintProgress), the invariant report and, from
// Pdr_ManStop, the statistics block. The counterexample bits come from the
// SAT models of the frame solvers, so they are compared too.
var pdrScripts = func() []string {
	var scripts []string
	for _, design := range []string{"test_vectors/seq/s713.blif", "test_vectors/seq/s1494.blif", "test_vectors/native_s27_drw.blif"} {
		for _, command := range []string{"pdr", "pdr -v", "pdr -v -a", "pdr -F 5 -v", "pdr -T 5 -v"} {
			scripts = append(scripts, "read "+design+"; strash; "+command+"; print_status")
		}
	}
	// self-miters against a resynthesized copy are proved
	for _, design := range []string{"s27", "s713", "s1494"} {
		scripts = append(scripts, "read test_vectors/seq/"+design+".blif; strash; dc2; miter test_vectors/seq/"+design+".blif; pdr -v; print_status")
	}
	return scripts
}()

var pdrPercent = regexp.MustCompile(`\( *[0-9.]+ %\)`)

// Progress lines ended by '\r' are overwritten on a terminal; the native
// binary (batch mode under -c) skips them, the in-process frame does not.
var pdrOverwritten = regexp.MustCompile(`[^\n\r]*\r`)

func pdrNormalize(out string) string {
	return pdrPercent.ReplaceAllString(bmc2VerboseNormalize(pdrOverwritten.ReplaceAllString(out, "")), "(P %)")
}

func TestPdrLikeNative(t *testing.T) {
	for _, script := range pdrScripts {
		if native, port := pdrNormalize(runNativeABC(t, script)), pdrNormalize(runScript(t, script)); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}

// write_cex of pdr's counterexample is byte-identical to the native one.
func TestPdrWriteCexLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, design := range []string{"test_vectors/seq/s713.blif", "test_vectors/native_s27_drw.blif"} {
		native, port := filepath.Join(dir, "native.cex"), filepath.Join(dir, "port.cex")
		script := "read " + design + "; strash; pdr; write_cex -n "
		runNativeABC(t, script+native)
		runScript(t, script+port)
		want, err := os.ReadFile(native)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(port)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("%s:\nport:\n%s\nnative:\n%s", design, got, want)
		}
	}
}
