package abci

import (
	"testing"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRrUpdateRemovesAProvenRedundantFanin(t *testing.T) {
	n := baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG, 1)
	x, y := baseabc.Abc_NtkCreatePi(n), baseabc.Abc_NtkCreatePi(n)
	and := baseabc.Abc_AigAnd(n.FunctionManager.(*baseabc.Abc_Aig_t), x.Signal(), y.Signal())
	po := baseabc.Abc_NtkCreatePo(n)
	if err := n.AddFanin(po, and); err != nil {
		t.Fatal(err)
	}
	if Abc_NtkRRUpdate(n, and.Object, x, nil) != 1 {
		t.Fatal("RR update failed")
	}
	if got := baseabc.Abc_ObjChild0(po); got != y.Signal() {
		t.Fatalf("output driver = %+v, want %v", got, y.Signal())
	}
}

func TestRrWindowCopiesConeAndBoundary(t *testing.T) {
	n := baseabc.Abc_NtkAlloc(baseabc.ABC_NTK_STRASH, baseabc.ABC_FUNC_AIG, 1)
	x, y := baseabc.Abc_NtkCreatePi(n), baseabc.Abc_NtkCreatePi(n)
	and := baseabc.Abc_AigAnd(n.FunctionManager.(*baseabc.Abc_Aig_t), x.Signal(), y.Signal())
	leaves, cone, roots := vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(1), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(leaves, x)
	vec.Vec_PtrPush(leaves, y)
	vec.Vec_PtrPush(cone, and.Object)
	vec.Vec_PtrPush(roots, and.Object)
	window := Abc_NtkWindow(n, leaves, cone, roots)
	if window == nil {
		t.Fatal("window construction failed")
	}
	defer baseabc.Abc_NtkDelete(window)
	if baseabc.Abc_NtkCiNum(window) != 2 || baseabc.Abc_NtkCoNum(window) != 1 || baseabc.Abc_NtkNodeNum(window) != 1 {
		t.Fatalf("window size: CI=%d CO=%d nodes=%d", baseabc.Abc_NtkCiNum(window), baseabc.Abc_NtkCoNum(window), baseabc.Abc_NtkNodeNum(window))
	}
}
