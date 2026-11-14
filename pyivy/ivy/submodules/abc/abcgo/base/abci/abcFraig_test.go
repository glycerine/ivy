package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/proof/fraig"
)

func fraigTestNetwork(t *testing.T) (*abc.Network, *abc.Object, *abc.Object) {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	x, y := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	x.SetName("x")
	y.SetName("y")
	po := abc.Abc_NtkCreatePo(n)
	po.SetName("f")
	m := n.FunctionManager.(*abc.Abc_Aig_t)
	root := abc.Abc_AigAnd(m, x.Signal(), y.Signal())
	if err := n.AddFanin(po, root.Not()); err != nil {
		t.Fatal(err)
	}
	return n, x, y
}

func TestFraigRoundTripKeepsComplementedOutput(t *testing.T) {
	n, _, _ := fraigTestNetwork(t)
	defer abc.Abc_NtkDelete(n)
	p := &fraig.Fraig_Params_t{}
	fraig.Fraig_ParamsSetDefault(p)
	manager := Abc_NtkToFraig(n, p, 0, 0).(*fraig.Fraig_Man_t)
	defer fraig.Fraig_ManFree(manager)
	if fraig.Fraig_ManReadOutputNum(manager) != 1 {
		t.Fatalf("FRAIG output count = %d, want 1", fraig.Fraig_ManReadOutputNum(manager))
	}
	gotNoChoices := Abc_NtkFromFraig2(manager, n)
	if gotNoChoices == nil {
		t.Fatal("no-choice reconstruction returned nil")
	}
	defer abc.Abc_NtkDelete(gotNoChoices)
	if !gotNoChoices.Outputs()[0].Fanins()[0].Complement {
		t.Fatal("no-choice reconstruction lost output polarity")
	}
	got := Abc_NtkFromFraig(manager, n)
	if got == nil {
		t.Fatal("FRAIG reconstruction returned nil")
	}
	defer abc.Abc_NtkDelete(got)
	if len(got.Outputs()) != 1 || len(got.Outputs()[0].Fanins()) != 1 {
		t.Fatal("reconstructed network has malformed output")
	}
	if !got.Outputs()[0].Fanins()[0].Complement {
		t.Fatal("reconstructed output lost its complement")
	}
}

func TestFraigTopLevelChecksReconstruction(t *testing.T) {
	n, _, _ := fraigTestNetwork(t)
	defer abc.Abc_NtkDelete(n)
	p := &fraig.Fraig_Params_t{}
	fraig.Fraig_ParamsSetDefault(p)
	p.FTryProve = 0
	got := Abc_NtkFraig(n, p, 0, 0)
	if got == nil {
		t.Fatal("top-level FRAIG rejected the reconstructed AIG")
	}
	defer abc.Abc_NtkDelete(got)
	if len(got.Outputs()) != 1 || len(got.Outputs()[0].Fanins()) != 1 || !got.Outputs()[0].Fanins()[0].Complement {
		t.Fatal("top-level FRAIG changed the output function")
	}
}

func TestFraigTrustCheckRejectsUnsupportedSOP(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	defer abc.Abc_NtkDelete(n)
	x, y, node := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreateNode(n)
	_ = n.AddFanin(node, x.Signal())
	_ = n.AddFanin(node, y.Signal())
	node.Data = "10 1\n01 1\n"
	if Abc_NtkFraigTrustCheck(n) != 0 {
		t.Fatal("trust check accepted a non-OR SOP")
	}
}

func TestFraigTrustPreservesOrChoices(t *testing.T) {
	source := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	defer abc.Abc_NtkDelete(source)
	xs := []*abc.Object{abc.Abc_NtkCreatePi(source), abc.Abc_NtkCreatePi(source), abc.Abc_NtkCreatePi(source)}
	node := abc.Abc_NtkCreateNode(source)
	for _, x := range xs {
		if err := source.AddFanin(node, x.Signal()); err != nil {
			t.Fatal(err)
		}
	}
	node.Data = abc.Abc_SopCreateOrMultiCube(nil, 3, nil)
	if Abc_NtkFraigTrustCheck(source) == 0 {
		t.Fatal("trust check rejected source OR-choice SOP")
	}

	dest := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	defer abc.Abc_NtkDelete(dest)
	for i, x := range xs {
		x.CopyEdge = abc.Abc_NtkCreatePi(dest).Signal()
		_ = i
	}
	edge := Abc_NodeFraigTrust(dest, node)
	if edge.Object != dest.Inputs()[0] || edge.Complement {
		t.Fatal("trust reconstruction did not preserve the OR root as the first choice")
	}
	if edge.Object.Choice != dest.Inputs()[1] || edge.Object.Choice.Choice != dest.Inputs()[2] {
		t.Fatal("trust reconstruction did not link all alternative choices")
	}
}
