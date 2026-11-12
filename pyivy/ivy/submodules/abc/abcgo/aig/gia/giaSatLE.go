package gia

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSatLE.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Mapping with edges.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatLE.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"
//#include "misc/tim/tim.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Cut printing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// returns 1 if Cut can represent LUT (Cut is equal or is contained in LUT)
/**Function*************************************************************

  Synopsis    [Derive CNF.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Add edge compatibility constraints.]

  Description [Returns 1 if constraints have been added.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sle_ManExplore(pGia *Gia_Man_t, nBTLimit int32, DelayInit int32, fDynamic int32, fTwoEdges int32, fVerbose int32) {
    var fVeryVerbose int32;
    var clk abctime  = Abc_Clock();
    var vEdges2 *Vec_Int_t   = Vec_IntAlloc(1000);
    var vMapping *Vec_Int_t  = Vec_IntAlloc(1000);
    var i, iLut, nConfs, status, Delay, iFirstVar int32;
    var DelayStart int32;
    if (DelayInit != 0) || !Gia_ManHasMapping(pGia) {
        DelayStart = DelayInit;
    } else {
        DelayStart = Gia_ManLutLevel(pGia, nil);
    }
    var p *Sle_Man_t  = Sle_ManAlloc( pGia, DelayStart, fVerbose );
    if fVerbose != 0 {
        fmt.Printf("Running solver with %d conflicts, %d initial delay, and %d edges. Dynamic constraints = %s.\n", nBTLimit, DelayInit, 1+fTwoEdges, fDynamic?"yes":"no"); }
    Sle_ManMarkupVariables( p );
    if fVerbose != 0 {
        fmt.Printf("Vars:  Total = %d.  Node = %d. Cut = %d. Edge = %d. Delay = %d.\n", p.nVarsTotal, p.nNodeVars, p.nCutVars, p.nEdgeVars, p.nDelayVars); }
    Sle_ManDeriveInit( p );
    Sle_ManDeriveCnf( p, nBTLimit, fDynamic || fTwoEdges );
    if fVerbose != 0 {
        fmt.Printf("Clas:  Total = %d.  Cut = %d. Edge = %d. EdgeEx = %d. Delay = %d.\n", sat_solver_nclauses(p.pSat), p.nCutClas, p.nEdgeClas, p.nEdgeClas2, p.nDelayClas); }
    //Sat_SolverWriteDimacs( p->pSat, "temp.cnf", NULL, NULL, 0 );
    for  Delay = p.nLevels; Delay >= 0; Delay--  {
        // we constrain COs, although it would be fine to constrain only POs
        if Delay < p.nLevels {
            // inlining of macro Gia_ManForEachCoDriverId
            for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((iLut) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
                if Vec_BitEntry(p.vMask, iLut) { // internal node
                
                    iFirstVar = Vec_IntEntry( p.vDelayFirst, iLut );
                    if !sat_solver_push(p.pSat, Abc_Var2Lit(iFirstVar + Delay, 1)) {
                        break; }
                }
            }
            if i < Gia_ManCoNum(p.pGia) {
                if fVerbose != 0 {
                    fmt.Printf("Proved UNSAT for delay %d.  ", Delay);
                    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
                }
                break;
            }
        }
        // solve with assumptions
        //clk = Abc_Clock();
        nConfs = sat_solver_nconflicts( p.pSat );
        for {
            p.nSatCalls++;
            status = sat_solver_solve_internal( p.pSat );
            if status != l_True {
                break; }
            if !(Sle_ManAddEdgeConstraints(p, 1+fTwoEdges) != 0) {
                break; }
        }
        nConfs = sat_solver_nconflicts( p.pSat ) - nConfs;
        if status == l_True {
            if fVerbose != 0 {
                var nNodes, nEdges int32;
                for  i = 0; i < p.nNodeVars; i++  {
                    nNodes += sat_solver_var_value(p.pSat, i);
                }
                for  i = 0; i < p.nEdgeVars; i++  {
                    nEdges += sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + i);
                }
                fmt.Printf("Solution with delay %2d, node count %5d, and edge count %5d exists. Conf = %8d.  ", Delay, nNodes, nEdges, nConfs);
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            // save the result
            Sle_ManDeriveResult( p, vEdges2, vMapping );
            if fVeryVerbose != 0 {
                fmt.Printf("Nodes:  ");
                for  i = 0; i < p.nNodeVars; i++  {
                    if sat_solver_var_value(p.pSat, i) {
                        fmt.Printf("%d ", i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vCutFirst );
                fmt.Printf("Cuts:   ");
                for  i = 0; i < p.nCutVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vEdgeFirst );
                fmt.Printf("Edges:  ");
                for  i = 0; i < p.nEdgeVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + p.nCutVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");

                Vec_IntPrint( p.vDelayFirst );
                fmt.Printf("Delays: ");
                for  i = 0; i < p.nDelayVars; i++  {
                    if sat_solver_var_value(p.pSat, p.nNodeVars + p.nCutVars + p.nEdgeVars + i) {
                        fmt.Printf("%d ", p.nNodeVars + p.nCutVars + p.nEdgeVars + i); }
                }
                fmt.Printf("\n");
                fmt.Printf("\n");
            }
        
        } else {
            if fVerbose != 0 {
                if status == l_False {
                    fmt.Printf("Proved UNSAT for delay %d. Conf = %8d.  ", Delay, nConfs);
                } else {
                    fmt.Printf("Resource limit reached for delay %d. Conf = %8d.  ", Delay, nConfs); }
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            break;
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Clas:  Total = %d.  Cut = %d. Edge = %d. EdgeEx = %d. Delay = %d.  Calls = %d.\n", sat_solver_nclauses(p.pSat), p.nCutClas, p.nEdgeClas, p.nEdgeClas2, p.nDelayClas, p.nSatCalls); }
    if Vec_IntSize(vMapping) > 0 {
        Gia_ManEdgeFromArray( p.pGia, vEdges2 );
        Vec_IntFree( vEdges2 );
        Vec_IntFreeP( &p.pGia.vMapping );
        p.pGia.vMapping = vMapping;
    
    } else {
        Vec_IntFree( vEdges2 );
        Vec_IntFree( vMapping );
    }
    Vec_IntFreeP( &p.pGia.vPacking );
    Sle_ManStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

