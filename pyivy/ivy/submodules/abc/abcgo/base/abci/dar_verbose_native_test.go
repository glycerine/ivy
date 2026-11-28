package abci

import (
	"strings"
	"testing"
)

// TestDarVerboseStatsLikeNative: drw -v and drf -v print Dar_ManPrintStats
// and Dar_ManRefPrintStats (node, cut and buffer counters); the port printed
// nothing.
func TestDarVerboseStatsLikeNative(t *testing.T) {
	out := runScript(t, "read ../../test_vectors/i10.aig; strash; drw -v")
	for _, want := range []string{
		"Tried =     2654. Beg =     2675. End =     2124. Gain =    551. ( 20.60 %).  Cut mem = 0 MB\n",
		"Cuts =    85270. Tried =    22011. Used =    17042. Bad =     0. Skipped =     5. Ave = 8.02.\n",
		"Bufs =     0. BufMax =     2. BufReplace =    134. BufFix =    291.  Levels =   48.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("drw -v: missing %q", want)
		}
	}
	out = runScript(t, "read ../../test_vectors/i10.aig; strash; drf -v")
	for _, want := range []string{
		"NodesBeg =     2675. NodesEnd =     2301. Gain =    374. ( 13.98 %).\n",
		"Tried =   1574. Below =     0. Extended =     0.  Used =   325.  Levels =   49.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("drf -v: missing %q", want)
		}
	}
}
