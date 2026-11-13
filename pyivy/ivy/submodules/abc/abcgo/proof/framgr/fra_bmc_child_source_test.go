package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestBMCChildSourceMappingAndParentPolarity(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	n := g.AND(a.Not(), b)
	m := Fra_BmcStart(g, 0, 1)
	Bmc_ObjSetFrames(m, a, 0, a)
	Bmc_ObjSetFrames(m, b, 0, b)
	if got := Bmc_ObjChild0Frames(m, n, 0); got != a.Not() {
		t.Fatalf("first mapped child = %v", got)
	}
	if got := Bmc_ObjChild1Frames(m, n, 0); got != b {
		t.Fatalf("second mapped child = %v", got)
	}
	for _, child := range []func(*Fra_Bmc_t, aig.Edge, int32) aig.Edge{Bmc_ObjChild0Frames, Bmc_ObjChild1Frames} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("source rejects complemented parent")
				}
			}()
			child(m, n.Not(), 0)
		}()
	}
}
