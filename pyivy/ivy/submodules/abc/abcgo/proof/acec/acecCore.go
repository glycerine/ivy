package acec

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [acecCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecCore.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "proof/cec/cec.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const TRUTH_UNUSED = 0x1234567812345678

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_ComputeEquivClasses(pOne *Gia_Man_t, pTwo *Gia_Man_t, pvMap1 **Vec_Int_t, pvMap2 **Vec_Int_t) {
    var clk abctime  = Abc_Clock();
    var pBase, pRepr *Gia_Man_t;
    pBase = Acec_CommonStart( nil, pOne );
    pBase = Acec_CommonStart( pBase, pTwo );
    Acec_CommonFinish( pBase );
    //Gia_ManShow( pBase, NULL, 0, 0, 0 );
    pRepr = Gia_ManComputeGiaEquivs( pBase, 100, 0 );
    *pvMap1 = Acec_CountRemap( pOne, pBase );
    *pvMap2 = Acec_CountRemap( pTwo, pBase );
    Gia_ManStop( pBase );
    Gia_ManStop( pRepr );
    fmt.Printf("Finished computing equivalent nodes.  ");
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}
func Acec_Solve(pGia0 *Gia_Man_t, pGia1 *Gia_Man_t, pPars *Acec_ParCec_t) int32 {
    var status int32  = -1;
    var clk abctime  = Abc_Clock();
    var pMiter *Gia_Man_t;
    var pGia0n *Gia_Man_t = pGia0;
    var pGia1n *Gia_Man_t = pGia1;
    var ParsCec Cec_ParCec_t;
    var pCecPars *Cec_ParCec_t = &ParsCec;
//    Vec_Bit_t * vIgnore0 = pPars->fBooth ? Acec_BoothFindPPG(pGia0) : NULL;
//    Vec_Bit_t * vIgnore1 = pPars->fBooth ? Acec_BoothFindPPG(pGia1) : NULL;
//    Acec_Box_t * pBox0 = Acec_DeriveBox( pGia0, vIgnore0, 0, 0, pPars->fVerbose );
//    Acec_Box_t * pBox1 = Acec_DeriveBox( pGia1, vIgnore1, 0, 0, pPars->fVerbose );
//    Vec_BitFreeP( &vIgnore0 );
//    Vec_BitFreeP( &vIgnore1 );
    var pBox0 *Acec_Box_t  = Acec_ProduceBox( pGia0, pPars.fVerbose );
    var pBox1 *Acec_Box_t  = Acec_ProduceBox( pGia1, pPars.fVerbose );
    if pBox0 == nil || pBox1 == nil { // cannot match
        fmt.Printf("Cannot find arithmetic boxes in both LHS and RHS. Trying regular CEC.\n");
    } else if !(Acec_MatchBoxes( pBox0, pBox1 ) != 0) { // cannot find matching
        fmt.Printf("Cannot match arithmetic boxes in LHS and RHS. Trying regular CEC.\n");
    } else {
        pGia0n = Acec_InsertBox( pBox0, 0 );
        pGia1n = Acec_InsertBox( pBox1, 0 );
        fmt.Printf("Matching of adder trees in LHS and RHS succeeded.  ");
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        // remove the last output
        Gia_ManPatchCoDriver( pGia0n, Gia_ManCoNum(pGia0n)-1, 0 );
        Gia_ManPatchCoDriver( pGia1n, Gia_ManCoNum(pGia1n)-1, 0 );

        Gia_ManPatchCoDriver( pGia0n, Gia_ManCoNum(pGia0n)-2, 0 );
        Gia_ManPatchCoDriver( pGia1n, Gia_ManCoNum(pGia1n)-2, 0 );
    }
    // solve regular CEC problem 
    Cec_ManCecSetDefaultParams( pCecPars );
    pCecPars.nBTLimit = pPars.nBTLimit;
    pMiter = Gia_ManMiter( pGia0n, pGia1n, 0, 1, 0, 0, pPars.fVerbose );
    if pMiter != nil {
        var fDumpMiter int32;
        if fDumpMiter != 0 {
            Abc_Print( 0, "The verification miter is written into file \"%s\".\n", "acec_miter.aig" );
            Gia_AigerWrite( pMiter, "acec_miter.aig", 0, 0, 0 );
        }
        status = Cec_ManVerify( pMiter, pCecPars );
        // inlining of macro ABC_SWAP
        { var t *Abc_Cex_t  = pGia0.pCexComb; pGia0.pCexComb = pMiter.pCexComb; pMiter.pCexComb = t; };
        Gia_ManStop( pMiter );
    
    } else {
        fmt.Printf("Miter computation has failed.\n"); }
    if pGia0n != pGia0 {
        Gia_ManStop( pGia0n ); }
    if pGia1n != pGia1 {
        Gia_ManStop( pGia1n ); }
    Acec_BoxFreeP( &pBox0 );
    Acec_BoxFreeP( &pBox1 );
    return status;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

