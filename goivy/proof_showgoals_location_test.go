package goivy

import (
	"strings"
	"testing"
)

// Python showgoals: print('{}Proof goals:'.format(proof.lineno)) prints the
// full location (file and line). goivy printed only "line N: ".
func TestShowGoalsPrintsFullLocation(t *testing.T) {
	pc := mkPCWithModule(NewWithSig(NewSig()))
	sg := &ShowGoalsTactic{}
	sg.SetLineno(Location{Filename: "x.ivy", Line: 10})
	out := captureStdout(t, func() { pc.showGoalsTactic(nil, sg) })
	if !strings.Contains(out, "\nx.ivy: line 10: Proof goals:\n") {
		t.Fatalf("output %q, want the full location before Proof goals:", out)
	}
}
