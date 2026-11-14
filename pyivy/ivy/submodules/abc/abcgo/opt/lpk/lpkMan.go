package lpk

import (
	kit "github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	vec "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [lpkMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkMan.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_ManStart(pPars *Lpk_Par_t) *Lpk_Man_t {
	if pPars == nil || pPars.nLutsMax < 0 || pPars.nLutsMax > 16 || pPars.nVarsMax < 1 || pPars.nVarsMax > 16 {
		return nil
	}
	p := &Lpk_Man_t{pPars: pPars, nCutsMax: LPK_CUTS_MAX}
	p.vTtElems = vec.Vec_PtrAllocTruthTables(int(pPars.nVarsMax))
	p.vTtNodes = vec.Vec_PtrAllocSimInfo(1024, int(kit.Kit_TruthWordNum(pPars.nVarsMax)))
	p.vCover = vec.Vec_IntAlloc(1 << 12)
	p.vLeaves = vec.Vec_PtrAlloc(32)
	p.vTemp = vec.Vec_PtrAlloc(32)
	for i := range p.vSets {
		p.vSets[i] = vec.Vec_IntAlloc(100)
	}
	p.pDsdMan = kit.Kit_DsdManAlloc(pPars.nVarsMax, 64)
	p.vMemory = vec.Vec_IntAlloc(1024 * 32)
	p.vBddDir = vec.Vec_IntAlloc(256)
	p.vBddInv = vec.Vec_IntAlloc(256)
	nWords := int(kit.Kit_TruthWordNum(pPars.nVarsMax))
	storage := make([]uint32, 32*nWords)
	at := 0
	for level := 0; level < 5; level++ {
		for index := 0; index < 1<<level; index++ {
			p.ppTruths[level][index] = storage[at : at+nWords]
			at += nWords
		}
	}
	return p
}

func Lpk_ManStop(p *Lpk_Man_t) {
	if p == nil {
		return
	}
	for i := range p.vSets {
		vec.Vec_IntFree(p.vSets[i])
		p.vSets[i] = nil
	}
	kit.Kit_DsdManFree(p.pDsdMan)
	vec.Vec_IntFree(p.vMemory)
	p.vMemory = nil
	vec.Vec_IntFree(p.vBddDir)
	p.vBddDir = nil
	vec.Vec_IntFree(p.vBddInv)
	p.vBddInv = nil
	if p.pIfMan != nil {
		ifpkg.If_ManStop(p.pIfMan)
		p.pIfMan = nil
	}
	if p.vLevels != nil {
		vec.Vec_VecFree(p.vLevels)
		p.vLevels = nil
	}
	if p.vVisited != nil {
		vec.Vec_VecFree(p.vVisited)
		p.vVisited = nil
	}
	vec.Vec_PtrFree(p.vLeaves)
	p.vLeaves = nil
	vec.Vec_PtrFree(p.vTemp)
	p.vTemp = nil
	vec.Vec_IntFree(p.vCover)
	p.vCover = nil
	vec.Vec_PtrFree(p.vTtElems)
	p.vTtElems = nil
	vec.Vec_PtrFree(p.vTtNodes)
	p.vTtNodes = nil
	p.ppTruths = [5][16][]uint32{}
	p.pPars = nil
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
