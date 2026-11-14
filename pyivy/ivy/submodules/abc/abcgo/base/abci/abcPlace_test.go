package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	physplace "github.com/glycerine/abcgo/phys/place"
)

func makePlacementTestNetwork(t *testing.T) (*abc.Network, *abc.Object, *abc.Object, abc.Edge) {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	ab := abc.Abc_AigAnd(manager, a, abc.Abc_ObjNot(b))
	root := abc.Abc_AigAnd(manager, ab, c)
	po := abc.Abc_NtkCreatePo(n)
	po.SetName("out")
	abc.Abc_ObjAddFanin(po, root)
	abc.Abc_NtkOrderCisCos(n)
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("invalid placement test network")
	}
	return n, a, b, root
}

func TestAbcPlaceBeginBuildsCellsAndFanoutNets(t *testing.T) {
	n, a, b, root := makePlacementTestNetwork(t)
	Abc_PlaceBegin(n)
	defer Abc_PlaceEnd(n)
	if cells[a.ID] == nil || cells[b.ID] == nil || cells[root.Object.ID] == nil {
		t.Fatal("placement omitted a PI or AND cell")
	}
	if nets[a.ID] == nil || nets[b.ID] == nil || nets[root.Object.ID] == nil {
		t.Fatal("placement omitted a CI or node fanout net")
	}
	if cells[a.ID].ID() != int32(a.ID) || cells[root.Object.ID].ID() != int32(root.Object.ID) {
		t.Fatal("placement cell IDs do not match ABC object IDs")
	}
	cut := vec.Vec_PtrAlloc(2)
	defer vec.Vec_PtrFree(cut)
	vec.Vec_PtrPush(cut, a)
	vec.Vec_PtrPush(cut, b)
	if cost := Abc_PlaceEvaluateCut(root.Object, cut); cost != 0 {
		t.Fatalf("source placeholder cut cost = %v, want 0", cost)
	}
	Abc_PlaceEnd(n)
	if cells != nil || nets != nil || abstractCells != nil {
		t.Fatal("placement records remain allocated after Abc_PlaceEnd")
	}
}

func TestAbcPlaceUpdateRefreshesNewCellsAndFanoutTerminals(t *testing.T) {
	n, a, _, root := makePlacementTestNetwork(t)
	Abc_PlaceBegin(n)
	defer Abc_PlaceEnd(n)
	oldNet := nets[a.ID]
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	added := abc.Abc_AigAnd(manager, a, root.Object)
	addedCell := vec.Vec_PtrAlloc(1)
	updatedNets := vec.Vec_PtrAlloc(1)
	defer vec.Vec_PtrFree(addedCell)
	defer vec.Vec_PtrFree(updatedNets)
	vec.Vec_PtrPush(addedCell, added.Object)
	vec.Vec_PtrPush(updatedNets, a)
	Abc_PlaceUpdate(addedCell, updatedNets)
	if cells[added.Object.ID] == nil || nets[added.Object.ID] == nil || nets[a.ID] == oldNet {
		t.Fatal("incremental update did not install the new cell and replace the changed net")
	}
	oldFanin := n.ObjectsOfKind(abc.ObjectNode)[0]
	cells[oldFanin.ID].SetPosition(10, 10)
	cells[added.Object.ID].SetPosition(30, 40)
	bounds := physplace.NetBounds(nets[a.ID])
	x, y, w, h := bounds.Coordinates()
	if x != 10 || y != 10 || w != 20 || h != 30 {
		t.Fatalf("updated net bounds = (%v,%v,%v,%v), want (10,10,20,30)", x, y, w, h)
	}
}
