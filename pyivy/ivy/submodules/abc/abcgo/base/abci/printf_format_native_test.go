package abci

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Abc_Print has no printf format attribute, so gcc never checked its
// arguments. Compiling with a prototype that carries one found reports with
// missing, extra or misplaced arguments (CBUGS_FIXED.md #22). This compiles
// the files that had them and requires no -Wformat mismatch.
func TestNativeAbcPrintFormatsMatchArguments(t *testing.T) {
	gcc, err := testoracle.LookPath("gcc")
	if err != nil {
		t.Skip("gcc not found")
	}
	src, err := filepath.Abs("../../../src")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testoracle.Stat(filepath.Join(src, "base/abci/abc.c")); err != nil {
		t.Skip("ABC C sources not found")
	}
	decl := filepath.Join(t.TempDir(), "fmtdecl.h")
	if err := os.WriteFile(decl, []byte("static inline void Abc_Print( int level, const char * format, ... ) __attribute__((format(printf,2,3)));\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{
		"base/abci/abc.c", "proof/cec/cecCore.c", "base/abci/abcNpnSave.c",
		"aig/gia/giaGiarf.c", "sat/proof/pr.c", "proof/pdr/pdrIncr.c",
	} {
		cmd := testoracle.Command(t, gcc, "-fsyntax-only", "-Wformat", "-Wno-format-zero-length", "-Wno-format-overflow",
			"-Wno-format-truncation", "-Wno-format-security",
			"-DLIN64", "-DSIZEOF_VOID_P=8", "-DSIZEOF_LONG=8", "-DSIZEOF_INT=4", "-DABC_USE_CUDD=1", "-DABC_USE_PTHREADS",
			"-I.", "-include", decl, file)
		cmd.Dir = src
		out, _ := cmd.CombinedOutput()
		for _, line := range strings.Split(string(out), "\n") {
			// a missing or misplaced argument, or a stray % directive; %d
			// given a long is a separate, benign class left as it is
			extra := strings.Contains(line, "[-Wformat-extra-args]") && (file == "proof/cec/cecCore.c" || file == "proof/pdr/pdrIncr.c")
			if extra || strings.Contains(line, "[-Wformat=]") && !strings.Contains(line, "but argument") {
				t.Errorf("%s", line)
			}
		}
	}
}

// powerdown's usage printed "5% means" through printf: "% m" is a
// directive and printed strerror(errno), "5No such file or directoryeans".
func TestNativePowerdownUsageText(t *testing.T) {
	bin := testoracle.NativeABCPath()
	if bin == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	bin, _ = filepath.Abs(bin)
	out, _ := testoracle.Command(t, bin, "-c", "powerdown -h").CombinedOutput()
	if !strings.Contains(string(out), "(e.g. 5% means hot wires") {
		t.Fatalf("powerdown -h:\n%s", out)
	}
}

// Cec_ManSimulation reported "rounds of random simulation of <rounds>
// frames with <frames> words"; the port copied the dropped argument.
func TestEquivSeqSimulationReport(t *testing.T) {
	const script = "read test_vectors/seq/s713.blif; strash; &get; &equiv -s"
	const want = "Performing 20 rounds of random simulation of 100 frames with 31 words."
	if out := runScript(t, script); !strings.Contains(out, want) {
		t.Errorf("port:\n%s", out)
	}
	if testoracle.NativeABCPath() != "" {
		if out := runNativeABC(t, script); !strings.Contains(out, want) {
			t.Errorf("native:\n%s", out)
		}
	}
}
