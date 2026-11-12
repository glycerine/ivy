package fra

/**CFile****************************************************************

  FileName    [fraCnf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraCnf.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

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

  Synopsis    [Addes clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_AddClausesMux(p *Fra_Man_t, pNode *Aig_Obj_t) {
    var pNodeI, pNodeT, pNodeE *Aig_Obj_t;
    var pLits [4]int32;
    var RetValue int32;
    var VarF int32;
    var VarI int32;
    var VarT int32;
    var VarE int32;
    var fCompT int32;
    var fCompE int32;

    assert(!Aig_IsComplement( pNode ), "!Aig_IsComplement( pNode )");
    assert(Aig_ObjIsMuxType( pNode ), "Aig_ObjIsMuxType( pNode )");
    // get nodes (I = if, T = then, E = else)
    pNodeI = Aig_ObjRecognizeMux( pNode, &pNodeT, &pNodeE );
    // get the variable numbers
    VarF = Fra_ObjSatNum(pNode);
    VarI = Fra_ObjSatNum(pNodeI);
    VarT = Fra_ObjSatNum(Aig_Regular(pNodeT));
    VarE = Fra_ObjSatNum(Aig_Regular(pNodeE));
    // get the complementation flags
    fCompT = Aig_IsComplement(pNodeT);
    fCompE = Aig_IsComplement(pNodeE);

    // f = ITE(i, t, e)

    // i' + t' + f
    // i' + t  + f'
    // i  + e' + f
    // i  + e  + f'

    // create four clauses
    pLits[0] = toLitCond(VarI, 1);
    pLits[1] = toLitCond(VarT, 1^fCompT);
    pLits[2] = toLitCond(VarF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 1);
    pLits[1] = toLitCond(VarT, 0^fCompT);
    pLits[2] = toLitCond(VarF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 0);
    pLits[1] = toLitCond(VarE, 1^fCompE);
    pLits[2] = toLitCond(VarF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 0);
    pLits[1] = toLitCond(VarE, 0^fCompE);
    pLits[2] = toLitCond(VarF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");

    // two additional clauses
    // t' & e' -> f'
    // t  & e  -> f 

    // t  + e   + f'
    // t' + e'  + f 

    if VarT == VarE {
//        assert( fCompT == !fCompE );
        return;
    }

    pLits[0] = toLitCond(VarT, 0^fCompT);
    pLits[1] = toLitCond(VarE, 0^fCompE);
    pLits[2] = toLitCond(VarF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarT, 1^fCompT);
    pLits[1] = toLitCond(VarE, 1^fCompE);
    pLits[2] = toLitCond(VarF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
}

/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_AddClausesSuper(p *Fra_Man_t, pNode *Aig_Obj_t, vSuper *Vec_Ptr_t) {
    var pFanin *Aig_Obj_t;
    var pLits *int32;
    var nLits int32;
    var RetValue int32;
    var i int32;
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert(Aig_ObjIsNode( pNode ), "Aig_ObjIsNode( pNode )");
    // create storage for literals
    nLits = Vec_PtrSize(vSuper) + 1;
    // inlining of macro ABC_ALLOC
    pLits = ((int *) malloc(sizeof(int) * (size_t)(nLits)));
    // suppose AND-gate is A & B = C
    // add !A => !C   or   A + !C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[0] = toLitCond(Fra_ObjSatNum(Aig_Regular(pFanin)), Aig_IsComplement(pFanin));
        pLits[1] = toLitCond(Fra_ObjSatNum(pNode), 1);
        RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[i] = toLitCond(Fra_ObjSatNum(Aig_Regular(pFanin)), !Aig_IsComplement(pFanin));
    }
    pLits[nLits-1] = toLitCond(Fra_ObjSatNum(pNode), 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + nLits );
    assert((RetValue != 0), "RetValue");
    // inlining of macro ABC_FREE
    ((pLits) ? (free((char *) (pLits)), (pLits) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_CollectSuper_rec(pObj *Aig_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Aig_IsComplement(pObj) || Aig_ObjIsCi(pObj) || (!(fFirst != 0) && Aig_ObjRefs(pObj) > 1) || 
         ((fUseMuxes != 0) && Aig_ObjIsMuxType(pObj)) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
    // go through the branches
    Fra_CollectSuper_rec( Aig_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Fra_CollectSuper_rec( Aig_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_CollectSuper(pObj *Aig_Obj_t, fUseMuxes int32) *Vec_Ptr_t {
    var vSuper *Vec_Ptr_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(!Aig_ObjIsCi(pObj), "!Aig_ObjIsCi(pObj)");
    vSuper = Vec_PtrAlloc( 4 );
    Fra_CollectSuper_rec( pObj, vSuper, 1, fUseMuxes );
    return vSuper;
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ObjAddToFrontier(p *Fra_Man_t, pObj *Aig_Obj_t, vFrontier *Vec_Ptr_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Fra_ObjSatNum(pObj) {
        return; }
    assert(Fra_ObjSatNum(pObj) == 0, "Fra_ObjSatNum(pObj) == 0");
    assert(Fra_ObjFaninVec(pObj) == nil, "Fra_ObjFaninVec(pObj) == nil");
    if Aig_ObjIsConst1(pObj) {
        return; }
    Fra_ObjSetSatNum( pObj, p.nSatVars++ );
    if Aig_ObjIsNode(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_CnfNodeAddToSolver(p *Fra_Man_t, pOld *Aig_Obj_t, pNew *Aig_Obj_t) { 
    var vFrontier, vFanins *Vec_Ptr_t;
    var pNode, pFanin *Aig_Obj_t;
    var i int32;
    var k int32;
    var fUseMuxes int32 = 1;
    assert((pOld != nil) || (pNew != nil), "pOld || pNew");
    // quit if CNF is ready
    if (!(pOld != nil) || Fra_ObjFaninVec(pOld)) && (!(pNew != nil) || Fra_ObjFaninVec(pNew)) {
        return; }
    // start the frontier
    vFrontier = Vec_PtrAlloc( 100 );
    if pOld != nil { Fra_ObjAddToFrontier( p, pOld, vFrontier ); }
    if pNew != nil { Fra_ObjAddToFrontier( p, pNew, vFrontier ); }
    // explore nodes in the frontier
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFrontier)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(vFrontier, i)), 1); i++  {
        // create the supergate
        assert(Fra_ObjSatNum(pNode), "Fra_ObjSatNum(pNode)");
        assert(Fra_ObjFaninVec(pNode) == nil, "Fra_ObjFaninVec(pNode) == nil");
        if (fUseMuxes != 0) && Aig_ObjIsMuxType(pNode) {
            vFanins = Vec_PtrAlloc( 4 );
            Vec_PtrPushUnique( vFanins, Aig_ObjFanin0( Aig_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( vFanins, Aig_ObjFanin0( Aig_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( vFanins, Aig_ObjFanin1( Aig_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( vFanins, Aig_ObjFanin1( Aig_ObjFanin1(pNode) ) );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFanins)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(vFanins, k)), 1); k++  {
                Fra_ObjAddToFrontier( p, Aig_Regular(pFanin), vFrontier );
            }
            Fra_AddClausesMux( p, pNode );
        
        } else {
            vFanins = Fra_CollectSuper( pNode, fUseMuxes );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFanins)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(vFanins, k)), 1); k++  {
                Fra_ObjAddToFrontier( p, Aig_Regular(pFanin), vFrontier );
            }
            Fra_AddClausesSuper( p, pNode, vFanins );
        }
        assert(Vec_PtrSize(vFanins) > 1, "Vec_PtrSize(vFanins) > 1");
        Fra_ObjSetFaninVec( pNode, vFanins );
    }
    Vec_PtrFree( vFrontier );
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

