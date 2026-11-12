package bmc

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [bmcBmci.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcBmci.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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
func Cnf_DeriveGiaRemapped(p *Gia_Man_t) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t;
    var pAig *Aig_Man_t  = Gia_ManToAigSimple( p );
    pAig.nRegs = 0;
    pCnf = Cnf_Derive( pAig, Aig_ManCoNum(pAig) );
    Aig_ManStop( pAig );
    return pCnf;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_DataLiftGia(p *Cnf_Dat_t, pGia *Gia_Man_t, nVarsPlus int32) {
    var pObj *Gia_Obj_t;
    var v int32;
    // inlining of macro Gia_ManForEachObj
    for  v = 0; (v < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, v)); v++  {
        if p.pVarNums[Gia_ObjId(pGia, pObj)] >= 0 {
            p.pVarNums[Gia_ObjId(pGia, pObj)] += nVarsPlus; }
    }
    for  v = 0; v < p.nLiterals; v++  {
        p.pClauses[0][v] += 2*nVarsPlus;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_BmciUnfold(pNew *Gia_Man_t, p *Gia_Man_t, vFFLits *Vec_Int_t, fPiReuse int32) {
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManRegNum(p) == Vec_IntSize(vFFLits), "Gia_ManRegNum(p) == Vec_IntSize(vFFLits)");
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.Value = Vec_IntEntry(vFFLits, i);
    }
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if fPiReuse != 0 {
            pObj.Value = Gia_ObjToLit(pNew, Gia_ManPi(pNew, Gia_ManPiNum(pNew)-Gia_ManPiNum(p)+i));
        } else {
            pObj.Value = Gia_ManAppendCi(pNew);
        }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        Vec_IntWriteEntry( vFFLits, i, Gia_ObjFanin0Copy(pObj) );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_BmciPart_rec(pNew *Gia_Man_t, vSatMap *Vec_Int_t, iIdNew int32, pPart *Gia_Man_t, vPartMap *Vec_Int_t, vCopies *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( pNew, iIdNew ); 
    var iLitPart0, iLitPart1, iRes int32;
    if Vec_IntEntry(vCopies, iIdNew) {
        return Vec_IntEntry(vCopies, iIdNew); }
    if Vec_IntEntry(vSatMap, iIdNew) >= 0 || Gia_ObjIsCi(pObj) {
        Vec_IntPush( vPartMap, iIdNew );
        iRes = Gia_ManAppendCi(pPart);
        Vec_IntWriteEntry( vCopies, iIdNew, iRes );
        return iRes;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    iLitPart0 = Bmc_BmciPart_rec( pNew, vSatMap, Gia_ObjFaninId0(pObj, iIdNew), pPart, vPartMap, vCopies );
    iLitPart1 = Bmc_BmciPart_rec( pNew, vSatMap, Gia_ObjFaninId1(pObj, iIdNew), pPart, vPartMap, vCopies );
    iLitPart0 = Abc_LitNotCond( iLitPart0, Gia_ObjFaninC0(pObj) );
    iLitPart1 = Abc_LitNotCond( iLitPart1, Gia_ObjFaninC1(pObj) );
    Vec_IntPush( vPartMap, iIdNew );
    iRes = Gia_ManAppendAnd( pPart, iLitPart0, iLitPart1 );
    Vec_IntWriteEntry( vCopies, iIdNew, iRes );
    return iRes;
}
func Bmc_BmciPart(pNew *Gia_Man_t, vSatMap *Vec_Int_t, vMiters *Vec_Int_t, vPartMap *Vec_Int_t, vCopies *Vec_Int_t) *Gia_Man_t {
    var pPart *Gia_Man_t;
    var i, iLit, iLitPart int32;
    Vec_IntFill( vCopies, Gia_ManObjNum(pNew), 0 );
    Vec_IntFillExtra( vSatMap, Gia_ManObjNum(pNew), -1 );
    pPart = Gia_ManStart( 1000 );
    pPart.pName = Abc_UtilStrsav( pNew.pName );
    Vec_IntClear( vPartMap );
    Vec_IntPush( vPartMap, 0 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vMiters)) && (((iLit) = Vec_IntEntry(vMiters, i)), 1); i++  {
        if iLit == -1 {
            continue; }
        assert(iLit >= 2, "iLit >= 2");
        iLitPart = Bmc_BmciPart_rec( pNew, vSatMap, Abc_Lit2Var(iLit), pPart, vPartMap, vCopies );
        Gia_ManAppendCo( pPart, Abc_LitNotCond(iLitPart, Abc_LitIsCompl(iLit)) );
        Vec_IntPush( vPartMap, -1 );
    }
    assert(Gia_ManPoNum(pPart) > 0, "Gia_ManPoNum(pPart) > 0");
    assert(Vec_IntSize(vPartMap) == Gia_ManObjNum(pPart), "Vec_IntSize(vPartMap) == Gia_ManObjNum(pPart)");
    return pPart;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_BmciPerform(p *Gia_Man_t, vInit0 *Vec_Int_t, vInit1 *Vec_Int_t, nFrames int32, nWords int32, nTimeOut int32, fVerbose int32) int32 {
    var nSatVars int32  = 1;
    var vLits0, vLits1, vMiters, vSatMap, vPartMap, vCopies *Vec_Int_t;
    var pNew, pPart *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var iVar0, iVar1, iLit, iLit0, iLit1 int32;
    var i int32;
    var f int32;
    var status int32;
    var nChanges int32;
    var nMiters int32;
    var RetValue int32 = 1;
    assert(Vec_IntSize(vInit0) == Gia_ManRegNum(p), "Vec_IntSize(vInit0) == Gia_ManRegNum(p)");
    assert(Vec_IntSize(vInit1) == Gia_ManRegNum(p), "Vec_IntSize(vInit1) == Gia_ManRegNum(p)");

    // start the SAT solver
    pSat = sat_solver_new();
    sat_solver_set_runtime_limit( pSat, nTimeOut ? nTimeOut * int64(time.Second) + Abc_Clock(): 0 );

    pNew = Gia_ManStart( 10000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;

    vLits0 = Vec_IntAlloc( Gia_ManRegNum(p) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vInit0)) && (((iLit) = Vec_IntEntry(vInit0, i)), 1); i++  {
        Vec_IntPush( vLits0, iLit < 2 ? iLit : Gia_ManAppendCi(pNew) );
    }

    vLits1 = Vec_IntAlloc( Gia_ManRegNum(p) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vInit1)) && (((iLit) = Vec_IntEntry(vInit1, i)), 1); i++  {
        Vec_IntPush( vLits1, iLit < 2 ? iLit : Gia_ManAppendCi(pNew) );
    }

    vMiters  = Vec_IntAlloc( 1000 );
    vSatMap  = Vec_IntAlloc( 1000 );
    vPartMap = Vec_IntAlloc( 1000 );
    vCopies  = Vec_IntAlloc( 1000 );
    for  f = 0; f < nFrames; f++  {
        var clk abctime  = Abc_Clock();
        Bmc_BmciUnfold( pNew, p, vLits0, 0 );
        Bmc_BmciUnfold( pNew, p, vLits1, 1 );
        assert(Vec_IntSize(vLits0) == Vec_IntSize(vLits1), "Vec_IntSize(vLits0) == Vec_IntSize(vLits1)");
        nMiters  = 0;
        Vec_IntClear( vMiters );
        // inlining of macro Vec_IntForEachEntryTwo
        for  i = 0; (i < Vec_IntSize(vLits0)) && (((iLit0) = Vec_IntEntry(vLits0, i)), 1) && (((iLit1) = Vec_IntEntry(vLits1, i)), 1); i++  {
            if (iLit0 >= 2 || iLit1 >= 2) && iLit0 != iLit1 {
                Vec_IntPush( vMiters, Gia_ManHashXor(pNew, iLit0, iLit1) );
nMiters++;
            } else {
                Vec_IntPush( vMiters, -1 ); }
        }
        assert(Vec_IntSize(vMiters) == Gia_ManRegNum(p), "Vec_IntSize(vMiters) == Gia_ManRegNum(p)");
        if Vec_IntSum(vMiters) + Vec_IntSize(vLits1) == 0 {
            if fVerbose != 0 {
                fmt.Printf("Reached a fixed point after %d frames.  \n", f+1); }
            break;
        }
        // create new part
        pPart = Bmc_BmciPart( pNew, vSatMap, vMiters, vPartMap, vCopies );
        pCnf = Cnf_DeriveGiaRemapped( pPart );
        Cnf_DataLiftGia( pCnf, pPart, nSatVars );
        nSatVars += pCnf.nVars;
        sat_solver_setnvars( pSat, nSatVars );
        for  i = 0; i < pCnf.nClauses; i++  {
            if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
                assert(false, "0"); }
        }
        // stitch the clauses
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(pPart)) && ((pObj) = Gia_ManCi(pPart, i)); i++  {
            iVar0 = pCnf.pVarNums[Gia_ObjId(pPart, pObj)];
            iVar1 = Vec_IntEntry( vSatMap, Vec_IntEntry(vPartMap, Gia_ObjId(pPart, pObj)) );
            if iVar1 == -1 {
                continue; }
            sat_solver_add_buffer( pSat, iVar0, iVar1, 0 );
        }
        // transfer variables
        // inlining of macro Gia_ManForEachCand
        for  i = 0; (i < pPart.nObjs) && ((pObj) = Gia_ManObj(pPart, i)); i++  {      if !Gia_ObjIsCand(pObj) {} else
            if pCnf.pVarNums[i] >= 0 {
                assert(Gia_ObjIsCi(pObj) || Vec_IntEntry( vSatMap, Vec_IntEntry(vPartMap, i) ) == -1, "Gia_ObjIsCi(pObj) || Vec_IntEntry( vSatMap, Vec_IntEntry(vPartMap, i) ) == -1");
                Vec_IntWriteEntry( vSatMap, Vec_IntEntry(vPartMap, i), pCnf.pVarNums[i] );
            }
        }
        Cnf_DataFree( pCnf );
        Gia_ManStop( pPart );
        // perform runs
        nChanges = 0;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vMiters)) && (((iLit) = Vec_IntEntry(vMiters, i)), 1); i++  {
            if iLit == -1 {
                continue; }
            assert(Vec_IntEntry(vSatMap, Abc_Lit2Var(iLit)) > 0, "Vec_IntEntry(vSatMap, Abc_Lit2Var(iLit)) > 0");
            iLit = Abc_Lit2LitV( Vec_IntArray(vSatMap), iLit );
            status = sat_solver_solve( pSat, &iLit, &iLit + 1, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
            if status == l_True {
                nChanges++;
                continue;
            }
            if status == l_Undef {
                fmt.Printf("Timeout reached after %d seconds.  \n", nTimeOut);
                RetValue = 0;
                goto cleanup;
            }
            assert(status == l_False, "status == l_False");
            iLit0 = Vec_IntEntry( vLits0, i );
            iLit1 = Vec_IntEntry( vLits1, i );
            assert((iLit0 >= 2 || iLit1 >= 2) && iLit0 != iLit1, "(iLit0 >= 2 || iLit1 >= 2) && iLit0 != iLit1");
            if iLit1 >= 2 {
                Vec_IntWriteEntry( vLits1, i, iLit0 );
            } else {
                Vec_IntWriteEntry( vLits0, i, iLit1 ); }
            iLit0 = Vec_IntEntry( vLits0, i );
            iLit1 = Vec_IntEntry( vLits1, i );
            assert(iLit0 == iLit1, "iLit0 == iLit1");
        }
        if fVerbose != 0 {
            fmt.Printf("Frame %4d : ", f+1);
            fmt.Printf("Vars =%7d  ", nSatVars);
            fmt.Printf("Clause =%10d  ", sat_solver_nclauses(pSat));
            fmt.Printf("Conflict =%10d  ", sat_solver_nconflicts(pSat));
            fmt.Printf("AIG =%7d  ", Gia_ManAndNum(pNew));
            fmt.Printf("Miters =%5d  ", nMiters);
            fmt.Printf("SAT =%5d  ", nChanges);
            Abc_PrintTime( 1, "Time",   Abc_Clock() - clk );
        }
        if nChanges == 0 {
            fmt.Printf("Reached a fixed point after %d frames.  \n", f+1);
            break;
        }
    }
cleanup:

    sat_solver_delete( pSat );
    Gia_ManStopP( &pNew );
    Vec_IntFree( vLits0 );
    Vec_IntFree( vLits1 );
    Vec_IntFree( vMiters );
    Vec_IntFree( vSatMap );
    Vec_IntFree( vPartMap );
    Vec_IntFree( vCopies );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManBmciTest(p *Gia_Man_t, vInit *Vec_Int_t, nFrames int32, nWords int32, nTimeOut int32, fSim int32, fVerbose int32) int32 {
    var vInit0 *Vec_Int_t  = Vec_IntStart( Gia_ManRegNum(p) );
    Bmc_BmciPerform( p, vInit, vInit0, nFrames, nWords, nTimeOut, fVerbose );
    Vec_IntFree( vInit0 );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

