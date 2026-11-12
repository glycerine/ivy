package saig

import "fmt"

/**CFile****************************************************************

  FileName    [saigOutDec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Output cone decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigOutDec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs decomposition of the property output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFindPrimes(pAig *Aig_Man_t, nLits int32, fVerbose int32) *Vec_Ptr_t {
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var pObj0, pObj1, pRoot, pMiter *Aig_Obj_t;
    var vPrimes, vNodes *Vec_Ptr_t;
    var vCube, vMarks *Vec_Int_t;
    var i0 int32;
    var i1 int32;
    var m int32;
    var RetValue int32;
    var Lits [10]int32;
    var fCompl0, fCompl1, nNodes1, nNodes2 int32;
    assert(nLits == 1 || nLits == 2, "nLits == 1 || nLits == 2");
    assert(nLits < 10, "nLits < 10");

    // create SAT solver
    pCnf = Cnf_DeriveSimple( pAig, Aig_ManCoNum(pAig) ); 
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );

    // collect nodes in the property output cone
    pMiter = Aig_ManCo( pAig, 0 );
    pRoot  = Aig_ObjFanin0( pMiter );
    vNodes = Aig_ManDfsNodes( pAig, &pRoot, 1 );
    // sort nodes by level and remove the last few
 
    // try single nodes
    vPrimes = Vec_PtrAlloc( 100 );
    // create assumptions
    vMarks = Vec_IntStart( Vec_PtrSize(vNodes) );
    Lits[0] = toLitCond( pCnf.pVarNums[Aig_ObjId(pMiter)], 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i0 = 0; (i0 < Vec_PtrSize(vNodes)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i0)), 1); i0++  {
    if pObj0 != pRoot {
        // create assumptions
        Lits[1] = toLitCond( pCnf.pVarNums[Aig_ObjId(pObj0)], pObj0.fPhase );
        // solve the problem
        RetValue = sat_solver_solve( pSat, Lits, Lits+2, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
        if RetValue == l_False {
            vCube = Vec_IntAlloc( 1 );
            Vec_IntPush( vCube, toLitCond(Aig_ObjId(pObj0), pObj0.fPhase) );
            Vec_PtrPush( vPrimes, vCube );
            if fVerbose != 0 {
            fmt.Printf("Adding prime %d%c\n", Aig_ObjId(pObj0), pObj0.fPhase?'-':'+'); }
            Vec_IntWriteEntry( vMarks, i0, 1 );
        }
    }
    }
    nNodes1 = Vec_PtrSize(vPrimes);
    if nLits > 1 {
        // try adding second literal
        // inlining of macro Vec_PtrForEachEntry
        for  i0 = 0; (i0 < Vec_PtrSize(vNodes)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i0)), 1); i0++  {
        if pObj0 != pRoot {
        // inlining of macro Vec_PtrForEachEntryStart
        for  i1 = i0+1; (i1 < Vec_PtrSize(vNodes)) && (((pObj1) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i1)), 1); i1++  {
        if pObj1 != pRoot {
            if Vec_IntEntry(vMarks,i0) || Vec_IntEntry(vMarks,i1) {
                continue; }
            for  m = 0; m < 3; m++  {
                fCompl0 =  m & 1;
                fCompl1 = (m >> 1) & 1;
                // create assumptions
                Lits[1] = toLitCond( pCnf.pVarNums[Aig_ObjId(pObj0)], fCompl0 ^ pObj0.fPhase );
                Lits[2] = toLitCond( pCnf.pVarNums[Aig_ObjId(pObj1)], fCompl1 ^ pObj1.fPhase );
                // solve the problem
                RetValue = sat_solver_solve( pSat, Lits, Lits+3, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
                if RetValue == l_False {
                    vCube = Vec_IntAlloc( 2 );
                    Vec_IntPush( vCube, toLitCond(Aig_ObjId(pObj0), fCompl0 ^ pObj0.fPhase) );
                    Vec_IntPush( vCube, toLitCond(Aig_ObjId(pObj1), fCompl1 ^ pObj1.fPhase) );
                    Vec_PtrPush( vPrimes, vCube );
                    if fVerbose != 0 {
                    fmt.Printf("Adding prime %d%c %d%c\n", Aig_ObjId(pObj0), (fCompl0 ^ pObj0.fPhase)?'-':'+', Aig_ObjId(pObj1), (fCompl1 ^ pObj1.fPhase)?'-':'+'); }
                    break;
                }
            }
        }
        } }
        }
    }
    nNodes2 = Vec_PtrSize(vPrimes) - nNodes1;

    fmt.Printf("Property cone size = %6d    1-lit primes = %5d    2-lit primes = %5d\n", Vec_PtrSize(vNodes), nNodes1, nNodes2);

    // clean up
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Vec_PtrFree( vNodes );
    Vec_IntFree( vMarks );
    return vPrimes;
}

/**Function*************************************************************

  Synopsis    [Performs decomposition of the property output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManDecPropertyOutput(pAig *Aig_Man_t, nLits int32, fVerbose int32) *Aig_Man_t {
    var pAigNew *Aig_Man_t  = nil;
    var pObj, pMiter *Aig_Obj_t;
    var vPrimes *Vec_Ptr_t;
    var vCube *Vec_Int_t;
    var i, k, Lit int32;

    // compute primes of the comb output function
    vPrimes = Saig_ManFindPrimes( pAig, nLits, fVerbose );

    // start the new manager
    pAigNew = Aig_ManStart( Aig_ManNodeNum(pAig) );
    pAigNew.pName = Abc_UtilStrsav( pAig.pName );
    pAigNew.nConstrs = pAig.nConstrs;
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pAigNew );
    // create variables for PIs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pAigNew );
    }
    // add internal nodes of this frame
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pAigNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create original POs of the circuit
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }
    // create prime POs of the circuit
    if vPrimes != nil {
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vPrimes)) && (((vCube) = (Vec_Int_t *)Vec_PtrEntry(vPrimes, k)), 1); k++  {
        pMiter = Aig_ManConst1( pAigNew );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCube)) && (((Lit) = Vec_IntEntry(vCube, i)), 1); i++  {
            pObj = Aig_NotCond( Aig_ObjCopy(Aig_ManObj(pAig, Abc_Lit2Var(Lit))), Abc_LitIsCompl(Lit) );
            pMiter = Aig_And( pAigNew, pMiter, pObj );
        }
        Aig_ObjCreateCo( pAigNew, pMiter );
    } }
    // transfer to register outputs
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Aig_ObjCreateCo( pAigNew, Aig_ObjChild0Copy(pObj) );
    }
    Aig_ManCleanup( pAigNew );
    Aig_ManSetRegNum( pAigNew, Aig_ManRegNum(pAig) );

    Vec_VecFreeP( (Vec_Vec_t **)&vPrimes );
    return pAigNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

