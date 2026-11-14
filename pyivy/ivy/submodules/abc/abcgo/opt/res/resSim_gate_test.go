package res

import (
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func TestSourceSimAllFaninPhasesAndOutputs(t *testing.T) {
	n := abc.NewNetwork("phases")
	n.Type, n.Function = uint8(abc.ABC_NTK_STRASH), uint8(abc.ABC_FUNC_AIG)
	mgr := abc.Abc_AigAlloc(n)
	n.FunctionManager = mgr
	a, b := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	for bits := 0; bits < 4; bits++ {
		x, y := a.Signal(), b.Signal()
		if bits&1 != 0 {
			x = x.Not()
		}
		if bits&2 != 0 {
			y = y.Not()
		}
		out := n.NewObject(abc.ObjectPO)
		abc.Abc_ObjAddFanin(out, abc.Abc_AigAnd(mgr, x, y))
	}
	sim := &Res_Sim_t{pAig: n, vPats: Vec_PtrAllocSimInfo(abc.Abc_NtkObjNumMax(n), 1)}
	resWords(Vec_PtrEntryAny(sim.vPats, a.ID))[0] = 0xaaaaaaaa
	resWords(Vec_PtrEntryAny(sim.vPats, b.ID))[0] = 0xcccccccc
	Res_SimPerformRound(sim, 1)
	for bits := 0; bits < 4; bits++ {
		aWord, bWord := uint32(0xaaaaaaaa), uint32(0xcccccccc)
		if bits&1 != 0 {
			aWord = ^aWord
		}
		if bits&2 != 0 {
			bWord = ^bWord
		}
		got := resWords(Vec_PtrEntryAny(sim.vPats, abc.Abc_NtkPo(n, int32(bits)).ID))[0]
		if got != aWord&bWord {
			t.Fatalf("phase %d: got %08x want %08x", bits, got, aWord&bWord)
		}
	}
}
