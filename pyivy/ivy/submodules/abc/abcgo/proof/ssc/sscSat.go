package ssc

/**CFile****************************************************************

  FileName    [sscSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sweeping under constraints.]

  Synopsis    [SAT procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: sscSat.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sscInt.h"
//#include "sat/cnf/cnf.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Ssc_ObjSatLit(p *Ssc_Man_t, Lit int32) int32 { return Abc_Var2Lit( Ssc_ObjSatVar(p, Abc_Lit2Var(Lit)), Abc_LitIsCompl(Lit) ); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAddClausesMux(p *Ssc_Man_t, pNode *Gia_Obj_t) {
    var pNodeI, pNodeT, pNodeE *Gia_Obj_t;
    var pLits [4]int32;
    var LitF int32;
    var LitI int32;
    var LitT int32;
    var LitE int32;
    var RetValue int32;
    assert(!Gia_IsComplement( pNode ), "!Gia_IsComplement( pNode )");
    assert(Gia_ObjIsMuxType( pNode ), "Gia_ObjIsMuxType( pNode )");
    // get nodes (I = if, T = then, E = else)
    pNodeI = Gia_ObjRecognizeMux( pNode, &pNodeT, &pNodeE );
    // get the Litiable numbers
    LitF = Ssc_ObjSatLit( p, Gia_Obj2Lit(p.pFraig,pNode) );
    LitI = Ssc_ObjSatLit( p, Gia_Obj2Lit(p.pFraig,pNodeI) );
    LitT = Ssc_ObjSatLit( p, Gia_Obj2Lit(p.pFraig,pNodeT) );
    LitE = Ssc_ObjSatLit( p, Gia_Obj2Lit(p.pFraig,pNodeE) );

    // f = ITE(i, t, e)

    // i' + t' + f
    // i' + t  + f'
    // i  + e' + f
    // i  + e  + f'

    // create four clauses
    pLits[0] = Abc_LitNotCond(LitI, 1);
    pLits[1] = Abc_LitNotCond(LitT, 1);
    pLits[2] = Abc_LitNotCond(LitF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_LitNotCond(LitI, 1);
    pLits[1] = Abc_LitNotCond(LitT, 0);
    pLits[2] = Abc_LitNotCond(LitF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_LitNotCond(LitI, 0);
    pLits[1] = Abc_LitNotCond(LitE, 1);
    pLits[2] = Abc_LitNotCond(LitF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_LitNotCond(LitI, 0);
    pLits[1] = Abc_LitNotCond(LitE, 0);
    pLits[2] = Abc_LitNotCond(LitF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");

    // two additional clauses
    // t' & e' -> f'
    // t  & e  -> f

    // t  + e   + f'
    // t' + e'  + f

    if LitT == LitE {
//        assert( fCompT == !fCompE );
        return;
    }

    pLits[0] = Abc_LitNotCond(LitT, 0);
    pLits[1] = Abc_LitNotCond(LitE, 0);
    pLits[2] = Abc_LitNotCond(LitF, 1);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_LitNotCond(LitT, 1);
    pLits[1] = Abc_LitNotCond(LitE, 1);
    pLits[2] = Abc_LitNotCond(LitF, 0);
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
}

/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAddClausesSuper(p *Ssc_Man_t, pNode *Gia_Obj_t, vSuper *Vec_Int_t) {
    var i int32;
    var RetValue int32;
    var Lit int32;
    var LitNode int32;
    var pLits [2]int32;
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    assert(Gia_ObjIsAnd( pNode ), "Gia_ObjIsAnd( pNode )");
    // suppose AND-gate is A & B = C
    // add !A => !C   or   A + !C
    // add !B => !C   or   B + !C
    LitNode = Ssc_ObjSatLit( p, Gia_Obj2Lit(p.pFraig,pNode) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((Lit) = Vec_IntEntry(vSuper, i)), 1); i++  {
        pLits[0] = Ssc_ObjSatLit( p, Lit );
        pLits[1] = Abc_LitNot( LitNode );
        RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
        assert((RetValue != 0), "RetValue");
        // update literals
        Vec_IntWriteEntry( vSuper, i, Abc_LitNot(pLits[0]) );
    }
    // add A & B => C   or   !A + !B + C
    Vec_IntPush( vSuper, LitNode );
    RetValue = sat_solver_addclause( p.pSat, Vec_IntArray(vSuper), Vec_IntArray(vSuper) + Vec_IntSize(vSuper) );
    assert((RetValue != 0), "RetValue");
    (void) RetValue;
}


/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_ManCollectSuper_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t) {
    // stop at complements, PIs, and MUXes
    if Gia_IsComplement(pObj) || Gia_ObjIsCi(pObj) || Gia_ObjIsMuxType(pObj) {
        Vec_IntPushUnique( vSuper, Gia_Obj2Lit(p, pObj) );
        return;
    }
    Ssc_ManCollectSuper_rec( p, Gia_ObjChild0(pObj), vSuper );
    Ssc_ManCollectSuper_rec( p, Gia_ObjChild1(pObj), vSuper );
}
func Ssc_ManCollectSuper(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Vec_IntClear( vSuper );
    Ssc_ManCollectSuper_rec( p, Gia_ObjChild0(pObj), vSuper );
    Ssc_ManCollectSuper_rec( p, Gia_ObjChild1(pObj), vSuper );
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_ManCnfAddToFrontier(p *Ssc_Man_t, Id int32, vFront *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    assert(Id > 0, "Id > 0");
    if Ssc_ObjSatVar(p, Id) {
        return; }
    pObj = Gia_ManObj( p.pFraig, Id );
    Ssc_ObjSetSatVar( p, Id, p.nSatVars++ );
    sat_solver_setnvars( p.pSat, p.nSatVars + 100 );
    if Gia_ObjIsAnd(pObj) {
        Vec_IntPush( vFront, Id ); }
}
func Ssc_ManCnfNodeAddToSolver(p *Ssc_Man_t, NodeId int32) {
    var pNode *Gia_Obj_t;
    var i, k, Id, Lit int32;
    var clk abctime;
    assert(NodeId > 0, "NodeId > 0");
    // quit if CNF is ready
    if Ssc_ObjSatVar(p, NodeId) {
        return; }
clk = Abc_Clock();
    // start the frontier
    Vec_IntClear( p.vFront );
    Ssc_ManCnfAddToFrontier( p, NodeId, p.vFront );
    // explore nodes in the frontier
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vFront)) && ((pNode) = Gia_ManObj(p.pFraig, Vec_IntEntry(p.vFront,i))); i++  {
        // create the supergate
        assert(Ssc_ObjSatVar(p, Gia_ObjId(p.pFraig, pNode)), "Ssc_ObjSatVar(p, Gia_ObjId(p.pFraig, pNode))");
        if Gia_ObjIsMuxType(pNode) {
            Vec_IntClear( p.vFanins );
            Vec_IntPushUnique( p.vFanins, Gia_ObjFaninId0p( p.pFraig, Gia_ObjFanin0(pNode) ) );
            Vec_IntPushUnique( p.vFanins, Gia_ObjFaninId0p( p.pFraig, Gia_ObjFanin1(pNode) ) );
            Vec_IntPushUnique( p.vFanins, Gia_ObjFaninId1p( p.pFraig, Gia_ObjFanin0(pNode) ) );
            Vec_IntPushUnique( p.vFanins, Gia_ObjFaninId1p( p.pFraig, Gia_ObjFanin1(pNode) ) );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(p.vFanins)) && (((Id) = Vec_IntEntry(p.vFanins, k)), 1); k++  {
                Ssc_ManCnfAddToFrontier( p, Id, p.vFront );
            }
            Gia_ManAddClausesMux( p, pNode );
        
        } else {
            Ssc_ManCollectSuper( p.pFraig, pNode, p.vFanins );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(p.vFanins)) && (((Lit) = Vec_IntEntry(p.vFanins, k)), 1); k++  {
                Ssc_ManCnfAddToFrontier( p, Abc_Lit2Var(Lit), p.vFront );
            }
            Gia_ManAddClausesSuper( p, pNode, p.vFanins );
        }
        assert(Vec_IntSize(p.vFanins) > 1, "Vec_IntSize(p.vFanins) > 1");
    }
p.timeCnfGen += Abc_Clock() - clk;
}


/**Function*************************************************************

  Synopsis    [Starts the SAT solver for constraints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_ManStartSolver(p *Ssc_Man_t) {
    var pMan *Aig_Man_t  = Gia_ManToAigSimple( p.pFraig );
    var pDat *Cnf_Dat_t  = Cnf_Derive( pMan, 0 );
    var pObj *Gia_Obj_t;
    var pSat *sat_solver;
    var i, status int32;
    assert(p.pSat == nil && p.vId2Var == nil, "p.pSat == nil && p.vId2Var == nil");
    assert(Aig_ManObjNumMax(pMan) == Gia_ManObjNum(p.pFraig), "Aig_ManObjNumMax(pMan) == Gia_ManObjNum(p.pFraig)");
    Aig_ManStop( pMan );
    // save variable mapping
    p.nSatVarsPivot = p.nSatVars = pDat.nVars;
    p.vId2Var = Vec_IntStart( Gia_ManCandNum(p.pAig) + Gia_ManCandNum(p.pCare) + 10 ); // mapping of each node into its SAT var
    p.vVar2Id = Vec_IntStart( Gia_ManCandNum(p.pAig) + Gia_ManCandNum(p.pCare) + 10 ); // mapping of each SAT var into its node
    Ssc_ObjSetSatVar( p, 0, pDat.pVarNums[0] );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pFraig.vCis)) && ((pObj) = Gia_ManCi(p.pFraig, i)); i++  {
        var iObj int32  = Gia_ObjId( p.pFraig, pObj );
        Ssc_ObjSetSatVar( p, iObj, pDat.pVarNums[iObj] );
    }
//Cnf_DataWriteIntoFile( pDat, "dump.cnf", 1, NULL, NULL );
    // start the SAT solver
    pSat = sat_solver_new();
    sat_solver_setnvars( pSat, pDat.nVars + 1000 );
    for  i = 0; i < pDat.nClauses; i++  {
        if !sat_solver_addclause( pSat, pDat.pClauses[i], pDat.pClauses[i+1] ) {
            Cnf_DataFree( pDat );
            sat_solver_delete( pSat );
            return;
        }
    }
    Cnf_DataFree( pDat );
    status = sat_solver_simplify( pSat );
    if status == 0 {
        sat_solver_delete( pSat );
        return;
    }
    p.pSat = pSat;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_ManCollectSatPattern(p *Ssc_Man_t, vPattern *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Vec_IntClear( vPattern );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pFraig.vCis)) && ((pObj) = Gia_ManCi(p.pFraig, i)); i++  {
        Vec_IntPush( vPattern, sat_solver_var_value(p.pSat, Ssc_ObjSatVar(p, Gia_ObjId(p.pFraig, pObj))) );
    }
}
func Ssc_ManFindPivotSat(p *Ssc_Man_t) *Vec_Int_t {
    var vInit *Vec_Int_t;
    var status int32  = sat_solver_solve( p.pSat, nil, nil, p.pPars.nBTLimit, 0, 0, 0 );
    if status == l_False {
        return (Vec_Int_t *)(ABC_PTRINT_T)1; }
    if status == l_Undef {
        return nil; }
    assert(status == l_True, "status == l_True");
    vInit = Vec_IntAlloc( Gia_ManCiNum(p.pFraig) );
    Ssc_ManCollectSatPattern( p, vInit );
    return vInit;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_ManCheckEquivalence(p *Ssc_Man_t, iRepr int32, iNode int32, fCompl int32) int32 {
    var pLitsSat [2]int32;
    var RetValue int32;
    var clk abctime;
    assert(iRepr != iNode, "iRepr != iNode");
    if iRepr > iNode {
        return l_Undef; }
    assert(iRepr < iNode, "iRepr < iNode");
//    if ( p->nTimeOut )
//        sat_solver_set_runtime_limit( p->pSat, p->nTimeOut * CLOCKS_PER_SEC + Abc_Clock() );

    // create CNF
    if iRepr != 0 {
    Ssc_ManCnfNodeAddToSolver( p, iRepr ); }
    Ssc_ManCnfNodeAddToSolver( p, iNode );
    sat_solver_compress( p.pSat );

    // order the literals
    pLitsSat[0] = Abc_Var2Lit( Ssc_ObjSatVar(p, iRepr), 0 );
    pLitsSat[1] = Abc_Var2Lit( Ssc_ObjSatVar(p, iNode), fCompl ^ int32((iRepr > 0)) );

    // solve under assumptions
    // A = 1; B = 0
    clk = Abc_Clock();
    RetValue = sat_solver_solve( p.pSat, pLitsSat, pLitsSat + 2, (ABC_INT64_T)p.pPars.nBTLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if RetValue == l_False {
        pLitsSat[0] = Abc_LitNot( pLitsSat[0] ); // compl
        pLitsSat[1] = Abc_LitNot( pLitsSat[1] ); // compl
        RetValue = sat_solver_addclause( p.pSat, pLitsSat, pLitsSat + 2 );
        assert((RetValue != 0), "RetValue");
        p.timeSatUnsat += Abc_Clock() - clk;
    
    } else if RetValue == l_True {
        Ssc_ManCollectSatPattern( p, p.vPattern );
        p.timeSatSat += Abc_Clock() - clk;
        return l_True;
    
    } else { // if ( RetValue1 == l_Undef )
    
        p.timeSatUndec += Abc_Clock() - clk;
        return l_Undef;
    }

    // if the old node was constant 0, we already know the answer
    if iRepr == 0 {
        return l_False; }

    // solve under assumptions
    // A = 0; B = 1
    clk = Abc_Clock();
    RetValue = sat_solver_solve( p.pSat, pLitsSat, pLitsSat + 2, (ABC_INT64_T)p.pPars.nBTLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if RetValue == l_False {
        pLitsSat[0] = Abc_LitNot( pLitsSat[0] );
        pLitsSat[1] = Abc_LitNot( pLitsSat[1] );
        RetValue = sat_solver_addclause( p.pSat, pLitsSat, pLitsSat + 2 );
        assert((RetValue != 0), "RetValue");
        p.timeSatUnsat += Abc_Clock() - clk;
    
    } else if RetValue == l_True {
        Ssc_ManCollectSatPattern( p, p.vPattern );
        p.timeSatSat += Abc_Clock() - clk;
        return l_True;
    
    } else { // if ( RetValue1 == l_Undef )
    
        p.timeSatUndec += Abc_Clock() - clk;
        return l_Undef;
    }
    return l_False;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

