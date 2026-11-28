// Port of src/proof/live/combination.c: n-choose-k disjunctions of
// registers or hint outputs (Knuth's combination generator) for nck.

package live

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

func countCombination(n, k int64) int64 {
	if n == k {
		return 1
	}
	if k == 1 {
		return n
	}
	return countCombination(n-1, k-1) + countCombination(n-1, k)
}

func listCombination(n, t int32) {
	combCounter := 0

	//Initialization
	vC := make([]int32, 0, t+3)
	for i := int32(-1); i < t; i++ {
		vC = append(vC, i)
	}
	vC = append(vC, n)
	vC = append(vC, 0)

	for {
		//visit combination
		combCounter++
		fmt.Printf("Comb-%3d : ", combCounter)
		for i := t; i > 0; i-- {
			fmt.Printf("vC[%d] = %d ", i, vC[i])
		}
		fmt.Printf("\n")

		j := int32(1)
		for vC[j]+1 == vC[j+1] {
			vC[j] = j - 1
			j = j + 1
		}
		if j > t {
			break
		}
		vC[j] = vC[j] + 1
	}
}

// knuthCombinations visits the combinations of combK out of combN as
// combination.c's loops do, with vC_KNUTH[1..combK] the chosen indices.
func knuthCombinations(combN, combK int32, visit func(vC []int32)) int32 {
	totalCombination_KNUTH := int32(0)

	//Knuth's Data Structure Initialization
	vC_KNUTH := make([]int32, 0, combK+3)
	for i_KNUTH := int32(-1); i_KNUTH < combK; i_KNUTH++ {
		vC_KNUTH = append(vC_KNUTH, i_KNUTH)
	}
	vC_KNUTH = append(vC_KNUTH, combN)
	vC_KNUTH = append(vC_KNUTH, 0)

	for {
		totalCombination_KNUTH++
		visit(vC_KNUTH)

		j_KNUTH := int32(1)
		for vC_KNUTH[j_KNUTH]+1 == vC_KNUTH[j_KNUTH+1] {
			vC_KNUTH[j_KNUTH] = j_KNUTH - 1
			j_KNUTH = j_KNUTH + 1
		}
		if j_KNUTH > combK {
			break
		}
		vC_KNUTH[j_KNUTH] = vC_KNUTH[j_KNUTH] + 1
	}
	return totalCombination_KNUTH
}

func generateCombinatorialStabil(pAigNew, pAigOld *Aig_Man_t,
	vCandidateMonotoneSignals_ []int32,
	vDisj_nCk_ *[]Aig_Obj_t,
	combN, combK int32) int32 {
	return knuthCombinations(combN, combK, func(vC_KNUTH []int32) {
		pObjMonoCand := aig.Aig_Not(aig.Aig_ManConst1(pAigNew))
		for i_KNUTH := combK; i_KNUTH > 0; i_KNUTH-- {
			targetPoIndex := vCandidateMonotoneSignals_[vC_KNUTH[i_KNUTH]]
			pObj := aig.Aig_ObjChild0Copy(pAigOld, aig.Aig_ManCo(pAigOld, targetPoIndex))
			pObjMonoCand = aig.Aig_Or(pAigNew, pObj, pObjMonoCand)
		}
		*vDisj_nCk_ = append(*vDisj_nCk_, pObjMonoCand)
	})
}

func generateCombinatorialStabilExhaust(pAigNew, pAigOld *Aig_Man_t,
	vDisj_nCk_ *[]Aig_Obj_t,
	combN, combK int32) int32 {
	return knuthCombinations(combN, combK, func(vC_KNUTH []int32) {
		pObjMonoCand := aig.Aig_Not(aig.Aig_ManConst1(pAigNew))
		for i_KNUTH := combK; i_KNUTH > 0; i_KNUTH-- {
			targetPoIndex := vC_KNUTH[i_KNUTH]
			pObj := aigData(pAigOld, aig.Aig_ManLo(pAigOld, targetPoIndex))
			pObjMonoCand = aig.Aig_Or(pAigNew, pObj, pObjMonoCand)
		}
		*vDisj_nCk_ = append(*vDisj_nCk_, pObjMonoCand)
	})
}

// nCkDuplicate is the AIG duplication shared by the two testers, with
// lCombinationCount new registers whose next states are the implications
// (also new POs) of the disjunctions that makeDisj creates.
func nCkDuplicate(pAig *Aig_Man_t, combK int32, combN_internal int32, makeDisj func(pNewAig *Aig_Man_t, vDisj_nCk *[]Aig_Obj_t)) *Aig_Man_t {
	var loCopied, loCreated int32

	pNewAig := aig.Aig_ManStart(aig.Aig_ManObjNumMax(pAig))
	pNewAig.Name = fmt.Sprintf("%s_%s", pAig.Name, "nCk")
	pNewAig.Spec = ""

	//Standard Mapping of Constant Nodes
	aigSetData(pAig, aig.Aig_ManConst1(pAig), aig.Aig_ManConst1(pNewAig))

	//Standard AIG PI duplication
	for i := int32(0); i < saigPiNum(pAig); i++ {
		aigSetData(pAig, aig.Aig_ManCi(pAig, i), aig.Aig_ObjCreateCi(pNewAig))
	}

	//Standard AIG LO duplication
	for i := int32(0); i < saigRegNum(pAig); i++ {
		loCopied++
		aigSetData(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ObjCreateCi(pNewAig))
	}

	//nCk LO creation
	lCombinationCount := int64(0)
	for combK_internal := int32(1); combK_internal <= combK; combK_internal++ {
		lCombinationCount += countCombination(int64(combN_internal), int64(combK_internal))
	}
	vLO_nCk := make([]Aig_Obj_t, 0, max(lCombinationCount, 0))
	for longI := int64(0); longI < lCombinationCount; longI++ {
		loCreated++
		vLO_nCk = append(vLO_nCk, aig.Aig_ObjCreateCi(pNewAig))
	}

	//Standard Node duplication
	nodes, _ := aigManNodes(pAig)
	for _, pObj := range nodes {
		aigSetData(pAig, pObj, aig.Aig_And(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj), aig.Aig_ObjChild1Copy(pAig, pObj)))
	}

	//nCk specific logic creation (i.e. nCk number of OR gates)
	vDisj_nCk := make([]Aig_Obj_t, 0, max(lCombinationCount, 0))
	makeDisj(pNewAig, &vDisj_nCk)

	//creation of implication logic
	vPODriver_nCk := make([]Aig_Obj_t, 0, max(lCombinationCount, 0))
	for longI := int64(0); longI < lCombinationCount; longI++ {
		pObjLO_nCk := vLO_nCk[longI]
		pObjDisj_nCk := vDisj_nCk[longI]

		pObj := aig.Aig_Or(pNewAig, aig.Aig_Not(pObjDisj_nCk), pObjLO_nCk)
		vPODriver_nCk = append(vPODriver_nCk, pObj)
	}

	//Standard PO duplication
	for i := int32(0); i < saigPoNum(pAig); i++ {
		pObj := aig.Aig_ManCo(pAig, i)
		aigSetData(pAig, pObj, aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj)))
	}

	//nCk specific PO creation
	for longI := int64(0); longI < lCombinationCount; longI++ {
		aig.Aig_ObjCreateCo(pNewAig, vPODriver_nCk[longI])
	}

	//Standard LI duplication
	for i := int32(0); i < saigRegNum(pAig); i++ {
		aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, saig.Saig_ManLi(pAig, i)))
	}

	//nCk specific LI creation
	for longI := int64(0); longI < lCombinationCount; longI++ {
		aig.Aig_ObjCreateCo(pNewAig, vPODriver_nCk[longI])
	}

	//clean-up, book-keeping
	nRegCount := loCopied + loCreated

	aig.Aig_ManSetRegNum(pNewAig, nRegCount)
	aig.Aig_ManCleanup(pNewAig)

	return pNewAig
}

// generateDisjunctiveTester is not called by any command.
func generateDisjunctiveTester(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, combN, combK int32) *Aig_Man_t {
	var combN_internal int32

	//Collecting target HINT signals
	vCandidateMonotoneSignals := findHintOutputs(pNtk)
	if vCandidateMonotoneSignals == nil {
		fmt.Printf("\nTraget Signal Set is Empty: Duplicating given AIG\n")
		combN_internal = 0
	} else {
		hintSingalBeginningMarker := vCandidateMonotoneSignals[0]
		hintSingalEndMarker := vCandidateMonotoneSignals[len(vCandidateMonotoneSignals)-1]
		combN_internal = hintSingalEndMarker - hintSingalBeginningMarker + 1
	}

	return nCkDuplicate(pAig, combK, combN_internal, func(pNewAig *Aig_Man_t, vDisj_nCk *[]Aig_Obj_t) {
		for combK_internal := int32(1); combK_internal <= combK; combK_internal++ {
			generateCombinatorialStabil(pNewAig, pAig, vCandidateMonotoneSignals,
				vDisj_nCk, combN_internal, combK_internal)
		}
	})
}

func generateGeneralDisjunctiveTester(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, combK int32) *Aig_Man_t {
	combN_internal := aig.Aig_ManRegNum(pAig)

	return nCkDuplicate(pAig, combK, combN_internal, func(pNewAig *Aig_Man_t, vDisj_nCk *[]Aig_Obj_t) {
		for combK_internal := int32(1); combK_internal <= combK; combK_internal++ {
			generateCombinatorialStabilExhaust(pNewAig, pAig,
				vDisj_nCk, combN_internal, combK_internal)
		}
	})
}
