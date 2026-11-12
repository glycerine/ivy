package saig

import "fmt"

/**CFile****************************************************************

  FileName    [saigRetMin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Min-area retiming for the AIG.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigRetMin.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//#include "opt/nwk/nwk.h"
//#include "sat/cnf/cnf.h"
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

  Synopsis    [Derives the initial state after backward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeInitState(p *Aig_Man_t) *Vec_Int_t {
    var nConfLimit int32  = 1000000;
    var vCiIds *Vec_Int_t;
    var vInit *Vec_Int_t = nil;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var pObj *Aig_Obj_t;
    var i int32;
    var RetValue int32;
    var pModel *int32;
    // solve the SAT problem
    pCnf = Cnf_DeriveSimpleForRetiming( p );
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    if pSat == nil {
        Cnf_DataFree( pCnf );
        return nil;
    }
    RetValue = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    assert(RetValue != l_Undef, "RetValue != l_Undef");
    // create counter-example
    if RetValue == l_True {
        // accumulate SAT variables of the CIs
        vCiIds = Vec_IntAlloc( Aig_ManCiNum(p) );
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            Vec_IntPush( vCiIds, pCnf.pVarNums[pObj.Id] );
        }
        // create the model
        pModel = Sat_SolverGetModel( pSat, vCiIds.pArray, vCiIds.nSize );
        vInit = Vec_IntAllocArray( pModel, Aig_ManCiNum(p) );
        Vec_IntFree( vCiIds );
    }
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    return vInit;
}

/**Function*************************************************************

  Synopsis    [Uses UNSAT core to find the part of AIG to be excluded.]

  Description [Returns the number of the PO that appears in the UNSAT core.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeUnsatCore(p *Aig_Man_t, fVerbose int32) int32 {
    var fVeryVerbose int32;
    var nConfLimit int32  = 1000000;
    var pSatCnf any  = nil; 
    var pManProof *Intp_Man_t;
    var vCore *Vec_Int_t  = nil;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var pObj *Aig_Obj_t;
    var pClause1, pClause2, pLit, pVars *int32;
    var i, RetValue, iBadPo, iClause, nVars, nPos int32;
    // create the SAT solver
    pCnf = Cnf_DeriveSimpleForRetiming( p );
    pSat = sat_solver_new();
    sat_solver_store_alloc( pSat ); 
    sat_solver_setnvars( pSat, pCnf.nVars );
    for  i = 0; i < pCnf.nClauses; i++  {
        if !sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1] ) {
            Cnf_DataFree( pCnf );
            sat_solver_delete( pSat );
            return -1;
        }
    }
    sat_solver_store_mark_roots( pSat ); 
    // solve the problem
    RetValue = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    assert(RetValue != l_Undef, "RetValue != l_Undef");
    assert(RetValue == l_False, "RetValue == l_False");
    pSatCnf = sat_solver_store_release( pSat ); 
    sat_solver_delete( pSat );
    // derive the UNSAT core
    pManProof = Intp_ManAlloc();
    vCore = (Vec_Int_t *)Intp_ManUnsatCore( pManProof, (Sto_Man_t *)pSatCnf, 0, fVeryVerbose );
    Intp_ManFree( pManProof );
    Sto_ManFree( (Sto_Man_t *)pSatCnf );
    // derive the set of variables on which the core depends
    // collect the variable numbers
    nVars = 0;
    // inlining of macro ABC_ALLOC
    pVars = ((int *) malloc(sizeof(int) * (size_t)(pCnf.nVars)));
    memset( pVars, 0, sizeof(int) * pCnf.nVars );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCore)) && (((iClause) = Vec_IntEntry(vCore, i)), 1); i++  {
        pClause1 = pCnf.pClauses[iClause];
        pClause2 = pCnf.pClauses[iClause+1];
        for  pLit = pClause1; pLit < pClause2; pLit++  {
            if pVars[ (*pLit) >> 1 ] == 0 {
                nVars++; }
            pVars[ (*pLit) >> 1 ] = 1;
            if fVeryVerbose != 0 {
            fmt.Printf("%s%d ", ((*pLit) & 1)? "-" : "+", (*pLit) >> 1); }
        }
        if fVeryVerbose != 0 {
        fmt.Printf("\n"); }
    }
    // collect the nodes
    if fVeryVerbose != 0 {
      // inlining of macro Aig_ManForEachObj
      // inlining of macro Vec_PtrForEachEntry
      for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else
          if pCnf.pVarNums[pObj.Id] >= 0 && pVars[ pCnf.pVarNums[pObj.Id] ] == 1 {
              Aig_ObjPrint( p, pObj );
              fmt.Printf("\n");
          }
      }
    }
    // pick the first PO in the list
    nPos = 0;
    iBadPo = -1;
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if pCnf.pVarNums[pObj.Id] >= 0 && pVars[ pCnf.pVarNums[pObj.Id] ] == 1 {
            if iBadPo == -1 {
                iBadPo = i; }
            nPos++;
        }
    }
    if fVerbose != 0 {
        fmt.Printf("UNSAT core: %d clauses, %d variables, %d POs.  ", Vec_IntSize(vCore), nVars, nPos); }
    pVars = nil;
    Vec_IntFree( vCore );
    Cnf_DataFree( pCnf );
    return iBadPo;
}

/**Function*************************************************************

  Synopsis    [Marks the TFI cone with the current traversal ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManMarkCone_rec(p *Aig_Man_t, pObj *Aig_Obj_t) {
    if pObj == nil {
        return; }
    if Aig_ObjIsTravIdCurrent( p, pObj ) {
        return; }
    Aig_ObjSetTravIdCurrent( p, pObj );
    Saig_ManMarkCone_rec( p, Aig_ObjFanin0(pObj) );
    Saig_ManMarkCone_rec( p, Aig_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Counts the number of nodes to get registers after retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeCountCut(p *Aig_Man_t, vCut *Vec_Ptr_t) int32 {
    var vNodes *Vec_Ptr_t;
    var pObj, pFanin *Aig_Obj_t;
    var i, RetValue int32;
    // mark the cones
    Aig_ManIncrementTravId( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        Saig_ManMarkCone_rec( p, pObj );
    }
    // collect the new cut
    vNodes = Vec_PtrAlloc( 1000 );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsTravIdCurrent(p, pObj) {
            continue; }
        pFanin = Aig_ObjFanin0( pObj );
        if (pFanin != nil) && !pFanin.fMarkA && Aig_ObjIsTravIdCurrent(p, pFanin) {
            Vec_PtrPush( vNodes, pFanin );
            pFanin.fMarkA = 1;
        }
        pFanin = Aig_ObjFanin1( pObj );
        if (pFanin != nil) && !pFanin.fMarkA && Aig_ObjIsTravIdCurrent(p, pFanin) {
            Vec_PtrPush( vNodes, pFanin );
            pFanin.fMarkA = 1;
        }
    }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.fMarkA = 0;
    }
    RetValue = Vec_PtrSize( vNodes );
    Vec_PtrFree( vNodes );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeDup_rec(pNew *Aig_Man_t, pObj *Aig_Obj_t) {
    if pObj.pData {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Saig_ManRetimeDup_rec( pNew, Aig_ObjFanin0(pObj) );
    Saig_ManRetimeDup_rec( pNew, Aig_ObjFanin1(pObj) );
    pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG while retiming the registers to the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeDupForward(p *Aig_Man_t, vCut *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    // mark the cones under the cut
//    assert( Vec_PtrSize(vCut) == Saig_ManRetimeCountCut(p, vCut) );
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nRegs = Vec_PtrSize(vCut);
    pNew.nTruePis = p.nTruePis;
    pNew.nTruePos = p.nTruePos;
    // create the true PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // create the registers
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        pObj.pData = Aig_NotCond( Aig_ObjCreateCi(pNew), pObj.fPhase );
    }
    // duplicate logic above the cut
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Saig_ManRetimeDup_rec( pNew, Aig_ObjFanin0(pObj) );
    }
    // create the true POs
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    // remember value in LI
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        pObj.pData = Aig_ObjChild0Copy(pObj);
    }
    // transfer values from the LIs to the LOs
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
        pObjLo.pData = pObjLi.pData;
    }
    // erase the data values on the internal nodes of the cut
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        if Aig_ObjIsNode(pObj) {
            pObj.pData = nil; }
    }
    // duplicate logic below the cut
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        Saig_ManRetimeDup_rec( pNew, pObj );
        Aig_ObjCreateCo( pNew, Aig_NotCond((Aig_Obj_t *)pObj.pData, pObj.fPhase) );
    }
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG while retiming the registers to the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeDupBackward(p *Aig_Man_t, vCut *Vec_Ptr_t, vInit *Vec_Int_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    // mark the cones under the cut
//    assert( Vec_PtrSize(vCut) == Saig_ManRetimeCountCut(p, vCut) );
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nRegs = Vec_PtrSize(vCut);
    pNew.nTruePis = p.nTruePis;
    pNew.nTruePos = p.nTruePos;
    // create the true PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // create the registers
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        pObj.pData = Aig_NotCond( Aig_ObjCreateCi(pNew), vInit?Vec_IntEntry(vInit,i):0 );
    }
    // duplicate logic above the cut and remember values
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        Saig_ManRetimeDup_rec( pNew, Aig_ObjFanin0(pObj) );
        pObj.pData = Aig_ObjChild0Copy(pObj);
    }
    // transfer values from the LIs to the LOs
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
        pObjLo.pData = pObjLi.pData;
    }
    // erase the data values on the internal nodes of the cut
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        if Aig_ObjIsNode(pObj) {
            pObj.pData = nil; }
    }
    // replicate the data on the constant node and the PIs
    pObj = Aig_ManConst1(p);
    pObj.pData = Aig_ManConst1(pNew);
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Aig_ManCi( pNew, i );
    }
    // duplicate logic below the cut
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Saig_ManRetimeDup_rec( pNew, Aig_ObjFanin0(pObj) );
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        Saig_ManRetimeDup_rec( pNew, pObj );
        Aig_ObjCreateCo( pNew, Aig_NotCond((Aig_Obj_t *)pObj.pData, vInit?Vec_IntEntry(vInit,i):0) );
    }
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derives AIG for the initial state computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeDupInitState(p *Aig_Man_t, vCut *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // mark the cones under the cut
//    assert( Vec_PtrSize(vCut) == Saig_ManRetimeCountCut(p, vCut) );
    // create the new manager
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    // create the true PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    // create the registers
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi(pNew);
    }
    // duplicate logic above the cut and create POs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        Saig_ManRetimeDup_rec( pNew, Aig_ObjFanin0(pObj) );
        Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Returns the array of bad registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManGetRegistersToExclude(p *Aig_Man_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj, pFanin *Aig_Obj_t;
    var i, Diffs int32;
    assert(Saig_ManRegNum(p) > 0, "Saig_ManRegNum(p) > 0");
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        if !Aig_ObjFaninC0(pObj) {
            pFanin.fMarkA = 1;
        } else {
            pFanin.fMarkB = 1; }
    }
    Diffs = 0;
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        Diffs += pFanin.fMarkA && pFanin.fMarkB;
    }
    vNodes = Vec_PtrAlloc( 100 );
    if Diffs > 0 {
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
            pFanin = Aig_ObjFanin0(pObj);
            if pFanin.fMarkA && pFanin.fMarkB {
                Vec_PtrPush( vNodes, pObj ); }
        }
        assert(Vec_PtrSize(vNodes) == Diffs, "Vec_PtrSize(vNodes) == Diffs");
    }
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        pFanin = Aig_ObjFanin0(pObj);
        pFanin.fMarkA = pFanin.fMarkB = 0;
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Hides the registers that cannot be backward retimed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManHideBadRegs(p *Aig_Man_t, vBadRegs *Vec_Ptr_t) int32 {
    var vPisNew, vPosNew *Vec_Ptr_t;
    var pObjLi, pObjLo *Aig_Obj_t;
    var nTruePi, nTruePo, nBadRegs, i int32;
    if Vec_PtrSize(vBadRegs) == 0 {
        return 0; }
    // attached LOs to LIs
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
        pObjLi.pData = pObjLo;
    }
    // reorder them by putting bad registers first
    vPisNew = Vec_PtrDup( p.vCis );
    vPosNew = Vec_PtrDup( p.vCos );
    nTruePi = Aig_ManCiNum(p) - Aig_ManRegNum(p);
    nTruePo = Aig_ManCoNum(p) - Aig_ManRegNum(p);
    assert(nTruePi == p.nTruePis, "nTruePi == p.nTruePis");
    assert(nTruePo == p.nTruePos, "nTruePo == p.nTruePos");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vBadRegs)) && (((pObjLi) = (Aig_Obj_t *)Vec_PtrEntry(vBadRegs, i)), 1); i++  {
        Vec_PtrWriteEntry( vPisNew, nTruePi++, pObjLi.pData );
        Vec_PtrWriteEntry( vPosNew, nTruePo++, pObjLi );
        pObjLi.fMarkA = 1;
    }
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
        if pObjLi.fMarkA {
            pObjLi.fMarkA = 0;
            continue;
        }
        Vec_PtrWriteEntry( vPisNew, nTruePi++, pObjLo );
        Vec_PtrWriteEntry( vPosNew, nTruePo++, pObjLi );
    }
    // check the sizes
    assert(nTruePi == Aig_ManCiNum(p), "nTruePi == Aig_ManCiNum(p)");
    assert(nTruePo == Aig_ManCoNum(p), "nTruePo == Aig_ManCoNum(p)");
    // transfer the arrays
    Vec_PtrFree( p.vCis ); p.vCis = vPisNew;
    Vec_PtrFree( p.vCos ); p.vCos = vPosNew;
    // update the PIs
    nBadRegs = Vec_PtrSize(vBadRegs);
    p.nRegs -= nBadRegs;
    p.nTruePis += nBadRegs;
    p.nTruePos += nBadRegs;
    return nBadRegs;
}

/**Function*************************************************************

  Synopsis    [Exposes bad registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManExposeBadRegs(p *Aig_Man_t, nBadRegs int32) {
    p.nRegs += nBadRegs;
    p.nTruePis -= nBadRegs;
    p.nTruePos -= nBadRegs;
}

/**Function*************************************************************

  Synopsis    [Performs min-area retiming backward with initial state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeMinAreaBackward(pNew *Aig_Man_t, fVerbose int32) *Aig_Man_t {
    var pInit, pFinal *Aig_Man_t;
    var vBadRegs, vCut *Vec_Ptr_t;
    var vInit *Vec_Int_t;
    var iBadReg int32;
    // transform the AIG to have no bad registers
    vBadRegs = Saig_ManGetRegistersToExclude( pNew );
    if (fVerbose != 0) && Vec_PtrSize(vBadRegs) {
        fmt.Printf("Excluding %d registers that cannot be backward retimed.\n", Vec_PtrSize(vBadRegs)); } 
    for {
        Saig_ManHideBadRegs( pNew, vBadRegs );
        Vec_PtrFree( vBadRegs );
        // compute cut
        vCut = Nwk_ManDeriveRetimingCut( pNew, 0, fVerbose );
        if Vec_PtrSize(vCut) >= Aig_ManRegNum(pNew) {
            Vec_PtrFree( vCut );
            return nil;
        }
        // derive the initial state
        pInit = Saig_ManRetimeDupInitState( pNew, vCut );
        vInit = Saig_ManRetimeInitState( pInit );
        if vInit != nil {
            pFinal = Saig_ManRetimeDupBackward( pNew, vCut, vInit );
            Vec_IntFree( vInit );
            Vec_PtrFree( vCut );
            Aig_ManStop( pInit );
            return pFinal;
        }
        Vec_PtrFree( vCut );
        // there is no initial state - find the offending output
        iBadReg = Saig_ManRetimeUnsatCore( pInit, fVerbose );
        Aig_ManStop( pInit );
        if fVerbose != 0 {
            fmt.Printf("Excluding register %d.\n", iBadReg); } 
        // prepare to remove this output
        vBadRegs = Vec_PtrAlloc( 1 );
        Vec_PtrPush( vBadRegs, Aig_ManCo( pNew, Saig_ManPoNum(pNew) + iBadReg ) );
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Performs min-area retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManRetimeMinArea(p *Aig_Man_t, nMaxIters int32, fForwardOnly int32, fBackwardOnly int32, fInitial int32, fVerbose int32) *Aig_Man_t {
    var vCut *Vec_Ptr_t;
    var pNew, pTemp, pCopy *Aig_Man_t;
    var i, fChanges int32;
    pNew = Aig_ManDupSimple( p );
    // perform several iterations of forward retiming
    fChanges = 0;
    if !(fBackwardOnly != 0) {
    for  i = 0; i < nMaxIters; i++  {
        if Saig_ManRegNum(pNew) == 0 {
            break; }
        vCut = Nwk_ManDeriveRetimingCut( pNew, 1, fVerbose );
        if Vec_PtrSize(vCut) >= Aig_ManRegNum(pNew) {
            if (fVerbose != 0) && !(fChanges != 0) {
                fmt.Printf("Forward retiming cannot reduce registers.\n"); }
            Vec_PtrFree( vCut );
            break;
        }
        pNew = Saig_ManRetimeDupForward( pTemp = pNew, vCut );
        Aig_ManStop( pTemp );
        Vec_PtrFree( vCut );
        if fVerbose != 0 {
            Aig_ManReportImprovement( p, pNew ); }
        fChanges = 1;
    } }
    // perform several iterations of backward retiming
    fChanges = 0;
    if !(fForwardOnly != 0) && !(fInitial != 0) {
    for  i = 0; i < nMaxIters; i++  {
        if Saig_ManRegNum(pNew) == 0 {
            break; }
        vCut = Nwk_ManDeriveRetimingCut( pNew, 0, fVerbose );
        if Vec_PtrSize(vCut) >= Aig_ManRegNum(pNew) {
            if (fVerbose != 0) && !(fChanges != 0) {
                fmt.Printf("Backward retiming cannot reduce registers.\n"); }
            Vec_PtrFree( vCut );
            break;
        }
        pNew = Saig_ManRetimeDupBackward( pTemp = pNew, vCut, nil );
        Aig_ManStop( pTemp );
        Vec_PtrFree( vCut );
        if fVerbose != 0 {
            Aig_ManReportImprovement( p, pNew ); }
        fChanges = 1;
    }
    } else if !(fForwardOnly != 0) && (fInitial != 0) {
    for  i = 0; i < nMaxIters; i++  {
        if Saig_ManRegNum(pNew) == 0 {
            break; }
        pCopy = Aig_ManDupSimple( pNew );
        pTemp = Saig_ManRetimeMinAreaBackward( pCopy, fVerbose );
        Aig_ManStop( pCopy );
        if pTemp == nil {
            if (fVerbose != 0) && !(fChanges != 0) {
                fmt.Printf("Backward retiming cannot reduce registers.\n"); }
            break;
        }
        Saig_ManExposeBadRegs( pTemp, Saig_ManPoNum(pTemp) - Saig_ManPoNum(pNew) );
        Aig_ManStop( pNew );
        pNew = pTemp;
        if fVerbose != 0 {
            Aig_ManReportImprovement( p, pNew ); }
        fChanges = 1;
    } }
    if !(fForwardOnly != 0) && !(fInitial != 0) && (fChanges != 0) {
        fmt.Printf("Assuming const-0 init-state after backward retiming. Result will not verify.\n"); }
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

