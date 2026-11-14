package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestNodeConeTruthTracksAIGPhases(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	ab := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(ab, a)
	abc.Abc_ObjAddFanin(ab, b)
	nab := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(nab, a)
	abc.Abc_ObjAddFanin(nab, b)
	abc.Abc_ObjSetFaninC(nab, 0)
	abc.Abc_ObjSetFaninC(nab, 1)
	root := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(root, ab)
	abc.Abc_ObjAddFanin(root, nab)
	abc.Abc_ObjSetFaninC(root, 0)
	abc.Abc_ObjSetFaninC(root, 1)

	manager := Abc_NtkManRefStart(2, 2, 0, 0)
	defer Abc_NtkManRefStop(manager)
	leaves, visited := vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(8)
	defer vec.Vec_PtrFree(leaves)
	defer vec.Vec_PtrFree(visited)
	vec.Vec_PtrPush(leaves, a)
	vec.Vec_PtrPush(leaves, b)
	root.MarkA, ab.MarkA = true, true
	truth := Abc_NodeConeTruth(manager.vVars, manager.vFuncs, int32(refTruthWords(2)), root, leaves, visited)
	if len(truth) != 1 || truth[0] != 0x6666666666666666 {
		t.Fatalf("XOR truth = %#v, want 0x6666666666666666", truth)
	}
	if !root.MarkA || !ab.MarkA {
		t.Fatal("cone truth collection changed preexisting MarkA state")
	}
}

func TestKitToDecRefactorGraphPreservesTruth(t *testing.T) {
	kitGraph := kit.Kit_TruthToGraph([]uint32{0x88888888}, 3, nil)
	defer kit.Kit_GraphFree(kitGraph)
	graph := refKitToDec(kitGraph)
	defer dec.Dec_GraphFree(graph)
	if got := dec.Dec_GraphDeriveTruth(graph); got != 0x88888888 {
		t.Fatalf("converted decomposition truth = %#08x, want %#08x", got, uint32(0x88888888))
	}
}

func TestRefactorManagerStopReleasesScratch(t *testing.T) {
	manager := Abc_NtkManRefStart(8, 12, 1, 0)
	if manager.vVars == nil || manager.vFuncs == nil || manager.vVisited == nil || manager.vCube == nil {
		t.Fatal("manager did not initialize scratch storage")
	}
	Abc_NtkManRefStop(manager)
	if manager.vVars != nil || manager.vFuncs != nil || manager.vVisited != nil || manager.vCube != nil || manager.vMemory != nil {
		t.Fatal("manager stop retained scratch storage")
	}
}
