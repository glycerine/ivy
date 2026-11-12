package gia



func Gia_Min2AddClausesMux(p *Gia_Man_t, pNode *Gia_Obj_t, pSat *satoko_t) {
    var fPolarFlip int32;
    var pNodeI, pNodeT, pNodeE *Gia_Obj_t;
    var pLits [4]int32;
    var RetValue int32;
    var VarF int32;
    var VarI int32;
    var VarT int32;
    var VarE int32;
    var fCompT int32;
    var fCompE int32;

    assert(!Gia_IsComplement( pNode ), "!Gia_IsComplement( pNode )");
    assert(pNode.fMark0, "pNode.fMark0");
    // get nodes (I = if, T = then, E = else)
    pNodeI = Gia_ObjRecognizeMux( pNode, &pNodeT, &pNodeE );
    // get the variable numbers
    VarF = Gia_Min2ObjSatId(p, pNode);
    VarI = Gia_Min2ObjSatId(p, pNodeI);
    VarT = Gia_Min2ObjSatId(p, Gia_Regular(pNodeT));
    VarE = Gia_Min2ObjSatId(p, Gia_Regular(pNodeE));
    // get the complementation flags
    fCompT = Gia_IsComplement(pNodeT);
    fCompE = Gia_IsComplement(pNodeE);

    // f = ITE(i, t, e)

    // i' + t' + f
    // i' + t  + f'
    // i  + e' + f
    // i  + e  + f'

    // create four clauses
    pLits[0] = Abc_Var2Lit(VarI, 1);
    pLits[1] = Abc_Var2Lit(VarT, 1^fCompT);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeT).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 1);
    pLits[1] = Abc_Var2Lit(VarT, 0^fCompT);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeT).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 0);
    pLits[1] = Abc_Var2Lit(VarE, 1^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 0);
    pLits[1] = Abc_Var2Lit(VarE, 0^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
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

    pLits[0] = Abc_Var2Lit(VarT, 0^fCompT);
    pLits[1] = Abc_Var2Lit(VarE, 0^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if Gia_Regular(pNodeT).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarT, 1^fCompT);
    pLits[1] = Abc_Var2Lit(VarE, 1^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if Gia_Regular(pNodeT).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
}
func Gia_Min2AddClausesSuper(p *Gia_Man_t, pNode *Gia_Obj_t, vSuper *Vec_Ptr_t, pSat *satoko_t) {
    var fPolarFlip int32;
    var pFanin *Gia_Obj_t;
    var pLits *int32;
    var nLits int32;
    var RetValue int32;
    var i int32;
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    assert(Gia_ObjIsAnd( pNode ), "Gia_ObjIsAnd( pNode )");
    // create storage for literals
    nLits = Vec_PtrSize(vSuper) + 1;
    pLits = ABC_ALLOC( int, nLits );
    // suppose AND-gate is A & B = C
    // add !A => !C   or   A + !C
    Vec_PtrForEachEntry( Gia_Obj_t *, vSuper, pFanin, i )
    {
        pLits[0] = Abc_Var2Lit(Gia_Min2ObjSatId(p, Gia_Regular(pFanin)), Gia_IsComplement(pFanin));
        pLits[1] = Abc_Var2Lit(Gia_Min2ObjSatId(p, pNode), 1);
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
            if pNode.fPhase {                pLits[1] = Abc_LitNot( pLits[1] ); }
        }
        RetValue = satoko_add_clause( pSat, pLits, 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    Vec_PtrForEachEntry( Gia_Obj_t *, vSuper, pFanin, i )
    {
        pLits[i] = Abc_Var2Lit(Gia_Min2ObjSatId(p, Gia_Regular(pFanin)), !Gia_IsComplement(pFanin));
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[i] = Abc_LitNot( pLits[i] ); }
        }
    }
    pLits[nLits-1] = Abc_Var2Lit(Gia_Min2ObjSatId(p, pNode), 0);
    if fPolarFlip != 0 {
        if pNode.fPhase {  pLits[nLits-1] = Abc_LitNot( pLits[nLits-1] ); }
    }
    RetValue = satoko_add_clause( pSat, pLits, nLits );
    assert((RetValue != 0), "RetValue");
    pLits = nil;
}

/**Function*************************************************************

  Synopsis    [Adds clauses and returns CNF variable of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Min2CollectSuper_rec(pObj *Gia_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Gia_IsComplement(pObj) || Gia_ObjIsCi(pObj) || 
         (!(fFirst != 0) && Gia_ObjValue(pObj) > 1) || 
         ((fUseMuxes != 0) && pObj.fMark0) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
    // go through the branches
    Gia_Min2CollectSuper_rec( Gia_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Gia_Min2CollectSuper_rec( Gia_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}
func Gia_Min2CollectSuper(pObj *Gia_Obj_t, fUseMuxes int32, vSuper *Vec_Ptr_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsCi(pObj), "!Gia_ObjIsCi(pObj)");
    Vec_PtrClear( vSuper );
    Gia_Min2CollectSuper_rec( pObj, vSuper, 1, fUseMuxes );
}
func Gia_Min2ObjAddToFrontier(p *Gia_Man_t, pObj *Gia_Obj_t, vFrontier *Vec_Ptr_t, pSat *satoko_t, vSatVars *Vec_Int_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsConst0(pObj), "!Gia_ObjIsConst0(pObj)");
    if Gia_Min2ObjSatId(p, pObj) >= 0 {
        return; }
    assert(Gia_Min2ObjSatId(p, pObj) == -1, "Gia_Min2ObjSatId(p, pObj) == -1");
    Vec_IntPush( vSatVars, Gia_ObjId(p, pObj) );
    Gia_Min2ObjSetSatId( p, pObj, satoko_add_variable(pSat, 0) );
    if Gia_ObjIsAnd(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}
func Gia_Min2ObjGetCnfVar(p *Gia_Man2Min_t, iObj int32) int32 { 
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var pNode, pFanin *Gia_Obj_t;
    var i int32;
    var k int32;
    var fUseMuxes int32 = 1;
    if Gia_Min2ObjSatId(p.pGia, pObj) >= 0 {
        return Gia_Min2ObjSatId(p.pGia, pObj); }
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( p.vSatVars, iObj );
        return Gia_Min2ObjSetSatId( p.pGia, pObj, satoko_add_variable(p.pSat, 0) );
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // start the frontier
    Vec_PtrClear( p.vFrontier );
    Gia_Min2ObjAddToFrontier( p.pGia, pObj, p.vFrontier, p.pSat, p.vSatVars );
    // explore nodes in the frontier
    Vec_PtrForEachEntry( Gia_Obj_t *, p.vFrontier, pNode, i )
    {
        assert(Gia_Min2ObjSatId(p.pGia,pNode) >= 0, "Gia_Min2ObjSatId(p.pGia,pNode) >= 0");
        if (fUseMuxes != 0) && pNode.fMark0 {
            Vec_PtrClear( p.vFanins );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin1(pNode) ) );
            Vec_PtrForEachEntry( Gia_Obj_t *, p.vFanins, pFanin, k )
                Gia_Min2ObjAddToFrontier( p.pGia, Gia_Regular(pFanin), p.vFrontier, p.pSat, p.vSatVars );
            Gia_Min2AddClausesMux( p.pGia, pNode, p.pSat );
        
        } else {
            Gia_Min2CollectSuper( pNode, fUseMuxes, p.vFanins );
            Vec_PtrForEachEntry( Gia_Obj_t *, p.vFanins, pFanin, k )
                Gia_Min2ObjAddToFrontier( p.pGia, Gia_Regular(pFanin), p.vFrontier, p.pSat, p.vSatVars );
            Gia_Min2AddClausesSuper( p.pGia, pNode, p.vFanins, p.pSat );
        }
        assert(Vec_PtrSize(p.vFanins) > 1, "Vec_PtrSize(p.vFanins) > 1");
    }
    return Gia_Min2ObjSatId(p.pGia,pObj);
}

/**Function*************************************************************

  Synopsis    [Returns 0 if the node is not a constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collect nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Takes a literal and returns a support-minized literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

