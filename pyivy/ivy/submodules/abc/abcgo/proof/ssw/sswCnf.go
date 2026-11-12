package ssw

/**CFile****************************************************************

  FileName    [sswCnf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Computation of CNF.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswCnf.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

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

  Synopsis    [Starts the SAT manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SatStart(fPolarFlip int32) *Ssw_Sat_t {
    var p *Ssw_Sat_t;
    var Lit int32;
    // inlining of macro ABC_ALLOC
    p = ((Ssw_Sat_t *) malloc(sizeof(Ssw_Sat_t) * (size_t)(1)));
    memset( p, 0, sizeof(Ssw_Sat_t) );
    p.pAig          = nil;
    p.fPolarFlip    = fPolarFlip;
    p.vSatVars      = Vec_IntStart( 10000 );
    p.vFanins       = Vec_PtrAlloc( 100 );
    p.vUsedPis      = Vec_PtrAlloc( 100 );
    p.pSat          = sat_solver_new();
    sat_solver_setnvars( p.pSat, 1000 );
    // var 0 is not used
    // var 1 is reserved for const1 node - add the clause
    p.nSatVars = 1;
    Lit = toLit( p.nSatVars );
    if fPolarFlip != 0 {
        Lit = lit_neg( Lit ); }
    sat_solver_addclause( p.pSat, &Lit, &Lit + 1 );
//    Ssw_ObjSetSatNum( p, Aig_ManConst1(p->pAig), p->nSatVars++ );
    Vec_IntWriteEntry( p.vSatVars, 0, p.nSatVars++ );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stop the SAT manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SatStop(p *Ssw_Sat_t) {
//    Abc_Print( 1, "Recycling SAT solver with %d vars and %d restarts.\n", 
//        p->pSat->size, p->pSat->stats.starts );
    if p.pSat {
        sat_solver_delete( p.pSat ); }
    Vec_IntFree( p.vSatVars );
    Vec_PtrFree( p.vFanins );
    Vec_PtrFree( p.vUsedPis );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_AddClausesMux(p *Ssw_Sat_t, pNode *Aig_Obj_t) {
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
    VarF = Ssw_ObjSatNum(p,pNode);
    VarI = Ssw_ObjSatNum(p,pNodeI);
    VarT = Ssw_ObjSatNum(p,Aig_Regular(pNodeT));
    VarE = Ssw_ObjSatNum(p,Aig_Regular(pNodeE));
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
    if p.fPolarFlip {
        if pNodeI.fPhase {               pLits[0] = lit_neg( pLits[0] ); }
        if Aig_Regular(pNodeT).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 1);
    pLits[1] = toLitCond(VarT, 0^fCompT);
    pLits[2] = toLitCond(VarF, 1);
    if p.fPolarFlip {
        if pNodeI.fPhase {               pLits[0] = lit_neg( pLits[0] ); }
        if Aig_Regular(pNodeT).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 0);
    pLits[1] = toLitCond(VarE, 1^fCompE);
    pLits[2] = toLitCond(VarF, 0);
    if p.fPolarFlip {
        if pNodeI.fPhase {               pLits[0] = lit_neg( pLits[0] ); }
        if Aig_Regular(pNodeE).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 0);
    pLits[1] = toLitCond(VarE, 0^fCompE);
    pLits[2] = toLitCond(VarF, 1);
    if p.fPolarFlip {
        if pNodeI.fPhase {               pLits[0] = lit_neg( pLits[0] ); }
        if Aig_Regular(pNodeE).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
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
    if p.fPolarFlip {
        if Aig_Regular(pNodeT).fPhase {  pLits[0] = lit_neg( pLits[0] ); }
        if Aig_Regular(pNodeE).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarT, 1^fCompT);
    pLits[1] = toLitCond(VarE, 1^fCompE);
    pLits[2] = toLitCond(VarF, 0);
    if p.fPolarFlip {
        if Aig_Regular(pNodeT).fPhase {  pLits[0] = lit_neg( pLits[0] ); }
        if Aig_Regular(pNodeE).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
}

/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_AddClausesSuper(p *Ssw_Sat_t, pNode *Aig_Obj_t, vSuper *Vec_Ptr_t) {
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
        pLits[0] = toLitCond(Ssw_ObjSatNum(p,Aig_Regular(pFanin)), Aig_IsComplement(pFanin));
        pLits[1] = toLitCond(Ssw_ObjSatNum(p,pNode), 1);
        if p.fPolarFlip {
            if Aig_Regular(pFanin).fPhase {  pLits[0] = lit_neg( pLits[0] ); }
            if pNode.fPhase {                pLits[1] = lit_neg( pLits[1] ); }
        }
        RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[i] = toLitCond(Ssw_ObjSatNum(p,Aig_Regular(pFanin)), !Aig_IsComplement(pFanin));
        if p.fPolarFlip {
            if Aig_Regular(pFanin).fPhase {  pLits[i] = lit_neg( pLits[i] ); }
        }
    }
    pLits[nLits-1] = toLitCond(Ssw_ObjSatNum(p,pNode), 0);
    if p.fPolarFlip {
        if pNode.fPhase {  pLits[nLits-1] = lit_neg( pLits[nLits-1] ); }
    }
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
func Ssw_CollectSuper_rec(pObj *Aig_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Aig_IsComplement(pObj) || Aig_ObjIsCi(pObj) ||
         (!(fFirst != 0) && Aig_ObjRefs(pObj) > 1) ||
         ((fUseMuxes != 0) && Aig_ObjIsMuxType(pObj)) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
//    pObj->fMarkA = 1;
    // go through the branches
    Ssw_CollectSuper_rec( Aig_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Ssw_CollectSuper_rec( Aig_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_CollectSuper(pObj *Aig_Obj_t, fUseMuxes int32, vSuper *Vec_Ptr_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(!Aig_ObjIsCi(pObj), "!Aig_ObjIsCi(pObj)");
    Vec_PtrClear( vSuper );
    Ssw_CollectSuper_rec( pObj, vSuper, 1, fUseMuxes );
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ObjAddToFrontier(p *Ssw_Sat_t, pObj *Aig_Obj_t, vFrontier *Vec_Ptr_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Ssw_ObjSatNum(p,pObj) {
        return; }
    assert(Ssw_ObjSatNum(p,pObj) == 0, "Ssw_ObjSatNum(p,pObj) == 0");
    if Aig_ObjIsConst1(pObj) {
        return; }
//    pObj->fMarkA = 1;
    // save PIs (used by register correspondence)
    if Aig_ObjIsCi(pObj) {
        Vec_PtrPush( p.vUsedPis, pObj ); }
    Ssw_ObjSetSatNum( p, pObj, p.nSatVars++ );
    sat_solver_setnvars( p.pSat, 100 * (1 + p.nSatVars / 100) );
    if Aig_ObjIsNode(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_CnfNodeAddToSolver(p *Ssw_Sat_t, pObj *Aig_Obj_t) {
    var vFrontier *Vec_Ptr_t;
    var pNode, pFanin *Aig_Obj_t;
    var i int32;
    var k int32;
    var fUseMuxes int32 = 1;
    // quit if CNF is ready
    if Ssw_ObjSatNum(p,pObj) {
        return; }
    // start the frontier
    vFrontier = Vec_PtrAlloc( 100 );
    Ssw_ObjAddToFrontier( p, pObj, vFrontier );
    // explore nodes in the frontier
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFrontier)) && (((pNode) = (Aig_Obj_t *)Vec_PtrEntry(vFrontier, i)), 1); i++  {
        // create the supergate
        assert(Ssw_ObjSatNum(p,pNode), "Ssw_ObjSatNum(p,pNode)");
        if (fUseMuxes != 0) && Aig_ObjIsMuxType(pNode) {
            Vec_PtrClear( p.vFanins );
            Vec_PtrPushUnique( p.vFanins, Aig_ObjFanin0( Aig_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Aig_ObjFanin0( Aig_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Aig_ObjFanin1( Aig_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Aig_ObjFanin1( Aig_ObjFanin1(pNode) ) );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(p.vFanins, k)), 1); k++  {
                Ssw_ObjAddToFrontier( p, Aig_Regular(pFanin), vFrontier );
            }
            Ssw_AddClausesMux( p, pNode );
        
        } else {
            Ssw_CollectSuper( pNode, fUseMuxes, p.vFanins );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Aig_Obj_t *)Vec_PtrEntry(p.vFanins, k)), 1); k++  {
                Ssw_ObjAddToFrontier( p, Aig_Regular(pFanin), vFrontier );
            }
            Ssw_AddClausesSuper( p, pNode, p.vFanins );
        }
        assert(Vec_PtrSize(p.vFanins) > 1, "Vec_PtrSize(p.vFanins) > 1");
    }
    Vec_PtrFree( vFrontier );
}


/**Function*************************************************************

  Synopsis    [Copy pattern from the solver into the internal storage.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_CnfGetNodeValue(p *Ssw_Sat_t, pObj *Aig_Obj_t) int32 {
    var Value0, Value1, nVarNum int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    nVarNum = Ssw_ObjSatNum( p, pObj );
    if nVarNum > 0 {
        return sat_solver_var_value( p.pSat, nVarNum ); }
//    if ( pObj->fMarkA == 1 )
//        return 0;
    if Aig_ObjIsCi(pObj) {
        return 0; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Value0 = Ssw_CnfGetNodeValue( p, Aig_ObjFanin0(pObj) );
    Value0 ^= Aig_ObjFaninC0(pObj);
    Value1 = Ssw_CnfGetNodeValue( p, Aig_ObjFanin1(pObj) );
    Value1 ^= Aig_ObjFaninC1(pObj);
    return Value0 & Value1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
