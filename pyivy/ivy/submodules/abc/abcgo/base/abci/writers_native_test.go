package abci

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// nativeWriterCase writes the same network with Go and (with ABC_BINARY)
// native ABC and compares the files, with lines matching mask removed.
func nativeWriterCase(t *testing.T, script, ext string, mask *regexp.Regexp, check func([]byte) string) {
	t.Helper()
	dir := t.TempDir()
	goFile, nativeFile := filepath.Join(dir, "go."+ext), filepath.Join(dir, "native."+ext)
	runScript(t, strings.ReplaceAll(script, "%F", goFile))
	got, err := os.ReadFile(goFile)
	if err != nil {
		t.Fatalf("%s: %v", script, err)
	}
	if check != nil {
		if msg := check(got); msg != "" {
			t.Fatalf("%s: %s", script, msg)
		}
	}
	if os.Getenv("ABC_BINARY") == "" {
		return
	}
	runNativeABC(t, strings.ReplaceAll(script, "%F", nativeFile))
	want, err := os.ReadFile(nativeFile)
	if err != nil {
		t.Fatal(err)
	}
	if mask != nil {
		got, want = mask.ReplaceAll(got, nil), mask.ReplaceAll(want, nil)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s: file differs from native (%d vs %d bytes)", script, len(got), len(want))
	}
}

// TestWriteCnfLikeNative: Cnf_DataWriteIntoFile starts with ABC's comment
// line and ends with an empty line.
func TestWriteCnfLikeNative(t *testing.T) {
	nativeWriterCase(t, "read ../../../i10.aig; strash; write_cnf %F", "cnf", nil, func(b []byte) string {
		if !bytes.HasPrefix(b, []byte("c Result of efficient AIG-to-CNF conversion using package CNF\np cnf ")) {
			return "missing ABC's CNF header"
		}
		if !bytes.HasSuffix(b, []byte("0\n\n")) {
			return "missing the final empty line"
		}
		return ""
	})
}

var abcDateLine = regexp.MustCompile(`written by ABC on [^\n]*\n`)

// TestWriteGmlLikeNative: Io_WriteGml dates its header and separates the
// PO, PI, latch, node and edge sections with empty lines.
func TestWriteGmlLikeNative(t *testing.T) {
	for _, script := range []string{"read ../../../i10.aig; strash; write_gml %F", "read ../../../i10.aig; strash; if -K 4; write_gml %F"} {
		nativeWriterCase(t, script, "gml", abcDateLine, func(b []byte) string {
			if !abcDateLine.Match(b) || !bytes.HasPrefix(b, []byte("# GML for \"")) {
				return "missing ABC's dated header"
			}
			if !bytes.Contains(b, []byte("graph [\n\n    node")) {
				return "missing the empty line before the POs"
			}
			return ""
		})
	}
}
