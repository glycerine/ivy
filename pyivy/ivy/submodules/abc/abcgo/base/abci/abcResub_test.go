package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestResubManagerSimulationStorageLifecycle(t *testing.T) {
	m := Abc_ManResubStart(3, 12)
	if m == nil || len(m.pInfo) != int(m.nWords*(m.nDivsMax+1)) {
		t.Fatal("manager did not allocate contiguous simulation storage")
	}
	if len(m.pCareSet) != int(m.nWords) || m.pCareSet[0] != ^uint32(0) {
		t.Fatal("manager care set is not initialized to all ones")
	}
	for i := int32(0); i < m.nLeavesMax; i++ {
		words := vec.Vec_PtrEntry(m.vSims, int(i)).([]uint32)
		if len(words) != int(m.nWords) {
			t.Fatalf("leaf %d has %d simulation words, want %d", i, len(words), m.nWords)
		}
	}
	Abc_ManResubStop(m)
	if m.pInfo != nil || m.pCareSet != nil || m.vDivs != nil || m.vSims != nil || m.vTemp != nil ||
		m.vDivs1UP != nil || m.vDivs1UN != nil || m.vDivs1B != nil || m.vDivs2UP0 != nil ||
		m.vDivs2UP1 != nil || m.vDivs2UN0 != nil || m.vDivs2UN1 != nil {
		t.Fatal("manager stop retained allocated scratch storage")
	}
}

func TestResubSimulationTruthAndComplementedFanin(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	root := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(root, a)
	abc.Abc_ObjAddFanin(root, b)
	abc.Abc_ObjSetFaninC(root, 0)
	m := Abc_ManResubStart(2, 8)
	defer Abc_ManResubStop(m)
	divs := vec.Vec_PtrAlloc(3)
	defer vec.Vec_PtrFree(divs)
	vec.Vec_PtrPush(divs, a)
	vec.Vec_PtrPush(divs, b)
	vec.Vec_PtrPush(divs, root)
	Abc_ManResubSimulate(divs, 2, m.vSims, m.nLeavesMax, m.nWords)
	words := root.Data.([]uint32)
	if got, want := words[0], uint32(0x4); got != want {
		t.Fatalf("simulation of !a & b = %#08x, want %#08x", got, want)
	}
	if root.Phase {
		t.Fatal("simulation phase should be zero after normalization")
	}
}
