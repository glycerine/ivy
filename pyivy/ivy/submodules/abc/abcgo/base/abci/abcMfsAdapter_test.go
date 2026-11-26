package abci

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestMfsParamsDefaultAndTranslation(t *testing.T) {
	var p Mfs_Par_t
	Abc_NtkMfsParsDefault(&p)
	if p.nWinTfoLevs != 2 || p.nFanoutsMax != 30 || p.nDepthMax != 20 || p.nWinMax != 300 || p.nGrowthLevel != 0 || p.nBTLimit != 5000 || p.fResub != 1 {
		t.Fatalf("unexpected ABC defaults: %+v", p)
	}
	p.nGrowthLevel, p.nBTLimit = 4, 123
	p.fRrOnly, p.fArea, p.fMoreEffort, p.fSwapEdge = 1, 1, 1, 1
	p.fOneHotness, p.fDelay, p.fPower, p.fGiaSat, p.fVerbose, p.fVeryVerbose = 1, 1, 1, 1, 1, 1
	got := mfsParamsFromABC(&p)
	if got.NWinTfoLevs != 2 || got.NFanoutsMax != 30 || got.NDepthMax != 20 || got.NWinMax != 300 || got.NGrowthLevel != 4 || got.NBTLimit != 123 {
		t.Fatalf("integer parameter translation: %+v", got)
	}
	if got.FRrOnly != 1 || got.FResub != 1 || got.FArea != 1 || got.FMoreEffort != 1 || got.FSwapEdge != 1 || got.FOneHotness != 1 || got.FDelay != 1 || got.FPower != 1 || got.FGiaSat != 1 || got.FVerbose != 1 || got.FVeryVerbose != 1 {
		t.Fatalf("flag parameter translation: %+v", got)
	}
}

// mfsEvaluate returns the CO values of the logic network for every CI
// assignment (CIs in network order), computed from its strashed AIG.
func mfsEvaluate(t *testing.T, n *Abc_Ntk_t) [][]bool {
	t.Helper()
	strashed := Abc_NtkStrash(n, 0, 1, 0)
	g := abc.Abc_NtkToDar(strashed, 0, 0)
	nCis := g.InputCount()
	var out [][]bool
	for m := 0; m < 1<<nCis; m++ {
		inputs := make([]bool, nCis)
		for i := range inputs {
			inputs[i] = m&(1<<i) != 0
		}
		values, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, values)
	}
	return out
}

func mfsRedundantNetwork() *Abc_Ntk_t {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	first := abc.Abc_NtkCreateNode(n)
	for _, input := range []*abc.Abc_Obj_t{a, b, c} {
		abc.Abc_ObjAddFanin(first, input)
	}
	first.Data = abc.Abc_SopCreateAnd(nil, 3, []int32{0, 0, 0})
	second := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(second, first)
	abc.Abc_ObjAddFanin(second, a)
	second.Data = abc.Abc_SopCreateOr(nil, 2, []int32{0, 0})
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, second)
	abc.Abc_NtkAddDummyPiNames(n)
	abc.Abc_NtkAddDummyPoNames(n)
	return n
}

// OR(AND(a,b,c), a) == a: resubstitution removes the AND fanin.
func TestMfsRemovesRedundantFaninAndPreservesFunction(t *testing.T) {
	for _, resub := range []int32{1, 0} {
		n := mfsRedundantNetwork()
		before := mfsEvaluate(t, n)
		edgesBefore := abc.Abc_NtkGetTotalFanins(n)
		var pars Mfs_Par_t
		Abc_NtkMfsParsDefault(&pars)
		pars.fResub = resub
		if !Abc_NtkMfs(n, &pars) {
			t.Fatal("Abc_NtkMfs returned failure")
		}
		if abc.Abc_NtkCheck(n) == 0 {
			t.Fatal("network check failed after mfs")
		}
		if got := mfsEvaluate(t, n); !reflect.DeepEqual(got, before) {
			t.Fatalf("resub=%d: mfs changed the function", resub)
		}
		if resub == 1 && abc.Abc_NtkGetTotalFanins(n) >= edgesBefore {
			t.Fatalf("resubstitution did not remove an edge: %d >= %d", abc.Abc_NtkGetTotalFanins(n), edgesBefore)
		}
	}
}

func TestMfsWithExternalCareSetPreservesCareFunction(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	inputs := make([]*abc.Abc_Obj_t, 3)
	for i := range inputs {
		inputs[i] = abc.Abc_NtkCreatePi(n)
	}
	node := abc.Abc_NtkCreateNode(n)
	for _, input := range inputs {
		abc.Abc_ObjAddFanin(node, input)
	}
	node.Data = abc.Abc_SopCreateAnd(nil, 3, []int32{0, 0, 0})
	other := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(other, node)
	abc.Abc_ObjAddFanin(other, inputs[0])
	other.Data = abc.Abc_SopCreateXor(nil, 2)
	po := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(po, other)
	abc.Abc_NtkAddDummyPiNames(n)
	abc.Abc_NtkAddDummyPoNames(n)

	care := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	careInputs := make([]*abc.Abc_Obj_t, 3)
	for i := range careInputs {
		careInputs[i] = abc.Abc_NtkCreatePi(care)
	}
	careNode := abc.Abc_NtkCreateNode(care)
	for _, input := range careInputs {
		abc.Abc_ObjAddFanin(careNode, input)
	}
	careNode.Data = abc.Abc_SopCreateOr(nil, 3, []int32{0, 0, 0})
	carePO := abc.Abc_NtkCreatePo(care)
	abc.Abc_ObjAddFanin(carePO, careNode)
	abc.Abc_NtkAddDummyPiNames(care)
	abc.Abc_NtkAddDummyPoNames(care)
	n.Excare = care

	before := mfsEvaluate(t, n)
	for _, resub := range []int32{1, 0} {
		var pars Mfs_Par_t
		Abc_NtkMfsParsDefault(&pars)
		pars.fResub = resub
		if !Abc_NtkMfs(n, &pars) {
			t.Fatal("Abc_NtkMfs returned failure")
		}
		after := mfsEvaluate(t, n)
		for m := range before {
			if m == 0 { // outside the care set (OR of the inputs)
				continue
			}
			if !reflect.DeepEqual(after[m], before[m]) {
				t.Fatalf("resub=%d: care assignment %03b changed outputs", resub, m)
			}
		}
	}
}
