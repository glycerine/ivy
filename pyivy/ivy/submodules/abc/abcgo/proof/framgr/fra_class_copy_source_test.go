package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCopyRepresentativesChecksEmptyClassesBeforeFailures(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	p := Fra_ClassesStart(g)
	p.Reprs[x.NodeID()] = 0
	failed := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(failed, x)
	defer func() {
		if recover() == nil {
			t.Error("source checks stale representatives before dropping failures")
		}
		if got := aig.Aig_ObjRepr(g, x); got != g.Const1() {
			t.Errorf("copied source representative = %v", got)
		}
	}()
	Fra_ClassesCopyReprs(p, failed)
}
