package abci

import (
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/proof/fraig"
)

func TestCecFraigProvesAssociativeAnd(t *testing.T) {
	makeNetwork := func(left bool) *abc.Network {
		n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
		var inputs [3]abc.Edge
		for i, name := range []string{"a", "b", "c"} {
			pi := abc.Abc_NtkCreatePi(n)
			pi.SetName(name)
			inputs[i] = pi.Signal()
		}
		a, b, c := inputs[0], inputs[1], inputs[2]
		if !left {
			a, c = c, a
		}
		root := n.AndEdges(n.AndEdges(a, b), c)
		po := abc.Abc_NtkCreatePo(n)
		po.SetName("out")
		abc.Abc_ObjAddFanin(po, root)
		return n
	}
	left, right := makeNetwork(true), makeNetwork(false)
	defer abc.Abc_NtkDelete(left)
	defer abc.Abc_NtkDelete(right)
	got := capturePrintStdout(t, func() { Abc_NtkCecFraig(left, right, 20, 0) })
	if !strings.Contains(got, "Networks are equivalent.") {
		t.Fatal(got)
	}
}

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

func TestVerifyPatternSimulatesGatesAfterDeletedIDs(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	defer abc.Abc_NtkDelete(n)
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	out := abc.Abc_NtkCreatePo(n)
	unused := n.AndEdges(a.Signal().Not(), b.Signal())
	abc.Abc_AigDeleteNode(n.FunctionManager.(*abc.Abc_Aig_t), unused.Object)
	abc.Abc_ObjAddFanin(out, n.AndEdges(a.Signal(), b.Signal()))
	for _, pattern := range [][]int32{{1, 1}, {1, 0}} {
		got := Abc_NtkVerifySimulatePatternSlice(n, pattern)
		if len(got) != 1 || got[0] != pattern[0]&pattern[1] {
			t.Fatalf("AND at last ID on %v returned %v", pattern, got)
		}
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

func TestVerifyCleanModelClearsStaleZeroInputModel(t *testing.T) {
	n := truthNetwork(t, 0, 1)
	defer abc.Abc_NtkDelete(n)
	n.Model = []int32{1}
	if got := Abc_NtkVerifyGetCleanModel(n, 1); got != nil || n.Model != nil {
		t.Fatalf("stale model retained: %v, %v", got, n.Model)
	}
}

func BenchmarkI10NativeResynthesisCEC(b *testing.B) {
	extra.Extra_ProgressBarSetMode(false, true)
	defer extra.Extra_ProgressBarSetMode(true, false)
	left, err := abcio.ReadAiger("../../../i10.aig", true)
	if err != nil {
		b.Fatal(err)
	}
	defer abc.Abc_NtkDelete(left)
	right, err := abcio.ReadAiger("test_vectors/i10_native_demo.aig", true)
	if err != nil {
		b.Fatal(err)
	}
	defer abc.Abc_NtkDelete(right)
	b.ReportAllocs()
	var nodes int32
	for b.Loop() {
		miter := Abc_NtkMiter(left, right, 1, 0, 0, 0)
		if miter == nil || Abc_NtkMiterIsConstant(miter) != -1 {
			b.Fatal("expected a structurally nontrivial miter")
		}
		nodes = abc.Abc_NtkNodeNum(miter)
		var params Prove_Params_t
		Prove_ParamsSetDefault(&params)
		params.nItersMax = 5
		result := Abc_NtkIvyProve(&miter, &params)
		abc.Abc_NtkDelete(miter)
		if result != 1 {
			b.Fatalf("expected equivalence, prover returned %d", result)
		}
	}
	b.ReportMetric(float64(nodes), "miter-ANDs/op")
}
