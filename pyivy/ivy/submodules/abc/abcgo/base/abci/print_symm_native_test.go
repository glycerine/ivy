package abci

import (
	"regexp"
	"strings"
	"testing"
)

var printSymmTime = regexp.MustCompile(`[0-9.]+ sec`)

func printSymmNormalize(out string) string {
	var lines []string
	for _, line := range strings.Split(printSymmTime.ReplaceAllString(out, "T sec"), "\n") {
		if strings.HasPrefix(line, "ABC command line") || strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return strings.Join(lines, "\n")
}

// print_symm's simulation/SAT engine (simSym*.c) is a line-by-line port:
// the same rand() stream, structural pass, simulation rounds, FRAIG
// SAT checks and Sym_ManPrintStats block as C.
func TestPrintSymmLikeNative(t *testing.T) {
	designs := []string{
		"test_vectors/native_mul4_gen.aig; strash",
		"test_vectors/native_mul8_canon.aig; strash",
		"test_vectors/i10_native_demo.aig; strash",
		"test_vectors/seq/s713.blif",
		"test_vectors/seq/s713.blif; comb",
		"test_vectors/seq/s713.blif; strash; comb",
	}
	flags := []string{"", "-v", "-n", "-b", "-r", "-bv", "-nv"}
	for _, design := range designs {
		for _, flag := range flags {
			if strings.Contains(design, "i10") && strings.Contains(flag, "n") {
				continue // the naive BDD engine is slow on i10
			}
			script := "read " + design + "; print_symm " + flag
			native := printSymmNormalize(runNativeABC(t, script))
			out, _ := runScriptStatus(t, script) // s713 without comb is an error in both
			port := printSymmNormalize(out)
			if native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
		}
	}
}
