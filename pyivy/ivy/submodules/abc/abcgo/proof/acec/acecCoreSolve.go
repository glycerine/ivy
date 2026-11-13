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
	a.ClearCombinationalCounterexample()
	// Box matching assumes identical interfaces. Reject a mismatch before
	// building arithmetic boxes, just as the combinational miter does.
	if gia.Gia_ManCiNum(a) != gia.Gia_ManCiNum(b) || gia.Gia_ManCoNum(a) != gia.Gia_ManCoNum(b) {
		return -1
	}
	left, right := a, b
	boxA, boxB := Acec_ProduceBox(a, pars.fVerbose), Acec_ProduceBox(b, pars.fVerbose)
	if boxA == nil || boxB == nil {
		fmt.Println("Cannot find arithmetic boxes in both LHS and RHS. Trying regular CEC.")
	} else if Acec_MatchBoxes(boxA, boxB) == 0 {
		fmt.Println("Cannot match arithmetic boxes in LHS and RHS. Trying regular CEC.")
	} else {
		left, right = Acec_InsertBox(boxA, 0), Acec_InsertBox(boxB, 0)
		fmt.Printf("Matching of adder trees in LHS and RHS succeeded. Time = %v\n", time.Since(start))
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
		if left == a && right == b && status == 0 {
			status = acecRecordOriginalWitness(a, b, miter)
		}
		gia.Gia_ManStop(miter)
	}
	// Box normalization can change individual outputs. Its result is useful
	// for arithmetic analysis but cannot certify equivalence of the original
	// interfaces, or supply a witness for them, without a separate proof.
	if left != a || right != b {
		miter = gia.Gia_ManMiter(a, b, 0, 1, 0, 0, pars.fVerbose)
		status = -1
		if miter != nil {
			status = cec.Cec_ManVerify(miter, &cecParams)
			if status == 0 {
				status = acecRecordOriginalWitness(a, b, miter)
			}
			gia.Gia_ManStop(miter)
		}
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
