package ssw

/**CFile****************************************************************

  FileName    [sswSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [On-demand uniqueness constraints.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswSat.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs computation of signal correspondence with constraints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_UniqueRegisterPairInfo(p *Ssw_Man_t) {
    var pObjLo, pObj0, pObj1 *Aig_Obj_t;
    var i, RetValue, Counter int32;
    if p.vDiffPairs == nil {
        p.vDiffPairs = Vec_IntAlloc( Saig_ManRegNum(p.pAig) ); }
    Vec_IntClear( p.vDiffPairs );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLo) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        pObj0 = Ssw_ObjFrame( p, pObjLo, 0 );
        pObj1 = Ssw_ObjFrame( p, pObjLo, 1 );
        if pObj0 == pObj1 {
            Vec_IntPush( p.vDiffPairs, 0 );
        } else if pObj0 == Aig_Not(pObj1) {
            Vec_IntPush( p.vDiffPairs, 1 );
//        else
//            Vec_IntPush( p->vDiffPairs, 1 );
        } else if Aig_ObjPhaseReal(pObj0) != Aig_ObjPhaseReal(pObj1) {
            Vec_IntPush( p.vDiffPairs, 1 );
        } else {
            RetValue = Ssw_NodesAreEquiv( p, Aig_Regular(pObj0), Aig_Regular(pObj1) );
            Vec_IntPush( p.vDiffPairs, RetValue!=1 );
        }
    }
    assert(Vec_IntSize(p.vDiffPairs) == Saig_ManRegNum(p.pAig), "Vec_IntSize(p.vDiffPairs) == Saig_ManRegNum(p.pAig)");
    // count the number of ones
    Counter = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vDiffPairs)) && (((RetValue) = Vec_IntEntry(p.vDiffPairs, i)), 1); i++  {
        Counter += RetValue;
    }
//    Abc_Print( 1, "The number of different register pairs = %d.\n", Counter );
}


/**Function*************************************************************

  Synopsis    [Returns 1 if uniqueness constraints can be added.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManUniqueOne(p *Ssw_Man_t, pRepr *Aig_Obj_t, pObj *Aig_Obj_t, fVerbose int32) int32 {
    var ppObjs [2]*Aig_Obj_t;
    var pTemp *Aig_Obj_t;
    var i, k, Value0, Value1, RetValue, fFeasible int32;

    assert(p.pPars.nFramesK > 1, "p.pPars.nFramesK > 1");
    assert(p.vDiffPairs && Vec_IntSize(p.vDiffPairs) == Saig_ManRegNum(p.pAig), "p.vDiffPairs && Vec_IntSize(p.vDiffPairs) == Saig_ManRegNum(p.pAig)");

    // compute the first support in terms of LOs
    ppObjs[0] = pRepr;
    ppObjs[1] = pObj;
    Aig_SupportNodes( p.pAig, ppObjs, 2, p.vCommon );
    // keep only LOs
    RetValue = Vec_PtrSize( p.vCommon );
    fFeasible = 0;
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCommon)) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(p.vCommon, i)), 1); i++  {
        assert(Aig_ObjIsCi(pTemp), "Aig_ObjIsCi(pTemp)");
        if !Saig_ObjIsLo(p.pAig, pTemp) {
            continue; }
        assert(Aig_ObjCioId(pTemp) > 0, "Aig_ObjCioId(pTemp) > 0");
        Vec_PtrWriteEntry( p.vCommon, k++, pTemp );
        if Vec_IntEntry(p.vDiffPairs, Aig_ObjCioId(pTemp) - Saig_ManPiNum(p.pAig)) {
            fFeasible = 1; }
    }
    Vec_PtrShrink( p.vCommon, k );

    if fVerbose != 0 {
        Abc_Print( 1, "Node = %5d : Supp = %3d. Regs = %3d. Feasible = %s. ",
            Aig_ObjId(pObj), RetValue, Vec_PtrSize(p.vCommon),
            fFeasible? "yes": "no " ); }

    // check the current values
    RetValue = 1;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCommon)) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(p.vCommon, i)), 1); i++  {
        Value0 = Ssw_ManGetSatVarValue( p, pTemp, 0 );
        Value1 = Ssw_ManGetSatVarValue( p, pTemp, 1 );
        if Value0 != Value1 {
            RetValue = 0; }
        if fVerbose != 0 {
            Abc_Print( 1, "%d", Value0 ^ Value1 ); }
    }
    if fVerbose != 0 {
        Abc_Print( 1, "\n" ); }

    return RetValue && fFeasible;
}

/**Function*************************************************************

  Synopsis    [Returns the output of the uniqueness constraint.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManUniqueAddConstraint(p *Ssw_Man_t, vCommon *Vec_Ptr_t, f1 int32, f2 int32) int32 {
    var pObj, pObj1New, pObj2New, pMiter, pTotal *Aig_Obj_t;
    var i int32;
    var pLits [2]int32;
//    int RetValue;
    assert(Vec_PtrSize(vCommon) > 0, "Vec_PtrSize(vCommon) > 0");
    // generate the constraint
    pTotal = Aig_ManConst0(p.pFrames);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCommon)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCommon, i)), 1); i++  {
        assert(Saig_ObjIsLo(p.pAig, pObj), "Saig_ObjIsLo(p.pAig, pObj)");
        pObj1New = Ssw_ObjFrame( p, pObj, f1 );
        pObj2New = Ssw_ObjFrame( p, pObj, f2 );
        pMiter = Aig_Exor( p.pFrames, pObj1New, pObj2New );
        pTotal = Aig_Or( p.pFrames, pTotal, pMiter );
    }
    if Aig_ObjIsConst1(Aig_Regular(pTotal)) {
//        Abc_Print( 1, "Skipped\n" );
        return 0;
    }
    // create CNF
    Ssw_CnfNodeAddToSolver( p.pMSat, Aig_Regular(pTotal) );
    // add output constraint
    pLits[0] = toLitCond( Ssw_ObjSatNum(p.pMSat,Aig_Regular(pTotal)), Aig_IsComplement(pTotal) );
/*
    RetValue = sat_solver_addclause( p->pSat, pLits, pLits + 1 );
    assert( RetValue );
    // simplify the solver
    if ( p->pSat->qtail != p->pSat->qhead )
    {
        RetValue = sat_solver_simplify(p->pSat);
        assert( RetValue != 0 );
    }
*/
    assert(p.iOutputLit == -1, "p.iOutputLit == -1");
    p.iOutputLit = pLits[0];
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
