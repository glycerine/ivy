package bmc

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [bmcICheck.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Performs specialized check.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcICheck.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_DeriveSolver(p *Gia_Man_t, pMiter *Gia_Man_t, pCnf *Cnf_Dat_t, nFramesMax int32, nTimeOut int32, fVerbose int32) *sat_solver {
    var pSat *sat_solver;
    var vLits *Vec_Int_t;
    var pObj, pObj0, pObj1 *Gia_Obj_t;
    var i, k, iVar0, iVar1, iVarOut int32;
    var VarShift int32;

    // start the SAT solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, Gia_ManRegNum(p) + Gia_ManCoNum(p) + pCnf.nVars * (nFramesMax + 1) );
    sat_solver_set_runtime_limit( pSat, nTimeOut ? nTimeOut * int64(time.Second) + Abc_Clock(): 0 );

    // add one large OR clause
    vLits = Vec_IntAlloc( Gia_ManCoNum(p) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_IntPush( vLits, Abc_Var2Lit(Gia_ManRegNum(p) + i, 0) );
    }
    sat_solver_addclause( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits) );

    // load the last timeframe
    Cnf_DataLiftGia( pCnf, pMiter, Gia_ManRegNum(p) + Gia_ManCoNum(p) );
    VarShift += Gia_ManRegNum(p) + Gia_ManCoNum(p);

    // add XOR clauses
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj0 = Gia_ManPo( pMiter, 2*i+0 );
        pObj1 = Gia_ManPo( pMiter, 2*i+1 );
        iVar0 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj0)];
        iVar1 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj1)];
        iVarOut = Gia_ManRegNum(p) + i;
        sat_solver_add_xor( pSat, iVar0, iVar1, iVarOut, 0 );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj0 = Gia_ManRi( pMiter, i );
        pObj1 = Gia_ManRi( pMiter, i + Gia_ManRegNum(p) );
        iVar0 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj0)];
        iVar1 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj1)];
        iVarOut = Gia_ManRegNum(p) + Gia_ManPoNum(p) + i;
        sat_solver_add_xor_and( pSat, iVarOut, iVar0, iVar1, i );
    }
    // add timeframe clauses
    for  i = 0; i < pCnf.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
            assert(false, "0"); }
    }

    // add other timeframes
    for  k = 0; k < nFramesMax; k++  {
        // collect variables of the RO nodes
        Vec_IntClear( vLits );
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(pMiter)) && ((pObj) = Gia_ManCi(pMiter, Gia_ManPiNum(pMiter)+i)); i++  {
            Vec_IntPush( vLits, pCnf.pVarNums[Gia_ObjId(pMiter, pObj)] );
        }
        // lift CNF again
        Cnf_DataLiftGia( pCnf, pMiter, pCnf.nVars );
        VarShift += pCnf.nVars;
        // stitch the clauses
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(pMiter)) && ((pObj) = Gia_ManCo(pMiter, Gia_ManPoNum(pMiter)+i)); i++  {
            iVar0 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj)];
            iVar1 = Vec_IntEntry( vLits, i );
            if iVar1 == -1 {
                continue; }
            sat_solver_add_buffer( pSat, iVar0, iVar1, 0 );
        }
        // add equality clauses for the COs
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            pObj0 = Gia_ManPo( pMiter, 2*i+0 );
            pObj1 = Gia_ManPo( pMiter, 2*i+1 );
            iVar0 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj0)];
            iVar1 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj1)];
            sat_solver_add_buffer( pSat, iVar0, iVar1, 0 );
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            pObj0 = Gia_ManRi( pMiter, i );
            pObj1 = Gia_ManRi( pMiter, i + Gia_ManRegNum(p) );
            iVar0 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj0)];
            iVar1 = pCnf.pVarNums[Gia_ObjId(pMiter, pObj1)];
            sat_solver_add_buffer_enable( pSat, iVar0, iVar1, i, 0 );
        }
        // add timeframe clauses
        for  i = 0; i < pCnf.nClauses; i++  {
            if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
                assert(false, "0"); }
        }
    }
//    sat_solver_compress( pSat );
    Cnf_DataLiftGia( pCnf, pMiter, -VarShift );
    Vec_IntFree( vLits );
    return pSat;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_PerformICheck(p *Gia_Man_t, nFramesMax int32, nTimeOut int32, fEmpty int32, fVerbose int32) {
    var fUseOldCnf int32;
    var pMiter, pTemp *Gia_Man_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var vLits, vUsed *Vec_Int_t;
    var i, status, Lit int32;
    var nLitsUsed int32;
    var nLits int32;
    var pLits *int32;
    var clkStart abctime  = Abc_Clock();
    assert(nFramesMax > 0, "nFramesMax > 0");
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");

    if fVerbose != 0 {
    fmt.Printf("Solving M-inductiveness for design %s with %d AND nodes and %d flip-flops:\n", Gia_ManName(p), Gia_ManAndNum(p), Gia_ManRegNum(p)); }

    // create miter
    pTemp = Gia_ManDup( p );
    pMiter = Gia_ManMiter( p, pTemp, 0, 1, 1, 0, 0 );
    Gia_ManStop( pTemp );
    assert(Gia_ManPoNum(pMiter)  == 2 * Gia_ManPoNum(p), "Gia_ManPoNum(pMiter)  == 2 * Gia_ManPoNum(p)");
    assert(Gia_ManRegNum(pMiter) == 2 * Gia_ManRegNum(p), "Gia_ManRegNum(pMiter) == 2 * Gia_ManRegNum(p)");
    // derive CNF
    if fUseOldCnf != 0 {
        pCnf = Cnf_DeriveGiaRemapped( pMiter );
    } else {
        pMiter = Jf_ManDeriveCnf( pTemp = pMiter, 0 );
        Gia_ManStop( pTemp );
        pCnf = (Cnf_Dat_t *)pMiter.pData; pMiter.pData = nil;
    }

    // collect positive literals
    vLits = Vec_IntAlloc( Gia_ManCoNum(p) );
    for  i = 0; i < Gia_ManRegNum(p); i++  {
        Vec_IntPush( vLits, Abc_Var2Lit(i, fEmpty) );
    }

    // iteratively compute a minimal M-inductive set of next-state functions
    if fEmpty != 0 {
        nLitsUsed = 0;
    } else {
        nLitsUsed = Vec_IntSize(vLits);
    }
    vUsed = Vec_IntAlloc( Vec_IntSize(vLits) );
    for {
        var fChanges int32;
        // derive SAT solver        
        pSat = Bmc_DeriveSolver( p, pMiter, pCnf, nFramesMax, nTimeOut, fVerbose );
//        sat_solver_bookmark( pSat );
        status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if status == l_Undef {
            fmt.Printf("Timeout reached after %d seconds.\n", nTimeOut);
            break;
        }
        if status == l_True {
            fmt.Printf("The problem is satisfiable (the current set is not M-inductive).\n");
            break;
        }
        assert(status == l_False, "status == l_False");
        // call analize_final
        nLits = sat_solver_final( pSat, &pLits );
        // mark used literals
        Vec_IntFill( vUsed, Vec_IntSize(vLits), 0 );
        for  i = 0; i < nLits; i++  {
            Vec_IntWriteEntry( vUsed, Abc_Lit2Var(pLits[i]), 1 );
        }

        // check if there are any positive unused
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
            assert(i == Abc_Lit2Var(Lit), "i == Abc_Lit2Var(Lit)");
            if Abc_LitIsCompl(Lit) {
                continue; }
            if Vec_IntEntry(vUsed, i) {
                continue; }
            // positive literal became unused
            Vec_IntWriteEntry( vLits, i, Abc_LitNot(Lit) );
            nLitsUsed--;
            fChanges = 1;
        }
        // report the results
        if fVerbose != 0 {
        fmt.Printf("M =%4d :  AIG =%8d.  SAT vars =%8d.  SAT conf =%8d.  S =%6d. (%6.2f %%)  ", nFramesMax, (nFramesMax+1) * Gia_ManAndNum(pMiter), Gia_ManRegNum(p) + Gia_ManCoNum(p) + sat_solver_nvars(pSat), sat_solver_nconflicts(pSat), nLitsUsed, 100.0 * nLitsUsed / Gia_ManRegNum(p)); }
        if fVerbose != 0 {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart ); }
        // count the number of negative literals
        sat_solver_delete( pSat );
        if !(fChanges != 0) || (fEmpty != 0) {
            break; }
//        break;
//        sat_solver_rollback( pSat );
    }
    Cnf_DataFree( pCnf );
    Gia_ManStop( pMiter );
    Vec_IntFree( vLits );
    Vec_IntFree( vUsed );
}

/**Function*************************************************************

  Synopsis    [Collect flops starting from the POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_PerformFindFlopOrder_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vRegs *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        if Gia_ObjIsRo(p, pObj) {
            Vec_IntPush( vRegs, Gia_ObjId(p, pObj) ); }
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Bmc_PerformFindFlopOrder_rec( p, Gia_ObjFanin0(pObj), vRegs );
    Bmc_PerformFindFlopOrder_rec( p, Gia_ObjFanin1(pObj), vRegs );
}
func Bmc_PerformFindFlopOrder(p *Gia_Man_t, vRegs *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i, iReg, k int32;
    // start with POs
    Vec_IntClear( vRegs );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_IntPush( vRegs, Gia_ObjId(p, pObj) );
    }
    // add flop outputs in the B
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vRegs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vRegs,i))); i++  {
        assert(Gia_ObjIsPo(p, pObj) || Gia_ObjIsRo(p, pObj), "Gia_ObjIsPo(p, pObj) || Gia_ObjIsRo(p, pObj)");
        if Gia_ObjIsRo(p, pObj) {
            pObj = Gia_ObjRoToRi( p, pObj ); }
        Bmc_PerformFindFlopOrder_rec( p, Gia_ObjFanin0(pObj), vRegs );
    }
    // add dangling flops
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if !Gia_ObjIsTravIdCurrent(p, pObj) {
            Vec_IntPush( vRegs, Gia_ObjId(p, pObj) ); }
    }
    // remove POs; keep flop outputs only; remap ObjId into CiId
    assert(Vec_IntSize(vRegs) == Gia_ManCoNum(p), "Vec_IntSize(vRegs) == Gia_ManCoNum(p)");
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vRegs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vRegs,i))); i++  {
        if i < Gia_ManPoNum(p) {
            continue; }
        iReg = Gia_ObjCioId(pObj) - Gia_ManPiNum(p);
        assert(iReg >= 0 && iReg < Gia_ManRegNum(p), "iReg >= 0 && iReg < Gia_ManRegNum(p)");
        Vec_IntWriteEntry( vRegs, k++, iReg );
    }
    Vec_IntShrink( vRegs, k );
    assert(Vec_IntSize(vRegs) == Gia_ManRegNum(p), "Vec_IntSize(vRegs) == Gia_ManRegNum(p)");
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_PerformISearchOne(p *Gia_Man_t, nFramesMax int32, nTimeOut int32, fReverse int32, fBackTopo int32, fVerbose int32, vLits *Vec_Int_t) int32 {
    var fUseOldCnf int32;
    var pMiter, pTemp *Gia_Man_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var vRegs *Vec_Int_t  = nil;
//    Vec_Int_t * vLits;
    var i, Iter, status int32;
    var nLitsUsed, RetValue int32;
    var clkStart abctime  = Abc_Clock();
    assert(nFramesMax > 0, "nFramesMax > 0");
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");

    // create miter
    pTemp = Gia_ManDup( p );
    pMiter = Gia_ManMiter( p, pTemp, 0, 1, 1, 0, 0 );
    Gia_ManStop( pTemp );
    assert(Gia_ManPoNum(pMiter)  == 2 * Gia_ManPoNum(p), "Gia_ManPoNum(pMiter)  == 2 * Gia_ManPoNum(p)");
    assert(Gia_ManRegNum(pMiter) == 2 * Gia_ManRegNum(p), "Gia_ManRegNum(pMiter) == 2 * Gia_ManRegNum(p)");
    // derive CNF
    if fUseOldCnf != 0 {
        pCnf = Cnf_DeriveGiaRemapped( pMiter );
    } else {
        //pMiter = Jf_ManDeriveCnf( pTemp = pMiter, 0 );
        //Gia_ManStop( pTemp );
        //pCnf = (Cnf_Dat_t *)pMiter->pData; pMiter->pData = NULL;
        pCnf = (Cnf_Dat_t *)Mf_ManGenerateCnf( pMiter, 8, 0, 0, 0, 0 );
    }
/*
    // collect positive literals
    vLits = Vec_IntAlloc( Gia_ManCoNum(p) );
    for ( i = 0; i < Gia_ManRegNum(p); i++ )
        Vec_IntPush( vLits, Abc_Var2Lit(i, 0) );
*/
    // derive SAT solver        
    pSat = Bmc_DeriveSolver( p, pMiter, pCnf, nFramesMax, nTimeOut, fVerbose );
    status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if status == l_True {
        fmt.Printf("I = %4d :  ", nFramesMax);
        fmt.Printf("Problem is satisfiable.\n");
        sat_solver_delete( pSat );
        Cnf_DataFree( pCnf );
        Gia_ManStop( pMiter );
        return 1;
    }
    if status == l_Undef {
        fmt.Printf("ICheck: Timeout reached after %d seconds.                                                                          \n", nTimeOut);
        RetValue = 1;
        goto cleanup;
    }
    assert(status == l_False, "status == l_False");

    // count the number of positive literals
    nLitsUsed = 0;
    for  i = 0; i < Gia_ManRegNum(p); i++  {
        if !Abc_LitIsCompl(Vec_IntEntry(vLits, i)) {
            nLitsUsed++; }
    }

    // try removing variables
    vRegs = Vec_IntStartNatural( Gia_ManRegNum(p) );
    if fBackTopo != 0 {
        Bmc_PerformFindFlopOrder( p, vRegs ); }
    if fReverse != 0 {
        Vec_IntReverseOrder( vRegs ); }
//    for ( Iter = 0; Iter < Gia_ManRegNum(p); Iter++ )
    // inlining of macro Vec_IntForEachEntry
    for  Iter = 0; (Iter < Vec_IntSize(vRegs)) && (((i) = Vec_IntEntry(vRegs, Iter)), 1); Iter++  {
//        i = fReverse ? Gia_ManRegNum(p) - 1 - Iter : Iter;
        if Abc_LitIsCompl(Vec_IntEntry(vLits, i)) {
            continue; }
        Vec_IntWriteEntry( vLits, i, Abc_LitNot(Vec_IntEntry(vLits, i)) );
        status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if status == l_Undef {
            fmt.Printf("ICheck: Timeout reached after %d seconds.                                                                          \n", nTimeOut);
            RetValue = 1;
            goto cleanup;
        }
        if status == l_True {
            Vec_IntWriteEntry( vLits, i, Abc_LitNot(Vec_IntEntry(vLits, i)) );
        } else if status == l_False {
            nLitsUsed--;
        } else { assert(false, "0"); }
        // report the results
        //printf( "Round %d:  ", o );
        if fVerbose != 0 {
            fmt.Printf("I = %4d :  AIG =%8d.  SAT vars =%8d.  SAT conf =%8d.  S =%6d. (%6.2f %%)  ", i, (nFramesMax+1) * Gia_ManAndNum(pMiter), Gia_ManRegNum(p) + Gia_ManCoNum(p) + sat_solver_nvars(pSat), sat_solver_nconflicts(pSat), nLitsUsed, 100.0 * nLitsUsed / Gia_ManRegNum(p));
            // inlining of macro ABC_PRTr
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\r", 1.0*(float64((Abc_Clock() - clkStart)))/(float64(int64(time.Second))));;
            fflush( os.Stdout );
        }
    }
    // report the results
    //printf( "Round %d:  ", o );
    if fVerbose != 0 {
        fmt.Printf("M = %4d :  AIG =%8d.  SAT vars =%8d.  SAT conf =%8d.  S =%6d. (%6.2f %%)  ", nFramesMax, (nFramesMax+1) * Gia_ManAndNum(pMiter), Gia_ManRegNum(p) + Gia_ManCoNum(p) + sat_solver_nvars(pSat), sat_solver_nconflicts(pSat), nLitsUsed, 100.0 * nLitsUsed / Gia_ManRegNum(p));
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
        fflush( os.Stdout );
    }
cleanup:
    // cleanup
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Gia_ManStop( pMiter );
    Vec_IntFree( vRegs );
//    Vec_IntFree( vLits );
    return RetValue;
}
func Bmc_PerformISearch(p *Gia_Man_t, nFramesMax int32, nTimeOut int32, fReverse int32, fBackTopo int32, fDump int32, fVerbose int32) *Vec_Int_t {
    var vLits, vFlops *Vec_Int_t;
    var i, f int32;
    if fVerbose != 0 {
    fmt.Printf("Solving M-inductiveness for design %s with %d AND nodes and %d flip-flops with %s %s flop order:\n", Gia_ManName(p), Gia_ManAndNum(p), Gia_ManRegNum(p), fReverse ? "reverse":"direct", fBackTopo ? "backward":"natural"); }
    fflush( os.Stdout );

    // collect positive literals
    vLits = Vec_IntAlloc( Gia_ManCoNum(p) );
    for  i = 0; i < Gia_ManRegNum(p); i++  {
        Vec_IntPush( vLits, Abc_Var2Lit(i, 0) );
    }

    for  f = 1; f <= nFramesMax; f++  {
        if Bmc_PerformISearchOne( p, f, nTimeOut, fReverse, fBackTopo, fVerbose, vLits ) != 0 {
            Vec_IntFree( vLits );
            return nil;
        }
    }

    // dump the numbers of the flops
    if fDump != 0 {
        var nLitsUsed int32;
        for  i = 0; i < Gia_ManRegNum(p); i++  {
            if !Abc_LitIsCompl(Vec_IntEntry(vLits, i)) {
                nLitsUsed++; }
        }
        fmt.Printf("The set contains %d (out of %d) next-state functions with 0-based numbers:\n", nLitsUsed, Gia_ManRegNum(p));
        for  i = 0; i < Gia_ManRegNum(p); i++  {
            if !Abc_LitIsCompl(Vec_IntEntry(vLits, i)) {
                fmt.Printf("%d ", i); }
        }
        fmt.Printf("\n");
    }       
    // save flop indexes
    vFlops = Vec_IntAlloc( Gia_ManRegNum(p) );
    for  i = 0; i < Gia_ManRegNum(p); i++  {
        if !Abc_LitIsCompl(Vec_IntEntry(vLits, i)) {
            Vec_IntPush( vFlops, 1 );
        } else {
            Vec_IntPush( vFlops, 0 ); }
    }
    Vec_IntFree( vLits );
    return vFlops;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

