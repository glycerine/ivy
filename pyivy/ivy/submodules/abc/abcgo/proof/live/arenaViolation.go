/**CFile****************************************************************

  FileName    [arenaViolation.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [module for addition of arena violator detector
        induced by stabilizing constraints]

  Author      [Sayak Ray]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 31, 2012.]

***********************************************************************/

package live

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	abc "github.com/glycerine/abcgo/base/abc"
)

//#define DISJUNCTIVE_CONSTRAINT_ENABLE_MODE
// #define BARRIER_MONOTONE_TEST

func createArenaLO(pAigNew *Aig_Man_t, vBarriers [][]int32) []Aig_Obj_t {
	if vBarriers == nil {
		return nil
	}
	barrierCount := len(vBarriers)
	if barrierCount <= 0 {
		return nil
	}
	vArenaLO := make([]Aig_Obj_t, 0, barrierCount)
	for i := 0; i < barrierCount; i++ {
		vArenaLO = append(vArenaLO, aig.Aig_ObjCreateCi(pAigNew))
	}
	return vArenaLO
}

func createArenaLi(pAigNew *Aig_Man_t, vBarriers [][]int32, vArenaSignal []Aig_Obj_t) []Aig_Obj_t {
	if vBarriers == nil {
		return nil
	}
	barrierCount := len(vBarriers)
	if barrierCount <= 0 {
		return nil
	}
	vArenaLi := make([]Aig_Obj_t, 0, barrierCount)
	for i := 0; i < barrierCount; i++ {
		pObjDriver := vArenaSignal[i]
		vArenaLi = append(vArenaLi, aig.Aig_ObjCreateCo(pAigNew, pObjDriver))
	}
	return vArenaLi
}

func createMonotoneBarrierLO(pAigNew *Aig_Man_t, vBarriers [][]int32) []Aig_Obj_t {
	if vBarriers == nil {
		return nil
	}
	barrierCount := len(vBarriers)
	if barrierCount <= 0 {
		return nil
	}
	vMonotoneLO := make([]Aig_Obj_t, 0, barrierCount)
	for i := 0; i < barrierCount; i++ {
		vMonotoneLO = append(vMonotoneLO, aig.Aig_ObjCreateCi(pAigNew))
	}
	return vMonotoneLO
}

// driverToPoNew is the image of a CO's driver of pAig in the new manager.
func driverToPoNew(pAig *Aig_Man_t, pObjPo Aig_Obj_t) Aig_Obj_t {
	poDriverOld := driverOld(pAig, pObjPo)
	return copyOf(pAig, poDriverOld)
}

func collectBarrierDisjunctions(pAigOld, pAigNew *Aig_Man_t, vBarriers [][]int32) []Aig_Obj_t {
	if vBarriers == nil {
		return nil
	}
	barrierCount := len(vBarriers)
	if barrierCount <= 0 {
		return nil
	}

	vNewBarrierSignals := make([]Aig_Obj_t, 0, barrierCount)

	for i := 0; i < barrierCount; i++ {
		vIthBarrier := vBarriers[i]
		pObjBarrier := aig.Aig_Not(aig.Aig_ManConst1(pAigNew))
		for _, jElem := range vIthBarrier {
			pObjTargetPoOld := aig.Aig_ManCo(pAigOld, jElem)
			pObjCurr := driverToPoNew(pAigOld, pObjTargetPoOld)
			pObjBarrier = aig.Aig_Or(pAigNew, pObjCurr, pObjBarrier)
		}
		vNewBarrierSignals = append(vNewBarrierSignals, pObjBarrier)
	}

	return vNewBarrierSignals
}

// Aig_Xor of arenaViolation.c (no trivial cases); GCC evaluates the
// arguments of Aig_Or right to left.
func Aig_Xor(pAig *Aig_Man_t, pObj1, pObj2 Aig_Obj_t) Aig_Obj_t {
	n1 := aig.Aig_And(pAig, aig.Aig_Not(pObj1), pObj2)
	n0 := aig.Aig_And(pAig, pObj1, aig.Aig_Not(pObj2))
	return aig.Aig_Or(pAig, n0, n1)
}

func createArenaViolation(
	pAigOld *Aig_Man_t,
	pAigNew *Aig_Man_t,
	pWindowBegins Aig_Obj_t,
	pWithinWindow Aig_Obj_t,
	vMasterBarriers [][]int32,
	vBarrierLo []Aig_Obj_t,
	vBarrierLiDriver *[]Aig_Obj_t,
	vMonotoneDisjunctionNodes *[]Aig_Obj_t,
) Aig_Obj_t {
	pWindowBeginsLocal := pWindowBegins
	pWithinWindowLocal := pWithinWindow

	pObjArenaViolation := aig.Aig_Not(aig.Aig_ManConst1(pAigNew))

	vBarrierSignals := collectBarrierDisjunctions(pAigOld, pAigNew, vMasterBarriers)

	*vMonotoneDisjunctionNodes = append(*vMonotoneDisjunctionNodes, vBarrierSignals...)

	for i, pObj := range vBarrierSignals {
		pObjAnd1 := aig.Aig_And(pAigNew, pObj, pWindowBeginsLocal)
		pObjBarrierLo := vBarrierLo[i]
		pObjOr1 := aig.Aig_Or(pAigNew, pObjAnd1, pObjBarrierLo)
		*vBarrierLiDriver = append(*vBarrierLiDriver, pObjOr1)

		pObjBarrierSwitch := Aig_Xor(pAigNew, pObj, pObjBarrierLo)
		pObjAnd2 := aig.Aig_And(pAigNew, pObjBarrierSwitch, pWithinWindowLocal)
		pObjArenaViolation = aig.Aig_Or(pAigNew, pObjAnd2, pObjArenaViolation)
	}

	return pObjArenaViolation
}

func createConstrained0LiveConeWithDSC(pNewAig *Aig_Man_t, pOldAig *Aig_Man_t, signalList []Aig_Obj_t) Aig_Obj_t {
	numSigAntecedent := len(signalList) - 1

	pAntecedent := aig.Aig_ManConst1(pNewAig)
	pConsequent := signalList[numSigAntecedent]
	pConsequentCopy := aig.Aig_NotCond(aigData(pOldAig, pConsequent), aig.Aig_IsComplement(pConsequent))

	for i := 0; i < numSigAntecedent; i++ {
		pObj := signalList[i]
		pAntecedent = aig.Aig_And(pNewAig, pAntecedent, aig.Aig_NotCond(aigData(pOldAig, pObj), aig.Aig_IsComplement(pObj)))
	}

	p0LiveCone := aig.Aig_Or(pNewAig, aig.Aig_Not(pAntecedent), pConsequentCopy)

	return p0LiveCone
}

func collectCSSignalsWithDSC(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) []Aig_Obj_t {
	return collectCSSignals(pNtk, pAig)
}

// the index of the first PO whose name contains substr, or -1
func findPoContaining(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, substr string) int32 {
	for i := int32(0); i < saigPoNum(pAig); i++ {
		if strings.Contains(abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)), substr) {
			return i
		}
	}
	return -1
}

func collectWindowBeginSignalWithDSC(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) int32 {
	return findPoContaining(pNtk, pAig, "windowBegins_")
}

func collectWithinWindowSignalWithDSC(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) int32 {
	return findPoContaining(pNtk, pAig, "withinWindow_")
}

func collectPendingSignalWithDSC(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) int32 {
	return findPoContaining(pNtk, pAig, "pendingSignal")
}

func createAndGateForMonotonicityVerification(
	pNewAig *Aig_Man_t,
	vDisjunctionSignals []Aig_Obj_t,
	vDisjunctionLo []Aig_Obj_t,
	pendingLo Aig_Obj_t,
	pendingSignal Aig_Obj_t,
) Aig_Obj_t {
	pObjBigAnd := aig.Aig_ManConst1(pNewAig)
	pObjPendingAndPendingLo := aig.Aig_And(pNewAig, pendingLo, pendingSignal)
	for i, pObj := range vDisjunctionSignals {
		pObjLo := vDisjunctionLo[i]
		pObjImply := aig.Aig_Or(pNewAig, aig.Aig_Not(aig.Aig_And(pNewAig, pObjPendingAndPendingLo, pObjLo)), pObj)
		pObjBigAnd = aig.Aig_And(pNewAig, pObjBigAnd, pObjImply)
	}
	return pObjBigAnd
}

func createNewAigWith0LivePoWithDSC(pAig *Aig_Man_t, signalList []Aig_Obj_t, index0Live *int32, windowBeginIndex, withinWindowIndex, pendingSignalIndex int32, vBarriers [][]int32) *Aig_Man_t {
	var vBarrierLiDriver, vMonotoneNodes []Aig_Obj_t

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNewAig := aig.Aig_ManStart(aig.Aig_ManObjNumMax(pAig))
	pNewAig.Name = fmt.Sprintf("%s_%s", pAig.Name, "0Live")
	pNewAig.Spec = ""

	//****************************************************************
	// Step 2: map constant nodes
	//****************************************************************
	aigSetData(pAig, aig.Aig_ManConst1(pAig), aig.Aig_ManConst1(pNewAig))

	//****************************************************************
	// Step 3: create true PIs
	//****************************************************************
	for i := int32(0); i < saigPiNum(pAig); i++ {
		aigSetData(pAig, aig.Aig_ManCi(pAig, i), aig.Aig_ObjCreateCi(pNewAig))
	}

	//****************************************************************
	// Step 4: create register outputs
	//****************************************************************
	for i := int32(0); i < saigRegNum(pAig); i++ {
		aigSetData(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ObjCreateCi(pNewAig))
	}

	//****************************************************************
	// Step 4.a: create register outputs for the barrier flops
	//****************************************************************
	vBarrierLo := createArenaLO(pNewAig, vBarriers)

	//****************************************************************
	// Step 4.b: create register output for arenaViolationFlop
	//****************************************************************
	pObjArenaViolationLo := aig.Aig_ObjCreateCi(pNewAig)

	//****************************************************************
	// Step 4.c: create register output for pendingFlop
	//****************************************************************
	pObjPendingFlopLo := aig.Aig_ObjCreateCi(pNewAig)

	//****************************************************************
	// Step 4.d: create register outputs for the barrier flops
	// for asserting monotonicity
	//****************************************************************
	vMonotoneBarrierLo := createMonotoneBarrierLO(pNewAig, vBarriers)

	//********************************************************************
	// Step 5: create internal nodes
	//********************************************************************
	nodes, _ := aigManNodes(pAig)
	for _, pObj := range nodes {
		aigSetData(pAig, pObj, aig.Aig_And(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj), aig.Aig_ObjChild1Copy(pAig, pObj)))
	}

	//********************************************************************
	// Step 5.a: create internal nodes corresponding to arenaViolation
	//********************************************************************
	pObjTarget := aig.Aig_ManCo(pAig, windowBeginIndex)
	pObjWindowBeginsNew := driverToPoNew(pAig, pObjTarget)

	pObjTarget = aig.Aig_ManCo(pAig, withinWindowIndex)
	pObjWithinWindowNew := driverToPoNew(pAig, pObjTarget)

	vBarrierLiDriver = make([]Aig_Obj_t, 0, len(vBarriers))
	vMonotoneNodes = make([]Aig_Obj_t, 0, len(vBarriers))

	pObjArenaViolation := createArenaViolation(pAig, pNewAig,
		pObjWindowBeginsNew, pObjWithinWindowNew,
		vBarriers, vBarrierLo, &vBarrierLiDriver, &vMonotoneNodes)

	pObjArenaViolationLiDriver := aig.Aig_Or(pNewAig, pObjArenaViolation, pObjArenaViolationLo)

	//********************************************************************
	// Step 5.b: Create internal nodes for monotone testing
	//********************************************************************

	pObjTarget = aig.Aig_ManCo(pAig, pendingSignalIndex)
	pObjPendingSignal := driverToPoNew(pAig, pObjTarget)

	pObjPendingAndPendingSignal := aig.Aig_And(pNewAig, pObjPendingSignal, pObjPendingFlopLo)
	pObjMonotoneAnd := aig.Aig_ManConst1(pNewAig)
	for i, pObj := range vMonotoneNodes {
		pObjCurrMonotoneLo := vMonotoneBarrierLo[i]
		pObjMonotoneAnd = aig.Aig_And(pNewAig, pObjMonotoneAnd,
			aig.Aig_Or(pNewAig,
				aig.Aig_Not(aig.Aig_And(pNewAig, pObjPendingAndPendingSignal, pObjCurrMonotoneLo)),
				pObj))
	}

	//********************************************************************
	// Step 6: create primary outputs
	//********************************************************************

	var i int32
	for i = 0; i < saigPoNum(pAig); i++ {
		pObj := aig.Aig_ManCo(pAig, i)
		aigSetData(pAig, pObj, aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj)))
	}

	pObjNewPoDriver := createConstrained0LiveConeWithDSC(pNewAig, pAig, signalList)
	pObjNewPoDriverArenaViolated := aig.Aig_Or(pNewAig, pObjNewPoDriver, pObjArenaViolationLo)
	pObjNewPoDriverArenaViolated = aig.Aig_And(pNewAig, pObjNewPoDriverArenaViolated, pObjMonotoneAnd)
	aig.Aig_ObjCreateCo(pNewAig, pObjNewPoDriverArenaViolated)

	*index0Live = i

	//********************************************************************
	// Step 7: create register inputs
	//********************************************************************

	loCopied := saigRegNum(pAig)
	for i := int32(0); i < saigRegNum(pAig); i++ {
		pObj := saig.Saig_ManLi(pAig, i)
		aigSetData(pAig, pObj, aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj)))
	}

	//********************************************************************
	// Step 7.a: create register inputs for barrier flops
	//********************************************************************
	createArenaLi(pNewAig, vBarriers, vBarrierLiDriver)

	//********************************************************************
	// Step 7.b: create register inputs for arenaViolation flop
	//********************************************************************
	aig.Aig_ObjCreateCo(pNewAig, pObjArenaViolationLiDriver)

	//****************************************************************
	// Step 7.c: create register input for pendingFlop
	//****************************************************************
	aig.Aig_ObjCreateCo(pNewAig, pObjPendingSignal)

	//********************************************************************
	// Step 7.d: create register inputs for the barrier flops
	// for asserting monotonicity
	//********************************************************************
	for _, pObj := range vMonotoneNodes {
		aig.Aig_ObjCreateCo(pNewAig, pObj)
	}

	loCreated := int32(len(vBarrierLo)) + 1 + 1 + int32(len(vMonotoneBarrierLo))
	//next step should be changed
	aig.Aig_ManSetRegNum(pNewAig, loCopied+loCreated)
	aig.Aig_ManCleanup(pNewAig)

	return pNewAig
}

func generateWorkingAigWithDSC(pAig *Aig_Man_t, pNtk *Abc_Ntk_t, pIndex0Live *int32, vMasterBarriers [][]int32) *Aig_Man_t {
	vSignalVector := collectCSSignalsWithDSC(pNtk, pAig)

	pObjWindowBegin := collectWindowBeginSignalWithDSC(pNtk, pAig)
	pObjWithinWindow := collectWithinWindowSignalWithDSC(pNtk, pAig)
	pObjPendingSignal := collectPendingSignalWithDSC(pNtk, pAig)

	pAigNew := createNewAigWith0LivePoWithDSC(pAig, vSignalVector, pIndex0Live, pObjWindowBegin, pObjWithinWindow, pObjPendingSignal, vMasterBarriers)

	return pAigNew
}
