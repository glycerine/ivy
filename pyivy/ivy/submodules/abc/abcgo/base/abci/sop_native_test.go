package abci

import (
	"strings"
	"testing"
)

// TestSopCubeLimitLikeNative: Abc_NtkBddToSop counts the ISOP cubes of both
// phases first (Extra_bddCountCubes, limit 1000000) to choose each node's
// phase, then builds only that phase's cover; on i10's collapsed network a
// cover exceeds ABC_MAX_CUBES and Abc_ConvertBddToSop says so.
func TestSopCubeLimitLikeNative(t *testing.T) {
	out := runScript(t, "read ../../test_vectors/i10.aig; collapse; sop; print_stats")
	for _, want := range []string{"The number of cubes exceeded the predefined limit (100000).\n", "bdd  = 79647"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}
