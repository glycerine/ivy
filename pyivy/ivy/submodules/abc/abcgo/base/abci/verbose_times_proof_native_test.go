package abci

import (
	"strings"
	"testing"
)

// The proof/* ports printed Go durations ("Time = 2.2ms") where ABC prints
// Abc_PrintTime's "Time =     0.00 sec" after the C text ("...  " with two
// spaces before the time).
func TestVerboseTimesProofLikeNative(t *testing.T) {
	for _, script := range []string{
		"&r test_vectors/native_mul8_canon.aig; &profile -a",
		"&r test_vectors/native_mul4_gen.aig; &fadds -n -v; &ps",
		"&r test_vectors/native_mul8_canon.aig; &atree",
		"read test_vectors/seq/s713.blif; strash; orpos; &get; &gla -n -v -F 20",
		"read test_vectors/seq/s713.blif; strash; ssweep -P 2 -v",
		"read test_vectors/seq/s1494.blif; strash; sim3 -a -F 4 -R 2",
		"miter test_vectors/i10_c_bal2.aig test_vectors/i10_c_rw.aig; &get; &splitprove -v -P 1",
		"miter -c test_vectors/i10_c_bal2.aig test_vectors/i10_c_rw.aig; psat -A 2 -P 300",
	} {
		compareTimesWithNative(t, script)
	}
}

// &acec: Acec_Solve's stage reports. The port's Cec_ManVerify does not
// print ABC's "Networks are equivalent." verdict line, so that line is
// dropped from the native output before comparing.
func TestAcecTimesLikeNative(t *testing.T) {
	script := "&acec -v test_vectors/native_mul8_canon.aig test_vectors/native_mul8_canon.aig"
	clean := func(s string) string {
		s = percents.ReplaceAllString(secTimes.ReplaceAllString(mbTimes.ReplaceAllString(bmc3Normalize(s), "M MB"), "T sec"), "(P %)")
		var keep []string
		for _, line := range strings.Split(s, "\n") {
			if !strings.HasPrefix(line, "Networks are equivalent.") {
				keep = append(keep, line)
			}
		}
		return strings.Join(keep, "\n")
	}
	port := clean(runScript(t, script))
	native := clean(runNativeABC(t, script))
	if native != port {
		t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
	}
}

// ssweep -N -v: Fra_BmcPerform's prefix report and Fra_ManPrint's runtime
// table (ABC_PRT lines). The port does not yet print fraInd.c's
// simulation and per-iteration lines, so only these lines are compared.
func TestSsweepPrefixTimesLikeNative(t *testing.T) {
	keep := func(s string) string {
		s = secTimes.ReplaceAllString(bmc3Normalize(s), "T sec")
		var lines []string
		for _, line := range strings.Split(s, "\n") {
			for _, prefix := range []string{"Original AIG", "Before BMC", "After  BMC", "After BMC", "AIG simulation", "AIG traversal", "SAT solving", "    Unsat", "    Sat", "    Fail", "Class refining", "TOTAL RUNTIME"} {
				if strings.HasPrefix(line, prefix) {
					lines = append(lines, line)
					break
				}
			}
		}
		return strings.Join(lines, "\n")
	}
	for _, script := range []string{
		"read test_vectors/seq/s713.blif; strash; ssweep -N 2 -v",
		"read test_vectors/seq/s1494.blif; strash; ssweep -N 3 -v",
	} {
		port, native := keep(runScript(t, script)), keep(runNativeABC(t, script))
		if native == "" || port != native {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
