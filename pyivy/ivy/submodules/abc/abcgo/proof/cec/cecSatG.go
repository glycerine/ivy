package cec

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [cecSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Detection of structural isomorphism.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "misc/util/utilTruth.h"
//#include "sat/glucose/AbcGlucose.h"
//#include "cec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// sweeping manager
   

// SAT solving manager
   
type Cec3_Man_t struct {
    pPars      *Cec3_Par_t;          // parameters
    pAig       *Gia_Man_t;           // user's AIG
    pNew       *Gia_Man_t;           // internal AIG
    // SAT solving
    pSat *bmcg_sat_solver;           // SAT solver
    vFrontier       *Vec_Ptr_t;      // CNF construction
    vFanins       *Vec_Ptr_t;        // CNF construction
    vSims       *Vec_Wrd_t;          // CI simulation info
    vNodesNew       *Vec_Int_t;      // nodes
    vSatVars       *Vec_Int_t;       // nodes
    vObjSatPairs       *Vec_Int_t;   // nodes
    vCexTriples       *Vec_Int_t;    // nodes
    // statistics
    nPatterns              int32;
    nSatSat              int32;
    nSatUnsat              int32;
    nSatUndec              int32;
    timeSatSat          abctime;
    timeSatUnsat          abctime;
    timeSatUndec          abctime;
    timeSim          abctime;
    timeRefine          abctime;
    timeExtra          abctime;
    timeStart          abctime;
}

func Cec3_ObjSatId(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjCopy2Array(p, Gia_ObjId(p, pObj));                                                     }
func Cec3_ObjSetSatId(p *Gia_Man_t, pObj *Gia_Obj_t, Num int32) int32 { assert(Cec3_ObjSatId(p, pObj) == -1, "Cec3_ObjSatId(p, pObj) == -1"); Gia_ObjSetCopy2Array(p, Gia_ObjId(p, pObj), Num); return Num;  }
func Cec3_ObjCleanSatId(p *Gia_Man_t, pObj *Gia_Obj_t) { assert(Cec3_ObjSatId(p, pObj) != -1, "Cec3_ObjSatId(p, pObj) != -1"); Gia_ObjSetCopy2Array(p, Gia_ObjId(p, pObj), -1);               }

func satoko_mark_cone(p *bmcg_sat_solver, pVars *int32, nVars int32) {}
func satoko_unmark_cone(p *bmcg_sat_solver, pVars *int32, nVars int32) {}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets parameter defaults.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_AddClausesMux(p *Gia_Man_t, pNode *Gia_Obj_t, pSat *bmcg_sat_solver) {
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
    VarF = Cec3_ObjSatId(p, pNode);
    VarI = Cec3_ObjSatId(p, pNodeI);
    VarT = Cec3_ObjSatId(p, Gia_Regular(pNodeT));
    VarE = Cec3_ObjSatId(p, Gia_Regular(pNodeE));
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
    RetValue = bmcg_sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 1);
    pLits[1] = Abc_Var2Lit(VarT, 0^fCompT);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeT).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = bmcg_sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 0);
    pLits[1] = Abc_Var2Lit(VarE, 1^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = bmcg_sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 0);
    pLits[1] = Abc_Var2Lit(VarE, 0^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = bmcg_sat_solver_addclause( pSat, pLits, 3 );
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
    RetValue = bmcg_sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarT, 1^fCompT);
    pLits[1] = Abc_Var2Lit(VarE, 1^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if Gia_Regular(pNodeT).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = bmcg_sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
}
func Cec3_AddClausesSuper(p *Gia_Man_t, pNode *Gia_Obj_t, vSuper *Vec_Ptr_t, pSat *bmcg_sat_solver) {
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
    // inlining of macro ABC_ALLOC
    pLits = ((int *) malloc(sizeof(int) * (size_t)(nLits)));
    // suppose AND-gate is A & B = C
    // add !A => !C   or   A + !C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[0] = Abc_Var2Lit(Cec3_ObjSatId(p, Gia_Regular(pFanin)), Gia_IsComplement(pFanin));
        pLits[1] = Abc_Var2Lit(Cec3_ObjSatId(p, pNode), 1);
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
            if pNode.fPhase {                pLits[1] = Abc_LitNot( pLits[1] ); }
        }
        RetValue = bmcg_sat_solver_addclause( pSat, pLits, 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[i] = Abc_Var2Lit(Cec3_ObjSatId(p, Gia_Regular(pFanin)), !Gia_IsComplement(pFanin));
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[i] = Abc_LitNot( pLits[i] ); }
        }
    }
    pLits[nLits-1] = Abc_Var2Lit(Cec3_ObjSatId(p, pNode), 0);
    if fPolarFlip != 0 {
        if pNode.fPhase {  pLits[nLits-1] = Abc_LitNot( pLits[nLits-1] ); }
    }
    RetValue = bmcg_sat_solver_addclause( pSat, pLits, nLits );
    assert((RetValue != 0), "RetValue");
    pLits = nil;
}

/**Function*************************************************************

  Synopsis    [Adds clauses and returns CNF variable of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_CollectSuper_rec(pObj *Gia_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Gia_IsComplement(pObj) || Gia_ObjIsCi(pObj) || 
         (!(fFirst != 0) && Gia_ObjValue(pObj) > 1) || 
         ((fUseMuxes != 0) && pObj.fMark0) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
    // go through the branches
    Cec3_CollectSuper_rec( Gia_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Cec3_CollectSuper_rec( Gia_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}
func Cec3_CollectSuper(pObj *Gia_Obj_t, fUseMuxes int32, vSuper *Vec_Ptr_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsCi(pObj), "!Gia_ObjIsCi(pObj)");
    Vec_PtrClear( vSuper );
    Cec3_CollectSuper_rec( pObj, vSuper, 1, fUseMuxes );
}
func Cec3_ObjAddToFrontier(p *Gia_Man_t, pObj *Gia_Obj_t, vFrontier *Vec_Ptr_t, pSat *bmcg_sat_solver) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsConst0(pObj), "!Gia_ObjIsConst0(pObj)");
    if Cec3_ObjSatId(p, pObj) >= 0 {
        return; }
    assert(Cec3_ObjSatId(p, pObj) == -1, "Cec3_ObjSatId(p, pObj) == -1");
    Cec3_ObjSetSatId( p, pObj, bmcg_sat_solver_addvar(pSat) );
    if Gia_ObjIsAnd(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}
func Cec3_ObjGetCnfVar(p *Cec3_Man_t, iObj int32) int32 { 
    var pNode, pFanin *Gia_Obj_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pNew, iObj);
    var i int32;
    var k int32;
    var fUseMuxes int32 = 1;
    // quit if CNF is ready
    if Cec3_ObjSatId(p.pNew,pObj) >= 0 {
        return Cec3_ObjSatId(p.pNew,pObj); }
    assert(iObj > 0, "iObj > 0");
    if Gia_ObjIsCi(pObj) {
        return Cec3_ObjSetSatId( p.pNew, pObj, bmcg_sat_solver_addvar(p.pSat) ); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // start the frontier
    Vec_PtrClear( p.vFrontier );
    Cec3_ObjAddToFrontier( p.pNew, pObj, p.vFrontier, p.pSat );
    // explore nodes in the frontier
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFrontier)) && (((pNode) = (Gia_Obj_t *)Vec_PtrEntry(p.vFrontier, i)), 1); i++  {
        // create the supergate
        assert(Cec3_ObjSatId(p.pNew,pNode) >= 0, "Cec3_ObjSatId(p.pNew,pNode) >= 0");
        if (fUseMuxes != 0) && pNode.fMark0 {
            Vec_PtrClear( p.vFanins );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin1(pNode) ) );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(p.vFanins, k)), 1); k++  {
                Cec3_ObjAddToFrontier( p.pNew, Gia_Regular(pFanin), p.vFrontier, p.pSat );
            }
            Cec3_AddClausesMux( p.pNew, pNode, p.pSat );
        
        } else {
            Cec3_CollectSuper( pNode, fUseMuxes, p.vFanins );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(p.vFanins, k)), 1); k++  {
                Cec3_ObjAddToFrontier( p.pNew, Gia_Regular(pFanin), p.vFrontier, p.pSat );
            }
            Cec3_AddClausesSuper( p.pNew, pNode, p.vFanins, p.pSat );
        }
        assert(Vec_PtrSize(p.vFanins) > 1, "Vec_PtrSize(p.vFanins) > 1");
    }
    return Cec3_ObjSatId(p.pNew,pObj);
}



/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_ObjSim(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, p.nSimWords * iObj );
}
func Cec3_ObjSimSetInputBit(p *Gia_Man_t, iObj int32, Bit int32) {
    var pSim *word  = Cec3_ObjSim( p, iObj );
    if Abc_InfoHasBit( (unsigned*)pSim, p.iPatsPi ) != Bit {
        Abc_InfoXorBit( (unsigned*)pSim, p.iPatsPi ); }
}
func Cec3_ObjSimRo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSimRo *word  = Cec3_ObjSim( p, iObj );
    var pSimRi *word  = Cec3_ObjSim( p, Gia_ObjRoToRiId(p, iObj) );
    for  w = 0; w < p.nSimWords; w++  {
        pSimRo[w] = pSimRi[w];
    }
}
func Cec3_ObjSimCo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSimCo *word   = Cec3_ObjSim( p, iObj );
    var pSimDri *word  = Cec3_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] = ~pSimDri[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] =  pSimDri[w];
        } }
}
func Cec3_ObjSimAnd(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSim *word   = Cec3_ObjSim( p, iObj );
    var pSim0 *word  = Cec3_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSim1 *word  = Cec3_ObjSim( p, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & ~pSim1[w];
        }
    } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & pSim1[w];
        }
    } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & ~pSim1[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & pSim1[w];
        } }
}
func Cec3_ObjSimEqual(p *Gia_Man_t, iObj0 int32, iObj1 int32) int32 {
    var w int32;
    var pSim0 *word  = Cec3_ObjSim( p, iObj0 );
    var pSim1 *word  = Cec3_ObjSim( p, iObj1 );
    if (pSim0[0] & 1) == (pSim1[0] & 1) {
        for  w = 0; w < p.nSimWords; w++  {
            if pSim0[w] != pSim1[w] {
                return 0; }
        }
        return 1;
    
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            if pSim0[w] != ~pSim1[w] {
                return 0; }
        }
        return 1;
    }
}
func Cec3_ObjSimCi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSim *word  = Cec3_ObjSim( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSim[w] = Gia_ManRandomW( 0 );
    }
    pSim[0] <<= 1;
}
func Cec3_ManSimulateCis(p *Gia_Man_t) {
    var i, Id int32;
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Cec3_ObjSimCi( p, Id );
    }
    p.iPatsPi = 0;
}
func Cec3_ManDeriveCex(p *Gia_Man_t, iOut int32, iPat int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i, Id int32;
    pCex = Abc_CexAlloc( 0, Gia_ManCiNum(p), 1 );
    pCex.iPo = iOut;
    if iPat == -1 {
        return pCex; }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        if Abc_InfoHasBit((unsigned *)Cec3_ObjSim(p, Id), iPat) {
            Abc_InfoSetBit( pCex.pData, i ); }
    }
    return pCex;
}
func Cec3_ManSimulateCos(p *Gia_Man_t) int32 {
    var i, Id int32;
    // check outputs and generate CEX if they fail
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        Cec3_ObjSimCo( p, Id );
        if Cec3_ObjSimEqual(p, Id, 0) != 0 {
            continue; }
        p.pCexSeq = Cec3_ManDeriveCex( p, i, Abc_TtFindFirstBit2(Cec3_ObjSim(p, Id), p.nSimWords) );
        return 0;
    }
    return 1;
}
func Cec3_ManSaveCis(p *Gia_Man_t) {
    var w, i, Id int32;
    assert(p.vSimsPi != nil, "p.vSimsPi != nil");
    for  w = 0; w < p.nSimWords; w++  {
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
            Vec_WrdPush( p.vSimsPi, Cec3_ObjSim(p, Id)[w] );
        }
    }
}
func Cec3_ManSimulate(p *Gia_Man_t, vTriples *Vec_Int_t, pMan *Cec3_Man_t) int32 {
    extern void Cec3_ManSimClassRefineOne( Gia_Man_t * p, int iRepr );
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; 
    var i, iRepr, iObj, Entry, Count int32;
    //Cec3_ManSaveCis( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Cec3_ObjSimAnd( p, i ); }
    }
    pMan.timeSim += Abc_Clock() - clk;
    if p.pReprs == nil {
        return 0; }
    if vTriples != nil {
        // inlining of macro Vec_IntForEachEntryTriple
        for  i = 0; (i+2 < Vec_IntSize(vTriples)) && (((iRepr) = Vec_IntEntry(vTriples, i)), 1) && (((iObj) = Vec_IntEntry(vTriples, i+1)), 1) && (((Entry) = Vec_IntEntry(vTriples, i+2)), 1); i += 3  {
            var pSim0 *word  = Cec3_ObjSim( p, iRepr );
            var pSim1 *word  = Cec3_ObjSim( p, iObj );
            var iPat int32      = Abc_Lit2Var(Entry);
            var fPhase int32    = Abc_LitIsCompl(Entry);
            if (fPhase ^ Abc_InfoHasBit((unsigned *)pSim0, iPat)) == Abc_InfoHasBit((unsigned *)pSim1, iPat) {
                //printf( "ERROR:  Pattern %d did not disprove pair %d and %d.\n", iPat, iRepr, iObj );
                Count++;
            }
        }
    }
    clk = Abc_Clock();
    // inlining of macro Gia_ManForEachClass0
    for  i = 0; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Cec3_ManSimClassRefineOne( p, i ); }
    }
    pMan.timeRefine += Abc_Clock() - clk;
    return Count;
}
func Cec3_ManSimAlloc(p *Gia_Man_t, nWords int32) {
    Vec_WrdFreeP( &p.vSims );
    Vec_WrdFreeP( &p.vSimsPi );
    p.vSims   = Vec_WrdStart( Gia_ManObjNum(p) * nWords );
    p.vSimsPi = Vec_WrdAlloc( Gia_ManCiNum(p) * nWords * 4 ); // storage for CI patterns
    p.nSimWords = nWords;
}


/**Function*************************************************************

  Synopsis    [Computes hash key of the simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creating initial equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_ManSimClassRefineOne(p *Gia_Man_t, iRepr int32) {
    var iObj int32;
    var iPrev int32 = iRepr;
    var iPrev2 int32;
    var iRepr2 int32;
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iRepr2 = Gia_ObjNext(p, iRepr); iRepr2 > 0; iRepr2 = Gia_ObjNext(p, iRepr2)  {
        if Cec3_ObjSimEqual(p, iRepr, iRepr2) != 0 {
            iPrev = iRepr2;
        } else {
            break; }
    }
    if iRepr2 <= 0 { // no refinement
        return; }
    // relink remaining nodes of the class
    // nodes that are equal to iRepr, remain in the class of iRepr
    // nodes that are not equal to iRepr, move to the class of iRepr2
    Gia_ObjSetRepr( p, iRepr2, GIA_VOID );
    iPrev2 = iRepr2;
    for  iObj = Gia_ObjNext(p, iRepr2); iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
        if Cec3_ObjSimEqual(p, iRepr, iObj) != 0 { // remains with iRepr
        
            Gia_ObjSetNext( p, iPrev, iObj );
            iPrev = iObj;
        
        } else { // moves to iRepr2
        
            Gia_ObjSetRepr( p, iObj, iRepr2 );
            Gia_ObjSetNext( p, iPrev2, iObj );
            iPrev2 = iObj;
        }
    }
    Gia_ObjSetNext( p, iPrev, -1 );
    Gia_ObjSetNext( p, iPrev2, -1 );
}
func Cec3_ManCreateClasses(p *Gia_Man_t, pMan *Cec3_Man_t) {
    var clk abctime;
    var pObj *Gia_Obj_t;
    var nWords int32  = p.nSimWords;
    var pTable *int32;
    var nTableSize int32;
    var i int32;
    var Key int32;
    // allocate representation
    p.pReprs = nil;
    p.pNexts = nil;
    // inlining of macro ABC_CALLOC
    p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    // inlining of macro ABC_FALLOC
    p.pNexts = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    // hash each node by its simulation info
    nTableSize = Abc_PrimeCudd( Gia_ManObjNum(p) );
    // inlining of macro ABC_FALLOC
    pTable = ((int *) memset(malloc(sizeof(int) * (size_t)(nTableSize)), 0xff, sizeof(int) * (size_t)(nTableSize)));
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        p.pReprs[i].iRepr = GIA_VOID;
        if Gia_ObjIsCo(pObj) {
            continue; }
        Key = Cec3_ManSimHashKey( Cec3_ObjSim(p, i), nWords, nTableSize );
        assert(Key >= 0 && Key < nTableSize, "Key >= 0 && Key < nTableSize");
        if pTable[Key] == -1 {
            pTable[Key] = i;
        } else {
            Gia_ObjSetRepr( p, i, pTable[Key] ); }
    }
    // create classes
    for  i = Gia_ManObjNum(p) - 1; i >= 0; i--  {
        var iRepr int32  = Gia_ObjRepr(p, i);
        if iRepr == GIA_VOID {
            continue; }
        Gia_ObjSetNext( p, i, Gia_ObjNext(p, iRepr) );
        Gia_ObjSetNext( p, iRepr, i );
    }
    pTable = nil;
    clk = Abc_Clock();
    // inlining of macro Gia_ManForEachClass0
    for  i = 0; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Cec3_ManSimClassRefineOne( p, i ); }
    }
    pMan.timeRefine += Abc_Clock() - clk;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_ManCreate(pAig *Gia_Man_t, pPars *Cec3_Par_t) *Cec3_Man_t {
    
    var pObj *Gia_Obj_t; var i int32;
    //satoko_opts_t Pars;
    //assert( Gia_ManRegNum(pAig) == 0 );
    // inlining of macro ABC_CALLOC
    p := &Cec3_Man_t{};
    memset( p, 0, sizeof(Cec3_Man_t) );
    p.timeStart    = Abc_Clock();
    p.pPars        = pPars;
    p.pAig         = pAig;
    // create new manager
    p.pNew         = Gia_ManStart( Gia_ManObjNum(pAig) );
    Gia_ManFillValue( pAig );
    Gia_ManConst0(pAig).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pAig.vCis)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
        pObj.Value = Gia_ManAppendCi( p.pNew );
    }
    Gia_ManHashAlloc( p.pNew );
    Vec_IntFill( &p.pNew.vCopies2, Gia_ManObjNum(p.pNew), -1 );
    // SAT solving
    //memset( &Pars, 0, sizeof(satoko_opts_t) );
    p.pSat         = bmcg_sat_solver_start();
    p.vFrontier    = Vec_PtrAlloc( 1000 );
    p.vFanins      = Vec_PtrAlloc( 100 );
    p.vNodesNew    = Vec_IntAlloc( 100 );
    p.vSatVars     = Vec_IntAlloc( 100 );
    p.vObjSatPairs = Vec_IntAlloc( 100 );
    p.vCexTriples  = Vec_IntAlloc( 100 );
    //Pars.conf_limit = pPars->nConfLimit;
    //satoko_configure(p->pSat, &Pars);
    // remember pointer to the solver in the AIG manager
    pAig.pData     = p.pSat;
    return p;
}
func Cec3_ManDestroy(p *Cec3_Man_t) {
    if p.pPars.fVerbose {
        var timeTotal abctime  = Abc_Clock() - p.timeStart;
        var timeSat abctime    = p.timeSatSat + p.timeSatUnsat + p.timeSatUndec;
        var timeOther abctime  = timeTotal - timeSat - p.timeSim - p.timeRefine - p.timeExtra;
//        Abc_Print( 1, "%d\n", p->Num );
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("SAT solving"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeSat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeSat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  sat      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  unsat    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUnsat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  fail     "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUndec)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUndec)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Simulation "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSim)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Refinement "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeRefine)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeRefine)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Rollback   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeExtra)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeExtra)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeOther)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeOther)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("TOTAL      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeTotal)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeTotal)))/(float64((timeTotal))) : 0.0);;
        fflush( os.Stdout );
    }

    Vec_WrdFreeP( &p.pAig.vSims );
    //Vec_WrdFreeP( &p->pAig->vSimsPi );
    Gia_ManCleanMark01( p.pAig );
    bmcg_sat_solver_stop( p.pSat );
    Gia_ManStopP( &p.pNew );
    Vec_PtrFreeP( &p.vFrontier );
    Vec_PtrFreeP( &p.vFanins );
    Vec_IntFreeP( &p.vNodesNew );
    Vec_IntFreeP( &p.vSatVars );
    Vec_IntFreeP( &p.vObjSatPairs );
    Vec_IntFreeP( &p.vCexTriples );
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Verify counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_ManVerify_rec(p *Gia_Man_t, iObj int32, pSat *bmcg_sat_solver) int32 {
    var Value0, Value1 int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if iObj == 0 { return 0; }
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return pObj.fMark1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if Gia_ObjIsCi(pObj) {
//        return pObj->fMark1 = satoko_var_polarity(pSat, Cec3_ObjSatId(p, pObj)) == SATOKO_LIT_TRUE;
        return pObj.fMark1 = bmcg_sat_solver_read_cex_varvalue(pSat, Cec3_ObjSatId(p, pObj)); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Cec3_ManVerify_rec( p, Gia_ObjFaninId0(pObj, iObj), pSat ) ^ Gia_ObjFaninC0(pObj);
    Value1 = Cec3_ManVerify_rec( p, Gia_ObjFaninId1(pObj, iObj), pSat ) ^ Gia_ObjFaninC1(pObj);
    return pObj.fMark1 = Value0 & Value1;
}
func Cec3_ManVerify(p *Gia_Man_t, iObj0 int32, iObj1 int32, fPhase int32, pSat *bmcg_sat_solver) {
//    int val0 = satoko_var_polarity(pSat, Cec3_ObjSatId(p, Gia_ManObj(p, iObj0))) == SATOKO_LIT_TRUE;
//    int val1 = satoko_var_polarity(pSat, Cec3_ObjSatId(p, Gia_ManObj(p, iObj1))) == SATOKO_LIT_TRUE;
    var Value0, Value1 int32;
    Gia_ManIncrementTravId( p );
    Value0 = Cec3_ManVerify_rec( p, iObj0, pSat );
    Value1 = Cec3_ManVerify_rec( p, iObj1, pSat );
    if (Value0 ^ Value1) == fPhase {
        fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", iObj0, iObj1); }
//    else
//        printf( "CEX verification succeeded for obj %d and obj %d.\n", iObj0, iObj1 );;
}

/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec3_ManCollect_rec(p *Cec3_Man_t, iObj int32) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p.pNew, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p.pNew, iObj);
    pObj = Gia_ManObj( p.pNew, iObj );
    if Cec3_ObjSatId(p.pNew, pObj) >= 0 {
        Vec_IntPush( p.vNodesNew, iObj );
        Vec_IntPush( p.vSatVars, Cec3_ObjSatId(p.pNew, pObj) );
    }
    if !(iObj != 0) {
        return; }
    if Gia_ObjIsAnd(pObj) {
        Cec3_ManCollect_rec( p, Gia_ObjFaninId0(pObj, iObj) );
        Cec3_ManCollect_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    
    } else {
        assert(Cec3_ObjSatId(p.pNew, pObj) >= 0, "Cec3_ObjSatId(p.pNew, pObj) >= 0");
        Vec_IntPushTwo( p.vObjSatPairs, Gia_ManCiIdToId(p.pAig, Gia_ObjCioId(pObj)), Cec3_ObjSatId(p.pNew, pObj) ); // SAT var
    }
}
func Cec3_ManSolveTwo(p *Cec3_Man_t, iObj0 int32, iObj1 int32, fPhase int32) int32 {
    var pObj *Gia_Obj_t;
    var status int32;
    var i int32;
    var iVar0 int32;
    var iVar1 int32;
    var Lits [2]int32;
    if iObj1 < iObj0 { 
        iObj1 ^= iObj0, iObj0 ^= iObj1, iObj1 ^= iObj0; }
    assert(iObj0 < iObj1, "iObj0 < iObj1");
    assert(p.pPars.fUseCones || bmcg_sat_solver_varnum(p.pSat) == 0, "p.pPars.fUseCones || bmcg_sat_solver_varnum(p.pSat) == 0");
    if !(iObj0 != 0) && Cec3_ObjSatId(p.pNew, Gia_ManConst0(p.pNew)) == -1 {
        Cec3_ObjSetSatId( p.pNew, Gia_ManConst0(p.pNew), bmcg_sat_solver_addvar(p.pSat) ); }
    iVar0 = Cec3_ObjGetCnfVar( p, iObj0 );
    iVar1 = Cec3_ObjGetCnfVar( p, iObj1 );
    // collect inputs and internal nodes
    Vec_IntClear( p.vNodesNew );
    Vec_IntClear( p.vSatVars );
    Vec_IntClear( p.vObjSatPairs );
    Gia_ManIncrementTravId( p.pNew );
    Cec3_ManCollect_rec( p, iObj0 );
    Cec3_ManCollect_rec( p, iObj1 );
//printf( "%d ", Vec_IntSize(p->vNodesNew) );
    // solve direct
    if p.pPars.fUseCones {  satoko_mark_cone( p.pSat, Vec_IntArray(p.vSatVars), Vec_IntSize(p.vSatVars) ); }
//    satoko_assump_push( p->pSat, Abc_Var2Lit(iVar0, 1) );
//    satoko_assump_push( p->pSat, Abc_Var2Lit(iVar1, fPhase) );
//    status = satoko_solve( p->pSat );
//    satoko_assump_pop( p->pSat );
//    satoko_assump_pop( p->pSat );
    Lits[0] = Abc_Var2Lit(iVar0, 1);
    Lits[1] = Abc_Var2Lit(iVar1, fPhase);
    bmcg_sat_solver_set_conflict_budget( p.pSat, p.pPars.nConfLimit );
    status = bmcg_sat_solver_solve( p.pSat, Lits, 2 );
    if status == GLUCOSE_UNSAT && iObj0 > 0 {
        // solve reverse
//        satoko_assump_push( p->pSat, Abc_Var2Lit(iVar0, 0) );
//        satoko_assump_push( p->pSat, Abc_Var2Lit(iVar1, !fPhase) );
//        status = satoko_solve( p->pSat );
//        satoko_assump_pop( p->pSat );
//        satoko_assump_pop( p->pSat );
        Lits[0] = Abc_Var2Lit(iVar0, 0);
        Lits[1] = Abc_Var2Lit(iVar1, !fPhase);
        bmcg_sat_solver_set_conflict_budget( p.pSat, p.pPars.nConfLimit );
        status = bmcg_sat_solver_solve( p.pSat, Lits, 2 );
    }
    if p.pPars.fUseCones {  satoko_unmark_cone( p.pSat, Vec_IntArray(p.vSatVars), Vec_IntSize(p.vSatVars) ); }
    //if ( status == SATOKO_SAT )
    //    Cec3_ManVerify( p->pNew, iObj0, iObj1, fPhase, p->pSat );
    if p.pPars.fUseCones {
        return status; }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vNodesNew)) && ((pObj) = Gia_ManObj(p.pNew, Vec_IntEntry(p.vNodesNew,i))); i++  {
        Cec3_ObjCleanSatId( p.pNew, pObj );
    }
    return status;
}

func Cec3_ManSweepNode(p *Cec3_Man_t, iObj int32) int32 {
    var clk abctime  = Abc_Clock();
    var i int32;
    var IdAig int32;
    var IdSat int32;
    var status int32;
    var RetValue int32 = 1;
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pAig, iObj );
    var pRepr *Gia_Obj_t  = Gia_ObjReprObj( p.pAig, iObj );
    var fCompl int32  = Abc_LitIsCompl(pObj.Value) ^ Abc_LitIsCompl(pRepr.Value) ^ pObj.fPhase ^ pRepr.fPhase;
    status = Cec3_ManSolveTwo( p, Abc_Lit2Var(pRepr.Value), Abc_Lit2Var(pObj.Value), fCompl );
    if status == GLUCOSE_SAT {
        p.nSatSat++;
        p.nPatterns++;
        if p.pAig.iPatsPi == 64 * p.pAig.nSimWords - 1 {
            p.pAig.iPatsPi = 1;
        } else {
            p.pAig.iPatsPi = p.pAig.iPatsPi + 1;
        }
        assert(p.pAig.iPatsPi > 0 && p.pAig.iPatsPi < 64 * p.pAig.nSimWords, "p.pAig.iPatsPi > 0 && p.pAig.iPatsPi < 64 * p.pAig.nSimWords");
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(p.vObjSatPairs)) && (((IdAig) = Vec_IntEntry(p.vObjSatPairs, i)), 1) && (((IdSat) = Vec_IntEntry(p.vObjSatPairs, i+1)), 1); i += 2  {
//            Cec3_ObjSimSetInputBit( p->pAig, IdAig, satoko_var_polarity(p->pSat, IdSat) == SATOKO_LIT_TRUE );
            Cec3_ObjSimSetInputBit( p.pAig, IdAig, bmcg_sat_solver_read_cex_varvalue(p.pSat, IdSat) );
        }
        p.timeSatSat += Abc_Clock() - clk;
        RetValue = 0;
    
    } else if status == GLUCOSE_UNSAT {
        p.nSatUnsat++;
        pObj.Value = Abc_LitNotCond( pRepr.Value, fCompl );
        Gia_ObjSetProved( p.pAig, iObj );
        p.timeSatUnsat += Abc_Clock() - clk;
        RetValue = 1;
    
    } else {
        p.nSatUndec++;
        assert(status == GLUCOSE_UNDEC, "status == GLUCOSE_UNDEC");
        Gia_ObjSetFailed( p.pAig, iObj );
        p.timeSatUndec += Abc_Clock() - clk;
        RetValue = 2;
    }
    if p.pPars.fUseCones {
        return RetValue; }
    clk = Abc_Clock();
    bmcg_sat_solver_reset( p.pSat );
    p.timeExtra += Abc_Clock() - clk;
//    satoko_stats(p->pSat)->n_conflicts = 0;
    return RetValue;
}
func Cec3_ManPrintStats(p *Gia_Man_t, pPars *Cec3_Par_t, pMan *Cec3_Man_t) {
    if !pPars.fVerbose {
        return; }
    fmt.Printf("S =%5d ", pMan ? pMan.nSatSat   : 0);
    fmt.Printf("U =%5d ", pMan ? pMan.nSatUnsat : 0);
    fmt.Printf("F =%5d ", pMan ? pMan.nSatUndec : 0);
    Gia_ManEquivPrintClasses( p, pPars.fVeryVerbose, 0 );
}
func Cec3_ManPerformSweeping(p *Gia_Man_t, pPars *Cec3_Par_t, ppNew **Gia_Man_t) int32 {
    var pMan *Cec3_Man_t  = Cec3_ManCreate( p, pPars ); 
    var pObj, pRepr, pObjNew *Gia_Obj_t; 
    var i int32;
    var Iter int32;
    var fDisproved int32 = 1;

    // check if any output trivially fails under all-0 pattern
    Gia_ManRandomW( 1 );
    Gia_ManSetPhase( p );
    if pPars.fIsMiter {
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if pObj.fPhase {
                p.pCexSeq = Cec3_ManDeriveCex( p, i, -1 );
                goto finalize;
            }
        }
    }

    // simulate one round and create classes
    Cec3_ManSimAlloc( p, pPars.nSimWords );
    Cec3_ManSimulateCis( p );
    Cec3_ManSimulate( p, nil, pMan );
    if pPars.fIsMiter && !(Cec3_ManSimulateCos(p) != 0) { // cex detected
        goto finalize; }
    Cec3_ManCreateClasses( p, pMan );
    Cec3_ManPrintStats( p, pPars, pMan );

    // perform additinal simulation
    for  i = 0; i < pPars.nSimRounds; i++  {
        Cec3_ManSimulateCis( p );
        Cec3_ManSimulate( p, nil, pMan );
        if pPars.fIsMiter && !(Cec3_ManSimulateCos(p) != 0) { // cex detected
            goto finalize; }
        Cec3_ManPrintStats( p, pPars, pMan );
    }
    // perform sweeping
    //pMan = Cec3_ManCreate( p, pPars );
    for  Iter = 0; (fDisproved != 0) && Iter < pPars.nItersMax; Iter++  {
        fDisproved = 0;
        pMan.nPatterns = 0;
        Cec3_ManSimulateCis( p );
        Vec_IntClear( pMan.vCexTriples );
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            if ~pObj.Value || Gia_ObjFailed(p, i) { // skip swept nodes and failed nodes
                continue; }
            if !~Gia_ObjFanin0(pObj).Value || !~Gia_ObjFanin1(pObj).Value { // skip fanouts of non-swept nodes
                continue; }
            assert(!Gia_ObjProved(p, i) && !Gia_ObjFailed(p, i), "!Gia_ObjProved(p, i) && !Gia_ObjFailed(p, i)");
            // duplicate the node
            pObj.Value = Gia_ManHashAnd( pMan.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            if Vec_IntSize(&pMan.pNew.vCopies2) == Abc_Lit2Var(pObj.Value) {
                pObjNew = Gia_ManObj( pMan.pNew, Abc_Lit2Var(pObj.Value) );
                pObjNew.fMark0 = Gia_ObjIsMuxType( pObjNew );
                Gia_ObjSetPhase( pMan.pNew, pObjNew );
                Vec_IntPush( &pMan.pNew.vCopies2, -1 );
            }
            assert(Vec_IntSize(&pMan.pNew.vCopies2) == Gia_ManObjNum(pMan.pNew), "Vec_IntSize(&pMan.pNew.vCopies2) == Gia_ManObjNum(pMan.pNew)");
            pRepr = Gia_ObjReprObj( p, i );
            if pRepr == nil || !~pRepr.Value {
                continue; }
            if Abc_Lit2Var(pObj.Value) == Abc_Lit2Var(pRepr.Value) {
                assert((pObj.Value ^ pRepr.Value) == (pObj.fPhase ^ pRepr.fPhase), "(pObj.Value ^ pRepr.Value) == (pObj.fPhase ^ pRepr.fPhase)");
                Gia_ObjSetProved( p, i );
                continue;
            }
            if Cec3_ManSweepNode(pMan, i) != 0 {
                if Gia_ObjProved(p, i) {
                    pObj.Value = Abc_LitNotCond( pRepr.Value, pObj.fPhase ^ pRepr.fPhase ); }
                continue;
            }
            pObj.Value = ~0;
            Vec_IntPushThree( pMan.vCexTriples, Gia_ObjId(p, pRepr), i, Abc_Var2Lit(p.iPatsPi, pObj.fPhase ^ pRepr.fPhase) );
            fDisproved = 1;
        }
        }
        if fDisproved != 0 {
            var Fails int32  = Cec3_ManSimulate( p, pMan.vCexTriples, pMan );
            if (Fails != 0) && pPars.fVerbose {
                fmt.Printf("Failed to resimulate %d times with pattern = %d  (total = %d).\n", Fails, pMan.nPatterns, pPars.nSimWords * 64); }
            if pPars.fIsMiter && !(Cec3_ManSimulateCos(p) != 0) { // cex detected
                break; }
        }
        Cec3_ManPrintStats( p, pPars, pMan );
    }
    // finish the AIG, if it is not finished
    if ppNew {
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
            if !~pObj.Value { 
                pObj.Value = Gia_ManHashAnd( pMan.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            pObj.Value = Gia_ManAppendCo( pMan.pNew, Gia_ObjFanin0Copy(pObj) );
        }
        *ppNew = Gia_ManCleanup( pMan.pNew );
        (*ppNew).pName = Abc_UtilStrsav( p.pName );
        (*ppNew).pSpec = Abc_UtilStrsav( p.pSpec );
    }
finalize:
    Cec3_ManDestroy( pMan );
    //Gia_ManEquivPrintClasses( p, 1, 0 );
    if p.pCexSeq {
        return 0;
    } else {
        return 1;
    }
}
func Cec3_ManSimulateTest(p *Gia_Man_t, pPars0 *Cec_ParFra_t) *Gia_Man_t {
    var pNew *Gia_Man_t  = nil;
    //abctime clk = Abc_Clock();
    var Pars Cec3_Par_t;
    var pPars *Cec3_Par_t = &Pars;
    Cec3_SetDefaultParams( pPars );
    // set resource limits
//    pPars->nSimWords  = pPars0->nWords;     // simulation words
//    pPars->nSimRounds = pPars0->nRounds;    // simulation rounds
//    pPars->nItersMax  = pPars0->nItersMax;  // max number of iterations
    pPars.nConfLimit = pPars0.nBTLimit;   // conflict limit at a node
    pPars.fUseCones  = pPars0.fUseCones;
    pPars.fVerbose   = pPars0.fVerbose;
//    Gia_ManComputeGiaEquivs( p, 100000, 0 );
//    Gia_ManEquivPrintClasses( p, 1, 0 );
    Cec3_ManPerformSweeping( p, pPars, &pNew );
    //Abc_PrintTime( 1, "SAT sweeping time", Abc_Clock() - clk );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

