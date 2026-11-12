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
//#include "sat/satoko/satoko.h"
//#include "cec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// sweeping manager
   
type Cec2_Par_t struct {
    nSimWords              int32;     // simulation words
    nSimRounds              int32;    // simulation rounds
    nItersMax              int32;     // max number of iterations
    nConfLimit              int32;    // SAT solver conflict limit
    fIsMiter              int32;      // this is a miter
    fUseCones              int32;     // use logic cones
    fVeryVerbose              int32;  // verbose stats
    fVerbose              int32;      // verbose stats
}

// SAT solving manager
   
type Cec2_Man_t struct {
    pPars      *Cec2_Par_t;          // parameters
    pAig       *Gia_Man_t;           // user's AIG
    pNew       *Gia_Man_t;           // internal AIG
    // SAT solving
    pSat        *satoko_t;           // SAT solver
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

func Cec2_ObjSatId(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjCopy2Array(p, Gia_ObjId(p, pObj));                                                     }
func Cec2_ObjSetSatId(p *Gia_Man_t, pObj *Gia_Obj_t, Num int32) int32 { assert(Cec2_ObjSatId(p, pObj) == -1, "Cec2_ObjSatId(p, pObj) == -1"); Gia_ObjSetCopy2Array(p, Gia_ObjId(p, pObj), Num); return Num;  }
func Cec2_ObjCleanSatId(p *Gia_Man_t, pObj *Gia_Obj_t) { assert(Cec2_ObjSatId(p, pObj) != -1, "Cec2_ObjSatId(p, pObj) != -1"); Gia_ObjSetCopy2Array(p, Gia_ObjId(p, pObj), -1);               }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets parameter defaults.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_SetDefaultParams(p *Cec2_Par_t) {
    memset( p, 0, sizeof(Cec2_Par_t) );
    p.nSimWords      =      12;    // simulation words
    p.nSimRounds     =       4;    // simulation rounds
    p.nItersMax      =      10;    // max number of iterations
    p.nConfLimit     =    1000;    // conflict limit at a node
    p.fIsMiter       =       0;    // this is a miter
    p.fUseCones      =       1;    // use logic cones
    p.fVeryVerbose   =       0;    // verbose stats
    p.fVerbose       =       0;    // verbose stats
}  

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_AddClausesMux(p *Gia_Man_t, pNode *Gia_Obj_t, pSat *satoko_t) {
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
    VarF = Cec2_ObjSatId(p, pNode);
    VarI = Cec2_ObjSatId(p, pNodeI);
    VarT = Cec2_ObjSatId(p, Gia_Regular(pNodeT));
    VarE = Cec2_ObjSatId(p, Gia_Regular(pNodeE));
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
func Cec2_AddClausesSuper(p *Gia_Man_t, pNode *Gia_Obj_t, vSuper *Vec_Ptr_t, pSat *satoko_t) {
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
        pLits[0] = Abc_Var2Lit(Cec2_ObjSatId(p, Gia_Regular(pFanin)), Gia_IsComplement(pFanin));
        pLits[1] = Abc_Var2Lit(Cec2_ObjSatId(p, pNode), 1);
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
            if pNode.fPhase {                pLits[1] = Abc_LitNot( pLits[1] ); }
        }
        RetValue = satoko_add_clause( pSat, pLits, 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[i] = Abc_Var2Lit(Cec2_ObjSatId(p, Gia_Regular(pFanin)), !Gia_IsComplement(pFanin));
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[i] = Abc_LitNot( pLits[i] ); }
        }
    }
    pLits[nLits-1] = Abc_Var2Lit(Cec2_ObjSatId(p, pNode), 0);
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
func Cec2_CollectSuper_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Gia_IsComplement(pObj) || Gia_ObjIsCi(pObj) || 
         (!fFirst && (p.pRefs ? Gia_ObjRefNum(p, pObj) : Gia_ObjValue(pObj)) > 1) || 
         (fUseMuxes && pObj.fMark0) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
    // go through the branches
    Cec2_CollectSuper_rec( p, Gia_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Cec2_CollectSuper_rec( p, Gia_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}
func Cec2_CollectSuper(p *Gia_Man_t, pObj *Gia_Obj_t, fUseMuxes int32, vSuper *Vec_Ptr_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsCi(pObj), "!Gia_ObjIsCi(pObj)");
    Vec_PtrClear( vSuper );
    Cec2_CollectSuper_rec( p, pObj, vSuper, 1, fUseMuxes );
}
func Cec2_ObjAddToFrontier(p *Gia_Man_t, pObj *Gia_Obj_t, vFrontier *Vec_Ptr_t, pSat *satoko_t) {
    var iVar int32;
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsConst0(pObj), "!Gia_ObjIsConst0(pObj)");
    if Cec2_ObjSatId(p, pObj) >= 0 {
        return; }
    assert(Cec2_ObjSatId(p, pObj) == -1, "Cec2_ObjSatId(p, pObj) == -1");
    iVar = satoko_add_variable(pSat, 0);
    if p.vVar2Obj {
        assert(Vec_IntSize(p.vVar2Obj) == iVar, "Vec_IntSize(p.vVar2Obj) == iVar");
        Vec_IntPush( p.vVar2Obj, Gia_ObjId(p, pObj) );
    }
    Cec2_ObjSetSatId( p, pObj, iVar );
    if Gia_ObjIsAnd(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}
func Gia_ObjGetCnfVar(pGia *Gia_Man_t, iObj int32, vFrontier *Vec_Ptr_t, vFanins *Vec_Ptr_t, pSat *satoko_t) int32 { 
    var pNode, pFanin *Gia_Obj_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(pGia, iObj);
    var i int32;
    var k int32;
    var fUseMuxes int32 = 1;
    if Vec_IntSize(&pGia.vCopies2) < Gia_ManObjNum(pGia) {
        Vec_IntFillExtra( &pGia.vCopies2, Gia_ManObjNum(pGia), -1 ); }
    // quit if CNF is ready
    if Cec2_ObjSatId(pGia,pObj) >= 0 {
        return Cec2_ObjSatId(pGia,pObj); }
    assert(iObj > 0, "iObj > 0");
    if Gia_ObjIsCi(pObj) {
        var iVar int32  = satoko_add_variable(pSat, 0);
        if pGia.vVar2Obj {
            assert(Vec_IntSize(pGia.vVar2Obj) == iVar, "Vec_IntSize(pGia.vVar2Obj) == iVar");
            Vec_IntPush( pGia.vVar2Obj, iObj );
        }
        return Cec2_ObjSetSatId( pGia, pObj, iVar );
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // start the frontier
    Vec_PtrClear( vFrontier );
    Cec2_ObjAddToFrontier( pGia, pObj, vFrontier, pSat );
    // explore nodes in the frontier
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFrontier)) && (((pNode) = (Gia_Obj_t *)Vec_PtrEntry(vFrontier, i)), 1); i++  {
        // create the supergate
        assert(Cec2_ObjSatId(pGia,pNode) >= 0, "Cec2_ObjSatId(pGia,pNode) >= 0");
        if (fUseMuxes != 0) && pNode.fMark0 {
            Vec_PtrClear( vFanins );
            Vec_PtrPushUnique( vFanins, Gia_ObjFanin0( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( vFanins, Gia_ObjFanin0( Gia_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( vFanins, Gia_ObjFanin1( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( vFanins, Gia_ObjFanin1( Gia_ObjFanin1(pNode) ) );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFanins)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(vFanins, k)), 1); k++  {
                Cec2_ObjAddToFrontier( pGia, Gia_Regular(pFanin), vFrontier, pSat );
            }
            Cec2_AddClausesMux( pGia, pNode, pSat );
        
        } else {
            Cec2_CollectSuper( pGia, pNode, fUseMuxes, vFanins );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFanins)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(vFanins, k)), 1); k++  {
                Cec2_ObjAddToFrontier( pGia, Gia_Regular(pFanin), vFrontier, pSat );
            }
            Cec2_AddClausesSuper( pGia, pNode, vFanins, pSat ); 
            //printf( "%d ", Vec_PtrSize(vFanins) );
        }
        assert(Vec_PtrSize(vFanins) > 1, "Vec_PtrSize(vFanins) > 1");
    }
    return Cec2_ObjSatId(pGia,pObj);
}
func Cec2_ObjGetCnfVar(p *Cec2_Man_t, iObj int32) int32 { 
    return Gia_ObjGetCnfVar( p.pNew, iObj, p.vFrontier, p.vFanins, p.pSat );
}


/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ObjSim(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, p.nSimWords * iObj );
}
func Cec2_ObjSimSetInputBit(p *Gia_Man_t, iObj int32, Bit int32) {
    var pSim *word  = Cec2_ObjSim( p, iObj );
    if Abc_InfoHasBit( (unsigned*)pSim, p.iPatsPi ) != Bit {
        Abc_InfoXorBit( (unsigned*)pSim, p.iPatsPi ); }
}
func Cec2_ObjSimRo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSimRo *word  = Cec2_ObjSim( p, iObj );
    var pSimRi *word  = Cec2_ObjSim( p, Gia_ObjRoToRiId(p, iObj) );
    for  w = 0; w < p.nSimWords; w++  {
        pSimRo[w] = pSimRi[w];
    }
}
func Cec2_ObjSimCo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSimCo *word   = Cec2_ObjSim( p, iObj );
    var pSimDri *word  = Cec2_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] = ~pSimDri[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] =  pSimDri[w];
        } }
}
func Cec2_ObjSimAnd(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSim *word   = Cec2_ObjSim( p, iObj );
    var pSim0 *word  = Cec2_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSim1 *word  = Cec2_ObjSim( p, Gia_ObjFaninId1(pObj, iObj) );
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
func Cec2_ObjSimEqual(p *Gia_Man_t, iObj0 int32, iObj1 int32) int32 {
    var w int32;
    var pSim0 *word  = Cec2_ObjSim( p, iObj0 );
    var pSim1 *word  = Cec2_ObjSim( p, iObj1 );
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
func Cec2_ObjSimCi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSim *word  = Cec2_ObjSim( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSim[w] = Gia_ManRandomW( 0 );
    }
    pSim[0] <<= 1;
}
func Cec2_ManSimulateCis(p *Gia_Man_t) {
    var i, Id int32;
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Cec2_ObjSimCi( p, Id );
    }
    p.iPatsPi = 0;
}
func Cec2_ManDeriveCex(p *Gia_Man_t, iOut int32, iPat int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i, Id int32;
    pCex = Abc_CexAlloc( 0, Gia_ManCiNum(p), 1 );
    pCex.iPo = iOut;
    if iPat == -1 {
        return pCex; }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        if Abc_InfoHasBit((unsigned *)Cec2_ObjSim(p, Id), iPat) {
            Abc_InfoSetBit( pCex.pData, i ); }
    }
    return pCex;
}
func Cec2_ManSimulateCos(p *Gia_Man_t) int32 {
    var i, Id int32;
    // check outputs and generate CEX if they fail
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        Cec2_ObjSimCo( p, Id );
        if Cec2_ObjSimEqual(p, Id, 0) != 0 {
            continue; }
        p.pCexSeq = Cec2_ManDeriveCex( p, i, Abc_TtFindFirstBit2(Cec2_ObjSim(p, Id), p.nSimWords) );
        return 0;
    }
    return 1;
}
func Cec2_ManSaveCis(p *Gia_Man_t) {
    var w, i, Id int32;
    assert(p.vSimsPi != nil, "p.vSimsPi != nil");
    for  w = 0; w < p.nSimWords; w++  {
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
            Vec_WrdPush( p.vSimsPi, Cec2_ObjSim(p, Id)[w] );
        }
    }
}
func Cec2_ManSimulate(p *Gia_Man_t, vTriples *Vec_Int_t, pMan *Cec2_Man_t) int32 {
    extern void Cec2_ManSimClassRefineOne( Gia_Man_t * p, int iRepr );
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; 
    var i, iRepr, iObj, Entry, Count int32;
    //Cec2_ManSaveCis( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Cec2_ObjSimAnd( p, i ); }
    }
    pMan.timeSim += Abc_Clock() - clk;
    if p.pReprs == nil {
        return 0; }
    if vTriples != nil {
        // inlining of macro Vec_IntForEachEntryTriple
        for  i = 0; (i+2 < Vec_IntSize(vTriples)) && (((iRepr) = Vec_IntEntry(vTriples, i)), 1) && (((iObj) = Vec_IntEntry(vTriples, i+1)), 1) && (((Entry) = Vec_IntEntry(vTriples, i+2)), 1); i += 3  {
            var pSim0 *word  = Cec2_ObjSim( p, iRepr );
            var pSim1 *word  = Cec2_ObjSim( p, iObj );
            var iPat int32      = Abc_Lit2Var(Entry);
            var fPhase int32    = Abc_LitIsCompl(Entry);
            if (fPhase ^ Abc_InfoHasBit((unsigned *)pSim0, iPat)) == Abc_InfoHasBit((unsigned *)pSim1, iPat) {
                Count++; }
        }
    }
    clk = Abc_Clock();
    // inlining of macro Gia_ManForEachClass0
    for  i = 0; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Cec2_ManSimClassRefineOne( p, i ); }
    }
    pMan.timeRefine += Abc_Clock() - clk;
    return Count;
}
func Cec2_ManSimAlloc(p *Gia_Man_t, nWords int32) {
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
var spStatic_Cec2_ManSimHashKey_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };
func Cec2_ManSimHashKey(pSim *word, nSims int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var pSimU *uint32 = (unsigned *)pSim;
    var i int32;
    var nSimsU int32 = 2 * nSims;
    if pSimU[0] & 1 {
        for  i = 0; i < nSimsU; i++  {
            uHash ^= ~pSimU[i] * spStatic_Cec2_ManSimHashKey_s_Primes[i & 0xf];
        }
    } else {
        for  i = 0; i < nSimsU; i++  {
            uHash ^= pSimU[i] * spStatic_Cec2_ManSimHashKey_s_Primes[i & 0xf];
        } }
    return int32((uHash % nTableSize));

}

/**Function*************************************************************

  Synopsis    [Creating initial equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ManSimClassRefineOne(p *Gia_Man_t, iRepr int32) {
    var iObj int32;
    var iPrev int32 = iRepr;
    var iPrev2 int32;
    var iRepr2 int32;
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iRepr2 = Gia_ObjNext(p, iRepr); iRepr2 > 0; iRepr2 = Gia_ObjNext(p, iRepr2)  {
        if Cec2_ObjSimEqual(p, iRepr, iRepr2) != 0 {
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
        if Cec2_ObjSimEqual(p, iRepr, iObj) != 0 { // remains with iRepr
        
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
func Cec2_ManCreateClasses(p *Gia_Man_t, pMan *Cec2_Man_t) {
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
        Key = Cec2_ManSimHashKey( Cec2_ObjSim(p, i), nWords, nTableSize );
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
        Cec2_ManSimClassRefineOne( p, i ); }
    }
    pMan.timeRefine += Abc_Clock() - clk;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec2_ManCreate(pAig *Gia_Man_t, pPars *Cec2_Par_t) *Cec2_Man_t {
    
    var pObj *Gia_Obj_t; var i int32;
    var Pars satoko_opts_t;
    //assert( Gia_ManRegNum(pAig) == 0 );
    // inlining of macro ABC_CALLOC
    p := &Cec2_Man_t{};
    memset( p, 0, sizeof(Cec2_Man_t) );
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
    memset( &Pars, 0, sizeof(satoko_opts_t) );
    p.pSat         = satoko_create();
    p.vFrontier    = Vec_PtrAlloc( 1000 );
    p.vFanins      = Vec_PtrAlloc( 100 );
    p.vNodesNew    = Vec_IntAlloc( 100 );
    p.vSatVars     = Vec_IntAlloc( 100 );
    p.vObjSatPairs = Vec_IntAlloc( 100 );
    p.vCexTriples  = Vec_IntAlloc( 100 );
    Pars.conf_limit = pPars.nConfLimit;
    satoko_configure(p.pSat, &Pars);
    // remember pointer to the solver in the AIG manager
    pAig.pData     = p.pSat;
    return p;
}
func Cec2_ManDestroy(p *Cec2_Man_t) {
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
    satoko_destroy( p.pSat );
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
func Cec2_ManVerify_rec(p *Gia_Man_t, iObj int32, pSat *satoko_t) int32 {
    var Value0, Value1 int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if iObj == 0 { return 0; }
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return pObj.fMark1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if Gia_ObjIsCi(pObj) {
        return pObj.fMark1 = satoko_var_polarity(pSat, Cec2_ObjSatId(p, pObj)) == SATOKO_LIT_TRUE; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Cec2_ManVerify_rec( p, Gia_ObjFaninId0(pObj, iObj), pSat ) ^ Gia_ObjFaninC0(pObj);
    Value1 = Cec2_ManVerify_rec( p, Gia_ObjFaninId1(pObj, iObj), pSat ) ^ Gia_ObjFaninC1(pObj);
    return pObj.fMark1 = Value0 & Value1;
}
func Cec2_ManVerify(p *Gia_Man_t, iObj0 int32, iObj1 int32, fPhase int32, pSat *satoko_t) {
    var Value0, Value1 int32;
    Gia_ManIncrementTravId( p );
    Value0 = Cec2_ManVerify_rec( p, iObj0, pSat );
    Value1 = Cec2_ManVerify_rec( p, iObj1, pSat );
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
func Cec2_ManCollect_rec(p *Cec2_Man_t, iObj int32) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p.pNew, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p.pNew, iObj);
    pObj = Gia_ManObj( p.pNew, iObj );
    if Cec2_ObjSatId(p.pNew, pObj) >= 0 {
        Vec_IntPush( p.vNodesNew, iObj );
        Vec_IntPush( p.vSatVars, Cec2_ObjSatId(p.pNew, pObj) );
    }
    if !(iObj != 0) {
        return; }
    if Gia_ObjIsAnd(pObj) {
        Cec2_ManCollect_rec( p, Gia_ObjFaninId0(pObj, iObj) );
        Cec2_ManCollect_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    
    } else {
        assert(Cec2_ObjSatId(p.pNew, pObj) >= 0, "Cec2_ObjSatId(p.pNew, pObj) >= 0");
        Vec_IntPushTwo( p.vObjSatPairs, Gia_ManCiIdToId(p.pAig, Gia_ObjCioId(pObj)), Cec2_ObjSatId(p.pNew, pObj) ); // SAT var
    }
}
func Cec2_ManSolveTwo(p *Cec2_Man_t, iObj0 int32, iObj1 int32, fPhase int32) int32 {
    var pObj *Gia_Obj_t;
    var status, i, iVar0, iVar1 int32;
    if iObj1 < iObj0 { 
        iObj1 ^= iObj0, iObj0 ^= iObj1, iObj1 ^= iObj0; }
    assert(iObj0 < iObj1, "iObj0 < iObj1");
    assert(p.pPars.fUseCones || satoko_varnum(p.pSat) == 0, "p.pPars.fUseCones || satoko_varnum(p.pSat) == 0");
    if !(iObj0 != 0) && Cec2_ObjSatId(p.pNew, Gia_ManConst0(p.pNew)) == -1 {
        Cec2_ObjSetSatId( p.pNew, Gia_ManConst0(p.pNew), satoko_add_variable(p.pSat, 0) ); }
    iVar0 = Cec2_ObjGetCnfVar( p, iObj0 );
    iVar1 = Cec2_ObjGetCnfVar( p, iObj1 );
    // collect inputs and internal nodes
    Vec_IntClear( p.vNodesNew );
    Vec_IntClear( p.vSatVars );
    Vec_IntClear( p.vObjSatPairs );
    Gia_ManIncrementTravId( p.pNew );
    Cec2_ManCollect_rec( p, iObj0 );
    Cec2_ManCollect_rec( p, iObj1 );
    // solve direct
    if p.pPars.fUseCones {  satoko_mark_cone( p.pSat, Vec_IntArray(p.vSatVars), Vec_IntSize(p.vSatVars) ); }
    satoko_assump_push( p.pSat, Abc_Var2Lit(iVar0, 1) );
    satoko_assump_push( p.pSat, Abc_Var2Lit(iVar1, fPhase) );
    status = satoko_solve( p.pSat );
    satoko_assump_pop( p.pSat );
    satoko_assump_pop( p.pSat );
    if status == SATOKO_UNSAT && iObj0 > 0 {
        // solve reverse
        satoko_assump_push( p.pSat, Abc_Var2Lit(iVar0, 0) );
        satoko_assump_push( p.pSat, Abc_Var2Lit(iVar1, !fPhase) );
        status = satoko_solve( p.pSat );
        satoko_assump_pop( p.pSat );
        satoko_assump_pop( p.pSat );
    }
    if p.pPars.fUseCones {  satoko_unmark_cone( p.pSat, Vec_IntArray(p.vSatVars), Vec_IntSize(p.vSatVars) ); }
    //if ( status == SATOKO_SAT )
    //    Cec2_ManVerify( p->pNew, iObj0, iObj1, fPhase, p->pSat );
    if p.pPars.fUseCones {
        return status; }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vNodesNew)) && ((pObj) = Gia_ManObj(p.pNew, Vec_IntEntry(p.vNodesNew,i))); i++  {
        Cec2_ObjCleanSatId( p.pNew, pObj );
    }
    return status;
}

func Cec2_ManSweepNode(p *Cec2_Man_t, iObj int32) int32 {
    var clk abctime  = Abc_Clock();
    var i int32;
    var IdAig int32;
    var IdSat int32;
    var status int32;
    var RetValue int32 = 1;
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pAig, iObj );
    var pRepr *Gia_Obj_t  = Gia_ObjReprObj( p.pAig, iObj );
    var fCompl int32  = Abc_LitIsCompl(pObj.Value) ^ Abc_LitIsCompl(pRepr.Value) ^ pObj.fPhase ^ pRepr.fPhase;
    status = Cec2_ManSolveTwo( p, Abc_Lit2Var(pRepr.Value), Abc_Lit2Var(pObj.Value), fCompl );
    if status == SATOKO_SAT {
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
            Cec2_ObjSimSetInputBit( p.pAig, IdAig, satoko_var_polarity(p.pSat, IdSat) == SATOKO_LIT_TRUE );
        }
        p.timeSatSat += Abc_Clock() - clk;
        RetValue = 0;
    
    } else if status == SATOKO_UNSAT {
        p.nSatUnsat++;
        pObj.Value = Abc_LitNotCond( pRepr.Value, fCompl );
        Gia_ObjSetProved( p.pAig, iObj );
        p.timeSatUnsat += Abc_Clock() - clk;
        RetValue = 1;
    
    } else {
        p.nSatUndec++;
        assert(status == SATOKO_UNDEC, "status == SATOKO_UNDEC");
        Gia_ObjSetFailed( p.pAig, iObj );
        p.timeSatUndec += Abc_Clock() - clk;
        RetValue = 2;
    }
    if p.pPars.fUseCones {
        return RetValue; }
    clk = Abc_Clock();
    satoko_rollback( p.pSat );
    p.timeExtra += Abc_Clock() - clk;
    satoko_stats(p.pSat).n_conflicts = 0;
    return RetValue;
}
func Cec2_ManPrintStats(p *Gia_Man_t, pPars *Cec2_Par_t, pMan *Cec2_Man_t) {
    if !pPars.fVerbose {
        return; }
    fmt.Printf("S =%5d ", pMan ? pMan.nSatSat   : 0);
    fmt.Printf("U =%5d ", pMan ? pMan.nSatUnsat : 0);
    fmt.Printf("F =%5d ", pMan ? pMan.nSatUndec : 0);
    Gia_ManEquivPrintClasses( p, pPars.fVeryVerbose, 0 );
}
func Cec2_ManPerformSweeping(p *Gia_Man_t, pPars *Cec2_Par_t, ppNew **Gia_Man_t) int32 {
    var pMan *Cec2_Man_t  = Cec2_ManCreate( p, pPars ); 
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
                p.pCexSeq = Cec2_ManDeriveCex( p, i, -1 );
                goto finalize;
            }
        }
    }

    // simulate one round and create classes
    Cec2_ManSimAlloc( p, pPars.nSimWords );
    Cec2_ManSimulateCis( p );
    Cec2_ManSimulate( p, nil, pMan );
    if pPars.fIsMiter && !(Cec2_ManSimulateCos(p) != 0) { // cex detected
        goto finalize; }
    Cec2_ManCreateClasses( p, pMan );
    Cec2_ManPrintStats( p, pPars, pMan );

    // perform additinal simulation
    for  i = 0; i < pPars.nSimRounds; i++  {
        Cec2_ManSimulateCis( p );
        Cec2_ManSimulate( p, nil, pMan );
        if pPars.fIsMiter && !(Cec2_ManSimulateCos(p) != 0) { // cex detected
            goto finalize; }
        Cec2_ManPrintStats( p, pPars, pMan );
    }
    // perform sweeping
    //pMan = Cec2_ManCreate( p, pPars );
    for  Iter = 0; (fDisproved != 0) && Iter < pPars.nItersMax; Iter++  {
        fDisproved = 0;
        pMan.nPatterns = 0;
        Cec2_ManSimulateCis( p );
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
            if Cec2_ManSweepNode(pMan, i) != 0 {
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
            var Fails int32  = Cec2_ManSimulate( p, pMan.vCexTriples, pMan );
            if (Fails != 0) && pPars.fVerbose {
                fmt.Printf("Failed to resimulate %d times with pattern = %d  (total = %d).\n", Fails, pMan.nPatterns, pPars.nSimWords * 64); }
            if pPars.fIsMiter && !(Cec2_ManSimulateCos(p) != 0) { // cex detected
                break; }
        }
        Cec2_ManPrintStats( p, pPars, pMan );
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
    Cec2_ManDestroy( pMan );
    //Gia_ManEquivPrintClasses( p, 1, 0 );
    if p.pCexSeq {
        return 0;
    } else {
        return 1;
    }
}
func Cec2_ManSimulateTest(p *Gia_Man_t, pPars0 *Cec_ParFra_t) *Gia_Man_t {
    var pNew *Gia_Man_t  = nil;
    //abctime clk = Abc_Clock();
    var Pars Cec2_Par_t;
    var pPars *Cec2_Par_t = &Pars;
    Cec2_SetDefaultParams( pPars );
    // set resource limits
//    pPars->nSimWords  = pPars0->nWords;     // simulation words
//    pPars->nSimRounds = pPars0->nRounds;    // simulation rounds
//    pPars->nItersMax  = pPars0->nItersMax;  // max number of iterations
    pPars.nConfLimit = pPars0.nBTLimit;   // conflict limit at a node
    pPars.fUseCones  = pPars0.fUseCones;
    pPars.fVerbose   = pPars0.fVerbose;
//    Gia_ManComputeGiaEquivs( p, 100000, 0 );
//    Gia_ManEquivPrintClasses( p, 1, 0 );
    Cec2_ManPerformSweeping( p, pPars, &pNew );
    //Abc_PrintTime( 1, "SAT sweeping time", Abc_Clock() - clk );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

