package rwr

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/misc/extra/npn"
)

func rwrFixtureManager() *Rwr_Man_t {
	canon, phases, perms, classes := npn.Extra_Truth4VarNPN()
	p := &Rwr_Man_t{nFuncs: 1 << 16, puCanons: canon, pPhases: phases, pPerms: perms, pMap: classes, pTable: make([]*Rwr_Node_t, 1<<16), nTravIds: 1}
	p.pPractical = Rwr_ManGetPractical(p)
	for _, truth := range []uint32{0, 0xaaaa, 0xcccc, 0xf0f0, 0xff00} {
		Rwr_ManAddVar(p, truth, 0)
	}
	return p
}
func TestSourceLibraryArrayAndBinaryRoundTrip(t *testing.T) {
	first := rwrFixtureManager()
	Rwr_ManLoadFromArray(first, 0)
	if len(first.vForest) != 1785 {
		t.Fatalf("source forest size %d", len(first.vForest))
	}
	path := filepath.Join(t.TempDir(), "library.bin")
	Rwr_ManWriteToFile(first, path)
	second := rwrFixtureManager()
	Rwr_ManLoadFromFile(second, path)
	if len(first.vForest) != len(second.vForest) {
		t.Fatal("forest size changed")
	}
	for i := range first.vForest {
		a, b := first.vForest[i], second.vForest[i]
		if a.uTruth != b.uTruth || a.Level != b.Level || a.Volume != b.Volume || a.fExor != b.fExor {
			t.Fatalf("node %d differs", i)
		}
	}
}
