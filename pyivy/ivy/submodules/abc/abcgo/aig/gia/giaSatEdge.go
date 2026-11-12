package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSatEdge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatEdge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/tim/tim.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
/**Function*************************************************************

  Synopsis    [Count the number of edges between internal nodes.]

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
func Seg_ManConvertResult(p *Seg_Man_t) *Vec_Int_t {
    var iFanin, iObj, i int32;
    var vEdges2 *Vec_Int_t  = Vec_IntAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(p.vEdges)) && (((iFanin) = Vec_IntEntry(p.vEdges, i)), 1) && (((iObj) = Vec_IntEntry(p.vEdges, i+1)), 1); i += 2  {
        if sat_solver_var_value(p.pSat, i/2) {
            Vec_IntPushTwo( vEdges2, iFanin, iObj ); }
    }
    return vEdges2;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Seg_ManComputeDelay(pGia *Gia_Man_t, DelayInit int32, nFanouts int32, fTwo int32, fVerbose int32) {
    var nBTLimit int32;
    var nTimeOut int32;
    var fVeryVerbose int32;

    var pObj *Gia_Obj_t;
    var clk abctime  = Abc_Clock();
    var vEdges2 *Vec_Int_t  = nil;
    var i, iLut, iFirst, nVars, status, Delay, nConfs int32;
    var p *Seg_Man_t  = Seg_ManAlloc( pGia, nFanouts );
    var DelayStart int32;
    if DelayInit != 0 {
        DelayStart = DelayInit;
    } else {
        DelayStart = p.DelayMax;
    }

    if fVerbose != 0 {
    fmt.Printf("Running SatEdge with starting delay %d and edge %d (edge vars %d, total vars %d)\n", DelayStart, fTwo+1, p.FirstVar, sat_solver_nvars(p.pSat)); }
    Seg_ManCreateCnf( p, fTwo, fVerbose );
    //Sat_SolverWriteDimacs( p->pSat, "test_edge.cnf", NULL, NULL, 0 );
    // set resource limits
    sat_solver_set_resource_limits( p.pSat, nBTLimit, 0, 0, 0 );
    sat_solver_set_runtime_limit( p.pSat, nTimeOut ? nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
    sat_solver_set_random( p.pSat, 1 );
    sat_solver_set_polarity( p.pSat, Vec_IntArray(p.vPolars), Vec_IntSize(p.vPolars) );
    //sat_solver_set_var_activity( p->pSat, NULL, p->nVars );
    // increment delay gradually
    for  Delay = p.DelayMax; Delay >= 0; Delay--  {
        // we constrain COs, although it would be fine to constrain only POs
        // inlining of macro Gia_ManForEachCoDriver
        for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ObjFanin0(Gia_ManCo(p.pGia, i))); i++  {
            iLut   = Gia_ObjId( p.pGia, pObj );
            iFirst = Vec_IntEntry( p.vFirsts, iLut );
            nVars  = Vec_IntEntry( p.vNvars, iLut );
            if Delay < nVars && !sat_solver_push(p.pSat, Abc_Var2Lit(iFirst + Delay, 1)) {
                break; }
        }
        if i < Gia_ManCoNum(p.pGia) {
            fmt.Printf("Proved UNSAT for delay %d.  ", Delay);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            break;
        }
        if Delay > DelayStart {
            continue; }
        // solve with assumptions
        //clk = Abc_Clock();
        nConfs = sat_solver_nconflicts( p.pSat );
        status = sat_solver_solve_internal( p.pSat );
        nConfs = sat_solver_nconflicts( p.pSat ) - nConfs;
        if status == l_True {
            if fVerbose != 0 {
                var Count int32;
                for  i = 0; i < p.nVars; i++  {
                    Count += sat_solver_var_value(p.pSat, i);
                }
                fmt.Printf("Solution with delay %2d and %5d edges exists. Conf = %8d.  ", Delay, Count, nConfs);
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            // save the result
            Vec_IntFreeP( &vEdges2 );
            vEdges2 = Seg_ManConvertResult( p );
            if fVeryVerbose != 0 {
                for  i = 0; i < p.nVars; i++  {
                    fmt.Printf("%d=%d ", i, sat_solver_var_value(p.pSat, i));
                }
                fmt.Printf("   ");

                for  i = p.nVars; i < sat_solver_nvars(p.pSat); i++  {
                    fmt.Printf("%d=%d ", i, sat_solver_var_value(p.pSat, i));
                }
                fmt.Printf("\n");
            }
        
        } else {
            if fVerbose != 0 {
                if status == l_False {
                    fmt.Printf("Proved UNSAT for delay %d.  ", Delay);
                } else {
                    fmt.Printf("Resource limit reached for delay %d.  ", Delay); }
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            break;
        }
    }
    Gia_ManEdgeFromArray( p.pGia, vEdges2 );
    Vec_IntFreeP( &vEdges2 );
    Seg_ManStop( p );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

