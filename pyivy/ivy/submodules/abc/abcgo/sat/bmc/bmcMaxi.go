package bmc

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [bmcMaxi.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcMaxi.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "aig/gia/giaAig.h"

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
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMaxiUnfold(p *Gia_Man_t, nFrames int32, fUseVars int32, vInit *Vec_Int_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, f int32;
    pNew = Gia_ManStart( fUseVars * 2 * Gia_ManRegNum(p) + nFrames * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // control/data variables
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ManAppendCi( pNew );
    }
    // build timeframes
    assert(!(vInit != nil) || Vec_IntSize(vInit) == Gia_ManRegNum(p), "!vInit || Vec_IntSize(vInit) == Gia_ManRegNum(p)");
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        var Value int32  = Vec_IntEntry( vInit, i );
        var iCtrl int32  = Gia_ManCiLit( pNew, i );
        var iData int32  = Gia_ManCiLit( pNew, Gia_ManRegNum(p)+i );
        // decide based on Value
        if Value == 0 {
            if fUseVars != 0 {
                pObj.Value = Gia_ManHashAnd(pNew, iCtrl, iData);
            } else {
                pObj.Value = 0;
            }
        } else if Value == 1 {
            if fUseVars != 0 {
                pObj.Value = Gia_ManHashOr(pNew, Abc_LitNot(iCtrl), iData);
            } else {
                pObj.Value = 1;
            }
        } else if Value == 2 {
            pObj.Value = Gia_ManHashAnd(pNew, iCtrl, iData);
        } else if Value == 3 {
            pObj.Value = Gia_ManHashOr(pNew, Abc_LitNot(iCtrl), iData);
        } else if Value == 4 {
            pObj.Value = 0;
        } else if Value == 5 {
            pObj.Value = 1;
        } else { assert(false, "0"); }
    }
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        }
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            pObj.Value = Gia_ObjRoToRi(p, pObj).Value;
        }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, pObj.Value );
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNew) == 2 * Gia_ManRegNum(p) + nFrames * Gia_ManPiNum(p), "Gia_ManPiNum(pNew) == 2 * Gia_ManRegNum(p) + nFrames * Gia_ManPiNum(p)");
    return pNew;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMaxiPerform(p *Gia_Man_t, vInit *Vec_Int_t, nFrames int32, nTimeOut int32, fVerbose int32) *Vec_Int_t {
    var nIterMax int32  = 1000000;
    var i, iLit, Iter, status int32;
    var nLits int32;
    var pLits *int32;
    var clkTotal abctime  = Abc_Clock();
    var clkSat abctime;
    var vLits, vMap *Vec_Int_t;
    var pSat *sat_solver;
    var pObj *Gia_Obj_t;
    var p0 *Gia_Man_t  = Gia_ManMaxiUnfold( p, nFrames, 0, vInit );
    var p1 *Gia_Man_t  = Gia_ManMaxiUnfold( p, nFrames, 1, vInit );
    var pM *Gia_Man_t  = Gia_ManMiter( p0, p1, 0, 0, 0, 0, 0 );
    var pCnf *Cnf_Dat_t  = Cnf_DeriveGiaRemapped( pM );
    Gia_ManStop( p0 );
    Gia_ManStop( p1 );
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    if fVerbose != 0 {
        fmt.Printf("Running with %d frames and %sgiven init state.\n", nFrames, vInit ? "":"no "); }

    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, pCnf.nVars );
    sat_solver_set_runtime_limit( pSat, nTimeOut ? nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
    for  i = 0; i < pCnf.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
            assert(false, "0"); }
    }

    // add one large OR clause
    vLits = Vec_IntAlloc( Gia_ManCoNum(p) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pM.vCos)) && ((pObj) = Gia_ManCo(pM, i)); i++  {
        Vec_IntPush( vLits, Abc_Var2Lit(pCnf.pVarNums[Gia_ObjId(pM, pObj)], 0) );
    }
    sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );

    // create assumptions
    Vec_IntClear( vLits );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(pM)) && ((pObj) = Gia_ManCi(pM, i)); i++  {
        if i == Gia_ManRegNum(p) {
            break;
        } else if Vec_IntEntry(vInit, i) == 0 || Vec_IntEntry(vInit, i) == 1 {
            Vec_IntPush( vLits, Abc_Var2Lit(pCnf.pVarNums[Gia_ObjId(pM, pObj)], 1) ); }
    }

    if fVerbose != 0 {
        fmt.Printf("Iter%6d : ", 0);
        fmt.Printf("Var =%10d  ", sat_solver_nvars(pSat));
        fmt.Printf("Clause =%10d  ", sat_solver_nclauses(pSat));
        fmt.Printf("Conflict =%10d  ", sat_solver_nconflicts(pSat));
        fmt.Printf("Subset =%6d  ", Vec_IntSize(vLits));
        Abc_PrintTime( 1, "Time", clkSat );
//      ABC_PRTr( "Solver time", clkSat );
    }
    for  Iter = 0; Iter < nIterMax; Iter++  {
        var clk abctime  = Abc_Clock();
        status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        clkSat += Abc_Clock() - clk;
        if status == l_Undef {
//            if ( fVerbose )
//                printf( "\n" );
            fmt.Printf("Timeout reached after %d seconds and %d iterations.  ", nTimeOut, Iter);
            break;
        }
        if status == l_True {
//            if ( fVerbose )
//                printf( "\n" );
            fmt.Printf("The problem is SAT after %d iterations.  ", Iter);
            break;
        }
        assert(status == l_False, "status == l_False");
        nLits = sat_solver_final( pSat, &pLits );
        if fVerbose != 0 {
            fmt.Printf("Iter%6d : ", Iter+1);
            fmt.Printf("Var =%10d  ", sat_solver_nvars(pSat));
            fmt.Printf("Clause =%10d  ", sat_solver_nclauses(pSat));
            fmt.Printf("Conflict =%10d  ", sat_solver_nconflicts(pSat));
            fmt.Printf("Subset =%6d  ", nLits);
            Abc_PrintTime( 1, "Time", clkSat );
//            ABC_PRTr( "Solver time", clkSat );
        }
        if Vec_IntSize(vLits) == nLits {
//            if ( fVerbose )
//                printf( "\n" );
            fmt.Printf("Reached fixed point with %d entries after %d iterations.  ", Vec_IntSize(vLits), Iter+1);
            break;
        }
        // collect used literals
        Vec_IntClear( vLits );
        for  i = 0; i < nLits; i++  {
            Vec_IntPush( vLits, Abc_LitNot(pLits[i]) );
        }
    }
    // create map
    vMap = Vec_IntStart( pCnf.nVars );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, i)), 1); i++  {
        Vec_IntWriteEntry( vMap, Abc_Lit2Var(iLit), 1 );
    }

    // create output
    Vec_IntFree( vLits );
    vLits = Vec_IntDup(vInit);
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(pM)) && ((pObj) = Gia_ManCi(pM, i)); i++  {
        if i == Gia_ManRegNum(p) {
            break;
        } else if Vec_IntEntry(vLits, i) == 4 || Vec_IntEntry(vLits, i) == 5 {
            Vec_IntWriteEntry( vLits, i, Vec_IntEntry(vLits, i) );
        } else if (Vec_IntEntry(vLits, i) == 0 || Vec_IntEntry(vLits, i) == 1) && !Vec_IntEntry(vMap, pCnf.pVarNums[Gia_ObjId(pM, pObj)]) {
            Vec_IntWriteEntry( vLits, i, Vec_IntEntry(vLits, i) | 2 ); }
    }
    Vec_IntFree( vMap );

    // cleanup
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Gia_ManStop( pM );
    Abc_PrintTime( 1, "Total runtime", Abc_Clock() - clkTotal );
    return vLits;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMaxiTest(p *Gia_Man_t, vInit0 *Vec_Int_t, nFrames int32, nWords int32, nTimeOut int32, fSim int32, fVerbose int32) *Vec_Int_t {
    var vRes, vInit *Vec_Int_t;
    if vInit0 != nil {
        vInit = vInit0;
    } else {
        vInit = Vec_IntStart( Gia_ManRegNum(p) );
    }
    vRes = Gia_ManMaxiPerform( p, vInit, nFrames, nTimeOut, fVerbose );
    if vInit != vInit0 {
        Vec_IntFree( vInit ); }
    return vRes;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

