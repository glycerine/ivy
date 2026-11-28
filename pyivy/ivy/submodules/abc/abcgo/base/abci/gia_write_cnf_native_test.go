package abci

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"testing"
)

// &write_cnf writes with Cnf_DataWriteIntoFile: a comment line before the
// header and an empty line at the end, which the port's writer omitted.
func TestGiaWriteCnfLikeNative(t *testing.T) {
	dir := t.TempDir()
	port, native := filepath.Join(dir, "port.cnf"), filepath.Join(dir, "native.cnf")
	const script = "read ../../test_vectors/i10.aig; &get; &write_cnf "
	runScript(t, script+port)
	data, err := testoracle.ReadFile(port)
	if err != nil {
		t.Fatal(err)
	}
	const head = "c Result of efficient AIG-to-CNF conversion using package CNF\np cnf "
	if len(data) < len(head) || string(data[:len(head)]) != head || string(data[len(data)-3:]) != "0\n\n" {
		t.Fatalf("port file starts %q, ends %q", data[:min(len(data), 80)], data[max(0, len(data)-10):])
	}
	if testoracle.NativeABCPath() == "" {
		return
	}
	runNativeABC(t, script+native)
	want, err := testoracle.ReadFile(native)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != string(data) {
		t.Fatalf("file differs from native (%d vs %d bytes)", len(want), len(data))
	}
}
