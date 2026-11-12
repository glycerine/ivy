package abc

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestNetworkOrderingAndCollections(t *testing.T) {
	n := NewNetwork("ordering")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	first := n.AndEdges(a.Signal(), b.Signal()).Object
	second := n.AndEdges(first.Signal(), a.Signal()).Object
	out := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(out, second)
	v := vec.Vec_PtrAlloc(3)
	for _, o := range []*Object{a, second, first, second} {
		Abc_VecObjPushUniqueOrderByLevel(v, o)
	}
	if got := vec.Vec_PtrArray(v); !reflect.DeepEqual(got, []any{second, first, a}) {
		t.Fatalf("ordered objects = %v", got)
	}
	if Abc_NodeCompareLevelsIncrease(&first, &second) >= 0 || Abc_NodeCompareLevelsDecrease(&first, &second) <= 0 {
		t.Fatal("incorrect level comparisons")
	}
	counts := vec.Vec_IntArray(Abc_NtkFanoutCounts(n))
	if counts[n.constant.ID] != -1 || counts[a.ID] != 2 || counts[first.ID] != 1 || counts[out.ID] != -1 {
		t.Fatalf("fanout counts = %v", counts)
	}
	if ids := vec.Vec_IntArray(Abc_NtkGetCiIds(n)); !reflect.DeepEqual(ids, []int32{int32(a.ID), int32(b.ID)}) {
		t.Fatalf("CI IDs = %v", ids)
	}
	if vec.Vec_PtrSize(Abc_NtkCollectObjects(n)) != len(n.Objects()) {
		t.Fatal("object collection has wrong size")
	}
}

func TestReassignIDsPreservesNamesAttributesAndHash(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	a := n.NewObject(ObjectPI)
	a.SetName("a")
	removed := n.NewObject(ObjectNode)
	if err := n.DeleteObject(removed); err != nil {
		t.Fatal(err)
	}
	removed = n.NewObject(ObjectNode)
	if err := n.DeleteObject(removed); err != nil {
		t.Fatal(err)
	}
	b := n.NewObject(ObjectPI)
	b.SetName("b")
	gate := n.AndEdges(a.Signal(), b.Signal()).Object
	gate.SetName("gate")
	out := n.NewObject(ObjectPO)
	out.SetName("out")
	Abc_ObjAddFanin(out, gate)
	oldID := gate.ID
	n.NameIDs = make([]int32, len(n.objects))
	n.NameIDs[oldID] = 42
	attribute := vec.Vec_AttAlloc(len(n.objects), nil, nil, nil, nil)
	vec.Vec_AttWriteEntry(attribute, oldID, "node attribute")
	n.Attributes = []*vec.Vec_Att_t{attribute}
	Abc_NtkReassignIds(n)
	if gate.ID == oldID || n.Object(gate.ID) != gate || gate.Name() != "gate" || out.Name() != "out" {
		t.Fatal("IDs or names were not reassigned correctly")
	}
	if n.NameIDs[gate.ID] != 42 || vec.Vec_AttEntry(attribute, gate.ID) != "node attribute" {
		t.Fatal("ID-indexed metadata did not follow its object")
	}
	if got := n.AndEdges(a.Signal(), b.Signal()); got.Object != gate || got.Complement {
		t.Fatal("structural hash was not restored")
	}
	if err := n.Check(); err != nil {
		t.Fatalf("invalid network: %v", err)
	}
}

func TestTransferPhases(t *testing.T) {
	src, dst := NewNetwork("source"), NewNetwork("destination")
	input := src.NewObject(ObjectPI)
	copy := dst.NewObject(ObjectPI)
	setCopy(input, copy.Signal().Not())
	src.Phases = make([]int32, len(src.objects))
	src.Phases[input.ID] = 7
	Abc_NtkTransferPhases(dst, src)
	if dst.Phases[copy.ID] != 7 {
		t.Fatal("phase did not follow copy")
	}
}
