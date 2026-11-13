package acec

// Top-level arithmetic equivalence flow from src/proof/acec/acecCore.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec"
	"time"
)

func Acec_ComputeEquivClasses(a, b *Gia_Man_t, mapA, mapB **Vec_Int_t) {
	acecComputeEquivClasses(a, b, mapA, mapB, cec.Gia_ManComputeGiaEquivs)
}

// acecComputeEquivClasses allows the SAT sweep to be exercised in tests.
func acecComputeEquivClasses(a, b *Gia_Man_t, mapA, mapB **Vec_Int_t, compute func(*gia.Gia_Man_t, int32, int32) *gia.Gia_Man_t) bool {
	start := time.Now()
	base := Acec_CommonStart(nil, a)
	defer gia.Gia_ManStop(base)
	Acec_CommonStart(base, b)
	Acec_CommonFinish(base)
	repr := compute(base, 100, 0)
	*mapA = Acec_CountRemap(a, base)
	*mapB = Acec_CountRemap(b, base)
	if repr == nil {
		panic("acec: equivalence sweep returned nil")
	}
	defer gia.Gia_ManStop(repr)
	fmt.Printf("Finished computing equivalent nodes. Time = %v\n", time.Since(start))
	return true
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
		for i := int32(1); i <= 2; i++ {
			gia.Gia_ManPatchCoDriver(left, gia.Gia_ManCoNum(left)-i, 0)
			gia.Gia_ManPatchCoDriver(right, gia.Gia_ManCoNum(right)-i, 0)
		}
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

// Replay the miter's CI assignment on both original graphs before publishing
// it, and use the first output that actually differs as the failure index.
func acecRecordOriginalWitness(a, b, miter *Gia_Man_t) int32 {
	witness := miter.CombinationalCounterexample()
	if witness == nil || int(witness.NPis) != len(a.InputLiterals()) || len(b.InputLiterals()) != len(a.InputLiterals()) || int64(len(witness.PData))*32 < int64(witness.NPis) {
		return -1
	}
	inputs := make([]bool, witness.NPis)
	for i := range inputs {
		inputs[i] = witness.PData[i>>5]&(uint32(1)<<uint(i&31)) != 0
	}
	x, err := a.Evaluate(inputs)
	if err != nil {
		return -1
	}
	y, err := b.Evaluate(inputs)
	if err != nil || len(x) != len(y) {
		return -1
	}
	for i := range x {
		if x[i] != y[i] {
			if a.SetCombinationalCounterexample(inputs, i) == nil {
				return 0
			}
			return -1
		}
	}
	return -1
}
