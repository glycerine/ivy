package mfs

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [mfsMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [Procedures working with the manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mfsMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mfsInt.h"

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
func Mfs_ManAlloc(pPars *Mfs_Par_t) *Mfs_Man_t {
    var p *Mfs_Man_t;
    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Mfs_Man_t *) malloc(sizeof(Mfs_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Mfs_Man_t) );
    p.pPars     = pPars;
    p.vProjVarsCnf = Vec_IntAlloc( 100 );
    p.vProjVarsSat = Vec_IntAlloc( 100 );
    p.vDivLits  = Vec_IntAlloc( 100 );
    p.nDivWords = Abc_BitWordNum(p.pPars.nWinMax + MFS_FANIN_MAX);
    p.vDivCexes = Vec_PtrAllocSimInfo( p.pPars.nWinMax + MFS_FANIN_MAX + 1, p.nDivWords );
    p.pMan      = Int_ManAlloc();
    p.vMem      = Vec_IntAlloc( 0 );
    p.vLevels   = Vec_VecStart( 32 );
    p.vMfsFanins= Vec_PtrAlloc( 32 );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mfs_ManClean(p *Mfs_Man_t) {
    if p.pAigWin {
        Aig_ManStop( p.pAigWin ); }
    if p.pCnf {
        Cnf_DataFree( p.pCnf ); }
    if p.pSat {
        sat_solver_delete( p.pSat ); }
    if p.vRoots {
        Vec_PtrFree( p.vRoots ); }
    if p.vSupp {
        Vec_PtrFree( p.vSupp ); }
    if p.vNodes {
        Vec_PtrFree( p.vNodes ); }
    if p.vDivs {
        Vec_PtrFree( p.vDivs ); }
    p.pAigWin = nil;
    p.pCnf    = nil;
    p.pSat    = nil;
    p.vRoots  = nil;
    p.vSupp   = nil;
    p.vNodes  = nil;
    p.vDivs   = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mfs_ManPrint(p *Mfs_Man_t) {
    if p.pPars.fResub {
        fmt.Printf("Nodes = %d. Try = %d. Resub = %d. Div = %d. SAT calls = %d. Timeouts = %d. MaxDivs = %d.\n", p.nTotalNodesBeg, p.nNodesTried, p.nNodesResub, p.nTotalDivs, p.nSatCalls, p.nTimeOuts, p.nMaxDivs);

        fmt.Printf("Attempts :   ");
        fmt.Printf("Remove %6d out of %6d (%6.2f %%)   ", p.nRemoves, p.nTryRemoves, 100.0*p.nRemoves/Abc_MaxInt(1, p.nTryRemoves));
        fmt.Printf("Resub  %6d out of %6d (%6.2f %%)   ", p.nResubs, p.nTryResubs, 100.0*p.nResubs /Abc_MaxInt(1, p.nTryResubs));
        fmt.Printf("\n");

        fmt.Printf("Reduction:   ");
        fmt.Printf("Nodes  %6d out of %6d (%6.2f %%)   ", p.nTotalNodesBeg-p.nTotalNodesEnd, p.nTotalNodesBeg, 100.0*(p.nTotalNodesBeg-p.nTotalNodesEnd)/Abc_MaxInt(1, p.nTotalNodesBeg));
        fmt.Printf("Edges  %6d out of %6d (%6.2f %%)   ", p.nTotalEdgesBeg-p.nTotalEdgesEnd, p.nTotalEdgesBeg, 100.0*(p.nTotalEdgesBeg-p.nTotalEdgesEnd)/Abc_MaxInt(1, p.nTotalEdgesBeg));
        fmt.Printf("\n");

        if p.pPars.fPower {
            fmt.Printf("Power( %5.2f, %4.2f%%) \n", p.TotalSwitchingBeg - p.TotalSwitchingEnd, 100.0*(p.TotalSwitchingBeg-p.TotalSwitchingEnd)/p.TotalSwitchingBeg); }
        if p.pPars.fSwapEdge {
            fmt.Printf("Swappable edges = %d. Total edges = %d. Ratio = %5.2f.\n", p.nNodesResub, Abc_NtkGetTotalFanins(p.pNtk), 1.00 * p.nNodesResub / Abc_NtkGetTotalFanins(p.pNtk)); }
//        printf( "Average ratio of DCs in the resubed nodes = %.2f.\n", 1.0*p->nDcMints/(64 * p->nNodesResub) );
    
    } else {
        fmt.Printf("Nodes = %d. Try = %d. Total mints = %d. Local DC mints = %d. Ratio = %5.2f.\n", p.nTotalNodesBeg, p.nNodesTried, p.nMintsTotal, p.nMintsTotal-p.nMintsCare, 1.0 * (p.nMintsTotal-p.nMintsCare) / p.nMintsTotal);
//        printf( "Average ratio of sequential DCs in the global space = %5.2f.\n", 
//            1.0-(p->dTotalRatios/p->nNodesTried) );
        fmt.Printf("Nodes resyn = %d. Ratio = %5.2f.  Total AIG node gain = %d. Timeouts = %d.\n", p.nNodesDec, 1.0 * p.nNodesDec / p.nNodesTried, p.nNodesGained, p.nTimeOuts);
    }

    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Win"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeWin)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeWin)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Div"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeDiv)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeDiv)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Aig"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeAig)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeAig)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Gia"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeGia)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeGia)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Cnf"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCnf)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeCnf)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Sat"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat-p.timeInt)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat-p.timeInt)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Int"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeInt)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeInt)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("ALL"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;

}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mfs_ManStop(p *Mfs_Man_t) {
    if p.pPars.fVerbose {
        Mfs_ManPrint( p ); }
    if p.vTruth {
        Vec_IntFree( p.vTruth ); }
    if p.pManDec {
        Bdc_ManFree( p.pManDec ); }
    if p.pCare {
        Aig_ManStop( p.pCare ); }
    if p.vSuppsInv {
        Vec_VecFree( (Vec_Vec_t *)p.vSuppsInv ); }
    if p.vProbs {
        Vec_IntFree( p.vProbs ); }
    Mfs_ManClean( p );
    Int_ManFree( p.pMan );
    Vec_IntFree( p.vMem );
    Vec_VecFree( p.vLevels );
    Vec_PtrFree( p.vMfsFanins );
    Vec_IntFree( p.vProjVarsCnf );
    Vec_IntFree( p.vProjVarsSat );
    Vec_IntFree( p.vDivLits );
    Vec_PtrFree( p.vDivCexes );
    p = nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

