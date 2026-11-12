package lpk

import (
	abc "github.com/glycerine/abcgo/base/abc"
	kit "github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
)

/**CFile****************************************************************

  FileName    [lpkMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkMap.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "lpkInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Transforms the decomposition graph into the AIG.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MapPrimeInternal(pIfMan *If_Man_t, pGraph *Kit_Graph_t) *If_Obj_t {
	var pNode *Kit_Node_t = nil // Suppress "might be used uninitialized"
	var pAnd0, pAnd1 *If_Obj_t
	var i int32
	// check for constant function
	if kit.Kit_GraphIsConst(pGraph) != 0 {
		return ifpkg.If_ManConst1(pIfMan)
	}
	// check for a literal
	if kit.Kit_GraphIsVar(pGraph) != 0 {
		return kit.Kit_GraphVar(pGraph).pFunc
	}
	// build the AIG nodes corresponding to the AND gates of the graph
	// inlining of macro Kit_GraphForEachNode
	for i = int32(pGraph.nLeaves); i < int32(pGraph.nSize); i++ {
		pNode = kit.Kit_GraphNode(pGraph, i)
		pAnd0 = kit.Kit_GraphNode(pGraph, pNode.eEdge0.Node).pFunc
		pAnd1 = kit.Kit_GraphNode(pGraph, pNode.eEdge1.Node).pFunc
		pNode.pFunc = ifpkg.If_ManCreateAnd(pIfMan,
			ifpkg.If_NotCond(ifpkg.If_Regular(pAnd0), ifpkg.If_IsComplement(pAnd0)^pNode.eEdge0.fCompl),
			ifpkg.If_NotCond(ifpkg.If_Regular(pAnd1), ifpkg.If_IsComplement(pAnd1)^pNode.eEdge1.fCompl))
	}
	return pNode.pFunc
}

/*
*Function*************************************************************

	Synopsis    [Strashes one logic node using its SOP.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MapPrime(p *Lpk_Man_t, pTruth []uint32, nVars int32, ppLeaves []*If_Obj_t) *If_Obj_t {
	var pGraph *Kit_Graph_t
	var pNode *Kit_Node_t
	var pRes *If_Obj_t
	var i int32
	// derive the factored form
	pGraph = kit.Kit_TruthToGraph(pTruth, nVars, p.vCover)
	if pGraph == nil {
		return nil
	}
	// collect the fanins
	// inlining of macro Kit_GraphForEachLeaf
	for i = 0; i < int32(pGraph.nLeaves); i++ {
		pNode = kit.Kit_GraphNode(pGraph, i)
		pNode.pFunc = ppLeaves[i]
	}
	// perform strashing
	pRes = Lpk_MapPrimeInternal(p.pIfMan, pGraph)
	pRes = ifpkg.If_NotCond(pRes, kit.Kit_GraphIsComplement(pGraph))
	kit.Kit_GraphFree(pGraph)
	return pRes
}

/*
*Function*************************************************************

	Synopsis    [Prepares the mapping manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MapTree_rec(p *Lpk_Man_t, pNtk *Kit_DsdNtk_t, ppLeaves []*If_Obj_t, iLit int32, pResult *If_Obj_t) *If_Obj_t {
	var pObj *Kit_DsdObj_t
	var pObjNew *If_Obj_t = nil
	var pObjNew2 *If_Obj_t = nil
	var pFansNew [16]*If_Obj_t
	var i, iLitFanin uint32

	assert(iLit >= 0, "iLit >= 0")

	// consider the case of a gate
	pObj = kit.Kit_DsdNtkObj(pNtk, abc.Abc_Lit2Var(iLit))
	if pObj == nil {
		pObjNew = ppLeaves[abc.Abc_Lit2Var(iLit)]
		return ifpkg.If_NotCond(pObjNew, abc.Abc_LitIsCompl(iLit))
	}
	if pObj.Type == KIT_DSD_CONST1 {
		return ifpkg.If_NotCond(ifpkg.If_ManConst1(p.pIfMan), abc.Abc_LitIsCompl(iLit))
	}
	if pObj.Type == KIT_DSD_VAR {
		pObjNew = ppLeaves[abc.Abc_Lit2Var(int32(pObj.pFans[0]))]
		return ifpkg.If_NotCond(pObjNew, abc.Abc_LitIsCompl(iLit)^abc.Abc_LitIsCompl(int32(pObj.pFans[0])))
	}
	if pObj.Type == KIT_DSD_AND {
		assert(pObj.nFans == 2, "pObj.nFans == 2")
		pFansNew[0] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(pObj.pFans[0]), nil)
		if pResult != nil {
			pFansNew[1] = pResult
		} else {
			pFansNew[1] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(pObj.pFans[1]), nil)
		}
		if pFansNew[0] == nil || pFansNew[1] == nil {
			return nil
		}
		pObjNew = ifpkg.If_ManCreateAnd(p.pIfMan, pFansNew[0], pFansNew[1])
		return ifpkg.If_NotCond(pObjNew, abc.Abc_LitIsCompl(iLit))
	}
	if pObj.Type == KIT_DSD_XOR {
		var fCompl int32 = abc.Abc_LitIsCompl(iLit)
		assert(pObj.nFans == 2, "pObj.nFans == 2")
		pFansNew[0] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(pObj.pFans[0]), nil)
		if pResult != nil {
			pFansNew[1] = pResult
		} else {
			pFansNew[1] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(pObj.pFans[1]), nil)
		}
		if pFansNew[0] == nil || pFansNew[1] == nil {
			return nil
		}
		fCompl ^= ifpkg.If_IsComplement(pFansNew[0]) ^ ifpkg.If_IsComplement(pFansNew[1])
		pObjNew = ifpkg.If_ManCreateXor(p.pIfMan, ifpkg.If_Regular(pFansNew[0]), ifpkg.If_Regular(pFansNew[1]))
		return ifpkg.If_NotCond(pObjNew, fCompl)
	}
	assert(pObj.Type == KIT_DSD_PRIME, "pObj.Type == KIT_DSD_PRIME")
	p.nBlocks[pObj.nFans]++

	// solve for the inputs
	// inlining of macro Kit_DsdObjForEachFanin
	for i = 0; i < int32(pObj.nFans); i++ {
		iLitFanin = int32(pObj.pFans[i])
		if i == 0 {
			if pResult != nil {
				pFansNew[i] = pResult
			} else {
				pFansNew[i] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], iLitFanin, nil)
			}
		} else {
			pFansNew[i] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], iLitFanin, nil)
		}
		if pFansNew[i] == nil {
			return nil
		}
	}
	/*
	   if ( !p->fCofactoring && p->pPars->nVarsShared > 0 && pObj->nFans > p->pPars->nLutSize )
	   {
	       pObjNew = Lpk_MapTreeMulti( p, kit.Kit_DsdObjTruth(pObj), pObj->nFans, pFansNew );
	       return ifpkg.If_NotCond( pObjNew, abc.Abc_LitIsCompl(iLit) );
	   }
	*/
	/*
	       if ( pObj->nFans > p->pPars->nLutSize )
	       {
	           pObjNew2 = Lpk_MapTreeMux_rec( p, kit.Kit_DsdObjTruth(pObj), pObj->nFans, pFansNew );
	   //        if ( pObjNew2 )
	   //            return If_NotCond( pObjNew2, Abc_LitIsCompl(iLit) );
	       }
	*/

	// find best cofactoring variable
	if p.pPars.nVarsShared > 0 && int32(pObj.nFans) > p.pPars.nLutSize {
		pObjNew2 = Lpk_MapSuppRedDec_rec(p, kit.Kit_DsdObjTruth(pObj), pObj.nFans, pFansNew[:])
		if pObjNew2 != nil {
			return ifpkg.If_NotCond(pObjNew2, abc.Abc_LitIsCompl(iLit))
		}
	}

	pObjNew = Lpk_MapPrime(p, kit.Kit_DsdObjTruth(pObj), pObj.nFans, pFansNew[:])

	// add choice
	if (pObjNew != nil) && (pObjNew2 != nil) {
		ifpkg.If_ObjSetChoice(ifpkg.If_Regular(pObjNew), ifpkg.If_Regular(pObjNew2))
		ifpkg.If_ManCreateChoice(p.pIfMan, ifpkg.If_Regular(pObjNew))
	}
	return ifpkg.If_NotCond(pObjNew, abc.Abc_LitIsCompl(iLit))
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
