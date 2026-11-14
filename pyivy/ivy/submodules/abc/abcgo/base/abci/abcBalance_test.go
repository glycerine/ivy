package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestBalanceFindLeftUsesEqualLevelRun(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	x := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	y := n.CreateAndEdges(a.Signal(), c.Signal()).Object
	x.Level, y.Level = 3, 3
	v := vec.Vec_PtrAlloc(3)
	vec.Vec_PtrPush(v, x.Signal())
	vec.Vec_PtrPush(v, y.Signal())
	vec.Vec_PtrPush(v, a.Signal())
	if got := Abc_NodeBalanceFindLeft(v); got != 0 {
		t.Fatalf("left bound=%d, want 0", got)
	}
}

func TestBalancePermuteMovesShareablePairToTail(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	shared := n.CreateAndEdges(a.Signal(), c.Signal())
	v := vec.Vec_PtrAlloc(3)
	vec.Vec_PtrPush(v, c.Signal())
	vec.Vec_PtrPush(v, b.Signal())
	vec.Vec_PtrPush(v, a.Signal())
	Abc_NodeBalancePermute(n, v, 0)
	if got := edgeFromAny(vec.Vec_PtrEntry(v, 1)); got != c.Signal() {
		t.Fatalf("tail candidate=%v, want shared input %v", got, c.Signal())
	}
	if shared.Object == nil {
		t.Fatal("fixture shared AND was not constructed")
	}
}

func TestBalanceConePreservesComplementedLeaves(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	root := n.CreateAndEdges(a.Signal(), b.Signal().Not()).Object
	storage := vec.Vec_VecStart(1)
	super := Abc_NodeBalanceCone(root, storage, 0, 1, 0)
	if got := vec.Vec_PtrSize(super); got != 2 {
		t.Fatalf("flattened leaves=%d, want 2", got)
	}
	vec.Vec_VecFree(storage)
}

func TestBalanceConeReturnsEmptyForContradiction(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a := abc.Abc_NtkCreatePi(n)
	root := n.CreateAndEdges(a.Signal(), a.Signal().Not()).Object
	storage := vec.Vec_VecStart(1)
	super := Abc_NodeBalanceCone(root, storage, 0, 1, 0)
	if got := vec.Vec_PtrSize(super); got != 0 {
		t.Fatalf("contradictory supergate leaves=%d, want 0", got)
	}
	if a.MarkB {
		t.Fatal("cone traversal left a temporary mark set")
	}
	vec.Vec_VecFree(storage)
}

func TestBalanceConeRecDetectsContradiction(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	root := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	v := vec.Vec_PtrAlloc(2)
	if got := Abc_NodeBalanceCone_rec(root.Fanins()[0].Object, v, 0, 1, 0); got != 0 {
		t.Fatalf("first branch ret=%d, want 0", got)
	}
	if got := balanceConeRec(root.Fanins()[0].Not(), v, 0, 1, 0); got != -1 {
		t.Fatalf("opposite-polarity ret=%d, want -1", got)
	}
}

func TestBalanceAttachLevelAndDetach(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	root := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	po := abc.Abc_NtkCreatePo(n)
	if err := n.AddFanin(po, root.Signal()); err != nil {
		t.Fatal(err)
	}
	Abc_NtkBalanceAttach(n)
	if _, ok := root.CopyData.(*Vec_Ptr_t); !ok {
		t.Fatal("root does not have an attached balance cone")
	}
	Abc_NtkBalanceLevel(n)
	if root.Level != 1 {
		t.Fatalf("root implication level=%d, want 1", root.Level)
	}
	Abc_NtkBalanceDetach(n)
	if root.CopyData != nil {
		t.Fatal("detach left root cone data attached")
	}
}
