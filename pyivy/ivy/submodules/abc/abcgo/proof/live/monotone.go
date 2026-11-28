/**CFile****************************************************************

  FileName    [monotone.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Main implementation module of the algorithm k-Liveness ]
              [invented by Koen Claessen, Niklas Sorensson. Implements]
              [the code for 'kcs'. 'kcs' pre-processes based on switch]
              [and then runs the (absorber circuit >> pdr) loop ]

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
	"github.com/glycerine/abcgo/proof/pdr"
)

type aigPoIndices struct {
	attrPendingSignalIndex        int32
	attrHintSingalBeginningMarker int32
	attrHintSingalEndMarker       int32
	attrSafetyInvarIndex          int32
}

func allocAigPoIndices() *aigPoIndices {
	return &aigPoIndices{
		attrPendingSignalIndex:        -1,
		attrHintSingalBeginningMarker: -1,
		attrHintSingalEndMarker:       -1,
		attrSafetyInvarIndex:          -1,
	}
}

func deallocAigPoIndices(toBeDeletedAigPoIndices *aigPoIndices) {}

type monotoneVectorsStruct struct {
	attrKnownMonotone []int32
	attrCandMonotone  []int32
	attrHintMonotone  []int32
}

func allocPointersToMonotoneVectors() *monotoneVectorsStruct {
	return &monotoneVectorsStruct{}
}

func deallocPointersToMonotoneVectors(toBeDeleted *monotoneVectorsStruct) {}

// Vec_IntFind
func intFind(v []int32, x int32) int {
	for i, e := range v {
		if e == x {
			return i
		}
	}
	return -1
}

func findHintOutputs(pNtk *Abc_Ntk_t) []int32 {
	vHintPoIntdex := make([]int32, 0)
	for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
		if strings.Contains(abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)), "hint_") {
			vHintPoIntdex = append(vHintPoIntdex, i)
		}
	}
	if len(vHintPoIntdex) == 0 {
		return nil
	}
	return vHintPoIntdex
}

func findPendingSignal(pNtk *Abc_Ntk_t) int32 {
	pendingSignalIndex := int32(-1)
	for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
		if strings.Contains(abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)), "pendingSignal") {
			pendingSignalIndex = i
			break
		}
	}
	return pendingSignalIndex
}

func checkSanityOfKnownMonotone(vKnownMonotone, vCandMonotone, vHintMonotone []int32) int32 {
	for _, iElem := range vKnownMonotone {
		fmt.Printf("%d ", iElem)
	}
	fmt.Printf("\n")
	for _, iElem := range vCandMonotone {
		fmt.Printf("%d ", iElem)
	}
	fmt.Printf("\n")
	for _, iElem := range vHintMonotone {
		fmt.Printf("%d ", iElem)
	}
	fmt.Printf("\n")
	return 1
}

func createMonotoneTester(pAig *Aig_Man_t, aigPoIndicesArg *aigPoIndices, monotoneVectorArg *monotoneVectorsStruct, startMonotonePropPo *int32) *Aig_Man_t {
	var liCopied, liCreated, loCopied, loCreated int32
	var poCopied, poCreated int32

	pendingSignalIndexLocal := aigPoIndicesArg.attrPendingSignalIndex
	hintSingalBeginningMarkerLocal := aigPoIndicesArg.attrHintSingalBeginningMarker

	vKnownMonotoneLocal := monotoneVectorArg.attrKnownMonotone
	vHintMonotoneLocal := monotoneVectorArg.attrHintMonotone

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNewAig := aig.Aig_ManStart(aig.Aig_ManObjNumMax(pAig))
	pNewAig.Name = fmt.Sprintf("%s_%s", pAig.Name, "_monotone")
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
	// Step 5: create register outputs
	//****************************************************************
	for i := int32(0); i < saigRegNum(pAig); i++ {
		loCopied++
		aigSetData(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ObjCreateCi(pNewAig))
	}

	//****************************************************************
	// Step 6: create "D" register output for PENDING flop
	//****************************************************************
	loCreated++
	pPendingFlop := aig.Aig_ObjCreateCi(pNewAig)

	//****************************************************************
	// Step 6.a: create "D" register output for HINT_MONOTONE flop
	//****************************************************************
	vHintMonotoneLocalFlopOutput := make([]Aig_Obj_t, 0, len(vHintMonotoneLocal))
	for range vHintMonotoneLocal {
		loCreated++
		vHintMonotoneLocalFlopOutput = append(vHintMonotoneLocalFlopOutput, aig.Aig_ObjCreateCi(pNewAig))
	}

	nRegCount := loCreated + loCopied
	fmt.Printf("\nnRegCount = %d\n", nRegCount)

	//********************************************************************
	// Step 7: create internal nodes
	//********************************************************************
	nodes, _ := aigManNodes(pAig)
	for _, pObj := range nodes {
		aigSetData(pAig, pObj, aig.Aig_And(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj), aig.Aig_ObjChild1Copy(pAig, pObj)))
	}

	//********************************************************************
	// Step 8: mapping appropriate new flop drivers
	//********************************************************************

	pObjPo := aig.Aig_ManCo(pAig, pendingSignalIndexLocal)
	pObjDriver := driverOld(pAig, pObjPo)
	pObjPendingDriverNew := copyOf(pAig, pObjDriver)

	pObjPendingAndNextPending := aig.Aig_And(pNewAig, pObjPendingDriverNew, pPendingFlop)

	oldPoNum := aig.Aig_ManCoNum(pAig) - aig.Aig_ManRegNum(pAig)
	pObjKnownMonotoneAnd := aig.Aig_ManConst1(pNewAig)
	if vKnownMonotoneLocal != nil {
		// assert( checkSanityOfKnownMonotone( ... ) ) is not executed in
		// the release build

		for _, iElemHint := range vKnownMonotoneLocal {
			iElem := (iElemHint - hintSingalBeginningMarkerLocal) + 1 + pendingSignalIndexLocal
			fmt.Printf("\nProcessing knownMonotone = %d\n", iElem)
			pObjPo = aig.Aig_ManCo(pAig, iElem)
			pObjDriver = driverOld(pAig, pObjPo)
			pObjDriverNew := copyOf(pAig, pObjDriver)
			pHintSignalLo := vHintMonotoneLocalFlopOutput[iElem-oldPoNum]
			// GCC evaluates the arguments of Aig_Or right to left
			pRight := aig.Aig_And(pNewAig, aig.Aig_Not(pObjDriverNew), aig.Aig_Not(pHintSignalLo))
			pLeft := aig.Aig_And(pNewAig, pObjDriverNew, pHintSignalLo)
			pObjTemp1 := aig.Aig_Or(pNewAig, pLeft, pRight)

			pObjKnownMonotoneAnd = aig.Aig_And(pNewAig, pObjKnownMonotoneAnd, pObjTemp1)
		}
		pObjPendingAndNextPending = aig.Aig_And(pNewAig, pObjPendingAndNextPending, pObjKnownMonotoneAnd)
	}

	vHintMonotoneLocalDriverNew := make([]Aig_Obj_t, 0, len(vHintMonotoneLocal))
	vHintMonotoneLocalProp := make([]Aig_Obj_t, 0, len(vHintMonotoneLocal))
	for _, iElem := range vHintMonotoneLocal {
		pObjPo = aig.Aig_ManCo(pAig, iElem)
		pObjDriver = driverOld(pAig, pObjPo)
		pObjDriverNew := copyOf(pAig, pObjDriver)

		if vKnownMonotoneLocal != nil && intFind(vKnownMonotoneLocal, iElem) != -1 {
			vHintMonotoneLocalDriverNew = append(vHintMonotoneLocalDriverNew, pObjDriverNew)
		} else {
			poSerialNum := intFind(vHintMonotoneLocal, iElem)
			pHintSignalLo := vHintMonotoneLocalFlopOutput[poSerialNum]
			pObjTemp1 := aig.Aig_And(pNewAig, pObjPendingAndNextPending, pHintSignalLo)
			pObjTemp2 := aig.Aig_Or(pNewAig, aig.Aig_Not(pObjTemp1), pObjDriverNew)
			vHintMonotoneLocalDriverNew = append(vHintMonotoneLocalDriverNew, pObjDriverNew)
			vHintMonotoneLocalProp = append(vHintMonotoneLocalProp, pObjTemp2)
		}
	}

	//********************************************************************
	// Step 9: create primary outputs
	//********************************************************************
	var i int32
	for i = 0; i < saigPoNum(pAig); i++ {
		poCopied++
		pObj := aig.Aig_ManCo(pAig, i)
		aigSetData(pAig, pObj, aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj)))
	}

	*startMonotonePropPo = i
	for _, pObj := range vHintMonotoneLocalProp {
		poCreated++
		aig.Aig_ObjCreateCo(pNewAig, pObj)
	}

	//********************************************************************
	// Step 9: create latch inputs
	//********************************************************************

	for i := int32(0); i < saigRegNum(pAig); i++ {
		liCopied++
		aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, saig.Saig_ManLi(pAig, i)))
	}

	//********************************************************************
	// Step 9.a: create latch input for PENDING_FLOP
	//********************************************************************

	liCreated++
	aig.Aig_ObjCreateCo(pNewAig, pObjPendingDriverNew)

	//********************************************************************
	// Step 9.b: create latch input for MONOTONE_FLOP
	//********************************************************************

	for _, pObj := range vHintMonotoneLocalDriverNew {
		liCreated++
		aig.Aig_ObjCreateCo(pNewAig, pObj)
	}

	fmt.Printf("\npoCopied = %d, poCreated = %d\n", poCopied, poCreated)
	fmt.Printf("\nliCreated++ = %d\n", liCreated)
	aig.Aig_ManSetRegNum(pNewAig, nRegCount)
	aig.Aig_ManCleanup(pNewAig)

	fmt.Printf("\nSaig_ManPoNum = %d\n", saigPoNum(pNewAig))
	_ = liCopied
	return pNewAig
}

func findNewMonotone(pAig *Aig_Man_t, aigPoIndicesArg *aigPoIndices, monotoneVectorArg *monotoneVectorsStruct) []int32 {
	var poMarker int32
	var Pars pdr.Pdr_Par_t
	pPars := &Pars

	pendingSignalIndexLocal := aigPoIndicesArg.attrPendingSignalIndex
	hintSingalBeginningMarkerLocal := aigPoIndicesArg.attrHintSingalBeginningMarker

	pAigNew := createMonotoneTester(pAig, aigPoIndicesArg, monotoneVectorArg, &poMarker)
	oldPoNum := aig.Aig_ManCoNum(pAig) - aig.Aig_ManRegNum(pAig)

	vMonotoneIndex := make([]int32, 0)
	fmt.Printf("\nSaig_ManPoNum(pAigNew) = %d, poMarker = %d\n", saigPoNum(pAigNew), poMarker)
	for i := poMarker; i < saigPoNum(pAigNew); i++ {
		pObjTargetPo := aig.Aig_ManCo(pAigNew, i)
		aig.Aig_ObjChild0Flip(pAigNew, pObjTargetPo)

		pdr.Pdr_ManSetDefaultParams(pPars)
		pPars.FVerbose = 0
		RetValue := pdr.Pdr_ManSolve(pAigNew, pPars)
		if RetValue == 1 {
			fmt.Printf("\ni = %d, RetValue = %d : %s (Frame %d)\n", i-oldPoNum+hintSingalBeginningMarkerLocal, RetValue, "Property Proved", -1)
			vMonotoneIndex = append(vMonotoneIndex, i-(pendingSignalIndexLocal+1)+hintSingalBeginningMarkerLocal)
		}
		aig.Aig_ObjChild0Flip(pAigNew, pObjTargetPo)
	}

	if len(vMonotoneIndex) > 0 {
		return vMonotoneIndex
	}
	return nil
}

func findRemainingMonotoneCandidates(vKnownMonotone, vHintMonotone []int32) []int32 {
	if vKnownMonotone == nil || len(vKnownMonotone) <= 0 {
		return vHintMonotone
	}
	vCandMonotone := make([]int32, 0)
	for _, iElem := range vHintMonotone {
		if intFind(vKnownMonotone, iElem) == -1 {
			vCandMonotone = append(vCandMonotone, iElem)
		}
	}
	return vCandMonotone
}

// findMonotoneSignals is not called by any command.
func findMonotoneSignals(pNtk *Abc_Ntk_t) []int32 {
	var pAig *Aig_Man_t

	/*******************************************/
	//Finding the PO index of the pending signal
	/*******************************************/
	pendingSignalIndex := findPendingSignal(pNtk)
	if pendingSignalIndex == -1 {
		fmt.Printf("\nNo Pending Signal Found\n")
		return nil
	}
	fmt.Printf("Po[%d] = %s\n", pendingSignalIndex, abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, pendingSignalIndex)))

	/*******************************************/
	//Finding the PO indices of all hint signals
	/*******************************************/
	vCandidateMonotoneSignals := findHintOutputs(pNtk)
	if vCandidateMonotoneSignals == nil {
		return nil
	}
	for _, iElem := range vCandidateMonotoneSignals {
		fmt.Printf("Po[%d] = %s\n", iElem, abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, iElem)))
	}
	hintSingalBeginningMarker := vCandidateMonotoneSignals[0]
	hintSingalEndMarker := vCandidateMonotoneSignals[len(vCandidateMonotoneSignals)-1]

	/**********************************************/
	//Allocating "struct" with necessary parameters
	/**********************************************/
	aigPoIndicesInstance := allocAigPoIndices()
	aigPoIndicesInstance.attrPendingSignalIndex = pendingSignalIndex
	aigPoIndicesInstance.attrHintSingalBeginningMarker = hintSingalBeginningMarker
	aigPoIndicesInstance.attrHintSingalEndMarker = hintSingalEndMarker

	/****************************************************/
	//Allocating "struct" with necessary monotone vectors
	/****************************************************/
	monotoneVectorsInstance := allocPointersToMonotoneVectors()
	monotoneVectorsInstance.attrCandMonotone = vCandidateMonotoneSignals
	monotoneVectorsInstance.attrHintMonotone = vCandidateMonotoneSignals

	/*******************************************/
	//Generate AIG from Ntk
	/*******************************************/
	if !(abc.Abc_NtkIsStrash(pNtk) != 0) {
		pNtkTemp := Abc_NtkStrash(pNtk, 0, 0, 0)
		pAig = Abc_NtkToDar(pNtkTemp, 0, 1)
	} else {
		pAig = Abc_NtkToDar(pNtk, 0, 1)
	}

	/*******************************************/
	//finding LEVEL 1 monotone signals
	/*******************************************/
	vKnownMonotoneSignals := findNewMonotone(pAig, aigPoIndicesInstance, monotoneVectorsInstance)
	monotoneVectorsInstance.attrKnownMonotone = vKnownMonotoneSignals

	deallocAigPoIndices(aigPoIndicesInstance)
	deallocPointersToMonotoneVectors(monotoneVectorsInstance)
	return nil
}
