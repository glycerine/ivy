package mpm

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestMappingManagerInitializesCutsAndTruths(t *testing.T) {
	for _, truth := range []int32{0, 1} {
		mig := Mig_ManStart()
		Mig_ManAppendCi(mig)
		pars := &Mpm_Par_t{PLib: Mpm_LibLutSetSimple(6), NNumCuts: 8, FUseTruth: truth}
		p := Mpm_ManStart(mig, pars)
		if mig.pMan != p || vec.Vec_PtrSize(&p.vFreeUnits) != 9 || vec.Vec_IntSize(&p.vCutBests) != int(mig.nObjs) {
			t.Fatalf("truth=%d manager storage not initialized", truth)
		}
		for _, unit := range vec.Vec_PtrArray(&p.vFreeUnits) {
			if len(unit.(*Mpm_Uni_t).pCut.pLeaves) != MPM_VAR_MAX {
				t.Fatal("cut unit leaf storage")
			}
		}
		if truth != 0 {
			if p.funcCst0 != 0 || p.funcVar0 != 1 || Mpm_CutTruth(p, p.funcVar0)[0] != 0xaaaaaaaaaaaaaaaa {
				t.Fatalf("unexpected initial truth table ids: %d %d", p.funcCst0, p.funcVar0)
			}
		} else if p.vTtMem != nil {
			t.Fatal("unexpected truth memory")
		}
		Mpm_ManStop(p)
		Mig_ManStop(mig)
	}
}
