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

