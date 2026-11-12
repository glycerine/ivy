package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaFalse.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Detection and elimination of false paths.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaFalse.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecQue.h"
//#include "misc/vec/vecWec.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reconstruct the AIG after detecting false paths.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFalseRebuildOne(pNew *Gia_Man_t, p *Gia_Man_t, vHook *Vec_Int_t, fVerbose int32, fVeryVerbose int32) {
    var pObj *Gia_Obj_t;
    var pObj1 *Gia_Obj_t;
    var pPrev *Gia_Obj_t = nil;
    var i int32;
    var CtrlValue int32;
    var iPrevValue int32 = -1;
    pObj = Gia_ManObj( p, Vec_IntEntry(vHook, 0) );
    if Vec_IntSize(vHook) == 1 {
        pObj.Value = 0; // what if stuck at 1???
        return;
    }
    assert(Vec_IntSize(vHook) >= 2, "Vec_IntSize(vHook) >= 2");
    // find controlling value
    pObj1 = Gia_ManObj( p, Vec_IntEntry(vHook, 1) );
    if Gia_ObjFanin0(pObj1) == pObj {
        CtrlValue = Gia_ObjFaninC0(pObj1);
    } else if Gia_ObjFanin1(pObj1) == pObj {
        CtrlValue = Gia_ObjFaninC1(pObj1);
    } else { assert(false, "0"); }
//    printf( "%d ", CtrlValue );
    // rewrite the path
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vHook)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vHook,i))); i++  {
        var iObjValue int32  = pObj.Value;
        if i != 0 {
            pObj.Value = Gia_ManHashAnd(pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj));
        } else {
            pObj.Value = CtrlValue;
        }
        if pPrev != nil {
            pPrev.Value = iPrevValue; }
        iPrevValue = iObjValue;
        pPrev = pObj;
    }
    if fVeryVerbose != 0 {
        fmt.Printf("Eliminated path: ");
        Vec_IntPrint( vHook );
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vHook)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vHook,i))); i++  {
            fmt.Printf("Level %3d : ", Gia_ObjLevel(p, pObj));
            Gia_ObjPrint( p, pObj );
        }
    }
}
func Gia_ManFalseRebuild(p *Gia_Man_t, vHooks *Vec_Wec_t, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    pNew = Gia_ManStart( 4 * Gia_ManObjNum(p) / 3 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            if Vec_WecLevelSize(vHooks, i) > 0 {
                if fVeryVerbose != 0 {
                    fmt.Printf("Path %d : ", Counter++); }
                Gia_ManFalseRebuildOne( pNew, p, Vec_WecEntry(vHooks, i), fVerbose, fVeryVerbose );
            
            } else {
                pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derive critical path by following minimum slacks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectPath_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vPath *Vec_Int_t) {
    if Gia_ObjIsAnd(pObj) {
        if Gia_ObjLevel(p, Gia_ObjFanin0(pObj)) > Gia_ObjLevel(p, Gia_ObjFanin1(pObj)) {
            Gia_ManCollectPath_rec( p, Gia_ObjFanin0(pObj), vPath );
        } else if Gia_ObjLevel(p, Gia_ObjFanin0(pObj)) < Gia_ObjLevel(p, Gia_ObjFanin1(pObj)) {
            Gia_ManCollectPath_rec( p, Gia_ObjFanin1(pObj), vPath );
//        else if ( rand() & 1 )
//            Gia_ManCollectPath_rec( p, Gia_ObjFanin0(pObj), vPath );
        } else {
            Gia_ManCollectPath_rec( p, Gia_ObjFanin1(pObj), vPath ); }
    }
    Vec_IntPush( vPath, Gia_ObjId(p, pObj) );
}
func Gia_ManCollectPath(p *Gia_Man_t, pObj *Gia_Obj_t) *Vec_Int_t {
    var vPath *Vec_Int_t  = Vec_IntAlloc( p.nLevels );
    Gia_ManCollectPath_rec( p, Gia_ObjIsCo(pObj) ? Gia_ObjFanin0(pObj) : pObj, vPath );
    return vPath;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckFalseOne(p *Gia_Man_t, iOut int32, nTimeOut int32, vHooks *Vec_Wec_t, fVerbose int32, fVeryVerbose int32) {
    var pSat *sat_solver;
    var pObj, pFanin *Gia_Obj_t;
    var pPivot *Gia_Obj_t  = Gia_ManCo( p, iOut );
    var vLits *Vec_Int_t  = Vec_IntAlloc( p.nLevels );
    var vPath *Vec_Int_t  = Gia_ManCollectPath( p, pPivot );
    var nLits int32;
    var pLits *int32 = nil;
    var i int32;
    var Shift [2]int32;
    var status int32;
    var clkStart abctime  = Abc_Clock();
    // collect objects and assign SAT variables
    var iFanin int32  = Gia_ObjFaninId0p( p, pPivot );
    var vObjs *Vec_Int_t  = Gia_ManCollectNodesCis( p, &iFanin, 1 );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        pObj.Value = Vec_IntSize(vObjs) - 1 - i;
    }
    assert(Gia_ObjIsCo(pPivot), "Gia_ObjIsCo(pPivot)");
    // create SAT solver
    pSat = sat_solver_new();
    sat_solver_set_runtime_limit( pSat, nTimeOut ? nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
    sat_solver_setnvars( pSat, 3 * Vec_IntSize(vPath) + 2 * Vec_IntSize(vObjs) );
    Shift[0] = 3 * Vec_IntSize(vPath);
    Shift[1] = 3 * Vec_IntSize(vPath) + Vec_IntSize(vObjs);
    // add CNF for the cone
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        if !Gia_ObjIsAnd(pObj) {
            continue; }
        sat_solver_add_and( pSat, pObj.Value + Shift[0], 
            Gia_ObjFanin0(pObj).Value + Shift[0], Gia_ObjFanin1(pObj).Value + Shift[0], 
            Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
        sat_solver_add_and( pSat, pObj.Value + Shift[1], 
            Gia_ObjFanin0(pObj).Value + Shift[1], Gia_ObjFanin1(pObj).Value + Shift[1], 
            Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
    }
    // add CNF for the path
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPath)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vPath,i))); i++  {
        if Gia_ObjIsAnd(pObj) {
            assert(i > 0, "i > 0");
            pFanin = Gia_ManObj( p, Vec_IntEntry(vPath, i-1) );
            if pFanin == Gia_ObjFanin0(pObj) {
                sat_solver_add_and( pSat, i + 1*Vec_IntSize(vPath), 
                    i-1 + 1*Vec_IntSize(vPath), Gia_ObjFanin1(pObj).Value + Shift[0], 
                    Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
                sat_solver_add_and( pSat, i + 2*Vec_IntSize(vPath), 
                    i-1 + 2*Vec_IntSize(vPath), Gia_ObjFanin1(pObj).Value + Shift[1], 
                    Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
            
            } else if pFanin == Gia_ObjFanin1(pObj) {
                sat_solver_add_and( pSat, i + 1*Vec_IntSize(vPath), 
                    Gia_ObjFanin0(pObj).Value + Shift[0], i-1 + 1*Vec_IntSize(vPath), 
                    Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
                sat_solver_add_and( pSat, i + 2*Vec_IntSize(vPath), 
                    Gia_ObjFanin0(pObj).Value + Shift[1], i-1 + 2*Vec_IntSize(vPath), 
                    Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
            
            } else { assert(false, "0"); }
            sat_solver_add_xor( pSat, i, i + 1*Vec_IntSize(vPath), i + 2*Vec_IntSize(vPath), 0 );
        
        } else if Gia_ObjIsCi(pObj) {
            sat_solver_add_xor( pSat, i, pObj.Value + Shift[0], pObj.Value + Shift[1], 0 );
        } else { assert(false, "0"); }
        Vec_IntPush( vLits, Abc_Var2Lit(i, 0) );
    }
    // call the SAT solver
    status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), (ABC_INT64_T)nTimeOut, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if status == l_False {
        var iBeg, iEnd int32;
        nLits = sat_solver_final( pSat, &pLits );
        iBeg  = Abc_Lit2Var(pLits[nLits-1]);
        iEnd  = Abc_Lit2Var(pLits[0]);
        if iEnd - iBeg < 20 {
            // check if nodes on the path are already used
            for  i = Abc_MaxInt(iBeg-1, 0); i <= iEnd; i++  {
                if Vec_WecLevelSize(vHooks, Vec_IntEntry(vPath, i)) > 0 {
                    break; }
            }
            if i > iEnd {
                var vHook *Vec_Int_t  = Vec_WecEntry(vHooks, Vec_IntEntry(vPath, iEnd));
                for  i = Abc_MaxInt(iBeg-1, 0); i <= iEnd; i++  {
                    Vec_IntPush( vHook, Vec_IntEntry(vPath, i) );
                }
            }
        }
    }
    if fVerbose != 0 {
        fmt.Printf("PO %6d : Level = %3d  ", iOut, Gia_ObjLevel(p, pPivot));
        if status == l_Undef {
            fmt.Printf("Timeout reached after %d seconds. ", nTimeOut);
        } else if status == l_True {
            fmt.Printf("There is no false path. ");
        } else {
            fmt.Printf("False path contains %d nodes (out of %d):  ", nLits, Vec_IntSize(vPath));
            fmt.Printf("top = %d  ", Vec_IntEntry(vPath, Abc_Lit2Var(pLits[0])));
            if fVeryVerbose != 0 {
                for  i = 0; i < nLits; i++  {
                    fmt.Printf("%d ", Abc_Lit2Var(pLits[i]));
                } }
            fmt.Printf("  ");
        }
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    }
    sat_solver_delete( pSat );
    Vec_IntFree( vObjs );
    Vec_IntFree( vPath );
    Vec_IntFree( vLits );
}
func Gia_ManCheckFalse2(p *Gia_Man_t, nSlackMax int32, nTimeOut int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vHooks *Vec_Wec_t;
    var vPrio *Vec_Que_t;
    var vWeights *Vec_Flt_t;
    var pObj *Gia_Obj_t;
    var i int32;
//    srand( 111 );
    Gia_ManLevelNum( p );
    // create PO weights
    vWeights = Vec_FltAlloc( Gia_ManCoNum(p) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_FltPush( vWeights, Gia_ObjLevel(p, pObj) );
    }
    // put POs into the queue
    vPrio = Vec_QueAlloc( Gia_ManCoNum(p) );
    Vec_QueSetPriority( vPrio, Vec_FltArrayP(vWeights) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_QuePush( vPrio, i );
    }
    // work on each PO in the queue
    vHooks = Vec_WecStart( Gia_ManObjNum(p) );
    for  Vec_QueTopPriority(vPrio) >= p.nLevels - nSlackMax  {
        Gia_ManCheckFalseOne( p, Vec_QuePop(vPrio), nTimeOut, vHooks, fVerbose, fVeryVerbose );
    }
    if fVerbose != 0 {
    fmt.Printf("Collected %d non-overlapping false paths.\n", Vec_WecSizeUsed(vHooks)); }
    // reconstruct the AIG
    pNew = Gia_ManFalseRebuild( p, vHooks, fVerbose, fVeryVerbose );
    // cleanup
    Vec_WecFree( vHooks );
    Vec_FltFree( vWeights );
    Vec_QueFree( vPrio );
    return pNew;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFalseRebuildPath(p *Gia_Man_t, vHooks *Vec_Int_t, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var iPathEnd int32 = Vec_IntEntryLast(vHooks);
    pNew = Gia_ManStart( 4 * Gia_ManObjNum(p) / 3 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            if iPathEnd == i {
                Gia_ManFalseRebuildOne( pNew, p, vHooks, fVerbose, fVeryVerbose );
            } else {
                pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Gia_ManCheckOne(p *Gia_Man_t, iOut int32, iObj int32, nTimeOut int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var pSat *sat_solver;
    var pNew *Gia_Man_t;
    var pObj, pFanin *Gia_Obj_t;
    var vLits *Vec_Int_t  = Vec_IntAlloc( p.nLevels );
    var vPath *Vec_Int_t  = Gia_ManCollectPath( p, Gia_ManObj(p, iObj) );
    var nLits int32;
    var pLits *int32 = nil;
    var i int32;
    var Shift [2]int32;
    var status int32;
    var clkStart abctime  = Abc_Clock();
    // collect objects and assign SAT variables
    var vObjs *Vec_Int_t  = Gia_ManCollectNodesCis( p, &iObj, 1 );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        pObj.Value = Vec_IntSize(vObjs) - 1 - i;
    }
    // create SAT solver
    pSat = sat_solver_new();
    sat_solver_set_runtime_limit( pSat, nTimeOut ? nTimeOut * int64(time.Second) + Abc_Clock(): 0 );
    sat_solver_setnvars( pSat, 3 * Vec_IntSize(vPath) + 2 * Vec_IntSize(vObjs) );
    Shift[0] = 3 * Vec_IntSize(vPath);
    Shift[1] = 3 * Vec_IntSize(vPath) + Vec_IntSize(vObjs);
    // add CNF for the cone
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        if !Gia_ObjIsAnd(pObj) {
            continue; }
        sat_solver_add_and( pSat, pObj.Value + Shift[0], 
            Gia_ObjFanin0(pObj).Value + Shift[0], Gia_ObjFanin1(pObj).Value + Shift[0], 
            Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
        sat_solver_add_and( pSat, pObj.Value + Shift[1], 
            Gia_ObjFanin0(pObj).Value + Shift[1], Gia_ObjFanin1(pObj).Value + Shift[1], 
            Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
    }
    // add CNF for the path
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPath)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vPath,i))); i++  {
        if !Gia_ObjIsAnd(pObj) {
            continue; }
        assert(i > 0, "i > 0");
        pFanin = Gia_ManObj( p, Vec_IntEntry(vPath, i-1) );
        if pFanin == Gia_ObjFanin0(pObj) {
            sat_solver_add_and( pSat, i + 1*Vec_IntSize(vPath), 
                i-1 + 1*Vec_IntSize(vPath), Gia_ObjFanin1(pObj).Value + Shift[0], 
                Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
            sat_solver_add_and( pSat, i + 2*Vec_IntSize(vPath), 
                i-1 + 2*Vec_IntSize(vPath), Gia_ObjFanin1(pObj).Value + Shift[1], 
                Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
        
        } else if pFanin == Gia_ObjFanin1(pObj) {
            sat_solver_add_and( pSat, i + 1*Vec_IntSize(vPath), 
                Gia_ObjFanin0(pObj).Value + Shift[0], i-1 + 1*Vec_IntSize(vPath), 
                Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
            sat_solver_add_and( pSat, i + 2*Vec_IntSize(vPath), 
                Gia_ObjFanin0(pObj).Value + Shift[1], i-1 + 2*Vec_IntSize(vPath), 
                Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); 
        
        } else { assert(false, "0"); }
        sat_solver_add_xor( pSat, i, i + 1*Vec_IntSize(vPath), i + 2*Vec_IntSize(vPath), 0 );
        Vec_IntPush( vLits, Abc_Var2Lit(i, 0) );
    }
    // call the SAT solver
    status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), (ABC_INT64_T)nTimeOut, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    Vec_IntClear( vLits );
    if status == l_False {
        var iBeg, iEnd int32;
        nLits = sat_solver_final( pSat, &pLits );
        iBeg  = Abc_Lit2Var(pLits[nLits-1]);
        iEnd  = Abc_Lit2Var(pLits[0]);
        assert(iBeg <= iEnd, "iBeg <= iEnd");
        // collect path
        for  i = Abc_MaxInt(iBeg-1, 0); i <= iEnd; i++  {
//        for ( i = 0; i < Vec_IntSize(vPath); i++ )
            Vec_IntPush( vLits, Vec_IntEntry(vPath, i) );
        }
    }
    if fVerbose != 0 {
        fmt.Printf("PO %6d : Level = %3d  ", iOut, Gia_ObjLevelId(p, iObj));
        if status == l_Undef {
            fmt.Printf("Timeout reached after %d seconds. ", nTimeOut);
        } else if status == l_True {
            fmt.Printf("There is no false path. ");
        } else {
            fmt.Printf("False path contains %d nodes (out of %d):  ", Vec_IntSize(vLits), Vec_IntSize(vPath));
            if fVeryVerbose != 0 {
                for  i = nLits-1; i >= 0; i--  {
                    fmt.Printf("%d ", Vec_IntEntry(vPath, Abc_Lit2Var(pLits[i])));
                } }
            fmt.Printf("  ");
        }
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    }
    sat_solver_delete( pSat );
    Vec_IntFree( vObjs );
    Vec_IntFree( vPath );
    // update the AIG
    if Vec_IntSize(vLits) {
        pNew = Gia_ManFalseRebuildPath( p, vLits, fVerbose, fVeryVerbose );
    } else {
        pNew = nil;
    }
    Vec_IntFree( vLits );
/*
    if ( pNew )
    {
        Gia_Man_t * pTemp = Gia_ManDupDfsNode( p, Gia_ManObj(p, iObj) );
        Gia_AigerWrite( pTemp, "false.aig", 0, 0, 0 );
        Abc_Print( 1, "Dumping cone with %d nodes into file \"%s\".\n", Gia_ManAndNum(pTemp), "false.aig" );
        Gia_ManStop( pTemp );
    }
*/
    return pNew;
}
func Gia_ManCheckFalseAll(p *Gia_Man_t, nSlackMax int32, nTimeOut int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var Tried, Changed int32;
    p = Gia_ManDup( p );
    for {
        var pNew *Gia_Man_t;
        var pObj *Gia_Obj_t;
        var i int32;
        var LevelMax int32;
        var Changed0 int32 = Changed;
        LevelMax = Gia_ManLevelNum( p );
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            if Gia_ObjLevel(p, pObj) > nSlackMax {
                continue; }
            Tried++;
            pNew = Gia_ManCheckOne( p, -1, i, nTimeOut, fVerbose, fVeryVerbose );
            if pNew == nil {
                continue; }
            Changed++;
            Gia_ManStop( p );
            p = pNew;
            LevelMax = Gia_ManLevelNum( p );
        }
        }
        if Changed0 == Changed {
            break; }
    }
//    if ( fVerbose )
        fmt.Printf("Performed %d attempts and %d changes.\n", Tried, Changed);
    return p;
}
func Gia_ManCheckFalse(p *Gia_Man_t, nSlackMax int32, nTimeOut int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var Tried, Changed int32;
    var vTried *Vec_Int_t;
//    srand( 111 );
    p = Gia_ManDup( p );
    vTried = Vec_IntStart( Gia_ManCoNum(p) );
    for {
        var pNew *Gia_Man_t;
        var pObj *Gia_Obj_t;
        var i int32;
        var LevelMax int32;
        var Changed0 int32 = Changed;
        LevelMax = Gia_ManLevelNum( p );
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if !Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) {
                continue; }
            if Gia_ObjLevel(p, Gia_ObjFanin0(pObj)) < LevelMax - nSlackMax {
                continue; }
            if Vec_IntEntry( vTried, i ) {
                continue; }
            Tried++;
            pNew = Gia_ManCheckOne( p, i, Gia_ObjFaninId0p(p, pObj), nTimeOut, fVerbose, fVeryVerbose );
/*
            if ( i != 126 && pNew )
            {
                Gia_ManStop( pNew );
                pNew = NULL;
            }
*/
            if pNew == nil {
                Vec_IntWriteEntry( vTried, i, 1 );
                continue;
            }
            Changed++;
            Gia_ManStop( p );
            p = pNew;
            LevelMax = Gia_ManLevelNum( p );
        }
        if Changed0 == Changed {
            break; }
    }
//    if ( fVerbose )
        fmt.Printf("Performed %d attempts and %d changes.\n", Tried, Changed);
    Vec_IntFree( vTried );
    return p;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

