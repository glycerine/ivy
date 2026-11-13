package abs

import "fmt"

/**CFile****************************************************************

  FileName    [absOut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Abstraction refinement outside of abstraction engines.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absOut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derive a new counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Refines gate-level abstraction using the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGlaRefine(p *Gia_Man_t, pCex *Abc_Cex_t, fMinCut int32, fVerbose int32) int32 {
    extern void Nwk_ManDeriveMinCut( Gia_Man_t * p, int fVerbose );
    var fAddOneLayer int32  = 1;
    var pCexNew *Abc_Cex_t  = nil;
    var pAbs *Gia_Man_t;
    var pAig *Aig_Man_t;
    var pCare *Abc_Cex_t;
    var vPis, vPPis *Vec_Int_t;
    var f, i, iObjId int32;
    var clk abctime  = Abc_Clock();
    var nOnes, Counter int32;
    if p.vGateClasses == nil {
        Abc_Print( 1, "Gia_ManGlaRefine(): Abstraction gate map is missing.\n" );
        return -1;
    }
    // derive abstraction
    pAbs = Gia_ManDupAbsGates( p, p.vGateClasses );
    Gia_ManStop( pAbs );
    pAbs = Gia_ManDupAbsGates( p, p.vGateClasses );
    if Gia_ManPiNum(pAbs) != pCex.nPis {
        Abc_Print( 1, "Gia_ManGlaRefine(): The PI counts in GLA and in CEX do not match.\n" );
        Gia_ManStop( pAbs );
        return -1;
    }
    if !Gia_ManVerifyCex( pAbs, pCex, 0 ) {
        Abc_Print( 1, "Gia_ManGlaRefine(): The initial counter-example is invalid.\n" );
//        Gia_ManStop( pAbs );
//        return -1;
    }
//    else
//        Abc_Print( 1, "Gia_ManGlaRefine(): The initial counter-example is correct.\n" );
    // get inputs
    Gia_ManGlaCollect( p, p.vGateClasses, &vPis, &vPPis, nil, nil );
    assert(Vec_IntSize(vPis) + Vec_IntSize(vPPis) == Gia_ManPiNum(pAbs), "Vec_IntSize(vPis) + Vec_IntSize(vPPis) == Gia_ManPiNum(pAbs)");
    // add missing logic
    if fAddOneLayer != 0 {
        var pObj *Gia_Obj_t;
        // check if this is a real counter-example
        Gia_ObjTerSimSet0( Gia_ManConst0(pAbs) );
        for  f = 0; f <= pCex.iFrame; f++  {
            // inlining of macro Gia_ManForEachPi
            for  i = 0; (i < Gia_ManPiNum(pAbs)) && ((pObj) = Gia_ManCi(pAbs, i)); i++  {
                if i >= Vec_IntSize(vPis) { // PPIs
                    Gia_ObjTerSimSetX( pObj );
                } else if Abc_InfoHasBit(pCex.pData, pCex.nRegs + pCex.nPis * f + i) {
                    Gia_ObjTerSimSet1( pObj );
                } else {
                    Gia_ObjTerSimSet0( pObj ); }
            }
            // inlining of macro Gia_ManForEachRo
            for  i = 0; (i < Gia_ManRegNum(pAbs)) && ((pObj) = Gia_ManCi(pAbs, Gia_ManPiNum(pAbs)+i)); i++  {
                if f == 0 {
                    Gia_ObjTerSimSet0( pObj );
                } else {
                    Gia_ObjTerSimRo( pAbs, pObj ); }
            }
            // inlining of macro Gia_ManForEachAnd
            for  i = 0; (i < pAbs.nObjs) && ((pObj) = Gia_ManObj(pAbs, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
                Gia_ObjTerSimAnd( pObj ); }
            }
            // inlining of macro Gia_ManForEachCo
            for  i = 0; (i < Vec_IntSize(pAbs.vCos)) && ((pObj) = Gia_ManCo(pAbs, i)); i++  {
                Gia_ObjTerSimCo( pObj );
            }
        }
        pObj = Gia_ManPo( pAbs, 0 );
        if Gia_ObjTerSimGet1(pObj) {
            pCexNew = Gia_ManCexRemap( p, pCex, vPis );
            Abc_Print( 1, "Procedure &gla_refine found a real counter-example in frame %d.\n", pCexNew.iFrame );
        }
//        else
//            Abc_Print( 1, "CEX is not real.\n" );
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < pAbs.nObjs) && ((pObj) = Gia_ManObj(pAbs, i)); i++  {
            Gia_ObjTerSimSetC( pObj );
        }
        if pCexNew == nil {
            // grow one layer
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vPPis)) && (((iObjId) = Vec_IntEntry(vPPis, i)), 1); i++  {
                assert(Vec_IntEntry( p.vGateClasses, iObjId ) == 0, "Vec_IntEntry( p.vGateClasses, iObjId ) == 0");
                Vec_IntWriteEntry( p.vGateClasses, iObjId, 1 );
            }
            if fVerbose != 0 {
                Abc_Print( 1, "Additional objects = %d.  ", Vec_IntSize(vPPis) );
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
        }
    
    } else {
        // minimize the CEX
        pAig = Gia_ManToAigSimple( pAbs );
        pCare = Saig_ManCbaFindCexCareBits( pAig, pCex, Vec_IntSize(vPis), fVerbose );
        Aig_ManStop( pAig );
        if pCare == nil {
            Abc_Print( 1, "Counter-example minimization has failed.\n" ); }
        // add new objects to the map
        iObjId = -1;
        for  f = 0; f <= pCare.iFrame; f++  {
            for  i = 0; i < pCare.nPis; i++  {
                if Abc_InfoHasBit( pCare.pData, pCare.nRegs + f * pCare.nPis + i ) {
                    nOnes++;
                    assert(i >= Vec_IntSize(vPis), "i >= Vec_IntSize(vPis)");
                    iObjId = Vec_IntEntry( vPPis, i - Vec_IntSize(vPis) );
                    assert(iObjId > 0 && iObjId < Gia_ManObjNum(p), "iObjId > 0 && iObjId < Gia_ManObjNum(p)");
                    if Vec_IntEntry( p.vGateClasses, iObjId ) > 0 {
                        continue; }
                    assert(Vec_IntEntry( p.vGateClasses, iObjId ) == 0, "Vec_IntEntry( p.vGateClasses, iObjId ) == 0");
                    Vec_IntWriteEntry( p.vGateClasses, iObjId, 1 );
    //                Abc_Print( 1, "Adding object %d.\n", iObjId );
    //                Gia_ObjPrint( p, Gia_ManObj(p, iObjId) );
                    Counter++;
                }
            }
        }
        Abc_CexFree( pCare );
        if fVerbose != 0 {
            Abc_Print( 1, "Essential bits = %d.  Additional objects = %d.  ", nOnes, Counter );
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        // consider the case of SAT
        if iObjId == -1 {
            pCexNew = Gia_ManCexRemap( p, pCex, vPis );
            Abc_Print( 1, "Procedure &gla_refine found a real counter-example in frame %d.\n", pCexNew.iFrame );
        }
    }
    Vec_IntFree( vPis );
    Vec_IntFree( vPPis );
    Gia_ManStop( pAbs );
    if pCexNew != nil {
        p.pCexSeq = nil;
        p.pCexSeq = pCexNew;
        return 0;
    }
    // extract abstraction to include min-cut
    if fMinCut != 0 {
        Nwk_ManDeriveMinCut( p, fVerbose ); }
    return -1;
}





/**Function*************************************************************

  Synopsis    [Resimulates the counter-example and returns flop values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Verify counter-example starting in the given timeframe.]

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
func Gia_ManNewRefine(p *Gia_Man_t, pCex *Abc_Cex_t, iFrameStart int32, iFrameExtra int32, fVerbose int32) int32 {
    var pAbs, pNew *Gia_Man_t;
    var vFlops, vInit *Vec_Int_t;
    var vCopy *Vec_Int_t;
//    abctime clk = Abc_Clock();
    var RetValue int32;
    p.pCexSeq = nil;
    if p.vGateClasses == nil {
        Abc_Print( 1, "Gia_ManNewRefine(): Abstraction gate map is missing.\n" );
        return -1;
    }
    vCopy = Vec_IntDup( p.vGateClasses );
    Abc_Print( 1, "Refining with %d-frame CEX, starting in frame %d, with %d extra frames.\n", pCex.iFrame, iFrameStart, iFrameExtra );
    // derive abstraction
    pAbs = Gia_ManDupAbsGates( p, p.vGateClasses );
    Gia_ManStop( pAbs );
    pAbs = Gia_ManDupAbsGates( p, p.vGateClasses );
    if Gia_ManPiNum(pAbs) != pCex.nPis {
        Abc_Print( 1, "Gia_ManNewRefine(): The PI counts in GLA and in CEX do not match.\n" );
        Gia_ManStop( pAbs );
        Vec_IntFree( vCopy );
        return -1;
    }
    // get the state in frame iFrameStart
    vInit = Gia_ManGetStateAndCheckCex( pAbs, pCex, iFrameStart );
    if vInit == nil {
        Abc_Print( 1, "Gia_ManNewRefine(): The initial counter-example is invalid.\n" );
        Gia_ManStop( pAbs );
        Vec_IntFree( vCopy );
        return -1;
    }
    if fVerbose != 0 {
        Abc_Print( 1, "Gia_ManNewRefine(): The initial counter-example is correct.\n" ); }
    // get inputs
    Gia_ManGlaCollect( p, p.vGateClasses, nil, nil, &vFlops, nil );
//    assert( Vec_IntSize(vPis) + Vec_IntSize(vPPis) == Gia_ManPiNum(pAbs) );
    Gia_ManStop( pAbs );
//Vec_IntPrint( vFlops );
//Vec_IntPrint( vInit );
    // transform the manager to have new init state
    pNew = Gia_ManTransformFlops( p, vFlops, vInit );
    Vec_IntFree( vFlops );
    Vec_IntFree( vInit );
    // verify abstraction
    {
        var pAbs *Gia_Man_t  = Gia_ManDupAbsGates( pNew, p.vGateClasses );
        Gia_ManCheckCex( pAbs, pCex, iFrameStart );
        Gia_ManStop( pAbs );
    }
    // transfer abstraction
    assert(pNew.vGateClasses == nil, "pNew.vGateClasses == nil");
    pNew.vGateClasses = Vec_IntDup( p.vGateClasses );
    // perform abstraction for the new AIG
    {
        var Pars Abs_Par_t;
        var pPars *Abs_Par_t = &Pars;
        Abs_ParSetDefaults( pPars );
        pPars.nFramesMax = pCex.iFrame - iFrameStart + 1 + iFrameExtra;
        pPars.fVerbose = fVerbose;
        RetValue = Gia_ManPerformGla( pNew, pPars );
        if RetValue == 0 { // spurious SAT
        
            Vec_IntFreeP( &pNew.vGateClasses );
            pNew.vGateClasses = Vec_IntDup( vCopy );
        }
    }
    // move the abstraction map
    Vec_IntFreeP( &p.vGateClasses );
    p.vGateClasses = pNew.vGateClasses;
    pNew.vGateClasses = nil;
    // cleanup
    Gia_ManStop( pNew );
    Vec_IntFree( vCopy );
    return -1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

