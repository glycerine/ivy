package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Io_ReadBench gives undriven signals constant-0 drivers
// (Abc_NtkFinalizeRead); the port rejected the file. Unknown gates and
// doubly driven nets are reported in ABC's words.
var benchCases = map[string]string{
	"undef": "INPUT(a)\nINPUT(b)\nOUTPUT(y)\ny = AND(a, c)\n",
	"dup":   "INPUT(a)\nINPUT(b)\nOUTPUT(y)\ny = AND(a, b)\ny = OR(a, b)\n",
	"gate":  "INPUT(a)\nOUTPUT(y)\ny = FOO(a)\n",
}

func writeBenchCase(t *testing.T, name string) string {
	path := filepath.Join(t.TempDir(), name+".bench")
	if err := os.WriteFile(path, []byte(benchCases[name]), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBenchReaderLikeABC(t *testing.T) {
	for name, want := range map[string]string{
		"undef": "undef\":\nc\n",
		"dup":   "NetworkCheck: Net \"y\" has more than one driver.\nIo_ReadBench: The network check has failed.\nReading network from file has failed.",
		"gate":  "Io_ReadBenchNetwork(): Cannot determine gate type \"FOO\" in line 3.\nReading network from file has failed.",
	} {
		out, _ := runScriptStatus(t, "read "+writeBenchCase(t, name)+"; print_stats")
		if !strings.Contains(out, want) {
			t.Errorf("%s: missing %q:\n%s", name, want, out)
		}
	}
}

func TestBenchReaderLikeNative(t *testing.T) {
	for name := range benchCases {
		path := writeBenchCase(t, name)
		script := "read " + path + "; print_stats"
		port, _ := runScriptStatus(t, script)
		clean := func(s string) string { return strings.ReplaceAll(bmc3Normalize(s), path, "FILE") }
		if native, port := clean(runNativeABC(t, script)), clean(port); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", name, native, port)
		}
	}
}
