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

