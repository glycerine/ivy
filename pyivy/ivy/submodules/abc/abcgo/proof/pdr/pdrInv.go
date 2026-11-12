package pdr

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [pdrInv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Invariant computation, printing, verification.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrInv.c,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pdrInt.h"
//#include "base/abc/abc.h"      // for Abc_NtkCollectCioNames()
//#include "base/main/main.h"    // for Abc_FrameReadGlobalFrame()
//#include "aig/ioa/ioa.h"

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
func Pdr_ManPrintProgress(p *Pdr_Man_t, fClose int32, Time abctime) {
    var vVec *Vec_Ptr_t;
    var i, ThisSize, Length, LengthStart int32;
    if Vec_PtrSize(p.vSolvers) < 2 {
        Abc_Print(1, "Frame " );
        Abc_Print(1, "Clauses                                                     " );
        Abc_Print(1, "Max Queue " );
        Abc_Print(1, "Flops " );
        Abc_Print(1, "Cex      " );
        Abc_Print(1, "Time" );
        Abc_Print(1, "\n" );
        return;
    }
    if Abc_FrameIsBatchMode() && !(fClose != 0) {
        return; }
    // count the total length of the printout
    Length = 0;
    Vec_VecForEachLevel( p.vClauses, vVec, i )
        Length += 1 + Abc_Base10Log(Vec_PtrSize(vVec)+1);
    // determine the starting point
    LengthStart = Abc_MaxInt( 0, Length - 60 );
    Abc_Print( 1, "%3d :", Vec_PtrSize(p.vSolvers)-1 );
    ThisSize = 5;
    if LengthStart > 0 {
        Abc_Print( 1, " ..." );
        ThisSize += 4;
    }
    Length = 0;
    Vec_VecForEachLevel( p.vClauses, vVec, i )
    {
        if Length < LengthStart {
            Length += 1 + Abc_Base10Log(Vec_PtrSize(vVec)+1);
            continue;
        }
        Abc_Print( 1, " %d", Vec_PtrSize(vVec) );
        Length += 1 + Abc_Base10Log(Vec_PtrSize(vVec)+1);
        ThisSize += 1 + Abc_Base10Log(Vec_PtrSize(vVec)+1);
    }
    for  i = ThisSize; i < 70; i++  {
        Abc_Print( 1, " " );
    }
    Abc_Print( 1, "%5d", p.nQueMax );
    Abc_Print( 1, "%6d", p.vAbsFlops ? Vec_IntCountPositive(p.vAbsFlops) : p.nAbsFlops );
    if p.pPars.fUseAbs {
    Abc_Print( 1, "%4d", p.nCexes ); }
    Abc_Print( 1, "%10.2f sec", 1.0*Time/int64(time.Second) );
    if p.pPars.fSolveAll {
        Abc_Print( 1, "  CEX =%4d", p.pPars.nFailOuts ); }
    if p.pPars.nTimeOutOne {
        Abc_Print( 1, "  T/O =%3d", p.pPars.nDropOuts ); }
    Abc_Print( 1, "%s", fClose ? "\n":"\r" );
    if fClose != 0 {
        p.nQueMax = 0, p.nCexes = 0; }
    fflush( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Counts how many times each flop appears in the set of cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManCountFlops(p *Pdr_Man_t, vCubes *Vec_Ptr_t) *Vec_Int_t {
    var vFlopCount *Vec_Int_t;
    var pCube *Pdr_Set_t;
    var i, n int32;
    vFlopCount = Vec_IntStart( Aig_ManRegNum(p.pAig) );
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        if pCube.nRefs == -1 {
            continue; }
        for  n = 0; n < pCube.nLits; n++  {
            assert(pCube.Lits[n] >= 0 && pCube.Lits[n] < 2*Aig_ManRegNum(p.pAig), "pCube.Lits[n] >= 0 && pCube.Lits[n] < 2*Aig_ManRegNum(p.pAig)");
            Vec_IntAddToEntry( vFlopCount, pCube.Lits[n] >> 1, 1 );
        }
    }
    return vFlopCount;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManFindInvariantStart(p *Pdr_Man_t) int32 {
    var vArrayK *Vec_Ptr_t;
    var k int32;
    var kMax int32 = Vec_PtrSize(p.vSolvers)-1;
    Vec_VecForEachLevelStartStop( p.vClauses, vArrayK, k, 1, kMax+1 )
        if ( Vec_PtrSize(vArrayK) == 0 )
            return k;
//    return -1;
    // if there is no starting point (as in case of SAT or undecided), return the last frame
//    Abc_Print( 1, "The last timeframe contains %d clauses.\n", Vec_PtrSize(Vec_VecEntry(p->vClauses, kMax)) );
    return kMax;
}

/**Function*************************************************************

  Synopsis    [Counts the number of variables used in the clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManCollectCubes(p *Pdr_Man_t, kStart int32) *Vec_Ptr_t {
    var vResult *Vec_Ptr_t;
    var vArrayK *Vec_Ptr_t;
    var pSet *Pdr_Set_t;
    var i, j int32;
    vResult = Vec_PtrAlloc( 100 );
    Vec_VecForEachLevelStart( p.vClauses, vArrayK, i, kStart )
        Vec_PtrForEachEntry( Pdr_Set_t *, vArrayK, pSet, j )
            Vec_PtrPush( vResult, pSet );
    return vResult;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManCountFlopsInv(p *Pdr_Man_t) *Vec_Int_t {
    var kStart int32  = Pdr_ManFindInvariantStart(p);
    var vCubes *Vec_Ptr_t  = Pdr_ManCollectCubes(p, kStart);
    var vInv *Vec_Int_t  = Pdr_ManCountFlops( p, vCubes );
    Vec_PtrFree(vCubes);
    return vInv;
}

/**Function*************************************************************

  Synopsis    [Counts the number of variables used in the clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManCountVariables(p *Pdr_Man_t, kStart int32) int32 {
    var vFlopCounts *Vec_Int_t;
    var vCubes *Vec_Ptr_t;
    var i, Entry, Counter int32;
    if p.vInfCubes == nil {
        vCubes = Pdr_ManCollectCubes( p, kStart );
    } else {
        vCubes = Vec_PtrDup( p.vInfCubes ); }
    vFlopCounts = Pdr_ManCountFlops( p, vCubes );
    Vec_IntForEachEntry( vFlopCounts, Entry, i )
        Counter += (Entry > 0);
    Vec_IntFreeP( &vFlopCounts );
    Vec_PtrFree( vCubes );
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManPrintClauses(p *Pdr_Man_t, kStart int32) {
    var vArrayK *Vec_Ptr_t;
    var pCube *Pdr_Set_t;
    var i, k, Counter int32;
    Vec_VecForEachLevelStart( p.vClauses, vArrayK, k, kStart )
    {
        Vec_PtrSort( vArrayK, (int (*)(void))Pdr_SetCompare );
        Vec_PtrForEachEntry( Pdr_Set_t *, vArrayK, pCube, i )
        {
            Abc_Print( 1, "C=%4d. F=%4d ", Counter++, k );
            Pdr_SetPrint( os.Stdout, pCube, Aig_ManRegNum(p.pAig), nil );  
            Abc_Print( 1, "\n" ); 
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Pdr_SetPrintOne(p *Pdr_Set_t) {
    var i int32;
    Abc_Print(1, "Clause: {" );
    for  i = 0; i < p.nLits; i++  {
        Abc_Print(1, " %s%d", Abc_LitIsCompl(p.Lits[i])? "!":"", Abc_Lit2Var(p.Lits[i]) );
    }
    Abc_Print(1, " }\n" );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Pdr_ManDupAigWithClauses(p *Aig_Man_t, vCubes *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjNew, pLit *Aig_Obj_t;
    var pCube *Pdr_Set_t;
    var i, n int32;
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    Aig_ManForEachCi( p, pObj, i )
        pObj.pData = Aig_ObjCreateCi( pNew );
    // create outputs for each cube
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
//        Pdr_SetPrintOne( pCube );
        pObjNew = Aig_ManConst1(pNew);
        for  n = 0; n < pCube.nLits; n++  {
            assert(Abc_Lit2Var(pCube.Lits[n]) < Saig_ManRegNum(p), "Abc_Lit2Var(pCube.Lits[n]) < Saig_ManRegNum(p)");
            pLit = Aig_NotCond( Aig_ManCi(pNew, Saig_ManPiNum(p) + Abc_Lit2Var(pCube.Lits[n])), Abc_LitIsCompl(pCube.Lits[n]) );
            pObjNew = Aig_And( pNew, pObjNew, pLit );
        }
        Aig_ObjCreateCo( pNew, pObjNew );
    }
    // duplicate internal nodes
    Aig_ManForEachNode( p, pObj, i )
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    // add the POs
    Saig_ManForEachLi( p, pObj, i )
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    Aig_ManCleanup( pNew );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting network
    if !Aig_ManCheck(pNew) {
        Abc_Print(1, "Aig_ManDupSimple(): The check has failed.\n" ); }
    return pNew;
}
func Pdr_ManDumpAig(p *Aig_Man_t, vCubes *Vec_Ptr_t) {
    var pNew *Aig_Man_t  = Pdr_ManDupAigWithClauses( p, vCubes );
    Ioa_WriteAiger( pNew, "aig_with_clauses.aig", 0, 0 );
    Aig_ManStop( pNew );
    Abc_Print(1, "Dumped modified AIG into file \"aig_with_clauses.aig\".\n" );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Pdr_ManDumpClauses(p *Pdr_Man_t, pFileName string, fProved int32) {
    var pFile *os.File;
    var vFlopCounts *Vec_Int_t;
    var vCubes *Vec_Ptr_t;
    var pCube *Pdr_Set_t;
    var pNamesCi *string;
    var i, kStart, Count int32;
    // create file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        Abc_Print( 1, "Cannot open file \"%s\" for writing invariant.\n", pFileName );
        return;
    } 
    // collect cubes
    kStart = Pdr_ManFindInvariantStart( p );
    if fProved != 0 {
        vCubes = Pdr_ManCollectCubes( p, kStart );
    } else {
        vCubes = Vec_PtrDup( p.vInfCubes ); }
    Vec_PtrSort( vCubes, (int (*)(void))Pdr_SetCompare );
//    Pdr_ManDumpAig( p->pAig, vCubes );
    // count cubes
    Count = 0;
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        if pCube.nRefs == -1 {
            continue; }
        Count++;
    }
    // collect variable appearances
    if p.pPars.fUseSupp {
        vFlopCounts = Pdr_ManCountFlops( p, vCubes );
    } else {
        vFlopCounts = nil;
    } 
    // output the header
    if fProved != 0 {
        fmt.Fprintf(pFile, "# Inductive invariant for \"%s\"\n", p.pAig.pName);
    } else {
        fmt.Fprintf(pFile, "# Clauses of the last timeframe for \"%s\"\n", p.pAig.pName); }
    fmt.Fprintf(pFile, "# generated by PDR in ABC on %s\n", Aig_TimeStamp());
    fmt.Fprintf(pFile, ".i %d\n", p.pPars.fUseSupp ? Pdr_ManCountVariables(p, kStart) : Aig_ManRegNum(p.pAig));
    fmt.Fprintf(pFile, ".o 1\n");
    fmt.Fprintf(pFile, ".p %d\n", Count);
    // output flop names
    pNamesCi = Abc_NtkCollectCioNames( Abc_FrameReadNtk( Abc_FrameReadGlobalFrame() ), 0 );
    if pNamesCi != nil {
        fmt.Fprintf(pFile, ".ilb");
        for  i = 0; i < Aig_ManRegNum(p.pAig); i++  {
            if !p.pPars.fUseSupp || Vec_IntEntry( vFlopCounts, i ) {
                fmt.Fprintf(pFile, " %s", pNamesCi[Saig_ManPiNum(p.pAig) + i]); }
        }
        fmt.Fprintf(pFile, "\n");
        ABC_FREE( pNamesCi );
        fmt.Fprintf(pFile, ".ob inv\n");
    }
    // output cubes
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        if pCube.nRefs == -1 {
            continue; }
        Pdr_SetPrint( pFile, pCube, Aig_ManRegNum(p.pAig), vFlopCounts );  
        fmt.Fprintf(pFile, " 1\n"); 
    }
    fmt.Fprintf(pFile, ".e\n\n");
    _ = (pFile).Close();
    Vec_IntFreeP( &vFlopCounts );
    Vec_PtrFree( vCubes );
    if fProved != 0 {
        Abc_Print( 1, "Inductive invariant was written into file \"%s\".\n", pFileName );
    } else {
        Abc_Print( 1, "Clauses of the last timeframe were written into file \"%s\".\n", pFileName ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Pdr_ManDumpString(p *Pdr_Man_t) *Vec_Str_t {
    var vStr *Vec_Str_t;
    var vFlopCounts *Vec_Int_t;
    var vCubes *Vec_Ptr_t;
    var pCube *Pdr_Set_t;
    var i, kStart int32;
    vStr = Vec_StrAlloc( 1000 );
    // collect cubes
    kStart = Pdr_ManFindInvariantStart( p );
    if p.vInfCubes == nil {
        vCubes = Pdr_ManCollectCubes( p, kStart );
    } else {
        vCubes = Vec_PtrDup( p.vInfCubes ); }
    Vec_PtrSort( vCubes, (int (*)(void))Pdr_SetCompare );
    // collect variable appearances
    if p.pPars.fUseSupp {
        vFlopCounts = Pdr_ManCountFlops( p, vCubes );
    } else {
        vFlopCounts = nil;
    } 
    // output cubes
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        if pCube.nRefs == -1 {
            continue; }
        Pdr_SetPrintStr( vStr, pCube, Aig_ManRegNum(p.pAig), vFlopCounts );  
    }
    Vec_IntFreeP( &vFlopCounts );
    Vec_PtrFree( vCubes );
    Vec_StrPush( vStr, '\0' );
    return vStr;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManReportInvariant(p *Pdr_Man_t) {
    var vCubes *Vec_Ptr_t;
    var kStart int32  = Pdr_ManFindInvariantStart( p );
    vCubes = Pdr_ManCollectCubes( p, kStart );
    Abc_Print( 1, "Invariant F[%d] : %d clauses with %d flops (out of %d) (cex = %d, ave = %.2f)\n", 
        kStart, Vec_PtrSize(vCubes), Pdr_ManCountVariables(p, kStart), Aig_ManRegNum(p.pAig), p.nCexesTotal, 1.0*p.nXsimLits/p.nXsimRuns );
//    Abc_Print( 1, "Invariant F[%d] : %d clauses with %d flops (out of %d)\n", 
//        kStart, Vec_PtrSize(vCubes), Pdr_ManCountVariables(p, kStart), Aig_ManRegNum(p->pAig) );
    Vec_PtrFree( vCubes );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManVerifyInvariant(p *Pdr_Man_t) {
    var pSat *sat_solver;
    var vLits *Vec_Int_t;
    var vCubes *Vec_Ptr_t;
    var pCube *Pdr_Set_t;
    var i, kStart, kThis, RetValue, Counter int32;
    var clk abctime  = Abc_Clock();
    // collect cubes used in the inductive invariant
    kStart = Pdr_ManFindInvariantStart( p );
    vCubes = Pdr_ManCollectCubes( p, kStart );
    // create solver with the cubes
    kThis = Vec_PtrSize(p.vSolvers);
    pSat  = Pdr_ManCreateSolver( p, kThis );
    // add the property output
//    Pdr_ManSetPropertyOutput( p, kThis );
    // add the clauses
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        vLits = Pdr_ManCubeToLits( p, kThis, pCube, 1, 0 );
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );
        assert((RetValue != 0), "RetValue");
        sat_solver_compress( pSat );
    }
    // check each clause
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        vLits = Pdr_ManCubeToLits( p, kThis, pCube, 0, 1 );
        RetValue = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), 0, 0, 0, 0 );
        if RetValue != l_False {
            Abc_Print( 1, "Verification of clause %d failed.\n", i );
            Counter++;
        }
    }
    if Counter != 0 {
        Abc_Print( 1, "Verification of %d clauses has failed.\n", Counter );
    } else {
        Abc_Print( 1, "Verification of invariant with %d clauses was successful.  ", Vec_PtrSize(vCubes) );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
//    sat_solver_delete( pSat );
    Vec_PtrFree( vCubes );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManDeriveMarkNonInductive(p *Pdr_Man_t, vCubes *Vec_Ptr_t) int32 {
    var pSat *sat_solver;
    var vLits *Vec_Int_t;
    var pCube *Pdr_Set_t;
    var i, kThis, RetValue, fChanges, Counter int32;
    // create solver with the cubes
    kThis = Vec_PtrSize(p.vSolvers);
    pSat  = Pdr_ManCreateSolver( p, kThis );
    // add the clauses
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        if pCube.nRefs == -1 { // skip non-inductive
            continue; }
        vLits = Pdr_ManCubeToLits( p, kThis, pCube, 1, 0 );
        RetValue = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );
        assert((RetValue != 0), "RetValue");
        sat_solver_compress( pSat );
    }
    // check each clause
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        if pCube.nRefs == -1 { // skip non-inductive
            continue; }
        vLits = Pdr_ManCubeToLits( p, kThis, pCube, 0, 1 );
        RetValue = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), 0, 0, 0, 0 );
        if RetValue != l_False { // mark as non-inductive
        
            pCube.nRefs = -1;
            fChanges = 1;
        
        } else {
            Counter++; }
    }
    //Abc_Print(1, "Clauses = %d.\n", Counter );
    //sat_solver_delete( pSat );
    return fChanges;
}
func Pdr_ManDeriveInfinityClauses(p *Pdr_Man_t, fReduce int32) *Vec_Int_t {
    var vResult *Vec_Int_t;
    var vCubes *Vec_Ptr_t;
    var pCube *Pdr_Set_t;
    var i, v, kStart int32;
    // collect cubes used in the inductive invariant
    kStart = Pdr_ManFindInvariantStart( p );
    vCubes = Pdr_ManCollectCubes( p, kStart );
    // refine as long as there are changes
    if fReduce != 0 {
        for  Pdr_ManDeriveMarkNonInductive(p, vCubes) != 0  {; } }
    // collect remaining clauses
    vResult = Vec_IntAlloc( 1000 );
    Vec_IntPush( vResult, 0 );
    Vec_PtrForEachEntry( Pdr_Set_t *, vCubes, pCube, i )
    {
        if pCube.nRefs == -1 { // skip non-inductive
            continue; }
        Vec_IntAddToEntry( vResult, 0, 1 );
        Vec_IntPush( vResult, pCube.nLits );
        for  v = 0; v < pCube.nLits; v++  {
            Vec_IntPush( vResult, pCube.Lits[v] );
        }
    }
    //Vec_PtrFree( vCubes );
    Vec_PtrFreeP( &p.vInfCubes );
    p.vInfCubes = vCubes;
    Vec_IntPush( vResult, Aig_ManRegNum(p.pAig) );
    return vResult;
}



/**Function*************************************************************

  Synopsis    [Remove clauses while maintaining the invariant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// Original macro:
// #define Pdr_ForEachCube( pList, pCut, i ) for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += pCut[0] + 1 )


func Pdr_InvMap(vCounts *Vec_Int_t) *Vec_Int_t {
    var i, k, Count int32;
    var vMap *Vec_Int_t  = Vec_IntStart( Vec_IntSize(vCounts) );
    Vec_IntForEachEntry( vCounts, Count, i )
        if ( Count )
            Vec_IntWriteEntry( vMap, i, k++ );
    return vMap;
}
func Pdr_InvCounts(vInv *Vec_Int_t) *Vec_Int_t {
    var i int32;
    var k int32;
    var pCube *int32;
    var pList *int32 = Vec_IntArray(vInv);
    var vCounts *Vec_Int_t  = Vec_IntStart( Vec_IntEntryLast(vInv) );
    // inlining of macro Pdr_ForEachCube
    for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
        for  k = 0; k < pCube[0]; k++  {
            Vec_IntAddToEntry( vCounts, Abc_Lit2Var(pCube[k+1]), 1 );
        }
    }
    return vCounts;
}
func Pdr_InvUsedFlopNum(vInv *Vec_Int_t) int32 {
    var vCounts *Vec_Int_t  = Pdr_InvCounts( vInv );
    var nZeros int32  = Vec_IntCountZero( vCounts );
    Vec_IntFree( vCounts );
    return Vec_IntEntryLast(vInv) - nZeros;
}

func Pdr_InvPrintStr(vInv *Vec_Int_t, vCounts *Vec_Int_t) *Vec_Str_t {
    var vStr *Vec_Str_t  = Vec_StrAlloc( 1000 );
    var vMap *Vec_Int_t  = Pdr_InvMap( vCounts );
    var nVars int32  = Vec_IntSize(vCounts) - Vec_IntCountZero(vCounts);
    var i int32;
    var k int32;
    var pCube *int32;
    var pList *int32 = Vec_IntArray(vInv);
    var pBuffer string  = ABC_ALLOC( char, (size_t)uint32(nVars) );
    for  i = 0; i < nVars; i++  {
        pBuffer[i] = '-';
    }
    // inlining of macro Pdr_ForEachCube
    for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
        for  k = 0; k < pCube[0]; k++  {
            pBuffer[Vec_IntEntry(vMap, Abc_Lit2Var(pCube[k+1]))] = '0' + !Abc_LitIsCompl(pCube[k+1]);
        }
        for  k = 0; k < nVars; k++  {
            Vec_StrPush( vStr, pBuffer[k] );
        }
        Vec_StrPush( vStr, ' ' );
        Vec_StrPush( vStr, '1' );
        Vec_StrPush( vStr, '\n' );
        for  k = 0; k < pCube[0]; k++  {
            pBuffer[Vec_IntEntry(vMap, Abc_Lit2Var(pCube[k+1]))] = '-';
        }
    }
    Vec_StrPush( vStr, '\0' );
    ABC_FREE( pBuffer );
    Vec_IntFree( vMap );
    return vStr;
}
func Pdr_InvPrint(vInv *Vec_Int_t, fVerbose int32) {
    Abc_Print(1, "Invariant contains %d clauses with %d literals and %d flops (out of %d).\n", Vec_IntEntry(vInv, 0), Vec_IntSize(vInv)-Vec_IntEntry(vInv, 0)-2, Pdr_InvUsedFlopNum(vInv), Vec_IntEntryLast(vInv) );
    if fVerbose != 0 {
        var vCounts *Vec_Int_t  = Pdr_InvCounts( vInv );
        var vStr *Vec_Str_t  = Pdr_InvPrintStr( vInv, vCounts );
        Abc_Print(1, "%s", Vec_StrArray( vStr ) );
        Vec_IntFree( vCounts );
        Vec_StrFree( vStr );
    }
}

func Pdr_InvCheck_int(p *Gia_Man_t, vInv *Vec_Int_t, fVerbose int32, pSat *sat_solver, fSkip int32) int32 {
    var nBTLimit int32;
    var fCheckProperty int32  = 1;
    var i, k, status, nFailed, nFailedOuts int32; 
    // collect cubes
    var pCube *int32;
    var pList *int32 = Vec_IntArray(vInv);
    // create variables
    var vLits *Vec_Int_t  = Vec_IntAlloc(100);
    var iFoVarBeg int32  = sat_solver_nvars(pSat) - Gia_ManRegNum(p);
    var iFiVarBeg int32  = 1 + Gia_ManPoNum(p);
    // add cubes
    // inlining of macro Pdr_ForEachCube
    for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
        // collect literals
        Vec_IntClear( vLits );
        for  k = 0; k < pCube[0]; k++  {
            if pCube[k+1] != -1 {
                Vec_IntPush( vLits, Abc_Var2Lit(iFoVarBeg + Abc_Lit2Var(pCube[k+1]), !Abc_LitIsCompl(pCube[k+1])) ); }
        }
        if Vec_IntSize(vLits) == 0 {
            Vec_IntFree( vLits );
            return 1;
        }
        // add it to the solver
        status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits) );
        assert(status == 1, "status == 1");
    }
    // verify property output
    if fCheckProperty != 0 {
        for  i = 0; i < Gia_ManPoNum(p); i++  {
            Vec_IntFill( vLits, 1, Abc_Var2Lit(1+i, 0) );
            status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits), nBTLimit, 0, 0, 0 );
            if status == l_Undef { // timeout
                break; }
            if status == l_True { // sat - property fails
            
                if fVerbose != 0 {
                    Abc_Print(1, "Coverage check failed for output %d.\n", i ); }
                nFailedOuts++;
                if fSkip != 0 {
                    Vec_IntFree( vLits );
                    return 1;
                }
                continue;
            }
            assert(status == l_False, "status == l_False"); // unsat - property holds
        }
    }
    // iterate through cubes in the direct order
    // inlining of macro Pdr_ForEachCube
    for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
        // collect cube
        Vec_IntClear( vLits );
        for  k = 0; k < pCube[0]; k++  {
            if pCube[k+1] != -1 {
                Vec_IntPush( vLits, Abc_Var2Lit(iFiVarBeg + Abc_Lit2Var(pCube[k+1]), Abc_LitIsCompl(pCube[k+1])) ); }
        }
        // check if this cube intersects with the complement of other cubes in the solver
        // if it does not intersect, then it is redundant and can be skipped
        status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits), nBTLimit, 0, 0, 0 );
        if status != l_True && (fVerbose != 0) {
            Abc_Print(1, "Finished checking clause %d (out of %d)...\r", i, pList[0] ); }
        if status == l_Undef { // timeout
            break; }
        if status == l_False { // unsat -- correct
            continue; }
        assert(status == l_True, "status == l_True");
        if fVerbose != 0 {
            Abc_Print(1, "Inductiveness check failed for clause %d.\n", i ); }
        nFailed++;
        if fSkip != 0 {
            Vec_IntFree( vLits );
            return 1;
        }
    }
    Vec_IntFree( vLits );
    return nFailed + nFailedOuts;
}

func Pdr_InvCheck(p *Gia_Man_t, vInv *Vec_Int_t, fVerbose int32) int32 {
    var RetValue int32;
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    assert(sat_solver_nvars(pSat) == pCnf.nVars, "sat_solver_nvars(pSat) == pCnf.nVars");
    Cnf_DataFree( pCnf );
    RetValue = Pdr_InvCheck_int( p, vInv, fVerbose, pSat, 0 );
    sat_solver_delete( pSat );
    return RetValue;
}

func Pdr_InvMinimize(p *Gia_Man_t, vInv *Vec_Int_t, fVerbose int32) *Vec_Int_t {
    var nBTLimit int32;
    var fCheckProperty int32  = 1;
    var clk abctime  = Abc_Clock();
    var n, i, k, status, nLits, fFailed, fCannot, nRemoved int32; 
    var vRes *Vec_Int_t  = nil;
    // create SAT solver
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    var pCube *int32;
    var pList *int32 = Vec_IntArray(vInv);
    var nCubes int32 = pList[0];
    // create variables
    var vLits *Vec_Int_t  = Vec_IntAlloc(100);
    var vRemoved *Vec_Bit_t  = Vec_BitStart( nCubes );
    var iFoVarBeg int32  = pCnf.nVars - Gia_ManRegNum(p);
    var iFiVarBeg int32  = 1 + Gia_ManPoNum(p);
    var iAuxVarBeg int32  = sat_solver_nvars(pSat);
    // allocate auxiliary variables
    assert(sat_solver_nvars(pSat) == pCnf.nVars, "sat_solver_nvars(pSat) == pCnf.nVars");
    sat_solver_setnvars( pSat, sat_solver_nvars(pSat) + nCubes );
    // add clauses
    // inlining of macro Pdr_ForEachCube
    for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
        // collect literals
        Vec_IntFill( vLits, 1, Abc_Var2Lit(iAuxVarBeg + i, 1) ); // neg aux literal
        for  k = 0; k < pCube[0]; k++  {
            Vec_IntPush( vLits, Abc_Var2Lit(iFoVarBeg + Abc_Lit2Var(pCube[k+1]), !Abc_LitIsCompl(pCube[k+1])) );
        }
        // add it to the solver
        status = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits) );
        assert(status == 1, "status == 1");
    }
    // iterate through clauses 
    // inlining of macro Pdr_ForEachCube
    for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
        if Vec_BitEntry(vRemoved, i) {
            continue; }
        // collect aux literals for remaining clauses
        Vec_IntClear( vLits );
        for  k = 0; k < nCubes; k++  {
            if k != i && !Vec_BitEntry(vRemoved, k) { // skip this cube and already removed cubes
                Vec_IntPush( vLits, Abc_Var2Lit(iAuxVarBeg + k, 0) ); }
        } // pos aux literal
        nLits = Vec_IntSize( vLits );
        // check if the property still holds
        if fCheckProperty != 0 {
            for  k = 0; k < Gia_ManPoNum(p); k++  {
                Vec_IntShrink( vLits, nLits );
                Vec_IntPush( vLits, Abc_Var2Lit(1+k, 0) );
                status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits), nBTLimit, 0, 0, 0 );
                if status == l_Undef { // timeout
                
                    fFailed = 1;
                    break;
                }
                if status == l_True { // sat - property fails
                    break; }
                assert(status == l_False, "status == l_False"); // unsat - property holds
            }
            if fFailed {
                break; }
            if k < Gia_ManPoNum(p) {
                continue; }
        }
        // check other clauses
        fCannot = 0;
        // inlining of macro Pdr_ForEachCube
        for  n, pCube = 0, pList + 1; n < pList[0]; n++, pCube += pCube[0] + 1  {
            if Vec_BitEntry(vRemoved, n) || n == i {
                continue; }
            // collect cube
            Vec_IntShrink( vLits, nLits );
            for  k = 0; k < pCube[0]; k++  {
               Vec_IntPush( vLits, Abc_Var2Lit(iFiVarBeg + Abc_Lit2Var(pCube[k+1]), Abc_LitIsCompl(pCube[k+1])) );
            }
            // check if this cube intersects with the complement of other cubes in the solver
            // if it does not intersect, then it is redundant and can be skipped
            status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits), nBTLimit, 0, 0, 0 );
            if status == l_Undef { // timeout
            
                fFailed = 1;
                break;
            }
            if status == l_False { // unsat -- correct
                continue; }
            assert(status == l_True, "status == l_True");
            // cannot remove
            fCannot = 1;
            break;
        }
        if fFailed {
            break; }
        if fCannot {
            continue; }
        if fVerbose != 0 {
        Abc_Print(1, "Removing clause %d.\n", i ); }
        Vec_BitWriteEntry( vRemoved, i, 1 );
        nRemoved++;
    }
    if nRemoved != 0 {
        Abc_Print(1, "Invariant minimization reduced %d clauses (out of %d).  ", nRemoved, nCubes );
    } else {
        Abc_Print(1, "Invariant minimization did not change the invariant.  " ); } 
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    // cleanup cover
    if !fFailed && nRemoved > 0 { // finished without timeout and removed some cubes
    
        vRes = Vec_IntAlloc( 1000 );
        Vec_IntPush( vRes, nCubes-nRemoved );
        // inlining of macro Pdr_ForEachCube
        for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
            if !Vec_BitEntry(vRemoved, i) {
                for  k = 0; k <= pCube[0]; k++  {
                    Vec_IntPush( vRes, pCube[k] );
                } }
        }
        Vec_IntPush( vRes, Vec_IntEntryLast(vInv) );
    }
    Cnf_DataFree( pCnf );
    sat_solver_delete( pSat );
    Vec_BitFree( vRemoved );
    Vec_IntFree( vLits );
    return vRes;
}

func Pdr_InvMinimizeLits(p *Gia_Man_t, vInv *Vec_Int_t, fVerbose int32) *Vec_Int_t {
    var vRes *Vec_Int_t  = nil;
    var clk abctime  = Abc_Clock();
    var i int32;
    var k int32;
    var nLits int32;
    var pCube *int32;
    var pList *int32 = Vec_IntArray(vInv);
    var nRemoved int32;
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat *sat_solver;
//    sat_solver * pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    // inlining of macro Pdr_ForEachCube
    for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
        nLits += pCube[0];
        for  k = 0; k < pCube[0]; k++  {
            var Save int32  = pCube[k+1];
            pCube[k+1] = -1;
            //sat_solver_bookmark( pSat );
            pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
            if Pdr_InvCheck_int(p, vInv, 0, pSat, 1) != 0 {
                pCube[k+1] = Save;
            } else {
                if fVerbose != 0 {
                Abc_Print(1, "Removing lit %d from clause %d.\n", k, i ); }
                nRemoved++;
            }
            sat_solver_delete( pSat );
            //sat_solver_rollback( pSat );
            //sat_solver_bookmark( pSat );
        }
    }
    Cnf_DataFree( pCnf );
    //sat_solver_delete( pSat );
    if nRemoved != 0 {
        Abc_Print(1, "Invariant minimization reduced %d literals (out of %d).  ", nRemoved, nLits );
    } else {
        Abc_Print(1, "Invariant minimization did not change the invariant.  " ); } 
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if nRemoved > 0 { // finished without timeout and removed some lits
    
        vRes = Vec_IntAlloc( 1000 );
        Vec_IntPush( vRes, pList[0] );
        // inlining of macro Pdr_ForEachCube
        for  i, pCube = 0, pList + 1; i < pList[0]; i++, pCube += pCube[0] + 1  {
            var nLits int32;
            for  k = 0; k < pCube[0]; k++  {
                if pCube[k+1] != -1 {
                    nLits++; }
            }
            Vec_IntPush( vRes, nLits );
            for  k = 0; k < pCube[0]; k++  {
                if pCube[k+1] != -1 {
                    Vec_IntPush( vRes, pCube[k+1] ); }
            }
        }
        Vec_IntPush( vRes, Vec_IntEntryLast(vInv) );
    }
    return vRes;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

