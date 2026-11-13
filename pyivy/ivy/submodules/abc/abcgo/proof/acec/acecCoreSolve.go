package acec

// Top-level arithmetic equivalence flow from src/proof/acec/acecCore.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec"
	"time"
)

func Acec_ComputeEquivClasses(a, b *Gia_Man_t, mapA, mapB **Vec_Int_t) {
	start := time.Now()
	base := Acec_CommonStart(nil, a)
	Acec_CommonStart(base, b)
	Acec_CommonFinish(base)
	repr := cec.Gia_ManComputeGiaEquivs(base, 100, 0)
	if repr == nil {
		panic("acec: equivalence computation failed")
	}
	*mapA = Acec_CountRemap(a, base)
	*mapB = Acec_CountRemap(b, base)
	gia.Gia_ManStop(base)
	gia.Gia_ManStop(repr)
	fmt.Printf("Finished computing equivalent nodes. Time = %v\n", time.Since(start))
}
func Acec_Solve(a, b *Gia_Man_t, pars *Acec_ParCec_t) int32 {
	start := time.Now()
	left, right := a, b
	boxA, boxB := Acec_ProduceBox(a, pars.fVerbose), Acec_ProduceBox(b, pars.fVerbose)
	if boxA == nil || boxB == nil {
		fmt.Println("Cannot find arithmetic boxes in both LHS and RHS. Trying regular CEC.")
	} else if Acec_MatchBoxes(boxA, boxB) == 0 {
		fmt.Println("Cannot match arithmetic boxes in LHS and RHS. Trying regular CEC.")
	} else {
		left, right = Acec_InsertBox(boxA, 0), Acec_InsertBox(boxB, 0)
		fmt.Printf("Matching of adder trees in LHS and RHS succeeded. Time = %v\n", time.Since(start))
		gia.Gia_ManPatchCoDriver(left, gia.Gia_ManCoNum(left)-1, 0)
		gia.Gia_ManPatchCoDriver(right, gia.Gia_ManCoNum(right)-1, 0)
		gia.Gia_ManPatchCoDriver(left, gia.Gia_ManCoNum(left)-2, 0)
		gia.Gia_ManPatchCoDriver(right, gia.Gia_ManCoNum(right)-2, 0)
	}
	var cecParams cec.Cec_ParCec_t
	cec.Cec_ManCecSetDefaultParams(&cecParams)
	cecParams.SetConflictLimit(pars.nBTLimit)
	miter := gia.Gia_ManMiter(left, right, 0, 1, 0, 0, pars.fVerbose)
	status := int32(-1)
	if miter == nil {
		fmt.Println("Miter computation has failed.")
	} else {
		status = cec.Cec_ManVerify(miter, &cecParams)
		gia.SwapCombinationalCounterexamples(a, miter)
		gia.Gia_ManStop(miter)
	}
	if left != a {
		gia.Gia_ManStop(left)
	}
	if right != b {
		gia.Gia_ManStop(right)
	}
	return status
}
