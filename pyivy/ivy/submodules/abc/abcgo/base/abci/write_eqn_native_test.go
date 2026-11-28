package abci

import (
	"bytes"
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"regexp"
	"testing"
)

var eqnStamp = regexp.MustCompile(`written by ABC on [^\n]*\n`)

// TestWriteEqnLikeNative: write_eqn follows Io_WriteEqn: names EQN cannot
// carry (here "V32(0)") make it print ABC's message and write nothing;
// otherwise the header, wrapped orders and Hop_ObjPrintEqn formulas match
// native's file, the date aside.
func TestWriteEqnLikeNative(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.eqn")
	runScript(t, "read ../../test_vectors/i10.aig; strash; write_eqn "+file)
	if _, err := testoracle.Stat(file); err == nil {
		t.Fatalf("write_eqn wrote a file for names EQN cannot carry")
	}
	for _, script := range []string{"strash; short_names", "strash; short_names; if -K 4", "strash; short_names; logic"} {
		goFile := filepath.Join(dir, "go.eqn")
		runScript(t, "read ../../test_vectors/i10.aig; "+script+"; write_eqn "+goFile)
		got, err := testoracle.ReadFile(goFile)
		if err != nil {
			t.Fatalf("%s: %v", script, err)
		}
		if !eqnStamp.Match(got) {
			t.Fatalf("%s: no ABC header", script)
		}
		if testoracle.NativeABCPath() == "" {
			continue
		}
		nativeFile := filepath.Join(dir, "native.eqn")
		runNativeABC(t, "read ../../test_vectors/i10.aig; "+script+"; write_eqn "+nativeFile)
		want, err := testoracle.ReadFile(nativeFile)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(eqnStamp.ReplaceAll(got, nil), eqnStamp.ReplaceAll(want, nil)) {
			t.Errorf("%s: EQN differs from native (%d vs %d bytes)", script, len(got), len(want))
		}
	}
}
