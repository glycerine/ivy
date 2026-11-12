package if_

import "fmt"

/**CFile****************************************************************

  FileName    [ifSelect.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Cut selection procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifSelect.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Prints the logic cone with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ObjConePrint_rec(pIfMan *If_Man_t, pIfObj *If_Obj_t, vVisited *Vec_Ptr_t) {
    var pCut *If_Cut_t;
    pCut = If_ObjCutBest(pIfObj);
    if If_CutDataInt(pCut) {
        return; }
    Vec_PtrPush( vVisited, pCut );
    // insert the worst case
    If_CutSetDataInt( pCut, ~0 );
    // skip in case of primary input
    if If_ObjIsCi(pIfObj) {
        return; }
    // compute the functions of the children
    if pIfObj.pEquiv { 
    If_ObjConePrint_rec( pIfMan, pIfObj.pEquiv, vVisited ); }
    If_ObjConePrint_rec( pIfMan, pIfObj.pFanin0, vVisited );
    If_ObjConePrint_rec( pIfMan, pIfObj.pFanin1, vVisited );
    fmt.Printf("%5d = %5d & %5d | %5d\n", pIfObj.Id, pIfObj.pFanin0.Id, pIfObj.pFanin1.Id, pIfObj.pEquiv ? pIfObj.pEquiv.Id : 0);
}
func If_ObjConePrint(pIfMan *If_Man_t, pIfObj *If_Obj_t) {
    var pCut *If_Cut_t;
    var pLeaf *If_Obj_t;
    var i int32;
    Vec_PtrClear( pIfMan.vTemp );
    If_ObjConePrint_rec( pIfMan, pIfObj, pIfMan.vTemp );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pIfMan.vTemp)) && (((pCut) = (If_Cut_t *)Vec_PtrEntry(pIfMan.vTemp, i)), 1); i++  {
        If_CutSetDataInt( pCut, 0 );
    }
    // print the leaf variables
    fmt.Printf("Cut ");
    pCut = If_ObjCutBest(pIfObj);
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        fmt.Printf("%d ", pLeaf.Id);
    }
    fmt.Printf("\n");
}


/**Function*************************************************************

  Synopsis    [Recursively derives local AIG for the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManNodeShapeMap_rec(pIfMan *If_Man_t, pIfObj *If_Obj_t, vVisited *Vec_Ptr_t, vShape *Vec_Int_t) int32 {
    var pCut *If_Cut_t;
    var pTemp *If_Obj_t;
    var i, iFunc0, iFunc1 int32;
    // get the best cut
    pCut = If_ObjCutBest(pIfObj);
    // if the cut is visited, return the result
    if If_CutDataInt(pCut) {
        return If_CutDataInt(pCut); }
    // mark the node as visited
    Vec_PtrPush( vVisited, pCut );
    // insert the worst case
    If_CutSetDataInt( pCut, ~0 );
    // skip in case of primary input
    if If_ObjIsCi(pIfObj) {
        return If_CutDataInt(pCut); }
    // compute the functions of the children
    for  i, pTemp = 0, pIfObj; (pTemp != nil); pTemp, i = pTemp.pEquiv, i+1  {
        iFunc0 = If_ManNodeShapeMap_rec( pIfMan, pTemp.pFanin0, vVisited, vShape );
        if iFunc0 == ~0 {
            continue; }
        iFunc1 = If_ManNodeShapeMap_rec( pIfMan, pTemp.pFanin1, vVisited, vShape );
        if iFunc1 == ~0 {
            continue; }
        // both branches are solved
        Vec_IntPush( vShape, pIfObj.Id );
        Vec_IntPush( vShape, pTemp.Id );
        If_CutSetDataInt( pCut, 1 );
        break;
    }
    return If_CutDataInt(pCut);
}
func If_ManNodeShapeMap(pIfMan *If_Man_t, pIfObj *If_Obj_t, vShape *Vec_Int_t) int32 {
    var pCut *If_Cut_t;
    var pLeaf *If_Obj_t;
    var i, iRes int32;
    // get the best cut
    pCut = If_ObjCutBest(pIfObj);
    assert(pCut.nLeaves > 1, "pCut.nLeaves > 1");
    // set the leaf variables
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        assert(If_CutDataInt( If_ObjCutBest(pLeaf) ) == 0, "If_CutDataInt( If_ObjCutBest(pLeaf) ) == 0");
        If_CutSetDataInt( If_ObjCutBest(pLeaf), 1 );
    }
    // recursively compute the function while collecting visited cuts
    Vec_IntClear( vShape );
    Vec_PtrClear( pIfMan.vTemp );
    iRes = If_ManNodeShapeMap_rec( pIfMan, pIfObj, pIfMan.vTemp, vShape ); 
    if iRes == ~0 {
        Abc_Print( -1, "If_ManNodeShapeMap(): Computing local AIG has failed.\n" );
        return 0;
    }
    // clean the cuts
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        If_CutSetDataInt( If_ObjCutBest(pLeaf), 0 );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pIfMan.vTemp)) && (((pCut) = (If_Cut_t *)Vec_PtrEntry(pIfMan.vTemp, i)), 1); i++  {
        If_CutSetDataInt( pCut, 0 );
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Recursively derives the local AIG for the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_WordCountOnes(uWord uint32) int32 {
    uWord = (uWord & 0x55555555) + ((uWord>>1) & 0x55555555);
    uWord = (uWord & 0x33333333) + ((uWord>>2) & 0x33333333);
    uWord = (uWord & 0x0F0F0F0F) + ((uWord>>4) & 0x0F0F0F0F);
    uWord = (uWord & 0x00FF00FF) + ((uWord>>8) & 0x00FF00FF);
    return  (uWord & 0x0000FFFF) + (uWord>>16);
}
func If_ManNodeShapeMap2_rec(pIfMan *If_Man_t, pIfObj *If_Obj_t, vVisited *Vec_Ptr_t, vShape *Vec_Int_t) int32 {
    var pCut *If_Cut_t;
    var pTemp *If_Obj_t;
    var pTempBest *If_Obj_t = nil;
    var i, iFunc, iFunc0, iFunc1, iBest int32;
    // get the best cut
    pCut = If_ObjCutBest(pIfObj);
    // if the cut is visited, return the result
    if If_CutDataInt(pCut) {
        return If_CutDataInt(pCut); }
    // mark the node as visited
    Vec_PtrPush( vVisited, pCut );
    // insert the worst case
    If_CutSetDataInt( pCut, ~0 );
    // skip in case of primary input
    if If_ObjIsCi(pIfObj) {
        return If_CutDataInt(pCut); }
    // compute the functions of the children
    for  i, pTemp = 0, pIfObj; (pTemp != nil); pTemp, i = pTemp.pEquiv, i+1  {
        iFunc0 = If_ManNodeShapeMap2_rec( pIfMan, pTemp.pFanin0, vVisited, vShape );
        if iFunc0 == ~0 {
            continue; }
        iFunc1 = If_ManNodeShapeMap2_rec( pIfMan, pTemp.pFanin1, vVisited, vShape );
        if iFunc1 == ~0 {
            continue; }
        iFunc = iFunc0 | iFunc1;
//        if ( If_WordCountOnes(iBest) <= If_WordCountOnes(iFunc) )
        if iBest < iFunc {
            iBest = iFunc;
            pTempBest = pTemp;
        }
    }
    if pTempBest != nil {
        Vec_IntPush( vShape, pIfObj.Id );
        Vec_IntPush( vShape, pTempBest.Id );
        If_CutSetDataInt( pCut, iBest );
    }
    return If_CutDataInt(pCut);
}
func If_ManNodeShapeMap2(pIfMan *If_Man_t, pIfObj *If_Obj_t, vShape *Vec_Int_t) int32 {
    var pCut *If_Cut_t;
    var pLeaf *If_Obj_t;
    var i, iRes int32;
    // get the best cut
    pCut = If_ObjCutBest(pIfObj);
    assert(pCut.nLeaves > 1, "pCut.nLeaves > 1");
    // set the leaf variables
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        If_CutSetDataInt( If_ObjCutBest(pLeaf), (1 << i) );
    }
    // recursively compute the function while collecting visited cuts
    Vec_IntClear( vShape );
    Vec_PtrClear( pIfMan.vTemp );
    iRes = If_ManNodeShapeMap2_rec( pIfMan, pIfObj, pIfMan.vTemp, vShape ); 
    if iRes == ~0 {
        Abc_Print( -1, "If_ManNodeShapeMap2(): Computing local AIG has failed.\n" );
        return 0;
    }
    // clean the cuts
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        If_CutSetDataInt( If_ObjCutBest(pLeaf), 0 );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pIfMan.vTemp)) && (((pCut) = (If_Cut_t *)Vec_PtrEntry(pIfMan.vTemp, i)), 1); i++  {
        If_CutSetDataInt( pCut, 0 );
    }
    return 1;
}



/**Function*************************************************************

  Synopsis    [Collects logic cone with choices]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManConeCollect_rec(pIfMan *If_Man_t, pIfObj *If_Obj_t, vVisited *Vec_Ptr_t, vCone *Vec_Ptr_t) int32 {
    var pCut *If_Cut_t;
    var pTemp *If_Obj_t;
    var iFunc0, iFunc1 int32;
    var fRootAdded int32;
    var fNodeAdded int32;
    // get the best cut
    pCut = If_ObjCutBest(pIfObj);
    // if the cut is visited, return the result
    if If_CutDataInt(pCut) {
        return If_CutDataInt(pCut); }
    // mark the node as visited
    Vec_PtrPush( vVisited, pCut );
    // insert the worst case
    If_CutSetDataInt( pCut, ~0 );
    // skip in case of primary input
    if If_ObjIsCi(pIfObj) {
        return If_CutDataInt(pCut); }
    // compute the functions of the children
    for  pTemp = pIfObj; (pTemp != nil); pTemp = pTemp.pEquiv  {
        iFunc0 = If_ManConeCollect_rec( pIfMan, pTemp.pFanin0, vVisited, vCone );
        if iFunc0 == ~0 {
            continue; }
        iFunc1 = If_ManConeCollect_rec( pIfMan, pTemp.pFanin1, vVisited, vCone );
        if iFunc1 == ~0 {
            continue; }
        fNodeAdded = 1;
        If_CutSetDataInt( pCut, 1 );
        Vec_PtrPush( vCone, pTemp );
        if fRootAdded == 0 && pTemp == pIfObj {
            fRootAdded = 1; }
    }
    if (fNodeAdded != 0) && !(fRootAdded != 0) {
        Vec_PtrPush( vCone, pIfObj ); }
    return If_CutDataInt(pCut);
}
func If_ManConeCollect(pIfMan *If_Man_t, pIfObj *If_Obj_t, pCut *If_Cut_t) *Vec_Ptr_t {
    var vCone *Vec_Ptr_t;
    var pLeaf *If_Obj_t;
    var i, RetValue int32;
    // set the leaf variables
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        assert(If_CutDataInt( If_ObjCutBest(pLeaf) ) == 0, "If_CutDataInt( If_ObjCutBest(pLeaf) ) == 0");
        If_CutSetDataInt( If_ObjCutBest(pLeaf), 1 );
    }
    // recursively compute the function while collecting visited cuts
    vCone = Vec_PtrAlloc( 100 );
    Vec_PtrClear( pIfMan.vTemp );
    RetValue = If_ManConeCollect_rec( pIfMan, pIfObj, pIfMan.vTemp, vCone ); 
    assert((RetValue != 0), "RetValue");
    // clean the cuts
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        If_CutSetDataInt( If_ObjCutBest(pLeaf), 0 );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pIfMan.vTemp)) && (((pCut) = (If_Cut_t *)Vec_PtrEntry(pIfMan.vTemp, i)), 1); i++  {
        If_CutSetDataInt( pCut, 0 );
    }
    return vCone;
}

/**Function*************************************************************

  Synopsis    [Adding clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func sat_solver_add_choice(pSat *sat_solver, iVar int32, vVars *Vec_Int_t) {
    var pVars *int32  = Vec_IntArray(vVars);
    var nVars int32  = Vec_IntSize(vVars);
    var i int32;
    var k int32;
    var Lits [2]int32;
    var Value int32;
    assert(Vec_IntSize(vVars) < Vec_IntCap(vVars), "Vec_IntSize(vVars) < Vec_IntCap(vVars)");
    // create literals
    for  i = 0; i < nVars; i++  {
        pVars[i] = Abc_Var2Lit( pVars[i], 0 );
    }
    pVars[i] = Abc_Var2Lit( iVar, 1 );
    // add clause
    Value = sat_solver_addclause( pSat, pVars, pVars + nVars + 1 );
    assert((Value != 0), "Value");
    // undo literals
    for  i = 0; i < nVars; i++  {
        pVars[i] = Abc_Lit2Var( pVars[i] );
    }
    // add !out => !in
    Lits[0] = Abc_Var2Lit( iVar, 0 );
    for  i = 0; i < nVars; i++  {
        Lits[1] = Abc_Var2Lit( pVars[i], 1 );
        Value = sat_solver_addclause( pSat, Lits, Lits + 2 );
        assert((Value != 0), "Value");
    }
    // add excluvisity
    for  i = 0; i < nVars; i++  {
    for  k = i+1; k < nVars; k++  {
        Lits[0] = Abc_Var2Lit( pVars[i], 1 );
        Lits[1] = Abc_Var2Lit( pVars[k], 1 );
        Value = sat_solver_addclause( pSat, Lits, Lits + 2 );
        assert((Value != 0), "Value");
    }
    }
}
func If_ObjSatVar(pIfObj *If_Obj_t) int32 { return If_CutDataInt(If_ObjCutBest(pIfObj));  }
func If_ObjSetSatVar(pIfObj *If_Obj_t, v int32) { If_CutSetDataInt( If_ObjCutBest(pIfObj), v ); }


/**Function*************************************************************

  Synopsis    [Recursively derives the local AIG for the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManNodeShape2_rec(pSat *sat_solver, pIfMan *If_Man_t, pIfObj *If_Obj_t, vShape *Vec_Int_t) {
    var pTemp *If_Obj_t;
    assert(sat_solver_var_value(pSat, If_ObjSatVar(pIfObj)) == 1, "sat_solver_var_value(pSat, If_ObjSatVar(pIfObj)) == 1");
    if pIfObj.fMark {
        return; }
    pIfObj.fMark = 1;
    for  pTemp = pIfObj; (pTemp != nil); pTemp = pTemp.pEquiv  {
        if sat_solver_var_value(pSat, If_ObjSatVar(pTemp)+1) == 1 {
            break; }
    }
    assert(pTemp != nil, "pTemp != nil");
    If_ManNodeShape2_rec( pSat, pIfMan, pTemp.pFanin0, vShape );
    If_ManNodeShape2_rec( pSat, pIfMan, pTemp.pFanin1, vShape );
    Vec_IntPush( vShape, pIfObj.Id );
    Vec_IntPush( vShape, pTemp.Id );
}

/**Function*************************************************************

  Synopsis    [Solve the problem of selecting choices for the given cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManNodeShapeSat(pIfMan *If_Man_t, pIfObj *If_Obj_t, vShape *Vec_Int_t) int32 {
    var pSat *sat_solver;
    var pCut *If_Cut_t;
    var vCone *Vec_Ptr_t;
    var vFanins *Vec_Int_t;
    var pObj, pTemp *If_Obj_t;
    var i, Lit, Status int32;

    // get best cut
    pCut = If_ObjCutBest(pIfObj);
    assert(pCut.nLeaves > 1, "pCut.nLeaves > 1");

    // collect the cone
    vCone = If_ManConeCollect( pIfMan, pIfObj, pCut );

    // assign SAT variables
    // EXTERNAL variable is even numbered
    // INTERNAL variable is odd numbered
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pObj) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        assert(If_ObjSatVar(pObj) == 0, "If_ObjSatVar(pObj) == 0");
        If_ObjSetSatVar( pObj, 2*(i+1) );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCone)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(vCone, i)), 1); i++  {
        assert(If_ObjSatVar(pObj) == 0, "If_ObjSatVar(pObj) == 0");
        If_ObjSetSatVar( pObj, 2*(i+1+pCut.nLeaves) );
    }

    // start SAT solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, 2 * (pCut.nLeaves + Vec_PtrSize(vCone) + 1) );

    // add constraints
    vFanins = Vec_IntAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCone)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(vCone, i)), 1); i++  {
        assert(If_ObjIsAnd(pObj), "If_ObjIsAnd(pObj)");
        Vec_IntClear( vFanins );
        for  pTemp = pObj; (pTemp != nil); pTemp = pTemp.pEquiv  {
            if If_ObjSatVar(pTemp) != 0 {
                Vec_IntPush( vFanins, If_ObjSatVar(pTemp)+1 ); }
        } // internal
        assert(Vec_IntSize(vFanins) > 0, "Vec_IntSize(vFanins) > 0");
        sat_solver_add_choice( pSat, If_ObjSatVar(pObj), vFanins ); // external
        assert(If_ObjSatVar(pObj) > 0, "If_ObjSatVar(pObj) > 0");
//        sat_solver_add_and( pSat, If_ObjSatVar(pObj)+1, If_ObjSatVar(pObj->pFanin0), If_ObjSatVar(pObj->pFanin1), 0, 0, 0 );
        if If_ObjSatVar(pObj.pFanin0) > 0 && If_ObjSatVar(pObj.pFanin1) > 0 {
            var Lits [2]int32;
            Lits[0] = Abc_Var2Lit( If_ObjSatVar(pObj)+1,        1 );
            Lits[1] = Abc_Var2Lit( If_ObjSatVar(pObj.pFanin0), 0 );
            Status = sat_solver_addclause( pSat, Lits, Lits + 2 );
            assert((Status != 0), "Status");

            Lits[0] = Abc_Var2Lit( If_ObjSatVar(pObj)+1,        1 );
            Lits[1] = Abc_Var2Lit( If_ObjSatVar(pObj.pFanin1), 0 );
            Status = sat_solver_addclause( pSat, Lits, Lits + 2 );
            assert((Status != 0), "Status");
        }
    }
    Vec_IntFree( vFanins );

    // set cut variables to 1
    pCut = If_ObjCutBest(pIfObj);
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pObj) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        Lit = Abc_Var2Lit( If_ObjSatVar(pObj), 0 ); // external
        Status = sat_solver_addclause( pSat, &Lit, &Lit + 1 );
        assert((Status != 0), "Status");
    }
    // set output variable to 1
    Lit = Abc_Var2Lit( If_ObjSatVar(pIfObj), 0 ); // external
    Status = sat_solver_addclause( pSat, &Lit, &Lit + 1 );
    assert((Status != 0), "Status");

    // solve the problem
    Status = sat_solver_solve( pSat, nil, nil, 0, 0, 0, 0 );
    assert(Status == l_True, "Status == l_True");

    // mark cut nodes
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pObj) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        assert(pObj.fMark == 0, "pObj.fMark == 0");
        pObj.fMark = 1;
    }

    // select the node's shape
    Vec_IntClear( vShape );
    assert(pIfObj.fMark == 0, "pIfObj.fMark == 0");
    If_ManNodeShape2_rec( pSat, pIfMan, pIfObj, vShape ); 

    // cleanup
    sat_solver_delete( pSat );
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pObj) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        If_ObjSetSatVar( pObj, 0 );
        pObj.fMark = 0;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCone)) && (((pObj) = (If_Obj_t *)Vec_PtrEntry(vCone, i)), 1); i++  {
        If_ObjSetSatVar( pObj, 0 );
        pObj.fMark = 0;
    }
    Vec_PtrFree( vCone );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Verify that the shape is correct.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManCheckShape(pIfMan *If_Man_t, pIfObj *If_Obj_t, vShape *Vec_Int_t) int32 {
    var pCut *If_Cut_t;
    var pLeaf *If_Obj_t;
    var i int32;
    var Entry1 int32;
    var Entry2 int32;
    var RetValue int32 = 1;
    // check that the marks are not set
    pCut = If_ObjCutBest(pIfObj);
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        assert(pLeaf.fMark == 0, "pLeaf.fMark == 0");
    }
    // set the marks of the shape
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vShape)) && (((Entry1) = Vec_IntEntry(vShape, i)), 1) && (((Entry2) = Vec_IntEntry(vShape, i+1)), 1); i += 2  {
        pLeaf = If_ManObj(pIfMan, Entry2);
        pLeaf.pFanin0.fMark = 1;
        pLeaf.pFanin1.fMark = 1;
    }
    // check that the leaves are marked
    // inlining of macro If_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nLeaves)) && ((pLeaf) = If_ManObj(pIfMan, (pCut).pLeaves[i])); i++  {
        if pLeaf.fMark == 0 {
            RetValue = 0;
        } else {
            pLeaf.fMark = 0; }
    }
    // clean the inner marks
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vShape)) && (((Entry1) = Vec_IntEntry(vShape, i)), 1) && (((Entry2) = Vec_IntEntry(vShape, i+1)), 1); i += 2  {
        pLeaf = If_ManObj(pIfMan, Entry2);
        pLeaf.pFanin0.fMark = 0;
        pLeaf.pFanin1.fMark = 0;
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Recursively compute the set of nodes supported by the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_ManNodeShape(pIfMan *If_Man_t, pIfObj *If_Obj_t, vShape *Vec_Int_t, fExact int32) int32 {
    var RetValue int32;
//    if ( pIfMan->nChoices == 0 )
    {
        RetValue = If_ManNodeShapeMap( pIfMan, pIfObj, vShape );
        assert((RetValue != 0), "RetValue");
        if !(fExact != 0) || (If_ManCheckShape(pIfMan, pIfObj, vShape) != 0) {
            return 1; }
    }
//    if ( pIfObj->Id == 1254 && If_ObjCutBest(pIfObj)->nLeaves == 7 )
//        If_ObjConePrint( pIfMan, pIfObj );
    RetValue = If_ManNodeShapeMap2( pIfMan, pIfObj, vShape );
    assert((RetValue != 0), "RetValue");
    RetValue = If_ManCheckShape(pIfMan, pIfObj, vShape);
//    assert( RetValue );
//    printf( "%d", RetValue );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

