package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcCexMin1.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [CEX minimization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcCexMin1.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/ioa/ioa.h"
//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexMinPerform(pAig *Aig_Man_t, pCex *Abc_Cex_t) *Abc_Cex_t {
    var fReasonPi int32;

    var pCexMin *Abc_Cex_t  = nil;
    var pManNew *Aig_Man_t  = nil;
    var vFrameReas *Vec_Vec_t;
    vFrameReas = Saig_ManCexMinComputeReason( pAig, pCex, fReasonPi );
    fmt.Printf("Reason size = %d.  Ave = %d.\n", Vec_VecSizeSize(vFrameReas), Vec_VecSizeSize(vFrameReas)/(pCex.iFrame + 1));

    // try reducing the frames
    if !(fReasonPi != 0) {
        var pFileName string  = "aigcube.aig";
        pManNew = Saig_ManCexMinDupWithCubes( pAig, vFrameReas );
        Ioa_WriteAiger( pManNew, pFileName, 0, 0 );
        Aig_ManStop( pManNew );
        fmt.Printf("Intermediate AIG is written into file \"%s\".\n", pFileName);
    }

    // find reduced counter-example
    Vec_VecFree( vFrameReas );
    return pCexMin;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

