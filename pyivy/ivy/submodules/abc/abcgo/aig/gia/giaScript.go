package gia

import (
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/proof/dch"
)

/**CFile****************************************************************

  FileName    [giaScript.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Various hardcoded scripts.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaScript.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "giaAig.h"
//#include "base/main/main.h"
//#include "base/cmd/cmd.h"
//#include "proof/dch/dch.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Synthesis script.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Synthesis script.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAigSyn2(pInit *Gia_Man_t, fOldAlgo, fCoarsen, fCutMin, nRelaxRatio, fDelayMin, fVerbose, fVeryVerbose int32) *Gia_Man_t {
	return giaAigSyn2(pInit, fOldAlgo, fCoarsen, fCutMin, nRelaxRatio, fDelayMin, fVerbose, fVeryVerbose, Gia_ManPerformDsdBalance)
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager.]

  Description [This duplicator works for AIGs with choices.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAigSynch2(pInit *Gia_Man_t, pars *dch.Params, nLutSize, nRelaxRatio int32) *Gia_Man_t {
	return giaAigSynch2(pInit, pars, nRelaxRatio, Gia_ManPerformDsdBalance)
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformFlow3(nLutSize int32, nCutNum int32, fBalance int32, fMinAve int32, fUseMfs int32, fUseLutLib int32, fVerbose int32) {
	giaPerformFlow3(abcmain.Abc_FrameGetGlobalFrame(), nLutSize, nCutNum, fUseMfs != 0, fUseLutLib != 0, giaRecordingActive(), fVerbose != 0)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
