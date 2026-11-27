package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestWriteVerilogDeclaresEveryInternalNet: Io_WriteVerilogWires skips the
// object with id 0, not the first node; the port skipped the first node, so
// its net was assigned without being declared.
func TestWriteVerilogDeclaresEveryInternalNet(t *testing.T) {
	file := filepath.Join(t.TempDir(), "x.v")
	runScript(t, "read ../../../i10.aig; strash; write_verilog "+file)
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, "  wire ")
	end := strings.Index(text[start:], ";")
	declared := map[string]bool{}
	for _, name := range regexp.MustCompile(`new_n\d+_`).FindAllString(text[start:start+end], -1) {
		declared[name] = true
	}
	for _, m := range regexp.MustCompile(`assign (new_n\d+_) =`).FindAllStringSubmatch(text, -1) {
		if !declared[m[1]] {
			t.Fatalf("%s is assigned but not declared", m[1])
		}
	}
}

// TestWriteVerilogLikeNative: with ABC_BINARY, write_verilog files equal
// native's (Hop_ObjPrintVerilog formulas, wire lists), the date aside.
func TestWriteVerilogLikeNative(t *testing.T) {
	if os.Getenv("ABC_BINARY") == "" {
		t.Skip("ABC_BINARY not set")
	}
	stamp := regexp.MustCompile(`written by ABC on [^\n]*\n`)
	dir := t.TempDir()
	for _, script := range []string{"strash", "strash; if -K 4", "strash; logic"} {
		goFile, nativeFile := filepath.Join(dir, "go.v"), filepath.Join(dir, "native.v")
		runScript(t, "read ../../../i10.aig; "+script+"; write_verilog "+goFile)
		runNativeABC(t, "read ../../../i10.aig; "+script+"; write_verilog "+nativeFile)
		got, err1 := os.ReadFile(goFile)
		want, err2 := os.ReadFile(nativeFile)
		if err1 != nil || err2 != nil {
			t.Fatal(err1, err2)
		}
		if stamp.ReplaceAllString(string(got), "") != stamp.ReplaceAllString(string(want), "") {
			t.Errorf("%s: Verilog differs from native (%d vs %d bytes)", script, len(got), len(want))
		}
	}
}
