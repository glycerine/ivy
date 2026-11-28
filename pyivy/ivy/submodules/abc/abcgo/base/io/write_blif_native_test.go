package io

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"strings"
	"testing"
)

// Io_WriteBlif wraps long .inputs/.outputs/.names lines with " \" at ABC's
// line length. The file must equal native ABC's output for the same design,
// apart from the timestamp comment and the path-derived model name.
// test_vectors/native_mul8.blif: abc -c "read mul8.aig; write_blif ...".
func TestWriteBlifMatchesNative(t *testing.T) {
	n, err := ReadAiger("../abci/test_vectors/bench/mul8.aig", true)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "mul8.blif")
	if err := Io_WriteBlifLogic(n, path, 1); err != nil {
		t.Fatal(err)
	}
	strip := func(file string) string {
		data, err := testoracle.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var keep []string
		for _, line := range strings.Split(string(data), "\n") {
			if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, ".model") {
				keep = append(keep, line)
			}
		}
		return strings.Join(keep, "\n")
	}
	if got, want := strip(path), strip("test_vectors/native_mul8.blif"); got != want {
		gl, wl := strings.Split(got, "\n"), strings.Split(want, "\n")
		for i := range min(len(gl), len(wl)) {
			if gl[i] != wl[i] {
				t.Fatalf("line %d:\n go:     %q\n native: %q", i, gl[i], wl[i])
			}
		}
		t.Fatalf("length %d lines, native %d", len(gl), len(wl))
	}
}
