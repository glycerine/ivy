package lpk

import (
	"fmt"
	hop "github.com/glycerine/abcgo/aig/hop"
	abc "github.com/glycerine/abcgo/base/abc"
	kit "github.com/glycerine/abcgo/bool/kit"
	vec "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [lpkCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkCut.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "lpkInt.h"
//#include "bool/kit/cloud.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Computes the truth table of one cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_CutTruthBdd_rec(dd *CloudManager, pMan *Hop_Man_t, pObj *Hop_Obj_t, nVars int32) *CloudNode {
	var pTruth, pTruth0, pTruth1 *CloudNode
	assert(hop.Hop_IsComplement(pObj) == 0, "hop.Hop_IsComplement(pObj) == 0")
	if cached, ok := hop.Hop_ObjData(pObj).(*CloudNode); ok && cached != nil {
		return cached
	}
	// get the plan for a new truth table
	if hop.Hop_ObjIsConst1(pObj) != 0 {
		pTruth = kit.Cloud_One(dd)
	} else {
		assert(hop.Hop_ObjIsAnd(pObj), "hop.Hop_ObjIsAnd(pObj)")
		// compute the truth tables of the fanins
		pTruth0 = Lpk_CutTruthBdd_rec(dd, pMan, hop.Hop_ObjFanin0(pObj), nVars)
		pTruth1 = Lpk_CutTruthBdd_rec(dd, pMan, hop.Hop_ObjFanin1(pObj), nVars)
		pTruth0 = kit.Cloud_NotCond(pTruth0, hop.Hop_ObjFaninC0(pObj) != 0)
		pTruth1 = kit.Cloud_NotCond(pTruth1, hop.Hop_ObjFaninC1(pObj) != 0)
		// creat the truth table of the node
		pTruth = kit.Cloud_bddAnd(dd, pTruth0, pTruth1)
	}
	hop.Hop_ObjSetData(pObj, pTruth)
	return pTruth
}

/*
*Function*************************************************************

	Synopsis    [Verifies that the factoring is correct.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_CutTruthBdd(p *Lpk_Man_t, pCut *Lpk_Cut_t) *CloudNode {
	var dd *CloudManager = kit.Kit_DsdCloud(p.pDsdMan)
	pManHop, _ := p.pNtk.FunctionManager.(*hop.Hop_Man_t)
	var pObjHop *Hop_Obj_t
	var pObj, pFanin *Abc_Obj_t
	var pTruth *CloudNode = nil // Suppress "might be used uninitialized"
	var i, k int32

	//    return NULL;
	//    Lpk_NodePrintCut( p, pCut );
	// initialize the leaves
	// inlining of macro Lpk_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pObj = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[i])
		pObj.Copy = &abc.Object{Data: kit.Cloud_Var(dd, int32(pCut.nLeaves)-1-i)}
	}

	// construct truth table in the topological order
	// inlining of macro Lpk_CutForEachNodeReverse
	for i = int32(pCut.nNodes) - 1; i >= 0; i-- {
		pObj = abc.Abc_NtkObj(p.pNtk, pCut.pNodes[i])
		// get the local AIG
		pObjHop = hop.Hop_Regular(pObj.Data.(*hop.Hop_Obj_t))
		// clean the data field of the nodes in the AIG subgraph
		hop.Hop_ObjCleanData_rec(pObjHop)
		// set the initial truth tables at the fanins
		for k = 0; k < int32(len(pObj.Fanins())); k++ {
			pFanin = pObj.Fanins()[k].Object
			hop.Hop_ObjSetData(hop.Hop_ManPi(pManHop, k), pFanin.Copy.Data)
		}
		// compute the truth table of internal nodes
		pTruth = Lpk_CutTruthBdd_rec(dd, pManHop, pObjHop, int32(pCut.nLeaves))
		if hop.Hop_IsComplement(pObj.Data.(*hop.Hop_Obj_t)) != 0 {
			pTruth = kit.Cloud_Not(pTruth)
		}
		// set the truth table at the node
		pObj.Copy = &abc.Object{Data: pTruth}

	}

	//    Cloud_bddPrint( dd, pTruth );
	//    printf( "Bdd size = %d. Total nodes = %d.\n", Cloud_DagSize( dd, pTruth ), dd->nNodesCur-dd->nVars-1 );
	return pTruth
}

/*
*Function*************************************************************

	Synopsis    [Computes the truth table of one cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_CutTruth_rec(pMan *Hop_Man_t, pObj *Hop_Obj_t, nVars int32, vTtNodes *Vec_Ptr_t, piCount *int32) []uint32 {
	var pTruth, pTruth0, pTruth1 []uint32
	assert(hop.Hop_IsComplement(pObj) == 0, "hop.Hop_IsComplement(pObj) == 0")
	if cached, ok := hop.Hop_ObjData(pObj).([]uint32); ok && cached != nil {
		return cached
	}
	// get the plan for a new truth table
	pTruth = vec.Vec_PtrEntry(vTtNodes, int(*piCount)).([]uint32)
	*piCount++
	if hop.Hop_ObjIsConst1(pObj) != 0 {
		kit.Kit_TruthFill(pTruth, nVars)
	} else {
		assert(hop.Hop_ObjIsAnd(pObj), "hop.Hop_ObjIsAnd(pObj)")
		// compute the truth tables of the fanins
		pTruth0 = Lpk_CutTruth_rec(pMan, hop.Hop_ObjFanin0(pObj), nVars, vTtNodes, piCount)
		pTruth1 = Lpk_CutTruth_rec(pMan, hop.Hop_ObjFanin1(pObj), nVars, vTtNodes, piCount)
		// creat the truth table of the node
		kit.Kit_TruthAndPhase(pTruth, pTruth0, pTruth1, nVars, hop.Hop_ObjFaninC0(pObj), hop.Hop_ObjFaninC1(pObj))
	}
	hop.Hop_ObjSetData(pObj, pTruth)
	return pTruth
}

/*
*Function*************************************************************

	Synopsis    [Computes the truth able of one cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_CutTruth(p *Lpk_Man_t, pCut *Lpk_Cut_t, fInv int32) []uint32 {
	pManHop, _ := p.pNtk.FunctionManager.(*hop.Hop_Man_t)
	var pObjHop *Hop_Obj_t
	var pObj *Abc_Obj_t = nil // Suppress "might be used uninitialized"
	var pFanin *Abc_Obj_t
	var pTruth []uint32 = nil // Suppress "might be used uninitialized"
	var i, k, iCount int32
	//    Lpk_NodePrintCut( p, pCut );
	assert(pCut.nNodes > 0, "pCut.nNodes > 0")

	// initialize the leaves
	// inlining of macro Lpk_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pObj = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[i])
		pObj.Copy = &abc.Object{Data: vec.Vec_PtrEntry(p.vTtElems, func() int {
			if fInv != 0 {
				return int(pCut.nLeaves) - 1 - int(i)
			}
			return int(i)
		}())}
	}

	// construct truth table in the topological order
	// inlining of macro Lpk_CutForEachNodeReverse
	for i = int32(pCut.nNodes) - 1; i >= 0; i-- {
		pObj = abc.Abc_NtkObj(p.pNtk, pCut.pNodes[i])
		// get the local AIG
		pObjHop = hop.Hop_Regular(pObj.Data.(*hop.Hop_Obj_t))
		// clean the data field of the nodes in the AIG subgraph
		hop.Hop_ObjCleanData_rec(pObjHop)
		// set the initial truth tables at the fanins
		for k = 0; k < int32(len(pObj.Fanins())); k++ {
			pFanin = pObj.Fanins()[k].Object
			hop.Hop_ObjSetData(hop.Hop_ManPi(pManHop, k), pFanin.Copy.Data)
		}
		// compute the truth table of internal nodes
		pTruth = Lpk_CutTruth_rec(pManHop, pObjHop, int32(pCut.nLeaves), p.vTtNodes, &iCount)
		if hop.Hop_IsComplement(pObj.Data.(*hop.Hop_Obj_t)) != 0 {
			kit.Kit_TruthNot(pTruth, pTruth, int32(pCut.nLeaves))
		}
		// set the truth table at the node
		pObj.Copy = &abc.Object{Data: pTruth}
	}

	// make sure direct truth table is stored elsewhere (assuming the first call for direct truth!!!)
	if fInv == 0 {
		pTruth = vec.Vec_PtrEntry(p.vTtNodes, int(iCount)).([]uint32)
		iCount++
		kit.Kit_TruthCopy(pTruth, pObj.Copy.Data.([]uint32), int32(pCut.nLeaves))
	}
	assert(int(iCount) <= vec.Vec_PtrSize(p.vTtNodes), "iCount <= vec.Vec_PtrSize(p.vTtNodes)")
	return pTruth
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if at least one entry has changed.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodeRecordImpact(p *Lpk_Man_t) {
	var pCut *Lpk_Cut_t
	var vNodes *Vec_Ptr_t = vec.Vec_VecEntry(p.vVisited, p.pObj.ID)
	var pNode *Abc_Obj_t
	var i, k int32
	// collect the nodes that impact the given node
	vec.Vec_PtrClear(vNodes)
	for i = 0; i < p.nCuts; i++ {
		pCut = &p.pCuts[i]
		for k = 0; k < int32(pCut.nLeaves); k++ {
			pNode = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[k])
			if pNode.MarkC {
				continue
			}
			pNode.MarkC = true
			vec.Vec_PtrPush(vNodes, pNode.ID)
			vec.Vec_PtrPush(vNodes, abc.Abc_ObjFanoutNum(pNode))
		}
	}
	// clear the marks
	for i = 0; i+1 < int32(vec.Vec_PtrSize(vNodes)); i += 2 {
		pNode = abc.Abc_NtkObj(p.pNtk, int32(vec.Vec_PtrEntry(vNodes, int(i)).(int)))
		pNode.MarkC = false
	}
	// printf( "%d ", Vec_PtrSize(vNodes) );
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the cut has structural DSD.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodeCutsCheckDsd(p *Lpk_Man_t, pCut *Lpk_Cut_t) int32 {
	var pObj, pFanin *Abc_Obj_t
	var i, k, nCands, fLeavesOnly, RetValue int32
	assert(pCut.nLeaves > 0, "pCut.nLeaves > 0")
	// clear ref counters
	clear(p.pRefs[:])
	// mark cut leaves
	// inlining of macro Lpk_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pObj = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[i])
		assert(!pObj.MarkA, "pObj.MarkA == 0")
		pObj.MarkA = true
		pObj.Temp = int(i)
	}
	// ref leaves pointed from the internal nodes
	nCands = 0
	// inlining of macro Lpk_CutForEachNode
	for i = 0; i < int32(pCut.nNodes); i++ {
		pObj = abc.Abc_NtkObj(p.pNtk, pCut.pNodes[i])
		fLeavesOnly = 1
		for k = 0; k < int32(len(pObj.Fanins())); k++ {
			pFanin = pObj.Fanins()[k].Object
			if pFanin.MarkA {
				p.pRefs[pFanin.Temp]++
			} else {
				fLeavesOnly = 0
			}
		}
		if fLeavesOnly != 0 {
			p.pCands[nCands] = int32(pObj.ID)
			nCands++
		}
	}
	// look at the nodes that only point to the leaves
	RetValue = 0
	for i = 0; i < nCands; i++ {
		pObj = abc.Abc_NtkObj(p.pNtk, p.pCands[i])
		for k = 0; k < int32(len(pObj.Fanins())); k++ {
			pFanin = pObj.Fanins()[k].Object
			assert(pFanin.MarkA, "pFanin.MarkA == 1")
			if p.pRefs[int32(pFanin.Temp)] > 1 {
				break
			}
		}
		if k == abc.Abc_ObjFaninNum(pObj) {
			RetValue = 1
			break
		}
	}
	// unmark cut leaves
	// inlining of macro Lpk_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pObj = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[i])
		pObj.MarkA = false
	}
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if pDom is contained in pCut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodeCutsOneDominance(pDom *Lpk_Cut_t, pCut *Lpk_Cut_t) int32 {
	var i, k int32
	for i = 0; i < int32(pDom.nLeaves); i++ {
		for k = 0; k < int32(pCut.nLeaves); k++ {
			if pDom.pLeaves[i] == pCut.pLeaves[k] {
				break
			}
		}
		if k == int32(pCut.nLeaves) { // node i in pDom is not contained in pCut
			return 0
		}
	}
	// every node in pDom is contained in pCut
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Check if the cut exists.]

	Description [Returns 1 if the cut exists.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodeCutsOneFilter(pCuts []Lpk_Cut_t, nCuts int32, pCutNew *Lpk_Cut_t) int32 {
	var pCut *Lpk_Cut_t
	var i, k int32
	assert(pCutNew.uSign[0] != 0 || pCutNew.uSign[1] != 0, "pCutNew.uSign[0] || pCutNew.uSign[1]")
	// try to find the cut
	for i = 0; i < nCuts; i++ {
		pCut = &pCuts[i]
		if pCut.nLeaves == 0 {
			continue
		}
		if pCut.nLeaves == pCutNew.nLeaves {
			if pCut.uSign[0] == pCutNew.uSign[0] && pCut.uSign[1] == pCutNew.uSign[1] {
				for k = 0; k < int32(pCutNew.nLeaves); k++ {
					if pCut.pLeaves[k] != pCutNew.pLeaves[k] {
						break
					}
				}
				if k == int32(pCutNew.nLeaves) {
					return 1
				}
			}
			continue
		}
		if pCut.nLeaves < pCutNew.nLeaves {
			// skip the non-contained cuts
			if (pCut.uSign[0] & pCutNew.uSign[0]) != pCut.uSign[0] {
				continue
			}
			if (pCut.uSign[1] & pCutNew.uSign[1]) != pCut.uSign[1] {
				continue
			}
			// check containment seriously
			if Lpk_NodeCutsOneDominance(pCut, pCutNew) != 0 {
				return 1
			}
			continue
		}
		// check potential containment of other cut

		// skip the non-contained cuts
		if (pCut.uSign[0] & pCutNew.uSign[0]) != pCutNew.uSign[0] {
			continue
		}
		if (pCut.uSign[1] & pCutNew.uSign[1]) != pCutNew.uSign[1] {
			continue
		}
		// check containment seriously
		if Lpk_NodeCutsOneDominance(pCutNew, pCut) != 0 {
			pCut.nLeaves = 0
		} // removed
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Prints the given cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodePrintCut(p *Lpk_Man_t, pCut *Lpk_Cut_t, fLeavesOnly int32) {
	var pObj *Abc_Obj_t
	var i int32
	if !(fLeavesOnly != 0) {
		fmt.Printf("LEAVES:\n")
	}
	// inlining of macro Lpk_CutForEachLeaf
	for i = 0; i < int32(pCut.nLeaves); i++ {
		pObj = abc.Abc_NtkObj(p.pNtk, pCut.pLeaves[i])
		fmt.Printf("%d,", pObj.ID)
	}
	if !(fLeavesOnly != 0) {
		fmt.Printf("\nNODES:\n")
		// inlining of macro Lpk_CutForEachNode
		for i = 0; i < int32(pCut.nNodes); i++ {
			pObj = abc.Abc_NtkObj(p.pNtk, pCut.pNodes[i])
			fmt.Printf("%d,", pObj.ID)
			assert(abc.Abc_ObjIsNode(pObj), "abc.Abc_ObjIsNode(pObj)")
		}
		fmt.Printf("\n")
	}
}

/*
*Function*************************************************************

	Synopsis    [Set the cut signature.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodeCutSignature(pCut *Lpk_Cut_t) {
	var i uint32
	pCut.uSign[0] = 0
	pCut.uSign[1] = 0
	for i = 0; i < pCut.nLeaves; i++ {
		pCut.uSign[(pCut.pLeaves[i]>>5)&1] |= (1 << (pCut.pLeaves[i] & 31))
		if i != pCut.nLeaves-1 {
			assert(pCut.pLeaves[i] < pCut.pLeaves[i+1], "pCut.pLeaves[i] < pCut.pLeaves[i+1]")
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Computes the set of all cuts.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodeCutsOne(p *Lpk_Man_t, pCut *Lpk_Cut_t, Node int32) {
	var pCutNew *Lpk_Cut_t
	var pObj, pFanin *Abc_Obj_t
	var i, k, j, nLeavesNew int32
	/*
	   printf( "Exploring cut " );
	   Lpk_NodePrintCut( p, pCut, 1 );
	   printf( "with node %d\n", Node );
	*/
	// check if the cut can stand adding one more internal node
	if pCut.nNodes == LPK_SIZE_MAX {
		return
	}

	// if the node is a PI, quit
	pObj = abc.Abc_NtkObj(p.pNtk, Node)
	if abc.Abc_ObjIsCi(pObj) != 0 {
		return
	}
	assert(abc.Abc_ObjIsNode(pObj), "abc.Abc_ObjIsNode(pObj)")
	//    assert( Abc_ObjFaninNum(pObj) <= p->pPars->NLutSize );

	// if the node is not in the MFFC, check the limit
	if abc.Abc_NodeIsTravIdCurrent(pObj) == 0 {
		if int32(pCut.nNodesDup) == p.pPars.NLutsOver {
			return
		}
		assert(int32(pCut.nNodesDup) < p.pPars.NLutsOver, "int32(pCut.nNodesDup) < p.pPars.NLutsOver")
	}

	// check the possibility of adding this node using the signature
	nLeavesNew = int32(pCut.nLeaves) - 1
	for i = 0; i < int32(len(pObj.Fanins())); i++ {
		pFanin = pObj.Fanins()[i].Object
		if pCut.uSign[(pFanin.ID>>5)&1]&(1<<(pFanin.ID&31)) != 0 {
			continue
		}
		nLeavesNew++
		if nLeavesNew > p.pPars.NVarsMax {
			return
		}
	}

	// initialize the set of leaves to the nodes in the cut
	assert(p.nCuts < LPK_CUTS_MAX, "p.nCuts < LPK_CUTS_MAX")
	pCutNew = &p.pCuts[p.nCuts]
	pCutNew.nLeaves = 0
	for i = 0; i < int32(pCut.nLeaves); i++ {
		if pCut.pLeaves[i] != Node {
			pCutNew.pLeaves[pCutNew.nLeaves] = pCut.pLeaves[i]
			pCutNew.nLeaves++
		}
	}

	// add new nodes
	for i = 0; i < int32(len(pObj.Fanins())); i++ {
		pFanin = pObj.Fanins()[i].Object
		// find the place where this node belongs
		for k = 0; k < int32(pCutNew.nLeaves); k++ {
			if pCutNew.pLeaves[k] >= int32(pFanin.ID) {
				break
			}
		}
		if k < int32(pCutNew.nLeaves) && pCutNew.pLeaves[k] == int32(pFanin.ID) {
			continue
		}
		// check if there is room
		if int32(pCutNew.nLeaves) == p.pPars.NVarsMax {
			return
		}
		// move all the nodes
		for j = int32(pCutNew.nLeaves); j > k; j-- {
			pCutNew.pLeaves[j] = pCutNew.pLeaves[j-1]
		}
		pCutNew.pLeaves[k] = int32(pFanin.ID)
		pCutNew.nLeaves++
		assert(pCutNew.nLeaves <= LPK_SIZE_MAX, "pCutNew.nLeaves <= LPK_SIZE_MAX")

	}
	// skip the contained cuts
	Lpk_NodeCutSignature(pCutNew)
	if Lpk_NodeCutsOneFilter(p.pCuts[:], p.nCuts, pCutNew) != 0 {
		return
	}

	// update the set of internal nodes
	assert(pCut.nNodes < LPK_SIZE_MAX, "pCut.nNodes < LPK_SIZE_MAX")
	copy(pCutNew.pNodes[:], pCut.pNodes[:pCut.nNodes])
	pCutNew.nNodes = pCut.nNodes
	pCutNew.nNodesDup = pCut.nNodesDup

	// check if the node is already there
	// if so, move the node to be the last
	for i = 0; i < int32(pCutNew.nNodes); i++ {
		if pCutNew.pNodes[i] == Node {
			for k = i; k < int32(pCutNew.nNodes)-1; k++ {
				pCutNew.pNodes[k] = pCutNew.pNodes[k+1]
			}
			pCutNew.pNodes[k] = Node
			break
		}
	}
	if i == int32(pCutNew.nNodes) { // new node

		pCutNew.pNodes[pCutNew.nNodes] = Node
		pCutNew.nNodes++
		if abc.Abc_NodeIsTravIdCurrent(pObj) == 0 {
			pCutNew.nNodesDup++
		}
	}
	// the number of nodes does not exceed MFFC plus duplications
	assert(int32(pCutNew.nNodes) <= p.nMffc+int32(pCutNew.nNodesDup), "pCutNew.nNodes <= p.nMffc + pCutNew.nNodesDup")
	// add the cut to storage
	assert(p.nCuts < LPK_CUTS_MAX, "p.nCuts < LPK_CUTS_MAX")
	p.nCuts++
}

/*
*Function*************************************************************

	Synopsis    [Count support.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_CountSupp(p *Abc_Ntk_t, vNodes *Vec_Ptr_t) int32 {
	var pObj, pFanin *Abc_Obj_t
	var i, Count int32
	for i = 0; i < int32(vec.Vec_PtrSize(vNodes)); i++ {
		pObj = vec.Vec_PtrEntry(vNodes, int(i)).(*abc.Object)
		for _, edge := range pObj.Fanins() {
			pFanin = edge.Object
			if abc.Abc_NodeIsTravIdCurrent(pFanin) == 0 {
				if !pFanin.Persist {
					Count++
				}
				pFanin.Persist = true
			}
		}
	}
	for i = 0; i < int32(vec.Vec_PtrSize(vNodes)); i++ {
		pObj = vec.Vec_PtrEntry(vNodes, int(i)).(*abc.Object)
		for _, edge := range pObj.Fanins() {
			edge.Object.Persist = false
		}
	}
	return Count
}

/*
*Function*************************************************************

	Synopsis    [Computes the set of all cuts.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_NodeCuts(p *Lpk_Man_t) int32 {
	var pCut, pCut2 *Lpk_Cut_t
	var i, k, Temp, nMffc, fChanges int32
	//int nSupp;

	// mark the MFFC of the node with the current trav ID
	vec.Vec_PtrClear(p.vTemp)
	var mffcNodes []*abc.Object
	nMffc = abc.Abc_NodeMffcLabel(p.pObj, &mffcNodes)
	for _, node := range mffcNodes {
		vec.Vec_PtrPush(p.vTemp, node)
	}
	p.nMffc = nMffc
	assert(nMffc > 0, "nMffc > 0")
	if nMffc == 1 {
		return 0
	}
	/*
	   // count the leaves
	   nSupp = Lpk_CountSupp( p->pNtk, p->vTemp );
	   if ( nMffc > 10 && nSupp <= 10 )
	       printf( "Obj = %4d : Supp = %4d. Mffc = %4d.\n", p->pObj->Id, nSupp, nMffc );
	*/
	// initialize the first cut
	pCut = &p.pCuts[0]
	p.nCuts = 1
	pCut.nNodes = 0
	pCut.nNodesDup = 0
	pCut.nLeaves = 1
	pCut.pLeaves[0] = int32(p.pObj.ID)
	// assign the signature
	Lpk_NodeCutSignature(pCut)

	// perform the cut computation
	for i = 0; i < p.nCuts; i++ {
		pCut = &p.pCuts[i]
		if pCut.nLeaves == 0 {
			continue
		}

		// try to expand the fanins of this cut
		for k = 0; k < int32(pCut.nLeaves); k++ {
			// create a new cut
			Lpk_NodeCutsOne(p, pCut, pCut.pLeaves[k])
			// quit if the number of cuts has exceeded the limit
			if p.nCuts == LPK_CUTS_MAX {
				break
			}
		}
		if p.nCuts == LPK_CUTS_MAX {
			break
		}
	}
	if p.nCuts == LPK_CUTS_MAX {
		p.nNodesOver++
	}

	// record the impact of this node
	if p.pPars.FSatur != 0 {
		Lpk_NodeRecordImpact(p)
	}

	// compress the cuts by removing empty ones, those with negative Weight, and decomposable ones
	p.nEvals = 0
	for i = 0; i < p.nCuts; i++ {
		pCut = &p.pCuts[i]
		if pCut.nLeaves < 2 {
			continue
		}
		// compute the minimum number of LUTs needed to implement this cut
		// V = N * (K-1) + 1  ~~~~~  N = Ceiling[(V-1)/(K-1)] = (V-1)/(K-1) + [(V-1)%(K-1) > 0]
		pCut.nLuts = uint32(Lpk_LutNumLuts(int32(pCut.nLeaves), p.pPars.NLutSize))
		//        pCut->Weight = 1.0 * (pCut->nNodes - pCut->nNodesDup - 1) / pCut->nLuts; //p->pPars->NLutsMax;
		pCut.Weight = float32(pCut.nNodes-pCut.nNodesDup) / float32(pCut.nLuts) //p->pPars->NLutsMax;
		if pCut.Weight <= 1.001 {
			//        if ( pCut->Weight <= 0.999 )
			continue
		}
		pCut.fHasDsd = uint32(Lpk_NodeCutsCheckDsd(p, pCut))
		if pCut.fHasDsd != 0 {
			continue
		}
		p.pEvals[p.nEvals] = i
		p.nEvals++

		//        if ( pCut->nLeaves <= 9 && pCut->nNodes > 15 )
		//        printf( "%5d : Obj = %4d  Leaves = %4d  Nodes = %4d  LUTs = %4d\n", i, p->pObj->Id, pCut->nLeaves, pCut->nNodes, pCut->nLuts );
	}
	if p.nEvals == 0 {
		return 0
	}

	// sort the cuts by Weight
	for {
		fChanges = 0
		for i = 0; i < p.nEvals-1; i++ {
			pCut = &p.pCuts[p.pEvals[i]]
			pCut2 = &p.pCuts[p.pEvals[i+1]]
			if pCut.Weight >= pCut2.Weight-0.001 {
				continue
			}
			Temp = p.pEvals[i]
			p.pEvals[i] = p.pEvals[i+1]
			p.pEvals[i+1] = Temp
			fChanges = 1
		}
		if fChanges == 0 {
			break
		}
	}
	/*
	   for ( i = 0; i < p->nEvals; i++ )
	   {
	       pCut = p->pCuts + p->pEvals[i];
	       printf( "Cut %3d : W = %5.2f.\n", i, pCut->Weight );
	   }
	   printf( "\n" );
	*/
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
