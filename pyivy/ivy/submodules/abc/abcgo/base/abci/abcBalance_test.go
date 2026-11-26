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

func TestBalancePermuteSinglePairDoesNotConsultNetwork(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	a, b := abc.Abc_NtkCreatePi(n).Signal(), abc.Abc_NtkCreatePi(n).Signal().Not()
	v := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(v, a)
	vec.Vec_PtrPush(v, b)
	Abc_NodeBalancePermute(nil, v, 0)
	if vec.Vec_PtrEntry(v, 0) != a || vec.Vec_PtrEntry(v, 1) != b {
		t.Fatal("a single candidate pair changed order")
	}
}

func TestBalanceWideSupergatePreservesEveryInputPhase(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	const width = 11
	root := n.One()
	for i := range width {
		edge := abc.Abc_NtkCreatePi(n).Signal()
		if i%2 != 0 {
			edge = edge.Not()
		}
		root = n.AndEdges(root, edge)
	}
	abc.Abc_ObjAddFanin(abc.Abc_NtkCreatePo(n), root)
	balanced := Abc_NtkBalance(n, 0, 0, 1)
	if balanced == nil {
		t.Fatal("balancing failed")
	}
	defer abc.Abc_NtkDelete(balanced)
	if got := abc.Abc_AigLevel(balanced); got != 4 || abc.Abc_NtkNodeNum(balanced) != width-1 {
		t.Fatalf("balanced %d-input AND has %d levels and %d nodes", width, got, abc.Abc_NtkNodeNum(balanced))
	}
	inputs := balanced.Inputs()
	values := make(map[*abc.Object]bool, width)
	output := abc.Abc_ObjChild0(balanced.Outputs()[0])
	for a := 0; a < 1<<width; a++ {
		for i, input := range inputs {
			values[input] = a>>i&1 != 0
		}
		got, err := balanced.Evaluate(output, values)
		if err != nil || got != (a == 0x555) {
			t.Fatalf("assignment %#x produced %t, err %v", a, got, err)
		}
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
	if got := Abc_NodeBalanceCone_rec(root.FaninAt(int32(0)).Object, v, 0, 1, 0); got != 0 {
		t.Fatalf("first branch ret=%d, want 0", got)
	}
	if got := balanceConeRec(root.FaninAt(int32(0)).Not(), v, 0, 1, 0); got != -1 {
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

func TestBalancePermuteUsesOwningAIGHash(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	m := n.FunctionManager.(*abc.Abc_Aig_t)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	shared := abc.Abc_AigAnd(m, a, c)
	abc.Abc_AigAndDelete(m, shared.Object)
	v := vec.Vec_PtrAlloc(3)
	vec.Vec_PtrPush(v, c.Signal())
	vec.Vec_PtrPush(v, b.Signal())
	vec.Vec_PtrPush(v, a.Signal())
	Abc_NodeBalancePermute(n, v, 0)
	if got := vec.Vec_PtrEntry(v, 1); got != b.Signal() {
		t.Errorf("selected an excluded AND's input %v, want %v", got, b.Signal())
	}
	if got := testing.AllocsPerRun(50, func() { Abc_NodeBalancePermute(n, v, 0) }); got != 0 {
		t.Errorf("balance pair search allocated %.0f objects, want 0", got)
	}
}

func TestBalanceLeafStorageKeepsSharedConeOrder(t *testing.T) {
	for _, mode := range []struct{ duplicate, selective int32 }{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
		n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
		a, b, c := n.NewObject(abc.ObjectPI).Signal(), n.NewObject(abc.ObjectPI).Signal(), n.NewObject(abc.ObjectPI).Signal()
		shared := n.CreateAndEdges(a, b.Not())
		branch := n.CreateAndEdges(shared, c)
		root := n.CreateAndEdges(branch, shared)
		want := []abc.Edge{a, b.Not(), c}
		if mode.duplicate == 0 && mode.selective == 0 {
			want = []abc.Edge{shared, c}
		}
		for _, vector := range []bool{false, true} {
			var leaves balanceLeaves
			if vector {
				leaves.vector = vec.Vec_PtrAlloc(0)
			}
			out, ret := collectBalanceEdges(root, leaves, 1, mode.duplicate, mode.selective)
			if ret != 1 {
				t.Fatalf("mode %+v repeated cone status %d", mode, ret)
			}
			got := out.edges
			if vector {
				for _, item := range vec.Vec_PtrArray(out.vector) {
					got = append(got, edgeFromAny(item))
				}
				vec.Vec_PtrFree(out.vector)
			}
			if len(got) != len(want) {
				t.Fatalf("mode %+v size %d, want %d", mode, len(got), len(want))
			}
			for i, e := range got {
				if e != want[i] || !e.Object.MarkB {
					t.Fatalf("mode %+v leaf %d changed phase, order or mark", mode, i)
				}
				e.Object.MarkB = false
			}
		}
		m := balanceAigManager(n)
		g := m.AndEdges(a.Not(), b)
		if g != abc.Abc_AigAnd(m, b, a.Not()) || m.AndEdges(c, c.Not()) != n.One().Not() {
			t.Fatal("typed manager lost a phase or hash hit")
		}
		abc.Abc_NtkDelete(n)
	}
}
