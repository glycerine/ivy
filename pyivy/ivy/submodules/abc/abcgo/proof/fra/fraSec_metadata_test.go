package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// fraSec.c writes nTruePis/nTruePos directly before several stages.
func TestSecRefreshTrueCountsPreservesCioIDs(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x, y := g.AddOutput(a), g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManCleanCioIds(g)
	g.SetTruePiCount(0)
	g.SetTruePoCount(0)
	secRefreshTrueCounts(g)
	if g.TruePiCount() != 1 || g.TruePoCount() != 1 {
		t.Fatal("true terminal counts were not refreshed")
	}
	for _, e := range []aig.Edge{a, b, x, y} {
		if aig.Aig_ObjCioId(g, e) != 0 {
			t.Fatal("refresh changed a stored terminal ID")
		}
	}
}
