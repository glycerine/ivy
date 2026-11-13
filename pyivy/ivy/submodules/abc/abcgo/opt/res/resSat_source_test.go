package res

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func TestSourceProofPathAssertsAtDisabledClauseDuplication(t *testing.T) {
	n := abc.NewNetwork("proof")
	n.Type, n.Function = uint8(abc.ABC_NTK_STRASH), uint8(abc.ABC_FUNC_AIG)
	m := abc.Abc_AigAlloc(n)
	n.FunctionManager = m
	x, y := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	and := abc.Abc_AigAnd(m, x.Signal(), y.Signal())
	care, onset := n.NewObject(abc.ObjectPO), n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(care, x.Signal())
	abc.Abc_ObjAddFanin(onset, and)
	roots := Vec_PtrAlloc(2)
	Vec_PtrPush(roots, care)
	Vec_PtrPush(roots, onset)
	defer func() {
		if recover() == nil {
			t.Fatal("the checked-in C clause duplication path asserts")
		}
	}()
	Res_SatProveUnsat(n, roots)
}

func TestSourceSatSimulationEnumeratesOnsetPatterns(t *testing.T) {
	n := abc.NewNetwork("simulate")
	n.Type, n.Function = uint8(abc.ABC_NTK_STRASH), uint8(abc.ABC_FUNC_AIG)
	m := abc.Abc_AigAlloc(n)
	n.FunctionManager = m
	x := n.NewObject(abc.ObjectPI)
	n.NewObject(abc.ObjectPI)
	care, onset := n.NewObject(abc.ObjectPO), n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(care, abc.Abc_AigConst1(n))
	abc.Abc_ObjAddFanin(onset, x.Signal())
	sim := Res_SimAlloc(1)
	sim.pAig, sim.nTruePis = n, 2
	if got := Res_SatSimulate(sim, 3, 1); got != 1 || sim.nPats1 != 2 {
		t.Fatalf("onset result=%d patterns=%d", got, sim.nPats1)
	}
	xBits := resWords(Vec_PtrEntryAny(sim.vPats1, 0))[0]
	yBits := resWords(Vec_PtrEntryAny(sim.vPats1, 1))[0]
	if xBits&3 != 3 || yBits&3 != 1 && yBits&3 != 2 {
		t.Fatalf("patterns x=%b y=%b", xBits&3, yBits&3)
	}
}
