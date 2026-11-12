package cec

import "os"

/**CFile****************************************************************

  FileName    [cecSolve.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Performs one round of SAT solving.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSolve.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Cec_ObjSatNum(p *Cec_ManSat_t, pObj *Gia_Obj_t) int32 { return p.pSatVars[Gia_ObjId(p.pAig,pObj)]; }
func Cec_ObjSetSatNum(p *Cec_ManSat_t, pObj *Gia_Obj_t, Num int32) { p.pSatVars[Gia_ObjId(p.pAig,pObj)] = Num;  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns value of the SAT variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_AddClausesMux(p *Cec_ManSat_t, pNode *Gia_Obj_t) {
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
    assert(Gia_ObjIsMuxType( pNode ), "Gia_ObjIsMuxType( pNode )");
    // get nodes (I = if, T = then, E = else)
    pNodeI = Gia_ObjRecognizeMux( pNode, &pNodeT, &pNodeE );
    // get the variable numbers
    VarF = Cec_ObjSatNum(p,pNode);
    VarI = Cec_ObjSatNum(p,pNodeI);
    VarT = Cec_ObjSatNum(p,Gia_Regular(pNodeT));
    VarE = Cec_ObjSatNum(p,Gia_Regular(pNodeE));
    // get the complementation flags
    fCompT = Gia_IsComplement(pNodeT);
    fCompE = Gia_IsComplement(pNodeE);

    // f = ITE(i, t, e)

    // i' + t' + f
    // i' + t  + f'
    // i  + e' + f
    // i  + e  + f'

    // create four clauses
    pLits[0] = toLitCond(VarI, 1);
    pLits[1] = toLitCond(VarT, 1^fCompT);
    pLits[2] = toLitCond(VarF, 0);
    if p.pPars.fPolarFlip {
        if pNodeI.fPhase {               pLits[0] = lit_neg( pLits[0] ); }
        if Gia_Regular(pNodeT).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 1);
    pLits[1] = toLitCond(VarT, 0^fCompT);
    pLits[2] = toLitCond(VarF, 1);
    if p.pPars.fPolarFlip {
        if pNodeI.fPhase {               pLits[0] = lit_neg( pLits[0] ); }
        if Gia_Regular(pNodeT).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 0);
    pLits[1] = toLitCond(VarE, 1^fCompE);
    pLits[2] = toLitCond(VarF, 0);
    if p.pPars.fPolarFlip {
        if pNodeI.fPhase {               pLits[0] = lit_neg( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarI, 0);
    pLits[1] = toLitCond(VarE, 0^fCompE);
    pLits[2] = toLitCond(VarF, 1);
    if p.pPars.fPolarFlip {
        if pNodeI.fPhase {               pLits[0] = lit_neg( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
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
    if p.pPars.fPolarFlip {
        if Gia_Regular(pNodeT).fPhase {  pLits[0] = lit_neg( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = lit_neg( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = toLitCond(VarT, 1^fCompT);
    pLits[1] = toLitCond(VarE, 1^fCompE);
    pLits[2] = toLitCond(VarF, 0);
    if p.pPars.fPolarFlip {
        if Gia_Regular(pNodeT).fPhase {  pLits[0] = lit_neg( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = lit_neg( pLits[1] ); }
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
func Cec_AddClausesSuper(p *Cec_ManSat_t, pNode *Gia_Obj_t, vSuper *Vec_Ptr_t) {
    var pFanin *Gia_Obj_t;
    var pLits *int32;
    var nLits int32;
    var RetValue int32;
    var i int32;
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    assert(Gia_ObjIsAnd( pNode ), "Gia_ObjIsAnd( pNode )");
    // create storage for literals
    nLits = Vec_PtrSize(vSuper) + 1;
    // inlining of macro ABC_ALLOC
    pLits = ((int *) malloc(sizeof(int) * (size_t)(nLits)));
    // suppose AND-gate is A & B = C
    // add !A => !C   or   A + !C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[0] = toLitCond(Cec_ObjSatNum(p,Gia_Regular(pFanin)), Gia_IsComplement(pFanin));
        pLits[1] = toLitCond(Cec_ObjSatNum(p,pNode), 1);
        if p.pPars.fPolarFlip {
            if Gia_Regular(pFanin).fPhase {  pLits[0] = lit_neg( pLits[0] ); }
            if pNode.fPhase {                pLits[1] = lit_neg( pLits[1] ); }
        }
        RetValue = sat_solver_addclause( p.pSat, pLits, pLits + 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[i] = toLitCond(Cec_ObjSatNum(p,Gia_Regular(pFanin)), !Gia_IsComplement(pFanin));
        if p.pPars.fPolarFlip {
            if Gia_Regular(pFanin).fPhase {  pLits[i] = lit_neg( pLits[i] ); }
        }
    }
    pLits[nLits-1] = toLitCond(Cec_ObjSatNum(p,pNode), 0);
    if p.pPars.fPolarFlip {
        if pNode.fPhase {  pLits[nLits-1] = lit_neg( pLits[nLits-1] ); }
    }
    RetValue = sat_solver_addclause( p.pSat, pLits, pLits + nLits );
    assert((RetValue != 0), "RetValue");
    pLits = nil;
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_CollectSuper_rec(pObj *Gia_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Gia_IsComplement(pObj) || Gia_ObjIsCi(pObj) || 
         (!(fFirst != 0) && Gia_ObjValue(pObj) > 1) || 
         ((fUseMuxes != 0) && Gia_ObjIsMuxType(pObj)) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
    // go through the branches
    Cec_CollectSuper_rec( Gia_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Cec_CollectSuper_rec( Gia_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_CollectSuper(pObj *Gia_Obj_t, fUseMuxes int32, vSuper *Vec_Ptr_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsCi(pObj), "!Gia_ObjIsCi(pObj)");
    Vec_PtrClear( vSuper );
    Cec_CollectSuper_rec( pObj, vSuper, 1, fUseMuxes );
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ObjAddToFrontier(p *Cec_ManSat_t, pObj *Gia_Obj_t, vFrontier *Vec_Ptr_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    if Cec_ObjSatNum(p,pObj) != 0 {
        return; }
    assert(Cec_ObjSatNum(p,pObj) == 0, "Cec_ObjSatNum(p,pObj) == 0");
    if Gia_ObjIsConst0(pObj) {
        return; }
    Vec_PtrPush( p.vUsedNodes, pObj );
    Cec_ObjSetSatNum( p, pObj, p.nSatVars++ );
    if Gia_ObjIsAnd(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}

/**Function*************************************************************

  Synopsis    [Updates the solver clause database.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_CnfNodeAddToSolver(p *Cec_ManSat_t, pObj *Gia_Obj_t) { 
    var vFrontier *Vec_Ptr_t;
    var pNode, pFanin *Gia_Obj_t;
    var i int32;
    var k int32;
    var fUseMuxes int32 = 1;
    // quit if CNF is ready
    if Cec_ObjSatNum(p,pObj) != 0 {
        return; }
    if Gia_ObjIsCi(pObj) {
        Vec_PtrPush( p.vUsedNodes, pObj );
        Cec_ObjSetSatNum( p, pObj, p.nSatVars++ );
        sat_solver_setnvars( p.pSat, p.nSatVars );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // start the frontier
    vFrontier = Vec_PtrAlloc( 100 );
    Cec_ObjAddToFrontier( p, pObj, vFrontier );
    // explore nodes in the frontier
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFrontier)) && (((pNode) = (Gia_Obj_t *)Vec_PtrEntry(vFrontier, i)), 1); i++  {
        // create the supergate
        assert((Cec_ObjSatNum(p,pNode) != 0), "Cec_ObjSatNum(p,pNode)");
        if (fUseMuxes != 0) && Gia_ObjIsMuxType(pNode) {
            Vec_PtrClear( p.vFanins );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin1(pNode) ) );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(p.vFanins, k)), 1); k++  {
                Cec_ObjAddToFrontier( p, Gia_Regular(pFanin), vFrontier );
            }
            Cec_AddClausesMux( p, pNode );
        
        } else {
            Cec_CollectSuper( pNode, fUseMuxes, p.vFanins );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(p.vFanins, k)), 1); k++  {
                Cec_ObjAddToFrontier( p, Gia_Regular(pFanin), vFrontier );
            }
            Cec_AddClausesSuper( p, pNode, p.vFanins );
        }
        assert(Vec_PtrSize(p.vFanins) > 1, "Vec_PtrSize(p.vFanins) > 1");
    }
    Vec_PtrFree( vFrontier );
}


/**Function*************************************************************

  Synopsis    [Recycles the SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Sets variable activities in the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SetActivityFactors_rec(p *Cec_ManSat_t, pObj *Gia_Obj_t, LevelMin int32, LevelMax int32) {
    var dActConeBumpMax float32  = 20.0;
    var iVar int32;
    // skip visited variables
    if Gia_ObjIsTravIdCurrent(p.pAig, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p.pAig, pObj);
    // add the PI to the list
    if Gia_ObjLevel(p.pAig, pObj) <= LevelMin || Gia_ObjIsCi(pObj) {
        return; }
    // set the factor of this variable
    // (LevelMax-LevelMin) / (pObj->Level-LevelMin) = p->pPars->dActConeBumpMax / ThisBump
    if iVar = Cec_ObjSatNum(p,pObj) {
        p.pSat.factors[iVar] = dActConeBumpMax * (Gia_ObjLevel(p.pAig, pObj) - LevelMin)/(LevelMax - LevelMin);
        veci_push(&p.pSat.act_vars, iVar);
    }
    // explore the fanins
    Cec_SetActivityFactors_rec( p, Gia_ObjFanin0(pObj), LevelMin, LevelMax );
    Cec_SetActivityFactors_rec( p, Gia_ObjFanin1(pObj), LevelMin, LevelMax );
}

/**Function*************************************************************

  Synopsis    [Sets variable activities in the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SetActivityFactors(p *Cec_ManSat_t, pObj *Gia_Obj_t) int32 {
    var dActConeRatio float32  = 0.5;
    var LevelMin, LevelMax int32;
    // reset the active variables
    veci_resize(&p.pSat.act_vars, 0);
    // prepare for traversal
    Gia_ManIncrementTravId( p.pAig );
    // determine the min and max level to visit
    assert(dActConeRatio > 0 && dActConeRatio < 1, "dActConeRatio > 0 && dActConeRatio < 1");
    LevelMax = Gia_ObjLevel(p.pAig,pObj);
    LevelMin = int32((LevelMax * (1.0 - dActConeRatio)));
    // traverse
    Cec_SetActivityFactors_rec( p, pObj, LevelMin, LevelMax );
//Cec_PrintActivity( p );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Runs equivalence test for the two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Runs equivalence test for the two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Performs one round of solving for the POs of the AIG.]

  Description [Labels the nodes that have been proved (pObj->fMark1) 
  and returns the set of satisfying assignments.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


func Cec_ManSatSolve(pPat *Cec_ManPat_t, pAig *Gia_Man_t, pPars *Cec_ParSat_t, vIdsOrig *Vec_Int_t, vMiterPairs *Vec_Int_t, vEquivPairs *Vec_Int_t, f0Proved int32) {
    var pProgress *Bar_Progress_t  = nil;
    var p *Cec_ManSat_t;
    var pObj *Gia_Obj_t;
    var i, status int32;
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    Vec_PtrFreeP( &pAig.vSeqModelVec );
    if pPars.fSaveCexes {
        pAig.vSeqModelVec = Vec_PtrStart( Gia_ManCoNum(pAig) ); }
    // reset the manager
    if pPat != nil {
        pPat.iStart = Vec_StrSize(pPat.vStorage);
        pPat.nPats = 0;
        pPat.nPatLits = 0;
        pPat.nPatLitsMin = 0;
    } 
    Gia_ManSetPhase( pAig );
    Gia_ManLevelNum( pAig );
    Gia_ManIncrementTravId( pAig );
    p = Cec_ManSatCreate( pAig, pPars );
    pProgress = Bar_ProgressStart( os.Stdout, Gia_ManPoNum(pAig) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pAig.vCos)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
        if Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) {
            status = !Gia_ObjFaninC0(pObj);
            pObj.fMark0 = (status == 0);
            pObj.fMark1 = (status == 1);
            if pPars.fSaveCexes {
                Vec_PtrWriteEntry( pAig.vSeqModelVec, i, status ? (Abc_Cex_t *)(ABC_PTRINT_T)1 : Cex_ManGenSimple(p, i) ); }
            continue;
        }
        Bar_ProgressUpdate( pProgress, i, "SAT..." );
clk2 = Abc_Clock();
        status = Cec_ManSatCheckNode( p, Gia_ObjChild0(pObj) );
        pObj.fMark0 = (status == 0);
        pObj.fMark1 = (status == 1);
        if status == 1 && (vIdsOrig != nil) {
            var iObj1 int32  = Vec_IntEntry(vMiterPairs, 2*i);
            var iObj2 int32  = Vec_IntEntry(vMiterPairs, 2*i+1);
            var OrigId1 int32  = Vec_IntEntry(vIdsOrig, iObj1);
            var OrigId2 int32  = Vec_IntEntry(vIdsOrig, iObj2);
            assert(OrigId1 >= 0 && OrigId2 >= 0, "OrigId1 >= 0 && OrigId2 >= 0");
            Vec_IntPushTwo( vEquivPairs, OrigId1, OrigId2 );
        }
        if pPars.fSaveCexes && status != -1 {
            Vec_PtrWriteEntry( pAig.vSeqModelVec, i, status ? (Abc_Cex_t *)(ABC_PTRINT_T)1 : Cex_ManGenCex(p, i) ); }

        if (f0Proved != 0) && status == 1 {
            Gia_ManPatchCoDriver( pAig, i, 0 ); }

/*
        if ( status == -1 )
        {
            Gia_Man_t * pTemp = Gia_ManDupDfsCone( pAig, pObj );
            Gia_AigerWrite( pTemp, "gia_hard.aig", 0, 0, 0 );
            Gia_ManStop( pTemp );
            Abc_Print( 1, "Dumping hard cone into file \"%s\".\n", "gia_hard.aig" );
        }
*/
        if status != 0 {
            continue; }
        // save the pattern
        if pPat != nil {
            var clk3 abctime  = Abc_Clock();
            Cec_ManPatSavePattern( pPat, p, pObj );
            pPat.timeTotalSave += Abc_Clock() - clk3;
        }
        // quit if one of them is solved
        if pPars.fCheckMiter {
            break; }
    }
    p.timeTotal = Abc_Clock() - clk;
    Bar_ProgressStop( pProgress );
    if pPars.fVerbose {
        Cec_ManSatPrintStats( p ); }
    Cec_ManSatStop( p );
}

/**Function*************************************************************

  Synopsis    [Performs one round of solving for the POs of the AIG.]

  Description [Labels the nodes that have been proved (pObj->fMark1) 
  and returns the set of satisfying assignments.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSatSolveExractPattern(vCexStore *Vec_Int_t, iStart int32, vPat *Vec_Int_t) int32 {
    var k, nSize int32;
    Vec_IntClear( vPat );
    // skip the output number
    iStart++;
    // get the number of items
    nSize = Vec_IntEntry( vCexStore, iStart++ );
    if nSize <= 0 {
        return iStart; }
    // extract pattern
    for  k = 0; k < nSize; k++  {
        Vec_IntPush( vPat, Vec_IntEntry( vCexStore, iStart++ ) );
    }
    return iStart;
}
func Cec_ManSatSolveCSat(pPat *Cec_ManPat_t, pAig *Gia_Man_t, pPars *Cec_ParSat_t) {
    var vStatus *Vec_Str_t;
    var vPat *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vCexStore *Vec_Int_t  = Cbs_ManSolveMiterNc( pAig, pPars.nBTLimit, &vStatus, 0, 0 );
    var pObj *Gia_Obj_t;
    var i, status, iStart int32;
    assert(Vec_StrSize(vStatus) == Gia_ManCoNum(pAig), "Vec_StrSize(vStatus) == Gia_ManCoNum(pAig)");
    // reset the manager
    if pPat != nil {
        pPat.iStart = Vec_StrSize(pPat.vStorage);
        pPat.nPats = 0;
        pPat.nPatLits = 0;
        pPat.nPatLitsMin = 0;
    } 
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pAig.vCos)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
        status = int32(Vec_StrEntry(vStatus, i));
        pObj.fMark0 = (status == 0);
        pObj.fMark1 = (status == 1);
        if Vec_IntSize(vCexStore) > 0 && status != 1 {
            iStart = Cec_ManSatSolveExractPattern( vCexStore, iStart, vPat ); }
        if status != 0 {
            continue; }
        // save the pattern
        if (pPat != nil) && Vec_IntSize(vPat) > 0 {
            var clk3 abctime  = Abc_Clock();
            Cec_ManPatSavePatternCSat( pPat, vPat );
            pPat.timeTotalSave += Abc_Clock() - clk3;
        }
        // quit if one of them is solved
        if pPars.fCheckMiter {
            break; }
    }
    assert(iStart == Vec_IntSize(vCexStore), "iStart == Vec_IntSize(vCexStore)");
    Vec_IntFree( vPat );
    Vec_StrFree( vStatus );
    Vec_IntFree( vCexStore );
}



/**Function*************************************************************

  Synopsis    [Returns the pattern stored.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Save values in the cone of influence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs one round of solving for the POs of the AIG.]

  Description [Labels the nodes that have been proved (pObj->fMark1) 
  and returns the set of satisfying assignments.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Save values in the cone of influence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Save values in the cone of influence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSatSolveMiter_rec(pSat *Cec_ManSat_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        pSat.nCexLits++;
        Vec_IntPush( pSat.vCex, Abc_Var2Lit( Gia_ObjCioId(pObj), !Cec_ObjSatVarValue(pSat, pObj) ) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Cec_ManSatSolveMiter_rec( pSat, p, Gia_ObjFanin0(pObj) );
    Cec_ManSatSolveMiter_rec( pSat, p, Gia_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Save patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs one round of solving for the POs of the AIG.]

  Description [Labels the nodes that have been proved (pObj->fMark1) 
  and returns the set of satisfying assignments.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

