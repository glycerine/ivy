package abs

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [absGla.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Gate-level abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absGla.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver2.h"
//#include "base/main/main.h"
//#include "abs.h"
//#include "absRef.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// iterator through abstracted objects
// Original macro:
// #define Gla_ManForEachObj( p, pObj )                       \
//     for ( pObj = p->pObjs + 1; pObj < p->pObjs + p->nObjs; pObj++ ) 

// Original macro:
// #define Gla_ManForEachObjAbs( p, pObj, i )                 \
//     for ( i = 0; i < Vec_IntSize(p->vAbs) && ((pObj = Gla_ManObj(p, Vec_IntEntry(p->vAbs, i))),1); i++) 

// Original macro:
// #define Gla_ManForEachObjAbsVec( vVec, p, pObj, i )        \
//     for ( i = 0; i < Vec_IntSize(vVec) && ((pObj = Gla_ManObj(p, Vec_IntEntry(vVec, i))),1); i++) 


// iterator through fanins of an abstracted object
// Original macro:
// #define Gla_ObjForEachFanin( p, pObj, pFanin, i )          \
//     for ( i = 0; (i < (int)pObj->nFanins) && ((pFanin = Gla_ManObj(p, pObj->Fanins[i])),1); i++ )


// some lessons learned from debugging mismatches between GIA and mapped CNF
// - inputs/output of AND-node may be PPIs (have SAT vars), but the node is not included in the abstraction
// - some logic node can be a PPI of one LUT and an internal node of another LUT

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Prepares CEX and vMap for refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Derives counter-example using current assignments.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collects GIA abstraction objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Drive implications of the given node towards primary outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Selects assignments to be refined.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs refinement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gla_ManRefinement2_Sign int32;
func Gla_ManRefinement2(p *Gla_Man_t) *Vec_Int_t {
    var fVerify int32  = 1;
    
    var vPis *Vec_Int_t;
    var vPPis *Vec_Int_t;
    var vCos *Vec_Int_t;
    var vRoAnds *Vec_Int_t;
    var vSelect *Vec_Int_t = nil;
    var pRef, pRef0, pRef1 *Rfn_Obj_t;
    var pObj *Gia_Obj_t;
    var i, f int32;
    spStatic_Gla_ManRefinement2_Sign++;

    // compute PIs and pseudo-PIs
    vCos = Vec_IntAlloc( 1000 );
    vPis = Vec_IntAlloc( 1000 );  
    vPPis = Vec_IntAlloc( 1000 );
    vRoAnds = Vec_IntAlloc( 1000 );  
    Gla_ManCollect( p, vPis, vPPis, vCos, vRoAnds );

/*
    // check how many pseudo PIs have variables
    Gla_ManForEachObjAbsVec( vPis, p, pGla, i )
    {
        Abc_Print( 1, "  %5d : ", Gla_ObjId(p, pGla) );
        for ( f = 0; f <= p->pPars->iFrame; f++ )
            Abc_Print( 1, "%d", Gla_ManCheckVar(p, Gla_ObjId(p, pGla), f) );
        Abc_Print( 1, "\n" );
    }    

    // check how many pseudo PIs have variables
    Gla_ManForEachObjAbsVec( vPPis, p, pGla, i )
    {
        Abc_Print( 1, "%5d : ", Gla_ObjId(p, pGla) );
        for ( f = 0; f <= p->pPars->iFrame; f++ )
            Abc_Print( 1, "%d", Gla_ManCheckVar(p, Gla_ObjId(p, pGla), f) );
        Abc_Print( 1, "\n" );
    }    
*/
    // propagate values
    for  f = 0; f <= p.pPars.iFrame; f++  {
        // constant
        pRef = Gla_ObjRef( p, Gia_ManConst0(p.pGia), f );  Gla_ObjClearRef( pRef );
        pRef.Value  = 0;
        pRef.Prio   = 0;
        pRef.Sign   = spStatic_Gla_ManRefinement2_Sign;
        // primary input
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPis,i))); i++  {
//            assert( f == p->pPars->iFrame || Gla_ManCheckVar(p, p->pObj2Obj[Gia_ObjId(p->pGia, pObj)], f) );
            pRef = Gla_ObjRef( p, pObj, f );   Gla_ObjClearRef( pRef );
            pRef.Value = Gla_ObjSatValue( p, Gia_ObjId(p.pGia, pObj), f );
            pRef.Prio  = 0;
            pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
            assert(pRef.fVisit == 0, "pRef.fVisit == 0");
        }
        // primary input
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vPPis)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vPPis,i))); i++  {
//            assert( f == p->pPars->iFrame || Gla_ManCheckVar(p, p->pObj2Obj[Gia_ObjId(p->pGia, pObj)], f) );
            assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj)");
            pRef = Gla_ObjRef( p, pObj, f );   Gla_ObjClearRef( pRef );
            pRef.Value = Gla_ObjSatValue( p, Gia_ObjId(p.pGia, pObj), f );
            pRef.Prio  = i+1;
            pRef.fPPi  = 1;
            pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
            assert(pRef.fVisit == 0, "pRef.fVisit == 0");
        }
        // internal nodes
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vRoAnds)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vRoAnds,i))); i++  {
            assert(Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsAnd(pObj) || Gia_ObjIsRo(p.pGia, pObj)");
            pRef = Gla_ObjRef( p, pObj, f );   Gla_ObjClearRef( pRef );
            if Gia_ObjIsRo(p.pGia, pObj) {
                if f == 0 {
                    pRef.Value = 0;
                    pRef.Prio  = 0;
                    pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
                
                } else {
                    pRef0 = Gla_ObjRef( p, Gia_ObjRoToRi(p.pGia, pObj), f-1 );
                    pRef.Value = pRef0.Value;
                    pRef.Prio  = pRef0.Prio;
                    pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
                }
                continue;
            }
            assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
            pRef0 = Gla_ObjRef( p, Gia_ObjFanin0(pObj), f );
            pRef1 = Gla_ObjRef( p, Gia_ObjFanin1(pObj), f );
            pRef.Value = (pRef0.Value ^ Gia_ObjFaninC0(pObj)) & (pRef1.Value ^ Gia_ObjFaninC1(pObj));

            if p.pObj2Obj[Gia_ObjId(p.pGia, pObj)] != ~0 && 
                 (Gla_ManCheckVar(p, p.pObj2Obj[Gia_ObjId(p.pGia, pObj)], f) != 0) &&
                 int32(pRef.Value) != Gla_ObjSatValue(p, Gia_ObjId(p.pGia, pObj), f) {
                    Abc_Print( 1, "Object has value mismatch    " );
                    Gia_ObjPrint( p.pGia, pObj );
            }

            if pRef.Value == 1 {
                pRef.Prio  = Abc_MaxInt( pRef0.Prio, pRef1.Prio );
            } else if (pRef0.Value ^ Gia_ObjFaninC0(pObj)) == 0 && (pRef1.Value ^ Gia_ObjFaninC1(pObj)) == 0 {
                pRef.Prio  = Abc_MinInt( pRef0.Prio, pRef1.Prio ); // choice
            } else if (pRef0.Value ^ Gia_ObjFaninC0(pObj)) == 0 {
                pRef.Prio  = pRef0.Prio;
            } else { 
                pRef.Prio  = pRef1.Prio; }
            assert(pRef.fVisit == 0, "pRef.fVisit == 0");
            pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
        }
        // output nodes
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vCos)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vCos,i))); i++  {
            pRef = Gla_ObjRef( p, pObj, f );    Gla_ObjClearRef( pRef );
            pRef0 = Gla_ObjRef( p, Gia_ObjFanin0(pObj), f ); 
            pRef.Value = (pRef0.Value ^ Gia_ObjFaninC0(pObj));
            pRef.Prio  = pRef0.Prio;
            assert(pRef.fVisit == 0, "pRef.fVisit == 0");
            pRef.Sign  = spStatic_Gla_ManRefinement2_Sign;
        }
    } 

    // make sure the output value is 1
    pObj = Gia_ManPo( p.pGia, 0 );
    pRef = Gla_ObjRef( p, pObj, p.pPars.iFrame );
    if pRef.Value != 1 {
        Abc_Print( 1, "\nCounter-example verification has failed!!!\n" ); }

    // check the CEX
    if pRef.Prio == 0 {
        p.pGia.pCexSeq = Gla_ManDeriveCex( p, vPis );
        Vec_IntFree( vPis );
        Vec_IntFree( vPPis );
        Vec_IntFree( vRoAnds );
        Vec_IntFree( vCos );
        return nil;
    }

    // select objects
    vSelect = Vec_IntAlloc( 100 );
    Vec_IntFill( p.vObjCounts, p.pPars.iFrame+1, 0 );
    Gla_ManRefSelect_rec( p, Gia_ObjFanin0(Gia_ManPo(p.pGia, 0)), p.pPars.iFrame, vSelect, spStatic_Gla_ManRefinement2_Sign );
    Vec_IntUniqify( vSelect );

/*
    for ( f = 0; f < p->pPars->iFrame; f++ )
        printf( "%2d", Vec_IntEntry(p->vObjCounts, f) );
    printf( "\n" );
*/
    if fVerify != 0 {
        Gla_ManVerifyUsingTerSim( p, vPis, vPPis, vRoAnds, vCos, vSelect ); }

    // remap them into GLA objects
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vSelect)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(vSelect,i))); i++  {
        Vec_IntWriteEntry( vSelect, i, p.pObj2Obj[Gia_ObjId(p.pGia, pObj)] );
    }

    Vec_IntFree( vPis );
    Vec_IntFree( vPPis );
    Vec_IntFree( vRoAnds );
    Vec_IntFree( vCos );

    p.nObjAdded += Vec_IntSize(vSelect);
    return vSelect;
}


/**Function*************************************************************

  Synopsis    [Adds clauses for the given obj in the given frame.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Duplicates AIG while decoupling nodes duplicated in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManStart2(pGia *Gia_Man_t, pPars *Abs_Par_t) *Gla_Man_t {
    
    var pAig *Aig_Man_t;
    var pObj *Gia_Obj_t;
    var pGla *Gla_Obj_t;
    var i int32;
    var pLits *int32;
    // start
    // inlining of macro ABC_CALLOC
    p := &Gla_Man_t{};
    p.pGia  = pGia;
    p.pPars = pPars;
    p.vAbs = Vec_IntAlloc( 100 );
    p.vTemp = Vec_IntAlloc( 100 );
    p.vAddedNew = Vec_IntAlloc( 100 );
    // internal data
    pAig = Gia_ManToAigSimple( p.pGia );
    p.pCnf  = Cnf_DeriveOther( pAig, 1 );
    Aig_ManStop( pAig );
    // count the number of variables
    p.nObjs = 1;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        if p.pCnf.pObj2Count[i] >= 0 {
            pObj.Value = p.nObjs++;
        } else {
            pObj.Value = ~0; }
    }
    // re-express CNF using new variable IDs
    pLits = p.pCnf.pClauses[0];
    for  i = 0; i < p.pCnf.nLiterals; i++  {
        pObj = Gia_ManObj( p.pGia, lit_var(pLits[i]) );
        assert(~pObj.Value, "~pObj.Value");
        pLits[i] = toLitCond( pObj.Value, lit_sign(pLits[i]) );
    }
    // create objects 
    // inlining of macro ABC_CALLOC
    p.pObjs    = ((Gla_Obj_t *) calloc((size_t)(p.nObjs), sizeof(Gla_Obj_t)));
    // inlining of macro ABC_FALLOC
    p.pObj2Obj = ((unsigned *) memset(malloc(sizeof(unsigned) * (size_t)(Gia_ManObjNum(p.pGia))), 0xff, sizeof(unsigned) * (size_t)(Gia_ManObjNum(p.pGia))));
//    p->pvRefis  = ABC_CALLOC( Vec_Int_t, Gia_ManObjNum(p->pGia) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        p.pObj2Obj[i] = pObj.Value;
        if !~pObj.Value {
            continue; }
        pGla = Gla_ManObj( p, pObj.Value );
        pGla.iGiaObj = i;
        pGla.fCompl0 = Gia_ObjFaninC0(pObj);
        pGla.fConst  = Gia_ObjIsConst0(pObj);
        pGla.fPi     = Gia_ObjIsPi(p.pGia, pObj);
        pGla.fPo     = Gia_ObjIsPo(p.pGia, pObj);
        pGla.fRi     = Gia_ObjIsRi(p.pGia, pObj);
        pGla.fRo     = Gia_ObjIsRo(p.pGia, pObj);
        pGla.fAnd    = Gia_ObjIsAnd(pObj);
        if Gia_ObjIsConst0(pObj) || Gia_ObjIsPi(p.pGia, pObj) {
            continue; }
        if Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj) {
            Gla_ManCollectFanins( p, pGla, pObj.Value, p.vTemp );
            pGla.nFanins = Vec_IntSize( p.vTemp );
            memcpy( pGla.Fanins, Vec_IntArray(p.vTemp), sizeof(int) * Vec_IntSize(p.vTemp) );
            continue;
        }
        assert(Gia_ObjIsRo(p.pGia, pObj), "Gia_ObjIsRo(p.pGia, pObj)");
        pGla.nFanins   = 1;
        pGla.Fanins[0] = Gia_ObjFanin0( Gia_ObjRoToRi(p.pGia, pObj) ).Value;
        pGla.fCompl0   = Gia_ObjFaninC0( Gia_ObjRoToRi(p.pGia, pObj) );
    }
    p.pObjRoot = Gla_ManObj( p, Gia_ManPo(p.pGia, 0).Value );
    // abstraction 
    assert(pGia.vGateClasses != nil, "pGia.vGateClasses != nil");
    // inlining of macro Gla_ManForEachObj
    for  pGla = p.pObjs + 1; pGla < p.pObjs + p.nObjs; pGla++  {
        if Vec_IntEntry( pGia.vGateClasses, pGla.iGiaObj ) == 0 {
            continue; }
        pGla.fAbs = 1;
        Vec_IntPush( p.vAbs, Gla_ObjId(p, pGla) );
    }
    // other 
    p.pSat      = sat_solver2_new();
    p.nSatVars  = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gla_ManStop(p *Gla_Man_t) {
    var pGla *Gla_Obj_t;
    var i int32;

    if p.pPars.fVerbose {
        Abc_Print( 1, "SAT solver:  Var = %d  Cla = %d  Conf = %d  Lrn = %d  Reduce = %d  Cex = %d  Objs+ = %d\n", 
            sat_solver2_nvars(p.pSat), sat_solver2_nclauses(p.pSat), sat_solver2_nconflicts(p.pSat), 
            sat_solver2_nlearnts(p.pSat), p.pSat.nDBreduces, p.nCexes, p.nObjAdded ); }

    // stop the refinement manager
//    Gia_ManStopP( &p->pGia2 );
    Rnm_ManStop( p.pRnm, 0 );

    if p.pvRefis {
    for  i = 0; i < Gia_ManObjNum(p.pGia); i++  {
        p.pvRefis[i].pArray = nil;
    } }
    // inlining of macro Gla_ManForEachObj
    for  pGla = p.pObjs + 1; pGla < p.pObjs + p.nObjs; pGla++  { 
        pGla.vFrames.pArray = nil;
    }
    Cnf_DataFree( p.pCnf );
    if p.pGia0 != nil {
        Gia_ManStop( p.pGia ); }
//    Gia_ManStaticFanoutStart( p->pGia0 );
    sat_solver2_delete( p.pSat );
    Vec_IntFreeP( &p.vObjCounts );
    Vec_IntFreeP( &p.vAddedNew );
    Vec_IntFreeP( &p.vCoreCounts );
    Vec_IntFreeP( &p.vProofIds );
    Vec_IntFreeP( &p.vTemp );
    Vec_IntFreeP( &p.vAbs );
    p.pvRefis = nil;
    p.pObj2Obj = nil;
    p.pObjs = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Creates a new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Derives new abstraction map.]

  Description [Returns 1 if node contains abstracted leaf on the path.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collect pseudo-PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Adds CNF for the given timeframe.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Finds the set of clauses involved in the UNSAT core.]

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
func Gla_ManAbsPrintFrame(p *Gla_Man_t, nCoreSize int32, nFrames int32, nConfls int32, nCexes int32, Time abctime) {
    if Abc_FrameIsBatchMode() && nCoreSize <= 0 {
        return; }
    Abc_Print( 1, "%4d :", nFrames-1 );
    Abc_Print( 1, "%4d", Abc_MinInt(100, 100 * Gia_GlaAbsCount(p, 0, 0) / (p.nObjs - Gia_ManPoNum(p.pGia) + Gia_ManCoNum(p.pGia) + 1)) ); 
    Abc_Print( 1, "%6d", Gia_GlaAbsCount(p, 0, 0) );
    Abc_Print( 1, "%5d", Gla_ManCountPPis(p) );
    Abc_Print( 1, "%5d", Gia_GlaAbsCount(p, 1, 0) );
    Abc_Print( 1, "%6d", Gia_GlaAbsCount(p, 0, 1) );
    Abc_Print( 1, "%8d", nConfls );
    if nCexes == 0 {
        Abc_Print( 1, "%5c", '-' ); 
    } else {
        Abc_Print( 1, "%5d", nCexes ); } 
//    Abc_Print( 1, " %9d", sat_solver2_nvars(p->pSat) ); 
    Abc_PrintInt( sat_solver2_nvars(p.pSat) );
    Abc_PrintInt( sat_solver2_nclauses(p.pSat) );
    Abc_PrintInt( sat_solver2_nlearnts(p.pSat) );
//    Abc_Print( 1, " %6d", nCoreSize > 0 ? nCoreSize : 0 ); 
    Abc_Print( 1, "%9.2f sec", 1.0*Time/int64(time.Second) );
    Abc_Print( 1, "%5.0f MB", (sat_solver2_memory_proof(p.pSat) + sat_solver2_memory(p.pSat, 0)) / (1<<20) );
//    Abc_PrintInt( p->nAbsNew );
//    Abc_PrintInt( p->nLrnNew );
//    Abc_Print( 1, "%4.1f MB", 4.0 * p->nLrnNew * Abc_BitWordNum(p->nAbsNew) / (1<<20) );
    Abc_Print( 1, "%s", (nCoreSize > 0 && nCexes > 0) ? "\n" : "\r" );
    fflush( os.Stdout );
}
func Gla_ManReportMemory(p *Gla_Man_t) {
    var pGla *Gla_Obj_t;
    var memTot float64;
    var memAig float64  = Gia_ManObjNum(p.pGia) * sizeof(Gia_Obj_t);
    var memSat float64  = sat_solver2_memory( p.pSat, 1 );
    var memPro float64  = sat_solver2_memory_proof( p.pSat );
    var memMap float64  = p.nObjs * sizeof(Gla_Obj_t) + Gia_ManObjNum(p.pGia) * sizeof(int);
    var memRef float64  = Rnm_ManMemoryUsage( p.pRnm );
    var memOth float64  = sizeof(Gla_Man_t);
    for  pGla = p.pObjs; pGla < p.pObjs + p.nObjs; pGla++  {
        memMap += Vec_IntCap(&pGla.vFrames) * sizeof(int);
    }
    memOth += Vec_IntCap(p.vAddedNew) * sizeof(int);
    memOth += Vec_IntCap(p.vTemp) * sizeof(int);
    memOth += Vec_IntCap(p.vAbs) * sizeof(int);
    memTot = memAig + memSat + memPro + memMap + memRef + memOth;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: AIG      "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memAig)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memAig)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: SAT      "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memSat)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memSat)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Proof    "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memPro)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memPro)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Map      "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memMap)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memMap)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Refine   "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memRef)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memRef)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: Other    "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memOth)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memOth)))/(float64((memTot))) : 0.0) );;
    // inlining of macro ABC_PRMP
    Abc_Print(1, "%s =", ("Memory: TOTAL    "));
Abc_Print(1, "%10.3f MB (%6.2f %%)\n",  (1.0*(float64((memTot)))/(1<<20)), ((float64((memTot)))? 100.0*(float64((memTot)))/(float64((memTot))) : 0.0) );;
}


/**Function*************************************************************

  Synopsis    [Send abstracted model or send cancel.]

  Description [Counter-example will be sent automatically when &vta terminates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaSendAbsracted(p *Gla_Man_t, fVerbose int32) {
    var pAbs *Gia_Man_t;
    var vGateClasses *Vec_Int_t;
    assert(Abc_FrameIsBridgeMode(), "Abc_FrameIsBridgeMode()");
//    if ( fVerbose )
//        Abc_Print( 1, "Sending abstracted model...\n" );
    // create abstraction (value of p->pGia is not used here)
    vGateClasses = Gla_ManTranslate( p );
    pAbs = Gia_ManDupAbsGates( p.pGia0, vGateClasses );
    Vec_IntFreeP( &vGateClasses );
    // send it out
    Gia_ManToBridgeAbsNetlist( os.Stdout, pAbs, BRIDGE_ABS_NETLIST );
    Gia_ManStop( pAbs );
}
func Gia_GlaSendCancel(p *Gla_Man_t, fVerbose int32) {
    extern int Gia_ManToBridgeBadAbs( FILE * pFile );
    assert(Abc_FrameIsBridgeMode(), "Abc_FrameIsBridgeMode()");
//    if ( fVerbose )
//        Abc_Print( 1, "Cancelling previously sent model...\n" );
    Gia_ManToBridgeBadAbs( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Send abstracted model or send cancel.]

  Description [Counter-example will be sent automatically when &vta terminates.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaDumpAbsracted(p *Gla_Man_t, fVerbose int32) {
    var pFileNameDef string  = "glabs.aig";
    var pFileName string;
    if p.pPars.pFileVabs {
        pFileName = p.pPars.pFileVabs;
    } else {
        pFileName = pFileNameDef;
    }
    var pAbs *Gia_Man_t;
    var vGateClasses *Vec_Int_t;
    if fVerbose != 0 {
        Abc_Print( 1, "Dumping abstracted model into file \"%s\"...\n", pFileName ); }
    // create abstraction
    vGateClasses = Gla_ManTranslate( p );
    pAbs = Gia_ManDupAbsGates( p.pGia0, vGateClasses );
    Vec_IntFreeP( &vGateClasses );
    // write into file
    Gia_AigerWrite( pAbs, pFileName, 0, 0, 0 );
    Gia_ManStop( pAbs );
}


/**Function*************************************************************

  Synopsis    [Performs gate-level abstraction]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformGlaOld(pAig *Gia_Man_t, pPars *Abs_Par_t, fStartVta int32) int32 {
    extern int Gia_VtaPerformInt( Gia_Man_t * pAig, Abs_Par_t * pPars );
    extern void Ga2_ManDumpStats( Gia_Man_t * pGia, Abs_Par_t * pPars, sat_solver2 * pSat, int iFrame, int fUseN );
    var p *Gla_Man_t;
    var vPPis, vCore *Vec_Int_t;//, * vCore2 = NULL;
    var pCex *Abc_Cex_t  = nil;
    var f int32;
    var i int32;
    var iPrev int32;
    var nConfls int32;
    var Status int32;
    var nVarsOld int32;
    var nCoreSize int32;
    var fOneIsSent int32;
    var RetValue int32 = -1;
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    // preconditions
    assert(Gia_ManPoNum(pAig) == 1, "Gia_ManPoNum(pAig) == 1");
    assert(pPars.nFramesMax == 0 || pPars.nFramesStart <= pPars.nFramesMax, "pPars.nFramesMax == 0 || pPars.nFramesStart <= pPars.nFramesMax");
    if Gia_ObjIsConst0(Gia_ObjFanin0(Gia_ManPo(pAig,0))) {
        if !Gia_ObjFaninC0(Gia_ManPo(pAig,0)) {
            fmt.Printf("Sequential miter is trivially UNSAT.\n");
            return 1;
        }
        pAig.pCexSeq = nil;
        pAig.pCexSeq = Abc_CexMakeTriv( Gia_ManRegNum(pAig), Gia_ManPiNum(pAig), 1, 0 );
        fmt.Printf("Sequential miter is trivially SAT.\n");
        return 0;
    }

    // compute intial abstraction
    if pAig.vGateClasses == nil {
        if fStartVta != 0 {
            var nFramesMaxOld int32    = pPars.nFramesMax;
            var nFramesStartOld int32  = pPars.nFramesStart;
            var nTimeOutOld int32      = pPars.nTimeOut;
            var nDumpOld int32         = pPars.fDumpVabs;
            pPars.nFramesMax   = pPars.nFramesStart;
            pPars.nFramesStart = Abc_MinInt( pPars.nFramesStart/2 + 1, 3 );
            pPars.nTimeOut     = 20;
            pPars.fDumpVabs    = 0;
            RetValue = Gia_VtaPerformInt( pAig, pPars );
            pPars.nFramesMax   = nFramesMaxOld;
            pPars.nFramesStart = nFramesStartOld;
            pPars.nTimeOut     = nTimeOutOld;
            pPars.fDumpVabs    = nDumpOld;
            // create gate classes
            Vec_IntFreeP( &pAig.vGateClasses );
            if pAig.vObjClasses {
                pAig.vGateClasses = Gia_VtaConvertToGla( pAig, pAig.vObjClasses ); }
            Vec_IntFreeP( &pAig.vObjClasses );
            // return if VTA solve the problem if could not start
            if RetValue == 0 || pAig.vGateClasses == nil {
                return RetValue; }
        
        } else {
            pAig.vGateClasses = Vec_IntStart( Gia_ManObjNum(pAig) );
            Vec_IntWriteEntry( pAig.vGateClasses, 0, 1 );
            Vec_IntWriteEntry( pAig.vGateClasses, Gia_ObjFaninId0p(pAig, Gia_ManPo(pAig, 0)), 1 );
        }
    }
    // start the manager
    p = Gla_ManStart( pAig, pPars );
    p.timeInit = Abc_Clock() - clk;
    // set runtime limit
    if p.pPars.nTimeOut {
        sat_solver2_set_runtime_limit( p.pSat, p.pPars.nTimeOut * int64(time.Second) + Abc_Clock() ); }
    // perform initial abstraction
    if p.pPars.fVerbose {
        Abc_Print( 1, "Running gate-level abstraction (GLA) with the following parameters:\n" );
        Abc_Print( 1, "FrameMax = %d  ConfMax = %d  Timeout = %d  RatioMin = %d %%.\n", 
            pPars.nFramesMax, pPars.nConfLimit, pPars.nTimeOut, pPars.nRatioMin );
        Abc_Print( 1, "LearnStart = %d  LearnDelta = %d  LearnRatio = %d %%.\n", 
            pPars.nLearnedStart, pPars.nLearnedDelta, pPars.nLearnedPerce );
        Abc_Print( 1, " Frame   %%   Abs  PPI   FF   LUT   Confl  Cex   Vars   Clas   Lrns     Time        Mem\n" );
    }
    for  f = i = iPrev = 0; !p.pPars.nFramesMax || f < p.pPars.nFramesMax; f, iPrev = f+1, i  {
        var nConflsBeg int32  = sat_solver2_nconflicts(p.pSat);
        p.pPars.iFrame = f;

        // load timeframe
        Gia_GlaAddTimeFrame( p, f );

        // iterate as long as there are counter-examples
        for  i = 0; ; i++  { 
            clk2 = Abc_Clock();
            vCore = Gla_ManUnsatCore( p, f, p.pSat, pPars.nConfLimit, pPars.fVerbose, &Status, &nConfls );
//            assert( (vCore != NULL) == (Status == 1) );
            if Status == -1 || (p.pSat.nRuntimeLimit && Abc_Clock() > p.pSat.nRuntimeLimit) { // resource limit is reached
            
                Prf_ManStopP( &p.pSat.pPrf2 );
//                if ( Gia_ManRegNum(p->pGia) > 1 ) // for comb cases, return the abstraction
//                    Vec_IntShrink( p->vAbs, p->nAbsOld );
                goto finish;
            }
            if Status == 1 {
                Prf_ManStopP( &p.pSat.pPrf2 );
                p.timeUnsat += Abc_Clock() - clk2;
                break;
            } 
            p.timeSat += Abc_Clock() - clk2;
            assert(Status == 0, "Status == 0");
            p.nCexes++;

            // cancel old one if it was sent
            if Abc_FrameIsBridgeMode() && (fOneIsSent != 0) {
                Gia_GlaSendCancel( p, pPars.fVerbose );
                fOneIsSent = 0;
            }

            // perform the refinement
            clk2 = Abc_Clock();
            if pPars.fAddLayer {
                vPPis = Gla_ManCollectPPis( p, nil );
//                Gla_ManExplorePPis( p, vPPis );
            
            } else {
                vPPis = Gla_ManRefinement( p );
                if vPPis == nil {
                    Prf_ManStopP( &p.pSat.pPrf2 );
                    pCex = p.pGia.pCexSeq; p.pGia.pCexSeq = nil;
                    break;
                }
            } 
            assert(pCex == nil, "pCex == nil");

            // start proof logging
            if i == 0 {
                // create bookmark to be used for rollback
                sat_solver2_bookmark( p.pSat );
                Vec_IntClear( p.vAddedNew );
                p.nAbsOld = Vec_IntSize( p.vAbs );
                nVarsOld = p.nSatVars;
//                p->nLrnOld = sat_solver2_nlearnts( p->pSat );
//                p->nAbsNew = 0;
//                p->nLrnNew = 0;

                // start incremental proof manager
                assert(p.pSat.pPrf2 == nil, "p.pSat.pPrf2 == nil");
                if p.pSat.pPrf1 == nil {
                    p.pSat.pPrf2 = Prf_ManAlloc(); }
                if p.pSat.pPrf2 {
                    p.nProofIds = 0;
                    Vec_IntFill( p.vProofIds, Gia_ManObjNum(p.pGia), -1 );
                    Prf_ManRestart( p.pSat.pPrf2, p.vProofIds, sat_solver2_nlearnts(p.pSat), Vec_IntSize(vPPis) );
                }
            
            } else {
                // resize the proof logger
                if p.pSat.pPrf2 {
                    Prf_ManGrow( p.pSat.pPrf2, p.nProofIds + Vec_IntSize(vPPis) ); }
            }

            Gia_GlaAddToAbs( p, vPPis, 1 );
            Gia_GlaAddOneSlice( p, f, vPPis );
            Vec_IntFree( vPPis );

            // print the result (do not count it towards change)
            if p.pPars.fVerbose {
            Gla_ManAbsPrintFrame( p, -1, f+1, sat_solver2_nconflicts(p.pSat)-nConflsBeg, i, Abc_Clock() - clk ); }
        }
        if pCex != nil {
            break; }
        assert(Status == 1, "Status == 1");

        // valid core is obtained
        nCoreSize = 1;
        if vCore != nil {
            nCoreSize += Vec_IntSize( vCore );
            Gia_GlaAddToCounters( p, vCore );
        }
        if i == 0 {
            p.pPars.nFramesNoChange++;
            Vec_IntFreeP( &vCore );
        
        } else {
            p.pPars.nFramesNoChange = 0;
//            p->nAbsNew = Vec_IntSize( p->vAbs ) - p->nAbsOld;
//            p->nLrnNew = Abc_AbsInt( sat_solver2_nlearnts( p->pSat ) - p->nLrnOld );
            // update the SAT solver
            sat_solver2_rollback( p.pSat );
            // update storage
            Gla_ManRollBack( p );
            p.nSatVars = nVarsOld;
            // load this timeframe
            Gia_GlaAddToAbs( p, vCore, 0 );
            Gia_GlaAddOneSlice( p, f, vCore );
            Vec_IntFree( vCore );
            // run SAT solver
            clk2 = Abc_Clock();
            vCore = Gla_ManUnsatCore( p, f, p.pSat, pPars.nConfLimit, p.pPars.fVerbose, &Status, &nConfls );
            p.timeUnsat += Abc_Clock() - clk2;
//            assert( (vCore != NULL) == (Status == 1) );
            Vec_IntFreeP( &vCore );
            if Status == -1 { // resource limit is reached
                break; }
            if Status == 0 {
                assert(false, "0");
    //            Vta_ManSatVerify( p );
                // make sure, there was no initial abstraction (otherwise, it was invalid)
                assert(pAig.vObjClasses == nil && f < p.pPars.nFramesStart, "pAig.vObjClasses == nil && f < p.pPars.nFramesStart");
    //            pCex = Vga_ManDeriveCex( p );
                break;
            }
        }
        // print the result
        if p.pPars.fVerbose {
        Gla_ManAbsPrintFrame( p, nCoreSize, f+1, sat_solver2_nconflicts(p.pSat)-nConflsBeg, i, Abc_Clock() - clk ); }

        if f > 2 && iPrev > 0 && i == 0 { // change has happened
        
            if Abc_FrameIsBridgeMode() {
                // cancel old one if it was sent
                if fOneIsSent != 0 {
                    Gia_GlaSendCancel( p, pPars.fVerbose ); }
                // send new one 
                Gia_GlaSendAbsracted( p, pPars.fVerbose );
                fOneIsSent = 1;
            }

            // dump the model into file
            if p.pPars.fDumpVabs {
                var Command string // was [1000]byte in C

                Abc_FrameSetStatus( -1 );
                Abc_FrameSetCex( nil );
                Abc_FrameSetNFrames( f+1 );
                Command = fmt.Sprintf("write_status %s", Extra_FileNameGenericAppend((char *)(p.pPars.pFileVabs ? p.pPars.pFileVabs : "glabs.aig"), ".status"));
                Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
                Gia_GlaDumpAbsracted( p, pPars.fVerbose );
            }
        }

        // check if the number of objects is below limit
        if Gia_GlaAbsCount(p,0,0) >= (p.nObjs - 1) * (100 - pPars.nRatioMin) / 100 {
            Status = -1;
            break;
        }
    }
finish:
    // analize the results
    if pCex == nil {
        if p.pPars.fVerbose && Status == -1 {
            fmt.Printf("\n"); }
//        if ( pAig->vGateClasses != NULL )
//            Abc_Print( 1, "Replacing the old abstraction by a new one.\n" );
        Vec_IntFreeP( &pAig.vGateClasses );
        pAig.vGateClasses = Gla_ManTranslate( p );
        if Status == -1 {
            if p.pPars.nTimeOut && Abc_Clock() >= p.pSat.nRuntimeLimit { 
                Abc_Print( 1, "Timeout %d sec in frame %d with a %d-stable abstraction.    ", p.pPars.nTimeOut, f, p.pPars.nFramesNoChange );
            } else if pPars.nConfLimit && sat_solver2_nconflicts(p.pSat) >= pPars.nConfLimit {
                Abc_Print( 1, "Exceeded %d conflicts in frame %d with a %d-stable abstraction.  ", pPars.nConfLimit, f, p.pPars.nFramesNoChange );
            } else if Gia_GlaAbsCount(p,0,0) >= (p.nObjs - 1) * (100 - pPars.nRatioMin) / 100 {
                Abc_Print( 1, "The ratio of abstracted objects is less than %d %% in frame %d.  ", pPars.nRatioMin, f );
            } else {
                Abc_Print( 1, "Abstraction stopped for unknown reason in frame %d.  ", f ); }
        
        } else {
            p.pPars.iFrame++;
            Abc_Print( 1, "GLA completed %d frames with a %d-stable abstraction.  ", f, p.pPars.nFramesNoChange );
        }
    
    } else {
        if p.pPars.fVerbose {
            fmt.Printf("\n"); }
        pAig.pCexSeq = nil;
        pAig.pCexSeq = pCex;
        if !Gia_ManVerifyCex( pAig, pCex, 0 ) {
            Abc_Print( 1, "    Gia_ManPerformGlaOld(): CEX verification has failed!\n" ); }
        Abc_Print( 1, "Counter-example detected in frame %d.  ", f );
        p.pPars.iFrame = pCex.iFrame - 1;
        Vec_IntFreeP( &pAig.vGateClasses );
        RetValue = 0;
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if p.pPars.fVerbose {
        p.timeOther = (Abc_Clock() - clk) - p.timeUnsat - p.timeSat - p.timeCex - p.timeInit;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Initializing"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeInit)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeInit)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Solver UNSAT"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeUnsat)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeUnsat)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Solver SAT  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeSat)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Refinement  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCex)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeCex)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: Other       "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeOther)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Runtime: TOTAL       "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((Abc_Clock() - clk)))/(float64((Abc_Clock() - clk))) : 0.0);;
        Gla_ManReportMemory( p );
    }
//    Ga2_ManDumpStats( pAig, p->pPars, p->pSat, p->pPars->iFrame, 1 );
    Gla_ManStop( p );
    fflush( os.Stdout );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

