package bmc

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [bmcBCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Performs recording of BMC core.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcBCore.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collect pivot variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_ManBCoreReadPivots(pName string) *Vec_Int_t {
    vPivots := Vec_IntAlloc(100)
    data, err := os.ReadFile(pName)
    if err != nil {
        return vPivots
    }
    for _, id := range parsePivotIDs(data) {
        Vec_IntPush(vPivots, id)
    }
    return vPivots;
}
func Bmc_ManBCoreCollectPivots(p *Gia_Man_t, pName string, vVarMap *Vec_Int_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var i, iVar int32;
    var vPivots *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vVars *Vec_Int_t  = Bmc_ManBCoreReadPivots( pName );
    // inlining of macro Gia_ManForEachObj
    for i = 0; i < p.nObjs; i++ {
        pObj = Gia_ManObj(p, i)
        pObj.fMark0 = 0;
    }
    // inlining of macro Vec_IntForEachEntry
    for i = 0; i < Vec_IntSize(vVars); i++ {
        iVar = Vec_IntEntry(vVars, i)
        if iVar > 0 && iVar < Gia_ManObjNum(p) {
            Gia_ManObj( p, iVar ).fMark0 = 1;
        } else {
            fmt.Printf("Cannot find object with Id %d in the given AIG.\n", iVar); }
    }
    // inlining of macro Vec_IntForEachEntryDouble
    for i = 0; i+1 < Vec_IntSize(vVarMap); i += 2 {
        iVar = Vec_IntEntry(vVarMap, i)
        _ = Vec_IntEntry(vVarMap, i+1) // each mapping entry is an (object, frame) pair
        if Gia_ManObj(p, iVar).fMark0 != 0 {
            Vec_IntPush( vPivots, Abc_Lit2Var(i) ); }
    }
    // inlining of macro Gia_ManForEachObj
    for i = 0; i < p.nObjs; i++ {
        pObj = Gia_ManObj(p, i)
        pObj.fMark0 = 0;
    }
    Vec_IntFree( vVars );
    return vPivots;
}

/**Function*************************************************************

  Synopsis    [Collect (Id; Frame) pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_ManBCoreAssignVar(p *Gia_Man_t, pObj *Gia_Obj_t, f int32, vNodes *Vec_Int_t) {
    pObj.Value = Abc_Lit2Var(Vec_IntSize(vNodes));
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
    Vec_IntPush( vNodes, f );
//    printf( "Obj %3d  Frame %3d  --->  Var %3d    ", Gia_ObjId(p, pObj), f, pObj->Value );
//    Gia_ObjPrint( p, pObj );
}
func Bmc_ManBCoreCollect_rec(p *Gia_Man_t, Id int32, f int32, vNodes *Vec_Int_t, vRootsNew *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, Id) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, Id);
    pObj = Gia_ManObj( p, Id );
    Bmc_ManBCoreAssignVar( p, pObj, f, vNodes );
    if Gia_ObjIsPi(p, pObj) {
        return; }
    if Gia_ObjIsRo(p, pObj) {
        Vec_IntPush( vRootsNew, Gia_ObjId(p, Gia_ObjRoToRi(p, pObj)) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Bmc_ManBCoreCollect_rec( p, Gia_ObjFaninId0p(p, pObj), f, vNodes, vRootsNew );
    Bmc_ManBCoreCollect_rec( p, Gia_ObjFaninId1p(p, pObj), f, vNodes, vRootsNew );
}
func Bmc_ManBCoreCollect(p *Gia_Man_t, iFrame int32, iOut int32, pSat *sat_solver) *Vec_Int_t {
    var pObj *Gia_Obj_t; 
    var f, i, iObj, nNodesOld int32;
    var vNodes *Vec_Int_t   = Vec_IntAlloc( 100 );
    var vRoots *Vec_Int_t   = Vec_IntAlloc( 100 );
    var vRoots2 *Vec_Int_t  = Vec_IntAlloc( 100 );
    assert(iFrame >= 0 && iOut >= 0, "iFrame >= 0 && iOut >= 0");
    // add first variables
    Vec_IntPush( vNodes, -1 );
    Vec_IntPush( vNodes, -1 );
    Bmc_ManBCoreAssignVar( p, Gia_ManPo(p, iOut), iFrame, vNodes );
    // start with root node
    Vec_IntPush( vRoots, Gia_ObjId(p, Gia_ManPo(p, iOut)) );
    // iterate through time frames
    for  f = iFrame; f >= 0; f--  {
        Gia_ManIncrementTravId( p );
        // add constant node
        Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
        Bmc_ManBCoreAssignVar( p, Gia_ManConst0(p), f, vNodes );
        sat_solver_add_const( pSat, Gia_ManConst0(p).Value, 1 );
        // collect nodes in this timeframe
        Vec_IntClear( vRoots2 );
        nNodesOld = Vec_IntSize(vNodes);
        // inlining of macro Gia_ManForEachObjVec
        for i = 0; i < Vec_IntSize(vRoots); i++ {
            pObj = Gia_ManObj(p, Vec_IntEntry(vRoots, i))
            Bmc_ManBCoreCollect_rec( p, Gia_ObjFaninId0p(p, pObj), f, vNodes, vRoots2 );
        }
        if f == iFrame {
            // add the final clause
            pObj = Gia_ManPo(p, iOut);
            assert(pObj.Value == 1, "pObj.Value == 1");
            assert(Gia_ObjFanin0(pObj).Value == 3, "Gia_ObjFanin0(pObj).Value == 3");
//            sat_solver_add_const( pSat, Gia_ObjFanin0(pObj)->Value, Gia_ObjFaninC0(pObj) );
            sat_solver_add_constraint( pSat, Gia_ObjFanin0(pObj).Value, pObj.Value, Gia_ObjFaninC0(pObj) );
        
        } else {
            // connect current RIs to previous ROs
            // inlining of macro Gia_ManForEachObjVec
            for i = 0; i < Vec_IntSize(vRoots); i++ {
                pObj = Gia_ManObj(p, Vec_IntEntry(vRoots, i))
                sat_solver_add_buffer( pSat, pObj.Value, Gia_ObjFanin0(pObj).Value, Gia_ObjFaninC0(pObj) );
            }
        }
        // inlining of macro Gia_ManForEachObjVec
        for i = 0; i < Vec_IntSize(vRoots2); i++ {
            pObj = Gia_ManObj(p, Vec_IntEntry(vRoots2, i))
            pObj.Value = Gia_ObjRiToRo(p, pObj).Value;
        }
        // add nodes of this timeframe
        // inlining of macro Vec_IntForEachEntryStart
        for i = nNodesOld; i < Vec_IntSize(vNodes); i += 2 {
            iObj = Vec_IntEntry(vNodes, i)
            pObj = Gia_ManObj(p, iObj)
            if Gia_ObjIsCi(pObj) {
                continue; }
            assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
            sat_solver_add_and( pSat, pObj.Value, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value, Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 );
        }
        // collect constant node
        // inlining of macro ABC_SWAP
        { var t *Vec_Int_t  = vRoots; vRoots = vRoots2; vRoots2 = t; };
    }
    // add constraint variables for the init state
    // inlining of macro Gia_ManForEachObjVec
    for i = 0; i < Vec_IntSize(vRoots); i++ {
        pObj = Gia_ManObj(p, Vec_IntEntry(vRoots, i))
        sat_solver_add_constraint( pSat, pObj.Value, Abc_Lit2Var(Vec_IntSize(vNodes)), 1 );
        pObj = Gia_ObjRiToRo(p, pObj);
        Bmc_ManBCoreAssignVar( p, pObj, -1, vNodes );
    }
    Vec_IntFree( vRoots );
    Vec_IntFree( vRoots2 );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_ManBCorePerform(p *Gia_Man_t, pPars *Bmc_BCorePar_t) {
    var clk int64  = time.Now().UnixNano();
    var pManProof *Intp_Man_t;
    var vVarMap, vCore *Vec_Int_t;
    var pSat *sat_solver;
    var pFile *os.File;
    var pSatCnf *Sto_Man_t
    var RetValue int32;
    // create SAT solver
    pSat = sat_solver_new();
    sat_solver_store_alloc( pSat ); 
    sat_solver_setnvars( pSat, 1000 );
    runtimeLimit := int64(0)
    if pPars.nTimeOut > 0 {
        runtimeLimit = int64(pPars.nTimeOut)*int64(time.Second) + Abc_Clock()
    }
    sat_solver_set_runtime_limit(pSat, runtimeLimit)
    vVarMap = Bmc_ManBCoreCollect( p, pPars.iFrame, pPars.iOutput, pSat );
    sat_solver_store_mark_roots( pSat ); 
    // create pivot variables
    if pPars.pFilePivots != "" {
        var vPivots *Vec_Int_t  = Bmc_ManBCoreCollectPivots(p, pPars.pFilePivots, vVarMap);
        sat_solver_set_pivot_variables( pSat, Vec_IntArray(vPivots), Vec_IntSize(vPivots) );
        Vec_IntReleaseArray( vPivots );
        Vec_IntFree( vPivots );
    }
    // solve the problem
    RetValue = sat_solver_solve(pSat, nil, 0, 0, 0, 0)
    if RetValue == l_Undef {
        Vec_IntFree( vVarMap );
        sat_solver_delete( pSat );
        fmt.Printf("Timeout of conflict limit is reached.\n");
        return;
    }
    if RetValue == l_True {
        Vec_IntFree( vVarMap );
        sat_solver_delete( pSat );
        fmt.Printf("The BMC problem is SAT.\n");
        return;
    }
    if pPars.fVerbose != 0 {
        fmt.Printf("SAT solver returned UNSAT after %7d conflicts.      ", int32(pSat.stats.conflicts));
        Abc_PrintTime( 1, "Time", time.Now().UnixNano() - clk );
    }
    assert(RetValue == l_False, "RetValue == l_False");
    released := sat_solver_store_release(pSat)
    pSatCnf, _ = released.(*Sto_Man_t)
//    Sto_ManDumpClauses( (Sto_Man_t *)pSatCnf, "cnf_store.txt" );
    // derive the UNSAT core
    clk = time.Now().UnixNano();
    pManProof = Intp_ManAlloc();
    core := Intp_ManUnsatCore(pManProof, pSatCnf, 1, pPars.fVerbose)
    vCore = Vec_IntAllocArrayCopy(core, len(core))
    Intp_ManFree( pManProof );
    if pPars.fVerbose != 0 {
        fmt.Printf("UNSAT core contains %d (out of %d) learned clauses.   ", Vec_IntSize(vCore), sat_solver_nconflicts(pSat));
        Abc_PrintTime( 1, "Time", time.Now().UnixNano() - clk );
    }
    // write the problem
    Vec_IntSort( vCore, 0 );
    if pPars.pFileProof != "" {
        var err error
        pFile, err = os.OpenFile(pPars.pFileProof, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
        if err != nil {
            Vec_IntFree(vCore)
            Vec_IntFree(vVarMap)
            sat_solver_delete(pSat)
            Sto_ManFree(pSatCnf)
            return
        }
    } else {
        pFile = os.Stdout;
    }
    Intp_ManUnsatCorePrintForBmc(pFile, pSatCnf, Vec_IntArray(vCore), Vec_IntArray(vVarMap))
    if pFile != os.Stdout {
        _ = (pFile).Close(); }
    // cleanup
    Sto_ManFree(pSatCnf)
    Vec_IntFree( vVarMap );
    Vec_IntFree( vCore );
    sat_solver_delete( pSat );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

