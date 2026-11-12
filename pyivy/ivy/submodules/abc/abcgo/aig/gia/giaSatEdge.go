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
func Seg_ManCreateCnf(p *Seg_Man_t, fTwo int32, fVerbose int32) {
    var pTim *Tim_Man_t  = (Tim_Man_t *)p.pGia.pManTime;
    var pObj *Gia_Obj_t;
    var vObjEdges *Vec_Wec_t;
    var vLevel *Vec_Int_t;
    var iLut, iFanin, iFirst int32;
    var pLits [3]int32;
    var Count int32;
    var i, k, nVars, Edge, value int32;
    var clk abctime  = Abc_Clock();
    // edge delay constraints
    var nConstr int32  = sat_solver_nclauses(p.pSat);
    // inlining of macro Gia_ManForEachObj
    for  iLut = 0; (iLut < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, iLut)); iLut++  {
        var iFirstLut int32  = Vec_IntEntry( p.vFirsts, iLut );
        var nVarsLut int32   = Vec_IntEntry( p.vNvars, iLut );
        if (pTim != nil) && Gia_ObjIsCi(pObj) {
            var iBox int32  = Tim_ManBoxForCi( pTim, Gia_ObjCioId(pObj) );
            if nVarsLut > 0 && iBox >= 0 {
                var iCiId int32  = Tim_ManBoxOutputFirst( pTim, iBox );
                if iCiId == Gia_ObjCioId(pObj) { // first input
                
                    var nIns int32  = Tim_ManBoxInputNum( pTim, iBox );
                    var iIn0 int32  = Tim_ManBoxInputFirst( pTim, iBox );
                    for  i = 0; i < nIns-1; i++  { // skip carry-in pin
                    
                        var pOut *Gia_Obj_t  = Gia_ManCo( p.pGia, iIn0+i );
                        var iDriverId int32  = Gia_ObjFaninId0p( p.pGia, pOut );
                        var AddOn int32;

                        iFirst = Vec_IntEntry( p.vFirsts, iDriverId );
                        nVars  = Vec_IntEntry( p.vNvars, iDriverId );
                        assert(nVars < nVarsLut, "nVars < nVarsLut");
                        AddOn = int32((nVars < nVarsLut));
                        for  k = 0; k < nVars; k++  {
                            pLits[0] = Abc_Var2Lit( iFirst+k, 1 );
                            pLits[1] = Abc_Var2Lit( iFirstLut+k+AddOn, 0 );
                            value = sat_solver_addclause( p.pSat, pLits, pLits+2 );
                            assert((value != 0), "value");
                        }
                    }
                
                } else { // intermediate input
                
                    var pIn *Gia_Obj_t  = Gia_ManCi( p.pGia, iCiId );
                    var iObjId int32  = Gia_ObjId( p.pGia, pIn );

                    iFirst = Vec_IntEntry( p.vFirsts, iObjId );
                    nVars  = Vec_IntEntry( p.vNvars, iObjId );
                    if nVars > 0 {
                        for  k = 0; k < nVars; k++  {
                            pLits[0] = Abc_Var2Lit( iFirst+k, 1 );
                            pLits[1] = Abc_Var2Lit( iFirstLut+k, 0 );
                            value = sat_solver_addclause( p.pSat, pLits, pLits+2 );
                            assert((value != 0), "value");
                        }
                    }
                }
            }
            continue;
        }
        if !Gia_ObjIsLut(p.pGia, iLut) {
            continue; }
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p.pGia,iLut) && ((iFanin = Gia_ObjLutFanins(p.pGia,iLut)[i]),1); i++  {
            if (pTim != nil) && Gia_ObjIsCi(Gia_ManObj(p.pGia, iFanin)) {
                iFirst = Vec_IntEntry( p.vFirsts, iFanin );
                nVars  = Vec_IntEntry( p.vNvars, iFanin );
                assert(nVars <= nVarsLut, "nVars <= nVarsLut");
                if nVars > 0 {
                    for  k = 0; k < nVars; k++  {
                        pLits[0] = Abc_Var2Lit( iFirst+k, 1 );
                        pLits[1] = Abc_Var2Lit( iFirstLut+k, 0 );
                        value = sat_solver_addclause( p.pSat, pLits, pLits+2 );
                        assert((value != 0), "value");
                    }
                }
            
            } else if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iFanin)) {
                iFirst = Vec_IntEntry( p.vFirsts, iFanin );
                nVars  = Vec_IntEntry( p.vNvars, iFanin );
                assert(nVars != 1 && nVars < nVarsLut, "nVars != 1 && nVars < nVarsLut");
                // add initial
                if nVars == 0 {
                    pLits[0] = Abc_Var2Lit( Count, 1 );
                    pLits[1] = Abc_Var2Lit( iFirstLut, 0 );
                    value = sat_solver_addclause( p.pSat, pLits, pLits+2 );
                    assert((value != 0), "value");

                    pLits[0] = Abc_Var2Lit( Count, 0 );
                    pLits[1] = Abc_Var2Lit( iFirstLut+1, 0 );
                    value = sat_solver_addclause( p.pSat, pLits, pLits+2 );
                    assert((value != 0), "value");
                }
                // add others
                for  k = 0; k < nVars; k++  {
                    pLits[0] = Abc_Var2Lit( iFirst+k, 1 );
                    pLits[1] = Abc_Var2Lit( Count, 1 );
                    pLits[2] = Abc_Var2Lit( iFirstLut+k, 0 );
                    value = sat_solver_addclause( p.pSat, pLits, pLits+3 );
                    assert((value != 0), "value");

                    pLits[0] = Abc_Var2Lit( iFirst+k, 1 );
                    pLits[1] = Abc_Var2Lit( Count, 0 );
                    pLits[2] = Abc_Var2Lit( iFirstLut+k+1, 0 );
                    value = sat_solver_addclause( p.pSat, pLits, pLits+3 );
                    assert((value != 0), "value");
                }
                Count++;
            }
        }
    }
    assert(Count == p.nVars, "Count == p.nVars");
    if fVerbose != 0 {
    fmt.Printf("Delay constraints = %d. ", sat_solver_nclauses(p.pSat) - nConstr); }
    nConstr = sat_solver_nclauses(p.pSat);
/*
    // delay relationship constraints
    Vec_IntForEachEntryTwo( p->vFirsts, p->vNvars, iFirst, nVars, iLut )
    {
        if ( nVars < 2 )
            continue;
        for ( i = 1; i < nVars; i++ )
        {
            pLits[0] = Abc_Var2Lit( iFirst + i - 1, 1 );
            pLits[1] = Abc_Var2Lit( iFirst + i,     0 );
            value = sat_solver_addclause( p->pSat, pLits, pLits+2 );
            assert( value );
        }
    }
*/
    // edge compatibility constraint
    vObjEdges = Seg_ManCollectObjEdges( p.vEdges, Gia_ManObjNum(p.pGia) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vObjEdges)) && (((vLevel) = Vec_WecEntry(vObjEdges, i)), 1); i++  {
        var v1, v2, v3, Var1, Var2, Var3 int32;
        if (!(fTwo != 0) && Vec_IntSize(vLevel) >= 2) || ((fTwo != 0) && Vec_IntSize(vLevel) > 10) {
            // inlining of macro Vec_IntForEachEntry
            for  v1 = 0; (v1 < Vec_IntSize(vLevel)) && (((Var1) = Vec_IntEntry(vLevel, v1)), 1); v1++  {
            // inlining of macro Vec_IntForEachEntryStart
            for  v2 = v1 + 1; (v2 < Vec_IntSize(vLevel)) && (((Var2) = Vec_IntEntry(vLevel, v2)), 1); v2++  {
                pLits[0] = Abc_Var2Lit( Var1, 1 );
                pLits[1] = Abc_Var2Lit( Var2, 1 );
                value = sat_solver_addclause( p.pSat, pLits, pLits+2 );
                assert((value != 0), "value");
            }
            }
        
        } else if (fTwo != 0) && Vec_IntSize(vLevel) >= 3 {
            // inlining of macro Vec_IntForEachEntry
            for  v1 = 0; (v1 < Vec_IntSize(vLevel)) && (((Var1) = Vec_IntEntry(vLevel, v1)), 1); v1++  {
            // inlining of macro Vec_IntForEachEntryStart
            for  v2 = v1 + 1; (v2 < Vec_IntSize(vLevel)) && (((Var2) = Vec_IntEntry(vLevel, v2)), 1); v2++  {
            // inlining of macro Vec_IntForEachEntryStart
            for  v3 = v2 + 1; (v3 < Vec_IntSize(vLevel)) && (((Var3) = Vec_IntEntry(vLevel, v3)), 1); v3++  {
                pLits[0] = Abc_Var2Lit( Var1, 1 );
                pLits[1] = Abc_Var2Lit( Var2, 1 );
                pLits[2] = Abc_Var2Lit( Var3, 1 );
                value = sat_solver_addclause( p.pSat, pLits, pLits+3 );
                assert((value != 0), "value");
            }
            }
            }
        }
    }
    Vec_WecFree( vObjEdges );
    // block forbidden edges
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vToSkip)) && (((Edge) = Vec_IntEntry(p.vToSkip, i)), 1); i++  {
        pLits[0] = Abc_Var2Lit( Edge, 1 );
        value = sat_solver_addclause( p.pSat, pLits, pLits+1 );
        assert((value != 0), "value");
    }
    if fVerbose != 0 {
    fmt.Printf("Edge constraints = %d. ", sat_solver_nclauses(p.pSat) - nConstr); }
    if fVerbose != 0 {
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
}

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

