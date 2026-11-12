package gia

import abcmain "github.com/glycerine/abcgo/base/main"

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
func Gia_ManAigSynch2(pInit *Gia_Man_t, pPars0 any, nLutSize int32, nRelaxRatio int32) *Gia_Man_t {
    extern Gia_Man_t * Gia_ManLutBalance( Gia_Man_t * p, int nLutSize, int fUseMuxes, int fRecursive, int fOptArea, int fVerbose );
    var pParsDch *Dch_Pars_t  = (Dch_Pars_t *)pPars0;
    var pGia1, pGia2, pGia3, pNew, pTemp *Gia_Man_t;
    var fVerbose int32  = pParsDch.fVerbose;
    var Pars Jf_Par_t;
    var pPars *Jf_Par_t = &Pars;
    Lf_ManSetDefaultPars( pPars );
    pPars.fCutMin     = 1;
    pPars.fCoarsen    = 1;
    pPars.nRelaxRatio = nRelaxRatio;
    pPars.nAreaTuner  = 5;
    pPars.nCutNum     = 12;
    pPars.fVerbose    = fVerbose;
    if fVerbose != 0 {  Gia_ManPrintStats( pInit, nil ); }
    pGia1 = Gia_ManDup( pInit );
    if Gia_ManAndNum(pGia1) == 0 {
        Gia_ManTransferTiming( pGia1, pInit );
        return pGia1;
    }
    if pGia1.pManTime && pGia1.vLevels == nil {
        Gia_ManLevelWithBoxes( pGia1 ); }
    // unmap if mapped
    if Gia_ManHasMapping(pInit) {
        Gia_ManTransferMapping( pGia1, pInit );
        pGia1 = (Gia_Man_t *)Dsm_ManDeriveGia( pTemp = pGia1, 0 );
        Gia_ManStop( pTemp );
    }
    // perform balancing
    if Gia_ManBufNum(pGia1) || true {
        pGia2 = Gia_ManAreaBalance( pGia1, 0, ABC_INFINITY, 0, 0 );
    } else {
        pGia2 = Gia_ManLutBalance( pGia1, nLutSize, 1, 1, 1, 0 );
        pGia2 = Gia_ManAreaBalance( pTemp = pGia2, 0, ABC_INFINITY, 0, 0 );
        Gia_ManStop( pTemp );
    }
    if fVerbose != 0 {     Gia_ManPrintStats( pGia2, nil ); }
    // perform mapping
    pGia2 = Lf_ManPerformMapping( pTemp = pGia2, pPars );
    if fVerbose != 0 {     Gia_ManPrintStats( pGia2, nil ); }
    if pTemp != pGia2 {
        Gia_ManStop( pTemp ); }
    // perform balancing
    if pParsDch.fLightSynth || Gia_ManBufNum(pGia2) {
        pGia3 = Gia_ManAreaBalance( pGia2, 0, ABC_INFINITY, 0, 0 );
    } else {
        assert(Gia_ManBufNum(pGia2) == 0, "Gia_ManBufNum(pGia2) == 0");
        pGia2 = Gia_ManAreaBalance( pTemp = pGia2, 0, ABC_INFINITY, 0, 0 );
        if fVerbose != 0 {     Gia_ManPrintStats( pGia2, nil ); }
        Gia_ManStop( pTemp );
        // perform DSD balancing
        pGia3 = Gia_ManPerformDsdBalance( pGia2, 6, 8, 0, 0 );
    }
    if fVerbose != 0 {     Gia_ManPrintStats( pGia3, nil ); }
    // perform choice computation
    if Gia_ManBufNum(pInit) {
        assert(Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia1), "Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia1)");
        pGia1 = Gia_ManDupFromBarBufs( pTemp = pGia1 );
        Gia_ManStop( pTemp );
        assert(Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia2), "Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia2)");
        pGia2 = Gia_ManDupFromBarBufs( pTemp = pGia2 );
        Gia_ManStop( pTemp );
        assert(Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia3), "Gia_ManBufNum(pInit) == Gia_ManBufNum(pGia3)");
        pGia3 = Gia_ManDupFromBarBufs( pTemp = pGia3 );
        Gia_ManStop( pTemp );
    }
    pNew = Gia_ManAigSynch2Choices( pGia1, pGia2, pGia3, pParsDch );
    Gia_ManStop( pGia1 );
    Gia_ManStop( pGia2 );
    Gia_ManStop( pGia3 );
    if Gia_ManBufNum(pInit) {
        pNew = Gia_ManDupToBarBufs( pTemp = pNew, Gia_ManBufNum(pInit) );
        Gia_ManStop( pTemp );
    }
    // copy names
    // inlining of macro ABC_FREE
    pNew.pName = "";
    // inlining of macro ABC_FREE
    pNew.pSpec = "";
    pNew.pName = Abc_UtilStrsav(pInit.pName);
    pNew.pSpec = Abc_UtilStrsav(pInit.pSpec);
    Gia_ManTransferTiming( pNew, pInit );
    return pNew;
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
    giaPerformFlow3(abcmain.Abc_FrameGetGlobalFrame(), nLutSize, nCutNum, fUseMfs != 0, fUseLutLib != 0, Abc_NtkRecIsRunning3() != 0, fVerbose != 0)
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

