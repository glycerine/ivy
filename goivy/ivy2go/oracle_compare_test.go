package ivy2go

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- OPEN 057: behavioural oracle harness tests ----------------------
//
// The harness runs two pre-built binaries on the same stdin and diffs
// stdouts. For these tests we use /bin/sh -c wrappers that simulate
// what ivy2cpp and ivy2go-emitted binaries would do — so we exercise
// the diff logic without needing real generated programs.

// shScriptBinary writes a small shell script + chmods it +x; returns
// the path. Caller is responsible for using a t.TempDir().
func shScriptBinary(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	full := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(path, []byte(full), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func TestCompareGoVsCpp_IdenticalStdoutMatches(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no /bin/sh available")
	}
	dir := t.TempDir()
	goBin := shScriptBinary(t, dir, "go.sh", `echo "hello"; echo "world"`)
	cppBin := shScriptBinary(t, dir, "cpp.sh", `echo "hello"; echo "world"`)
	if err := CompareGoVsCpp(goBin, cppBin, ""); err != nil {
		t.Fatalf("identical stdouts should match: %v", err)
	}
}

func TestCompareGoVsCpp_DivergingStdoutReportsLine(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no /bin/sh available")
	}
	dir := t.TempDir()
	goBin := shScriptBinary(t, dir, "go.sh", `echo "alpha"; echo "beta"`)
	cppBin := shScriptBinary(t, dir, "cpp.sh", `echo "alpha"; echo "gamma"`)
	err := CompareGoVsCpp(goBin, cppBin, "")
	if err == nil {
		t.Fatal("divergent stdouts should not match")
	}
	diff, ok := err.(*OracleDiff)
	if !ok {
		t.Fatalf("expected *OracleDiff, got %T: %v", err, err)
	}
	if diff.Line != 2 {
		t.Errorf("divergence line = %d, want 2", diff.Line)
	}
	if diff.Go != "beta" || diff.Cpp != "gamma" {
		t.Errorf("diff entries unexpected: %+v", diff)
	}
}

func TestCompareGoVsCpp_DifferingLineCountFlags(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no /bin/sh available")
	}
	dir := t.TempDir()
	goBin := shScriptBinary(t, dir, "go.sh", `echo "a"; echo "b"; echo "c"`)
	cppBin := shScriptBinary(t, dir, "cpp.sh", `echo "a"; echo "b"`)
	err := CompareGoVsCpp(goBin, cppBin, "")
	if err == nil {
		t.Fatal("extra Go line should be flagged")
	}
	diff, ok := err.(*OracleDiff)
	if !ok {
		t.Fatalf("expected *OracleDiff, got %T: %v", err, err)
	}
	if diff.Line != 3 || diff.Go != "c" || diff.Cpp != "<EOF>" {
		t.Errorf("expected line 3 with Go=c Cpp=<EOF>, got: %+v", diff)
	}
}

func TestCompareGoVsCpp_StdinIsPipedToBothBinaries(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no /bin/sh available")
	}
	dir := t.TempDir()
	// Both scripts echo stdin back; if the same input is piped to
	// both, they should produce matching output.
	goBin := shScriptBinary(t, dir, "go.sh", `cat`)
	cppBin := shScriptBinary(t, dir, "cpp.sh", `cat`)
	if err := CompareGoVsCpp(goBin, cppBin, "shared input\n"); err != nil {
		t.Fatalf("identical cat output should match: %v", err)
	}
}

func TestCompareGoVsCpp_NormalisesTrailingWhitespace(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no /bin/sh available")
	}
	dir := t.TempDir()
	goBin := shScriptBinary(t, dir, "go.sh", `printf "hello  \n"`)
	cppBin := shScriptBinary(t, dir, "cpp.sh", `printf "hello\n"`)
	if err := CompareGoVsCpp(goBin, cppBin, ""); err != nil {
		t.Fatalf("trailing-whitespace-only diff should not flag: %v", err)
	}
}

func TestSplitNormalisedLines_DropsTrailingEmpty(t *testing.T) {
	got := splitNormalisedLines("a\nb\n")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("splitNormalisedLines = %v, want [a, b]", got)
	}
}

func TestSplitNormalisedLines_HandlesCRLF(t *testing.T) {
	got := splitNormalisedLines("a\r\nb\r\n")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("CRLF split = %v, want [a, b]", got)
	}
}

func TestOracleDiffError(t *testing.T) {
	d := &OracleDiff{Line: 5, Go: "alpha", Cpp: "beta"}
	s := d.Error()
	if !strings.Contains(s, "line 5") || !strings.Contains(s, "alpha") || !strings.Contains(s, "beta") {
		t.Errorf("Error() should mention line / go / cpp values, got %q", s)
	}
}
