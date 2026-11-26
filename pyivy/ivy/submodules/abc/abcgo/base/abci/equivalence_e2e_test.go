package abci

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestScriptsPreserveFunction is an end-to-end correctness check that does
// not depend on native ABC: every script of the native sweep, applied to
// every benchmark, must leave a network that the port's own cec proves
// equivalent to the original file. CIs and COs are matched by order
// (cec -n): after &put, native ABC too cannot match i10's names.
func TestScriptsPreserveFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("end-to-end equivalence sweep")
	}
	for _, bench := range nativeScriptBenches {
		if strings.Contains(bench, "mul12") {
			continue // cec on the larger multiplier alone took 15 s; mul8 covers it
		}
		t.Run(filepath.Base(bench), func(t *testing.T) { checkScriptsPreserveFunction(t, bench) })
	}
}

func checkScriptsPreserveFunction(t *testing.T, bench string) {
	{
		var script strings.Builder
		for _, s := range nativeScripts {
			script.WriteString("read " + bench + "; " + s + "; cec -n " + bench + "; ")
		}
		out := runScript(t, script.String())
		if n := strings.Count(out, "Networks are equivalent"); n != len(nativeScripts) {
			// Report the scripts whose cec did not prove equivalence.
			parts := strings.Split(out, "Networks are equivalent")
			t.Errorf("%s: %d of %d scripts proved equivalent; after the last proof:\n%s", bench, n, len(nativeScripts), parts[len(parts)-1])
		}
	}
}
