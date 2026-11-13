package abs

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigAbsCba.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [CEX-based abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigAbsCba.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// local manager
   


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Selects the best flops from the given array.]

  Description [Selects the best 'nFfsToSelect' flops among the array 
  'vAbsFfsToAdd' of flops that should be added to the abstraction.
  To this end, this procedure simulates the original AIG (pAig) using
  the given CEX (pAbsCex), which was detected for the abstraction.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Duplicate with literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Maps array of frame PI IDs into array of additional PI IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns reasons for the property to fail.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCbaFindReason_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vPrios *Vec_Int_t, vReasons *Vec_Int_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsConst1(pObj) {
        return; }
    if Aig_ObjIsCi(pObj) {
        Vec_IntPush( vReasons, Aig_ObjCioId(pObj) );
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    if pObj.fPhase {
        Saig_ManCbaFindReason_rec( p, Aig_ObjFanin0(pObj), vPrios, vReasons );
        Saig_ManCbaFindReason_rec( p, Aig_ObjFanin1(pObj), vPrios, vReasons );
    
    } else {
        var fPhase0 int32  = Aig_ObjFaninC0(pObj) ^ Aig_ObjFanin0(pObj).fPhase;
        var fPhase1 int32  = Aig_ObjFaninC1(pObj) ^ Aig_ObjFanin1(pObj).fPhase;
        assert(!(fPhase0 != 0) || !(fPhase1 != 0), "!fPhase0 || !fPhase1");
        if !(fPhase0 != 0) && (fPhase1 != 0) {
            Saig_ManCbaFindReason_rec( p, Aig_ObjFanin0(pObj), vPrios, vReasons );
        } else if (fPhase0 != 0) && !(fPhase1 != 0) {
            Saig_ManCbaFindReason_rec( p, Aig_ObjFanin1(pObj), vPrios, vReasons );
        } else {
            var iPrio0 int32  = Vec_IntEntry( vPrios, Aig_ObjFaninId0(pObj) );
            var iPrio1 int32  = Vec_IntEntry( vPrios, Aig_ObjFaninId1(pObj) );
            if iPrio0 <= iPrio1 {
                Saig_ManCbaFindReason_rec( p, Aig_ObjFanin0(pObj), vPrios, vReasons );
            } else {
                Saig_ManCbaFindReason_rec( p, Aig_ObjFanin1(pObj), vPrios, vReasons ); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Returns reasons for the property to fail.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCbaFindReason(p *Saig_ManCba_t) *Vec_Int_t {
    var pObj *Aig_Obj_t;
    var vPrios, vReasons *Vec_Int_t;
    var i int32;

    // set PI values according to CEX
    vPrios = Vec_IntStartFull( Aig_ManObjNumMax(p.pFrames) );
    Aig_ManConst1(p.pFrames).fPhase = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vCis, i)), 1); i++  {
        var iInput int32  = Vec_IntEntry( p.vMapPiF2A, 2*i );
        var iFrame int32  = Vec_IntEntry( p.vMapPiF2A, 2*i+1 );
        pObj.fPhase = Abc_InfoHasBit( p.pCex.pData, p.pCex.nRegs + p.pCex.nPis * iFrame + iInput );
        Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), i );
    }

    // traverse and set the priority
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pFrames.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pFrames.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        var fPhase0 int32  = Aig_ObjFaninC0(pObj) ^ Aig_ObjFanin0(pObj).fPhase;
        var fPhase1 int32  = Aig_ObjFaninC1(pObj) ^ Aig_ObjFanin1(pObj).fPhase;
        var iPrio0 int32   = Vec_IntEntry( vPrios, Aig_ObjFaninId0(pObj) );
        var iPrio1 int32   = Vec_IntEntry( vPrios, Aig_ObjFaninId1(pObj) );
        pObj.fPhase = fPhase0 && fPhase1;
        if (fPhase0 != 0) && (fPhase1 != 0) { // both are one
            Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), Abc_MaxInt(iPrio0, iPrio1) );
        } else if !(fPhase0 != 0) && (fPhase1 != 0) { 
            Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), iPrio0 );
        } else if (fPhase0 != 0) && !(fPhase1 != 0) {
            Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), iPrio1 );
        } else { // both are zero
            Vec_IntWriteEntry( vPrios, Aig_ObjId(pObj), Abc_MinInt(iPrio0, iPrio1) ); }
    }
    }
    // check the property output
    pObj = Aig_ManCo( p.pFrames, 0 );
    pObj.fPhase = Aig_ObjFaninC0(pObj) ^ Aig_ObjFanin0(pObj).fPhase;
    assert(!pObj.fPhase, "!pObj.fPhase");

    // select the reason
    vReasons = Vec_IntAlloc( 100 );
    Aig_ManIncrementTravId( p.pFrames );
    Saig_ManCbaFindReason_rec( p.pFrames, Aig_ObjFanin0(pObj), vPrios, vReasons );
    Vec_IntFree( vPrios );
//    assert( !Aig_ObjIsTravIdCurrent(p->pFrames, Aig_ManConst1(p->pFrames)) );
    return vReasons;
}


/**Function*************************************************************

  Synopsis    [Collect nodes in the unrolled timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Creates refinement manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Destroys refinement manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCbaShrink(p *Saig_ManCba_t) {
    var pManNew *Aig_Man_t;
    var pObjLi, pObjFrame *Aig_Obj_t;
    var vLevel, vLevel2 *Vec_Int_t;
    var f, k, ObjId, Lit int32;
    // assuming that important objects are labeled in Saig_ManCbaFindReason()
    // inlining of macro Vec_VecForEachLevelInt
    for  f = 0; (f < Vec_VecSize(p.vReg2Frame)) && (((vLevel) = Vec_VecEntryInt(p.vReg2Frame, f)), 1); f++  {
        // inlining of macro Vec_IntForEachEntryDouble
        for  k = 0; (k+1 < Vec_IntSize(vLevel)) && (((ObjId) = Vec_IntEntry(vLevel, k)), 1) && (((Lit) = Vec_IntEntry(vLevel, k+1)), 1); k += 2  {
            pObjFrame = Aig_ManObj( p.pFrames, Abc_Lit2Var(Lit) );
            if pObjFrame == nil || (!Aig_ObjIsConst1(pObjFrame) && !Aig_ObjIsTravIdCurrent(p.pFrames, pObjFrame)) {
                continue; }
            pObjLi = Aig_ManObj( p.pAig, ObjId );
            assert(Saig_ObjIsLi(p.pAig, pObjLi), "Saig_ObjIsLi(p.pAig, pObjLi)");
            Vec_VecPushInt( p.vReg2Value, f, Abc_Var2Lit( Aig_ObjCioId(pObjLi) - Saig_ManPoNum(p.pAig), Abc_LitIsCompl(Lit) ^ !pObjFrame.fPhase ) );
        }
    }
    // print statistics
    // inlining of macro Vec_VecForEachLevelInt
    for  k = 0; (k < Vec_VecSize(p.vReg2Frame)) && (((vLevel) = Vec_VecEntryInt(p.vReg2Frame, k)), 1); k++  {
        vLevel2 = Vec_VecEntryInt( p.vReg2Value, k );
        fmt.Printf("Level = %4d   StateBits = %4d (%6.2f %%)  CareBits = %4d (%6.2f %%)\n", k, Vec_IntSize(vLevel)/2, 100.0 * (Vec_IntSize(vLevel)/2) / Aig_ManRegNum(p.pAig), Vec_IntSize(vLevel2), 100.0 * Vec_IntSize(vLevel2) / Aig_ManRegNum(p.pAig));
    }
    // try reducing the frames
    pManNew = Saig_ManDupWithCubes( p.pAig, p.vReg2Value );
//    Ioa_WriteAiger( pManNew, "aigcube.aig", 0, 0 );
    Aig_ManStop( pManNew );
}



/**Function*************************************************************

  Synopsis    [SAT-based refinement of the counter-example.]

  Description [The first parameter (nInputs) indicates how many first 
  primary inputs to skip without considering as care candidates.]
               

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCbaFindCexCareBits(pAig *Aig_Man_t, pCex *Abc_Cex_t, nInputs int32, fVerbose int32) *Abc_Cex_t {
    var p *Saig_ManCba_t;
    var vReasons *Vec_Int_t;
    var pCare *Abc_Cex_t;
    var clk abctime  = Abc_Clock();

    clk = Abc_Clock();
    p = Saig_ManCbaStart( pAig, pCex, nInputs, fVerbose );

//    p->vReg2Frame = Vec_VecStart( pCex->iFrame );
//    p->vReg2Value = Vec_VecStart( pCex->iFrame );
    p.pFrames = Saig_ManCbaUnrollWithCex( pAig, pCex, nInputs, &p.vMapPiF2A, &p.vReg2Frame );
    vReasons = Saig_ManCbaFindReason( p );
    if p.vReg2Frame {
        Saig_ManCbaShrink( p ); }


//if ( fVerbose )
//Aig_ManPrintStats( p->pFrames );

    if fVerbose != 0 {
        var vRes *Vec_Int_t  = Saig_ManCbaReason2Inputs( p, vReasons );
        fmt.Printf("Frame PIs = %4d (essential = %4d)   AIG PIs = %4d (essential = %4d)   ", Aig_ManCiNum(p.pFrames), Vec_IntSize(vReasons), Saig_ManPiNum(p.pAig) - p.nInputs, Vec_IntSize(vRes));
        Vec_IntFree( vRes );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

    pCare = Saig_ManCbaReason2Cex( p, vReasons );
    Vec_IntFree( vReasons );
    Saig_ManCbaStop( p );

if fVerbose != 0 {
fmt.Printf("Real ");
Abc_CexPrintStats( pCex );
}
if fVerbose != 0 {
fmt.Printf("Care ");
Abc_CexPrintStats( pCare );
}
/*
    // verify the reduced counter-example using ternary simulation
    if ( !Saig_ManCexVerifyUsingTernary( pAig, pCex, pCare ) )
        printf( "Saig_ManCbaFindCexCareBits(): Minimized counter-example verification has failed!!!\n" );
    else if ( fVerbose )
        printf( "Saig_ManCbaFindCexCareBits(): Minimized counter-example verification is successful.\n" );
*/
    Aig_ManCleanMarkAB( pAig );
    return pCare;
}

/**Function*************************************************************

  Synopsis    [Returns the array of PIs for flops that should not be absracted.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCbaFilterInputs(pAig *Aig_Man_t, iFirstFlopPi int32, pCex *Abc_Cex_t, fVerbose int32) *Vec_Int_t {
    var p *Saig_ManCba_t;
    var vRes, vReasons *Vec_Int_t;
    var clk abctime;
    if Saig_ManPiNum(pAig) != pCex.nPis {
        fmt.Printf("Saig_ManCbaFilterInputs(): The PI count of AIG (%d) does not match that of cex (%d).\n", Aig_ManCiNum(pAig), pCex.nPis);
        return nil;
    }

clk = Abc_Clock();
    p = Saig_ManCbaStart( pAig, pCex, iFirstFlopPi, fVerbose );
    p.pFrames = Saig_ManCbaUnrollWithCex( pAig, pCex, iFirstFlopPi, &p.vMapPiF2A, &p.vReg2Frame );
    vReasons = Saig_ManCbaFindReason( p );
    vRes = Saig_ManCbaReason2Inputs( p, vReasons );
    if fVerbose != 0 {
        fmt.Printf("Frame PIs = %4d (essential = %4d)   AIG PIs = %4d (essential = %4d)   ", Aig_ManCiNum(p.pFrames), Vec_IntSize(vReasons), Saig_ManPiNum(p.pAig) - p.nInputs, Vec_IntSize(vRes));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

    Vec_IntFree( vReasons );
    Saig_ManCbaStop( p );
    return vRes;
}




/**Function*************************************************************

  Synopsis    [Checks the abstracted model for a counter-example.]

  Description [Returns the array of abstracted flops that should be added
  to the abstraction.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCbaPerform(pAbs *Aig_Man_t, nInputs int32, pPars *Saig_ParBmc_t) *Vec_Int_t {
    var vAbsFfsToAdd *Vec_Int_t;
    var RetValue int32;
    var clk abctime  = Abc_Clock();
//    assert( pAbs->nRegs > 0 );
    // perform BMC
    RetValue = Saig_ManBmcScalable( pAbs, pPars );
    if RetValue == -1 { // time out - nothing to add
    
        fmt.Printf("Resource limit is reached during BMC.\n");
        assert(pAbs.pSeqModel == nil, "pAbs.pSeqModel == nil");
        return Vec_IntAlloc( 0 );
    }
    if pAbs.pSeqModel == nil {
        fmt.Printf("BMC did not detect a CEX with the given depth.\n");
        return Vec_IntAlloc( 0 );
    }
    if pPars.fVerbose {
        Abc_CexPrintStats( pAbs.pSeqModel ); }
    // CEX is detected - refine the flops
    vAbsFfsToAdd = Saig_ManCbaFilterInputs( pAbs, nInputs, pAbs.pSeqModel, pPars.fVerbose );
    if Vec_IntSize(vAbsFfsToAdd) == 0 {
        Vec_IntFree( vAbsFfsToAdd );
        return nil;
    }
    if pPars.fVerbose {
        fmt.Printf("Adding %d registers to the abstraction (total = %d).  ", Vec_IntSize(vAbsFfsToAdd), Aig_ManRegNum(pAbs)+Vec_IntSize(vAbsFfsToAdd));
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    return vAbsFfsToAdd;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

