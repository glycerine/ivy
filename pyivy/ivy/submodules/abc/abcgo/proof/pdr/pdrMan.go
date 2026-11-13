package pdr

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [pdrMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Manager procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrMan.c,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pdrInt.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Structural analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManDeriveCexAbs(p *Pdr_Man_t) *Abc_Cex_t {
    extern Gia_Man_t * Gia_ManDupAbs( Gia_Man_t * p, Vec_Int_t * vMapPpi2Ff, Vec_Int_t * vMapFf2Ppi );

    var pAbs *Gia_Man_t;
    var pCex, pCexCare *Abc_Cex_t;
    var pObl *Pdr_Obl_t;
    var i, f, Lit, Flop, nFrames int32;
    var nPis int32  = Saig_ManPiNum(p.pAig);
    var nFfRefined int32;
    if !p.pPars.fUseAbs || !p.vMapPpi2Ff {
        return Pdr_ManDeriveCex(p); }
    // restore previous map
    Vec_IntForEachEntry( p.vMapPpi2Ff, Flop, i )
    {
        assert(Vec_IntEntry( p.vMapFf2Ppi, Flop ) == i, "Vec_IntEntry( p.vMapFf2Ppi, Flop ) == i");
        Vec_IntWriteEntry( p.vMapFf2Ppi, Flop, -1 );
    }
    Vec_IntClear( p.vMapPpi2Ff );
    // count the number of frames
    for  pObl = p.pQueue; (pObl != nil); pObl = pObl.pNext  {
        for  i = pObl.pState.nLits; i < pObl.pState.nTotal; i++  {
            Lit = pObl.pState.Lits[i]; 
            if Abc_Lit2Var(Lit) < nPis { // PI literal
                continue; }
            Flop = Abc_Lit2Var(Lit) - nPis;
            if Vec_IntEntry(p.vMapFf2Ppi, Flop) >= 0 { // already used PPI literal
                continue; }
            Vec_IntWriteEntry( p.vMapFf2Ppi, Flop, Vec_IntSize(p.vMapPpi2Ff) );
            Vec_IntPush( p.vMapPpi2Ff, Flop );
        }
        nFrames++;
    }
    if Vec_IntSize(p.vMapPpi2Ff) == 0 { // no PPIs -- this is a real CEX
        return Pdr_ManDeriveCex(p); }
    if p.pPars.fUseSimpleRef {
        // rely on ternary simulation to perform refinement
        Vec_IntForEachEntry( p.vMapPpi2Ff, Flop, i )
        {
            assert(Vec_IntEntry(p.vAbsFlops, Flop) == 0, "Vec_IntEntry(p.vAbsFlops, Flop) == 0");
            Vec_IntWriteEntry( p.vAbsFlops, Flop, 1 );
            nFfRefined++;
        }
    
    } else {
        // create the counter-example
        pCex = Abc_CexAlloc( Aig_ManRegNum(p.pAig) - Vec_IntSize(p.vMapPpi2Ff), Saig_ManPiNum(p.pAig) + Vec_IntSize(p.vMapPpi2Ff), nFrames );
        pCex.iPo    = p.iOutCur;
        pCex.iFrame = nFrames-1;
        for  pObl, f = p.pQueue, 0; (pObl != nil); pObl, f = pObl.pNext, f+1  {
            for  i = pObl.pState.nLits; i < pObl.pState.nTotal; i++  {
                Lit = pObl.pState.Lits[i];
                if Abc_LitIsCompl(Lit) {
                    continue; }
                if Abc_Lit2Var(Lit) < nPis { // PI literal
                    Abc_InfoSetBit( pCex.pData, pCex.nRegs + f * pCex.nPis + Abc_Lit2Var(Lit) );
                } else {
                    var iPPI int32  = nPis + Vec_IntEntry(p.vMapFf2Ppi, Abc_Lit2Var(Lit) - nPis);
                    assert(iPPI < pCex.nPis, "iPPI < pCex.nPis");
                    Abc_InfoSetBit( pCex.pData, pCex.nRegs + f * pCex.nPis + iPPI );
                }
            }
        }
        assert(f == nFrames, "f == nFrames");
        // perform CEX minimization
        pAbs = Gia_ManDupAbs( p.pGia, p.vMapPpi2Ff, p.vMapFf2Ppi );
        pCexCare = Bmc_CexCareMinimizeAig( pAbs, nPis, pCex, 1, 0, 0 );
        Gia_ManStop( pAbs );
        assert(pCexCare.nPis == pCex.nPis, "pCexCare.nPis == pCex.nPis");
        Abc_CexFree( pCex );
        // detect care PPIs
        for  f = 0; f < nFrames; f++  {
            for  i = nPis; i < pCexCare.nPis; i++  {
                if Abc_InfoHasBit(pCexCare.pData, pCexCare.nRegs + pCexCare.nPis * f + i) {
                    if Vec_IntEntry(p.vAbsFlops, Vec_IntEntry(p.vMapPpi2Ff, i-nPis)) == 0 { // currently abstracted
                        Vec_IntWriteEntry( p.vAbsFlops, Vec_IntEntry(p.vMapPpi2Ff, i-nPis), 1 );
nFfRefined++; }
                }
            }
        }
        Abc_CexFree( pCexCare );
        if nFfRefined == 0 { // no refinement -- this is a real CEX
            return Pdr_ManDeriveCex(p); }
    }
    //printf( "CEX-based refinement refined %d flops.\n", nFfRefined );
    p.nCexesTotal++;
    p.nCexes++;
    return nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

