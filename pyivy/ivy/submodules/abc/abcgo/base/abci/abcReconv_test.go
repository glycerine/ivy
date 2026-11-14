package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestReconvCutExpansionAndCleanup(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b, c, d := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	x := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	y := n.CreateAndEdges(x.Signal(), c.Signal()).Object
	root := n.CreateAndEdges(y.Signal(), d.Signal()).Object
	m := Abc_NtkManCutStart(2, 4, 100, 100)
	defer Abc_NtkManCutStop(m)
	cut := Abc_NodeFindCut(m, root, 1)
	if vec.Vec_PtrSize(cut) != 2 || vec.Vec_PtrSize(Abc_NtkManCutReadCutLarge(m)) < 2 {
		t.Fatalf("unexpected cuts: small=%d large=%d", vec.Vec_PtrSize(cut), vec.Vec_PtrSize(Abc_NtkManCutReadCutLarge(m)))
	}
	for _, obj := range vec.Vec_PtrArray(Abc_NtkManCutReadVisited(m)) {
		if obj.(*abc.Object).MarkB {
			t.Fatal("cut construction left mark B set")
		}
	}
}

func TestReconvTwoLevelCostRecognizesSingleNewLeaf(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	x := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	root := n.CreateAndEdges(x.Signal(), c.Signal()).Object
	a.MarkB = true
	b.MarkB = true
	root.MarkB = true
	var leaf, m1, m2 *Abc_Obj_t
	if got := Abc_NodeGetLeafCostTwo(root, 10, &leaf, &m1, &m2); got != 1 || leaf != c || m1 != x || m2 != nil {
		t.Fatalf("cost-two result=%d leaf=%p marked=(%p,%p)", got, leaf, m1, m2)
	}
}
