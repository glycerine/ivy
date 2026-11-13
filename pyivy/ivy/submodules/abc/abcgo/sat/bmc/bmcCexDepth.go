package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcCexDepth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [CEX depth minimization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcCexDepth.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Abc_Cex_t * Bmc_CexInnerStates( Gia_Man_t * p, Abc_Cex_t * pCex, Abc_Cex_t ** ppCexImpl, int fVerbose );

// extern Abc_Cex_t * Bmc_CexCareBits( Gia_Man_t * p, Abc_Cex_t * pCexState, Abc_Cex_t * pCexImpl, Abc_Cex_t * pCexEss, int fFindAll, int fVerbose );

// extern int Bmc_CexVerify( Gia_Man_t * p, Abc_Cex_t * pCex, Abc_Cex_t * pCexCare );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes CE-induced network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexBuildNetwork2(p *Gia_Man_t, pCex *Abc_Cex_t, fStart int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var fCompl0, fCompl1 int32;
    var i, k, iBit int32;
    assert(pCex.nRegs == 0, "pCex.nRegs == 0");
    assert(pCex.nPis == Gia_ManCiNum(p), "pCex.nPis == Gia_ManCiNum(p)");
    assert(fStart <= pCex.iFrame, "fStart <= pCex.iFrame");
    // start the manager
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( "unate" );
    // set const0
    Gia_ManConst0(p).fMark0 = 0; // value
    Gia_ManConst0(p).fMark1 = 1; // care
    Gia_ManConst0(p).Value  = ~0;
    // set init state
    // inlining of macro Gia_ManForEachRi
    for k = 0; k < Gia_ManRegNum(p); k++ {
    	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+k)
    	if pObj == nil { break }
        pObj.fMark0 = Abc_InfoHasBit( pCex.pData, pCex.nRegs + fStart * Gia_ManCiNum(p) + Gia_ManPiNum(p) + k );
        pObj.fMark1 = 0;
        pObj.Value  = Abc_LitNotCond( Gia_ManAppendCi(pNew), !pObj.fMark0 );
    }
    Gia_ManHashAlloc( pNew );
    iBit = pCex.nRegs + fStart * Gia_ManCiNum(p);
    for  i = fStart; i <= pCex.iFrame; i++  {
        //  primary inputs
        // inlining of macro Gia_ManForEachPi
        for k = 0; k < Gia_ManPiNum(p); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Abc_InfoHasBit( pCex.pData, iBit++ );
            pObj.fMark1 = 1;
            pObj.Value  = ~0;
        }
        iBit += Gia_ManRegNum(p);
        // transfer 
        // inlining of macro Gia_ManForEachRiRo
        for k = 0; k < Gia_ManRegNum(p); k++ {
        	pObjRi = Gia_ManCo(p, Gia_ManPoNum(p)+k)
        	if pObjRi == nil { break }
        	pObjRo = Gia_ManCi(p, Gia_ManPiNum(p)+k)
        	if pObjRo == nil { break }
            pObjRo.fMark0 = pObjRi.fMark0;
            pObjRo.fMark1 = pObjRi.fMark1;
            pObjRo.Value  = pObjRi.Value;
        }
        // internal nodes
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
            pObj.fMark0 = fCompl0 & fCompl1;
            if pObj.fMark0 {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1 & Gia_ObjFanin1(pObj).fMark1;
            } else if !fCompl0 && !(fCompl1 != 0) {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1 | Gia_ObjFanin1(pObj).fMark1;
            } else if !fCompl0 {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1;
            } else if !(fCompl1 != 0) {
                pObj.fMark1 = Gia_ObjFanin1(pObj).fMark1;
            } else { assert(false, "0"); }
            pObj.Value = ~0;
            if pObj.fMark1 {
                continue; }
            if pObj.fMark0 {
                pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else if !fCompl0 && !(fCompl1 != 0) {
                pObj.Value = Gia_ManHashOr( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else if !fCompl0 {
                pObj.Value = Gia_ObjFanin0(pObj).Value;
            } else if !(fCompl1 != 0) {
                pObj.Value = Gia_ObjFanin1(pObj).Value;
            } else { assert(false, "0"); }
            assert(pObj.Value > 0, "pObj.Value > 0");
        }
        }
        // combinational outputs
        // inlining of macro Gia_ManForEachCo
        for k = 0; k < Vec_IntSize(p.vCos); k++ {
        	pObj = Gia_ManCo(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1;
            pObj.Value  = Gia_ObjFanin0(pObj).Value;
        }
    }
    Gia_ManHashStop( pNew );
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // create primary output
    pObj = Gia_ManPo(p, pCex.iPo);
    assert(pObj.fMark0 == 1, "pObj.fMark0 == 1");
    assert(pObj.fMark1 == 0, "pObj.fMark1 == 0");
    assert(pObj.Value > 0, "pObj.Value > 0");
    Gia_ManAppendCo( pNew, pObj.Value );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Bmc_CexBuildNetwork2_(p *Gia_Man_t, pCex *Abc_Cex_t, fStart int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var fCompl0, fCompl1 int32;
    var i, k, iBit int32;
    assert(pCex.nRegs == 0, "pCex.nRegs == 0");
    assert(pCex.nPis == Gia_ManCiNum(p), "pCex.nPis == Gia_ManCiNum(p)");
    assert(fStart <= pCex.iFrame, "fStart <= pCex.iFrame");
    // start the manager
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( "unate" );
    // set const0
    Gia_ManConst0(p).fMark0 = 0; // value
    Gia_ManConst0(p).Value  = 1;
    // set init state
    // inlining of macro Gia_ManForEachRi
    for k = 0; k < Gia_ManRegNum(p); k++ {
    	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+k)
    	if pObj == nil { break }
        pObj.fMark0 = Abc_InfoHasBit( pCex.pData, pCex.nRegs + fStart * Gia_ManCiNum(p) + Gia_ManPiNum(p) + k );
        pObj.Value  = Abc_LitNotCond( Gia_ManAppendCi(pNew), !pObj.fMark0 );
    }
    Gia_ManHashAlloc( pNew );
    iBit = pCex.nRegs + fStart * Gia_ManCiNum(p);
    for  i = fStart; i <= pCex.iFrame; i++  {
        //  primary inputs
        // inlining of macro Gia_ManForEachPi
        for k = 0; k < Gia_ManPiNum(p); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Abc_InfoHasBit( pCex.pData, iBit++ );
            pObj.Value  = 1;
        }
        iBit += Gia_ManRegNum(p);
        // transfer 
        // inlining of macro Gia_ManForEachRiRo
        for k = 0; k < Gia_ManRegNum(p); k++ {
        	pObjRi = Gia_ManCo(p, Gia_ManPoNum(p)+k)
        	if pObjRi == nil { break }
        	pObjRo = Gia_ManCi(p, Gia_ManPiNum(p)+k)
        	if pObjRo == nil { break }
            pObjRo.fMark0 = pObjRi.fMark0;
            pObjRo.Value  = pObjRi.Value;
        }
        // internal nodes
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
            pObj.fMark0 = fCompl0 & fCompl1;
            if pObj.fMark0 {
                pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else if !fCompl0 && !(fCompl1 != 0) {
                pObj.Value = Gia_ManHashOr( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else if !fCompl0 {
                pObj.Value = Gia_ObjFanin0(pObj).Value;
            } else if !(fCompl1 != 0) {
                pObj.Value = Gia_ObjFanin1(pObj).Value;
            } else { assert(false, "0"); }
        }
        }
        // combinational outputs
        // inlining of macro Gia_ManForEachCo
        for k = 0; k < Vec_IntSize(p.vCos); k++ {
        	pObj = Gia_ManCo(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            pObj.Value  = Gia_ObjFanin0(pObj).Value;
        }
    }
    Gia_ManHashStop( pNew );
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // create primary output
    pObj = Gia_ManPo(p, pCex.iPo);
    assert(pObj.fMark0 == 1, "pObj.fMark0 == 1");
    assert(pObj.Value > 0, "pObj.Value > 0");
    Gia_ManAppendCo( pNew, pObj.Value );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

func Bmc_CexBuildNetwork2Test(p *Gia_Man_t, pCex *Abc_Cex_t, nFramesMax int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var vCones *Vec_Ptr_t;
    var clk abctime  = Abc_Clock();
    var i int32;
    nFramesMax = Abc_MinInt( nFramesMax, pCex.iFrame );
    fmt.Printf("Processing CEX in frame %d (max frames %d).\n", pCex.iFrame, nFramesMax);
    vCones = Vec_PtrAlloc( nFramesMax );
    for  i = pCex.iFrame; i > pCex.iFrame - nFramesMax; i--  {
        fmt.Printf("Frame %5d : ", i);
        pNew = Bmc_CexBuildNetwork2_( p, pCex, i );
        Gia_ManPrintStats( pNew, nil );
        Vec_PtrPush( vCones, pNew );
    }
    pNew = Gia_ManDupAppendCones( p, (Gia_Man_t **)Vec_PtrArray(vCones), Vec_PtrSize(vCones), 1 );
    Gia_AigerWrite( pNew, "miter2.aig", 0, 0, 0 );
//Bmc_CexDumpAogStats( pNew, Abc_Clock() - clk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCones)) && (((pTemp) = (Gia_Man_t *)Vec_PtrEntry(vCones, i)), 1); i++  {
        Gia_ManStop( pTemp );
    }
    Vec_PtrFree( vCones );
    fmt.Printf("GIA with additional properties is written into \"miter2.aig\".\n");
//    printf( "CE-induced network is written into file \"unate.aig\".\n" );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
//    Gia_ManStop( pNew );
    return pNew;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexDepthTest(p *Gia_Man_t, pCex *Abc_Cex_t, nFrames int32, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var clk abctime  = Abc_Clock();
    var pCexImpl *Abc_Cex_t    = nil;
    var pCexStates *Abc_Cex_t  = Bmc_CexInnerStates( p, pCex, &pCexImpl, fVerbose );
    var pCexCare *Abc_Cex_t    = Bmc_CexCareBits( p, pCexStates, pCexImpl, nil, 1, fVerbose );
//    Abc_Cex_t * pCexEss, * pCexMin;

    if !Bmc_CexVerify( p, pCex, pCexCare ) {
        fmt.Printf("Counter-example care-set verification has failed.\n"); }

//    pCexEss = Bmc_CexEssentialBits( p, pCexStates, pCexCare, fVerbose );
//    pCexMin = Bmc_CexCareBits( p, pCexStates, pCexImpl, pCexEss, 0, fVerbose );

//    if ( !Bmc_CexVerify( p, pCex, pCexMin ) )
//        printf( "Counter-example min-set verification has failed.\n" );

//    Bmc_CexDumpStats( p, pCex, pCexCare, pCexEss, pCexMin, Abc_Clock() - clk );

    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    pNew = Bmc_CexBuildNetwork2Test( p, pCexStates, nFrames );
//    Bmc_CexPerformUnrollingTest( p, pCex );

    Abc_CexFreeP( &pCexStates );
    Abc_CexFreeP( &pCexImpl );
    Abc_CexFreeP( &pCexCare );
//    Abc_CexFreeP( &pCexEss );
//    Abc_CexFreeP( &pCexMin );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

