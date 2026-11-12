package cec

import "time"

/**CFile****************************************************************

  FileName    [cecSeq.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Refinement of sequential equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSeq.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets register values from the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSeqDeriveInfoFromCex(vInfo *Vec_Ptr_t, pAig *Gia_Man_t, pCex *Abc_Cex_t) {
    var pInfo *uint32;
    var k, i, w, nWords int32;
    assert(pCex.nBits == pCex.nRegs + pCex.nPis * (pCex.iFrame + 1), "pCex.nBits == pCex.nRegs + pCex.nPis * (pCex.iFrame + 1)");
    assert(pCex.nBits - pCex.nRegs + Gia_ManRegNum(pAig) <= Vec_PtrSize(vInfo), "pCex.nBits - pCex.nRegs + Gia_ManRegNum(pAig) <= Vec_PtrSize(vInfo)");
    nWords = Vec_PtrReadWordsSimInfo( vInfo );
/*
    // user register values
    assert( pCex->nRegs == Gia_ManRegNum(pAig) );
    for ( k = 0; k < pCex->nRegs; k++ )
    {
        pInfo = (unsigned *)Vec_PtrEntry( vInfo, k );
        for ( w = 0; w < nWords; w++ )
            pInfo[w] = Abc_InfoHasBit( pCex->pData, k )? ~0 : 0;
    }
*/
    // print warning about register values
    for  k = 0; k < pCex.nRegs; k++  {
        if Abc_InfoHasBit( pCex.pData, k ) {
            break; }
    }
    if k < pCex.nRegs {
        Abc_Print( 0, "The CEX has flop values different from 0, but they are currently not used by \"resim\".\n" ); }

    // assign zero register values
    for  k = 0; k < Gia_ManRegNum(pAig); k++  {
        pInfo = (unsigned *)Vec_PtrEntry( vInfo, k );
        for  w = 0; w < nWords; w++  {
            pInfo[w] = 0;
        }
    }
    for  i = pCex.nRegs; i < pCex.nBits; i++  {
        pInfo = (unsigned *)Vec_PtrEntry( vInfo, k++ );
        for  w = 0; w < nWords; w++  {
            pInfo[w] = Gia_ManRandom(0);
        }
        // set simulation pattern and make sure it is second (first will be erased during simulation)
        pInfo[0] = (pInfo[0] << 1) | Abc_InfoHasBit( pCex.pData, i ); 
        pInfo[0] <<= 1;
    }
    for  ; k < Vec_PtrSize(vInfo); k++  {
        pInfo = (unsigned *)Vec_PtrEntry( vInfo, k );
        for  w = 0; w < nWords; w++  {
            pInfo[w] = Gia_ManRandom(0);
        }
    }
}

/**Function*************************************************************

  Synopsis    [Sets register values from the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSeqDeriveInfoInitRandom(vInfo *Vec_Ptr_t, pAig *Gia_Man_t, pCex *Abc_Cex_t) {
    var pInfo *uint32;
    var k, w, nWords int32;
    nWords = Vec_PtrReadWordsSimInfo( vInfo );
    assert(pCex == nil || Gia_ManRegNum(pAig) == pCex.nRegs, "pCex == nil || Gia_ManRegNum(pAig) == pCex.nRegs");
    assert(Gia_ManRegNum(pAig) <= Vec_PtrSize(vInfo), "Gia_ManRegNum(pAig) <= Vec_PtrSize(vInfo)");
    for  k = 0; k < Gia_ManRegNum(pAig); k++  {
        pInfo = (unsigned *)Vec_PtrEntry( vInfo, k );
        for  w = 0; w < nWords; w++  {
            pInfo[w] = (pCex && Abc_InfoHasBit(pCex.pData, k))? ~0 : 0;
        }
    }

    for  ; k < Vec_PtrSize(vInfo); k++  {
        pInfo = (unsigned *)Vec_PtrEntry( vInfo, k );
        for  w = 0; w < nWords; w++  {
            pInfo[w] = Gia_ManRandom( 0 );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Resimulates the classes using sequential simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSeqResimulate(p *Cec_ManSim_t, vInfo *Vec_Ptr_t) int32 {
    var pInfo0, pInfo1 *uint32;
    var f, i, k, w int32;
//    assert( Gia_ManRegNum(p->pAig) > 0 );
    assert(Vec_PtrSize(vInfo) == Gia_ManRegNum(p.pAig) + Gia_ManPiNum(p.pAig) * p.pPars.nFrames, "Vec_PtrSize(vInfo) == Gia_ManRegNum(p.pAig) + Gia_ManPiNum(p.pAig) * p.pPars.nFrames");
    for  k = 0; k < Gia_ManRegNum(p.pAig); k++  {
        pInfo0 = (unsigned *)Vec_PtrEntry( vInfo, k );
        pInfo1 = (unsigned *)Vec_PtrEntry( p.vCoSimInfo, Gia_ManPoNum(p.pAig) + k );
        for  w = 0; w < p.nWords; w++  {
            pInfo1[w] = pInfo0[w];
        }
    }
    for  f = 0; f < p.pPars.nFrames; f++  {
        for  i = 0; i < Gia_ManPiNum(p.pAig); i++  {
            pInfo0 = (unsigned *)Vec_PtrEntry( vInfo, k++ );
            pInfo1 = (unsigned *)Vec_PtrEntry( p.vCiSimInfo, i );
            for  w = 0; w < p.nWords; w++  {
                pInfo1[w] = pInfo0[w];
            }
        }
        for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
            pInfo0 = (unsigned *)Vec_PtrEntry( p.vCoSimInfo, Gia_ManPoNum(p.pAig) + i );
            pInfo1 = (unsigned *)Vec_PtrEntry( p.vCiSimInfo, Gia_ManPiNum(p.pAig) + i );
            for  w = 0; w < p.nWords; w++  {
                pInfo1[w] = pInfo0[w];
            }
        }
        if Cec_ManSimSimulateRound( p, p.vCiSimInfo, p.vCoSimInfo ) {
            return 1; }
    }
    assert(k == Vec_PtrSize(vInfo), "k == Vec_PtrSize(vInfo)");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Resimulates information to refine equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSeqResimulateInfo(pAig *Gia_Man_t, vSimInfo *Vec_Ptr_t, pBestState *Abc_Cex_t, fCheckMiter int32) int32 {
    var ParsSim Cec_ParSim_t;
    var pParsSim *Cec_ParSim_t = &ParsSim;
    var pSim *Cec_ManSim_t;
    var RetValue int32;//, clkTotal = Abc_Clock();
    assert((Vec_PtrSize(vSimInfo) - Gia_ManRegNum(pAig)) % Gia_ManPiNum(pAig) == 0, "(Vec_PtrSize(vSimInfo) - Gia_ManRegNum(pAig)) % Gia_ManPiNum(pAig) == 0");
    Cec_ManSimSetDefaultParams( pParsSim );
    pParsSim.nFrames = (Vec_PtrSize(vSimInfo) - Gia_ManRegNum(pAig)) / Gia_ManPiNum(pAig);
    pParsSim.nWords  = Vec_PtrReadWordsSimInfo( vSimInfo );
    pParsSim.fCheckMiter = fCheckMiter;
    Gia_ManCreateValueRefs( pAig );
    pSim = Cec_ManSimStart( pAig, pParsSim );
    if pBestState != nil {
        pSim.pBestState = pBestState; }
    RetValue = Cec_ManSeqResimulate( pSim, vSimInfo );
    pSim.pBestState = nil;
    Cec_ManSimStop( pSim );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Resimuates one counter-example to refine equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSeqResimulateCounter(pAig *Gia_Man_t, pPars *Cec_ParSim_t, pCex *Abc_Cex_t) int32 {
    var vSimInfo *Vec_Ptr_t;
    var RetValue int32;
    var clkTotal abctime  = Abc_Clock();
    if pCex == nil {
        Abc_Print( 1, "Cec_ManSeqResimulateCounter(): Counter-example is not available.\n" );
        return -1;
    }
    if pAig.pReprs == nil {
        Abc_Print( 1, "Cec_ManSeqResimulateCounter(): Equivalence classes are not available.\n" );
        return -1;
    }
    if Gia_ManRegNum(pAig) == 0 {
        Abc_Print( 1, "Cec_ManSeqResimulateCounter(): Not a sequential AIG.\n" );
        return -1;
    }
//    if ( Gia_ManRegNum(pAig) != pCex->nRegs || Gia_ManPiNum(pAig) != pCex->nPis )
    if Gia_ManPiNum(pAig) != pCex.nPis {
        Abc_Print( 1, "Cec_ManSeqResimulateCounter(): The number of PIs in the AIG and the counter-example differ.\n" );
        return -1;
    }
    if pPars.fVerbose {
        Abc_Print( 1, "Resimulating %d timeframes.\n", pPars.nFrames + pCex.iFrame + 1 ); }
    Gia_ManRandom( 1 );
    vSimInfo = Vec_PtrAllocSimInfo( Gia_ManRegNum(pAig) + 
        Gia_ManPiNum(pAig) * (pPars.nFrames + pCex.iFrame + 1), 1 );
    Cec_ManSeqDeriveInfoFromCex( vSimInfo, pAig, pCex );
    if pPars.fVerbose {
        Gia_ManEquivPrintClasses( pAig, 0, 0 ); }
    RetValue = Cec_ManSeqResimulateInfo( pAig, vSimInfo, nil, pPars.fCheckMiter );
    if pPars.fVerbose {
        Gia_ManEquivPrintClasses( pAig, 0, 0 ); }
    Vec_PtrFree( vSimInfo );
    if pPars.fVerbose {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));; }
//    if ( RetValue && pPars->fCheckMiter )
//        Abc_Print( 1, "Cec_ManSeqResimulateCounter(): An output of the miter is asserted!\n" );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Returns the number of POs that are not const0 cands.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCountNonConstOutputs(pAig *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    if pAig.pReprs == nil {
        return -1; }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
        if !Gia_ObjIsConst( pAig, Gia_ObjFaninId0p(pAig, pObj) ) {
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns the number of POs that are not const0 cands.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCheckNonTrivialCands(pAig *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, RetValue int32;
    if pAig.pReprs == nil {
        return 0; }
    // label internal nodes driving POs
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
        Gia_ObjFanin0(pObj).fMark0 = 1;
    }
    // check if there are non-labled equivs
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {
        if Gia_ObjIsCand(pObj) && !pObj.fMark0 && Gia_ObjRepr(pAig, i) != GIA_VOID {
            RetValue = 1;
            break;
        }
    }
    // clean internal nodes driving POs
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
        Gia_ObjFanin0(pObj).fMark0 = 0;
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Performs semiformal refinement of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSeqSemiformal(pAig *Gia_Man_t, pPars *Cec_ParSmf_t) int32 { 
    var nAddFrames int32  = 16; // additional timeframes to simulate
    var nCountNoRef int32;
    var nFramesReal int32;
    var ParsSat Cec_ParSat_t;
    var pParsSat *Cec_ParSat_t = &ParsSat;
    var vSimInfo *Vec_Ptr_t; 
    var vStatus *Vec_Str_t;
    var pState *Abc_Cex_t;
    var pSrm, pReduce, pAux *Gia_Man_t;
    var r, nPats, RetValue int32;
    if pAig.pReprs == nil { 
        Abc_Print( 1, "Cec_ManSeqSemiformal(): Equivalence classes are not available.\n" );
        return -1;
    }
    if Gia_ManRegNum(pAig) == 0 {
        Abc_Print( 1, "Cec_ManSeqSemiformal(): Not a sequential AIG.\n" );
        return -1;
    }
    Gia_ManRandom( 1 );
    // prepare starting pattern
    pState = Abc_CexAlloc( Gia_ManRegNum(pAig), 0, 0 );
    pState.iFrame = -1;
    pState.iPo = -1;
    // prepare SAT solving
    Cec_ManSatSetDefaultParams( pParsSat );
    pParsSat.nBTLimit = pPars.nBTLimit;
    pParsSat.fVerbose = pPars.fVerbose;
    if pParsSat.fVerbose {
        Abc_Print( 1, "Starting: " );
        Gia_ManEquivPrintClasses( pAig, 0, 0 );
    }
    // perform the given number of BMC rounds
    Gia_ManCleanMark0( pAig );
    for  r = 0; r < pPars.nRounds; r++  {
        if !(Cec_ManCheckNonTrivialCands(pAig) != 0) {
            Abc_Print( 1, "Cec_ManSeqSemiformal: There are only trivial equiv candidates left (PO drivers). Quitting.\n" );
            break;
        }
//        Abc_CexPrint( pState );
        // derive speculatively reduced model
//        pSrm = Gia_ManSpecReduceInit( pAig, pState, pPars->nFrames, pPars->fDualOut );
        pSrm = Gia_ManSpecReduceInitFrames( pAig, pState, pPars.nFrames, &nFramesReal, pPars.fDualOut, pPars.nMinOutputs );
        if pSrm == nil {
            Abc_Print( 1, "Quitting refinement because miter could not be unrolled.\n" );
            break;
        }
        assert(Gia_ManRegNum(pSrm) == 0 && Gia_ManPiNum(pSrm) == (Gia_ManPiNum(pAig) * nFramesReal), "Gia_ManRegNum(pSrm) == 0 && Gia_ManPiNum(pSrm) == (Gia_ManPiNum(pAig) * nFramesReal)");
        if pPars.fVerbose {
            Abc_Print( 1, "Unrolled for %d frames.\n", nFramesReal ); }
        // allocate room for simulation info
        vSimInfo = Vec_PtrAllocSimInfo( Gia_ManRegNum(pAig) + 
            Gia_ManPiNum(pAig) * (nFramesReal + nAddFrames), pPars.nWords );
        Cec_ManSeqDeriveInfoInitRandom( vSimInfo, pAig, pState );
        // fill in simulation info with counter-examples
        vStatus = Cec_ManSatSolveSeq( vSimInfo, pSrm, pParsSat, Gia_ManRegNum(pAig), &nPats );
        Vec_StrFree( vStatus );
        Gia_ManStop( pSrm );
        // resimulate and refine the classes
        RetValue = Cec_ManSeqResimulateInfo( pAig, vSimInfo, pState, pPars.fCheckMiter );
        Vec_PtrFree( vSimInfo );
        assert(pState.iPo >= 0, "pState.iPo >= 0"); // hit counter
        pState.iPo = -1;
        if pPars.fVerbose {
            Abc_Print( 1, "BMC = %3d ", nPats );
            Gia_ManEquivPrintClasses( pAig, 0, 0 );
        }

        // write equivalence classes
        Gia_AigerWrite( pAig, "gore.aig", 0, 0, 0 );
        // reduce the model
        pReduce = Gia_ManSpecReduce( pAig, 0, 0, 1, 0, 0 );
        if pReduce {
            pReduce = Gia_ManSeqStructSweep( pAux = pReduce, 1, 1, 0 );
            Gia_ManStop( pAux );
            Gia_AigerWrite( pReduce, "gsrm.aig", 0, 0, 0 );
//            Abc_Print( 1, "Speculatively reduced model was written into file \"%s\".\n", "gsrm.aig" );
//          Gia_ManPrintStatsShort( pReduce );
            Gia_ManStop( pReduce );
        }

        if RetValue != 0 {
            Abc_Print( 1, "Cec_ManSeqSemiformal(): An output of the miter is asserted. Refinement stopped.\n" );
            break;
        }
        // decide when to stop
        if nPats > 0 {
            nCountNoRef = 0;
        } else if nCountNoRef++; nCountNoRef == pPars.nNonRefines {
            break; }
    }
    pState = nil;
    if pPars.fCheckMiter {
        var nNonConsts int32  = Cec_ManCountNonConstOutputs( pAig );
        if nNonConsts != 0 {
            Abc_Print( 1, "The number of POs that are not const-0 candidates = %d.\n", nNonConsts ); }
    }
    return RetValue;
}

//&r s13207.aig; &ps; &equiv; &ps; &semi -R 2 -vm
//&r bug/50/temp.aig; &ps; &equiv -smv; &semi -v
//r mentor/1_05c.blif; st; &get; &ps; &equiv -smv; &semi -mv
//&r bug/50/hdl1.aig; &ps; &equiv -smv; &semi -mv; &srm; &r gsrm.aig; &ps

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

