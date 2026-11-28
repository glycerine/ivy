/**CFile****************************************************************

  FileName    [liveness_sim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Main implementation module.]

  Author      [Sayak Ray]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2009.]

  Revision    [$Id: liveness_sim.c,v 1.00 2009/01/01 00:00:00 alanmi Exp $]

***********************************************************************/

package live

import (
	"fmt"
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	abc "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
)

// The static helpers of liveness_sim.c that differ from those of
// liveness.c carry the suffix Sim; printVecPtrOfString, getPoIndex,
// retrieveTruePiName, retrieveLOName and the Biere cleanups are the same.

func LivenessToSafetyTransformationSim(pNtk *Abc_Ntk_t, p *Aig_Man_t, vLive, vFair []Aig_Obj_t) *Aig_Man_t {
	var nRegCount int32

	vecPis = make([]Aig_Obj_t, 0, saigPiNum(p)+1)
	vecPiNames = make([]string, 0, saigPiNum(p)+1)
	vecLos = make([]Aig_Obj_t, 0, 2*saigRegNum(p)+1+int32(len(vLive)+len(vFair)))
	vecLoNames = make([]string, 0, 2*saigRegNum(p)+1+int32(len(vLive)+len(vFair)))

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNew := aig.Aig_ManStart(2 * aig.Aig_ManObjNumMax(p))
	pNew.Name = "live2safe"
	pNew.Spec = ""

	//****************************************************************
	// Step 2: map constant nodes
	//****************************************************************
	aigSetData(p, aig.Aig_ManConst1(p), aig.Aig_ManConst1(pNew))

	//****************************************************************
	// Step 3: create true PIs
	//****************************************************************
	for i := int32(0); i < saigPiNum(p); i++ {
		pObj := aig.Aig_ManCi(p, i)
		aigSetData(p, pObj, aig.Aig_ObjCreateCi(pNew))
		vecPis = append(vecPis, aigData(p, pObj))
		vecPiNames = append(vecPiNames, abc.Abc_ObjName(abc.Abc_NtkPi(pNtk, i)))
	}

	//****************************************************************
	// Step 4: create the special Pi corresponding to SAVE
	//****************************************************************
	pObjSavePi := aig.Aig_ObjCreateCi(pNew)
	vecPiNames = append(vecPiNames, "SAVE_BIERE")

	//****************************************************************
	// Step 5: create register outputs
	//****************************************************************
	for i := int32(0); i < saigRegNum(p); i++ {
		pObj := saig.Saig_ManLo(p, i)
		aigSetData(p, pObj, aig.Aig_ObjCreateCi(pNew))
		vecLos = append(vecLos, aigData(p, pObj))
		vecLoNames = append(vecLoNames, abc.Abc_ObjName(abc.Abc_NtkCi(pNtk, abc.Abc_NtkPiNum(pNtk)+i)))
	}

	//****************************************************************
	// Step 6: create "saved" register output
	//****************************************************************
	pObjSavedLo := aig.Aig_ObjCreateCi(pNew)
	vecLos = append(vecLos, pObjSavedLo)
	vecLoNames = append(vecLoNames, "SAVED_LO")

	//****************************************************************
	// Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
	//****************************************************************
	pObjSaveOrSaved := aig.Aig_Or(pNew, pObjSavePi, pObjSavedLo)

	//********************************************************************
	// Step 8: create internal nodes
	//********************************************************************
	nodes, pObj := aigManNodes(p)
	for _, pNode := range nodes {
		aigSetData(p, pNode, aig.Aig_And(pNew, aig.Aig_ObjChild0Copy(p, pNode), aig.Aig_ObjChild1Copy(p, pNode)))
	}

	//********************************************************************
	// Step 9: create the safety property output gate
	// create the safety property output gate, this will be the sole true PO
	// of the whole circuit, discuss with Sat/Alan for an alternative implementation
	//********************************************************************
	pObjSafetyPropertyOutput := aig.Aig_ObjCreateCo(pNew, aigData(p, aig.Aig_ObjFanin0(p, pObj)))

	// create register inputs for the original registers
	nRegCount = 0
	for i := int32(0); i < saigRegNum(p); i++ {
		pMatch := saig.Saig_ObjLoToLi(p, saig.Saig_ManLo(p, i))
		aig.Aig_ObjCreateCo(pNew, driverNew(p, pMatch))
		nRegCount++
	}

	// create register input corresponding to the register "saved"
	aig.Aig_ObjCreateCo(pNew, pObjSaveOrSaved)
	nRegCount++

	pObjAndAcc := aigNull

	// create the family of shadow registers, then create the cascade of Xnor and And gates for the comparator
	for i := int32(0); i < saigRegNum(p); i++ {
		pObj := saig.Saig_ManLo(p, i)
		pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

		vecLos = append(vecLos, pObjShadowLo)
		vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkCi(pNtk, abc.Abc_NtkPiNum(pNtk)+i)), "SHADOW"))

		pObjShadowLiDriver := aigMux(pNew, pObjSavePi, aigData(p, pObj), pObjShadowLo)
		aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
		nRegCount++

		pObjXor := aig.Aig_Exor(pNew, aigData(p, pObj), pObjShadowLo)
		pObjXnor := aig.Aig_Not(pObjXor)
		if pObjAndAcc == aigNull {
			pObjAndAcc = pObjXnor
		} else {
			pObjAndAccDummy := pObjAndAcc
			pObjAndAcc = aig.Aig_And(pNew, pObjXnor, pObjAndAccDummy)
		}
	}

	// with no registers the equality holds trivially (as in liveness.c;
	// C ANDed NULL here: CBUGS_FIXED.md #53)
	if pObjAndAcc == aigNull {
		pObjAndAcc = aig.Aig_ManConst1(pNew)
	}
	// create the AND gate whose output will be the signal "looped"
	pObjSavedLoAndEquality := aig.Aig_And(pNew, pObjSavedLo, pObjAndAcc)

	// create the master AND gate and corresponding AND and OR logic for the liveness properties
	pObjAndAcc = aigNull
	if len(vLive) == 0 {
		fmt.Printf("\nCircuit without any liveness property\n")
	} else {
		for _, pObj := range vLive {
			pDriverImage := driverNew(p, pObj)
			pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

			vecLos = append(vecLos, pObjShadowLo)
			vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, getPoIndex(p, pObj))), "LIVENESS"))

			// GCC evaluates the arguments of Aig_Or right to left
			pAnd := aig.Aig_And(pNew, pDriverImage, pObjSaveOrSaved)
			pMux := aigMux(pNew, pObjSavePi, aig.Aig_Not(aig.Aig_ManConst1(pNew)), pObjShadowLo)
			pObjShadowLiDriver := aig.Aig_Or(pNew, pMux, pAnd)
			aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
			nRegCount++

			if pObjAndAcc == aigNull {
				pObjAndAcc = pObjShadowLo
			} else {
				pObjAndAccDummy := pObjAndAcc
				pObjAndAcc = aig.Aig_And(pNew, pObjShadowLo, pObjAndAccDummy)
			}
		}
	}

	var pObjLive Aig_Obj_t
	if pObjAndAcc != aigNull {
		pObjLive = pObjAndAcc
	} else {
		pObjLive = aig.Aig_ManConst1(pNew)
	}

	// create the master AND gate and corresponding AND and OR logic for the fairness properties
	pObjAndAcc = aigNull
	if len(vFair) == 0 {
		fmt.Printf("\nCircuit without any fairness property\n")
	} else {
		for _, pObj := range vFair {
			pDriverImage := driverNew(p, pObj)
			pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

			vecLos = append(vecLos, pObjShadowLo)
			vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, getPoIndex(p, pObj))), "FAIRNESS"))

			// GCC evaluates the arguments of Aig_Or right to left
			pAnd := aig.Aig_And(pNew, pDriverImage, pObjSaveOrSaved)
			pMux := aigMux(pNew, pObjSavePi, aig.Aig_Not(aig.Aig_ManConst1(pNew)), pObjShadowLo)
			pObjShadowLiDriver := aig.Aig_Or(pNew, pMux, pAnd)
			aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
			nRegCount++

			if pObjAndAcc == aigNull {
				pObjAndAcc = pObjShadowLo
			} else {
				pObjAndAccDummy := pObjAndAcc
				pObjAndAcc = aig.Aig_And(pNew, pObjShadowLo, pObjAndAccDummy)
			}
		}
	}

	var pObjFair Aig_Obj_t
	if pObjAndAcc != aigNull {
		pObjFair = pObjAndAcc
	} else {
		pObjFair = aig.Aig_ManConst1(pNew)
	}

	pObjSafetyGate := aig.Aig_And(pNew, pObjSavedLoAndEquality, aig.Aig_And(pNew, pObjFair, aig.Aig_Not(pObjLive)))

	aig.Aig_ObjPatchFanin0(pNew, pObjSafetyPropertyOutput, pObjSafetyGate)

	aig.Aig_ManSetRegNum(pNew, nRegCount)

	Aig_ManCiCleanupBiere(pNew)
	Aig_ManCoCleanupBiere(pNew)

	aig.Aig_ManCleanup(pNew)

	return pNew
}

func LivenessToSafetyTransformationOneStepLoopSim(pNtk *Abc_Ntk_t, p *Aig_Man_t, vLive, vFair []Aig_Obj_t) *Aig_Man_t {
	var nRegCount int32

	vecPis = make([]Aig_Obj_t, 0, saigPiNum(p)+1)
	vecPiNames = make([]string, 0, saigPiNum(p)+1)
	vecLos = make([]Aig_Obj_t, 0, 2*saigRegNum(p)+1+int32(len(vLive)+len(vFair)))
	vecLoNames = make([]string, 0, 2*saigRegNum(p)+1+int32(len(vLive)+len(vFair)))

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNew := aig.Aig_ManStart(2 * aig.Aig_ManObjNumMax(p))
	pNew.Name = "live2safe"
	pNew.Spec = ""

	//****************************************************************
	// Step 2: map constant nodes
	//****************************************************************
	aigSetData(p, aig.Aig_ManConst1(p), aig.Aig_ManConst1(pNew))

	//****************************************************************
	// Step 3: create true PIs
	//****************************************************************
	for i := int32(0); i < saigPiNum(p); i++ {
		pObj := aig.Aig_ManCi(p, i)
		aigSetData(p, pObj, aig.Aig_ObjCreateCi(pNew))
		vecPis = append(vecPis, aigData(p, pObj))
		vecPiNames = append(vecPiNames, abc.Abc_ObjName(abc.Abc_NtkPi(pNtk, i)))
	}

	//****************************************************************
	// Step 4: create the special Pi corresponding to SAVE
	//****************************************************************
	pObjSavePi := aig.Aig_ObjCreateCi(pNew)
	vecPiNames = append(vecPiNames, "SAVE_BIERE")

	//****************************************************************
	// Step 5: create register outputs
	//****************************************************************
	for i := int32(0); i < saigRegNum(p); i++ {
		pObj := saig.Saig_ManLo(p, i)
		aigSetData(p, pObj, aig.Aig_ObjCreateCi(pNew))
		vecLos = append(vecLos, aigData(p, pObj))
		vecLoNames = append(vecLoNames, abc.Abc_ObjName(abc.Abc_NtkCi(pNtk, abc.Abc_NtkPiNum(pNtk)+i)))
	}

	//********************************************************************
	// Step 8: create internal nodes
	//********************************************************************
	nodes, pObj := aigManNodes(p)
	for _, pNode := range nodes {
		aigSetData(p, pNode, aig.Aig_And(pNew, aig.Aig_ObjChild0Copy(p, pNode), aig.Aig_ObjChild1Copy(p, pNode)))
	}

	//********************************************************************
	// Step 9: create the safety property output gate
	//********************************************************************
	pObjSafetyPropertyOutput := aig.Aig_ObjCreateCo(pNew, aigData(p, aig.Aig_ObjFanin0(p, pObj)))

	// create register inputs for the original registers
	nRegCount = 0
	for i := int32(0); i < saigRegNum(p); i++ {
		pMatch := saig.Saig_ObjLoToLi(p, saig.Saig_ManLo(p, i))
		aig.Aig_ObjCreateCo(pNew, driverNew(p, pMatch))
		nRegCount++
	}

	pObjAndAcc := aigNull

	//****************************************************************************************************
	//For detection of loop of length 1 we do not need any shadow register, we only need equality detector
	//between Lo_j and Li_j and then a cascade of AND gates
	//****************************************************************************************************
	for i := int32(0); i < saigRegNum(p); i++ {
		pObj := saig.Saig_ManLo(p, i)
		pObjCorrespondingLi := saig.Saig_ObjLoToLi(p, pObj)

		pObjXor := aig.Aig_Exor(pNew, aigData(p, pObj), driverNew(p, pObjCorrespondingLi))
		pObjXnor := aig.Aig_Not(pObjXor)

		if pObjAndAcc == aigNull {
			pObjAndAcc = pObjXnor
		} else {
			pObjAndAccDummy := pObjAndAcc
			pObjAndAcc = aig.Aig_And(pNew, pObjXnor, pObjAndAccDummy)
		}
	}

	// with no registers the equality holds trivially (C ANDed NULL here:
	// CBUGS_FIXED.md #55)
	if pObjAndAcc == aigNull {
		pObjAndAcc = aig.Aig_ManConst1(pNew)
	}
	// create the AND gate whose output will be the signal "looped"
	pObjSavedLoAndEquality := aig.Aig_And(pNew, pObjSavePi, pObjAndAcc)

	// create the master AND gate and corresponding AND and OR logic for the liveness properties
	pObjAndAcc = aigNull
	if len(vLive) == 0 {
		fmt.Printf("\nCircuit without any liveness property\n")
	} else {
		pObjAndAcc = andOfDriversNull(p, pNew, vLive)
	}

	var pObjLive Aig_Obj_t
	if pObjAndAcc != aigNull {
		pObjLive = pObjAndAcc
	} else {
		pObjLive = aig.Aig_ManConst1(pNew)
	}

	// create the master AND gate and corresponding AND and OR logic for the fairness properties
	pObjAndAcc = aigNull
	if len(vFair) == 0 {
		fmt.Printf("\nCircuit without any fairness property\n")
	} else {
		pObjAndAcc = andOfDriversNull(p, pNew, vFair)
	}

	var pObjFair Aig_Obj_t
	if pObjAndAcc != aigNull {
		pObjFair = pObjAndAcc
	} else {
		pObjFair = aig.Aig_ManConst1(pNew)
	}

	pObjSafetyGate := aig.Aig_And(pNew, pObjSavedLoAndEquality, aig.Aig_And(pNew, pObjFair, aig.Aig_Not(pObjLive)))

	aig.Aig_ObjPatchFanin0(pNew, pObjSafetyPropertyOutput, pObjSafetyGate)

	aig.Aig_ManSetRegNum(pNew, nRegCount)

	fmt.Printf("\nSaig_ManPiNum = %d, Reg Num = %d, before everything, before Pi cleanup\n", aig.Aig_ManCiNum(pNew), aig.Aig_ManRegNum(pNew))

	Aig_ManCiCleanupBiere(pNew)
	Aig_ManCoCleanupBiere(pNew)

	aig.Aig_ManCleanup(pNew)

	return pNew
}

func populateLivenessVectorSim(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) []Aig_Obj_t {
	vLive := populatePoVector(pNtk, pAig, func(pNode *Abc_Obj_t) bool { return strings.Contains(abc.Abc_ObjName(pNode), "assert_fair") })
	fmt.Printf("\nNumber of liveness property found = %d\n", len(vLive))
	return vLive
}

func populateFairnessVectorSim(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) []Aig_Obj_t {
	vFair := populatePoVector(pNtk, pAig, func(pNode *Abc_Obj_t) bool { return strings.Contains(abc.Abc_ObjName(pNode), "assume_fair") })
	fmt.Printf("\nNumber of fairness property found = %d\n", len(vFair))
	return vFair
}

// updateNewNetworkNameManagerSim names only the CIs.
func updateNewNetworkNameManagerSim(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, vPiNames, vLoNames []string) {
	storeCiNames(pNtk, pAig, vPiNames, vLoNames)
}

func Abc_CommandAbcLivenessToSafetySim(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pAig, pAigNew *Aig_Man_t
	var vLive, vFair []Aig_Obj_t

	pNtk := frameNtk(pAbc)
	pErr := abcmain.Abc_FrameReadErr(pAbc)

	if pNtk == nil {
		fmt.Fprintf(pErr, "Empty network.\n")
		return 1
	}

	if !(abc.Abc_NtkIsStrash(pNtk) != 0) {
		fmt.Printf("\nThe input network was not strashed, strashing....\n")
		pNtkTemp := Abc_NtkStrash(pNtk, 0, 0, 0)
		pAig = Abc_NtkToDar(pNtkTemp, 0, 1)
		vLive = populateLivenessVectorSim(pNtk, pAig)
		vFair = populateFairnessVectorSim(pNtk, pAig)
	} else {
		pAig = Abc_NtkToDar(pNtk, 0, 1)
		vLive = populateLivenessVectorSim(pNtk, pAig)
		vFair = populateFairnessVectorSim(pNtk, pAig)
	}

	// no Extra_UtilGetoptReset() here: the option state is the previous
	// command's, as in C
	loadLiveNames(pAbc)
	c := extra.Extra_UtilGetopt(argc, argv, "1")
	if c == '1' {
		pAigNew = LivenessToSafetyTransformationOneStepLoopSim(pNtk, pAig, vLive, vFair)
	} else {
		pAigNew = LivenessToSafetyTransformationSim(pNtk, pAig, vLive, vFair)
	}

	pNtkNew := Abc_NtkFromAigPhase(pAigNew)

	if abc.Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Printf("Abc_NtkCreateCone(): Network check has failed.\n")
	}

	updateNewNetworkNameManagerSim(pNtkNew, pAigNew, vecPiNames, vecLoNames)
	saveLiveNames(pAbc)
	abcmain.Abc_FrameSetCurrentNetwork(pAbc, pNtkNew)

	return 0
}
