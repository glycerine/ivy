package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Io_ReadBlifMv appends a spare ".end", gives undriven nets constant-0
// drivers (Abc_NtkFinalizeRead's warning) and reports malformed files as
// "Line N: ..." followed by "Reading network from file has failed.". The
// port rejected the first two kinds of file and worded its errors itself.
var blifCases = map[string]string{
	"dup":      ".model t\n.inputs a b\n.outputs y\n.names a b y\n11 1\n.names a y\n1 1\n.end\n",
	"undef":    ".model t\n.inputs a b\n.outputs y\n.names a c y\n11 1\n.end\n",
	"noend":    ".model t\n.inputs a b\n.outputs y\n.names a b y\n11 1\n",
	"nodrv":    ".model t\n.inputs a b\n.outputs y z\n.names a b y\n11 1\n.end\n",
	"cube":     ".model t\n.inputs a b\n.outputs y\n.names a b y\n111 1\n.end\n",
	"outval":   ".model t\n.inputs a b\n.outputs y\n.names a b y\n11 2\n.end\n",
	"polarity": ".model t\n.inputs a b\n.outputs y\n.names a b y\n11 1\n00 0\n.end\n",
}

func writeBlifCase(t *testing.T, name string) string {
	path := filepath.Join(t.TempDir(), name+".blif")
	if err := os.WriteFile(path, []byte(blifCases[name]), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBlifReaderLenientLikeABC(t *testing.T) {
	for name, want := range map[string][]string{
		"noend": {"i/o =    2/    1  lat =    0  nd =     1"},
		"undef": {"Warning: Constant-0 drivers added to 1 non-driven nets in network \"t\":\nc\n", "i/o =    2/    1"},
		"nodrv": {"Warning: Constant-0 drivers added to 1 non-driven nets in network \"t\":\nz\n", "i/o =    2/    2"},
		"dup":   {"Line 6: Signal \"y\" is defined more than once.\nReading network from file has failed."},
		"cube":  {"Line 5: Cube \"111\" has size different from the fanin count (2).\nReading network from file has failed."},
	} {
		out, _ := runScriptStatus(t, "read "+writeBlifCase(t, name)+"; print_stats")
		for _, w := range want {
			if !strings.Contains(out, w) {
				t.Errorf("%s: missing %q:\n%s", name, w, out)
			}
		}
	}
}

func TestBlifReaderLikeNative(t *testing.T) {
	for name := range blifCases {
		path := writeBlifCase(t, name)
		script := "read " + path + "; print_stats"
		clean := func(s string) string { return strings.ReplaceAll(bmc3Normalize(s), path, "FILE") }
		port, _ := runScriptStatus(t, script)
		if native, port := clean(runNativeABC(t, script)), clean(port); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", name, native, port)
		}
	}
}
