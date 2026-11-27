package abci

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var giaStamp = regexp.MustCompile(`This file was produced by the GIA package in ABC on [^\n]*\n`)

// TestGiaWriteLikeNative: &w writes ABC's trailer after the extensions
// ("This file was produced by the GIA package in ABC on <date>" and the
// AIGER pointer), whatever comment the input file carried. With ABC_BINARY
// the files must equal native's byte for byte, the date aside.
func TestGiaWriteLikeNative(t *testing.T) {
	dir := t.TempDir()
	for i, script := range []string{"&get", "&get; &lf", "&get; &if -K 4"} {
		goFile := filepath.Join(dir, "go.aig")
		runScript(t, "read ../../../i10.aig; strash; "+script+"; &w "+goFile)
		got, err := os.ReadFile(goFile)
		if err != nil {
			t.Fatal(err)
		}
		if !giaStamp.Match(got) || !bytes.HasSuffix(got, []byte("For information about AIGER format, refer to http://fmv.jku.at/aiger\n")) {
			t.Fatalf("%s: written AIG lacks ABC's trailer", script)
		}
		if os.Getenv("ABC_BINARY") == "" {
			continue
		}
		nativeFile := filepath.Join(dir, "native.aig")
		runNativeABC(t, "read ../../../i10.aig; strash; "+script+"; &w "+nativeFile)
		want, err := os.ReadFile(nativeFile)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(giaStamp.ReplaceAll(got, nil), giaStamp.ReplaceAll(want, nil)) {
			t.Errorf("case %d %q: written AIG differs from native (%d vs %d bytes)", i, script, len(got), len(want))
		}
	}
}
