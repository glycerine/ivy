package int2

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [int2Core.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 1, 2013.]

  Revision    [$Id: int2Core.c,v 1.00 2013/12/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "int2Int.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default values of interpolation parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManSetDefaultParams(p *Int2_ManPars_t) { 
    memset( p, 0, sizeof(Int2_ManPars_t) );
    p.nBTLimit      =  0;     // limit on the number of conflicts
    p.nFramesS      =  1;     // the starting number timeframes
    p.nFramesMax    =  0;     // the max number timeframes to unroll
    p.nSecLimit     =  0;     // time limit in seconds
    p.nFramesK      =  1;     // the number of timeframes to use in induction
    p.fRewrite      =  0;     // use additional rewriting to simplify timeframes
    p.fTransLoop    =  0;     // add transition into the init state under new PI var
    p.fDropInvar    =  0;     // dump inductive invariant into file
    p.fVerbose      =  0;     // print verbose statistics
    p.iFrameMax     = -1;
}
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManUnroll(p *Gia_Man_t, nFrames int32) *Gia_Man_t {
    var pFrames, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, f int32;
    assert(Gia_ManRegNum(pAig) > 0, "Gia_ManRegNum(pAig) > 0");
    pFrames = Gia_ManStart( Gia_ManObjNum(pAig) );
    pFrames.pName = Abc_UtilStrsav( pAig.pName );
    pFrames.pSpec = Abc_UtilStrsav( pAig.pSpec );
    Gia_ManHashAlloc( pFrames );
    Gia_ManConst0(pAig).Value = 0;
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
            if f != 0 {
                pObj.Value = Gia_ObjRoToRi( pAig, pObj ).Value;
            } else {
                pObj.Value = 0;
            }
        }
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pFrames );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pFrames, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+i)); i++  {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+i)); i++  {
        Gia_ManAppendCo( pFrames, pObj.Value );
    }
    Gia_ManHashStop( pFrames );
    pFrames = Gia_ManCleanup( pTemp = pFrames );
    Gia_ManStop( pTemp );
    return pFrames;
}
func Int2_ManPreparePrefix(p *Gia_Man_t, f int32, pvCiMap **Vec_Int_t) *sat_solver {
    var pPref, pNew *Gia_Man_t;
    var pSat *sat_solver;
    // create subset of the timeframe
    pPref = Int2_ManUnroll( p, f );
    // create SAT solver
    pNew = Jf_ManDeriveCnf( pPref, 0 );  
    pCnf = (Cnf_Dat_t *)pPref.pData; pPref.pData = nil;
    Gia_ManStop( pPref );
    // derive the SAT solver
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    // collect indexes of CO variables
    *pvCiMap = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pNew)) && ((pObj) = Gia_ManCo(pNew, i)); i++  {     
        Vec_IntPush( *pvCiMap, pCnf.pVarNums[ Gia_ObjId(pNew, pObj) ] );
    }
    // cleanup
    Cnf_DataFree( pCnf );
    Gia_ManStop( pNew );
    return pSat;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManPrepareSuffix(p *Gia_Man_t, vImageOne *Vec_Int_t, vImagesAll *Vec_Int_t, pvCoMap **Vec_Int_t, ppSuff **Gia_Man_t) *sat_solver {
    var pSuff, pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var vLits *Vec_Int_t;
    var i, Lit, Limit int32;
    // create subset of the timeframe
    pSuff = Int2_ManProbToGia( p, vImageOne );
    assert(Gia_ManPiNum(pSuff) == Gia_ManCiNum(p), "Gia_ManPiNum(pSuff) == Gia_ManCiNum(p)");
    // create SAT solver
    pNew = Jf_ManDeriveCnf( pSuff, 0 );  
    pCnf = (Cnf_Dat_t *)pSuff.pData; pSuff.pData = nil;
    Gia_ManStop( pSuff );
    // derive the SAT solver
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    // create new constraints
    vLits = Vec_IntAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(vImagesAll)) && (((Limit) = Vec_IntEntry(vImagesAll, i)), 1); i++  {
        Vec_IntClear( vLits );
        for  k = 0; k < Limit; k++  {
            i++;
            Lit = Vec_IntEntry( vSop, i + k );
            Vec_IntPush( vLits, Abc_LitNot(Lit) );
        }
        if !sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) ) {  // UNSAT
        
            Vec_IntFree( vLits );
            Cnf_DataFree( pCnf );
            Gia_ManStop( pNew );
            *pvCoMap = nil;
            return nil;
        }
    }
    Vec_IntFree( vLits );
    // collect indexes of CO variables
    *pvCoMap = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj = Gia_ManPi( pNew, i + Gia_ManPiNum(p) );
        Vec_IntPush( *pvCoMap, pCnf.pVarNums[ Gia_ObjId(pNew, pObj) ] );
    }
    // cleanup
    Cnf_DataFree( pCnf );
    if ppSuff {
        *ppSuff = pNew;
    } else {
        Gia_ManStop( pNew ); }
    return pSat;
}

/**Function*************************************************************

  Synopsis    [Returns the cube cover and status.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManComputePreimage(pSuff *Gia_Man_t, pSatPref *sat_solver, pSatSuff *sat_solver, vCiMap *Vec_Int_t, vCoMap *Vec_Int_t, vPrio *Vec_Int_t) *Vec_Int_t {
    var i, iVar, status int32;
    Vec_IntClear( p.vImage );
    for {
        // run suffix solver
        status = sat_solver_solve( p.pSatSuff, nil, nil, 0, 0, 0, 0 );
        if status == l_Undef {
            return nil; } // timeout
        if status == l_False {
            return INT2_COMPUTED; }
        assert(status == l_True, "status == l_True");
        // collect assignment
        Vec_IntClear( p.vAssign );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vCiMap)) && (((iVar) = Vec_IntEntry(p.vCiMap, i)), 1); i++  {
            Vec_IntPush( p.vAssign, sat_solver_var_value(p.pSatSuff, iVar) );
        }
        // derive initial cube
        vCube = Int2_ManRefineCube( p.pSuff, p.vAssign, p.vPrio );
        // expend the cube using prefix
        status = sat_solver_solve( p.pSatPref, Vec_IntArray(vCube), Vec_IntArray(vCube) + Vec_IntSize(vCube), 0, 0, 0, 0 );
        if status == l_False {
            var k int32;
            var nCoreLits int32;
            var pCoreLits *int32;
            nCoreLits = sat_solver_final( p.pSatPref, &pCoreLits );
            // create cube
            Vec_IntClear( vCube );
            Vec_IntPush( vImage, nCoreLits );
            for  k = 0; k < nCoreLits; k++  {
                Vec_IntPush( vCube, pCoreLits[k] );
                Vec_IntPush( vImage, pCoreLits[k] );
            }
            // add cube to the solver
            if !sat_solver_addclause( p.pSatSuff, Vec_IntArray(vCube), Vec_IntArray(vCube) + Vec_IntSize(vCube) ) {
                Vec_IntFree( vCube );
                return INT2_COMPUTED;
            }
        }
        Vec_IntFree( vCube );
        if status == l_Undef {
            return INT2_TIME_OUT; }
        if status == l_True {
            return INT2_FALSE_NEG; }
        assert(status == l_False, "status == l_False");
        continue;
    }
    return p.vImage;
}

/**Function*************************************************************

  Synopsis    [Interpolates while the number of conflicts is not exceeded.]

  Description [Returns 1 if proven. 0 if failed. -1 if undecided.]
               
  SideEffects [Does not check the property in 0-th frame.]

  SeeAlso     []

***********************************************************************/
func Int2_ManPerformInterpolation(pInit *Gia_Man_t, pPars *Int2_ManPars_t) int32 {
    var p *Int2_Man_t;
    var f int32;
    var i int32;
    var RetValue int32 = -1;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    var timeTemp abctime;
    var nTimeToStop abctime;
    if pPars.nSecLimit {
        nTimeToStop = pPars.nSecLimit * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStop = 0;
    }

    // sanity checks
    assert(Gia_ManPiNum(pInit) > 0, "Gia_ManPiNum(pInit) > 0");
    assert(Gia_ManPoNum(pInit) > 0, "Gia_ManPoNum(pInit) > 0");
    assert(Gia_ManRegNum(pInit) > 0, "Gia_ManRegNum(pInit) > 0");

    // create manager
    p = Int2_ManCreate( pInit, pPars );

    // create SAT solver
    p.pSatPref = sat_solver_new();
    sat_solver_setnvars( p.pSatPref, 1000 );
    sat_solver_set_runtime_limit( p.pSatPref, nTimeToStop );

    // check outputs in the first frame
    for  i = 0; i < Gia_ManPoNum(pInit); i++  {
        Vec_IntPush( p.vPrefCos, i );
    }
    Int2_ManCreateFrames( p, 0, p.vPrefCos );
    RetValue = Int2_ManCheckBmc( p, nil );
    if RetValue != 1 {
        return RetValue; }

    // create original image
    for  f = pPars.nFramesS; f < p.nFramesMax; f++  {
        for  i = 0; i < p.nFramesMax; i++  {
            p.pSatSuff = Int2_ManPrepareSuffix( p, vImageOne. vImagesAll, &vCoMap, &pGiaSuff );
            sat_solver_set_runtime_limit( p.pSatSuff, nTimeToStop );
            Vec_IntFreeP( &vImageOne );
            vImageOne = Int2_ManComputePreimage( pGiaSuff, p.pSatPref, p.pSatSuff, vCiMap, vCoMap );
            Vec_IntFree( vCoMap );
            Gia_ManStop( pGiaSuff );
            if nTimeToStop && Abc_Clock() > nTimeToStop {
                return -1; }
            if vImageOne == nil {
                if i == 0 {
                    fmt.Printf("Satisfiable in frame %d.\n", f);
                    Vec_IntFree( vCiMap );
                    sat_solver_delete( p.pSatPref ); p.pSatPref = nil;
                    return 0;
                }
                f += i;
                break;
            }
            Vec_IntAppend( vImagesAll, vImageOne );
            sat_solver_delete( p.pSatSuff ); p.pSatSuff = nil;
        }
        Vec_IntFree( vCiMap );
        sat_solver_delete( p.pSatPref ); p.pSatPref = nil;
    }
    Abc_PrintTime( "Time", Abc_Clock() - clk );


p.timeSatPref += Abc_Clock() - clk;

    Int2_ManStop( p );
    return RetValue;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

