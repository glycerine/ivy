package abc

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestNetworkStatistics(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	buffer := Abc_NtkCreateNodeBuf(n, a)
	gate := Abc_NtkCreateNodeAnd(n, []*Object{buffer, b})
	for range 2 {
		Abc_ObjAddFanin(n.NewObject(ObjectPO), gate)
	}
	if Abc_NtkGetBufNum(n) != 1 || Abc_NtkGetLargeNodeNum(n) != 1 || Abc_NtkGetFaninMax(n) != 2 || Abc_NtkGetFanoutMax(n) != 2 || Abc_NtkGetTotalFanins(n) != 3 {
		t.Fatal("incorrect network statistics")
	}
	gate.Choice = buffer
	if Abc_NtkGetChoiceNum(n) != 0 {
		t.Fatal("non-AIG network reported choices")
	}
	n.Type = uint8(ABC_NTK_STRASH)
	if Abc_NtkGetChoiceNum(n) != 1 {
		t.Fatal("AIG choice was not counted")
	}
}

func TestSOPAndBDDStatistics(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	gate := Abc_NtkCreateNodeAnd(n, []*Object{a, b})
	Abc_NtkCreateNodeConst1(n)
	if Abc_NtkGetCubeNum(n) != 1 || Abc_NtkGetCubePairNum(n) != 0 || Abc_NtkGetLitNum(n) != 2 || Abc_NtkGetLitFactNum(n) < 1 {
		t.Fatal("incorrect SOP statistics")
	}
	gate.Data = "10 1\n01 1\n"
	if Abc_NtkGetCubeNum(n) != 2 || Abc_NtkGetCubePairNum(n) != 1 || Abc_NtkGetLitNum(n) != 4 {
		t.Fatal("incorrect multi-cube statistics")
	}
	if Abc_NtkSopToBdd(n) == 0 {
		t.Fatal("SOP to BDD conversion failed")
	}
	dd := n.FunctionManager.(*core.DdManager)
	if got, want := Abc_NtkGetBddNodeNum(n), int32(core.Cudd_DagSize(gate.Data.(*core.DdNode))-1); got != want || len(dd.Vars) < 2 {
		t.Fatalf("BDD node count %d, want %d", got, want)
	}
	if got := Abc_NtkGetClauseNum(n); got != 5 {
		t.Fatalf("XOR and constant clause count = %d, want 5", got)
	}
}

func TestNetworkAttributeFree(t *testing.T) {
	n := NewNetwork("attributes")
	manager := new(int)
	n.Attributes = []*vec.Vec_Att_t{vec.Vec_AttAlloc(2, manager, nil, nil, nil)}
	if Abc_NtkAttrFree(n, 0, 0) != manager || n.Attributes[0] != nil {
		t.Fatal("attribute manager was not transferred")
	}
	freed := false
	n.Attributes[0] = vec.Vec_AttAlloc(2, manager, func(any) { freed = true }, nil, nil)
	Abc_NtkDelete(n)
	if !freed {
		t.Fatal("network deletion did not release attribute manager")
	}
}

func TestMappedAreaSkipsBarriers(t *testing.T) {
	n, _, _, _, driver, consumer := abcSCLFixture(t)
	if got := Abc_NtkGetMappedArea(n); got != 2 {
		t.Fatalf("mapped area = %v, want 2", got)
	}
	driver.Barrier = true
	if got := Abc_NtkGetMappedArea(n); got != 1 {
		t.Fatalf("mapped area with barrier = %v, want 1", got)
	}
	consumer.Data = nil
	if got := Abc_NtkGetMappedArea(n); got != 0 {
		t.Fatalf("unmapped node should not add area: %v", got)
	}
}
