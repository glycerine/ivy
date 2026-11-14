package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

// The simulator is the functional basis for LUT decomposition. This checks
// that a multi-node cone is transformed into the same truth table as its
// composition, with variable order preserved from vLeaves.
func TestAbcNodeSuperChoiceTruthComposesCone(t *testing.T) {
	network := abc.NewNetwork("lut-truth-composition")
	x0 := network.NewObject(abc.ObjectPI)
	x1 := network.NewObject(abc.ObjectPI)
	x2 := network.NewObject(abc.ObjectPI)
	and01 := abc.Abc_NtkCreateNode(network)
	abc.Abc_ObjAddFanin(and01, x0)
	abc.Abc_ObjAddFanin(and01, x1)
	and012 := abc.Abc_NtkCreateNode(network)
	abc.Abc_ObjAddFanin(and012, and01)
	abc.Abc_ObjAddFanin(and012, x2)
	and01.Data = "11 1\n"
	and012.Data = "11 1\n"

	leaves := vec.Vec_PtrAlloc(3)
	volume := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(leaves, x0)
	vec.Vec_PtrPush(leaves, x1)
	vec.Vec_PtrPush(leaves, x2)
	vec.Vec_PtrPush(volume, and01)
	vec.Vec_PtrPush(volume, and012)
	defer vec.Vec_PtrFree(leaves)
	defer vec.Vec_PtrFree(volume)

	manager := Abc_ManSclStart(3, 3, 2)
	defer Abc_ManSclStop(manager)
	manager.vLeaves = leaves
	manager.vVolume = volume
	got := Abc_NodeSuperChoiceTruth(manager)
	const want uint32 = 0x80808080
	if len(got) != 1 || got[0] != want {
		t.Fatalf("three-input AND truth table = %#v, want [%#08x]", got, want)
	}
}
