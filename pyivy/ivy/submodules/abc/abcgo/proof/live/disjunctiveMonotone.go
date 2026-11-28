/**CFile****************************************************************

  FileName    [disjunctiveMonotone.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Constraint analysis module for the k-Liveness algorithm
        invented by Koen Classen, Niklas Sorensson.]

  Author      [Sayak Ray]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 31, 2012.]

***********************************************************************/

package live

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	abc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/pdr"
)

type antecedentConsequentVectorsStruct struct {
	attrAntecedents          []int32
	attrConsequentCandidates []int32
}

func allocAntecedentConsequentVectorsStruct() *antecedentConsequentVectorsStruct {
	return &antecedentConsequentVectorsStruct{}
}

func deallocAntecedentConsequentVectorsStruct(toBeDeleted *antecedentConsequentVectorsStruct) {}

func createDisjunctiveMonotoneTester(pAig *Aig_Man_t, aigPoIndicesArg *aigPoIndices,
	anteConseVectors *antecedentConsequentVectorsStruct, startMonotonePropPo *int32) *Aig_Man_t {
	var liCopied, liCreated, loCopied, loCreated int32

	pendingSignalIndexLocal := aigPoIndicesArg.attrPendingSignalIndex

	vAntecedentsLocal := anteConseVectors.attrAntecedents
	vConsequentCandidatesLocal := anteConseVectors.attrConsequentCandidates

	if vConsequentCandidatesLocal == nil {
		return nil //no candidates for consequent is provided, hence no need to generate a new AIG
	}

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNewAig := aig.Aig_ManStart(aig.Aig_ManObjNumMax(pAig))
	pNewAig.Name = fmt.Sprintf("%s_%s", pAig.Name, "monotone")
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
	vConseCandFlopOutput := make([]Aig_Obj_t, 0, len(vConsequentCandidatesLocal))
	for range vConsequentCandidatesLocal {
		loCreated++
		vConseCandFlopOutput = append(vConseCandFlopOutput, aig.Aig_ObjCreateCi(pNewAig))
	}

	nRegCount := loCreated + loCopied

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

	var pObjSafetyInvariantPoDriver Aig_Obj_t
	if aigPoIndicesArg.attrSafetyInvarIndex != -1 {
		pObjPo := aig.Aig_ManCo(pAig, aigPoIndicesArg.attrSafetyInvarIndex)
		pObjDriver := driverOld(pAig, pObjPo)
		pObjSafetyInvariantPoDriver = copyOf(pAig, pObjDriver)
	} else {
		pObjSafetyInvariantPoDriver = aig.Aig_ManConst1(pNewAig)
	}

	pObjPo := aig.Aig_ManCo(pAig, pendingSignalIndexLocal)
	pObjDriver := driverOld(pAig, pObjPo)
	pObjPendingDriverNew := copyOf(pAig, pObjDriver)

	pObjPendingAndNextPending := aig.Aig_And(pNewAig, pObjPendingDriverNew, pPendingFlop)

	pObjAnteDisjunction := aig.Aig_Not(aig.Aig_ManConst1(pNewAig))
	if vAntecedentsLocal != nil {
		for _, iElem := range vAntecedentsLocal {
			pObjPo := aig.Aig_ManCo(pAig, iElem)
			pObjDriver := driverOld(pAig, pObjPo)
			pObjDriverNew := copyOf(pAig, pObjDriver)

			pObjAnteDisjunction = aig.Aig_Or(pNewAig, pObjDriverNew, pObjAnteDisjunction)
		}
	}

	vCandMonotoneProp := make([]Aig_Obj_t, 0, len(vConsequentCandidatesLocal))
	vCandMonotoneFlopInput := make([]Aig_Obj_t, 0, len(vConsequentCandidatesLocal))
	for i, iElem := range vConsequentCandidatesLocal {
		pObjPo := aig.Aig_ManCo(pAig, iElem)
		pObjConsecDriver := driverOld(pAig, pObjPo)
		pObjConsecDriverNew := copyOf(pAig, pObjConsecDriver)

		pObjCandMonotone := aig.Aig_Or(pNewAig, pObjConsecDriverNew, pObjAnteDisjunction)
		pObjPrevCandMonotone := vConseCandFlopOutput[i]
		pObjMonotonePropDriver := aig.Aig_Or(pNewAig, aig.Aig_Not(aig.Aig_And(pNewAig, pObjPendingAndNextPending, pObjPrevCandMonotone)),
			pObjCandMonotone)

		//Conjunting safety invar
		pObjMonotonePropDriver = aig.Aig_And(pNewAig, pObjMonotonePropDriver, pObjSafetyInvariantPoDriver)

		vCandMonotoneFlopInput = append(vCandMonotoneFlopInput, pObjCandMonotone)
		vCandMonotoneProp = append(vCandMonotoneProp, pObjMonotonePropDriver)
	}

	//********************************************************************
	// Step 9: create primary outputs
	//********************************************************************
	var i int32
	for i = 0; i < saigPoNum(pAig); i++ {
		pObj := aig.Aig_ManCo(pAig, i)
		aigSetData(pAig, pObj, aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj)))
	}

	*startMonotonePropPo = i
	for _, pObj := range vCandMonotoneProp {
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

	for _, pObj := range vCandMonotoneFlopInput {
		liCreated++
		aig.Aig_ObjCreateCo(pNewAig, pObj)
	}

	aig.Aig_ManSetRegNum(pNewAig, nRegCount)
	aig.Aig_ManCleanup(pNewAig)

	_, _ = liCopied, liCreated
	return pNewAig
}

func findNewDisjunctiveMonotone(pAig *Aig_Man_t, aigPoIndicesArg *aigPoIndices, anteConseVectors *antecedentConsequentVectorsStruct) []int32 {
	var poMarker int32
	var Pars pdr.Pdr_Par_t
	pPars := &Pars

	pendingSignalIndexLocal := aigPoIndicesArg.attrPendingSignalIndex

	pAigNew := createDisjunctiveMonotoneTester(pAig, aigPoIndicesArg, anteConseVectors, &poMarker)

	vMonotoneIndex := make([]int32, 0)

	for i := int32(0); i < saigPoNum(pAigNew); i++ {
		pObjTargetPo := aig.Aig_ManCo(pAigNew, i)
		aig.Aig_ObjChild0Flip(pAigNew, pObjTargetPo)
	}

	pdr.Pdr_ManSetDefaultParams(pPars)
	pPars.FVerbose = 0
	pPars.FNotVerbose = 1
	pPars.FSolveAll = 1
	pAigNew.SetSequentialCounterexamples(nil)
	pdr.Pdr_ManSolve(pAigNew, pPars)

	if vSeqModelVec := pAigNew.SequentialCounterexamples(); vSeqModelVec != nil {
		for i := 0; i < vec.Vec_PtrSize(vSeqModelVec); i++ {
			if cexIsNull(vec.Vec_PtrEntry(vSeqModelVec, i)) && int32(i) >= pendingSignalIndexLocal+1 {
				poSerialNum := int32(i) - (pendingSignalIndexLocal + 1)
				vMonotoneIndex = append(vMonotoneIndex, anteConseVectors.attrConsequentCandidates[poSerialNum])
			}
		}
	}
	for i := int32(0); i < saigPoNum(pAigNew); i++ {
		pObjTargetPo := aig.Aig_ManCo(pAigNew, i)
		aig.Aig_ObjChild0Flip(pAigNew, pObjTargetPo)
	}

	aig.Aig_ManStop(pAigNew)

	if len(vMonotoneIndex) > 0 {
		return vMonotoneIndex
	}
	return nil
}

// cexIsNull reports a NULL entry of vSeqModelVec.
func cexIsNull(entry any) bool {
	if entry == nil {
		return true
	}
	cex, ok := entry.(*util.Abc_Cex_t)
	return ok && cex == nil
}

func updateAnteConseVectors(anteConse *antecedentConsequentVectorsStruct) []int32 {
	if anteConse.attrAntecedents == nil || len(anteConse.attrAntecedents) <= 0 {
		return anteConse.attrConsequentCandidates
	}
	vCandMonotone := make([]int32, 0)
	for _, iElem := range anteConse.attrConsequentCandidates {
		if intFind(anteConse.attrAntecedents, iElem) == -1 {
			vCandMonotone = append(vCandMonotone, iElem)
		}
	}
	return vCandMonotone
}

func vectorDifference(A, B []int32) []int32 {
	C := make([]int32, 0)
	for _, iElem := range A {
		if intFind(B, iElem) == -1 {
			C = append(C, iElem)
		}
	}
	return C
}

func createSingletonIntVector(iElem int32) []int32 {
	return []int32{iElem}
}

func Vec_IntPushUniqueLocal(p *[]int32, Entry int32) int32 {
	for _, e := range *p {
		if e == Entry {
			return 1
		}
	}
	*p = append(*p, Entry)
	return 0
}

func findNextLevelDisjunctiveMonotone(
	pAig *Aig_Man_t,
	aigPoIndicesInstance *aigPoIndices,
	anteConsecInstance *antecedentConsequentVectorsStruct,
	previousMonotoneVectors [][]int32) [][]int32 {

	newLevelPtrVec := make([][]int32, 0)
	vUnionPrevMonotoneVector := make([]int32, 0)
	for _, vElem := range previousMonotoneVectors {
		for _, iElem := range vElem {
			Vec_IntPushUniqueLocal(&vUnionPrevMonotoneVector, iElem)
		}
	}

	for _, vElem := range previousMonotoneVectors {
		anteConsecInstanceLocal := allocAntecedentConsequentVectorsStruct()

		anteConsecInstanceLocal.attrAntecedents = append([]int32(nil), vElem...)
		vDiffVector := vectorDifference(anteConsecInstance.attrConsequentCandidates, vUnionPrevMonotoneVector)
		anteConsecInstanceLocal.attrConsequentCandidates = vDiffVector

		vNewDisjunctVector := findNewDisjunctiveMonotone(pAig, aigPoIndicesInstance, anteConsecInstanceLocal)

		if vNewDisjunctVector != nil {
			for _, iElem := range vNewDisjunctVector {
				newDisjunction := append([]int32(nil), vElem...)
				newDisjunction = append(newDisjunction, iElem)
				newLevelPtrVec = append(newLevelPtrVec, newDisjunction)
			}
		}
		deallocAntecedentConsequentVectorsStruct(anteConsecInstanceLocal)
	}

	return newLevelPtrVec
}

// printDisjunctions writes "( a || b )" lines for printAllIntVectors and
// printAllIntVectorsStabil (the latter also prints them).
func printDisjunctions(vDisjunctions [][]int32, pNtk *Abc_Ntk_t, fileName, marker string, echo bool) {
	fp, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer fp.Close()

	for i, vElem := range vDisjunctions {
		if echo {
			fmt.Printf("INT[%d] : ( ", i)
		}
		fmt.Fprintf(fp, "( ")
		for j, iElem := range vElem {
			name := abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, iElem))
			hintSubStr := name[strings.Index(name, marker):]
			if echo {
				fmt.Printf("%s", hintSubStr)
			}
			fmt.Fprintf(fp, "%s", hintSubStr)
			if j < len(vElem)-1 {
				if echo {
					fmt.Printf(" || ")
				}
				fmt.Fprintf(fp, " || ")
			} else {
				if echo {
					fmt.Printf(" )\n")
				}
				fmt.Fprintf(fp, " )\n")
			}
		}
	}
}

func printAllIntVectors(vDisjunctions [][]int32, pNtk *Abc_Ntk_t, fileName string) {
	printDisjunctions(vDisjunctions, pNtk, fileName, "hint", false)
}

func printAllIntVectorsStabil(vDisjunctions [][]int32, pNtk *Abc_Ntk_t, fileName string) {
	printDisjunctions(vDisjunctions, pNtk, fileName, "csLevel1Stabil", true)
}

func appendVecToMasterVecInt(masterVec *[][]int32, candVec [][]int32) {
	for _, vCand := range candVec {
		*masterVec = append(*masterVec, append([]int32(nil), vCand...))
	}
}

func deallocateVecOfIntVec(vecOfIntVec [][]int32) {}

func findDisjunctiveMonotoneSignals(pNtk *Abc_Ntk_t) [][]int32 {
	var pAig *Aig_Man_t

	/*******************************************/
	//Finding the PO index of the pending signal
	/*******************************************/
	pendingSignalIndex := findPendingSignal(pNtk)
	if pendingSignalIndex == -1 {
		fmt.Printf("\nNo Pending Signal Found\n")
		return nil
	}

	/*******************************************/
	//Finding the PO indices of all hint signals
	/*******************************************/
	vCandidateMonotoneSignals := findHintOutputs(pNtk)
	if vCandidateMonotoneSignals == nil {
		return nil
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
	aigPoIndicesInstance.attrSafetyInvarIndex = collectSafetyInvariantPOIndex(pNtk)

	/****************************************************/
	//Allocating "struct" with necessary monotone vectors
	/****************************************************/
	anteConsecInstance := allocAntecedentConsequentVectorsStruct()
	anteConsecInstance.attrAntecedents = nil
	anteConsecInstance.attrConsequentCandidates = vCandidateMonotoneSignals

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
	vKnownMonotoneSignals := findNewDisjunctiveMonotone(pAig, aigPoIndicesInstance, anteConsecInstance)
	levelOneMonotne := make([][]int32, 0)
	// nil when no hint is monotone (C iterated over NULL: CBUGS_FIXED.md #56)
	for _, iElem := range vKnownMonotoneSignals {
		levelOneMonotne = append(levelOneMonotne, createSingletonIntVector(iElem))
	}

	vMasterDisjunctions := make([][]int32, 0, len(levelOneMonotne))
	appendVecToMasterVecInt(&vMasterDisjunctions, levelOneMonotne)

	/*******************************************/
	//finding LEVEL >1 monotone signals
	/*******************************************/
	levelTwoMonotne := findNextLevelDisjunctiveMonotone(pAig, aigPoIndicesInstance, anteConsecInstance, levelOneMonotne)
	appendVecToMasterVecInt(&vMasterDisjunctions, levelTwoMonotne)

	deallocAigPoIndices(aigPoIndicesInstance)
	deallocAntecedentConsequentVectorsStruct(anteConsecInstance)

	deallocateVecOfIntVec(levelOneMonotne)
	deallocateVecOfIntVec(levelTwoMonotne)

	aig.Aig_ManStop(pAig)

	return vMasterDisjunctions
}
