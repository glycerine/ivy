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

// TestWriteSmvLikeNative: write_smv follows Io_Write and Io_WriteSmv: AIGs
// only, through Abc_NtkToNetlistBench; names with parentheses are refused
// (nothing written); otherwise ABC's layout, gates and init values.
func TestWriteSmvLikeNative(t *testing.T) {
	file := filepath.Join(t.TempDir(), "x.smv")
	runScript(t, "read ../../../i10.aig; strash; write_smv "+file)
	if _, err := os.Stat(file); err == nil {
		t.Fatal("write_smv wrote a file for names with parentheses")
	}
	nativeWriterCase(t, "read ../../../i10.aig; strash; short_names; write_smv %F", "smv", abcDateLine, func(b []byte) string {
		if !bytes.HasPrefix(b, []byte("-- benchmark \"")) || !bytes.Contains(b, []byte("MODULE main\n\nVAR  -- inputs\n")) {
			return "missing ABC's SMV header"
		}
		return ""
	})
}

// TestWriteBenchLikeNative: write_bench (LUT form by default, gates with -l)
// refuses names with parentheses without creating the file, and otherwise
// writes ABC's dated header and body.
func TestWriteBenchLikeNative(t *testing.T) {
	file := filepath.Join(t.TempDir(), "x.bench")
	runScript(t, "read ../../../i10.aig; strash; write_bench "+file)
	if _, err := os.Stat(file); err == nil {
		t.Fatal("write_bench created a file for names with parentheses")
	}
	for _, script := range []string{
		"read ../../../i10.aig; strash; short_names; write_bench %F",
		"read ../../../i10.aig; strash; short_names; write_bench -l %F",
		"read ../../../i10.aig; strash; short_names; if -K 4; write_bench %F",
	} {
		nativeWriterCase(t, script, "bench", abcDateLine, func(b []byte) string {
			if !bytes.HasPrefix(b, []byte("# Benchmark \"")) || !abcDateLine.Match(b) {
				return "missing ABC's dated header"
			}
			return ""
		})
	}
}

// TestReadEqnLikeNative: an EQN file (written by ABC from mapped i10) reads
// back into the same AIG as native's (one AIG node per equation,
// Parse_FormulaParserEqn's structure).
func TestReadEqnLikeNative(t *testing.T) {
	eqn := filepath.Join(t.TempDir(), "i10.eqn")
	runScript(t, "read ../../../i10.aig; strash; short_names; if -K 4; write_eqn "+eqn)
	nativeWriterCase(t, "read "+eqn+"; strash; write_aiger %F", "aig", regexp.MustCompile(`This file was written by ABC on [^\n]*\n`), nil)
}
