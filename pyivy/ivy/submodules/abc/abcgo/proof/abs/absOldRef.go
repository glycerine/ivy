package abs

import "fmt"

/**CFile****************************************************************

  FileName    [saigAbsStart.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Counter-example-based abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigAbsStart.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"
//#include "proof/ssw/ssw.h"
//#include "proof/fra/fra.h"
//#include "bdd/bbr/bbr.h"
//#include "proof/pdr/pdr.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Derive a new counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Find the first PI corresponding to the flop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Refines abstraction using one step.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCexRefine(p *Aig_Man_t, pAbs *Aig_Man_t, vFlops *Vec_Int_t, nFrames int32, nConfMaxOne int32, fUseBdds int32, fUseDprove int32, fVerbose int32, pnUseStart *int32, piRetValue *int32, pnFrames *int32) *Aig_Man_t { 
    var vFlopsNew *Vec_Int_t;
    var i, Entry, RetValue int32;
    *piRetValue = -1;
    if (fUseDprove != 0) && Aig_ManRegNum(pAbs) > 0 {
/*
        Fra_Sec_t SecPar, * pSecPar = &SecPar;
        Fra_SecSetDefaultParams( pSecPar );
        pSecPar->fVerbose       = fVerbose;
        RetValue = Fra_FraigSec( pAbs, pSecPar, NULL );
*/
        var pAbsOrpos *Aig_Man_t  = Saig_ManDupOrpos( pAbs );
        var Pars Pdr_Par_t;
        var pPars *Pdr_Par_t = &Pars;
        Pdr_ManSetDefaultParams( pPars );
        pPars.nTimeOut = 10;
        pPars.fVerbose = fVerbose;
        if pPars.fVerbose {
            fmt.Printf("Running property directed reachability...\n"); }
        RetValue = Pdr_ManSolve( pAbsOrpos, pPars );
        if pAbsOrpos.pSeqModel {
            pAbsOrpos.pSeqModel.iPo = Saig_ManFindFailedPoCex( pAbs, pAbsOrpos.pSeqModel ); }
        pAbs.pSeqModel = pAbsOrpos.pSeqModel;
        pAbsOrpos.pSeqModel = nil;
        Aig_ManStop( pAbsOrpos );
        if RetValue != 0 {
            *piRetValue = 1; }

    
    } else if (fUseBdds != 0) && (Aig_ManRegNum(pAbs) > 0 && Aig_ManRegNum(pAbs) <= 80) {
        var Pars Saig_ParBbr_t;
        var pPars *Saig_ParBbr_t = &Pars;
        Bbr_ManSetDefaultParams( pPars );
        pPars.TimeLimit     = 0;
        pPars.nBddMax       = 1000000;
        pPars.nIterMax      = nFrames;
        pPars.fPartition    = 1;
        pPars.fReorder      = 1;
        pPars.fReorderImage = 1;
        pPars.fVerbose      = fVerbose;
        pPars.fSilent       = 0;
        RetValue = Aig_ManVerifyUsingBdds( pAbs, pPars );
        if RetValue != 0 {
            *piRetValue = 1; }
    
    } else {
        Saig_BmcPerform( pAbs, pnUseStart? *pnUseStart: 0, nFrames, 2000, 0, nConfMaxOne, 0, fVerbose, 0, pnFrames, 0, 0 );
    }
    if pAbs.pSeqModel == nil {
        return nil; }
    if pnUseStart != nil {
        *pnUseStart = pAbs.pSeqModel.iFrame; }
//    vFlopsNew = Saig_ManExtendCounterExampleTest( pAbs, Saig_ManCexFirstFlopPi(p, pAbs), pAbs->pSeqModel, 1, fVerbose );
    vFlopsNew = Saig_ManExtendCounterExampleTest3( pAbs, Saig_ManCexFirstFlopPi(p, pAbs), pAbs.pSeqModel, fVerbose );
    if vFlopsNew == nil {
        return nil; }
    if Vec_IntSize(vFlopsNew) == 0 {
        fmt.Printf("Discovered a true counter-example!\n");
        p.pSeqModel = Saig_ManCexRemap( p, pAbs, pAbs.pSeqModel );
        Vec_IntFree( vFlopsNew );
        *piRetValue = 0;
        return nil;
    }
    // vFlopsNew contains PI numbers that should be kept in pAbs
    if fVerbose != 0 {
        fmt.Printf("Adding %d registers to the abstraction (total = %d).\n\n", Vec_IntSize(vFlopsNew), Aig_ManRegNum(pAbs)+Vec_IntSize(vFlopsNew)); }
    // add to the abstraction
    Vec_IntForEachEntry( vFlopsNew, Entry, i )
    {
        Entry = Vec_IntEntry(pAbs.vCiNumsOrig, Entry);
        assert(Entry >= Saig_ManPiNum(p), "Entry >= Saig_ManPiNum(p)");
        assert(Entry < Aig_ManCiNum(p), "Entry < Aig_ManCiNum(p)");
        Vec_IntPush( vFlops, Entry-Saig_ManPiNum(p) );
    }
    Vec_IntFree( vFlopsNew );

    Vec_IntSort( vFlops, 0 );
    Vec_IntForEachEntryStart( vFlops, Entry, i, 1 )
        assert(Vec_IntEntry(vFlops, i-1) != Entry, "Vec_IntEntry(vFlops, i-1) != Entry");

    return Saig_ManDupAbstraction( p, vFlops );
}
 
/**Function*************************************************************

  Synopsis    [Refines abstraction using one step.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Transform flop map into flop list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Transform flop list into flop map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Refines abstraction using the latch map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCexAbstractionRefine(pGia *Gia_Man_t, pCex *Abc_Cex_t, nFfToAddMax int32, fTryFour int32, fSensePath int32, fVerbose int32) int32 {
    var pNew *Aig_Man_t;
    var vFlops *Vec_Int_t;
    if pGia.vFlopClasses == nil {
        fmt.Printf("Gia_ManCexAbstractionRefine(): Abstraction latch map is missing.\n");
        return -1;
    }
    pNew = Gia_ManToAig( pGia, 0 );
    vFlops = Gia_ManClasses2Flops( pGia.vFlopClasses );
    if !(Saig_ManCexRefineStep( pNew, vFlops, pGia.vFlopClasses, pCex, nFfToAddMax, fTryFour, fSensePath, fVerbose ) != 0) {
        pGia.pCexSeq = pNew.pSeqModel; pNew.pSeqModel = nil;
        Vec_IntFree( vFlops );
        Aig_ManStop( pNew );
        return 0;
    }
    Vec_IntFree( pGia.vFlopClasses );
    pGia.vFlopClasses = Gia_ManFlops2Classes( pGia, vFlops );
    Vec_IntFree( vFlops );
    Aig_ManStop( pNew );
    return -1;
}


/**Function*************************************************************

  Synopsis    [Computes the flops to remain after abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Saig_ManCexAbstractionFlops(p *Aig_Man_t, pPars *Gia_ParAbs_t) *Vec_Int_t {
    var nUseStart int32;
    var pAbs, pTemp *Aig_Man_t;
    var vFlops *Vec_Int_t;
    var Iter int32;//, clk = Abc_Clock(), clk2 = Abc_Clock();//, iFlop;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    if pPars.fVerbose {
        fmt.Printf("Performing counter-example-based refinement.\n"); }
    Aig_ManSetCioIds( p );
    vFlops = Vec_IntStartNatural( 1 );
/*
    iFlop = Saig_ManFindFirstFlop( p );
    assert( iFlop >= 0 );
    vFlops = Vec_IntAlloc( 1 );
    Vec_IntPush( vFlops, iFlop );
*/
    // create the resulting AIG
    pAbs = Saig_ManDupAbstraction( p, vFlops );
    if !pPars.fVerbose {
        fmt.Printf("Init : ");
        Aig_ManPrintStats( pAbs );
    }
    fmt.Printf("Refining abstraction...\n");
    for  Iter = 0; ; Iter++  {
        pTemp = Saig_ManCexRefine( p, pAbs, vFlops, pPars.nFramesBmc, pPars.nConfMaxBmc, pPars.fUseBdds, pPars.fUseDprove, pPars.fVerbose, pPars.fUseStart?&nUseStart:nil, &pPars.Status, &pPars.nFramesDone );
        if pTemp == nil {
            ABC_FREE( p.pSeqModel );
            p.pSeqModel = pAbs.pSeqModel;
            pAbs.pSeqModel = nil;
            Aig_ManStop( pAbs );
            break;
        }
        Aig_ManStop( pAbs );
        pAbs = pTemp;
        fmt.Printf("ITER %4d : ", Iter);
        if !pPars.fVerbose {
            Aig_ManPrintStats( pAbs ); }
        // output the intermediate result of abstraction
        Ioa_WriteAiger( pAbs, "gabs.aig", 0, 0 );
//            printf( "Intermediate abstracted model was written into file \"%s\".\n", "gabs.aig" );
        // check if the ratio is reached
        if 100.0*(Aig_ManRegNum(p)-Aig_ManRegNum(pAbs))/Aig_ManRegNum(p) < 1.0*pPars.nRatio {
            fmt.Printf("Refinements is stopped because flop reduction is less than %d%%\n", pPars.nRatio);
            Aig_ManStop( pAbs );
            pAbs = nil;
            Vec_IntFree( vFlops );
            vFlops = nil;
            break;
        }
    }
    return vFlops;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

