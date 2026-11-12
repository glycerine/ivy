package cec

import "fmt"

/**CFile****************************************************************

  FileName    [cecMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Manager procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSatCreate(pAig *Gia_Man_t, pPars *Cec_ParSat_t) *Cec_ManSat_t {
    var p *Cec_ManSat_t;
    // create interpolation manager
    // inlining of macro ABC_ALLOC
    p = ((Cec_ManSat_t *) malloc(sizeof(Cec_ManSat_t) * (size_t)(1)));
    memset( p, 0, sizeof(Cec_ManSat_t) );
    p.pPars        = pPars;
    p.pAig         = pAig;
    // SAT solving
    p.nSatVars     = 1;
    // inlining of macro ABC_CALLOC
    p.pSatVars     = ((int *) calloc((size_t)(Gia_ManObjNum(pAig)), sizeof(int)));
    p.vUsedNodes   = Vec_PtrAlloc( 1000 );
    p.vFanins      = Vec_PtrAlloc( 100 );
    p.vCex         = Vec_IntAlloc( 100 );
    p.vVisits      = Vec_IntAlloc( 100 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Prints statistics of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSatPrintStats(p *Cec_ManSat_t) {
    fmt.Printf("SAT solver statistics:\n");
    Abc_Print( 1, "CO = %8d  ", Gia_ManCoNum(p.pAig) );
    Abc_Print( 1, "AND = %8d  ", Gia_ManAndNum(p.pAig) );
    Abc_Print( 1, "Conf = %5d  ", p.pPars.nBTLimit );
    Abc_Print( 1, "MinVar = %5d  ", p.pPars.nSatVarMax );
    Abc_Print( 1, "MinCalls = %5d\n", p.pPars.nCallsRecycle );
    Abc_Print( 1, "Unsat calls %6d  (%6.2f %%)   Ave conf = %8.1f   ", 
        p.nSatUnsat, p.nSatTotal? 100.0*p.nSatUnsat/p.nSatTotal : 0.0, p.nSatUnsat? 1.0*p.nConfUnsat/p.nSatUnsat :0.0 );
    Abc_PrintTimeP( 1, "Time", p.timeSatUnsat, p.timeTotal );
    Abc_Print( 1, "Sat   calls %6d  (%6.2f %%)   Ave conf = %8.1f   ", 
        p.nSatSat,   p.nSatTotal? 100.0*p.nSatSat/p.nSatTotal : 0.0,   p.nSatSat? 1.0*p.nConfSat/p.nSatSat : 0.0 );
    Abc_PrintTimeP( 1, "Time", p.timeSatSat,   p.timeTotal );
    Abc_Print( 1, "Undef calls %6d  (%6.2f %%)   Ave conf = %8.1f   ", 
        p.nSatUndec, p.nSatTotal? 100.0*p.nSatUndec/p.nSatTotal : 0.0, p.nSatUndec? 1.0*p.nConfUndec/p.nSatUndec : 0.0 );
    Abc_PrintTimeP( 1, "Time", p.timeSatUndec, p.timeTotal );
    Abc_PrintTime( 1, "Total time", p.timeTotal );
}

/**Function*************************************************************

  Synopsis    [Frees the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSatStop(p *Cec_ManSat_t) {
    if p.pSat {
        sat_solver_delete( p.pSat ); }
    Vec_IntFree( p.vCex );
    Vec_IntFree( p.vVisits );
    Vec_PtrFree( p.vUsedNodes );
    Vec_PtrFree( p.vFanins );
    // inlining of macro ABC_FREE
    ((p.pSatVars) ? (free((char *) (p.pSatVars)), (p.pSatVars) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}



/**Function*************************************************************

  Synopsis    [Creates AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatStart() *Cec_ManPat_t { 
    
    // inlining of macro ABC_CALLOC
    p := &Cec_ManPat_t{};
    p.vStorage  = Vec_StrAlloc( 1<<20 );
    p.vPattern1 = Vec_IntAlloc( 1000 );
    p.vPattern2 = Vec_IntAlloc( 1000 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatPrintStats(p *Cec_ManPat_t) { 
    Abc_Print( 1, "Latest: P = %8d.  L = %10d.  Lm = %10d. Ave = %6.1f. MEM =%6.2f MB\n", 
        p.nPats, p.nPatLits, p.nPatLitsMin, 1.0 * p.nPatLitsMin/p.nPats, 
        1.0*(Vec_StrSize(p.vStorage)-p.iStart)/(1<<20) );
    Abc_Print( 1, "Total:  P = %8d.  L = %10d.  Lm = %10d. Ave = %6.1f. MEM =%6.2f MB\n", 
        p.nPatsAll, p.nPatLitsAll, p.nPatLitsMinAll, 1.0 * p.nPatLitsMinAll/p.nPatsAll, 
        1.0*Vec_StrSize(p.vStorage)/(1<<20) );
    Abc_PrintTimeP( 1, "Finding  ", p.timeFind,   p.timeTotal );
    Abc_PrintTimeP( 1, "Shrinking", p.timeShrink, p.timeTotal );
    Abc_PrintTimeP( 1, "Verifying", p.timeVerify, p.timeTotal );
    Abc_PrintTimeP( 1, "Sorting  ", p.timeSort,   p.timeTotal );
    Abc_PrintTimeP( 1, "Packing  ", p.timePack,   p.timeTotal );
    Abc_PrintTime( 1, "TOTAL    ",  p.timeTotal );
}

/**Function*************************************************************

  Synopsis    [Deletes AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManPatStop(p *Cec_ManPat_t) {
    Vec_StrFree( p.vStorage );
    Vec_IntFree( p.vPattern1 );
    Vec_IntFree( p.vPattern2 );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}



/**Function*************************************************************

  Synopsis    [Creates AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimStart(pAig *Gia_Man_t, pPars *Cec_ParSim_t) *Cec_ManSim_t { 
    var p *Cec_ManSim_t;
    // inlining of macro ABC_ALLOC
    p = ((Cec_ManSim_t *) malloc(sizeof(Cec_ManSim_t) * (size_t)(1)));
    memset( p, 0, sizeof(Cec_ManSim_t) );
    p.pAig  = pAig;
    p.pPars = pPars;
    p.nWords = pPars.nWords;
    // inlining of macro ABC_CALLOC
    p.pSimInfo = ((int *) calloc((size_t)(Gia_ManObjNum(pAig)), sizeof(int)));
    p.vClassOld  = Vec_IntAlloc( 1000 );
    p.vClassNew  = Vec_IntAlloc( 1000 );
    p.vClassTemp = Vec_IntAlloc( 1000 );
    p.vRefinedC  = Vec_IntAlloc( 10000 );
    p.vCiSimInfo = Vec_PtrAllocSimInfo( Gia_ManCiNum(p.pAig), pPars.nWords );
    if pPars.fCheckMiter || Gia_ManRegNum(p.pAig) {
        p.vCoSimInfo = Vec_PtrAllocSimInfo( Gia_ManCoNum(p.pAig), pPars.nWords );
        Vec_PtrCleanSimInfo( p.vCoSimInfo, 0, pPars.nWords );
    }
    p.iOut = -1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimStop(p *Cec_ManSim_t) {
    Vec_IntFree( p.vClassOld );
    Vec_IntFree( p.vClassNew );
    Vec_IntFree( p.vClassTemp );
    Vec_IntFree( p.vRefinedC );
    if p.vCiSimInfo { 
        Vec_PtrFree( p.vCiSimInfo ); }
    if p.vCoSimInfo { 
        Vec_PtrFree( p.vCoSimInfo ); }
    // inlining of macro ABC_FREE
    ((p.pScores) ? (free((char *) (p.pScores)), (p.pScores) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pCexComb) ? (free((char *) (p.pCexComb)), (p.pCexComb) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pCexes) ? (free((char *) (p.pCexes)), (p.pCexes) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMems) ? (free((char *) (p.pMems)), (p.pMems) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSimInfo) ? (free((char *) (p.pSimInfo)), (p.pSimInfo) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Creates AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManFraStart(pAig *Gia_Man_t, pPars *Cec_ParFra_t) *Cec_ManFra_t { 
    var p *Cec_ManFra_t;
    // inlining of macro ABC_ALLOC
    p = ((Cec_ManFra_t *) malloc(sizeof(Cec_ManFra_t) * (size_t)(1)));
    memset( p, 0, sizeof(Cec_ManFra_t) );
    p.pAig  = pAig;
    p.pPars = pPars;
    p.vXorNodes  = Vec_IntAlloc( 1000 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManFraStop(p *Cec_ManFra_t) {
    Vec_IntFree( p.vXorNodes );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

