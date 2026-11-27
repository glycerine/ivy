package abci

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

// Verbose reports in aig/aig, aig/gia, aig/ivy and aig/saig print ABC times
// ("Time =%9.2f sec") and the same surrounding text as ABC; the ports
// printed Go durations ("Time=1.04ms") and abbreviated the lines.
func TestVerboseTimesAigLikeNative(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/i10_c_bal2.aig; &get; &profile -m",
		"read ../../../i10.aig; &get; &popart",
		"read ../../../i10.aig; &get; &test",
		"read test_vectors/native_mul8_canon.aig; &get; &satoko -s",
		"read test_vectors/native_mul8_canon.aig; &get; &sat3 -s",
		"read test_vectors/native_mul8_canon.aig; &get; &sim_gen -W 1; &simrsb",
		"read test_vectors/native_mul8_canon.aig; &get; &sim_gen -W 1; &sim_gen -R 3 -v",
		"read test_vectors/native_mul8_canon.aig; &get; &sim_gen -W 1; &sim_gen -d -v",
	} {
		util.C_Srand(1) // native runs each script in a fresh process: rand() starts unseeded
		compareTimesWithNative(t, script)
	}
	// Aig_ManRandom keeps its state across scripts in this process (a fresh
	// native process starts unseeded), so the chosen cofactoring variables,
	// and the cofactor count, depend on earlier tests: compare the format.
	compareTimeLinesWithNative(t, "read ../../../i10.aig; cone -O 5; strash; &get; &posplit -N 3 -v")
}
