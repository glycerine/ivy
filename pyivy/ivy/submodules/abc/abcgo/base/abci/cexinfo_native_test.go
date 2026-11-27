package abci

import (
	"strings"
	"testing"
)

// TestCexInfoLikeNative: &cexinfo (Bmc_CexTest) prints the inner states,
// implications, care bits, essentials and minimized counterexample with
// Abc_CexPrintStatsInputs/Bmc_CexPrint, then the time; the port printed
// nothing.
func TestCexInfoLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/native_s27_drw.blif; strash; bmc3; &get; &cexinfo -v")
	want := `Inner states: CEX: Po =   0  Fr =   0  FF = 0  PI = 7  Bit =      7  1 =       0 ( 0.00 %) 1pi =       0 ( 0.00 %) 1ppi =       0 ( 0.00 %)
  0 : 0000   0 000   0
Implications: CEX: Po =   0  Fr =   0  FF = 0  PI = 7  Bit =      7  1 =       3 (42.86 %) 1pi =       0 ( 0.00 %) 1ppi =       3 (100.00 %)
  0 : 0000   0 111   3
Care bits:    CEX: Po =   0  Fr =   0  FF = 0  PI = 7  Bit =      7  1 =       1 (14.29 %) 1pi =       1 (25.00 %) 1ppi =       0 ( 0.00 %)
  0 : 0001   1 000   0
Essentials:   CEX: Po =   0  Fr =   0  FF = 0  PI = 7  Bit =      7  1 =       1 (14.29 %) 1pi =       1 (25.00 %) 1ppi =       0 ( 0.00 %)
  0 : 0001   1 000   0
Minimized:    CEX: Po =   0  Fr =   0  FF = 0  PI = 7  Bit =      7  1 =       1 (14.29 %) 1pi =       1 (25.00 %) 1ppi =       0 ( 0.00 %)
  0 : 0001   1 000   0
Time =`
	if !strings.Contains(out, want) {
		t.Fatalf("&cexinfo -v:\n%s", out)
	}
}
