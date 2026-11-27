package abci

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var abcStamp = regexp.MustCompile(`This file was written by ABC on [^\n]*\n`)

// TestWriteAigerLikeNative: write_aiger ends files as Io_WriteAiger does
// ("c", the network name, ABC's stamp and the AIGER pointer; just "c" with
// -u). With ABC_BINARY the files must equal native's, the date aside.
func TestWriteAigerLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, cmd := range []string{"write_aiger", "write_aiger -s", "write_aiger -u", "write_aiger -c"} {
		goFile := filepath.Join(dir, "go.aig")
		runScript(t, "read ../../../i10.aig; strash; "+cmd+" "+goFile)
		got, err := os.ReadFile(goFile)
		if err != nil {
			t.Fatal(err)
		}
		if cmd == "write_aiger -u" {
			if abcStamp.Match(got) || !bytes.HasSuffix(got, []byte("c")) {
				t.Fatalf("%s: unique AIG must end with a bare \"c\"", cmd)
			}
		} else if !abcStamp.Match(got) || !bytes.HasSuffix(got, []byte("refer to http://fmv.jku.at/aiger\n")) {
			t.Fatalf("%s: written AIG lacks ABC's trailer", cmd)
		}
		if os.Getenv("ABC_BINARY") == "" {
			continue
		}
		nativeFile := filepath.Join(dir, "native.aig")
		runNativeABC(t, "read ../../../i10.aig; strash; "+cmd+" "+nativeFile)
		want, err := os.ReadFile(nativeFile)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(abcStamp.ReplaceAll(got, nil), abcStamp.ReplaceAll(want, nil)) {
			t.Errorf("%q: written AIG differs from native (%d vs %d bytes)", cmd, len(got), len(want))
		}
	}
}
