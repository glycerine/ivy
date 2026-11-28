/**CFile****************************************************************

  FileName    [liveness.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Liveness property checking.]

  Synopsis    [Main implementation module.]

  Author      [Sayak Ray]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2009.]

  Revision    [$Id: liveness.c,v 1.00 2009/01/01 00:00:00 alanmi Exp $]

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
	"github.com/glycerine/abcgo/misc/nm"
)

// #define PROPAGATE_NAMES
// #define MULTIPLE_LTL_FORMULA
// #define ALLOW_SAFETY_PROPERTIES

const (
	FULL_BIERE_MODE                           = 0
	IGNORE_LIVENESS_KEEP_SAFETY_MODE          = 1
	IGNORE_SAFETY_KEEP_LIVENESS_MODE          = 2
	IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE = 3
	FULL_BIERE_ONE_LOOP_MODE                  = 4
)

/*******************************************************************
LAYOUT OF PI VECTOR:

+------------------------------------------------------------------------------------------------------------------------------------+
| TRUE ORIGINAL PI (n) | SAVE(PI) (1) | ORIGINAL LO (k) | SAVED(LO) (1) | SHADOW_ORIGINAL LO (k) | LIVENESS LO (l) | FAIRNESS LO (f) |
+------------------------------------------------------------------------------------------------------------------------------------+
<------------True PI----------------->|<----------------------------LO--------------------------------------------------------------->

LAYOUT OF PO VECTOR:

+-----------------------------------------------------------------------------------------------------------+
| SOLE PO (1) | ORIGINAL LI (k) | SAVED LI (1) | SHADOW_ORIGINAL LI (k) | LIVENESS LI (l) | FAIRNESS LI (f) |
+-----------------------------------------------------------------------------------------------------------+
<--True PO--->|<--------------------------------------LI---------------------------------------------------->

********************************************************************/

func nodeName_starts_with(pNode *Abc_Obj_t, prefix string) bool {
	return strings.HasPrefix(abc.Abc_ObjName(pNode), prefix)
}

func printVecPtrOfString(vec []string) {
	for i := 0; i < len(vec); i++ {
		fmt.Printf("vec[%d] = %s\n", i, vec[i])
	}
}

func getPoIndex(pAig *Aig_Man_t, pPivot Aig_Obj_t) int32 {
	for i := int32(0); i < saigPoNum(pAig); i++ {
		if aig.Aig_ManCo(pAig, i) == pPivot {
			return i
		}
	}
	return -1
}

func retrieveTruePiName(pNtkOld *Abc_Ntk_t, pAigOld, pAigNew *Aig_Man_t, pObjPivot Aig_Obj_t) string {
	var index int32
	for index = 0; index < aig.Aig_ManCiNum(pAigNew); index++ {
		if aig.Aig_ManCi(pAigNew, index) == pObjPivot {
			break
		}
	}
	if index == saigPiNum(pAigNew)-1 {
		return "SAVE_BIERE"
	}
	pNode := abc.Abc_NtkPi(pNtkOld, index)
	return abc.Abc_ObjName(pNode)
}

func retrieveLOName(pNtkOld *Abc_Ntk_t, pAigOld, pAigNew *Aig_Man_t, pObjPivot Aig_Obj_t, vLive, vFair []Aig_Obj_t) string {
	originalLatchNum := saigRegNum(pAigOld)
	var index int32
	for index = 0; index < saigRegNum(pAigNew); index++ {
		if saig.Saig_ManLo(pAigNew, index) == pObjPivot {
			break
		}
	}
	nLive, nFair := int32(len(vLive)), int32(len(vFair))
	if index < originalLatchNum {
		oldIndex := saigPiNum(pAigOld) + index
		pNode := abc.Abc_NtkCi(pNtkOld, oldIndex)
		return abc.Abc_ObjName(pNode)
	} else if index == originalLatchNum {
		return "SAVED_LO"
	} else if index > originalLatchNum && index < 2*originalLatchNum+1 {
		oldIndex := saigPiNum(pAigOld) + index - originalLatchNum - 1
		pNode := abc.Abc_NtkCi(pNtkOld, oldIndex)
		return fmt.Sprintf("%s__%s", abc.Abc_ObjName(pNode), "SHADOW")
	} else if index >= 2*originalLatchNum+1 && index < 2*originalLatchNum+1+nLive {
		oldIndex := index - 2*originalLatchNum - 1
		strMatch := int32(0)
		dummyStr := ""
		for i := int32(0); i < saigPoNum(pAigOld); i++ {
			pNode := abc.Abc_NtkPo(pNtkOld, i)
			if nodeName_starts_with(pNode, "assert_fair") {
				if strMatch == oldIndex {
					dummyStr = fmt.Sprintf("%s__%s", abc.Abc_ObjName(pNode), "LIVENESS")
					break
				}
				strMatch++
			}
		}
		return dummyStr
	} else if index >= 2*originalLatchNum+1+nLive && index < 2*originalLatchNum+1+nLive+nFair {
		oldIndex := index - 2*originalLatchNum - 1 - nLive
		strMatch := int32(0)
		dummyStr := ""
		for i := int32(0); i < saigPoNum(pAigOld); i++ {
			pNode := abc.Abc_NtkPo(pNtkOld, i)
			if nodeName_starts_with(pNode, "assume_fair") {
				if strMatch == oldIndex {
					dummyStr = fmt.Sprintf("%s__%s", abc.Abc_ObjName(pNode), "FAIRNESS")
					break
				}
				strMatch++
			}
		}
		return dummyStr
	}
	return "UNKNOWN"
}

// The names of the new PIs and LOs, left by the last transformation for
// the command that installs its network (globals in C; nil is NULL).
var vecPis, vecLos []Aig_Obj_t
var vecPiNames, vecLoNames []string

// Aig_ManCiCleanupBiere and Aig_ManCoCleanupBiere recompute the counts
// from the CI/CO arrays; the Go manager keeps them current.
func Aig_ManCiCleanupBiere(p *Aig_Man_t) int32 { return 0 }
func Aig_ManCoCleanupBiere(p *Aig_Man_t) int32 { return 0 }

// collective AND of the CO drivers, starting from constant 1
func andOfDrivers(p, pNew *Aig_Man_t, vObjs []Aig_Obj_t) Aig_Obj_t {
	pObjAndAcc := aig.Aig_ManConst1(pNew)
	for _, pObj := range vObjs {
		pArgument := driverNew(p, pObj)
		pObjAndAcc = aig.Aig_And(pNew, pArgument, pObjAndAcc)
	}
	return pObjAndAcc
}

func LivenessToSafetyTransformation(mode int, pNtk *Abc_Ntk_t, p *Aig_Man_t,
	vLive, vFair, vAssertSafety, vAssumeSafety []Aig_Obj_t) *Aig_Man_t {
	var pObjSavePi, pObjSavedLo Aig_Obj_t = aigNull, aigNull
	var pObjSaveOrSaved, pObjSaveAndNotSaved Aig_Obj_t = aigNull, aigNull
	var pObjSafetyPropertyOutput Aig_Obj_t = aigNull
	var nRegCount int32
	liveLatch, fairLatch := 0, 0

	vecPis = make([]Aig_Obj_t, 0, saigPiNum(p)+1)
	vecPiNames = make([]string, 0, saigPiNum(p)+1)
	vecLos = make([]Aig_Obj_t, 0, 2*saigRegNum(p)+1+int32(len(vLive)+len(vFair)))
	vecLoNames = make([]string, 0, 2*saigRegNum(p)+1+int32(len(vLive)+len(vFair)))

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNew := aig.Aig_ManStart(2 * aig.Aig_ManObjNumMax(p))
	pNew.Name = fmt.Sprintf("%s_%s", pNtk.Name, "l2s")
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
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSavePi = aig.Aig_ObjCreateCi(pNew)
		vecPiNames = append(vecPiNames, "SAVE_BIERE")
	}

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
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSavedLo = aig.Aig_ObjCreateCi(pNew)
		vecLos = append(vecLos, pObjSavedLo)
		vecLoNames = append(vecLoNames, "SAVED_LO")
	}

	//****************************************************************
	// Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
	//****************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSaveOrSaved = aig.Aig_Or(pNew, pObjSavePi, pObjSavedLo)
		pObjSaveAndNotSaved = aig.Aig_And(pNew, pObjSavePi, aig.Aig_Not(pObjSavedLo))
	}

	//********************************************************************
	// Step 8: create internal nodes
	//********************************************************************
	nodes, pObj := aigManNodes(p)
	for _, pNode := range nodes {
		aigSetData(p, pNode, aig.Aig_And(pNew, aig.Aig_ObjChild0Copy(p, pNode), aig.Aig_ObjChild1Copy(p, pNode)))
	}

	//********************************************************************
	// Step 8.x : create PO for each safety assertions
	// NOTE : Here the output is purposely inverted as it will be thrown to
	// dprove
	//********************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
		if len(vAssertSafety) != 0 && len(vAssumeSafety) == 0 {
			pObjAndAcc := andOfDrivers(p, pNew, vAssertSafety)
			pObj = vAssertSafety[len(vAssertSafety)-1]
			aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(pObjAndAcc))
		} else if len(vAssertSafety) != 0 && len(vAssumeSafety) != 0 {
			collectiveAssertSafety := andOfDrivers(p, pNew, vAssertSafety)
			collectiveAssumeSafety := andOfDrivers(p, pNew, vAssumeSafety)
			pObj = vAssumeSafety[len(vAssumeSafety)-1]
			aig.Aig_ObjCreateCo(pNew, aig.Aig_And(pNew, aig.Aig_Not(collectiveAssertSafety), collectiveAssumeSafety))
		} else {
			fmt.Printf("WARNING!! No safety property is found, a new (negated) constant 1 output is created\n")
			aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(aig.Aig_ManConst1(pNew)))
		}
	}

	//********************************************************************
	// Step 9: create the safety property output gate for the liveness properties
	// discuss with Sat/Alan for an alternative implementation
	//********************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		// pObj is the last object the loops above visited
		pObjSafetyPropertyOutput = aig.Aig_ObjCreateCo(pNew, aigData(p, aig.Aig_ObjFanin0(p, pObj)))
	}

	// create register inputs for the original registers
	nRegCount = 0
	for i := int32(0); i < saigRegNum(p); i++ {
		pMatch := saig.Saig_ObjLoToLi(p, saig.Saig_ManLo(p, i))
		aig.Aig_ObjCreateCo(pNew, driverNew(p, pMatch))
		nRegCount++
	}

	// create register input corresponding to the register "saved"
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		aig.Aig_ObjCreateCo(pNew, pObjSaveOrSaved)
		nRegCount++

		//Changed on October 13, 2009
		pObjAndAcc := aig.Aig_ManConst1(pNew)

		// create the family of shadow registers, then create the cascade of Xnor and And gates for the comparator
		for i := int32(0); i < saigRegNum(p); i++ {
			pObj := saig.Saig_ManLo(p, i)
			pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

			vecLos = append(vecLos, pObjShadowLo)
			vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkCi(pNtk, abc.Abc_NtkPiNum(pNtk)+i)), "SHADOW"))

			pObjShadowLiDriver := aigMux(pNew, pObjSaveAndNotSaved, aigData(p, pObj), pObjShadowLo)
			aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
			nRegCount++

			pObjXor := aig.Aig_Exor(pNew, aigData(p, pObj), pObjShadowLo)
			pObjXnor := aig.Aig_Not(pObjXor)

			pObjAndAcc = aig.Aig_And(pNew, pObjXnor, pObjAndAcc)
		}

		// create the AND gate whose output will be the signal "looped"
		pObjSavedLoAndEquality := aig.Aig_And(pNew, pObjSavedLo, pObjAndAcc)

		// create the master AND gate and corresponding AND and OR logic for the liveness properties
		pObjAndAcc = aig.Aig_ManConst1(pNew)
		if len(vLive) == 0 {
			fmt.Printf("Circuit without any liveness property\n")
		} else {
			for _, pObj := range vLive {
				liveLatch++
				pDriverImage := driverNew(p, pObj)
				pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

				vecLos = append(vecLos, pObjShadowLo)
				vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, getPoIndex(p, pObj))), "LIVENESS"))

				pObjShadowLiDriver := aig.Aig_Or(pNew, pObjShadowLo, aig.Aig_And(pNew, pDriverImage, pObjSaveOrSaved))
				aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
				nRegCount++

				pObjAndAcc = aig.Aig_And(pNew, pObjShadowLo, pObjAndAcc)
			}
		}

		pObjLive := pObjAndAcc

		pObjAndAcc = aig.Aig_ManConst1(pNew)
		if len(vFair) == 0 {
			fmt.Printf("Circuit without any fairness property\n")
		} else {
			for _, pObj := range vFair {
				fairLatch++
				pDriverImage := driverNew(p, pObj)
				pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

				vecLos = append(vecLos, pObjShadowLo)
				vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, getPoIndex(p, pObj))), "FAIRNESS"))

				pObjShadowLiDriver := aig.Aig_Or(pNew, pObjShadowLo, aig.Aig_And(pNew, pDriverImage, pObjSaveOrSaved))
				aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
				nRegCount++

				pObjAndAcc = aig.Aig_And(pNew, pObjShadowLo, pObjAndAcc)
			}
		}

		pObjFair := pObjAndAcc

		//Following is the actual Biere translation
		pObjSafetyGate := aig.Aig_And(pNew, pObjSavedLoAndEquality, aig.Aig_And(pNew, pObjFair, aig.Aig_Not(pObjLive)))

		aig.Aig_ObjPatchFanin0(pNew, pObjSafetyPropertyOutput, pObjSafetyGate)
	}

	aig.Aig_ManSetRegNum(pNew, nRegCount)

	Aig_ManCiCleanupBiere(pNew)
	Aig_ManCoCleanupBiere(pNew)

	aig.Aig_ManCleanup(pNew)

	_, _ = liveLatch, fairLatch
	return pNew
}

func LivenessToSafetyTransformationAbs(mode int, pNtk *Abc_Ntk_t, p *Aig_Man_t, vFlops []int32,
	vLive, vFair, vAssertSafety, vAssumeSafety []Aig_Obj_t) *Aig_Man_t {
	var pObjSavePi, pObjSavedLo Aig_Obj_t = aigNull, aigNull
	var pObjSaveOrSaved, pObjSaveAndNotSaved Aig_Obj_t = aigNull, aigNull
	var pObjSafetyPropertyOutput Aig_Obj_t = aigNull
	var nRegCount int32

	vecPis = make([]Aig_Obj_t, 0, saigPiNum(p)+1)
	vecPiNames = make([]string, 0, saigPiNum(p)+1)
	vecLos = make([]Aig_Obj_t, 0, saigRegNum(p)+int32(len(vFlops))+1+int32(len(vLive)+len(vFair)))
	vecLoNames = make([]string, 0, saigRegNum(p)+int32(len(vFlops))+1+int32(len(vLive)+len(vFair)))

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNew := aig.Aig_ManStart(2 * aig.Aig_ManObjNumMax(p))
	pNew.Name = fmt.Sprintf("%s_%s", pNtk.Name, "l2s")
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
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSavePi = aig.Aig_ObjCreateCi(pNew)
		vecPiNames = append(vecPiNames, "SAVE_BIERE")
	}

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
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSavedLo = aig.Aig_ObjCreateCi(pNew)
		vecLos = append(vecLos, pObjSavedLo)
		vecLoNames = append(vecLoNames, "SAVED_LO")
	}

	//****************************************************************
	// Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
	//****************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSaveOrSaved = aig.Aig_Or(pNew, pObjSavePi, pObjSavedLo)
		pObjSaveAndNotSaved = aig.Aig_And(pNew, pObjSavePi, aig.Aig_Not(pObjSavedLo))
	}

	//********************************************************************
	// Step 8: create internal nodes
	//********************************************************************
	nodes, pObj := aigManNodes(p)
	for _, pNode := range nodes {
		aigSetData(p, pNode, aig.Aig_And(pNew, aig.Aig_ObjChild0Copy(p, pNode), aig.Aig_ObjChild1Copy(p, pNode)))
	}

	//********************************************************************
	// Step 8.x : create PO for each safety assertions
	//********************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
		if len(vAssertSafety) != 0 && len(vAssumeSafety) == 0 {
			pObjAndAcc := andOfDrivers(p, pNew, vAssertSafety)
			pObj = vAssertSafety[len(vAssertSafety)-1]
			aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(pObjAndAcc))
		} else if len(vAssertSafety) != 0 && len(vAssumeSafety) != 0 {
			collectiveAssertSafety := andOfDrivers(p, pNew, vAssertSafety)
			collectiveAssumeSafety := andOfDrivers(p, pNew, vAssumeSafety)
			pObj = vAssumeSafety[len(vAssumeSafety)-1]
			aig.Aig_ObjCreateCo(pNew, aig.Aig_And(pNew, aig.Aig_Not(collectiveAssertSafety), collectiveAssumeSafety))
		} else {
			fmt.Printf("WARNING!! No safety property is found, a new (negated) constant 1 output is created\n")
			aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(aig.Aig_ManConst1(pNew)))
		}
	}

	//********************************************************************
	// Step 9: create the safety property output gate for the liveness properties
	//********************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSafetyPropertyOutput = aig.Aig_ObjCreateCo(pNew, aigData(p, aig.Aig_ObjFanin0(p, pObj)))
	}

	// create register inputs for the original registers
	nRegCount = 0
	for i := int32(0); i < saigRegNum(p); i++ {
		pMatch := saig.Saig_ObjLoToLi(p, saig.Saig_ManLo(p, i))
		aig.Aig_ObjCreateCo(pNew, driverNew(p, pMatch))
		nRegCount++
	}

	// create register input corresponding to the register "saved"
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		aig.Aig_ObjCreateCo(pNew, pObjSaveOrSaved)
		nRegCount++

		pObjAndAcc := aig.Aig_ManConst1(pNew)

		// create the family of shadow registers, then create the cascade of Xnor and And gates for the comparator
		for i := int32(0); i < saigRegNum(p); i++ {
			fmt.Printf("Flop[%d] = %s\n", i, abc.Abc_ObjName(abc.Abc_NtkCi(pNtk, abc.Abc_NtkPiNum(pNtk)+i)))
		}
		for _, iEntry := range vFlops {
			pObjShadowLo := aig.Aig_ObjCreateCi(pNew)
			pObj := aig.Aig_ManLo(p, iEntry)

			vecLos = append(vecLos, pObjShadowLo)
			nodeName := fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkCi(pNtk, abc.Abc_NtkPiNum(pNtk)+iEntry)), "SHADOW")
			fmt.Printf("Flop copied [%d] = %s\n", iEntry, nodeName)
			vecLoNames = append(vecLoNames, nodeName)

			pObjShadowLiDriver := aigMux(pNew, pObjSaveAndNotSaved, aigData(p, pObj), pObjShadowLo)
			aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
			nRegCount++

			pObjXor := aig.Aig_Exor(pNew, aigData(p, pObj), pObjShadowLo)
			pObjXnor := aig.Aig_Not(pObjXor)

			pObjAndAcc = aig.Aig_And(pNew, pObjXnor, pObjAndAcc)
		}

		// create the AND gate whose output will be the signal "looped"
		pObjSavedLoAndEquality := aig.Aig_And(pNew, pObjSavedLo, pObjAndAcc)

		// create the master AND gate and corresponding AND and OR logic for the liveness properties
		pObjAndAcc = aig.Aig_ManConst1(pNew)
		if len(vLive) == 0 {
			fmt.Printf("Circuit without any liveness property\n")
		} else {
			for _, pObj := range vLive {
				pDriverImage := driverNew(p, pObj)
				pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

				vecLos = append(vecLos, pObjShadowLo)
				vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, getPoIndex(p, pObj))), "LIVENESS"))

				pObjShadowLiDriver := aig.Aig_Or(pNew, pObjShadowLo, aig.Aig_And(pNew, pDriverImage, pObjSaveOrSaved))
				aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
				nRegCount++

				pObjAndAcc = aig.Aig_And(pNew, pObjShadowLo, pObjAndAcc)
			}
		}

		pObjLive := pObjAndAcc

		pObjAndAcc = aig.Aig_ManConst1(pNew)
		if len(vFair) == 0 {
			fmt.Printf("Circuit without any fairness property\n")
		} else {
			for _, pObj := range vFair {
				pDriverImage := driverNew(p, pObj)
				pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

				vecLos = append(vecLos, pObjShadowLo)
				vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, getPoIndex(p, pObj))), "FAIRNESS"))

				pObjShadowLiDriver := aig.Aig_Or(pNew, pObjShadowLo, aig.Aig_And(pNew, pDriverImage, pObjSaveOrSaved))
				aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
				nRegCount++

				pObjAndAcc = aig.Aig_And(pNew, pObjShadowLo, pObjAndAcc)
			}
		}

		pObjFair := pObjAndAcc

		//Following is the actual Biere translation
		pObjSafetyGate := aig.Aig_And(pNew, pObjSavedLoAndEquality, aig.Aig_And(pNew, pObjFair, aig.Aig_Not(pObjLive)))

		aig.Aig_ObjPatchFanin0(pNew, pObjSafetyPropertyOutput, pObjSafetyGate)
	}

	aig.Aig_ManSetRegNum(pNew, nRegCount)

	Aig_ManCiCleanupBiere(pNew)
	Aig_ManCoCleanupBiere(pNew)

	aig.Aig_ManCleanup(pNew)

	return pNew
}

// conjunction of the CO drivers, starting from the first (NULL when empty)
func andOfDriversNull(p, pNew *Aig_Man_t, vObjs []Aig_Obj_t) Aig_Obj_t {
	pObjAndAcc := aigNull
	for _, pObj := range vObjs {
		pArgument := driverNew(p, pObj)
		if pObjAndAcc == aigNull {
			pObjAndAcc = pArgument
		} else {
			pObjAndAccDummy := pObjAndAcc
			pObjAndAcc = aig.Aig_And(pNew, pArgument, pObjAndAccDummy)
		}
	}
	return pObjAndAcc
}

func LivenessToSafetyTransformationOneStepLoop(mode int, pNtk *Abc_Ntk_t, p *Aig_Man_t,
	vLive, vFair, vAssertSafety, vAssumeSafety []Aig_Obj_t) *Aig_Man_t {
	var pObjSavePi Aig_Obj_t = aigNull
	var pObjSafetyPropertyOutput Aig_Obj_t = aigNull
	var nRegCount int32

	if aig.Aig_ManRegNum(p) == 0 {
		fmt.Printf("The input AIG contains no register, returning the original AIG as it is\n")
		return p
	}

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
	if mode == FULL_BIERE_ONE_LOOP_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE {
		pObjSavePi = aig.Aig_ObjCreateCi(pNew)
		vecPiNames = append(vecPiNames, "SAVE_BIERE")
	}

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

	if mode == FULL_BIERE_ONE_LOOP_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
		if len(vAssertSafety) != 0 && len(vAssumeSafety) == 0 {
			pObjAndAcc := andOfDriversNull(p, pNew, vAssertSafety)
			pObj = vAssertSafety[len(vAssertSafety)-1]
			aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(pObjAndAcc))
		} else if len(vAssertSafety) != 0 && len(vAssumeSafety) != 0 {
			collectiveAssertSafety := andOfDriversNull(p, pNew, vAssertSafety)
			collectiveAssumeSafety := andOfDriversNull(p, pNew, vAssumeSafety)
			pObj = vAssumeSafety[len(vAssumeSafety)-1]
			aig.Aig_ObjCreateCo(pNew, aig.Aig_And(pNew, aig.Aig_Not(collectiveAssertSafety), collectiveAssumeSafety))
		} else {
			fmt.Printf("No safety property is specified, hence no safety gate is created\n")
		}
	}

	//********************************************************************
	// Step 9: create the safety property output gate
	// create the safety property output gate, this will be the sole true PO
	// of the whole circuit, discuss with Sat/Alan for an alternative implementation
	//********************************************************************
	if mode == FULL_BIERE_ONE_LOOP_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE {
		pObjSafetyPropertyOutput = aig.Aig_ObjCreateCo(pNew, aigData(p, aig.Aig_ObjFanin0(p, pObj)))
	}

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

	if mode == FULL_BIERE_ONE_LOOP_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE {
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

		// create the AND gate whose output will be the signal "looped"
		pObjSavedLoAndEquality := aig.Aig_And(pNew, pObjSavePi, pObjAndAcc)

		// create the master AND gate and corresponding AND and OR logic for the liveness properties
		pObjAndAcc = aigNull
		if len(vLive) == 0 {
			fmt.Printf("Circuit without any liveness property\n")
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
			fmt.Printf("Circuit without any fairness property\n")
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
	}

	aig.Aig_ManSetRegNum(pNew, nRegCount)

	Aig_ManCiCleanupBiere(pNew)
	Aig_ManCoCleanupBiere(pNew)

	aig.Aig_ManCleanup(pNew)

	return pNew
}

// the COs of the POs whose names pass the test
func populatePoVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, match func(*Abc_Obj_t) bool) []Aig_Obj_t {
	v := make([]Aig_Obj_t, 0, 100)
	for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
		if match(abc.Abc_NtkPo(pNtk, i)) {
			v = append(v, aig.Aig_ManCo(pAig, i))
		}
	}
	return v
}

func populateLivenessVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) []Aig_Obj_t {
	vLive := populatePoVector(pNtk, pAig, func(pNode *Abc_Obj_t) bool { return nodeName_starts_with(pNode, "assert_fair") })
	fmt.Printf("Number of liveness property found = %d\n", len(vLive))
	return vLive
}

func populateFairnessVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) []Aig_Obj_t {
	vFair := populatePoVector(pNtk, pAig, func(pNode *Abc_Obj_t) bool { return nodeName_starts_with(pNode, "assume_fair") })
	fmt.Printf("Number of fairness property found = %d\n", len(vFair))
	return vFair
}

func populateSafetyAssertionVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) []Aig_Obj_t {
	vAssertSafety := populatePoVector(pNtk, pAig, func(pNode *Abc_Obj_t) bool {
		return nodeName_starts_with(pNode, "assert_safety") || nodeName_starts_with(pNode, "Assert")
	})
	fmt.Printf("Number of safety property found = %d\n", len(vAssertSafety))
	return vAssertSafety
}

func populateSafetyAssumptionVector(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) []Aig_Obj_t {
	vAssumeSafety := populatePoVector(pNtk, pAig, func(pNode *Abc_Obj_t) bool {
		return nodeName_starts_with(pNode, "assume_safety") || nodeName_starts_with(pNode, "Assume")
	})
	fmt.Printf("Number of assume_safety property found = %d\n", len(vAssumeSafety))
	return vAssumeSafety
}

// storeCiNames stores the PI and LO names as liveness.c and
// liveness_sim.c do, in a new name manager (pNtk->pManName =
// Nm_ManCreate(...)), with the AIG object type AIG_OBJ_CI.
func storeCiNames(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, vPiNames, vLoNames []string) {
	abc.Abc_NtkResetNames(pNtk)
	pManName := abc.Abc_NtkReadNameManager(pNtk)
	if vPiNames != nil {
		for i := int32(0); i < saigPiNum(pAig); i++ {
			ntkObjId := abc.Abc_ObjId(abc.Abc_NtkCi(pNtk, i))
			nm.Nm_ManStoreIdName(pManName, ntkObjId, int32(aig.AIG_OBJ_CI), vPiNames[i], "")
		}
	}
	if vLoNames != nil {
		for i := int32(0); i < saigRegNum(pAig); i++ {
			ntkObjId := abc.Abc_ObjId(abc.Abc_NtkCi(pNtk, saigPiNum(pAig)+i))
			nm.Nm_ManStoreIdName(pManName, ntkObjId, int32(aig.AIG_OBJ_CI), vLoNames[i], "")
		}
	}
}

func updateNewNetworkNameManager(pNtk *Abc_Ntk_t, pAig *Aig_Man_t, vPiNames, vLoNames []string) {
	storeCiNames(pNtk, pAig, vPiNames, vLoNames)
	pManName := abc.Abc_NtkReadNameManager(pNtk)

	for i := int32(0); i < abc.Abc_NtkPoNum(pNtk); i++ {
		pNode := abc.Abc_NtkPo(pNtk, i)
		abc.Abc_ObjAssignName(pNode, "assert_safety_", abc.Abc_ObjName(pNode))
	}

	// assign latch input names
	for _, pNode := range pNtk.ObjectsOfKind(abc.ObjectLatch) {
		pFanin := abc.Abc_ObjFanin0(pNode)
		if nm.Nm_ManFindNameById(pManName, abc.Abc_ObjId(pFanin)) == "" {
			abc.Abc_ObjAssignName(pFanin, abc.Abc_ObjName(pFanin), "")
		}
	}
}

// liveOptions is the option loop shared by l2s, l3s and the abstraction
// command (with "1slh", or "1slhf" for l3s); usage is false for -h and
// for a combination of options the C loop rejects.
func liveOptions(argc int32, argv []string, spec string) (directive int, ok bool) {
	directive = -1
	if argc == 1 {
		return FULL_BIERE_MODE, true
	}
	extra.Extra_UtilGetoptReset()
	for {
		c := extra.Extra_UtilGetopt(argc, argv, spec)
		if c == -1 {
			break
		}
		switch c {
		case '1':
			if directive == -1 {
				directive = FULL_BIERE_ONE_LOOP_MODE
			} else {
				if directive == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
					directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE
				} else {
					directive = IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE
				}
			}
		case 's':
			if directive == -1 {
				directive = IGNORE_SAFETY_KEEP_LIVENESS_MODE
			} else {
				if directive != FULL_BIERE_ONE_LOOP_MODE {
					return directive, false
				}
				directive = IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE
			}
		case 'l':
			if directive == -1 {
				directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE
			} else {
				if directive != FULL_BIERE_ONE_LOOP_MODE {
					return directive, false
				}
				directive = IGNORE_LIVENESS_KEEP_SAFETY_MODE
			}
		case 'f':
			fmt.Printf("\nILLEGAL FLAG: aborting....\n")
			exit0()
		default: // 'h' and unknown options
			return directive, false
		}
	}
	return directive, true
}

// liveInputs reads the current network and collects the property outputs
// as the three commands do; pAig is nil for an empty network.
func liveInputs(pAbc *Abc_Frame_t, pNtk *Abc_Ntk_t) (pAig *Aig_Man_t, vLive, vFair, vAssertSafety, vAssumeSafety []Aig_Obj_t) {
	if !(abc.Abc_NtkIsStrash(pNtk) != 0) {
		fmt.Printf("The input network was not strashed, strashing....\n")
		pNtkTemp := Abc_NtkStrash(pNtk, 0, 0, 0)
		pAig = Abc_NtkToDar(pNtkTemp, 0, 1)
	} else {
		pAig = Abc_NtkToDar(pNtk, 0, 1)
	}
	vLive = populateLivenessVector(pNtk, pAig)
	vFair = populateFairnessVector(pNtk, pAig)
	vAssertSafety = populateSafetyAssertionVector(pNtk, pAig)
	vAssumeSafety = populateSafetyAssumptionVector(pNtk, pAig)
	return
}

// installLiveNetwork converts the new AIG and makes it the current network.
func installLiveNetwork(pAbc *Abc_Frame_t, pAigNew *Aig_Man_t, setName bool) {
	pNtkNew := Abc_NtkFromAigPhase(pAigNew)
	if setName {
		pNtkNew.Name = pAigNew.Name
	}

	if abc.Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Printf("Abc_NtkCreateCone(): Network check has failed.\n")
	}

	updateNewNetworkNameManager(pNtkNew, pAigNew, vecPiNames, vecLoNames)
	saveLiveNames(pAbc)
	abcmain.Abc_FrameSetCurrentNetwork(pAbc, pNtkNew)
}

func Abc_CommandAbcLivenessToSafety(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pAigNew *Aig_Man_t

	pNtk := frameNtk(pAbc)
	pErr := abcmain.Abc_FrameReadErr(pAbc)

	directive, ok := liveOptions(argc, argv, "1slh")
	if !ok {
		goto usage
	}

	if pNtk == nil {
		fmt.Fprintf(pErr, "Empty network.\n")
		return 1
	}
	{
		loadLiveNames(pAbc)
		pAig, vLive, vFair, vAssertSafety, vAssumeSafety := liveInputs(pAbc, pNtk)

		switch directive {
		case FULL_BIERE_MODE:
			pAigNew = LivenessToSafetyTransformation(FULL_BIERE_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n")
			}
		case FULL_BIERE_ONE_LOOP_MODE:
			pAigNew = LivenessToSafetyTransformationOneStepLoop(FULL_BIERE_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n")
			}
		case IGNORE_LIVENESS_KEEP_SAFETY_MODE:
			pAigNew = LivenessToSafetyTransformation(IGNORE_LIVENESS_KEEP_SAFETY_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t1 PO - only for safety property; liveness properties are ignored, if any.\n\tno additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n")
			}
		case IGNORE_SAFETY_KEEP_LIVENESS_MODE:
			pAigNew = LivenessToSafetyTransformation(IGNORE_SAFETY_KEEP_LIVENESS_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t1 PO - only for liveness property; safety properties are ignored, if any.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n")
			}
		case IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE:
			pAigNew = LivenessToSafetyTransformationOneStepLoop(IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("New circuit is produced ignoring safety outputs!\nOnly liveness and fairness outputs are considered.\nShadow registers are not created\n")
			}
		}

		installLiveNetwork(pAbc, pAigNew, true)
		return 0
	}

usage:
	fmt.Printf("usage: l2s [-1lsh]\n")
	fmt.Printf("\t         performs Armin Biere's live-to-safe transformation\n")
	fmt.Printf("\t-1 : no shadow logic, presume all loops are self loops\n")
	fmt.Printf("\t-l : ignore liveness and fairness outputs\n")
	fmt.Printf("\t-s : ignore safety assertions and assumptions\n")
	fmt.Printf("\t-h : print command usage\n")
	return 1
}

func prepareFlopVector(pAig *Aig_Man_t, vectorLength int32) []int32 {
	vFlops := make([]int32, 0, vectorLength)
	for i := int32(0); i < vectorLength; i++ {
		vFlops = append(vFlops, i)
	}
	return vFlops
}

// Abc_CommandAbcLivenessToSafetyAbstraction is not registered as a
// command in abc.c.
func Abc_CommandAbcLivenessToSafetyAbstraction(pAbc *Abc_Frame_t, argc int32, argv []string) int32 {
	var pAigNew *Aig_Man_t

	pNtk := frameNtk(pAbc)
	pErr := abcmain.Abc_FrameReadErr(pAbc)

	directive, ok := liveOptions(argc, argv, "1slh")
	if !ok {
		goto usage
	}

	if pNtk == nil {
		fmt.Fprintf(pErr, "Empty network.\n")
		return 1
	}
	{
		loadLiveNames(pAbc)
		pAig, vLive, vFair, vAssertSafety, vAssumeSafety := liveInputs(pAbc, pNtk)

		var nFlops int32
		if aig.Aig_ManRegNum(pAig)%2 == 0 {
			nFlops = aig.Aig_ManRegNum(pAig) / 2
		} else {
			nFlops = (aig.Aig_ManRegNum(pAig) - 1) / 2
		}
		vFlops := prepareFlopVector(pAig, nFlops)

		switch directive {
		case FULL_BIERE_MODE:
			pAigNew = LivenessToSafetyTransformationAbs(FULL_BIERE_MODE, pNtk, pAig, vFlops, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n")
			}
		case FULL_BIERE_ONE_LOOP_MODE:
			pAigNew = LivenessToSafetyTransformationOneStepLoop(FULL_BIERE_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n")
			}
		case IGNORE_LIVENESS_KEEP_SAFETY_MODE:
			pAigNew = LivenessToSafetyTransformationAbs(IGNORE_LIVENESS_KEEP_SAFETY_MODE, pNtk, pAig, vFlops, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t1 PO - only for safety property; liveness properties are ignored, if any.\n\tno additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n")
			}
		case IGNORE_SAFETY_KEEP_LIVENESS_MODE:
			pAigNew = LivenessToSafetyTransformationAbs(IGNORE_SAFETY_KEEP_LIVENESS_MODE, pNtk, pAig, vFlops, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t1 PO - only for liveness property; safety properties are ignored, if any.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n")
			}
		case IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE:
			pAigNew = LivenessToSafetyTransformationOneStepLoop(IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("New circuit is produced ignoring safety outputs!\nOnly liveness and fairness outputs are considered.\nShadow registers are not created\n")
			}
		}

		installLiveNetwork(pAbc, pAigNew, true)
		return 0
	}

usage:
	fmt.Printf("usage: l2s [-1lsh]\n")
	fmt.Printf("\t         performs Armin Biere's live-to-safe transformation\n")
	fmt.Printf("\t-1 : no shadow logic, presume all loops are self loops\n")
	fmt.Printf("\t-l : ignore liveness and fairness outputs\n")
	fmt.Printf("\t-s : ignore safety assertions and assumptions\n")
	fmt.Printf("\t-h : print command usage\n")
	return 1
}

func LivenessToSafetyTransformationWithLTL(mode int, pNtk *Abc_Ntk_t, p *Aig_Man_t,
	vLive, vFair, vAssertSafety, vAssumeSafety []Aig_Obj_t,
	numLtlProcessed *int, ltlBuffer []string) *Aig_Man_t {
	var pObjSavePi, pObjSavedLo Aig_Obj_t = aigNull, aigNull
	var pObjSaveOrSaved, pObjSaveAndNotSaved Aig_Obj_t = aigNull, aigNull
	var pNegatedSafetyConjunction Aig_Obj_t = aigNull
	var vTopASTNodeArray []*ltlNode
	var vPoForLtlProps []Aig_Obj_t
	var nRegCount int32

	vecPis = make([]Aig_Obj_t, 0, saigPiNum(p)+1)
	vecPiNames = make([]string, 0, saigPiNum(p)+1)
	vecLos = make([]Aig_Obj_t, 0, 2*saigRegNum(p)+1+int32(len(vLive)+len(vFair)))
	vecLoNames = make([]string, 0, 2*saigRegNum(p)+1+int32(len(vLive)+len(vFair)))

	//****************************************************************
	//step0: Parsing the LTL formula
	//****************************************************************

	//***************************************************************************
	//Reading input LTL formulae from Ntk data-structure and creating
	//AST for them, Steps involved:
	//        parsing -> AST creation -> well-formedness check -> signal name check
	//***************************************************************************

	//resetting numLtlProcessed
	*numLtlProcessed = 0

	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		vecInputLtlFormulae := ltlBuffer
		if vecInputLtlFormulae != nil {
			vTopASTNodeArray = make([]*ltlNode, 0, len(vecInputLtlFormulae))
			fmt.Printf("\n")
			for i, pFormula := range vecInputLtlFormulae {
				tempTopASTNode := parseFormulaCreateAST(pFormula)
				if tempTopASTNode != nil {
					fmt.Printf("Formula %d: AST is created, ", i+1)
					if isWellFormed(tempTopASTNode) != 0 {
						fmt.Printf("Well-formedness check PASSED, ")
					} else {
						fmt.Printf("Well-formedness check FAILED!!\n")
						fmt.Printf("AST will be ignored for formula %d, no extra logic will be added for this formula\n", i+1)
						continue
					}
					if checkSignalNameExistence(pNtk, tempTopASTNode) != 0 {
						fmt.Printf("Signal check PASSED\n")
					} else {
						fmt.Printf("Signal check FAILED!!")
						fmt.Printf("AST will be ignored for formula %d, no extra logic will be added for this formula\n", i+1)
						continue
					}
					vTopASTNodeArray = append(vTopASTNodeArray, tempTopASTNode)
					(*numLtlProcessed)++
				} else {
					fmt.Printf("\nNo AST has been created for formula %d, no extra logic will be added\n", i+1)
				}
			}
		}
		fmt.Printf("\n")
		// without an LTL database (vTopASTNodeArray is NULL; the network did
		// not come from "read") C crashed here: CBUGS_FIXED.md #55
		if len(vTopASTNodeArray) == 0 {
			fmt.Printf("\nCurrently aborting, need to take care when Vec_PtrSize( vTopASTNodeArray ) == 0\n")
			exit0()
		}
	}

	//****************************************************************
	// Step1: create the new manager
	//****************************************************************
	pNew := aig.Aig_ManStart(2 * aig.Aig_ManObjNumMax(p))
	pNew.Name = fmt.Sprintf("%s_%s", pNtk.Name, "l3s")
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
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSavePi = aig.Aig_ObjCreateCi(pNew)
		vecPiNames = append(vecPiNames, "SAVE_BIERE")
	}

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
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSavedLo = aig.Aig_ObjCreateCi(pNew)
		vecLos = append(vecLos, pObjSavedLo)
		vecLoNames = append(vecLoNames, "SAVED_LO")
	}

	//****************************************************************
	// Step 7: create the OR gate and the AND gate directly fed by "SAVE" Pi
	//****************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		pObjSaveOrSaved = aig.Aig_Or(pNew, pObjSavePi, pObjSavedLo)
		pObjSaveAndNotSaved = aig.Aig_And(pNew, pObjSavePi, aig.Aig_Not(pObjSavedLo))
	}

	//********************************************************************
	// Step 8: create internal nodes
	//********************************************************************
	nodes, pObj := aigManNodes(p)
	for _, pNode := range nodes {
		aigSetData(p, pNode, aig.Aig_And(pNew, aig.Aig_ObjChild0Copy(p, pNode), aig.Aig_ObjChild1Copy(p, pNode)))
	}

	//********************************************************************
	// Step 8.x : create PO for each safety assertions
	// NOTE : Here the output is purposely inverted as it will be thrown to
	// dprove
	//********************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		if len(vAssertSafety) != 0 && len(vAssumeSafety) == 0 {
			pObjAndAcc := andOfDrivers(p, pNew, vAssertSafety)
			pObj = vAssertSafety[len(vAssertSafety)-1]
			pNegatedSafetyConjunction = aig.Aig_Not(pObjAndAcc)
			if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
				aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(pObjAndAcc))
			}
		} else if len(vAssertSafety) != 0 && len(vAssumeSafety) != 0 {
			collectiveAssertSafety := andOfDrivers(p, pNew, vAssertSafety)
			collectiveAssumeSafety := andOfDrivers(p, pNew, vAssumeSafety)
			pObj = vAssumeSafety[len(vAssumeSafety)-1]
			pNegatedSafetyConjunction = aig.Aig_And(pNew, aig.Aig_Not(collectiveAssertSafety), collectiveAssumeSafety)
			if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
				aig.Aig_ObjCreateCo(pNew, aig.Aig_And(pNew, aig.Aig_Not(collectiveAssertSafety), collectiveAssumeSafety))
			}
		} else {
			fmt.Printf("WARNING!! No safety property is found, a new (negated) constant 1 output is created\n")
			pNegatedSafetyConjunction = aig.Aig_Not(aig.Aig_ManConst1(pNew))
			if mode == FULL_BIERE_MODE || mode == IGNORE_LIVENESS_KEEP_SAFETY_MODE {
				aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(aig.Aig_ManConst1(pNew)))
			}
		}
	}

	//********************************************************************
	// Step 9: create the safety property output gate for the liveness properties
	// discuss with Sat/Alan for an alternative implementation
	//********************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		vPoForLtlProps = make([]Aig_Obj_t, 0, len(vTopASTNodeArray))
		for i := 0; i < len(vTopASTNodeArray); i++ {
			pObjSafetyPropertyOutput := aig.Aig_ObjCreateCo(pNew, aigData(p, aig.Aig_ObjFanin0(p, pObj)))
			vPoForLtlProps = append(vPoForLtlProps, pObjSafetyPropertyOutput)
		}
	}

	//*************************************************************************************
	// Step 10: Placeholder PO's were created for Liveness property outputs in the
	// last step. FYI, # of new liveness property outputs = # of LTL properties in the circuit
	// It is time for creation of loop LI's and other stuff
	// Now creating register inputs for the original flops
	//*************************************************************************************
	nRegCount = 0
	for i := int32(0); i < saigRegNum(p); i++ {
		pMatch := saig.Saig_ObjLoToLi(p, saig.Saig_ManLo(p, i))
		aig.Aig_ObjCreateCo(pNew, driverNew(p, pMatch))
		nRegCount++
	}

	//*************************************************************************************
	// Step 11: create register input corresponding to the register "saved"
	//*************************************************************************************
	if mode == FULL_BIERE_MODE || mode == IGNORE_SAFETY_KEEP_LIVENESS_MODE {
		aig.Aig_ObjCreateCo(pNew, pObjSaveOrSaved)
		nRegCount++

		pObjAndAcc := aig.Aig_ManConst1(pNew)

		//*************************************************************************************
		// Step 11: create the family of shadow registers, then create the cascade of Xnor
		// and And gates for the comparator
		//*************************************************************************************
		for i := int32(0); i < saigRegNum(p); i++ {
			pObj := saig.Saig_ManLo(p, i)
			pObjShadowLo := aig.Aig_ObjCreateCi(pNew)

			vecLos = append(vecLos, pObjShadowLo)
			vecLoNames = append(vecLoNames, fmt.Sprintf("%s__%s", abc.Abc_ObjName(abc.Abc_NtkCi(pNtk, abc.Abc_NtkPiNum(pNtk)+i)), "SHADOW"))

			pObjShadowLiDriver := aigMux(pNew, pObjSaveAndNotSaved, aigData(p, pObj), pObjShadowLo)
			aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)
			nRegCount++

			pObjXor := aig.Aig_Exor(pNew, aigData(p, pObj), pObjShadowLo)
			pObjXnor := aig.Aig_Not(pObjXor)

			pObjAndAcc = aig.Aig_And(pNew, pObjXnor, pObjAndAcc)
		}

		// create the AND gate whose output will be the signal "looped"
		pObjSavedLoAndEquality := aig.Aig_And(pNew, pObjSavedLo, pObjAndAcc)

		//*************************************************************************************
		// Step 11: logic for LTL properties:- (looped & ~theta) where theta is the input ltl
		// property
		//*************************************************************************************

		var vSignalMemory, vGFFlopMemory []Aig_Obj_t

		for iii, topNodeOfAST := range vTopASTNodeArray {
			var vSignal []Aig_Obj_t
			var vAigGFMap [][]*ltlNode

			//*************************************************************************************
			//Step 11a: for the current AST, find out the leaf level Boolean signal pointers from
			// the NEW aig.
			//*************************************************************************************
			populateBoolWithAigNodePtr(pNtk, p, pNew, topNodeOfAST)

			//*************************************************************************************
			//Step 11b: for each GF node, compute the pointer in AIG that it should point to
			//*************************************************************************************
			populateAigPointerUnitGF(pNew, topNodeOfAST, &vSignal, &vAigGFMap)

			//*************************************************************************************
			//Step 11c: everything below GF are computed. Now, it is time to create logic for individual
			// GF nodes (i.e. the OR gate and the latch and the Boolean crown of the AST
			//*************************************************************************************
			for i, pObj := range vSignal {
				var pObjShadowLo Aig_Obj_t
				//*********************************************************
				// Step 11c.1: if the OR+Latch of the particular signal is
				// not already created, create it. It may have already been
				// created from another property, so check it before creation
				//*********************************************************
				if aigObjFind(vSignalMemory, pObj) == -1 {
					pDriverImage := pObj
					pObjShadowLo = aig.Aig_ObjCreateCi(pNew)
					pObjShadowLiDriver := aig.Aig_Or(pNew, pObjShadowLo, aig.Aig_And(pNew, pDriverImage, pObjSaveOrSaved))
					aig.Aig_ObjCreateCo(pNew, pObjShadowLiDriver)

					nRegCount++

					vSignalMemory = append(vSignalMemory, pObj)
					vGFFlopMemory = append(vGFFlopMemory, pObjShadowLo)

					vecLos = append(vecLos, pObjShadowLo)
					vecLoNames = append(vecLoNames, fmt.Sprintf("n%d__%s", aig.Aig_ObjId(pObjShadowLo), "GF_flop"))
				} else {
					pObjShadowLo = vGFFlopMemory[aigObjFind(vSignalMemory, pObj)]
				}

				if i < len(vAigGFMap) {
					for _, pEnrtyGLOBALLY := range vAigGFMap[i] {
						setAIGNodePtrOfGloballyNode(pEnrtyGLOBALLY, pObjShadowLo)
					}
				}
			}

			//*********************************************************
			//Step 11c.2: creating the Boolean crown
			//*********************************************************
			buildLogicFromLTLNode(pNew, topNodeOfAST)

			//*********************************************************
			//Step 11c.3: creating logic for (looped & ~theta) and patching
			// it with the proper PO
			//*********************************************************
			pObjLive := retriveAIGPointerFromLTLNode(topNodeOfAST)
			pObjSafetyGate := aig.Aig_And(pNew, pObjSavedLoAndEquality, aig.Aig_Not(pObjLive))
			fmt.Printf("liveness output is conjoined with safety assertions\n")
			pObjSafetyAndLiveToSafety := aig.Aig_Or(pNew, pObjSafetyGate, pNegatedSafetyConjunction)
			pObjSafetyPropertyOutput := vPoForLtlProps[iii]
			aig.Aig_ObjPatchFanin0(pNew, pObjSafetyPropertyOutput, pObjSafetyAndLiveToSafety)
		}
	}

	aig.Aig_ManSetRegNum(pNew, nRegCount)

	Aig_ManCiCleanupBiere(pNew)
	Aig_ManCoCleanupBiere(pNew)

	aig.Aig_ManCleanup(pNew)

	return pNew
}

// Vec_PtrFind
func aigObjFind(v []Aig_Obj_t, pObj Aig_Obj_t) int {
	for i, e := range v {
		if e == pObj {
			return i
		}
	}
	return -1
}

func Abc_CommandAbcLivenessToSafetyWithLTL(pAbc *Abc_Frame_t, argc int32, argv []string) (status int32) {
	var pAigNew *Aig_Man_t
	var numOfLtlPropOutput int
	defer recoverExit(&status)

	pNtk := frameNtk(pAbc)
	pErr := abcmain.Abc_FrameReadErr(pAbc)

	directive, ok := liveOptions(argc, argv, "1slhf")
	if !ok {
		goto usage
	}

	if pNtk == nil {
		fmt.Fprintf(pErr, "Empty network.\n")
		return 1
	}
	{
		loadLiveNames(pAbc)
		pAig, vLive, vFair, vAssertSafety, vAssumeSafety := liveInputs(pAbc, pNtk)

		ltlBuffer := Abc_FrameReadLTLProperties(pAbc)

		switch directive {
		case FULL_BIERE_MODE:
			pAigNew = LivenessToSafetyTransformationWithLTL(FULL_BIERE_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety, &numOfLtlPropOutput, ltlBuffer)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t%d POs - one for safety and %d for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n", numOfLtlPropOutput+1, numOfLtlPropOutput)
			}
		case FULL_BIERE_ONE_LOOP_MODE:
			pAigNew = LivenessToSafetyTransformationOneStepLoop(FULL_BIERE_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t2 POs - one for safety and one for liveness.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n")
			}
		case IGNORE_LIVENESS_KEEP_SAFETY_MODE:
			pAigNew = LivenessToSafetyTransformationWithLTL(IGNORE_LIVENESS_KEEP_SAFETY_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety, &numOfLtlPropOutput, ltlBuffer)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t1 PO - only for safety property; liveness properties are ignored, if any.\n\tno additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created\n\tnon-property POs are suppressed\n")
			}
		case IGNORE_SAFETY_KEEP_LIVENESS_MODE:
			pAigNew = LivenessToSafetyTransformationWithLTL(IGNORE_SAFETY_KEEP_LIVENESS_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety, &numOfLtlPropOutput, ltlBuffer)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("A new circuit is produced with\n\t%d PO - only for liveness property; safety properties are ignored, if any.\n\tone additional input is added (due to Biere's nondeterminism)\n\tshadow flops are not created if the original circuit is combinational\n\tnon-property POs are suppressed\n", numOfLtlPropOutput)
			}
		case IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE:
			pAigNew = LivenessToSafetyTransformationOneStepLoop(IGNORE_SAFETY_KEEP_LIVENESS_ONE_LOOP_MODE, pNtk, pAig, vLive, vFair, vAssertSafety, vAssumeSafety)
			if aig.Aig_ManRegNum(pAigNew) != 0 {
				fmt.Printf("New circuit is produced ignoring safety outputs!\nOnly liveness and fairness outputs are considered.\nShadow registers are not created\n")
			}
		}

		installLiveNetwork(pAbc, pAigNew, true)
		return 0
	}

usage:
	fmt.Printf("usage: l3s [-1lsh]\n")
	fmt.Printf("\t         performs Armin Biere's live-to-safe transformation\n")
	fmt.Printf("\t-1 : no shadow logic, presume all loops are self loops\n")
	fmt.Printf("\t-l : ignore liveness and fairness outputs\n")
	fmt.Printf("\t-s : ignore safety assertions and assumptions\n")
	fmt.Printf("\t-h : print command usage\n")
	return 1
}
