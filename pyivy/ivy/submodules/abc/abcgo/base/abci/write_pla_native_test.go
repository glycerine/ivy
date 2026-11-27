package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// write_pla is Io_Write's PLA path: collapsed networks only, converted to a
// netlist of positive SOPs and written cube by cube (Io_WritePlaOne). The
// port enumerated minterms instead (a 12x larger file for mul8).
var writePlaScripts = []string{
	"read .scratch_mul8; collapse; write_pla %s",
	"read test_vectors/native_s27_drw.blif; collapse; write_pla %s",
	"read test_vectors/native_s27_drw.blif; strash; write_pla %s",
}

func writePlaRun(t *testing.T, run func(string) string, script string) (string, string) {
	path := filepath.Join(t.TempDir(), "out.pla")
	script = strings.ReplaceAll(strings.ReplaceAll(script, ".scratch_mul8", "../../.scratch/bench/mul8.blif"), "%s", path)
	out := run(script)
	data, _ := os.ReadFile(path)
	text := string(data)
	if i := strings.Index(text, "\n"); strings.HasPrefix(text, "# Benchmark") && i >= 0 {
		text = text[i+1:]
	}
	return text, bmc3Normalize(strings.ReplaceAll(out, path, "FILE"))
}

func TestWritePlaLikeNative(t *testing.T) {
	if _, err := os.Stat("../../.scratch/bench/mul8.blif"); err != nil {
		t.Skip("mul8 benchmark not present")
	}
	for _, script := range writePlaScripts {
		port, portOut := writePlaRun(t, func(s string) string { out, _ := runScriptStatus(t, s); return out }, script)
		if strings.Contains(script, "strash") {
			if !strings.Contains(portOut, "PLA writing is available for collapsed networks.") || port != "" {
				t.Errorf("%s: port:\n%s\n%s", script, portOut, port)
			}
		} else if !strings.Contains(port, ".p ") || strings.Count(port, "\n") > 50000 {
			t.Errorf("%s: port file has %d lines", script, strings.Count(port, "\n"))
		}
		if os.Getenv("ABC_BINARY") == "" {
			continue
		}
		native, nativeOut := writePlaRun(t, func(s string) string { return runNativeABC(t, s) }, script)
		if native != port {
			t.Errorf("%s: file differs from native (%d vs %d bytes)", script, len(native), len(port))
		}
		if nativeOut != portOut {
			t.Errorf("%s: output\nnative:\n%s\nport:\n%s", script, nativeOut, portOut)
		}
	}
}
