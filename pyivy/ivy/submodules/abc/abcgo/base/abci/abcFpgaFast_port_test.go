package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/aig/ivy"
	"github.com/glycerine/abcgo/base/abc"
)

func TestIvyFastMapReconstructsCutFunctionAndComplementedPO(t *testing.T) {
	manager := ivy.Ivy_ManStart()
	a := ivy.Ivy_ObjCreatePi(manager)
	b := ivy.Ivy_ObjCreatePi(manager)
	root := ivy.Ivy_And(manager, a, ivy.Ivy_Not(b))
	ivy.Ivy_ObjCreatePo(manager, root)
	ivy.Ivy_ObjCreatePo(manager, ivy.Ivy_Not(root))
	ivy.Ivy_FastMapPerform(manager, 2, 0, 0)

	source := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	abc.Abc_NtkCreatePi(source)
	abc.Abc_NtkCreatePi(source)
	abc.Abc_NtkCreatePo(source)
	abc.Abc_NtkCreatePo(source)
	mapped := Ivy_ManFpgaToAbc(source, manager)
	ivy.Ivy_FastMapStop(manager)
	ivy.Ivy_ManStop(manager)
	defer abc.Abc_NtkDelete(mapped)

	if got := abc.Abc_NtkCoNum(mapped); got != 2 {
		t.Fatalf("mapped outputs = %d, want 2", got)
	}
	first := abc.Abc_ObjFanin0(abc.Abc_NtkCo(mapped, 0))
	second := abc.Abc_ObjFanin0(abc.Abc_NtkCo(mapped, 1))
	if len(first.Fanins()) != 2 || len(second.Fanins()) != 2 {
		t.Fatalf("mapped node fanins = %d and %d, want 2 each", len(first.Fanins()), len(second.Fanins()))
	}
	one := first.Data.(*hop.Hop_Obj_t)
	two := second.Data.(*hop.Hop_Obj_t)
	truth := hop.Hop_ManConvertAigToTruth(mapped.FunctionManager.(*hop.Hop_Man_t), one, 2, nil, 0)
	truthInv := hop.Hop_ManConvertAigToTruth(mapped.FunctionManager.(*hop.Hop_Man_t), two, 2, nil, 0)
	if truth[0] != 0x22222222 || truthInv[0] != 0xdddddddd {
		t.Fatalf("mapped truth tables = %#x and %#x, want x&!y and its complement", truth[0], truthInv[0])
	}
}
