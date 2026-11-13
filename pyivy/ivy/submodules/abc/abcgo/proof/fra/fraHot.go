package fra

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [fraHot.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [Computing and using one-hotness conditions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraHot.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////





////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns 1 if simulation info is composed of all zeros.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if simulation infos are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns 1 if implications holds.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes one-hot implications.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_OneHotCompute(p *Fra_Man_t, pSim *Fra_Sml_t) *Vec_Int_t {
    var fSkipConstEqu int32  = 1;
    var vOneHots *Vec_Int_t;
    var pObj1, pObj2 *Aig_Obj_t;
    var i, k int32;
    var nTruePis int32  = Aig_ManCiNum(pSim.pAig) - Aig_ManRegNum(pSim.pAig);
    assert(pSim.pAig == p.pManAig, "pSim.pAig == p.pManAig");
    vOneHots = Vec_IntAlloc( 100 );
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(pSim.pAig)-Aig_ManRegNum(pSim.pAig); (i < Vec_PtrSize(pSim.pAig.vCis)) && (((pObj1) = (Aig_Obj_t *)Vec_PtrEntry(pSim.pAig.vCis, i)), 1); i++  {
        if (fSkipConstEqu != 0) && (Fra_OneHotNodeIsConst(pSim, pObj1) != 0) {
            continue; }
        assert(i-nTruePis >= 0, "i-nTruePis >= 0");
//        Aig_ManForEachLoSeq( pSim->pAig, pObj2, k )
//        Vec_PtrForEachEntryStart( Aig_Obj_t *, pSim->pAig->vPis, pObj2, k, Aig_ManCiNum(p)-Aig_ManRegNum(p) )
        // inlining of macro Vec_PtrForEachEntryStart
        for  k = i+1; (k < Vec_PtrSize(pSim.pAig.vCis)) && (((pObj2) = (Aig_Obj_t *)Vec_PtrEntry(pSim.pAig.vCis, k)), 1); k++  {
            if (fSkipConstEqu != 0) && (Fra_OneHotNodeIsConst(pSim, pObj2) != 0) {
                continue; }
            if (fSkipConstEqu != 0) && (Fra_OneHotNodesAreEqual( pSim, pObj1, pObj2 ) != 0) {
                continue; }
            assert(k-nTruePis >= 0, "k-nTruePis >= 0");
            if Fra_OneHotNodesAreClause( pSim, pObj1, pObj2, 1, 1 ) != 0 {
                Vec_IntPush( vOneHots, Fra_RegToLit(i-nTruePis, 1) );
                Vec_IntPush( vOneHots, Fra_RegToLit(k-nTruePis, 1) );
                continue;
            }
            if Fra_OneHotNodesAreClause( pSim, pObj1, pObj2, 0, 1 ) != 0 {
                Vec_IntPush( vOneHots, Fra_RegToLit(i-nTruePis, 0) );
                Vec_IntPush( vOneHots, Fra_RegToLit(k-nTruePis, 1) );
                continue;
            }
            if Fra_OneHotNodesAreClause( pSim, pObj1, pObj2, 1, 0 ) != 0 {
                Vec_IntPush( vOneHots, Fra_RegToLit(i-nTruePis, 1) );
                Vec_IntPush( vOneHots, Fra_RegToLit(k-nTruePis, 0) );
                continue;
            }
        }
    }
    return vOneHots;
}

/**Function*************************************************************

  Synopsis    [Assumes one-hot implications in the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

**********************************************************************/
func Fra_OneHotAssume(p *Fra_Man_t, vOneHots *Vec_Int_t) {
    var pObj1, pObj2 *Aig_Obj_t;
    var i int32;
    var Out1 int32;
    var Out2 int32;
    var pLits [2]int32;
    var nPiNum int32  = Aig_ManCiNum(p.pManFraig) - Aig_ManRegNum(p.pManFraig);
    assert(p.pPars.nFramesK == 1, "p.pPars.nFramesK == 1"); // add to only one frame
    for  i = 0; i < Vec_IntSize(vOneHots); i += 2  {
        Out1 = Vec_IntEntry( vOneHots, i );
        Out2 = Vec_IntEntry( vOneHots, i+1 );
        if Out1 == 0 && Out2 == 0 {
            continue; }
        pObj1 = Aig_ManCi( p.pManFraig, nPiNum + Fra_LitReg(Out1) );
        pObj2 = Aig_ManCi( p.pManFraig, nPiNum + Fra_LitReg(Out2) );
        pLits[0] = toLitCond( Fra_ObjSatNum(pObj1), Fra_LitSign(Out1) );
        pLits[1] = toLitCond( Fra_ObjSatNum(pObj2), Fra_LitSign(Out2) );
        // add constraint to solver
        if !sat_solver_addclause( p.pSat, pLits, pLits + 2 ) {
            fmt.Printf("Fra_OneHotAssume(): Adding clause makes SAT solver unsat.\n");
            sat_solver_delete( p.pSat );
            p.pSat = nil;
            return;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Checks one-hot implications.]

  Description []
               
  SideEffects []

  SeeAlso     []

**********************************************************************/
func Fra_OneHotCheck(p *Fra_Man_t, vOneHots *Vec_Int_t) {
    var pObj1, pObj2 *Aig_Obj_t;
    var RetValue, i, Out1, Out2 int32;
    var nTruePos int32  = Aig_ManCoNum(p.pManFraig) - Aig_ManRegNum(p.pManFraig);
    for  i = 0; i < Vec_IntSize(vOneHots); i += 2  {
        Out1 = Vec_IntEntry( vOneHots, i );
        Out2 = Vec_IntEntry( vOneHots, i+1 );
        if Out1 == 0 && Out2 == 0 {
            continue; }
        pObj1 = Aig_ManCo( p.pManFraig, nTruePos + Fra_LitReg(Out1) );
        pObj2 = Aig_ManCo( p.pManFraig, nTruePos + Fra_LitReg(Out2) );
        RetValue = Fra_NodesAreClause( p, pObj1, pObj2, Fra_LitSign(Out1), Fra_LitSign(Out2) );
        if RetValue != 1 {
            p.pCla.fRefinement = 1;
            if RetValue == 0 {
                Fra_SmlResimulate( p ); }
            if Vec_IntEntry(vOneHots, i) != 0 {
                fmt.Printf("Fra_OneHotCheck(): Clause is not refined!\n"); }
            assert(Vec_IntEntry(vOneHots, i) == 0, "Vec_IntEntry(vOneHots, i) == 0");
        }
    }
}

/**Function*************************************************************

  Synopsis    [Removes those implications that no longer hold.]

  Description [Returns 1 if refinement has happened.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Removes those implications that no longer hold.]

  Description [Returns 1 if refinement has happened.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Estimates the coverage of state space by clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Creates one-hotness EXDC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_OneHotCreateExdc(p *Fra_Man_t, vOneHots *Vec_Int_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj1, pObj2, pObj *Aig_Obj_t;
    var i, Out1, Out2, nTruePis int32;
    pNew = Aig_ManStart( Vec_IntSize(vOneHots)/2 );
//    for ( i = 0; i < Aig_ManRegNum(p->pManAig); i++ )
//        Aig_ObjCreateCi(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pManAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pManAig.vCis, i)), 1); i++  {
        Aig_ObjCreateCi(pNew);
    }
    nTruePis = Aig_ManCiNum(p.pManAig) - Aig_ManRegNum(p.pManAig);
    for  i = 0; i < Vec_IntSize(vOneHots); i += 2  {
        Out1 = Vec_IntEntry( vOneHots, i );
        Out2 = Vec_IntEntry( vOneHots, i+1 );
        if Out1 == 0 && Out2 == 0 {
            continue; }
        pObj1 = Aig_ManCi( pNew, nTruePis + Fra_LitReg(Out1) );
        pObj2 = Aig_ManCi( pNew, nTruePis + Fra_LitReg(Out2) );
        pObj1 = Aig_NotCond( pObj1, Fra_LitSign(Out1) );
        pObj2 = Aig_NotCond( pObj2, Fra_LitSign(Out2) );
        pObj  = Aig_Or( pNew, pObj1, pObj2 );
        Aig_ObjCreateCo( pNew, pObj );
    }
    Aig_ManCleanup(pNew);
//    printf( "Created AIG with %d nodes and %d outputs.\n", Aig_ManNodeNum(pNew), Aig_ManCoNum(pNew) );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Assumes one-hot implications in the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

**********************************************************************/
func Fra_OneHotAddKnownConstraint(p *Fra_Man_t, vOnehots *Vec_Ptr_t) {
    var vGroup *Vec_Int_t;
    var pObj1, pObj2 *Aig_Obj_t;
    var k int32;
    var i int32;
    var j int32;
    var Out1 int32;
    var Out2 int32;
    var pLits [2]int32;
    //
    // these constrants should be added to different timeframes!
    // (also note that PIs follow first - then registers)
    //
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vOnehots)) && (((vGroup) = (Vec_Int_t *)Vec_PtrEntry(vOnehots, k)), 1); k++  {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vGroup)) && (((Out1) = Vec_IntEntry(vGroup, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  j = i+1; (j < Vec_IntSize(vGroup)) && (((Out2) = Vec_IntEntry(vGroup, j)), 1); j++  {
            pObj1 = Aig_ManCi( p.pManFraig, Out1 );
            pObj2 = Aig_ManCi( p.pManFraig, Out2 );
            pLits[0] = toLitCond( Fra_ObjSatNum(pObj1), 1 );
            pLits[1] = toLitCond( Fra_ObjSatNum(pObj2), 1 );
            // add constraint to solver
            if !sat_solver_addclause( p.pSat, pLits, pLits + 2 ) {
                fmt.Printf("Fra_OneHotAddKnownConstraint(): Adding clause makes SAT solver unsat.\n");
                sat_solver_delete( p.pSat );
                p.pSat = nil;
                return;
            }
        }
        }
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

