package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSatLut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatLut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/tim/tim.h"
//#include "sat/bsat/satStore.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
/**Function*************************************************************

  Synopsis    [For each node in the window, create fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Timing computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Given mapping in p->vSolCur, check if mapping meets delay.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Sbl_ManTestSat(p *Sbl_Man_t, iPivot int32) int32 {
    var fKeepTrying int32  = 1;
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var i, status, Root, Count, StartSol, nConfTotal, nIters int32;
    var nEntries int32  = Hsh_VecSize( p.pHash );
    p.nTried++;

    Sbl_ManClean( p );

    // compute one window
    Count = Sbl_ManWindow2( p, iPivot );
    if Count == 0 {
        if p.fVeryVerbose {
        fmt.Printf("Obj %d: Window with less than %d nodes does not exist.\n", iPivot, p.nVars); }
        p.nSmallWins++;
        return 0;
    }
    Hsh_VecManAdd( p.pHash, p.vAnds );
    if nEntries == Hsh_VecSize(p.pHash) {
        if p.fVeryVerbose {
        fmt.Printf("Obj %d: This window was already tried.\n", iPivot); }
        p.nHashWins++;
        return 0;
    }
    if p.fVeryVerbose {
    fmt.Printf("\nObj = %6d : Leaf = %2d.  AND = %2d.  Root = %2d.    LUT = %2d.\n", iPivot, Vec_IntSize(p.vLeaves), Vec_IntSize(p.vAnds), Vec_IntSize(p.vRoots), Vec_IntSize(p.vNodes)); } 

    if Vec_IntSize(p.vLeaves) > 128 || Vec_IntSize(p.vAnds) > p.nVars {
        if p.fVeryVerbose {
        fmt.Printf("Obj %d: Encountered window with %d inputs and %d internal nodes.\n", iPivot, Vec_IntSize(p.vLeaves), Vec_IntSize(p.vAnds)); }
        p.nLargeWins++;
        return 0;
    }
    if Vec_IntSize(p.vAnds) < 10 {
        if p.fVeryVerbose {
        fmt.Printf("Skipping.\n"); }
        return 0;
    }

    // derive cuts
    Sbl_ManComputeCuts( p );
    // derive SAT instance
    Sbl_ManCreateCnf( p );

    if p.fVeryVeryVerbose {
    fmt.Printf("All clauses = %d.  Multi clauses = %d.  Binary clauses = %d.  Other clauses = %d.\n\n", sat_solver_nclauses(p.pSat), Vec_IntSize(p.vAnds), Vec_WrdSize(p.vCutsI1) - Vec_IntSize(p.vAnds), sat_solver_nclauses(p.pSat) - Vec_WrdSize(p.vCutsI1)); }

    // create assumptions
    // cardinality
    Vec_IntClear( p.vAssump );
    Vec_IntPush( p.vAssump, -1 );
    // unused variables
    for  i = Vec_IntSize(p.vAnds); i < p.Power2; i++  {
        Vec_IntPush( p.vAssump, Abc_Var2Lit(i, 1) );
    }
    // root variables
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vRootVars)) && (((Root) = Vec_IntEntry(p.vRootVars, i)), 1); i++  {
        Vec_IntPush( p.vAssump, Abc_Var2Lit(Root, 0) );
    }
//    Vec_IntPrint( p->vAssump );

    StartSol = Vec_IntSize(p.vSolInit) + 1;
//    StartSol = 30;
    for  (fKeepTrying != 0) && StartSol-fKeepTrying > 0  {
        var Count, LitCount int32;
        var nConfBef, nConfAft int32;
        if p.fVeryVerbose {
            fmt.Printf("Trying to find mapping with %d LUTs.\n", StartSol-fKeepTrying); }
    //    for ( i = Vec_IntSize(p->vSolInit)-5; i < nVars; i++ )
    //        Vec_IntPush( p->vAssump, Abc_Var2Lit(Vec_IntEntry(p->vCardVars, i), 1) );
        Vec_IntWriteEntry( p.vAssump, 0, Abc_Var2Lit(Vec_IntEntry(p.vCardVars, StartSol-fKeepTrying), 1) );
        // solve the problem
        clk2 = Abc_Clock();
        nConfBef = int32(p.pSat.stats.conflicts);
        status = sat_solver_solve( p.pSat, Vec_IntArray(p.vAssump), Vec_IntLimit(p.vAssump), p.nBTLimit, 0, 0, 0 );
        p.timeSat += Abc_Clock() - clk2;
        nConfAft = int32(p.pSat.stats.conflicts);
        nConfTotal += nConfAft - nConfBef;
        nIters++;
        p.nRuns++;
        if status == l_True {
            p.timeSatSat += Abc_Clock() - clk2;
        } else if status == l_False {
            p.timeSatUns += Abc_Clock() - clk2;
        } else { 
            p.timeSatUnd += Abc_Clock() - clk2; }
        if status == l_Undef {
            break; }
        if status == l_True {
            if p.fVeryVeryVerbose {
                for  i = 0; i < Vec_IntSize(p.vAnds); i++  {
                    fmt.Printf("%d", sat_solver_var_value(p.pSat, i));
                }
                fmt.Printf("\n");
                for  i = 0; i < Vec_IntSize(p.vAnds); i++  {
                    if sat_solver_var_value(p.pSat, i) {
                        fmt.Printf("%d=%d ", i, sat_solver_var_value(p.pSat, i));
                        Count++;
                    }
                }
                fmt.Printf("Count = %d\n", Count);
            }
//            for ( i = p->FirstVar; i < sat_solver_nvars(p->pSat); i++ )
//                printf( "%d", sat_solver_var_value(p->pSat, i) );
//            printf( "\n" );
            Count = 1;
            Vec_IntClear( p.vSolCur );
            for  i = p.FirstVar; i < sat_solver_nvars(p.pSat); i++  {
                if sat_solver_var_value(p.pSat, i) {
                    if p.fVeryVeryVerbose {
                        fmt.Printf("Cut %3d : Node = %3d %6d  ", Count++, Vec_IntEntry(p.vCutsObj, i-p.FirstVar), Vec_IntEntry(p.vAnds, Vec_IntEntry(p.vCutsObj, i-p.FirstVar))); }
                    if p.fVeryVeryVerbose {
                        LitCount += Sbl_ManFindAndPrintCut( p, i-p.FirstVar ); }
                    Vec_IntPush( p.vSolCur, i-p.FirstVar );
                }
            }
            //Vec_IntPrint( p->vRootVars );
            //Vec_IntPrint( p->vRoots );
            //Vec_IntPrint( p->vAnds );
            //Vec_IntPrint( p->vLeaves );
        }

//        fKeepTrying = status == l_True ? fKeepTrying + 1 : 0;
        // check the timing
        if status == l_True {
            if p.fDelay && !(Sbl_ManEvaluateMapping(p, p.DelayMax) != 0) {
                var iLit, value int32;
                if p.fVeryVerbose {
                    fmt.Printf("Critical path of length (%d) is detected:   ", Vec_IntSize(p.vPath));
                    // inlining of macro Vec_IntForEachEntry
                    for  i = 0; (i < Vec_IntSize(p.vPath)) && (((iLit) = Vec_IntEntry(p.vPath, i)), 1); i++  {
                        fmt.Printf("%d=%d ", i, Vec_IntEntry(p.vAnds, Abc_Lit2Var(iLit)));
                    }
                    fmt.Printf("\n");
                }
                // add the clause
                value = sat_solver_addclause( p.pSat, Vec_IntArray(p.vPath), Vec_IntLimit(p.vPath)  );
                assert((value != 0), "value");
            
            } else {
                Vec_IntClear( p.vSolBest );
                Vec_IntAppend( p.vSolBest, p.vSolCur );
                fKeepTrying++;
            }
        
        } else {
            fKeepTrying = 0; }
        if p.fVeryVerbose {
            if status == l_False {
                fmt.Printf("UNSAT ");
            } else if status == l_True {
                fmt.Printf("SAT   ");
            } else { 
                fmt.Printf("UNDEC "); }
            fmt.Printf("confl =%8d.    ", nConfAft - nConfBef);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk2 );

            fmt.Printf("Total ");
            fmt.Printf("confl =%8d.    ", nConfTotal);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            if p.fVeryVeryVerbose && status == l_True {
                fmt.Printf("LitCount = %d.\n", LitCount); }
            fmt.Printf("\n");
        }
        if nIters == 10 {
            p.nIterOuts++;
            //printf( "Obj %d : Quitting after %d iterations.\n", iPivot, nIters );
            break;
        }
    }

    // update solution
    if Vec_IntSize(p.vSolBest) > 0 && Vec_IntSize(p.vSolBest) < Vec_IntSize(p.vSolInit) {
        var nDelayCur, nEdgesCur int32;
        Sbl_ManUpdateMapping( p );
        if p.pGia.vEdge1 {
            nDelayCur = Gia_ManEvalEdgeDelay( p.pGia );
            nEdgesCur = Gia_ManEvalEdgeCount( p.pGia );
        
        } else {
            nDelayCur = Sbl_ManCreateTiming( p, p.DelayMax ); }
        if p.fVerbose {
        fmt.Printf("Object %5d : Saved %2d nodes  (Conf =%8d)  Iter =%3d  Delay = %d  Edges = %4d\n", iPivot, Vec_IntSize(p.vSolInit)-Vec_IntSize(p.vSolBest), nConfTotal, nIters, nDelayCur, nEdgesCur); }
        p.timeTotal += Abc_Clock() - p.timeStart;
        p.nImproved++;
        return 2;
    
    } else {
//        printf( "Object %5d : Saved %2d nodes  (Conf =%8d)  Iter =%3d\n", iPivot, 0, nConfTotal, nIters );
    }
    p.timeTotal += Abc_Clock() - p.timeStart;
    return 1;
}
func Sbl_ManPrintRuntime(p *Sbl_Man_t) {
    fmt.Printf("Runtime breakdown:\n");
    p.timeOther = p.timeTotal - p.timeWin - p.timeCut - p.timeSat - p.timeTime;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Win   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeWin)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeWin)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Cut   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCut)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeCut)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Sat   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Sat  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Unsat"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUns)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUns)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Undec"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnd)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUnd)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Timing"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTime)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTime)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Other "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("ALL   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
}
func Gia_ManLutSat(pGia *Gia_Man_t, LutSize int32, nNumber int32, nImproves int32, nBTLimit int32, DelayMax int32, nEdges int32, fDelay int32, fReverse int32, fVerbose int32, fVeryVerbose int32) {
    var iLut, nImproveCount int32;
    var p *Sbl_Man_t    = Sbl_ManAlloc( pGia, nNumber );
    p.LutSize      = LutSize;      // LUT size
    p.nBTLimit     = nBTLimit;     // conflicts
    p.DelayMax     = DelayMax;     // external delay
    p.nEdges       = nEdges;       // the number of edges
    p.fDelay       = fDelay;       // delay mode
    p.fReverse     = fReverse;     // reverse windowing
    p.fVerbose     = fVerbose | fVeryVerbose;
    p.fVeryVerbose = fVeryVerbose;
    if p.fVerbose {
    fmt.Printf("Parameters: WinSize = %d AIG nodes.  Conf = %d.  DelayMax = %d.\n", p.nVars, p.nBTLimit, p.DelayMax); }
    // determine delay limit
    if (fDelay != 0) && pGia.vEdge1 && p.DelayMax == 0 {
        p.DelayMax = Gia_ManEvalEdgeDelay( pGia ); }
    // iterate through the internal nodes
    Gia_ManComputeOneWinStart( pGia, nNumber, fReverse );
    // inlining of macro Gia_ManForEachLut2
    for  iLut = 1; iLut < Gia_ManObjNum(pGia); iLut++  { if !Gia_ObjIsLut2(pGia, iLut) {} else {
        if Sbl_ManTestSat( p, iLut ) != 2 {
            continue; }
        if nImproveCount++; nImproveCount == nImproves {
            break; }
    }
    }
    Gia_ManComputeOneWin( pGia, -1, nil, nil, nil, nil );
    if p.fVerbose {
    fmt.Printf("Tried = %d. Used = %d. HashWin = %d. SmallWin = %d. LargeWin = %d. IterOut = %d.  SAT runs = %d.\n", p.nTried, p.nImproved, p.nHashWins, p.nSmallWins, p.nLargeWins, p.nIterOuts, p.nRuns); }
    if p.fVerbose {
    Sbl_ManPrintRuntime( p ); }
    Sbl_ManStop( p );
    Vec_IntFreeP( &pGia.vPacking );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

