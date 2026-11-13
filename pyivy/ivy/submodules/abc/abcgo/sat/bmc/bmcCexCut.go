package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcCexCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Derives characterization of bad states.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcCexCut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Generate GIA for target bad states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_GiaTargetStates(p *Gia_Man_t, pCex *Abc_Cex_t, iFrBeg int32, iFrEnd int32, fCombOnly int32, fUseOne int32, fAllFrames int32, fVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var vInitNew *Vec_Bit_t;

    if iFrBeg < 0 { fmt.Printf("Starting frame is less than 0.\n"); return nil; }
    if iFrEnd < 0 { fmt.Printf("Stopping frame is less than 0.\n"); return nil; }
    if iFrBeg > pCex.iFrame { fmt.Printf("Starting frame is more than the last frame of CEX (%d).\n", pCex.iFrame); return nil; }
    if iFrEnd > pCex.iFrame { fmt.Printf("Stopping frame is more than the last frame of CEX (%d).\n", pCex.iFrame); return nil; }
    if iFrBeg > iFrEnd { fmt.Printf("Starting frame (%d) should be less than stopping frame (%d).\n", iFrBeg, iFrEnd); return nil; }
    assert(iFrBeg >= 0 && iFrBeg <= pCex.iFrame, "iFrBeg >= 0 && iFrBeg <= pCex.iFrame");
    assert(iFrEnd >= 0 && iFrEnd <= pCex.iFrame, "iFrEnd >= 0 && iFrEnd <= pCex.iFrame");
    assert(iFrBeg < iFrEnd, "iFrBeg < iFrEnd");

    if fUseOne != 0 {
        pNew = Bmc_GiaGenerateGiaOne( p, pCex, &vInitNew, iFrBeg, iFrEnd );
    } else if fAllFrames != 0 {
        pNew = Bmc_GiaGenerateGiaAllFrames( p, pCex, &vInitNew, iFrBeg, iFrEnd );
    } else { 
        pNew = Bmc_GiaGenerateGiaAllOne( p, pCex, &vInitNew, iFrBeg, iFrEnd ); }

    if !(fCombOnly != 0) {
        // create new GIA
        pNew = Gia_ManDupWithNewPo( p, pTemp = pNew ); 
        Gia_ManStop( pTemp );

        // create new initial state
        pNew = Gia_ManDupFlip( pTemp = pNew, Vec_BitArray(vInitNew) );
        Gia_ManStop( pTemp );
    }

    Vec_BitFree( vInitNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Generate AIG for target bad states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_AigTargetStates(p *Aig_Man_t, pCex *Abc_Cex_t, iFrBeg int32, iFrEnd int32, fCombOnly int32, fUseOne int32, fAllFrames int32, fVerbose int32) *Aig_Man_t {
    var pAig *Aig_Man_t;
    var pGia, pRes *Gia_Man_t;
    pGia = Gia_ManFromAigSimple( p );
    if !Gia_ManVerifyCex( pGia, pCex, 0 ) {
        Abc_Print( 1, "Current CEX does not fail AIG \"%s\".\n", p.pName ); 
        Gia_ManStop( pGia );
        return nil;
    }
    pRes = Bmc_GiaTargetStates( pGia, pCex, iFrBeg, iFrEnd, fCombOnly, fUseOne, fAllFrames, fVerbose );
    Gia_ManStop( pGia );
    pAig = Gia_ManToAigSimple( pRes );
    Gia_ManStop( pRes );
    return pAig;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

