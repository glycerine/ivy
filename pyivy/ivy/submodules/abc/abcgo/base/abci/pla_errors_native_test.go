package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Io_ReadPlaNetwork reports malformed cubes as "<file> (line N): ...".
var plaCases = map[string]string{
	"incube":  ".i 2\n.o 1\n11 1\n1 1\n.e\n",
	"outcube": ".i 2\n.o 1\n11 1\n10 11\n.e\n",
	"onetok":  ".i 2\n.o 1\n11 1\n10\n.e\n",
	"threetk": ".i 2\n.o 1\n11 1\n10 1 1\n.e\n",
}

func TestPlaReaderLikeNative(t *testing.T) {
	for name, source := range plaCases {
		path := filepath.Join(t.TempDir(), name+".pla")
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		script := "read " + path + "; print_stats"
		port, _ := runScriptStatus(t, script)
		if !strings.Contains(port, path+" (line ") || !strings.Contains(port, "Reading network from file has failed.") {
			t.Errorf("%s: port:\n%s", name, port)
		}
		if os.Getenv("ABC_BINARY") == "" {
			continue
		}
		if native, port := bmc3Normalize(runNativeABC(t, script)), bmc3Normalize(port); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", name, native, port)
		}
	}
}
