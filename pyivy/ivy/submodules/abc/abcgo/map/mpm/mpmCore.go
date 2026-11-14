package mpm

import "github.com/glycerine/abcgo/aig/gia"

/**CFile****************************************************************

  FileName    [mpmCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmCore.c,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "mpmInt.h"

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
/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Mpm_ManPerformLutMapping(pMig *Mig_Man_t, pPars *Mpm_Par_t) *gia.Gia_Man_t {
	var pNew *gia.Gia_Man_t
	var p *Mpm_Man_t
	p = Mpm_ManStart(pMig, pPars)
	if p.pPars.FVerbose != 0 {
		Mpm_ManPrintStatsInit(p)
	}
	Mpm_ManPrepare(p)
	Mpm_ManPerform(p)
	if p.pPars.FVerbose != 0 {
		Mpm_ManPrintStats(p)
	}
	if mapped := Mpm_ManFromIfLogic(p); mapped != nil {
		pNew = mapped.(*gia.Gia_Man_t)
	}
	Mpm_ManStop(p)
	return pNew
}
func Mpm_ManLutMapping(pGia *gia.Gia_Man_t, pPars *Mpm_Par_t) *gia.Gia_Man_t {
	var p *Mig_Man_t
	var pNew *gia.Gia_Man_t
	assert(pPars.PLib.LutMax <= MPM_VAR_MAX, "pPars.pLib.LutMax <= MPM_VAR_MAX")
	assert(pPars.NNumCuts <= MPM_CUT_MAX, "pPars.nNumCuts <= MPM_CUT_MAX")
	if pPars.FUseGates != 0 {
		pGia = gia.Gia_ManDupMuxes(pGia, 2)
		p = Mig_ManCreate(pGia)
		gia.Gia_ManStop(pGia)

	} else {
		p = Mig_ManCreate(pGia)
	}
	pNew = Mpm_ManPerformLutMapping(p, pPars)
	Mig_ManStop(p)
	return pNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
