package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var balanceExorTimes = regexp.MustCompile(`[0-9.]+ sec|[0-9.]+ MB`)

// balance -x runs Abc_NtkBalanceExor: Abc_NtkToDar with fCatchExor
// captures XOR objects, Dar_BalancePrintStats prints their supergates
// under -v, and Dar_ManBalance rebalances AND and XOR supergates. The port
// ignored fExors (so balance -x was plain balance on a different AIG) and
// printed "XOR gates: N" instead of the supergate sizes. The balanced
// network must be the native one (write_bench) and stay equivalent (cec).
func TestBalanceExorLikeNative(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"read test_vectors/native_mul8_canon.aig; strash",
		"read test_vectors/i10_native_demo.aig; strash",
		"read test_vectors/native_mul4_gen.aig; strash",
		"read test_vectors/seq/s713.blif; strash",
	}
	origs := []string{
		"test_vectors/native_mul8_canon.aig",
		"test_vectors/i10_native_demo.aig",
		"test_vectors/native_mul4_gen.aig",
		"test_vectors/seq/s713.blif",
	}
	for i, read := range files {
		for _, flags := range []string{"-x", "-x -v", "-x -l", "-x -d"} {
			script := func(out string) string {
				s := read + "; balance " + flags + "; print_stats"
				if !strings.Contains(read, "seq/") {
					s += "; cec " + origs[i]
				} else {
					s += "; dsec " + origs[i]
				}
				return s + "; short_names; write_bench " + out
			}
			goOut := filepath.Join(dir, "go.bench")
			cOut := filepath.Join(dir, "c.bench")
			os.Remove(goOut)
			os.Remove(cOut)
			clean := func(s string) string {
				var lines []string
				for _, line := range strings.Split(balanceExorTimes.ReplaceAllString(s, "T"), "\n") {
					if strings.HasPrefix(line, "ABC command line") || strings.TrimSpace(line) == "" {
						continue
					}
					lines = append(lines, strings.TrimRight(line, " "))
				}
				return strings.Join(lines, "\n")
			}
			native := clean(strings.ReplaceAll(runNativeABC(t, script(cOut)), cOut, "OUT"))
			port := clean(strings.ReplaceAll(runScript(t, script(goOut)), goOut, "OUT"))
			if native != port {
				t.Errorf("%s balance %s\nnative:\n%s\nport:\n%s", read, flags, native, port)
			}
			if !strings.Contains(port, "equivalent") || strings.Contains(port, "NOT EQUIVALENT") || strings.Contains(port, "not equivalent") {
				t.Errorf("%s balance %s: not proved equivalent:\n%s", read, flags, port)
			}
			cb, err := os.ReadFile(cOut)
			if err != nil {
				t.Fatal(err)
			}
			gb, err := os.ReadFile(goOut)
			if err != nil {
				t.Fatal(err)
			}
			// the first line is "# Benchmark ... written by ABC on <date>"
			if benchBody(cb) != benchBody(gb) {
				t.Errorf("%s balance %s: write_bench differs from native (%d vs %d bytes)", read, flags, len(cb), len(gb))
			}
		}
	}
}

// dcec -v's "Balance-X:" stage is Dar_ManBalanceXor (Aig_ManDupExor then
// Dar_ManBalance); its node counts must be native's.
func TestDcecBalanceXorLikeNative(t *testing.T) {
	balx := func(out string) string {
		var lines []string
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, "Balance-X:") {
				lines = append(lines, balanceExorTimes.ReplaceAllString(line, "T"))
			}
		}
		return strings.Join(lines, "\n")
	}
	for _, script := range []string{
		"&r test_vectors/native_mul8_canon.aig; &put; dcec -v test_vectors/cec_mul8_syn2.aig",
		"&r test_vectors/i10_native_demo.aig; &put; dcec -v test_vectors/cec_i10_syn2.aig",
	} {
		native := balx(runNativeABC(t, script))
		port := balx(runScript(t, script))
		if native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}

func benchBody(b []byte) string {
	s := string(b)
	if i := strings.IndexByte(s, '\n'); i >= 0 && strings.HasPrefix(s, "# Benchmark") {
		return s[i+1:]
	}
	return s
}
