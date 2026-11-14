package lpk

import (
	kit "github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	util "github.com/glycerine/abcgo/misc/util"
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
func Lpk_MapPrimeInternal(pIfMan *If_Man_t, pGraph *Kit_Graph_t, leaves []*If_Obj_t) *If_Obj_t {
	// check for constant function
	if kit.Kit_GraphIsConst(pGraph) != 0 {
		return ifpkg.If_ManConst1(pIfMan)
	}
	// check for a literal
	if kit.Kit_GraphIsVar(pGraph) != 0 {
		return leaves[kit.Kit_GraphVarInt(pGraph)]
	}
	// build the AIG nodes corresponding to the AND gates of the graph
	// inlining of macro Kit_GraphForEachNode
	values := make([]*If_Obj_t, kit.Kit_GraphLeaveNum(pGraph)+kit.Kit_GraphNodeNum(pGraph))
	copy(values, leaves)
	var result *If_Obj_t
	for i, node := range kit.Kit_GraphForEachNode(pGraph) {
		e0, e1 := kit.Kit_NodeEdges(node)
		pAnd0, pAnd1 := values[e0.Node], values[e1.Node]
		result = ifpkg.If_ManCreateAnd(pIfMan,
			ifpkg.If_NotCond(ifpkg.If_Regular(pAnd0), ifpkg.If_IsComplement(pAnd0)^int32(kit.Kit_EdgeToInt(e0)&1)),
			ifpkg.If_NotCond(ifpkg.If_Regular(pAnd1), ifpkg.If_IsComplement(pAnd1)^int32(kit.Kit_EdgeToInt(e1)&1)))
		values[i] = result
	}
	return result
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
	var pRes *If_Obj_t
	// derive the factored form
	pGraph = kit.Kit_TruthToGraph(pTruth, nVars, nil)
	if pGraph == nil {
		return nil
	}
	// perform strashing
	pRes = Lpk_MapPrimeInternal(p.pIfMan, pGraph, ppLeaves)
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
	var i, iLitFanin int32

	assert(iLit >= 0, "iLit >= 0")

	// consider the case of a gate
	pObj = kit.Kit_DsdNtkObj(pNtk, util.Abc_Lit2Var(iLit))
	if pObj == nil {
		pObjNew = ppLeaves[util.Abc_Lit2Var(iLit)]
		return ifpkg.If_NotCond(pObjNew, util.Abc_LitIsCompl(iLit))
	}
	if pObj.Type == kit.KIT_DSD_CONST1 {
		return ifpkg.If_NotCond(ifpkg.If_ManConst1(p.pIfMan), util.Abc_LitIsCompl(iLit))
	}
	if pObj.Type == kit.KIT_DSD_VAR {
		pObjNew = ppLeaves[util.Abc_Lit2Var(int32(kit.Kit_DsdObjFans(pObj)[0]))]
		return ifpkg.If_NotCond(pObjNew, util.Abc_LitIsCompl(iLit)^util.Abc_LitIsCompl(int32(kit.Kit_DsdObjFans(pObj)[0])))
	}
	if pObj.Type == kit.KIT_DSD_AND {
		assert(kit.Kit_DsdObjFanNum(pObj) == 2, "kit.Kit_DsdObjFanNum(pObj) == 2")
		pFansNew[0] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(kit.Kit_DsdObjFans(pObj)[0]), nil)
		if pResult != nil {
			pFansNew[1] = pResult
		} else {
			pFansNew[1] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(kit.Kit_DsdObjFans(pObj)[1]), nil)
		}
		if pFansNew[0] == nil || pFansNew[1] == nil {
			return nil
		}
		pObjNew = ifpkg.If_ManCreateAnd(p.pIfMan, pFansNew[0], pFansNew[1])
		return ifpkg.If_NotCond(pObjNew, util.Abc_LitIsCompl(iLit))
	}
	if pObj.Type == kit.KIT_DSD_XOR {
		var fCompl int32 = util.Abc_LitIsCompl(iLit)
		assert(kit.Kit_DsdObjFanNum(pObj) == 2, "kit.Kit_DsdObjFanNum(pObj) == 2")
		pFansNew[0] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(kit.Kit_DsdObjFans(pObj)[0]), nil)
		if pResult != nil {
			pFansNew[1] = pResult
		} else {
			pFansNew[1] = Lpk_MapTree_rec(p, pNtk, ppLeaves[:], int32(kit.Kit_DsdObjFans(pObj)[1]), nil)
		}
		if pFansNew[0] == nil || pFansNew[1] == nil {
			return nil
		}
		fCompl ^= ifpkg.If_IsComplement(pFansNew[0]) ^ ifpkg.If_IsComplement(pFansNew[1])
		pObjNew = ifpkg.If_ManCreateXor(p.pIfMan, ifpkg.If_Regular(pFansNew[0]), ifpkg.If_Regular(pFansNew[1]))
		return ifpkg.If_NotCond(pObjNew, fCompl)
	}
	assert(pObj.Type == kit.KIT_DSD_PRIME, "pObj.Type == kit.KIT_DSD_PRIME")
	p.nBlocks[kit.Kit_DsdObjFanNum(pObj)]++

	// solve for the inputs
	// inlining of macro Kit_DsdObjForEachFanin
	for i = 0; i < kit.Kit_DsdObjFanNum(pObj); i++ {
		iLitFanin = int32(kit.Kit_DsdObjFans(pObj)[i])
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
	       return ifpkg.If_NotCond( pObjNew, util.Abc_LitIsCompl(iLit) );
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
	if p.pPars.nVarsShared > 0 && int32(kit.Kit_DsdObjFanNum(pObj)) > p.pPars.nLutSize {
		pObjNew2 = Lpk_MapSuppRedDec_rec(p, kit.Kit_DsdObjTruth(pObj), kit.Kit_DsdObjFanNum(pObj), pFansNew[:])
		if pObjNew2 != nil {
			return ifpkg.If_NotCond(pObjNew2, util.Abc_LitIsCompl(iLit))
		}
	}

	pObjNew = Lpk_MapPrime(p, kit.Kit_DsdObjTruth(pObj), kit.Kit_DsdObjFanNum(pObj), pFansNew[:])

	// add choice
	if (pObjNew != nil) && (pObjNew2 != nil) {
		ifpkg.If_ObjSetChoice(ifpkg.If_Regular(pObjNew), ifpkg.If_Regular(pObjNew2))
		ifpkg.If_ManCreateChoice(p.pIfMan, ifpkg.If_Regular(pObjNew))
	}
	return ifpkg.If_NotCond(pObjNew, util.Abc_LitIsCompl(iLit))
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
