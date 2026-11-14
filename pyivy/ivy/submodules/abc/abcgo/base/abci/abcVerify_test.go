package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/proof/fraig"
)

func TestVerifyCleanModelReturnsZeroedCIsAcrossFrames(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0)
	defer abc.Abc_NtkDelete(n)
	abc.Abc_NtkCreatePi(n)
	abc.Abc_NtkCreatePi(n)
	model := Abc_NtkVerifyGetCleanModel(n, 3)
	if model == nil {
		t.Fatal("clean model is nil")
	}
	got := model
	if len(got) != 6 {
		t.Fatalf("model length = %d, want 6", len(got))
	}
	for i, value := range got {
		if value != 0 {
			t.Errorf("model[%d]=%d, want 0", i, value)
		}
	}
}

func TestVerifyValidCexReturnsCStyleBoolean(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0)
	defer abc.Abc_NtkDelete(n)
	abc.Abc_NtkCreatePi(n)
	if got := Abc_NtkIsValidCex(n, &util.Abc_Cex_t{NPis: 1}); got != 1 {
		t.Fatalf("matching PI count returned %d, want 1", got)
	}
	if got := Abc_NtkIsValidCex(n, &util.Abc_Cex_t{NPis: 2}); got != 0 {
		t.Fatalf("mismatched PI count returned %d, want 0", got)
	}
}

func TestVerifyFraigAdapterPreservesComplementedConstantOutput(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, abc.Abc_AigConst1(n).Not())
	var params fraig.Fraig_Params_t
	fraig.Fraig_ParamsSetDefault(&params)
	manager := abcVerifyNtkToFraig(n, &params)
	if manager == nil {
		t.Fatal("FRAIG conversion returned nil")
	}
	defer fraig.Fraig_ManFree(manager)
	fraig.Fraig_ManProveMiter(manager)
	if got := fraig.Fraig_ManCheckMiter(manager); got != 1 {
		t.Fatalf("constant-false miter status=%d, want equivalent (1)", got)
	}
}

func TestVerifySequentialModelMatchesCFrameAndLatchOrder(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	defer abc.Abc_NtkDelete(n)
	pi := abc.Abc_NtkCreatePi(n)
	latchOut := abc.Abc_NtkAddLatch(n, pi, abc.ABC_INIT_ONE)
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, latchOut)
	model := []int32{0, 1}
	values := abcVerifySeqValues(n, &model[0], 2)
	if got := values[po]; len(got) != 2 || !got[0] || got[1] {
		t.Fatalf("PO values=%v, want [true false]", got)
	}
	if got := values[latchOut]; len(got) != 2 || !got[0] || got[1] {
		t.Fatalf("latch state=%v, want [true false]", got)
	}
}
