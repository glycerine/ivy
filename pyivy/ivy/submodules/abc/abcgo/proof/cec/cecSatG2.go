package cec

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [cecSatG2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Detection of structural isomorphism.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSatG2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "misc/util/utilTruth.h"
//#include "cec.h"

#define USE_GLUCOSE2

                   

//#include "sat/glucose2/AbcGlucose2.h"

#define sat_solver                         bmcg2_sat_solver
#define sat_solver_start                   bmcg2_sat_solver_start
#define sat_solver_stop                    bmcg2_sat_solver_stop
#define sat_solver_addclause               bmcg2_sat_solver_addclause
#define sat_solver_add_and                 bmcg2_sat_solver_add_and
#define sat_solver_add_xor                 bmcg2_sat_solver_add_xor
#define sat_solver_addvar                  bmcg2_sat_solver_addvar
#define sat_solver_read_cex_varvalue       bmcg2_sat_solver_read_cex_varvalue
#define sat_solver_reset                   bmcg2_sat_solver_reset
#define sat_solver_set_conflict_budget     bmcg2_sat_solver_set_conflict_budget
#define sat_solver_conflictnum             bmcg2_sat_solver_conflictnum
#define sat_solver_solve                   bmcg2_sat_solver_solve
#define sat_solver_read_cex_varvalue       bmcg2_sat_solver_read_cex_varvalue
#define sat_solver_read_cex                bmcg2_sat_solver_read_cex
#define sat_solver_jftr                    bmcg2_sat_solver_jftr
#define sat_solver_set_jftr                bmcg2_sat_solver_set_jftr
#define sat_solver_set_var_fanin_lit       bmcg2_sat_solver_set_var_fanin_lit
#define sat_solver_start_new_round         bmcg2_sat_solver_start_new_round
#define sat_solver_mark_cone               bmcg2_sat_solver_mark_cone

     

                                   

                                                          
                                                                
                                                               
                                                                    
                                                                  
                                                                  
                                                                 
                                                                            
                                                                
                                                                              
                                                                      
                                                                
                                                                            
                                                                   
                                                               
                                                                   
                                                                            
                                                                          
                                                                    

      

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// SAT solving manager
   
type Cec4_Man_t struct {
    pPars    *Cec_ParFra_t;          // parameters
    pAig       *Gia_Man_t;           // user's AIG
    pNew       *Gia_Man_t;           // internal AIG
    // SAT solving
    pSat      *sat_solver;           // SAT solver
    vFrontier       *Vec_Ptr_t;      // CNF construction
    vFanins       *Vec_Ptr_t;        // CNF construction
    vCexMin       *Vec_Int_t;        // minimized CEX
    vClassUpdates       *Vec_Int_t;  // updated equiv classes
    vCexStamps       *Vec_Int_t;     // time stamps
    vCands       *Vec_Int_t;
    vVisit       *Vec_Int_t;
    vPat       *Vec_Int_t;
    vDisprPairs       *Vec_Int_t;
    vFails       *Vec_Bit_t;
    iPosRead              int32;       // candidate reading position
    iPosWrite              int32;      // candidate writing position
    iLastConst              int32;     // last const node proved
    // refinement
    vRefClasses       *Vec_Int_t;
    vRefNodes       *Vec_Int_t;
    vRefBins       *Vec_Int_t;
    pTable             *int32;
    nTableSize              int32;
    // statistics
    nItersSim              int32;
    nItersSat              int32;
    nAndNodes              int32;
    nPatterns              int32;
    nSatSat              int32;
    nSatUnsat              int32;
    nSatUndec              int32;
    nCallsSince              int32;
    nSimulates              int32;
    nRecycles              int32;
    nConflicts [2][3]int32;
    nGates [2]int32;
    nFaster [2]int32;
    timeCnf          abctime;
    timeGenPats          abctime;
    timeSatSat0          abctime;
    timeSatUnsat0          abctime;
    timeSatSat          abctime;
    timeSatUnsat          abctime;
    timeSatUndec          abctime;
    timeSim          abctime;
    timeRefine          abctime;
    timeResimGlo          abctime;
    timeResimLoc          abctime;
    timeStart          abctime;
}

func Cec4_ObjSatId(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Gia_ObjCopy2Array(p, Gia_ObjId(p, pObj));                                                     }
func Cec4_ObjSetSatId(p *Gia_Man_t, pObj *Gia_Obj_t, Num int32) int32 { assert(Cec4_ObjSatId(p, pObj) == -1, "Cec4_ObjSatId(p, pObj) == -1"); Gia_ObjSetCopy2Array(p, Gia_ObjId(p, pObj), Num); Vec_IntPush(&p.vSuppVars, Gia_ObjId(p, pObj)); if Gia_ObjIsCi(pObj) { Vec_IntPushTwo(&p.vCopiesTwo, Gia_ObjId(p, pObj), Num); }
 assert(Vec_IntSize(&p.vVarMap) == Num, "Vec_IntSize(&p.vVarMap) == Num"); Vec_IntPush(&p.vVarMap, Gia_ObjId(p, pObj)); return Num;  }
func Cec4_ObjCleanSatId(p *Gia_Man_t, pObj *Gia_Obj_t) { assert(Cec4_ObjSatId(p, pObj) != -1, "Cec4_ObjSatId(p, pObj) != -1"); Gia_ObjSetCopy2Array(p, Gia_ObjId(p, pObj), -1);               }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Default parameter settings.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_ManCreate(pAig *Gia_Man_t, pPars *Cec_ParFra_t) *Cec4_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Cec4_Man_t  = &Cec4_Man_t{};
    memset( p, 0, sizeof(Cec4_Man_t) );
    p.timeStart     = Abc_Clock();
    p.pPars         = pPars;
    p.pAig          = pAig;
    p.pSat          = sat_solver_start();  
    sat_solver_set_jftr( p.pSat, pPars.jType );
    p.vFrontier     = Vec_PtrAlloc( 1000 );
    p.vFanins       = Vec_PtrAlloc( 100 );
    p.vCexMin       = Vec_IntAlloc( 100 );
    p.vClassUpdates = Vec_IntAlloc( 100 );
    p.vCexStamps    = Vec_IntStart( Gia_ManObjNum(pAig) );
    p.vCands        = Vec_IntAlloc( 100 );
    p.vVisit        = Vec_IntAlloc( 100 );
    p.vPat          = Vec_IntAlloc( 100 );
    p.vDisprPairs   = Vec_IntAlloc( 100 );
    p.vFails        = Vec_BitStart( Gia_ManObjNum(pAig) );
    //pAig->pData     = p->pSat; // point AIG manager to the solver
    return p;
}
func Cec4_ManDestroy(p *Cec4_Man_t) {
    if p.pPars.fVerbose {
        var timeTotal abctime  = Abc_Clock() - p.timeStart;
        var timeSat abctime    = p.timeSatSat0 + p.timeSatSat + p.timeSatUnsat0 + p.timeSatUnsat + p.timeSatUndec;
        var timeOther abctime  = timeTotal - timeSat - p.timeSim - p.timeRefine - p.timeResimLoc - p.timeGenPats;// - p->timeResimGlo;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("SAT solving  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeSat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeSat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  sat(easy)  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat0)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatSat0)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  sat        "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  unsat(easy)"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat0)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUnsat0)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  unsat      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUnsat)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("  fail       "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUndec)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSatUndec)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Generate CNF "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCnf)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeCnf)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Generate pats"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeGenPats)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeGenPats)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Simulation   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeSim)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Refinement   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeRefine)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeRefine)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Resim global "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeResimGlo)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeResimGlo)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Resim local  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeResimLoc)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((p.timeResimLoc)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other        "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeOther)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeOther)))/(float64((timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("TOTAL        "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((timeTotal)))/(float64(int64(time.Second))), (float64((timeTotal)))? 100.0*(float64((timeTotal)))/(float64((timeTotal))) : 0.0);;
        fflush( os.Stdout );
    }
    Vec_WrdFreeP( &p.pAig.vSims );
    Vec_WrdFreeP( &p.pAig.vSimsPi );
    Gia_ManCleanMark01( p.pAig );
    sat_solver_stop( p.pSat );
    Gia_ManStopP( &p.pNew );
    Vec_PtrFreeP( &p.vFrontier );
    Vec_PtrFreeP( &p.vFanins );
    Vec_IntFreeP( &p.vCexMin );
    Vec_IntFreeP( &p.vClassUpdates );
    Vec_IntFreeP( &p.vCexStamps );
    Vec_IntFreeP( &p.vCands );
    Vec_IntFreeP( &p.vVisit );
    Vec_IntFreeP( &p.vPat );
    Vec_IntFreeP( &p.vDisprPairs );
    Vec_BitFreeP( &p.vFails );
    Vec_IntFreeP( &p.vRefClasses );
    Vec_IntFreeP( &p.vRefNodes );
    Vec_IntFreeP( &p.vRefBins );
    p.pTable = nil;
    p = nil;
}
func Cec4_ManStartNew(pAig *Gia_Man_t) *Gia_Man_t {
    var pObj *Gia_Obj_t; var i int32;
    var pNew *Gia_Man_t  = Gia_ManStart( Gia_ManObjNum(pAig) );
    pNew.pName = Abc_UtilStrsav( pAig.pName );
    pNew.pSpec = Abc_UtilStrsav( pAig.pSpec );
    if pAig.pMuxes {
        // inlining of macro ABC_CALLOC
        pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned))); }
    Gia_ManFillValue( pAig );
    Gia_ManConst0(pAig).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pAig.vCis)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    Gia_ManHashAlloc( pNew );
    Vec_IntFill( &pNew.vCopies2, Gia_ManObjNum(pAig), -1 );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(pAig) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_AddClausesMux(p *Gia_Man_t, pNode *Gia_Obj_t, pSat *sat_solver) {
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
    VarF = Cec4_ObjSatId(p, pNode);
    VarI = Cec4_ObjSatId(p, pNodeI);
    VarT = Cec4_ObjSatId(p, Gia_Regular(pNodeT));
    VarE = Cec4_ObjSatId(p, Gia_Regular(pNodeE));
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
    RetValue = sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 1);
    pLits[1] = Abc_Var2Lit(VarT, 0^fCompT);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeT).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 0);
    pLits[1] = Abc_Var2Lit(VarE, 1^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarI, 0);
    pLits[1] = Abc_Var2Lit(VarE, 0^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 1);
    if fPolarFlip != 0 {
        if pNodeI.fPhase {               pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( pSat, pLits, 3 );
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
    RetValue = sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
    pLits[0] = Abc_Var2Lit(VarT, 1^fCompT);
    pLits[1] = Abc_Var2Lit(VarE, 1^fCompE);
    pLits[2] = Abc_Var2Lit(VarF, 0);
    if fPolarFlip != 0 {
        if Gia_Regular(pNodeT).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
        if Gia_Regular(pNodeE).fPhase {  pLits[1] = Abc_LitNot( pLits[1] ); }
        if pNode.fPhase {                pLits[2] = Abc_LitNot( pLits[2] ); }
    }
    RetValue = sat_solver_addclause( pSat, pLits, 3 );
    assert((RetValue != 0), "RetValue");
}
func Cec4_AddClausesSuper(p *Gia_Man_t, pNode *Gia_Obj_t, vSuper *Vec_Ptr_t, pSat *sat_solver) {
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
        pLits[0] = Abc_Var2Lit(Cec4_ObjSatId(p, Gia_Regular(pFanin)), Gia_IsComplement(pFanin));
        pLits[1] = Abc_Var2Lit(Cec4_ObjSatId(p, pNode), 1);
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[0] = Abc_LitNot( pLits[0] ); }
            if pNode.fPhase {                pLits[1] = Abc_LitNot( pLits[1] ); }
        }
        RetValue = sat_solver_addclause( pSat, pLits, 2 );
        assert((RetValue != 0), "RetValue");
    }
    // add A & B => C   or   !A + !B + C
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        pLits[i] = Abc_Var2Lit(Cec4_ObjSatId(p, Gia_Regular(pFanin)), !Gia_IsComplement(pFanin));
        if fPolarFlip != 0 {
            if Gia_Regular(pFanin).fPhase {  pLits[i] = Abc_LitNot( pLits[i] ); }
        }
    }
    pLits[nLits-1] = Abc_Var2Lit(Cec4_ObjSatId(p, pNode), 0);
    if fPolarFlip != 0 {
        if pNode.fPhase {  pLits[nLits-1] = Abc_LitNot( pLits[nLits-1] ); }
    }
    RetValue = sat_solver_addclause( pSat, pLits, nLits );
    assert((RetValue != 0), "RetValue");
    pLits = nil;
}

/**Function*************************************************************

  Synopsis    [Adds clauses and returns CNF variable of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_CollectSuper_rec(pObj *Gia_Obj_t, vSuper *Vec_Ptr_t, fFirst int32, fUseMuxes int32) {
    // if the new node is complemented or a PI, another gate begins
    if Gia_IsComplement(pObj) || Gia_ObjIsCi(pObj) || 
         (!(fFirst != 0) && Gia_ObjValue(pObj) > 1) || 
         ((fUseMuxes != 0) && pObj.fMark0) {
        Vec_PtrPushUnique( vSuper, pObj );
        return;
    }
    // go through the branches
    Cec4_CollectSuper_rec( Gia_ObjChild0(pObj), vSuper, 0, fUseMuxes );
    Cec4_CollectSuper_rec( Gia_ObjChild1(pObj), vSuper, 0, fUseMuxes );
}
func Cec4_CollectSuper(pObj *Gia_Obj_t, fUseMuxes int32, vSuper *Vec_Ptr_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsCi(pObj), "!Gia_ObjIsCi(pObj)");
    Vec_PtrClear( vSuper );
    Cec4_CollectSuper_rec( pObj, vSuper, 1, fUseMuxes );
}
func Cec4_ObjAddToFrontier(p *Gia_Man_t, pObj *Gia_Obj_t, vFrontier *Vec_Ptr_t, pSat *sat_solver) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsConst0(pObj), "!Gia_ObjIsConst0(pObj)");
    if Cec4_ObjSatId(p, pObj) >= 0 {
        return; }
    assert(Cec4_ObjSatId(p, pObj) == -1, "Cec4_ObjSatId(p, pObj) == -1");
    Cec4_ObjSetSatId( p, pObj, sat_solver_addvar(pSat) );
    if Gia_ObjIsAnd(pObj) {
        Vec_PtrPush( vFrontier, pObj ); }
}
func Cec4_ObjGetCnfVar(p *Cec4_Man_t, iObj int32) int32 { 
    var fUseSimple int32  = 1; // enable simple CNF
    var fUseMuxes int32   = 1; // enable MUXes when using complex CNF
    var pNode, pFanin *Gia_Obj_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pNew, iObj);
    var i, k int32;
    // quit if CNF is ready
    if Cec4_ObjSatId(p.pNew,pObj) >= 0 {
        return Cec4_ObjSatId(p.pNew,pObj); }
    assert(iObj > 0, "iObj > 0");
    if Gia_ObjIsCi(pObj) {
        return Cec4_ObjSetSatId( p.pNew, pObj, sat_solver_addvar(p.pSat) ); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if fUseSimple != 0 {
        var pFan0, pFan1 *Gia_Obj_t;
        //if ( Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) )
        //    printf( "%d", (Gia_IsComplement(pFan1) << 1) + Gia_IsComplement(pFan0) );
        if p.pNew.pMuxes == nil && Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) && Gia_IsComplement(pFan0) == Gia_IsComplement(pFan1) {
            var iVar0 int32  = Cec4_ObjGetCnfVar( p, Gia_ObjId(p.pNew, Gia_Regular(pFan0)) );
            var iVar1 int32  = Cec4_ObjGetCnfVar( p, Gia_ObjId(p.pNew, Gia_Regular(pFan1)) );
            var iVar int32   = Cec4_ObjSetSatId( p.pNew, pObj, sat_solver_addvar(p.pSat) );
            if p.pPars.jType < 2 {
                sat_solver_add_xor( p.pSat, iVar, iVar0, iVar1, 0 ); }
            if p.pPars.jType > 0 {
                var Lit0 int32  = Abc_Var2Lit( iVar0, 0 );
                var Lit1 int32  = Abc_Var2Lit( iVar1, 0 );
                if Lit0 < Lit1 {
                     Lit1 ^= Lit0, Lit0 ^= Lit1, Lit1 ^= Lit0; }
                assert(Lit0 > Lit1, "Lit0 > Lit1");
                sat_solver_set_var_fanin_lit( p.pSat, iVar, Lit0, Lit1 );
                p.nGates[1]++;
            }
        
        } else {
            var iVar0 int32  = Cec4_ObjGetCnfVar( p, Gia_ObjFaninId0(pObj, iObj) );
            var iVar1 int32  = Cec4_ObjGetCnfVar( p, Gia_ObjFaninId1(pObj, iObj) );
            var iVar int32   = Cec4_ObjSetSatId( p.pNew, pObj, sat_solver_addvar(p.pSat) );
            if p.pPars.jType < 2 {
                if Gia_ObjIsXor(pObj) {
                    sat_solver_add_xor( p.pSat, iVar, iVar0, iVar1, Gia_ObjFaninC0(pObj) ^ Gia_ObjFaninC1(pObj) );
                } else {
                    sat_solver_add_and( p.pSat, iVar, iVar0, iVar1, Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 ); }
            }
            if p.pPars.jType > 0 {
                var Lit0 int32  = Abc_Var2Lit( iVar0, Gia_ObjFaninC0(pObj) );
                var Lit1 int32  = Abc_Var2Lit( iVar1, Gia_ObjFaninC1(pObj) );
                if (Lit0 > Lit1) ^ Gia_ObjIsXor(pObj) {
                     Lit1 ^= Lit0, Lit0 ^= Lit1, Lit1 ^= Lit0; }
                sat_solver_set_var_fanin_lit( p.pSat, iVar, Lit0, Lit1 );
                p.nGates[Gia_ObjIsXor(pObj)]++;
            }
        }
        return Cec4_ObjSatId( p.pNew, pObj );
    }
    assert(!Gia_ObjIsXor(pObj), "!Gia_ObjIsXor(pObj)");
    // start the frontier
    Vec_PtrClear( p.vFrontier );
    Cec4_ObjAddToFrontier( p.pNew, pObj, p.vFrontier, p.pSat );
    // explore nodes in the frontier
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFrontier)) && (((pNode) = (Gia_Obj_t *)Vec_PtrEntry(p.vFrontier, i)), 1); i++  {
        // create the supergate
        assert(Cec4_ObjSatId(p.pNew,pNode) >= 0, "Cec4_ObjSatId(p.pNew,pNode) >= 0");
        if (fUseMuxes != 0) && pNode.fMark0 {
            Vec_PtrClear( p.vFanins );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin0( Gia_ObjFanin1(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin0(pNode) ) );
            Vec_PtrPushUnique( p.vFanins, Gia_ObjFanin1( Gia_ObjFanin1(pNode) ) );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(p.vFanins, k)), 1); k++  {
                Cec4_ObjAddToFrontier( p.pNew, Gia_Regular(pFanin), p.vFrontier, p.pSat );
            }
            Cec4_AddClausesMux( p.pNew, pNode, p.pSat );
        
        } else {
            Cec4_CollectSuper( pNode, fUseMuxes, p.vFanins );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanins)) && (((pFanin) = (Gia_Obj_t *)Vec_PtrEntry(p.vFanins, k)), 1); k++  {
                Cec4_ObjAddToFrontier( p.pNew, Gia_Regular(pFanin), p.vFrontier, p.pSat );
            }
            Cec4_AddClausesSuper( p.pNew, pNode, p.vFanins, p.pSat );
        }
        assert(Vec_PtrSize(p.vFanins) > 1, "Vec_PtrSize(p.vFanins) > 1");
    }
    return Cec4_ObjSatId(p.pNew,pObj);
}


/**Function*************************************************************

  Synopsis    [Refinement of equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_ObjSim(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, p.nSimWords * iObj );
}
func Cec4_ObjSimEqual(p *Gia_Man_t, iObj0 int32, iObj1 int32) int32 {
    var w int32;
    var pSim0 *word  = Cec4_ObjSim( p, iObj0 );
    var pSim1 *word  = Cec4_ObjSim( p, iObj1 );
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

func Cec4_RefineOneClassIter(p *Gia_Man_t, iRepr int32) {
    var iObj int32;
    var iPrev int32 = iRepr;
    var iPrev2 int32;
    var iRepr2 int32;
    assert(Gia_ObjRepr(p, iRepr) == GIA_VOID, "Gia_ObjRepr(p, iRepr) == GIA_VOID");
    assert(Gia_ObjNext(p, iRepr) > 0, "Gia_ObjNext(p, iRepr) > 0");
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iRepr2 = Gia_ObjNext(p, iRepr); iRepr2 > 0; iRepr2 = Gia_ObjNext(p, iRepr2)  {
        if Cec4_ObjSimEqual(p, iRepr, iRepr2) != 0 {
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
        if Cec4_ObjSimEqual(p, iRepr, iObj) != 0 { // remains with iRepr
        
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
    // refine incrementally
    if Gia_ObjNext(p, iRepr2) > 0 {
        Cec4_RefineOneClassIter( p, iRepr2 ); }
}
func Cec4_RefineOneClass(p *Gia_Man_t, pMan *Cec4_Man_t, vNodes *Vec_Int_t) {
    var k, iObj, Bin int32;
    Vec_IntClear( pMan.vRefBins );
    // inlining of macro Vec_IntForEachEntryReverse
    for  k = Vec_IntSize(vNodes) - 1; (k >= 0) && (((iObj) = Vec_IntEntry(vNodes, k)), 1); k--  {
        var Key int32  = Cec4_ManSimHashKey( Cec4_ObjSim(p, iObj), p.nSimWords, pMan.nTableSize );
        assert(Key >= 0 && Key < pMan.nTableSize, "Key >= 0 && Key < pMan.nTableSize");
        if pMan.pTable[Key] == -1 {
            Vec_IntPush( pMan.vRefBins, Key ); }
        p.pNexts[iObj] = pMan.pTable[Key];
        pMan.pTable[Key] = iObj;
    }
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(pMan.vRefBins)) && (((Bin) = Vec_IntEntry(pMan.vRefBins, k)), 1); k++  {
        var iRepr int32  = pMan.pTable[Bin];
        pMan.pTable[Bin] = -1;
        assert(p.pReprs[iRepr].iRepr == GIA_VOID, "p.pReprs[iRepr].iRepr == GIA_VOID");
        assert(p.pNexts[iRepr] != 0, "p.pNexts[iRepr] != 0");
        if p.pNexts[iRepr] == -1 {
            continue; }
        for  iObj = p.pNexts[iRepr]; iObj > 0; iObj = p.pNexts[iObj]  {
            p.pReprs[iObj].iRepr = iRepr;
        }
        Cec4_RefineOneClassIter( p, iRepr );
    }
    Vec_IntClear( pMan.vRefBins );
}
func Cec4_RefineClasses(p *Gia_Man_t, pMan *Cec4_Man_t, vClasses *Vec_Int_t) {
    if Vec_IntSize(pMan.vRefClasses) == 0 {
        return; }
    if Vec_IntSize(pMan.vRefNodes) > 0 {
        Cec4_RefineOneClass( p, pMan, pMan.vRefNodes );
    } else {
        var i, k, iObj, iRepr int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(pMan.vRefClasses)) && (((iRepr) = Vec_IntEntry(pMan.vRefClasses, i)), 1); i++  {
            assert(p.pReprs[iRepr].fColorA, "p.pReprs[iRepr].fColorA");
            p.pReprs[iRepr].fColorA = 0;
            Vec_IntClear( pMan.vRefNodes );
            Vec_IntPush( pMan.vRefNodes, iRepr );
            // inlining of macro Gia_ClassForEachObj1
            for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), k = Gia_ObjNext(p, iRepr); k > 0; k = Gia_ObjNext(p, k)  {
                Vec_IntPush( pMan.vRefNodes, k );
            }
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(pMan.vRefNodes)) && (((iObj) = Vec_IntEntry(pMan.vRefNodes, k)), 1); k++  {
                p.pReprs[iObj].iRepr = GIA_VOID;
                p.pNexts[iObj] = -1;
            }
            Cec4_RefineOneClass( p, pMan, pMan.vRefNodes );
        }
    }
    Vec_IntClear( pMan.vRefClasses );
    Vec_IntClear( pMan.vRefNodes );
}
func Cec4_RefineInit(p *Gia_Man_t, pMan *Cec4_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    p.pReprs = nil;
    p.pNexts = nil;
    // inlining of macro ABC_CALLOC
    p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    // inlining of macro ABC_FALLOC
    p.pNexts = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    pMan.nTableSize  = Abc_PrimeCudd( Gia_ManObjNum(p) );
    // inlining of macro ABC_FALLOC
    pMan.pTable      = ((int *) memset(malloc(sizeof(int) * (size_t)(pMan.nTableSize)), 0xff, sizeof(int) * (size_t)(pMan.nTableSize)));
    pMan.vRefNodes   = Vec_IntAlloc( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        p.pReprs[i].iRepr = GIA_VOID;
        if !Gia_ObjIsCo(pObj) && (!pMan.pPars.nLevelMax || Gia_ObjLevel(p, pObj) <= pMan.pPars.nLevelMax) {
            Vec_IntPush( pMan.vRefNodes, i ); }
    }
    pMan.vRefBins    = Vec_IntAlloc( Gia_ManObjNum(p)/2 );
    pMan.vRefClasses = Vec_IntAlloc( Gia_ManObjNum(p)/2 );
    Vec_IntPush( pMan.vRefClasses, 0 );
}


/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_ObjSimSetInputBit(p *Gia_Man_t, iObj int32, Bit int32) {
    var pSim *word  = Cec4_ObjSim( p, iObj );
    if Abc_InfoHasBit( (unsigned*)pSim, p.iPatsPi ) != Bit {
        Abc_InfoXorBit( (unsigned*)pSim, p.iPatsPi ); }
}
func Cec4_ObjSimGetInputBit(p *Gia_Man_t, iObj int32) int32 {
    var pSim *word  = Cec4_ObjSim( p, iObj );
    return Abc_InfoHasBit( (unsigned*)pSim, p.iPatsPi );
}
func Cec4_ObjSimRo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSimRo *word  = Cec4_ObjSim( p, iObj );
    var pSimRi *word  = Cec4_ObjSim( p, Gia_ObjRoToRiId(p, iObj) );
    for  w = 0; w < p.nSimWords; w++  {
        pSimRo[w] = pSimRi[w];
    }
}
func Cec4_ObjSimCo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSimCo *word   = Cec4_ObjSim( p, iObj );
    var pSimDri *word  = Cec4_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] = ~pSimDri[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] =  pSimDri[w];
        } }
}
func Cec4_ObjSimAnd(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSim *word   = Cec4_ObjSim( p, iObj );
    var pSim0 *word  = Cec4_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSim1 *word  = Cec4_ObjSim( p, Gia_ObjFaninId1(pObj, iObj) );
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
func Cec4_ObjSimXor(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSim *word   = Cec4_ObjSim( p, iObj );
    var pSim0 *word  = Cec4_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSim1 *word  = Cec4_ObjSim( p, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) ^ Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] ^ pSim1[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] =  pSim0[w] ^ pSim1[w];
        } }
}
func Cec4_ObjSimCi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSim *word  = Cec4_ObjSim( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSim[w] = Gia_ManRandomW( 0 );
    }
    pSim[0] <<= 1;
}
func Cec4_ObjClearSimCi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSim *word  = Cec4_ObjSim( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSim[w] = 0;
    }
}
func Cec4_ManSimulateCis(p *Gia_Man_t) {
    var i, Id int32;
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Cec4_ObjSimCi( p, Id );
    }
    p.iPatsPi = 0;
}
func Cec4_ManClearCis(p *Gia_Man_t) {
    var i, Id int32;
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Cec4_ObjClearSimCi( p, Id );
    }
    p.iPatsPi = 0;
}
func Cec4_ManDeriveCex(p *Gia_Man_t, iOut int32, iPat int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var i, Id int32;
    pCex = Abc_CexAlloc( 0, Gia_ManCiNum(p), 1 );
    pCex.iPo = iOut;
    if iPat == -1 {
        return pCex; }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        if Abc_InfoHasBit((unsigned *)Cec4_ObjSim(p, Id), iPat) {
            Abc_InfoSetBit( pCex.pData, i ); }
    }
    return pCex;
}
func Cec4_ManSimulateCos(p *Gia_Man_t) int32 {
    var i, Id int32;
    // check outputs and generate CEX if they fail
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        Cec4_ObjSimCo( p, Id );
        if Cec4_ObjSimEqual(p, Id, 0) != 0 {
            continue; }
        p.pCexSeq = Cec4_ManDeriveCex( p, i, Abc_TtFindFirstBit2(Cec4_ObjSim(p, Id), p.nSimWords) );
        return 0;
    }
    return 1;
}
func Cec4_ManSimulate(p *Gia_Man_t, pMan *Cec4_Man_t) {
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; var i int32;
    pMan.nSimulates++;
    if pMan.pTable == nil {
        Cec4_RefineInit( p, pMan );
    } else {
        assert(Vec_IntSize(pMan.vRefClasses) == 0, "Vec_IntSize(pMan.vRefClasses) == 0"); }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var iRepr int32  = Gia_ObjRepr( p, i );
        if Gia_ObjIsXor(pObj) {
            Cec4_ObjSimXor( p, i );
        } else {
            Cec4_ObjSimAnd( p, i ); }
        if iRepr == GIA_VOID || p.pReprs[iRepr].fColorA || (Cec4_ObjSimEqual(p, iRepr, i) != 0) {
            continue; }
        p.pReprs[iRepr].fColorA = 1;
        Vec_IntPush( pMan.vRefClasses, iRepr );
    }
    }
    pMan.timeSim += Abc_Clock() - clk;
    clk = Abc_Clock();
    Cec4_RefineClasses( p, pMan, pMan.vRefClasses );
    pMan.timeRefine += Abc_Clock() - clk;
}
func Cec4_ManSimulate_rec(p *Gia_Man_t, pMan *Cec4_Man_t, iObj int32) {
    var pObj *Gia_Obj_t; 
    if !(iObj != 0) || Vec_IntEntry(pMan.vCexStamps, iObj) == p.iPatsPi {
        return; }
    Vec_IntWriteEntry( pMan.vCexStamps, iObj, p.iPatsPi );
    pObj = Gia_ManObj(p, iObj);
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Cec4_ManSimulate_rec( p, pMan, Gia_ObjFaninId0(pObj, iObj) );
    Cec4_ManSimulate_rec( p, pMan, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjIsXor(pObj) {
        Cec4_ObjSimXor( p, iObj );
    } else {
        Cec4_ObjSimAnd( p, iObj ); }
}
func Cec4_ManSimAlloc(p *Gia_Man_t, nWords int32) {
    Vec_WrdFreeP( &p.vSims );
    Vec_WrdFreeP( &p.vSimsPi );
    p.vSims     = Vec_WrdStart( Gia_ManObjNum(p) * nWords );
    p.vSimsPi   = Vec_WrdStart( (Gia_ManCiNum(p) + 1) * nWords );
    p.nSimWords = nWords;
}


/**Function*************************************************************

  Synopsis    [Creating initial equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_ManPrintTfiConeStats(p *Gia_Man_t) {
    var vRoots *Vec_Int_t   = Vec_IntAlloc( 100 );
    var vNodes *Vec_Int_t   = Vec_IntAlloc( 100 );
    var vLeaves *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i, k int32;
    // inlining of macro Gia_ManForEachClass0
    for  i = 0; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Vec_IntClear( vRoots );
        if i % 100 != 0 {
            continue; }
        Vec_IntPush( vRoots, i );
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), k = Gia_ObjNext(p, i); k > 0; k = Gia_ObjNext(p, k)  {
            Vec_IntPush( vRoots, k );
        }
        Gia_ManCollectTfi( p, vRoots, vNodes );
        fmt.Printf("Class %6d : ", i);
        fmt.Printf("Roots = %6d  ", Vec_IntSize(vRoots));
        fmt.Printf("Nodes = %6d  ", Vec_IntSize(vNodes));
        fmt.Printf("\n");
    }
    }
    Vec_IntFree( vRoots );
    Vec_IntFree( vNodes );
    Vec_IntFree( vLeaves );
}
var spStatic_Cec4_ManPrintStats_clk abctime;
func Cec4_ManPrintStats(p *Gia_Man_t, pPars *Cec_ParFra_t, pMan *Cec4_Man_t, fSim int32) {
    
    var clkThis abctime;
    var i, nLits, Counter, Counter0, CounterX int32;
    if !pPars.fVerbose {
        return; }
    if pMan.nItersSim + pMan.nItersSat {
        clkThis = Abc_Clock() - spStatic_Cec4_ManPrintStats_clk; }
    spStatic_Cec4_ManPrintStats_clk = Abc_Clock();
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        if Gia_ObjIsHead(p, i) {
            Counter++;
        } else if Gia_ObjIsConst(p, i) {
            Counter0++;
        } else if Gia_ObjIsNone(p, i) {
            CounterX++; }
    }
    nLits = Gia_ManObjNum(p) - Counter - CounterX;
    if fSim != 0 {
        fmt.Printf("Sim %4d : ", pMan.nItersSim++ + pMan.nItersSat);
        fmt.Printf("%6.2f %%  ", 100.0*nLits/Gia_ManCandNum(p));
    
    } else {
        fmt.Printf("SAT %4d : ", pMan.nItersSim + pMan.nItersSat++);
        fmt.Printf("%6.2f %%  ", 100.0*pMan.nAndNodes/Gia_ManAndNum(p));
    }
    fmt.Printf("P =%7d  ", pMan ? pMan.nSatUnsat : 0);
    fmt.Printf("D =%7d  ", pMan ? pMan.nSatSat   : 0);
    fmt.Printf("F =%8d  ", pMan ? pMan.nSatUndec : 0);
    //printf( "Last =%6d  ", pMan ? pMan->iLastConst : 0 );
    Abc_Print( 1, "cst =%9d  cls =%8d  lit =%9d   ",  Counter0, Counter, nLits );
    Abc_PrintTime( 1, "Time", clkThis );
}
func Cec4_ManPrintClasses2(p *Gia_Man_t) {
    var i, k int32;
    // inlining of macro Gia_ManForEachClass0
    for  i = 0; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        fmt.Printf("Class %d : ", i);
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), k = Gia_ObjNext(p, i); k > 0; k = Gia_ObjNext(p, k)  {
            fmt.Printf("%d ", k);
        }
        fmt.Printf("\n");
    }
    }
}
func Cec4_ManPrintClasses(p *Gia_Man_t) {
    var k, Count int32;
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, 0), "Gia_ObjIsHead(p, 0)"), k = Gia_ObjNext(p, 0); k > 0; k = Gia_ObjNext(p, k)  {
        Count++;
    }
    fmt.Printf("Const0 class has %d entries.\n", Count);
}


/**Function*************************************************************

  Synopsis    [Verify counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_ManVerify_rec(p *Gia_Man_t, iObj int32, pSat *sat_solver) int32 {
    var Value0, Value1 int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if iObj == 0 { return 0; }
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return pObj.fMark1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if Gia_ObjIsCi(pObj) {
        return pObj.fMark1 = sat_solver_read_cex_varvalue(pSat, Cec4_ObjSatId(p, pObj)); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Cec4_ManVerify_rec( p, Gia_ObjFaninId0(pObj, iObj), pSat ) ^ Gia_ObjFaninC0(pObj);
    Value1 = Cec4_ManVerify_rec( p, Gia_ObjFaninId1(pObj, iObj), pSat ) ^ Gia_ObjFaninC1(pObj);
    if pObj.fMark1 = Gia_ObjIsXor(pObj) {
        return Value0 ^ Value1;
    } else {
        return Value0 & Value1;
    }
}
func Cec4_ManVerify(p *Gia_Man_t, iObj0 int32, iObj1 int32, fPhase int32, pSat *sat_solver) {
    var Value0, Value1 int32;
    Gia_ManIncrementTravId( p );
    Value0 = Cec4_ManVerify_rec( p, iObj0, pSat );
    Value1 = Cec4_ManVerify_rec( p, iObj1, pSat );
    if (Value0 ^ Value1) == fPhase {
        fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", iObj0, iObj1); }
//    else
//        printf( "CEX verification succeeded for obj %d and obj %d.\n", iObj0, iObj1 );;
}


/**Function*************************************************************

  Synopsis    [Verify counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_ManCexVerify_rec(p *Gia_Man_t, iObj int32) int32 {
    var Value0, Value1 int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if iObj == 0 { return 0; }
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return pObj.fMark1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if Gia_ObjIsCi(pObj) {
        return pObj.fMark1 = Cec4_ObjSimGetInputBit(p, iObj); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Value0 = Cec4_ManCexVerify_rec( p, Gia_ObjFaninId0(pObj, iObj) ) ^ Gia_ObjFaninC0(pObj);
    Value1 = Cec4_ManCexVerify_rec( p, Gia_ObjFaninId1(pObj, iObj) ) ^ Gia_ObjFaninC1(pObj);
    if pObj.fMark1 = Gia_ObjIsXor(pObj) {
        return Value0 ^ Value1;
    } else {
        return Value0 & Value1;
    }
}
func Cec4_ManCexVerify(p *Gia_Man_t, iObj0 int32, iObj1 int32, fPhase int32) {
    var Value0, Value1 int32;
    Gia_ManIncrementTravId( p );
    Value0 = Cec4_ManCexVerify_rec( p, iObj0 );
    Value1 = Cec4_ManCexVerify_rec( p, iObj1 );
    if (Value0 ^ Value1) == fPhase {
        fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", iObj0, iObj1); }
//    else
//        printf( "CEX verification succeeded for obj %d and obj %d.\n", iObj0, iObj1 );;
}

/**Function*************************************************************

  Synopsis    [Packs simulation patterns into array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

*************************************`**********************************/
func Cec4_ManPackAddPatterns(p *Gia_Man_t, iBit int32, vLits *Vec_Int_t) {
    var k int32;
    var Limit int32 = Abc_MinInt( Vec_IntSize(vLits), 64 * p.nSimWords - 1 );
    for  k = 0; k < Limit; k++  {
        var i int32;
        var Lit int32;
        var iBitLocal int32 = (iBit + k + 1) % Limit + 1;
        assert(iBitLocal > 0 && iBitLocal < 64 * p.nSimWords, "iBitLocal > 0 && iBitLocal < 64 * p.nSimWords");
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
            var pInfo *word  = Vec_WrdEntryP( p.vSims,   p.nSimWords * Abc_Lit2Var(Lit) );
            var pPres *word  = Vec_WrdEntryP( p.vSimsPi, p.nSimWords * Abc_Lit2Var(Lit) );
            if Abc_InfoHasBit( (unsigned *)pPres, iBitLocal ) {
                continue; }
            if Abc_InfoHasBit( (unsigned *)pInfo, iBitLocal ) != Abc_LitIsCompl(Lit ^ (i == k)) {
                 Abc_InfoXorBit( (unsigned *)pInfo, iBitLocal ); }
        }
    }
}
func Cec4_ManPackAddPatternTry(p *Gia_Man_t, iBit int32, vLits *Vec_Int_t) int32 {
    var i, Lit int32;
    assert(p.iPatsPi > 0 && p.iPatsPi < 64 * p.nSimWords, "p.iPatsPi > 0 && p.iPatsPi < 64 * p.nSimWords");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
        var pInfo *word  = Vec_WrdEntryP( p.vSims,   p.nSimWords * Abc_Lit2Var(Lit) );
        var pPres *word  = Vec_WrdEntryP( p.vSimsPi, p.nSimWords * Abc_Lit2Var(Lit) );
        if Abc_InfoHasBit( (unsigned *)pPres, iBit ) && 
             Abc_InfoHasBit( (unsigned *)pInfo, iBit ) != Abc_LitIsCompl(Lit) {
             return 0; }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLits)) && (((Lit) = Vec_IntEntry(vLits, i)), 1); i++  {
        var pInfo *word  = Vec_WrdEntryP( p.vSims,   p.nSimWords * Abc_Lit2Var(Lit) );
        var pPres *word  = Vec_WrdEntryP( p.vSimsPi, p.nSimWords * Abc_Lit2Var(Lit) );
        Abc_InfoSetBit( (unsigned *)pPres, iBit );
        if Abc_InfoHasBit( (unsigned *)pInfo, iBit ) != Abc_LitIsCompl(Lit) {
             Abc_InfoXorBit( (unsigned *)pInfo, iBit ); }
    }
    return 1;
}
func Cec4_ManPackAddPattern(p *Gia_Man_t, vLits *Vec_Int_t, fExtend int32) int32 {
    var k int32;
    for  k = 1; k < 64 * p.nSimWords - 1; k++  {
        if ++p.iPatsPi == 64 * p.nSimWords - 1 {
            p.iPatsPi = 1; }
        if Cec4_ManPackAddPatternTry( p, p.iPatsPi, vLits ) != 0 {
            if fExtend != 0 {
                Cec4_ManPackAddPatterns( p, p.iPatsPi, vLits ); }
            break;
        }
    }
    if k == 64 * p.nSimWords - 1 {
        p.iPatsPi = k;
        if !(Cec4_ManPackAddPatternTry( p, p.iPatsPi, vLits ) != 0) {
            fmt.Printf("Internal error.\n");
        } else if fExtend != 0 {
            Cec4_ManPackAddPatterns( p, p.iPatsPi, vLits ); }
        return 64 * p.nSimWords;
    }
    return k;
}

/**Function*************************************************************

  Synopsis    [Generates counter-examples to refine the candidate equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/









func Cec4_ManCandIterStart(p *Cec4_Man_t) {
    var i int32;
    var pArray *int32;
    assert(p.iPosWrite == 0, "p.iPosWrite == 0");
    assert(p.iPosRead == 0, "p.iPosRead == 0");
    assert(Vec_IntSize(p.vCands) == 0, "Vec_IntSize(p.vCands) == 0");
    for  i = 1; i < Gia_ManObjNum(p.pAig); i++  {
        if Gia_ObjRepr(p.pAig, i) != GIA_VOID {
            Vec_IntPush( p.vCands, i ); }
    }
    pArray = Vec_IntArray( p.vCands );
    for  i = 0; i < Vec_IntSize(p.vCands); i++  {
        var iNew int32  = Abc_Random(0) % Vec_IntSize(p.vCands);
        // inlining of macro ABC_SWAP
        { var t int32  = pArray[i]; pArray[i] = pArray[iNew]; pArray[iNew] = t; };
    }
}
func Cec4_ManCandIterNext(p *Cec4_Man_t) int32 {
    for  Vec_IntSize(p.vCands) > 0  {
        var fStop int32;
        var iCand int32 = Vec_IntEntry( p.vCands, p.iPosRead );
        if fStop = (Gia_ObjRepr(p.pAig, iCand) != GIA_VOID) {
            Vec_IntWriteEntry( p.vCands, p.iPosWrite++, iCand ); }
        if ++p.iPosRead == Vec_IntSize(p.vCands) {
            Vec_IntShrink( p.vCands, p.iPosWrite );
            p.iPosWrite = 0;
            p.iPosRead = 0;
        }
        if fStop != 0 {
            return iCand; }
    }
    return 0;
}
func Cec4_ManGeneratePatterns(p *Cec4_Man_t) int32 {
    var clk abctime  = Abc_Clock();
    var i int32;
    var iCand int32;
    var nPats int32 = 100 * 64 * p.pAig.nSimWords;
    var CountPat int32;
    var Packs int32;
    //int iRepr;
    //Vec_IntForEachEntryDouble( p->vDisprPairs, iRepr, iCand, i )
    //    if ( iRepr == Gia_ObjRepr(p->pAig, iCand) )
    //        printf( "Pair %6d (%6d, %6d) (new repr = %9d) is FAILED to disprove.\n", i, iRepr, iCand, Gia_ObjRepr(p->pAig, iCand) );
    //    else
    //        printf( "Pair %6d (%6d, %6d) (new repr = %9d) is disproved.\n", i, iRepr, iCand, Gia_ObjRepr(p->pAig, iCand) );
    //Vec_IntClear( p->vDisprPairs );
    p.pAig.iPatsPi = 0;
    Vec_WrdFill( p.pAig.vSimsPi, Vec_WrdSize(p.pAig.vSimsPi), 0 );
    for  i = 0; i < nPats; i++  {
        if iCand = Cec4_ManCandIterNext(p) {
            var iRepr int32     = Gia_ObjRepr( p.pAig, iCand );
            var iCandVal int32  = Gia_ManObj(p.pAig, iCand).fPhase;
            var iReprVal int32  = Gia_ManObj(p.pAig, iRepr).fPhase;
            var Res int32  = Cec4_ManGeneratePatternOne( p.pAig, iRepr,  iReprVal, iCand, !iCandVal, p.vPat, p.vVisit );
            if !(Res != 0) {
                Res = Cec4_ManGeneratePatternOne( p.pAig, iRepr, !iReprVal, iCand,  iCandVal, p.vPat, p.vVisit ); }
            if Res != 0 {
                var Ret int32  = Cec4_ManPackAddPattern( p.pAig, p.vPat, 1 );
                //Vec_IntPushTwo( p->vDisprPairs, iRepr, iCand );
                Packs += Ret;
                if Ret == 64 * p.pAig.nSimWords {
                    break; }
                if CountPat++; CountPat == 8 * 64 * p.pAig.nSimWords {
                    break; }
                //Cec4_ManCexVerify( p->pAig, iRepr, iCand, iReprVal ^ iCandVal );
                //Gia_ManCleanMark01( p->pAig );
            }
        }
    }
    p.timeGenPats += Abc_Clock() - clk;
    p.nSatSat += CountPat;
    //printf( "%3d : %6.2f %% : Generated %6d CEXs after trying %6d pairs.  Ave packs = %9.2f  Ave tries = %9.2f  (Limit = %9.2f)\n", 
    //    p->nItersSim++, 100.0*Vec_IntSize(p->vCands)/Gia_ManAndNum(p->pAig),
    //    CountPat, i, (float)Packs / Abc_MaxInt(1, CountPat), (float)i / Abc_MaxInt(1, CountPat), (float)nPats / p->pPars->nItersMax );
    return CountPat >= i / p.pPars.nItersMax;
}


/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_ManSatSolverRecycle(p *Cec4_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    //printf( "Solver size = %d.\n", sat_solver_varnum(p->pSat) );
    p.nRecycles++;
    p.nCallsSince = 0;
    sat_solver_reset( p.pSat );
    // clean mapping of AigIds into SatIds
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(&p.pNew.vSuppVars)) && ((pObj) = Gia_ManObj(p.pNew, Vec_IntEntry(&p.pNew.vSuppVars,i))); i++  {
        Cec4_ObjCleanSatId( p.pNew, pObj );
    }
    Vec_IntClear( &p.pNew.vSuppVars  );  // AigIds for which SatId is defined
    Vec_IntClear( &p.pNew.vCopiesTwo );  // pairs (CiAigId, SatId)
    Vec_IntClear( &p.pNew.vVarMap    );  // mapping of SatId into AigId
}
func Cec4_ManSolveTwo(p *Cec4_Man_t, iObj0 int32, iObj1 int32, fPhase int32, pfEasy *int32, fVerbose int32) int32 {
    var clk abctime;
    var nBTLimit int32;
    if Vec_BitEntry(p.vFails, iObj0) || Vec_BitEntry(p.vFails, iObj1) {
        nBTLimit = Abc_MaxInt(1, p.pPars.nBTLimit/10);
    } else {
        nBTLimit = p.pPars.nBTLimit;
    }
    var nConfEnd int32;
    var nConfBeg int32;
    var status int32;
    var iVar0 int32;
    var iVar1 int32;
    var Lits [2]int32;
    var UnsatConflicts [3]int32 ;
    if iObj1 <  iObj0 { 
         iObj1 ^= iObj0, iObj0 ^= iObj1, iObj1 ^= iObj0; }
    assert(iObj0 < iObj1, "iObj0 < iObj1");    
    *pfEasy = 0;
    // check if SAT solver needs recycling
    p.nCallsSince++; 
    if p.nCallsSince > p.pPars.nCallsRecycle && 
         Vec_IntSize(&p.pNew.vSuppVars) > p.pPars.nSatVarMax && p.pPars.nSatVarMax {
        Cec4_ManSatSolverRecycle( p ); }
    // add more logic to the solver
    if !(iObj0 != 0) && Cec4_ObjSatId(p.pNew, Gia_ManConst0(p.pNew)) == -1 {
        Cec4_ObjSetSatId( p.pNew, Gia_ManConst0(p.pNew), sat_solver_addvar(p.pSat) ); }
    clk = Abc_Clock();
    iVar0 = Cec4_ObjGetCnfVar( p, iObj0 );
    iVar1 = Cec4_ObjGetCnfVar( p, iObj1 );
    if p.pPars.jType > 0 {
        sat_solver_start_new_round( p.pSat );
        sat_solver_mark_cone( p.pSat, Cec4_ObjSatId(p.pNew, Gia_ManObj(p.pNew, iObj0)) );
        sat_solver_mark_cone( p.pSat, Cec4_ObjSatId(p.pNew, Gia_ManObj(p.pNew, iObj1)) );
    }
    p.timeCnf += Abc_Clock() - clk;
    // perform solving
    Lits[0] = Abc_Var2Lit(iVar0, 1);
    Lits[1] = Abc_Var2Lit(iVar1, fPhase);
    sat_solver_set_conflict_budget( p.pSat, nBTLimit );
    nConfBeg = sat_solver_conflictnum( p.pSat );
    status = sat_solver_solve( p.pSat, Lits, 2 );
    nConfEnd = sat_solver_conflictnum( p.pSat );
    assert(nConfEnd >= nConfBeg, "nConfEnd >= nConfBeg");
    if fVerbose != 0 {
        if status == GLUCOSE_SAT {
            p.nConflicts[0][0] += nConfEnd == nConfBeg;
            p.nConflicts[0][1] += nConfEnd -  nConfBeg;
            p.nConflicts[0][2]  = Abc_MaxInt(p.nConflicts[0][2], nConfEnd - nConfBeg);
            *pfEasy = nConfEnd == nConfBeg;
        
        } else if status == GLUCOSE_UNSAT {
            if iObj0 > 0 {
                UnsatConflicts[0] = nConfEnd == nConfBeg;
                UnsatConflicts[1] = nConfEnd -  nConfBeg;
                UnsatConflicts[2] = Abc_MaxInt(p.nConflicts[1][2], nConfEnd - nConfBeg);
            
            } else {
                p.nConflicts[1][0] += nConfEnd == nConfBeg;
                p.nConflicts[1][1] += nConfEnd -  nConfBeg;
                p.nConflicts[1][2]  = Abc_MaxInt(p.nConflicts[1][2], nConfEnd - nConfBeg);
                *pfEasy = nConfEnd == nConfBeg;
            }
        
        } else { 
            return status; }
    }
    if status == GLUCOSE_UNSAT && iObj0 > 0 {
        Lits[0] = Abc_Var2Lit(iVar0, 0);
        Lits[1] = Abc_Var2Lit(iVar1, !fPhase);
        sat_solver_set_conflict_budget( p.pSat, nBTLimit );
        nConfBeg = sat_solver_conflictnum( p.pSat );
        status = sat_solver_solve( p.pSat, Lits, 2 );
        nConfEnd = sat_solver_conflictnum( p.pSat );
        assert(nConfEnd >= nConfBeg, "nConfEnd >= nConfBeg");
        if fVerbose != 0 {
            if status == GLUCOSE_SAT {
                p.nConflicts[0][0] += nConfEnd == nConfBeg;
                p.nConflicts[0][1] += nConfEnd -  nConfBeg;
                p.nConflicts[0][2]  = Abc_MaxInt(p.nConflicts[0][2], nConfEnd - nConfBeg);
                *pfEasy = nConfEnd == nConfBeg;
            
            } else if status == GLUCOSE_UNSAT {
                UnsatConflicts[0]  &= nConfEnd == nConfBeg;
                UnsatConflicts[1]  += nConfEnd -  nConfBeg;
                UnsatConflicts[2]   = Abc_MaxInt(p.nConflicts[1][2], nConfEnd - nConfBeg);

                p.nConflicts[1][0] += UnsatConflicts[0];
                p.nConflicts[1][1] += UnsatConflicts[1];
                p.nConflicts[1][2]  = Abc_MaxInt(p.nConflicts[1][2], UnsatConflicts[2]);
                *pfEasy = UnsatConflicts[0];
            }
        }
    }
    return status;
}
func Cec4_ManSweepNode(p *Cec4_Man_t, iObj int32, iRepr int32) int32 {
    var clk abctime  = Abc_Clock();
    var i int32;
    var IdAig int32;
    var IdSat int32;
    var status int32;
    var fEasy int32;
    var RetValue int32 = 1;
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pAig, iObj );
    var pRepr *Gia_Obj_t  = Gia_ManObj( p.pAig, iRepr );
    var fCompl int32  = Abc_LitIsCompl(pObj.Value) ^ Abc_LitIsCompl(pRepr.Value) ^ pObj.fPhase ^ pRepr.fPhase;
    status = Cec4_ManSolveTwo( p, Abc_Lit2Var(pRepr.Value), Abc_Lit2Var(pObj.Value), fCompl, &fEasy, p.pPars.fVerbose );
    if status == GLUCOSE_SAT {
        var iLit int32;
        //int iPatsOld = p->pAig->iPatsPi;
        //printf( "Disproved: %d == %d.\n", Abc_Lit2Var(pRepr->Value), Abc_Lit2Var(pObj->Value) );
        p.nSatSat++;
        p.nPatterns++;
        Vec_IntClear( p.vPat );
        if p.pPars.jType == 0 {
            // inlining of macro Vec_IntForEachEntryDouble
            for  i = 0; (i+1 < Vec_IntSize(&p.pNew.vCopiesTwo)) && (((IdAig) = Vec_IntEntry(&p.pNew.vCopiesTwo, i)), 1) && (((IdSat) = Vec_IntEntry(&p.pNew.vCopiesTwo, i+1)), 1); i += 2  {
                Vec_IntPush( p.vPat, Abc_Var2Lit(IdAig, sat_solver_read_cex_varvalue(p.pSat, IdSat)) );
            }
        
        } else {
            var pCex *int32  = sat_solver_read_cex( p.pSat );
            var pMap *int32  = Vec_IntArray(&p.pNew.vVarMap);
            for  i = 0; i < pCex[0];  {
                Vec_IntPush( p.vPat, Abc_Lit2LitV(pMap, Abc_LitNot(pCex[++i])) );
            }
        }
        assert(p.pAig.iPatsPi >= 0 && p.pAig.iPatsPi < 64 * p.pAig.nSimWords - 1, "p.pAig.iPatsPi >= 0 && p.pAig.iPatsPi < 64 * p.pAig.nSimWords - 1");
        p.pAig.iPatsPi++;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vPat)) && (((iLit) = Vec_IntEntry(p.vPat, i)), 1); i++  {
            Cec4_ObjSimSetInputBit( p.pAig, Abc_Lit2Var(iLit), Abc_LitIsCompl(iLit) );
        }
        //Cec4_ManPackAddPattern( p->pAig, p->vPat, 0 );
        //assert( iPatsOld + 1 == p->pAig->iPatsPi );
        if fEasy != 0 {
            p.timeSatSat0 += Abc_Clock() - clk;
        } else {
            p.timeSatSat += Abc_Clock() - clk; }
        RetValue = 0;
        // this is not needed, but we keep it here anyway, because it takes very little time
        //Cec4_ManVerify( p->pNew, Abc_Lit2Var(pRepr->Value), Abc_Lit2Var(pObj->Value), fCompl, p->pSat );
        // resimulated once in a while
        if p.pAig.iPatsPi == 64 * p.pAig.nSimWords - 2 {
            var clk2 abctime  = Abc_Clock();
            Cec4_ManSimulate( p.pAig, p );
            //printf( "FasterSmall = %d.  FasterBig = %d.\n", p->nFaster[0], p->nFaster[1] );
            p.nFaster[0] = p.nFaster[1] = 0;
            //if ( p->nSatSat && p->nSatSat % 100 == 0 )
                Cec4_ManPrintStats( p.pAig, p.pPars, p, 0 );
            Vec_IntFill( p.vCexStamps, Gia_ManObjNum(p.pAig), 0 );
            p.pAig.iPatsPi = 0;
            Vec_WrdFill( p.pAig.vSimsPi, Vec_WrdSize(p.pAig.vSimsPi), 0 );
            p.timeResimGlo += Abc_Clock() - clk2;
        }
    
    } else if status == GLUCOSE_UNSAT {
        //printf( "Proved: %d == %d.\n", Abc_Lit2Var(pRepr->Value), Abc_Lit2Var(pObj->Value) );
        p.nSatUnsat++;
        pObj.Value = Abc_LitNotCond( pRepr.Value, fCompl );
        Gia_ObjSetProved( p.pAig, iObj );
        if iRepr == 0 {
            p.iLastConst = iObj; }
        if fEasy != 0 {
            p.timeSatUnsat0 += Abc_Clock() - clk;
        } else {
            p.timeSatUnsat += Abc_Clock() - clk; }
        RetValue = 1;
    
    } else {
        p.nSatUndec++;
        assert(status == GLUCOSE_UNDEC, "status == GLUCOSE_UNDEC");
        Gia_ObjSetFailed( p.pAig, iObj );
        Vec_BitWriteEntry( p.vFails, iObj, 1 );
        //if ( iRepr )
        //Vec_BitWriteEntry( p->vFails, iRepr, 1 );
        p.timeSatUndec += Abc_Clock() - clk;
        RetValue = 2;
    }
    return RetValue;
}
func Cec4_ManFindRepr(p *Gia_Man_t, pMan *Cec4_Man_t, iObj int32) *Gia_Obj_t {
    var clk abctime  = Abc_Clock();
    var iMem, iRepr int32;
    assert(Gia_ObjHasRepr(p, iObj), "Gia_ObjHasRepr(p, iObj)");
    assert(!Gia_ObjProved(p, iObj), "!Gia_ObjProved(p, iObj)");
    iRepr = Gia_ObjRepr(p, iObj);
    assert(iRepr != iObj, "iRepr != iObj");
    assert(!Gia_ObjProved(p, iRepr), "!Gia_ObjProved(p, iRepr)");
    Cec4_ManSimulate_rec( p, pMan, iObj );
    Cec4_ManSimulate_rec( p, pMan, iRepr );
    if Cec4_ObjSimEqual(p, iObj, iRepr) != 0 {
        pMan.timeResimLoc += Abc_Clock() - clk;
        return Gia_ManObj(p, iRepr);
    }
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iMem = Gia_ObjNext(p, iRepr); iMem > 0; iMem = Gia_ObjNext(p, iMem)  {
        if iObj == iMem {
            break; }
        if Gia_ObjProved(p, iMem) || Gia_ObjFailed(p, iMem) {
            continue; }
        Cec4_ManSimulate_rec( p, pMan, iMem );
        if Cec4_ObjSimEqual(p, iObj, iMem) != 0 {
            pMan.nFaster[0]++;
            pMan.timeResimLoc += Abc_Clock() - clk;
            return Gia_ManObj(p, iMem);
        }
    }
    pMan.nFaster[1]++;
    pMan.timeResimLoc += Abc_Clock() - clk;
    return nil;
}
func Cec4_ManPerformSweeping(p *Gia_Man_t, pPars *Cec_ParFra_t, ppNew **Gia_Man_t, fSimOnly int32) int32 {
    var pMan *Cec4_Man_t  = Cec4_ManCreate( p, pPars ); 
    var pObj, pRepr *Gia_Obj_t; 
    var i int32;
    var fSimulate int32 = 1;
    if pPars.fVerbose {
        fmt.Printf("Solver type = %d. Simulate %d words in %d rounds. SAT with %d confs. Recycle after %d SAT calls.\n", pPars.jType, pPars.nWords, pPars.nRounds, pPars.nBTLimit, pPars.nCallsRecycle); }

    // this is currently needed to have a correct mapping
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        assert(Gia_ObjId(p, pObj) == i+1, "Gia_ObjId(p, pObj) == i+1");
    }

    // check if any output trivially fails under all-0 pattern
    Gia_ManRandom( 1 );
    Gia_ManSetPhase( p );
    if pPars.nLevelMax {
        Gia_ManLevelNum(p); }
    //Gia_ManStaticFanoutStart( p );
    if pPars.fCheckMiter {
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if pObj.fPhase {
                p.pCexSeq = Cec4_ManDeriveCex( p, i, -1 );
                goto finalize;
            }
        }
    }

    // simulate one round and create classes
    Cec4_ManSimAlloc( p, pPars.nWords );
    Cec4_ManSimulateCis( p );
    Cec4_ManSimulate( p, pMan );
    if pPars.fCheckMiter && !(Cec4_ManSimulateCos(p) != 0) { // cex detected
        goto finalize; }
    if pPars.fVerbose {
        Cec4_ManPrintStats( p, pPars, pMan, 1 ); }

    // perform simulation
    for  i = 0; i < pPars.nRounds; i++  {
        Cec4_ManSimulateCis( p );
        Cec4_ManSimulate( p, pMan );
        if pPars.fCheckMiter && !(Cec4_ManSimulateCos(p) != 0) { // cex detected
            goto finalize; }
        if (i != 0) && i % (pPars.nRounds / 5) == 0 && pPars.fVerbose {
            Cec4_ManPrintStats( p, pPars, pMan, 1 ); }
    }
    if fSimOnly != 0 {
        goto finalize; }

    // perform additional simulation
    Cec4_ManCandIterStart( pMan );
    for  i = 0; (fSimulate != 0) && i < pPars.nGenIters; i++  {
        Cec4_ManSimulateCis( p );
        fSimulate = Cec4_ManGeneratePatterns( pMan );
        Cec4_ManSimulate( p, pMan );
        if pPars.fCheckMiter && !(Cec4_ManSimulateCos(p) != 0) { // cex detected
            goto finalize; }
        if (i != 0) && i % 5 == 0 && pPars.fVerbose {
            Cec4_ManPrintStats( p, pPars, pMan, 1 ); }
    }
    if (i != 0) && i % 5 && pPars.fVerbose {
        Cec4_ManPrintStats( p, pPars, pMan, 1 ); }

    p.iPatsPi = 0;
    Vec_WrdFill( p.vSimsPi, Vec_WrdSize(p.vSimsPi), 0 );
    pMan.nSatSat = 0;
    pMan.pNew = Cec4_ManStartNew( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var pObjNew *Gia_Obj_t; 
        pMan.nAndNodes++;
        if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManHashXorReal( pMan.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else {
            pObj.Value = Gia_ManHashAnd( pMan.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        if pPars.nLevelMax && Gia_ObjLevel(p, pObj) > pPars.nLevelMax {
            continue; }
        pObjNew = Gia_ManObj( pMan.pNew, Abc_Lit2Var(pObj.Value) );
        if Gia_ObjIsAnd(pObjNew) {
        if Vec_BitEntry(pMan.vFails, Gia_ObjFaninId0(pObjNew, Abc_Lit2Var(pObj.Value))) || 
             Vec_BitEntry(pMan.vFails, Gia_ObjFaninId1(pObjNew, Abc_Lit2Var(pObj.Value))) {
            Vec_BitWriteEntry( pMan.vFails, Abc_Lit2Var(pObjNew.Value), 1 ); } }
        //if ( Gia_ObjIsAnd(pObjNew) )
        //    Gia_ObjSetAndLevel( pMan->pNew, pObjNew );
        // select representative based on candidate equivalence classes
        pRepr = Gia_ObjReprObj( p, i );
        if pRepr == nil {
            continue; }
        if true { // select representative based on recent counter-examples
        
            pRepr = Cec4_ManFindRepr( p, pMan, i );
            if pRepr == nil {
                continue; }
        }
        if Abc_Lit2Var(pObj.Value) == Abc_Lit2Var(pRepr.Value) {
            assert((pObj.Value ^ pRepr.Value) == (pObj.fPhase ^ pRepr.fPhase), "(pObj.Value ^ pRepr.Value) == (pObj.fPhase ^ pRepr.fPhase)");
            Gia_ObjSetProved( p, i );
            if Gia_ObjId(p, pRepr) == 0 {
                pMan.iLastConst = i; }
            continue;
        }
        if (Cec4_ManSweepNode(pMan, i, Gia_ObjId(p, pRepr)) != 0) && Gia_ObjProved(p, i) {
            pObj.Value = Abc_LitNotCond( pRepr.Value, pObj.fPhase ^ pRepr.fPhase ); }
    }
    }
    if p.iPatsPi > 0 {
        var clk2 abctime  = Abc_Clock();
        Cec4_ManSimulate( p, pMan );
        p.iPatsPi = 0;
        Vec_IntFill( pMan.vCexStamps, Gia_ManObjNum(p), 0 );
        pMan.timeResimGlo += Abc_Clock() - clk2;
    }
    if pPars.fVerbose {
        Cec4_ManPrintStats( p, pPars, pMan, 0 ); }
    if ppNew {
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            pObj.Value = Gia_ManAppendCo( pMan.pNew, Gia_ObjFanin0Copy(pObj) );
        }
        *ppNew = Gia_ManCleanup( pMan.pNew );
    }
finalize:
    if pPars.fVerbose {
        fmt.Printf("SAT calls = %d:  P = %d (0=%d a=%.2f m=%d)  D = %d (0=%d a=%.2f m=%d)  F = %d   Sim = %d  Recyc = %d  Xor = %.2f %%\n", pMan.nSatUnsat + pMan.nSatSat + pMan.nSatUndec, pMan.nSatUnsat, pMan.nConflicts[1][0], float32(pMan.nConflicts[1][1])/Abc_MaxInt(1, pMan.nSatUnsat-pMan.nConflicts[1][0]), pMan.nConflicts[1][2], pMan.nSatSat, pMan.nConflicts[0][0], float32(pMan.nConflicts[0][1])/Abc_MaxInt(1, pMan.nSatSat  -pMan.nConflicts[0][0]), pMan.nConflicts[0][2], pMan.nSatUndec, pMan.nSimulates, pMan.nRecycles, 100.0*pMan.nGates[1]/Abc_MaxInt(1, pMan.nGates[0]+pMan.nGates[1])); }
    Cec4_ManDestroy( pMan );
    //Gia_ManStaticFanoutStop( p );
    //Gia_ManEquivPrintClasses( p, 1, 0 );
    if p.pCexSeq {
        return 0;
    } else {
        return 1;
    }
}
func Cec4_ManSimulateTest(p *Gia_Man_t, pPars *Cec_ParFra_t) *Gia_Man_t {
    var pNew *Gia_Man_t  = nil;
    Cec4_ManPerformSweeping( p, pPars, &pNew, 0 );
    return pNew;
}
func Cec4_ManSimulateTest2(p *Gia_Man_t, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var ParsFra Cec_ParFra_t;
    var pPars *Cec_ParFra_t = &ParsFra;
    Cec4_ManSetParams( pPars );
    Cec4_ManPerformSweeping( p, pPars, nil, 0 );
    pPars.fVerbose = fVerbose;
    if fVerbose != 0 {
        Abc_PrintTime( 1, "New choice computation time", Abc_Clock() - clk ); }
}
func Cec4_ManSimulateTest3(p *Gia_Man_t, nBTLimit int32, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t  = nil;
    var ParsFra Cec_ParFra_t;
    var pPars *Cec_ParFra_t = &ParsFra;
    Cec4_ManSetParams( pPars );
    pPars.fVerbose = fVerbose;
    pPars.nBTLimit = nBTLimit;
    Cec4_ManPerformSweeping( p, pPars, &pNew, 0 );
    return pNew;
}
func Cec4_ManSimulateOnlyTest(p *Gia_Man_t, fVerbose int32) int32 {
    var ParsFra Cec_ParFra_t;
    var pPars *Cec_ParFra_t = &ParsFra;
    Cec4_ManSetParams( pPars );
    pPars.fVerbose = fVerbose;
    return Cec4_ManPerformSweeping( p, pPars, nil, 1 );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

