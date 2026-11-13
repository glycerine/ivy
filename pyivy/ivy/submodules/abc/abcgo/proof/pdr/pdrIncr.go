package pdr

import (
	"fmt"
	"os"
	"time"

 "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [pdrIncr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [PDR with incremental solving.]

  Author      [Yen-Sheng Ho, Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Feb. 17, 2017.]

  Revision    [$Id: pdrIncr.c$]

***********************************************************************/

//#include "pdrInt.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

// extern int Pdr_ManBlockCube( Pdr_Man_t * p, Pdr_Set_t * pCube );

// extern int Pdr_ManPushClauses( Pdr_Man_t * p );

// extern Pdr_Set_t * Pdr_ManReduceClause( Pdr_Man_t * p, int k, Pdr_Set_t * pCube );

// extern int Gia_ManToBridgeAbort( FILE * pFile, int Size, unsigned char * pBuffer );

// extern int Gia_ManToBridgeResult( FILE * pFile, int Result, Abc_Cex_t * pCex, int iPoProved );



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func IPdr_ManPushClausesK(p *Pdr_Man_t, k int32) *Vec_Ptr_t {
 if vec.Vec_VecSize(p.vClauses) != int(k)+1 { panic("pdr: unexpected final frame") }
 current := vec.Vec_VecEntry(p.vClauses, int(k))
 vec.Vec_PtrSort(current, func(a, b *any) int {
  left, right := (*a).(*Pdr_Set_t), (*b).(*Pdr_Set_t)
  return int(Pdr_SetCompare(&left, &right))
 })
 next := vec.Vec_PtrAlloc(100)
 remove := func(frame *Vec_Ptr_t, index int) {
  Pdr_SetDeref(vec.Vec_PtrEntry(frame, index).(*Pdr_Set_t))
  vec.Vec_PtrWriteEntry(frame, index, vec.Vec_PtrEntryLast(frame))
  vec.Vec_PtrPop(frame)
 }
 for j := 0; j < vec.Vec_PtrSize(current); {
  cube := vec.Vec_PtrEntry(current, j).(*Pdr_Set_t)
  for m := j+1; m < vec.Vec_PtrSize(current); {
   if Pdr_SetContains(vec.Vec_PtrEntry(current, m).(*Pdr_Set_t), cube) == 0 { m++; continue }
   remove(current, m)
  }
  result := Pdr_ManCheckCube(p, k, cube, nil, 0, 0, 1)
  if result == -1 { panic("pdr: incremental clause push returned unknown") }
  if result == 0 { j++; continue }
  if reduced := Pdr_ManReduceClause(p, k, cube); reduced != nil { Pdr_SetDeref(cube); cube = reduced }
  for i := 0; i < vec.Vec_PtrSize(next); {
   if Pdr_SetContains(vec.Vec_PtrEntry(next, i).(*Pdr_Set_t), cube) == 0 { i++; continue }
   remove(next, i)
  }
  vec.Vec_PtrPush(next, cube)
  vec.Vec_PtrWriteEntry(current, j, vec.Vec_PtrEntryLast(current))
  vec.Vec_PtrPop(current)
 }
 return next
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IPdr_ManPrintClauses(clauses *Vec_Vec_t, start, nRegs int32) {
 var count int
 for frame := int(start); frame < vec.Vec_VecSize(clauses); frame++ {
  cubes := vec.Vec_VecEntry(clauses, frame)
  vec.Vec_PtrSort(cubes, func(a, b *any) int {
   left, right := (*a).(*Pdr_Set_t), (*b).(*Pdr_Set_t)
   return int(Pdr_SetCompare(&left, &right))
  })
  for _, item := range vec.Vec_PtrArray(cubes) {
   fmt.Printf("Frame[%4d]Cube[%4d] = ", frame, count)
   ZPdr_SetPrint(item.(*Pdr_Set_t))
   fmt.Println()
   count++
  }
 }
}


/**Function*************************************************************

  Synopsis    [ Check if each cube c_k in frame k satisfies the query
                R_{k-1} && T && !c_k && c_k' (must be UNSAT).
                Return True if all cubes pass the check. ]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IPdr_ManCheckClauses(p *Pdr_Man_t) int32 {
 for frame := 1; frame < len(p.vSolvers); frame++ {
  for i, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
   result := Pdr_ManCheckCube(p, int32(frame-1), item.(*Pdr_Set_t), nil, 0, 0, 1)
   if result == 0 { fmt.Printf("Cube[%d][%d] not inductive!\n", frame, i) }
   if result == -1 { return -1 }
  }
 }
 return 1
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IPdr_ManSaveClauses(p *Pdr_Man_t, dropLast int32) *Vec_Vec_t {
 n := vec.Vec_VecSize(p.vClauses)
 if n == 1 || (n == 2 && dropLast != 0) { return nil }
 if dropLast != 0 { n-- }
 saved := vec.Vec_VecStart(n)
 for frame := 0; frame < n; frame++ {
  for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
   vec.Vec_VecPush(saved, frame, Pdr_SetDup(item.(*Pdr_Set_t)))
  }
 }
 return saved
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IPdr_ManSetSolver(p *Pdr_Man_t, k int32, fSetPropOutput int32) *sat_solver {
    var pSat *sat_solver;
    var vArrayK *Vec_Ptr_t;
    var pCube *Pdr_Set_t;
    var i, j int32;

    assert(Vec_PtrSize(p.vSolvers) == k, "Vec_PtrSize(p.vSolvers) == k");
    assert(Vec_IntSize(p.vActVars) == k, "Vec_IntSize(p.vActVars) == k");

    pSat = zsat_solver_new_seed(p.pPars.nRandomSeed);
    pSat = Pdr_ManNewSolver( pSat, p, k, int32((k == 0)) );
    Vec_PtrPush( p.vSolvers, pSat );
    Vec_IntPush( p.vActVars, 0 );

    // set the property output
    if fSetPropOutput != 0 {
        Pdr_ManSetPropertyOutput( p, k ); }

    if k == 0 {
        return pSat; }

    // add the clauses
    Vec_VecForEachLevelStart( p.vClauses, vArrayK, i, k )
        Vec_PtrForEachEntry( Pdr_Set_t *, vArrayK, pCube, j )
            Pdr_ManSolverAddClause( p, k, pCube );
    return pSat;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IPdr_ManRebuildClauses(p *Pdr_Man_t, vClauses *Vec_Vec_t) int32 {
    var vArrayK *Vec_Ptr_t;
    var pCube *Pdr_Set_t;
    var i, j int32;
    var RetValue int32  = -1;
    var nCubes int32;

    if vClauses == nil {
        return RetValue; }

    assert(Vec_VecSize(vClauses) >= 2, "Vec_VecSize(vClauses) >= 2");
    assert(Vec_VecSize(p.vClauses) == 0, "Vec_VecSize(p.vClauses) == 0"); 
    Vec_VecExpand( p.vClauses, 1 );

    IPdr_ManSetSolver( p, 0, 1 );

    // push clauses from R0 to R1
    Vec_VecForEachLevelStart( vClauses, vArrayK, i, 1 )
    {
        Vec_PtrForEachEntry( Pdr_Set_t *, vArrayK, pCube, j )
        {
            ++nCubes;

            RetValue = Pdr_ManCheckCube( p, 0, pCube, nil, 0, 0, 1 );
            Vec_IntWriteEntry( p.vActVars, 0, 0 );

            assert(RetValue != -1, "RetValue != -1");

            if RetValue == 0 {
                Abc_Print( 1, "Cube[%d][%d] cannot be pushed from R0 to R1.\n", i, j );
                Pdr_SetDeref( pCube );
                continue;
            }

            Vec_VecPush( p.vClauses, 1, pCube );
        }
    }
    Abc_Print( 1, "RebuildClauses: %d out of %d cubes reused in R1.\n", Vec_PtrSize(Vec_VecEntry(p.vClauses, 1)), nCubes );
    IPdr_ManSetSolver( p, 1, 0 );
    Vec_VecFree( vClauses );

    /*
    for ( i = 1; i < Vec_VecSize( vClauses ) - 1; ++i )
    {
        vArrayK = IPdr_ManPushClausesK( p, i );
        if ( Vec_PtrSize(vArrayK) == 0 )
        {
            Abc_Print( 1, "RebuildClauses: stopped at frame %d.\n", i );
            break;
        }

        Vec_PtrGrow( (Vec_Ptr_t *)(p->vClauses), i + 2 );
        p->vClauses->nSize = i + 2;
        p->vClauses->pArray[i+1] = vArrayK;
        IPdr_ManSetSolver( p, i + 1, 0 );
    }

    Abc_Print( 1, "After rebuild:" );
    Vec_VecForEachLevel( p->vClauses, vArrayK, i )
        Abc_Print( 1, " %d", Vec_PtrSize( vArrayK ) );
    Abc_Print( 1, "\n" );
    */

    return 0;
}

func IPdr_ManRestoreAbsFlops(p *Pdr_Man_t) int32 {
 for frame := 0; frame < vec.Vec_VecSize(p.vClauses); frame++ {
  for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
   cube := item.(*Pdr_Set_t)
   for _, lit := range cube.Lits[:cube.nLits] { vec.Vec_IntWriteEntry(p.vAbsFlops, int(lit>>1), 1) }
  }
 }
 return 0
}

func IPdr_ManRestoreClauses(p *Pdr_Man_t, clauses *Vec_Vec_t, mapping *Vec_Int_t) int32 {
 if clauses == nil { panic("pdr: nil saved clauses") }
 vec.Vec_VecFree(p.vClauses)
 p.vClauses = clauses
 if mapping != nil {
  for frame := 0; frame < vec.Vec_VecSize(clauses); frame++ {
   for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(clauses, frame)) {
    cube := item.(*Pdr_Set_t)
    for i := int32(0); i < cube.nLits; i++ {
     lit := cube.Lits[i]
     cube.Lits[i] = 2*vec.Vec_IntEntry(mapping, int(lit>>1)) + (lit & 1)
    }
   }
  }
 }
 for frame := 0; frame < vec.Vec_VecSize(clauses); frame++ {
  property := int32(0)
  if frame < vec.Vec_VecSize(clauses)-1 { property = 1 }
  IPdr_ManSetSolver(p, int32(frame), property)
 }
 return 0
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IPdr_ManSolveInt(p *Pdr_Man_t, fCheckClauses int32, fPushClauses int32) int32 {
    var fPrintClauses int32;
    var pCube *Pdr_Set_t  = nil;
    var pObj *Aig_Obj_t;
    var pCexNew *Abc_Cex_t;
    var iFrame int32;
    var RetValue int32 = -1;
    var nOutDigits int32  = Abc_Base10Log( Saig_ManPoNum(p.pAig) );
    var clkStart abctime = Abc_Clock();
    var clkOne abctime;
    if p.pPars.nTimeOut {
        p.timeToStop = p.pPars.nTimeOut * int64(time.Second) + Abc_Clock();
    } else {
        p.timeToStop = 0;
    }
    // assert( Vec_PtrSize(p->vSolvers) == 0 );
    // in the multi-output mode, mark trivial POs (those fed by const0) as solved 
    if ( p.pPars.fSolveAll )
        Saig_ManForEachPo( p.pAig, pObj, iFrame )
            if Aig_ObjChild0(pObj) == Aig_ManConst0(p.pAig) {
                Vec_IntWriteEntry( p.pPars.vOutMap, iFrame, 1 ); // unsat
                p.pPars.nProveOuts++;
                if p.pPars.fUseBridge {
                    Gia_ManToBridgeResult( os.Stdout, 1, nil, iFrame ); }
            }
    // create the first timeframe
    p.pPars.timeLastSolved = Abc_Clock();

    if Vec_VecSize(p.vClauses) == 0 {
        Pdr_ManCreateSolver( p, (iFrame = 0) );
    } else {
        iFrame = Vec_VecSize(p.vClauses) - 1;

        if fCheckClauses != 0 {
            if p.pPars.fVerbose {
               Abc_Print( 1, "IPDR: Checking the reloaded length-%d trace...", iFrame + 1 ) ; }
            IPdr_ManCheckClauses( p );
            if p.pPars.fVerbose {
               Abc_Print( 1, " Passed!\n" ) ; }
        }

        if fPushClauses != 0 {
            p.iUseFrame = Abc_MaxInt(iFrame, 1);

            if p.pPars.fVerbose {
                Abc_Print( 1, "IPDR: Pushing the reloaded clauses. Before:\n" );
                Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart );
            }

            RetValue = Pdr_ManPushClauses( p );

            if p.pPars.fVerbose {
                Abc_Print( 1, "IPDR: Finished pushing. After:\n" );
                Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart );
            }

            if RetValue != 0 {
                Pdr_ManReportInvariant( p );
                Pdr_ManVerifyInvariant( p );
                return 1;
            }
        }

        if p.pPars.fUseAbs && p.vAbsFlops == nil && iFrame >= 1 {
            assert(p.vAbsFlops == nil, "p.vAbsFlops == nil");
            p.vAbsFlops  = Vec_IntStart( Saig_ManRegNum(p.pAig) );
            p.vMapFf2Ppi = Vec_IntStartFull( Saig_ManRegNum(p.pAig) );
            p.vMapPpi2Ff = Vec_IntAlloc( 100 );
            
            IPdr_ManRestoreAbsFlops( p );
        }
 
    }
    for {
        var fRefined int32;
        if p.pPars.fUseAbs && p.vAbsFlops == nil && iFrame == 1 {
//            int i, Prio;
            assert(p.vAbsFlops == nil, "p.vAbsFlops == nil");
            p.vAbsFlops  = Vec_IntStart( Saig_ManRegNum(p.pAig) );
            p.vMapFf2Ppi = Vec_IntStartFull( Saig_ManRegNum(p.pAig) );
            p.vMapPpi2Ff = Vec_IntAlloc( 100 );
//            Vec_IntForEachEntry( p->vPrio, Prio, i )
//                if ( Prio >> p->nPrioShift  )
//                    Vec_IntWriteEntry( p->vAbsFlops, i, 1 );
        }
        //if ( p->pPars->fUseAbs && p->vAbsFlops )
        //    printf( "Starting frame %d with %d (%d) flops.\n", iFrame, Vec_IntCountPositive(p->vAbsFlops), Vec_IntCountPositive(p->vPrio) );
        p.nFrames = iFrame;
        assert(iFrame == Vec_PtrSize(p.vSolvers)-1, "iFrame == Vec_PtrSize(p.vSolvers)-1");
        p.iUseFrame = Abc_MaxInt(iFrame, 1);
        Saig_ManForEachPo( p.pAig, pObj, p.iOutCur )
        {
            // skip disproved outputs
            if p.vCexes && Vec_PtrEntry(p.vCexes, p.iOutCur) {
                continue; }
            // skip output whose time has run out
            if p.pTime4Outs && p.pTime4Outs[p.iOutCur] == 0 {
                continue; }
            // check if the output is trivially solved
            if Aig_ObjChild0(pObj) == Aig_ManConst0(p.pAig) {
                continue; }
            // check if the output is trivially solved
            if Aig_ObjChild0(pObj) == Aig_ManConst1(p.pAig) {
                if !p.pPars.fSolveAll {
                    pCexNew = Abc_CexMakeTriv( Aig_ManRegNum(p.pAig), Saig_ManPiNum(p.pAig), Saig_ManPoNum(p.pAig), iFrame*Saig_ManPoNum(p.pAig)+p.iOutCur );
                    p.pAig.pSeqModel = pCexNew;
                    return 0; // SAT
                }
                if p.pPars.fUseBridge || p.pPars.fStoreCex {
                    pCexNew = Abc_CexMakeTriv( Aig_ManRegNum(p.pAig), Saig_ManPiNum(p.pAig), Saig_ManPoNum(p.pAig), iFrame*Saig_ManPoNum(p.pAig)+p.iOutCur );
                } else {
                    pCexNew = (Abc_Cex_t *)(ABC_PTRINT_T)1;
                }
                p.pPars.nFailOuts++;
                if p.pPars.vOutMap { Vec_IntWriteEntry( p.pPars.vOutMap, p.iOutCur, 0 ); }
                if !p.pPars.fNotVerbose {
                Abc_Print( 1, "Output %*d was trivially asserted in frame %2d (solved %*d out of %*d outputs).\n",
                    nOutDigits, p.iOutCur, iFrame, nOutDigits, p.pPars.nFailOuts, nOutDigits, Saig_ManPoNum(p.pAig) ); }
                assert(Vec_PtrEntry(p.vCexes, p.iOutCur) == nil, "Vec_PtrEntry(p.vCexes, p.iOutCur) == nil");
                if p.pPars.fUseBridge {
                    Gia_ManToBridgeResult( os.Stdout, 0, pCexNew, pCexNew.iPo ); }
                Vec_PtrWriteEntry( p.vCexes, p.iOutCur, pCexNew );
                if p.pPars.pFuncOnFail && p.pPars.pFuncOnFail(p.iOutCur, p.pPars.fStoreCex ? (Abc_Cex_t *)Vec_PtrEntry(p.vCexes, p.iOutCur) : nil) {
                    if p.pPars.fVerbose {
                        Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
                    if !p.pPars.fSilent {
                        Abc_Print( 1, "Quitting due to callback on fail in frame %d.\n", iFrame ); }
                    p.pPars.iFrame = iFrame;
                    return -1;
                }
                if p.pPars.nFailOuts + p.pPars.nDropOuts == Saig_ManPoNum(p.pAig) {
                    if p.pPars.nFailOuts {
                        return 0;
                    } else {
                        return -1;
                    } } // SAT or UNDEC
                p.pPars.timeLastSolved = Abc_Clock();
                continue;
            }
            // try to solve this output
            if p.pTime4Outs {
                assert(p.pTime4Outs[p.iOutCur] > 0, "p.pTime4Outs[p.iOutCur] > 0");
                clkOne = Abc_Clock();
                p.timeToStopOne = p.pTime4Outs[p.iOutCur] + Abc_Clock();
            }
            for {
                if p.pPars.nTimeOutGap && p.pPars.timeLastSolved && Abc_Clock() > p.pPars.timeLastSolved + p.pPars.nTimeOutGap * int64(time.Second) {
                    if p.pPars.fVerbose {
                        Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
                    if !p.pPars.fSilent {
                        Abc_Print( 1, "Reached gap timeout (%d seconds) in frame %d.\n",  p.pPars.nTimeOutGap, iFrame ); }
                    p.pPars.iFrame = iFrame;
                    return -1;
                }
                RetValue = Pdr_ManCheckCube( p, iFrame, nil, &pCube, p.pPars.nConfLimit, 0, 1 );
                if RetValue == 1 {
                    break; }
                if RetValue == -1 {
                    if p.pPars.fVerbose {
                        Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
                    if p.timeToStop && Abc_Clock() > p.timeToStop {
                        Abc_Print( 1, "Reached timeout (%d seconds) in frame %d.\n",  p.pPars.nTimeOut, iFrame );
                    } else if p.pPars.nTimeOutGap && p.pPars.timeLastSolved && Abc_Clock() > p.pPars.timeLastSolved + p.pPars.nTimeOutGap * int64(time.Second) {
                        Abc_Print( 1, "Reached gap timeout (%d seconds) in frame %d.\n",  p.pPars.nTimeOutGap, iFrame );
                    } else if p.timeToStopOne && Abc_Clock() > p.timeToStopOne {
                        Pdr_QueueClean( p );
                        pCube = nil;
                        break; // keep solving
                    
                    } else if p.pPars.nConfLimit {
                        Abc_Print( 1, "Reached conflict limit (%d) in frame %d.\n",  p.pPars.nConfLimit, iFrame );
                    } else if p.pPars.fVerbose {
                        Abc_Print( 1, "Computation cancelled by the callback in frame %d.\n", iFrame ); }
                    p.pPars.iFrame = iFrame;
                    return -1;
                }
                if RetValue == 0 {
                    RetValue = Pdr_ManBlockCube( p, pCube );
                    if RetValue == -1 {
                        if p.pPars.fVerbose {
                            Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
                        if p.timeToStop && Abc_Clock() > p.timeToStop {
                            Abc_Print( 1, "Reached timeout (%d seconds) in frame %d.\n",  p.pPars.nTimeOut, iFrame );
                        } else if p.pPars.nTimeOutGap && p.pPars.timeLastSolved && Abc_Clock() > p.pPars.timeLastSolved + p.pPars.nTimeOutGap * int64(time.Second) {
                            Abc_Print( 1, "Reached gap timeout (%d seconds) in frame %d.\n",  p.pPars.nTimeOutGap, iFrame );
                        } else if p.timeToStopOne && Abc_Clock() > p.timeToStopOne {
                            Pdr_QueueClean( p );
                            pCube = nil;
                            break; // keep solving
                        
                        } else if p.pPars.nConfLimit {
                            Abc_Print( 1, "Reached conflict limit (%d) in frame %d.\n",  p.pPars.nConfLimit, iFrame );
                        } else if p.pPars.fVerbose {
                            Abc_Print( 1, "Computation cancelled by the callback in frame %d.\n", iFrame ); }
                        p.pPars.iFrame = iFrame;
                        return -1;
                    }
                    if RetValue == 0 {
                        if fPrintClauses != 0 {
                            Abc_Print( 1, "*** Clauses after frame %d:\n", iFrame );
                            Pdr_ManPrintClauses( p, 0 );
                        }
                        if p.pPars.fVerbose && !p.pPars.fUseAbs {
                            Pdr_ManPrintProgress( p, !p.pPars.fSolveAll, Abc_Clock() - clkStart ); }
                        p.pPars.iFrame = iFrame;
                        if !p.pPars.fSolveAll {
                            var clk abctime  = Abc_Clock();
                            var pCex *Abc_Cex_t  = Pdr_ManDeriveCexAbs(p);
                            p.tAbs += Abc_Clock() - clk;
                            if pCex == nil {
                                assert(p.pPars.fUseAbs, "p.pPars.fUseAbs");
                                Pdr_QueueClean( p );
                                pCube = nil;
                                fRefined = 1;
                                break; // keep solving
                            }
                            p.pAig.pSeqModel = pCex;

                            if p.pPars.fVerbose && p.pPars.fUseAbs {
                                Pdr_ManPrintProgress( p, !p.pPars.fSolveAll, Abc_Clock() - clkStart ); }
                            return 0; // SAT
                        }
                        p.pPars.nFailOuts++;
                        if p.pPars.fUseBridge || p.pPars.fStoreCex {
                            pCexNew = Pdr_ManDeriveCex(p);
                        } else {
                            pCexNew = (Abc_Cex_t *)(ABC_PTRINT_T)1;
                        }
                        if p.pPars.vOutMap { Vec_IntWriteEntry( p.pPars.vOutMap, p.iOutCur, 0 ); }
                        assert(Vec_PtrEntry(p.vCexes, p.iOutCur) == nil, "Vec_PtrEntry(p.vCexes, p.iOutCur) == nil");
                        if p.pPars.fUseBridge {
                            Gia_ManToBridgeResult( os.Stdout, 0, pCexNew, pCexNew.iPo ); }
                        Vec_PtrWriteEntry( p.vCexes, p.iOutCur, pCexNew );
                        if p.pPars.pFuncOnFail && p.pPars.pFuncOnFail(p.iOutCur, p.pPars.fStoreCex ? (Abc_Cex_t *)Vec_PtrEntry(p.vCexes, p.iOutCur) : nil) {
                            if p.pPars.fVerbose {
                                Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
                            if !p.pPars.fSilent {
                                Abc_Print( 1, "Quitting due to callback on fail in frame %d.\n", iFrame ); }
                            p.pPars.iFrame = iFrame;
                            return -1;
                        }
                        if !p.pPars.fNotVerbose {
                            Abc_Print( 1, "Output %*d was asserted in frame %2d (%2d) (solved %*d out of %*d outputs).\n",
                                nOutDigits, p.iOutCur, iFrame, iFrame, nOutDigits, p.pPars.nFailOuts, nOutDigits, Saig_ManPoNum(p.pAig) ); }
                        if p.pPars.nFailOuts == Saig_ManPoNum(p.pAig) {
                            return 0; } // all SAT
                        Pdr_QueueClean( p );
                        pCube = nil;
                        break; // keep solving
                    }
                    if p.pPars.fVerbose {
                        Pdr_ManPrintProgress( p, 0, Abc_Clock() - clkStart ); }
                }
            }
            if fRefined != 0 {
                break; }
            if p.pTime4Outs {
                var timeSince abctime  = Abc_Clock() - clkOne;
                assert(p.pTime4Outs[p.iOutCur] > 0, "p.pTime4Outs[p.iOutCur] > 0");
                p.pTime4Outs[p.iOutCur] = (p.pTime4Outs[p.iOutCur] > timeSince) ? p.pTime4Outs[p.iOutCur] - timeSince : 0;
                if p.pTime4Outs[p.iOutCur] == 0 && Vec_PtrEntry(p.vCexes, p.iOutCur) == nil { // undecided
                
                    p.pPars.nDropOuts++;
                    if p.pPars.vOutMap { 
                        Vec_IntWriteEntry( p.pPars.vOutMap, p.iOutCur, -1 ); }
                    if !p.pPars.fNotVerbose { 
                        Abc_Print( 1, "Timing out on output %*d in frame %d.\n", nOutDigits, p.iOutCur, iFrame ); }
                }
                p.timeToStopOne = 0;
            }
        }
        /*
        if ( p->pPars->fUseAbs && p->vAbsFlops && !fRefined )
        {
            int i, Used;
            Vec_IntForEachEntry( p->vAbsFlops, Used, i )
                if ( Used && (Vec_IntEntry(p->vPrio, i) >> p->nPrioShift) == 0 )
                    Vec_IntWriteEntry( p->vAbsFlops, i, 0 );
        }
        */
        if p.pPars.fUseAbs && p.vAbsFlops && !(fRefined != 0) {
            var pSet *Pdr_Set_t;
            var i, j, k int32;
            Vec_IntFill( p.vAbsFlops, Vec_IntSize( p.vAbsFlops ), 0 );
            Vec_VecForEachEntry( Pdr_Set_t *, p.vClauses, pSet, i, j )
                for ( k = 0; k < pSet.nLits; k++ )
                    Vec_IntWriteEntry( p.vAbsFlops, Abc_Lit2Var(pSet.Lits[k]), 1 );
        }

        if p.pPars.fVerbose {
            Pdr_ManPrintProgress( p, !fRefined, Abc_Clock() - clkStart ); }
        if fRefined != 0 {
            continue; }
        //if ( p->pPars->fUseAbs && p->vAbsFlops )
        //    printf( "Finished frame %d with %d (%d) flops.\n", iFrame, Vec_IntCountPositive(p->vAbsFlops), Vec_IntCountPositive(p->vPrio) );
        // open a new timeframe
        p.nQueLim = p.pPars.nRestLimit;
        assert(pCube == nil, "pCube == nil");
        Pdr_ManSetPropertyOutput( p, iFrame );
        Pdr_ManCreateSolver( p, ++iFrame );
        if fPrintClauses != 0 {
            Abc_Print( 1, "*** Clauses after frame %d:\n", iFrame );
            Pdr_ManPrintClauses( p, 0 );
        }
        // push clauses into this timeframe
        RetValue = Pdr_ManPushClauses( p );
        if RetValue == -1 {
            if p.pPars.fVerbose {
                Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
            if !p.pPars.fSilent {
                if p.timeToStop && Abc_Clock() > p.timeToStop {
                    Abc_Print( 1, "Reached timeout (%d seconds) in frame %d.\n",  p.pPars.nTimeOut, iFrame );
                } else {
                    Abc_Print( 1, "Reached conflict limit (%d) in frame.\n",  p.pPars.nConfLimit, iFrame ); }
            }
            p.pPars.iFrame = iFrame;
            return -1;
        }
        if RetValue != 0 {
            if p.pPars.fVerbose {
                Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
            if !p.pPars.fSilent {
                Pdr_ManReportInvariant( p ); }
            if !p.pPars.fSilent {
                Pdr_ManVerifyInvariant( p ); }
            p.pPars.iFrame = iFrame;
            // count the number of UNSAT outputs
            p.pPars.nProveOuts = Saig_ManPoNum(p.pAig) - p.pPars.nFailOuts - p.pPars.nDropOuts;
            // convert previously 'unknown' into 'unsat'
            if p.pPars.vOutMap {
                for  iFrame = 0; iFrame < Saig_ManPoNum(p.pAig); iFrame++  {
                    if Vec_IntEntry(p.pPars.vOutMap, iFrame) == -2 { // unknown
                    
                        Vec_IntWriteEntry( p.pPars.vOutMap, iFrame, 1 ); // unsat
                        if p.pPars.fUseBridge {
                            Gia_ManToBridgeResult( os.Stdout, 1, nil, iFrame ); }
                    }
                } }
            if p.pPars.nProveOuts == Saig_ManPoNum(p.pAig) {
                return 1; } // UNSAT
            if p.pPars.nFailOuts > 0 {
                return 0; } // SAT
            return -1;
        }
        if p.pPars.fVerbose {
            Pdr_ManPrintProgress( p, 0, Abc_Clock() - clkStart ); }

        // check termination
        if p.pPars.pFuncStop && p.pPars.pFuncStop(p.pPars.RunId) {
            p.pPars.iFrame = iFrame;
            return -1;
        }
        if p.timeToStop && Abc_Clock() > p.timeToStop {
            if fPrintClauses != 0 {
                Abc_Print( 1, "*** Clauses after frame %d:\n", iFrame );
                Pdr_ManPrintClauses( p, 0 );
            }
            if p.pPars.fVerbose {
                Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
            if !p.pPars.fSilent {
                Abc_Print( 1, "Reached timeout (%d seconds) in frame %d.\n",  p.pPars.nTimeOut, iFrame ); }
            p.pPars.iFrame = iFrame;
            return -1;
        }
        if p.pPars.nTimeOutGap && p.pPars.timeLastSolved && Abc_Clock() > p.pPars.timeLastSolved + p.pPars.nTimeOutGap * int64(time.Second) {
            if fPrintClauses != 0 {
                Abc_Print( 1, "*** Clauses after frame %d:\n", iFrame );
                Pdr_ManPrintClauses( p, 0 );
            }
            if p.pPars.fVerbose {
                Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
            if !p.pPars.fSilent {
                Abc_Print( 1, "Reached gap timeout (%d seconds) in frame %d.\n",  p.pPars.nTimeOutGap, iFrame ); }
            p.pPars.iFrame = iFrame;
            return -1;
        }
        if p.pPars.nFrameMax && iFrame >= p.pPars.nFrameMax {
            if p.pPars.fVerbose {
                Pdr_ManPrintProgress( p, 1, Abc_Clock() - clkStart ); }
            if !p.pPars.fSilent {
                Abc_Print( 1, "Reached limit on the number of timeframes (%d).\n", p.pPars.nFrameMax ); }
            p.pPars.iFrame = iFrame;
            return -1;
        }
    }
    assert(false, "0");
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func IPdr_ManSolve(pAig *Aig_Man_t, pPars *Pdr_Par_t) int32 {
    var p *Pdr_Man_t;
    var k, RetValue int32;
    var vClausesSaved *Vec_Vec_t;

    var clk abctime  = Abc_Clock();
    if pPars.nTimeOutOne && !pPars.fSolveAll {
        pPars.nTimeOutOne = 0; }
    if pPars.nTimeOutOne && pPars.nTimeOut == 0 {
        pPars.nTimeOut = pPars.nTimeOutOne * Saig_ManPoNum(pAig) / 1000 + int32(((pPars.nTimeOutOne * Saig_ManPoNum(pAig) % 1000) > 0)); }
    if pPars.fVerbose {
//    Abc_Print( 1, "Running PDR by Niklas Een (aka IC3 by Aaron Bradley) with these parameters:\n" );
        Abc_Print( 1, "VarMax = %d. FrameMax = %d. QueMax = %d. TimeMax = %d. ",
            pPars.nRecycle,
            pPars.nFrameMax,
            pPars.nRestLimit,
            pPars.nTimeOut );
        Abc_Print( 1, "MonoCNF = %s. SkipGen = %s. SolveAll = %s.\n",
            pPars.fMonoCnf ?     "yes" : "no",
            pPars.fSkipGeneral ? "yes" : "no",
            pPars.fSolveAll ?    "yes" : "no" );
    }
    ABC_FREE( pAig.pSeqModel );


    p = Pdr_ManStart( pAig, pPars, nil );
    for {
        RetValue = IPdr_ManSolveInt( p, 1, 0 );

        if RetValue == -1 && pPars.iFrame == pPars.nFrameMax {
            vClausesSaved = IPdr_ManSaveClauses( p, 1 );

            Pdr_ManStop( p );

            p = Pdr_ManStart( pAig, pPars, nil );
            IPdr_ManRestoreClauses( p, vClausesSaved, nil );

            pPars.nFrameMax = pPars.nFrameMax << 1;

            continue;
        }

        if RetValue == 0 {
            assert(pAig.pSeqModel != nil || p.vCexes != nil, "pAig.pSeqModel != nil || p.vCexes != nil"); }
        if p.vCexes {
            assert(p.pAig.vSeqModelVec == nil, "p.pAig.vSeqModelVec == nil");
            p.pAig.vSeqModelVec = p.vCexes;
            p.vCexes = nil;
        }
        if p.pPars.fDumpInv {
            var pFileName string;
            if pPars.pInvFileName {
                pFileName = pPars.pInvFileName;
            } else {
                pFileName = Extra_FileNameGenericAppend(p.pAig.pName, "_inv.pla");
            }
                Abc_FrameSetInv( Pdr_ManDeriveInfinityClauses( p, RetValue!=1 ) );
                Pdr_ManDumpClauses( p, pFileName, RetValue==1 );
        
        } else if RetValue == 1 {
            Abc_FrameSetInv( Pdr_ManDeriveInfinityClauses( p, RetValue!=1 ) ); }

        p.tTotal += Abc_Clock() - clk;
        Pdr_ManStop( p );

        break;
    }
    
    
    pPars.iFrame--;
    // convert all -2 (unknown) entries into -1 (undec)
    if pPars.vOutMap {
        for  k = 0; k < Saig_ManPoNum(pAig); k++  {
            if Vec_IntEntry(pPars.vOutMap, k) == -2 { // unknown
                Vec_IntWriteEntry( pPars.vOutMap, k, -1 ); }
        } } // undec
    if pPars.fUseBridge {
        Gia_ManToBridgeAbort( os.Stdout, 7, (unsigned char *)"timeout" ); }
    return RetValue;
}

func IPdr_ManCheckCombUnsat(p *Pdr_Man_t) int32 {
    var iFrame int32;
    var RetValue int32 = -1;

    Pdr_ManCreateSolver( p, (iFrame = 0) );
    Pdr_ManCreateSolver( p, (iFrame = 1) );

    p.nFrames = iFrame;
    p.iUseFrame = Abc_MaxInt(iFrame, 1);

    RetValue = Pdr_ManCheckCube( p, iFrame, nil, nil, p.pPars.nConfLimit, 0, 1 );
    return RetValue;
}

func IPdr_ManCheckCubeReduce(p *Pdr_Man_t, vClauses *Vec_Ptr_t, pCube *Pdr_Set_t, nConfLimit int32) int32 { 
    var pSat *sat_solver;
    var vLits, vLitsA *Vec_Int_t;
    var Lit int32;
    var RetValue int32 = l_True;
    var i int32;
    var pCla *Pdr_Set_t;
    var iActVar int32;
    var clk abctime  = Abc_Clock();

    pSat = Pdr_ManSolver( p, 1 );

    if pCube == nil { // solve the property
    
        Lit = Abc_Var2Lit( Pdr_ObjSatVar(p, 1, 2, Aig_ManCo(p.pAig, p.iOutCur)), 0 ); // pos literal (property fails)
        RetValue = sat_solver_addclause( pSat, &Lit, &Lit+1 );
        assert(RetValue == 1, "RetValue == 1");

        vLitsA = Vec_IntStart( Vec_PtrSize( vClauses ) );
        iActVar = Pdr_ManFreeVar( p, 1 );
        for  i = 1; i < Vec_PtrSize( vClauses ); ++i  {
            Pdr_ManFreeVar( p, 1 );
        }
        Vec_PtrForEachEntry( Pdr_Set_t *, vClauses, pCla, i )
        {
            vLits = Pdr_ManCubeToLits( p, 1, pCla, 1, 0 );
            Lit = Abc_Var2Lit( iActVar + i, 1 );
            Vec_IntPush( vLits, Lit );

            RetValue = sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );
            assert(RetValue == 1, "RetValue == 1");
            Vec_IntWriteEntry( vLitsA, i, Abc_Var2Lit( iActVar + i, 0 )  );
        }
        sat_solver_compress( pSat );

        // solve 
        RetValue = sat_solver_solve( pSat, Vec_IntArray(vLitsA), Vec_IntArray(vLitsA) + Vec_IntSize(vLitsA), nConfLimit, 0, 0, 0 );
        Vec_IntFree( vLitsA );

        if RetValue == l_Undef {
            return -1; }
    }

    assert(RetValue != l_Undef, "RetValue != l_Undef");
    if RetValue == l_False { // UNSAT
    
        var ncorelits int32;
        var pcorelits *int32;
        var vTemp *Vec_Ptr_t  = nil;
        var vMark *Vec_Bit_t  = nil;

        ncorelits = sat_solver_final(pSat, &pcorelits);
        Abc_Print( 1, "UNSAT at the last frame. nCores = %d (out of %d).", ncorelits, Vec_PtrSize( vClauses ) );
        Abc_PrintTime( 1, "    Time", Abc_Clock() - clk );

        vTemp = Vec_PtrDup( vClauses );
        vMark = Vec_BitStart( Vec_PtrSize( vClauses) );
        Vec_PtrClear( vClauses );
        for  i = 0; i < ncorelits; ++i  {
            //Abc_Print( 1, "Core[%d] = lit(%d) = var(%d) = %d-th set\n", i, pcorelits[i], Abc_Lit2Var(pcorelits[i]), Abc_Lit2Var(pcorelits[i]) - iActVar );
            Vec_BitWriteEntry( vMark, Abc_Lit2Var( pcorelits[i] ) - iActVar, 1 );
        }
        
        Vec_PtrForEachEntry( Pdr_Set_t *, vTemp, pCla, i )
        {
            if Vec_BitEntry( vMark, i ) { 
                Vec_PtrPush( vClauses, pCla );
                continue;
            }
            Pdr_SetDeref( pCla );
        }

        Vec_PtrFree( vTemp );
        Vec_BitFree( vMark );
        RetValue = 1;
    
    } else { // SAT
    
        Abc_Print( 1, "SAT at the last frame." );
        Abc_PrintTime( 1, "    Time", Abc_Clock() - clk );
        RetValue = 0;
    }

    return RetValue;
}

func IPdr_ManReduceClauses(p *Pdr_Man_t, vClauses *Vec_Vec_t) int32 {
    var iFrame int32;
    var RetValue int32 = -1;
    var vLast *Vec_Ptr_t  = nil;

    Pdr_ManCreateSolver( p, (iFrame = 0) );
    Pdr_ManCreateSolver( p, (iFrame = 1) );

    p.nFrames = iFrame;
    p.iUseFrame = Abc_MaxInt(iFrame, 1);

    vLast = Vec_VecEntry( vClauses, Vec_VecSize( vClauses ) - 1 );
    RetValue = IPdr_ManCheckCubeReduce( p, vLast, nil, p.pPars.nConfLimit );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarIPdr(pNtk *Abc_Ntk_t, pPars *Pdr_Par_t) int32 {
    var RetValue int32  = -1;
    var clk abctime  = Abc_Clock();
    var pMan *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );

    RetValue = IPdr_ManSolve( pMan, pPars );
 
    if RetValue == 1 {
        Abc_Print( 1, "Property proved.  " );
    } else {
        if RetValue == 0 {
            if pMan.pSeqModel == nil {
                Abc_Print( 1, "Counter-example is not available.\n" );
            } else {
                Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", pMan.pSeqModel.iPo, pNtk.pName, pMan.pSeqModel.iFrame );
                if !Saig_ManVerifyCex( pMan, pMan.pSeqModel ) {
                    Abc_Print( 1, "Counter-example verification has FAILED.\n" ); }
            }
        
        } else if RetValue == -1 {
            Abc_Print( 1, "Property UNDECIDED.  " );
        } else {
            assert(false, "0"); }
    }
    ABC_PRT( "Time", Abc_Clock() - clk );


    ABC_FREE( pNtk.pSeqModel );
    pNtk.pSeqModel = pMan.pSeqModel;
    pMan.pSeqModel = nil;
    if pNtk.vSeqModelVec {
        Vec_PtrFreeFree( pNtk.vSeqModelVec ); }
    pNtk.vSeqModelVec = pMan.vSeqModelVec;
    pMan.vSeqModelVec = nil;
    Aig_ManStop( pMan );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
