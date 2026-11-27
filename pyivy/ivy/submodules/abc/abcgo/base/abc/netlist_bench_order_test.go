package abc

import "testing"

// TestAigToLogicSopBenchCreatesInvertersFirst: Abc_NtkAigToLogicSopBench
// creates the constant's and the CIs' inverters before duplicating the AND
// nodes, and each node's inverter right after the node, so object ids (and
// every id-ordered writer) follow ABC.
func TestAigToLogicSopBenchCreatesInvertersFirst(t *testing.T) {
	src := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	a, b := Abc_NtkCreatePi(src), Abc_NtkCreatePi(src)
	n1 := Abc_AigAnd(src.FunctionManager.(*Abc_Aig_t), Abc_ObjNot(a), b)
	n2 := Abc_AigAnd(src.FunctionManager.(*Abc_Aig_t), Abc_ObjNot(n1), a)
	po := Abc_NtkCreatePo(src)
	Abc_ObjAddFanin(po, n2)
	Abc_NtkAddDummyPiNames(src)
	Abc_NtkAddDummyPoNames(src)
	dst := Abc_NtkAigToLogicSopBench(src)
	var kinds []string
	for i := int32(0); i < Abc_NtkObjNumMax(dst); i++ {
		o := Abc_NtkObj(dst, i)
		if o == nil || Abc_ObjIsNode(o) == 0 {
			continue
		}
		if Abc_ObjFaninNum(o) == 1 {
			kinds = append(kinds, "inv")
		} else {
			kinds = append(kinds, "and")
		}
	}
	// inverter of PI a, then n1, then n1's inverter, then n2
	want := []string{"inv", "and", "inv", "and"}
	if len(kinds) != len(want) {
		t.Fatalf("nodes %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("nodes in id order %v, want %v", kinds, want)
		}
	}
}
