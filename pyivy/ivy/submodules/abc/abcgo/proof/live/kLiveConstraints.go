/**CFile****************************************************************

  FileName    [kLiveConstraints.c]

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
	"strings"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	abc "github.com/glycerine/abcgo/base/abc"
)

func createConstrained0LiveCone(pNewAig *Aig_Man_t, pOldAig *Aig_Man_t, signalList []Aig_Obj_t) Aig_Obj_t {
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

func collectCSSignals(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) []Aig_Obj_t {
	pConsequent := aigNull
	vNodeArray := make([]Aig_Obj_t, 0, 1)

	for i := int32(0); i < saigPoNum(pAig); i++ {
		pObj := aig.Aig_ManCo(pAig, i)
		if strings.Contains(abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)), "csLiveConst_") {
			vNodeArray = append(vNodeArray, driverOld(pAig, pObj))
		} else if strings.Contains(abc.Abc_ObjName(abc.Abc_NtkPo(pNtk, i)), "csLiveTarget_") {
			pConsequent = driverOld(pAig, pObj)
		}
	}
	vNodeArray = append(vNodeArray, pConsequent)
	return vNodeArray
}

func createNewAigWith0LivePo(pAig *Aig_Man_t, signalList []Aig_Obj_t, index0Live *int32) *Aig_Man_t {
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
	// Step 5: create register outputs
	//****************************************************************
	for i := int32(0); i < saigRegNum(pAig); i++ {
		aigSetData(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ObjCreateCi(pNewAig))
	}

	//********************************************************************
	// Step 7: create internal nodes
	//********************************************************************
	nodes, _ := aigManNodes(pAig)
	for _, pObj := range nodes {
		aigSetData(pAig, pObj, aig.Aig_And(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj), aig.Aig_ObjChild1Copy(pAig, pObj)))
	}

	var i int32
	for i = 0; i < saigPoNum(pAig); i++ {
		pObj := aig.Aig_ManCo(pAig, i)
		aigSetData(pAig, pObj, aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj)))
	}

	pObjNewPoDriver := createConstrained0LiveCone(pNewAig, pAig, signalList)
	aig.Aig_ObjCreateCo(pNewAig, pObjNewPoDriver)
	*index0Live = i

	for i := int32(0); i < saigRegNum(pAig); i++ {
		pObj := saig.Saig_ManLi(pAig, i)
		aigSetData(pAig, pObj, aig.Aig_ObjCreateCo(pNewAig, aig.Aig_ObjChild0Copy(pAig, pObj)))
	}

	aig.Aig_ManSetRegNum(pNewAig, aig.Aig_ManRegNum(pAig))
	aig.Aig_ManCleanup(pNewAig)

	return pNewAig
}

func checkMonotoneSignal() []Aig_Obj_t {
	return nil
}

func gatherMonotoneSignals(pAig *Aig_Man_t) []Aig_Obj_t {
	nodes, _ := aigManNodes(pAig)
	for _, pObj := range nodes {
		aig.Aig_ObjPrint(pAig, pObj)
		fmt.Printf("\n")
	}
	return nil
}

func generateWorkingAig(pAig *Aig_Man_t, pNtk *Abc_Ntk_t, pIndex0Live *int32) *Aig_Man_t {
	vSignalVector := collectCSSignals(pNtk, pAig)
	pAigNew := createNewAigWith0LivePo(pAig, vSignalVector, pIndex0Live)
	return pAigNew
}
