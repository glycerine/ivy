package res

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func makeWindowChain() (*abc.Network, *abc.Object, *abc.Object, *abc.Object, *abc.Object) {
	n := abc.NewNetwork("window")
	n.Type = uint8(abc.ABC_NTK_STRASH)
	n.Function = uint8(abc.ABC_FUNC_AIG)
	m := abc.Abc_AigAlloc(n)
	n.FunctionManager = m
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	inner := abc.Abc_AigAnd(m, a.Signal(), b.Signal()).Object
	root := abc.Abc_AigAnd(m, inner.Signal(), c.Signal()).Object
	abc.Abc_ObjSetLevel(inner, 1)
	abc.Abc_ObjSetLevel(root, 2)
	co := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(co, root.Signal())
	return n, a, b, c, root
}

func TestResWindowCollectsBoundedTFI(t *testing.T) {
	_, a, b, c, root := makeWindowChain()
	w := Res_WinAlloc()
	defer Res_WinFree(w)
	if got := Res_WinCompute(root, 1, 0, w); got != 1 {
		t.Fatalf("window computation returned %d, want 1", got)
	}
	if got := Vec_PtrSize(w.vNodes); got != 2 {
		t.Fatalf("window has %d internal nodes, want 2", got)
	}
	if got := Vec_PtrSize(w.vLeaves); got != 3 {
		t.Fatalf("window has %d leaves, want 3", got)
	}
	want := map[*abc.Object]bool{a: true, b: true, c: true}
	for i := int32(0); i < Vec_PtrSize(w.vLeaves); i++ {
		leaf := Vec_PtrEntry(w.vLeaves, i)
		if !want[leaf] {
			t.Fatalf("unexpected window leaf %q", leaf.Name())
		}
		delete(want, leaf)
	}
	if len(want) != 0 {
		t.Fatalf("window omitted %d primary inputs", len(want))
	}
}

func TestResWindowFreeReleasesOwnedVectors(t *testing.T) {
	w := Res_WinAlloc()
	Res_WinFree(w)
	Res_WinFree(w)
	if w.vRoots != nil || w.vLeaves != nil || w.vBranches != nil || w.vNodes != nil || w.vDivs != nil || w.vMatrix != nil {
		t.Fatal("freed window retains vector storage")
	}
}

func TestResWindowMFFCLabelsWithoutChangingGraph(t *testing.T) {
	n, _, _, _, root := makeWindowChain()
	before := len(n.Objects())
	abc.Abc_NtkIncrementTravId(n)
	if got := Res_WinVisitMffc(root); got != 2 {
		t.Fatalf("MFFC size %d, want 2", got)
	}
	if len(n.Objects()) != before {
		t.Fatal("MFFC labeling changed network objects")
	}
	if abc.Abc_NodeIsTravIdCurrent(root) == 0 || abc.Abc_NodeIsTravIdCurrent(abc.Abc_ObjFanin0(root)) == 0 {
		t.Fatal("MFFC nodes were not labeled with the current traversal ID")
	}
}
