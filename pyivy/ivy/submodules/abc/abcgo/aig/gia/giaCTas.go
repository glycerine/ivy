package gia


import (
	"fmt"
	"time"
)

type Tas_Man_t = Tas_Man_t_

/**CFile****************************************************************

  FileName    [giaCSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [A simple circuit-based solver.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaCSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


//#define gia_assert(exp)     ((void)0)
//#define gia_assert(exp)     (assert(exp))

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Tas_Par_t struct {
    // conflict limits
    nBTLimit           int32;     // limit on the number of conflicts
    nJustLimit           int32;   // limit on the size of justification queue
    // current parameters
    nBTThis           int32;      // number of conflicts
    nBTThisNc           int32;    // number of conflicts
    nJustThis           int32;    // max size of the frontier
    nBTTotal           int32;     // total number of conflicts
    nJustTotal           int32;   // total size of the frontier
    // activity
    VarDecay         float32;     // variable activity decay
    VarInc           int32;       // variable increment
    // decision heuristics
    fUseActive           int32;   // use most active
    fUseHighest           int32;  // use node with the highest ID
    fUseLowest           int32;   // use node with the highest ID
    fUseMaxFF           int32;    // use node with the largest fanin fanout
    // other
    fVerbose           int32;
}

   
type Tas_Cls_t struct {
    iNext [2]int32;     // beginning of the queue
    nLits           int32;        // the number of literals
    pLits [0]int32;     // clause literals
}

   
type Tas_Sto_t struct {
    iCur           int32;         // current position
    nSize           int32;        // allocated size
    pData          *int32;        // clause information
}

   
type Tas_Que_t struct {
    iHead           int32;        // beginning of the queue
    iTail           int32;        // end of the queue
    nSize           int32;        // allocated size
    pData   **Gia_Obj_t;        // nodes stored in the queue
}
 
type Tas_Man_t_ struct {
    Pars     Tas_Par_t;         // parameters
    pAig    *Gia_Man_t;         // AIG manager
    pProp     Tas_Que_t;        // propagation queue
    pJust     Tas_Que_t;        // justification queue
    pClauses     Tas_Que_t;     // clause queue
    pIter   **Gia_Obj_t;        // iterator through clause vars
    vLevReas    *Vec_Int_t;     // levels and decisions
    vModel    *Vec_Int_t;       // satisfying assignment
    vTemp    *Vec_Ptr_t;        // temporary storage
    // watched clauses
    pStore     Tas_Sto_t;       // storage for watched clauses
    pWatches          *int32;     // watched lists for each literal
    vWatchLits    *Vec_Int_t;   // lits whose watched are assigned
    nClauses           int32;     // the counter of clauses
    // activity
    pActivity        *float32;    // variable activity
    vActiveVars    *Vec_Int_t;  // variables with activity
    // SAT calls statistics
    nSatUnsat           int32;    // the number of proofs
    nSatSat           int32;      // the number of failure
    nSatUndec           int32;    // the number of timeouts
    nSatTotal           int32;    // the number of calls
    // conflicts
    nConfUnsat           int32;   // conflicts in unsat problems
    nConfSat           int32;     // conflicts in sat problems
    nConfUndec           int32;   // conflicts in undec problems
    // runtime stats
    timeSatUnsat       abctime; // unsat
    timeSatSat       abctime;   // sat
    timeSatUndec       abctime; // undecided
    timeTotal       abctime;    // total runtime
}

func Tas_VarIsAssigned(pVar *Gia_Obj_t) int32 { return pVar.fMark0;                        }
func Tas_VarAssign(pVar *Gia_Obj_t) { assert(!pVar.fMark0, "!pVar.fMark0"); pVar.fMark0 = 1;    }
func Tas_VarUnassign(pVar *Gia_Obj_t) { assert(pVar.fMark0, "pVar.fMark0");  pVar.fMark0 = 0; pVar.fMark1 = 0; pVar.Value = ~0; }
func Tas_VarValue(pVar *Gia_Obj_t) int32 { assert(pVar.fMark0, "pVar.fMark0");  return pVar.fMark1; }
func Tas_VarSetValue(pVar *Gia_Obj_t, v int32) { assert(pVar.fMark0, "pVar.fMark0");  pVar.fMark1 = v;    }
func Tas_VarIsJust(pVar *Gia_Obj_t) int32 { return Gia_ObjIsAnd(pVar) && !Tas_VarIsAssigned(Gia_ObjFanin0(pVar)) && !Tas_VarIsAssigned(Gia_ObjFanin1(pVar)); } 
func Tas_VarFanin0Value(pVar *Gia_Obj_t) int32 { if !(Tas_VarIsAssigned(Gia_ObjFanin0(pVar)) != 0) {
    return 2;
} else {
    return (Tas_VarValue(Gia_ObjFanin0(pVar)) ^ Gia_ObjFaninC0(pVar));
} }
func Tas_VarFanin1Value(pVar *Gia_Obj_t) int32 { if !(Tas_VarIsAssigned(Gia_ObjFanin1(pVar)) != 0) {
    return 2;
} else {
    return (Tas_VarValue(Gia_ObjFanin1(pVar)) ^ Gia_ObjFaninC1(pVar));
} }
func Tas_VarToLit(p *Tas_Man_t, pObj *Gia_Obj_t) int32 { assert((Tas_VarIsAssigned(pObj) != 0), "Tas_VarIsAssigned(pObj)"); return Abc_Var2Lit( Gia_ObjId(p.pAig, pObj), !Tas_VarValue(pObj) );     }
func Tas_LitIsTrue(pObj *Gia_Obj_t, Lit int32) int32 { assert((Tas_VarIsAssigned(pObj) != 0), "Tas_VarIsAssigned(pObj)"); return Tas_VarValue(pObj) != Abc_LitIsCompl(Lit);                             }

func Tas_ClsHandle(p *Tas_Man_t, pClause *Tas_Cls_t) int32 { return ((int *)pClause) - p.pStore.pData;   }
func Tas_ClsFromHandle(p *Tas_Man_t, h int32) *Tas_Cls_t { return (Tas_Cls_t *)(p.pStore.pData + h);   }

func Tas_VarDecLevel(p *Tas_Man_t, pVar *Gia_Obj_t) int32 { assert(pVar.Value != ~0, "pVar.Value != ~0"); return Vec_IntEntry(p.vLevReas, 3*pVar.Value);          }
func Tas_VarReason0(p *Tas_Man_t, pVar *Gia_Obj_t) *Gia_Obj_t { assert(pVar.Value != ~0, "pVar.Value != ~0"); return pVar + Vec_IntEntry(p.vLevReas, 3*pVar.Value+1); }
func Tas_VarReason1(p *Tas_Man_t, pVar *Gia_Obj_t) *Gia_Obj_t { assert(pVar.Value != ~0, "pVar.Value != ~0"); return pVar + Vec_IntEntry(p.vLevReas, 3*pVar.Value+2); }
func Tas_ClauseDecLevel(p *Tas_Man_t, hClause int32) int32 { return Tas_VarDecLevel( p, p.pClauses.pData[hClause] );                               }

func Tas_VarHasReasonCls(p *Tas_Man_t, pVar *Gia_Obj_t) int32 { assert(pVar.Value != ~0, "pVar.Value != ~0"); return Vec_IntEntry(p.vLevReas, 3*pVar.Value+1) == 0 && Vec_IntEntry(p.vLevReas, 3*pVar.Value+2) != 0; }
func Tas_VarReasonCls(p *Tas_Man_t, pVar *Gia_Obj_t) *Tas_Cls_t { assert(pVar.Value != ~0, "pVar.Value != ~0"); return Tas_ClsFromHandle( p, Vec_IntEntry(p.vLevReas, 3*pVar.Value+2) );                                 }

// Original macro:
// #define Tas_QueForEachEntry( Que, pObj, i )                         \
//     for ( i = (Que).iHead; (i < (Que).iTail) && ((pObj) = (Que).pData[i]); i++ )


// Original macro:
// #define Tas_ClauseForEachVar( p, hClause, pObj )                    \
//     for ( (p)->pIter = (p)->pClauses.pData + hClause; (pObj = *pIter); (p)->pIter++ )

// Original macro:
// #define Tas_ClauseForEachVar1( p, hClause, pObj )                   \
//     for ( (p)->pIter = (p)->pClauses.pData+hClause+1; (pObj = *pIter); (p)->pIter++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets default values of the parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_SetDefaultParams(pPars *Tas_Par_t) {
    *pPars = Tas_Par_t{};
    pPars.nBTLimit    =        2000;   // limit on the number of conflicts
    pPars.nJustLimit  =        2000;   // limit on the size of justification queue
    pPars.fUseActive  =           0;   // use node with the highest activity
    pPars.fUseHighest =           1;   // use node with the highest ID
    pPars.fUseLowest  =           0;   // use node with the lowest ID
    pPars.fUseMaxFF   =           0;   // use node with the largest fanin fanout
    pPars.fVerbose    =           1;   // print detailed statistics
    pPars.VarDecay    = float32(0.95);   // variable decay
    pPars.VarInc      =         1.0;   // variable increment
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManAlloc(pAig *Gia_Man_t, nBTLimit int32) *Tas_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Tas_Man_t{};
    Tas_SetDefaultParams( &p.Pars );
    p.pAig           = pAig;
    p.Pars.nBTLimit  = nBTLimit;
    p.pProp.nSize    = p.pJust.nSize = p.pClauses.nSize = 10000;
    // inlining of macro ABC_ALLOC
    p.pProp.pData    = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.pProp.nSize)));
    // inlining of macro ABC_ALLOC
    p.pJust.pData    = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.pJust.nSize)));
    // inlining of macro ABC_ALLOC
    p.pClauses.pData = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.pClauses.nSize)));
    p.pClauses.iHead = p.pClauses.iTail = 1;
    p.vModel         = Vec_IntAlloc( 1000 );
    p.vLevReas       = Vec_IntAlloc( 1000 );
    p.vTemp          = Vec_PtrAlloc( 1000 );
    p.pStore.iCur    = 16;
    p.pStore.nSize   = 10000;
    // inlining of macro ABC_ALLOC
    p.pStore.pData   = ((int *) malloc(sizeof(int) * (size_t)(p.pStore.nSize)));
    // inlining of macro ABC_CALLOC
    p.pWatches       = ((int *) calloc((size_t)(2 * Gia_ManObjNum(pAig)), sizeof(int)));
    p.vWatchLits     = Vec_IntAlloc( 100 );
    // inlining of macro ABC_CALLOC
    p.pActivity      = ((float *) calloc((size_t)(Gia_ManObjNum(pAig)), sizeof(float)));
    p.vActiveVars    = Vec_IntAlloc( 100 );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManStop(p *Tas_Man_t) {
    Vec_IntFree( p.vActiveVars );
    Vec_IntFree( p.vWatchLits );
    Vec_IntFree( p.vLevReas );
    Vec_IntFree( p.vModel );
    Vec_PtrFree( p.vTemp );
    // inlining of macro ABC_FREE
    p.pActivity = nil;
    // inlining of macro ABC_FREE
    p.pWatches = nil;
    // inlining of macro ABC_FREE
    p.pStore.pData = nil;
    // inlining of macro ABC_FREE
    p.pClauses.pData = nil;
    // inlining of macro ABC_FREE
    p.pProp.pData = nil;
    // inlining of macro ABC_FREE
    p.pJust.pData = nil;
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Returns satisfying assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ReadModel(p *Tas_Man_t) *Vec_Int_t {
    return p.vModel;
}




/**Function*************************************************************

  Synopsis    [Returns 1 if the solver is out of limits.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManCheckLimits(p *Tas_Man_t) int32 {
    return p.Pars.nJustThis > p.Pars.nJustLimit || p.Pars.nBTThis > p.Pars.nBTLimit;
}

/**Function*************************************************************

  Synopsis    [Saves the satisfying assignment as an array of literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManSaveModel(p *Tas_Man_t, vCex *Vec_Int_t) {
    var pVar *Gia_Obj_t;
    var i int32;
    Vec_IntClear( vCex );
    p.pProp.iHead = 0;
//    printf( "\n" );
    // inlining of macro Tas_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
        if Gia_ObjIsCi(pVar) {
//            Vec_IntPush( vCex, Abc_Var2Lit(Gia_ObjId(p->pAig,pVar), !Tas_VarValue(pVar)) );
            Vec_IntPush( vCex, Abc_Var2Lit(Gia_ObjCioId(pVar), !Tas_VarValue(pVar)) ); }
/*
        printf( "%5d(%d) = ", Gia_ObjId(p->pAig, pVar), Tas_VarValue(pVar) );
        if ( Gia_ObjIsCi(pVar) )
            printf( "pi %d\n", Gia_ObjCioId(pVar) );
        else
        {
            printf( "%5d %d  &  ", Gia_ObjFaninId0p(p->pAig, pVar), Gia_ObjFaninC0(pVar) );
            printf( "%5d %d     ", Gia_ObjFaninId1p(p->pAig, pVar), Gia_ObjFaninC1(pVar) );
            printf( "\n" );
        }
*/
    }
} 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_QueIsEmpty(p *Tas_Que_t) int32 {
    return p.iHead == p.iTail;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_QuePush(p *Tas_Que_t, pObj *Gia_Obj_t) {
    if p.iTail == p.nSize {
        p.nSize *= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pData {
            p.pData = ((Gia_Obj_t * *) realloc((char *)(p.pData), sizeof(Gia_Obj_t *) * (size_t)(p.nSize)));
        } else {
            p.pData = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.nSize)));
        } 
    }
    p.pData[p.iTail++] = pObj;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the object in the queue.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_QueHasNode(p *Tas_Que_t, pObj *Gia_Obj_t) int32 {
    var pTemp *Gia_Obj_t;
    var i int32;
    // inlining of macro Tas_QueForEachEntry
    for  i = (*p).iHead; (i < (*p).iTail) && ((pTemp) = (*p).pData[i]); i++  {
        if pTemp == pObj {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_QueStore(p *Tas_Que_t, piHeadOld *int32, piTailOld *int32) {
    var i int32;
    *piHeadOld = p.iHead;
    *piTailOld = p.iTail;
    for  i = *piHeadOld; i < *piTailOld; i++  {
        Tas_QuePush( p, p.pData[i] );
    }
    p.iHead = *piTailOld;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_QueRestore(p *Tas_Que_t, iHeadOld int32, iTailOld int32) {
    p.iHead = iHeadOld;
    p.iTail = iTailOld;
}

/**Function*************************************************************

  Synopsis    [Finalized the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_QueFinish(p *Tas_Que_t) int32 {
    var iHeadOld int32  = p.iHead;
    assert(p.iHead < p.iTail, "p.iHead < p.iTail");
    Tas_QuePush( p, nil );
    p.iHead = p.iTail;
    return iHeadOld;
}


/**Function*************************************************************

  Synopsis    [Max number of fanins fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_VarFaninFanoutMax(p *Tas_Man_t, pObj *Gia_Obj_t) int32 {
    var Count0, Count1 int32;
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Count0 = Gia_ObjRefNum( p.pAig, Gia_ObjFanin0(pObj) );
    Count1 = Gia_ObjRefNum( p.pAig, Gia_ObjFanin1(pObj) );
    return Abc_MaxInt( Count0, Count1 );
}



/**Function*************************************************************

  Synopsis    [Find variable with the highest activity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManFindActive(p *Tas_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var BestCost float32  = 0.0;
    var i, ObjId int32;
    // inlining of macro Tas_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        ObjId = Gia_ObjId( p.pAig, pObj );
        if pObjMax == nil || 
             p.pActivity[Gia_ObjFaninId0(pObj,ObjId)] > BestCost || 
            (p.pActivity[Gia_ObjFaninId0(pObj,ObjId)] == BestCost && pObjMax < Gia_ObjFanin0(pObj)) {
            pObjMax  = Gia_ObjFanin0(pObj);
            BestCost = p.pActivity[Gia_ObjFaninId0(pObj,ObjId)];
        }
        if p.pActivity[Gia_ObjFaninId1(pObj,ObjId)] > BestCost || 
            (p.pActivity[Gia_ObjFaninId1(pObj,ObjId)] == BestCost && pObjMax < Gia_ObjFanin1(pObj)) {
            pObjMax  = Gia_ObjFanin1(pObj);
            BestCost = p.pActivity[Gia_ObjFaninId1(pObj,ObjId)];
        }
    }
    return pObjMax;
}

/**Function*************************************************************

  Synopsis    [Find variable with the highest activity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManDecideHighestFanin(p *Tas_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i, ObjId int32;
    // inlining of macro Tas_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        ObjId = Gia_ObjId( p.pAig, pObj );
        if pObjMax == nil || pObjMax < Gia_ObjFanin0(pObj) {
            pObjMax  = Gia_ObjFanin0(pObj); }
        if pObjMax < Gia_ObjFanin1(pObj) {
            pObjMax  = Gia_ObjFanin1(pObj); }
    }
    return pObjMax;
}

/**Function*************************************************************

  Synopsis    [Find variable with the highest ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManDecideHighest(p *Tas_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i int32;
    // inlining of macro Tas_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
//printf( "%d %6.2f   ", Gia_ObjId(p->pAig, pObj), p->pActivity[Gia_ObjId(p->pAig, pObj)] );
        if pObjMax == nil || pObjMax < pObj {
            pObjMax = pObj; }
    }
//printf( "\n" );
    return pObjMax;
}

/**Function*************************************************************

  Synopsis    [Find variable with the highest ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManDecideHighestA(p *Tas_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i int32;
    // inlining of macro Tas_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
        if pObjMax == nil || 
             p.pActivity[Gia_ObjId(p.pAig, pObjMax)] < p.pActivity[Gia_ObjId(p.pAig, pObj)] {
            pObjMax = pObj; }
    }
    return pObjMax;
}

/**Function*************************************************************

  Synopsis    [Find variable with the lowest ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManDecideLowest(p *Tas_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMin *Gia_Obj_t = nil;
    var i int32;
    // inlining of macro Tas_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
        if pObjMin == nil || pObjMin > pObj {
            pObjMin = pObj; }
    }
    return pObjMin;
}

/**Function*************************************************************

  Synopsis    [Find variable with the maximum number of fanin fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManDecideMaxFF(p *Tas_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i, iMaxFF, iCurFF int32;
    assert(p.pAig.Refs != nil, "p.pAig.Refs != nil");
    // inlining of macro Tas_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
        iCurFF = Tas_VarFaninFanoutMax( p, pObj );
        assert(iCurFF > 0, "iCurFF > 0");
        if iMaxFF < iCurFF {
            iMaxFF = iCurFF;
            pObjMax = pObj;
        }
    }
    return pObjMax; 
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManCancelUntil(p *Tas_Man_t, iBound int32) {
    var pVar *Gia_Obj_t;
    var i int32;
    assert(iBound <= p.pProp.iTail, "iBound <= p.pProp.iTail");
    p.pProp.iHead = iBound;
    // inlining of macro Tas_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
        Tas_VarUnassign( pVar );
    }
    p.pProp.iTail = iBound;
    Vec_IntShrink( p.vLevReas, 3*iBound );
}

var s_Counter2 int32;
var s_Counter3 int32;
var s_Counter4 int32;

/**Function*************************************************************

  Synopsis    [Assigns the variables a value.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManAssign(p *Tas_Man_t, pObj *Gia_Obj_t, Level int32, pRes0 *Gia_Obj_t, pRes1 *Gia_Obj_t) {
    var pObjR *Gia_Obj_t  = Gia_Regular(pObj);
    assert(Gia_ObjIsCand(pObjR), "Gia_ObjIsCand(pObjR)");
    assert(!(Tas_VarIsAssigned(pObjR) != 0), "!Tas_VarIsAssigned(pObjR)");
    Tas_VarAssign( pObjR );
    Tas_VarSetValue( pObjR, !Gia_IsComplement(pObj) );
    assert(pObjR.Value == ~0, "pObjR.Value == ~0");
    pObjR.Value = p.pProp.iTail;
    Tas_QuePush( &p.pProp, pObjR );
    Vec_IntPush( p.vLevReas, Level );
    if pRes0 == nil && pRes1 != 0 { // clause
    
        Vec_IntPush( p.vLevReas, 0 );
        Vec_IntPush( p.vLevReas, Tas_ClsHandle( p, (Tas_Cls_t *)pRes1 ) );
    
    } else {
        Vec_IntPush( p.vLevReas, pRes0 ? pRes0-pObjR : 0 );
        Vec_IntPush( p.vLevReas, pRes1 ? pRes1-pObjR : 0 );
    }
    assert(Vec_IntSize(p.vLevReas) == 3 * p.pProp.iTail, "Vec_IntSize(p.vLevReas) == 3 * p.pProp.iTail");
    s_Counter2++;
}


/**Function*************************************************************

  Synopsis    [Returns clause size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManClauseSize(p *Tas_Man_t, hClause int32) int32 {
    var pQue *Tas_Que_t  = &(p.pClauses);
    var pIter **Gia_Obj_t;
    for  pIter = pQue.pData + hClause; *pIter; pIter++  {}
    return pIter - pQue.pData - hClause ;
}

/**Function*************************************************************

  Synopsis    [Prints conflict clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManPrintClause(p *Tas_Man_t, Level int32, hClause int32) {
    var pQue *Tas_Que_t  = &(p.pClauses);
    var pObj *Gia_Obj_t;
    var i int32;
    assert((Tas_QueIsEmpty( pQue ) != 0), "Tas_QueIsEmpty( pQue )");
    fmt.Printf("Level %2d : ", Level);
    for  i = hClause; (pObj = pQue.pData[i]); i++  {
        fmt.Printf("%d=%d(%d) ", Gia_ObjId(p.pAig, pObj), Tas_VarValue(pObj), Tas_VarDecLevel(p, pObj));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Prints conflict clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManPrintClauseNew(p *Tas_Man_t, Level int32, hClause int32) {
    var pQue *Tas_Que_t  = &(p.pClauses);
    var pObj *Gia_Obj_t;
    var i int32;
    assert((Tas_QueIsEmpty( pQue ) != 0), "Tas_QueIsEmpty( pQue )");
    fmt.Printf("Level %2d : ", Level);
    for  i = hClause; (pObj = pQue.pData[i]); i++  {
        fmt.Printf("%c%d ", Tas_VarValue(pObj)? '+':'-', Gia_ObjId(p.pAig, pObj));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Returns conflict clause.]

  Description [Performs conflict analysis.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManDeriveReason(p *Tas_Man_t, Level int32) {
    var pQue *Tas_Que_t  = &(p.pClauses);
    var pObj, pReason *Gia_Obj_t;
    var i, k, j, iLitLevel, iLitLevel2 int32;//, Id;
    assert(pQue.pData[pQue.iHead] == nil, "pQue.pData[pQue.iHead] == nil");
    assert(pQue.iHead + 1 < pQue.iTail, "pQue.iHead + 1 < pQue.iTail");
/*
    for ( i = pQue->iHead + 1; i < pQue->iTail; i++ )
    {
        pObj = pQue->pData[i];
        assert( pObj->fPhase == 0 );
    }
*/
    // compact literals
    Vec_PtrClear( p.vTemp );
    for  i = k = pQue.iHead + 1; i < pQue.iTail; i++  {
        pObj = pQue.pData[i];
        if pObj.fPhase { // unassigned - seen again
            continue; }
        // assigned - seen first time
        pObj.fPhase = 1;
        Vec_PtrPush( p.vTemp, pObj );
        // bump activity
//        Id = Gia_ObjId( p->pAig, pObj );
//        if ( p->pActivity[Id] == 0.0 )
//            Vec_IntPush( p->vActiveVars, Id );
//        p->pActivity[Id] += p->Pars.VarInc;
        // check decision level
        iLitLevel = Tas_VarDecLevel( p, pObj );
        if iLitLevel < Level {
            pQue.pData[k++] = pObj;
            continue;
        }
        assert(iLitLevel == Level, "iLitLevel == Level");
        if Tas_VarHasReasonCls( p, pObj ) != 0 {
            var pCls *Tas_Cls_t  = Tas_VarReasonCls( p, pObj );
            pReason = Gia_ManObj( p.pAig, Abc_Lit2Var(pCls.pLits[0]) );
            assert(pReason == pObj, "pReason == pObj");
            for  j = 1; j < pCls.nLits; j++  {
                pReason = Gia_ManObj( p.pAig, Abc_Lit2Var(pCls.pLits[j]) );
                iLitLevel2 = Tas_VarDecLevel( p, pReason );
                assert((Tas_VarIsAssigned( pReason ) != 0), "Tas_VarIsAssigned( pReason )");
                assert(!(Tas_LitIsTrue( pReason, pCls.pLits[j] ) != 0), "!Tas_LitIsTrue( pReason, pCls.pLits[j] )");
                Tas_QuePush( pQue, pReason );
            }
        
        } else {
            pReason = Tas_VarReason0( p, pObj );
            if pReason == pObj { // no reason
            
                assert(pQue.pData[pQue.iHead] == nil || Level == 0, "pQue.pData[pQue.iHead] == nil || Level == 0");
                if pQue.pData[pQue.iHead] == nil {
                    pQue.pData[pQue.iHead] = pObj;
                } else {
                    Tas_QuePush( pQue, pObj ); }
                continue;
            }
            Tas_QuePush( pQue, pReason );
            pReason = Tas_VarReason1( p, pObj );
            if pReason != pObj { // second reason
                Tas_QuePush( pQue, pReason ); }
        }
    }
    assert(pQue.pData[pQue.iHead] != nil, "pQue.pData[pQue.iHead] != nil");
    if pQue.pData[pQue.iHead] == nil {
        fmt.Printf("Tas_ManDeriveReason(): Failed to derive the clause!!!\n"); }
    pQue.iTail = k;
    // clear the marks
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTemp)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(p.vTemp, i)), 1); i++  {
        pObj.fPhase = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Returns conflict clause.]

  Description [Performs conflict analysis.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManAnalyze(p *Tas_Man_t, Level int32, pVar *Gia_Obj_t, pFan0 *Gia_Obj_t, pFan1 *Gia_Obj_t) int32 {
    var pQue *Tas_Que_t  = &(p.pClauses);
    assert((Tas_VarIsAssigned(pVar) != 0), "Tas_VarIsAssigned(pVar)");
    assert((Tas_VarIsAssigned(pFan0) != 0), "Tas_VarIsAssigned(pFan0)");
    assert(pFan1 == nil || (Tas_VarIsAssigned(pFan1) != 0), "pFan1 == nil || Tas_VarIsAssigned(pFan1)");
    assert((Tas_QueIsEmpty( pQue ) != 0), "Tas_QueIsEmpty( pQue )");
    Tas_QuePush( pQue, nil );
    Tas_QuePush( pQue, pVar );
    Tas_QuePush( pQue, pFan0 );
    if pFan1 != nil {
        Tas_QuePush( pQue, pFan1 ); }
    Tas_ManDeriveReason( p, Level );
    return Tas_QueFinish( pQue );
}


/**Function*************************************************************

  Synopsis    [Performs resolution of two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManResolve(p *Tas_Man_t, Level int32, hClause0 int32, hClause1 int32) int32 {
    var pQue *Tas_Que_t  = &(p.pClauses);
    var pObj *Gia_Obj_t;
    var i int32;
    var LevelMax int32 = -1;
    var LevelCur int32;
    assert(pQue.pData[hClause0] != nil, "pQue.pData[hClause0] != nil");
    assert(pQue.pData[hClause0] == pQue.pData[hClause1], "pQue.pData[hClause0] == pQue.pData[hClause1]");
/*
    for ( i = hClause0 + 1; (pObj = pQue->pData[i]); i++ )
        assert( pObj->fPhase == 0 );
    for ( i = hClause1 + 1; (pObj = pQue->pData[i]); i++ )
        assert( pObj->fPhase == 0 );
*/
    assert((Tas_QueIsEmpty( pQue ) != 0), "Tas_QueIsEmpty( pQue )");
    Tas_QuePush( pQue, nil );
    for  i = hClause0 + 1; (pObj = pQue.pData[i]); i++  {
        if pObj.fPhase { // unassigned - seen again
            continue; }
        // assigned - seen first time
        pObj.fPhase = 1;
        Tas_QuePush( pQue, pObj );
        LevelCur = Tas_VarDecLevel( p, pObj );
        if LevelMax < LevelCur {
            LevelMax = LevelCur; }
    }
    for  i = hClause1 + 1; (pObj = pQue.pData[i]); i++  {
        if pObj.fPhase { // unassigned - seen again
            continue; }
        // assigned - seen first time
        pObj.fPhase = 1;
        Tas_QuePush( pQue, pObj );
        LevelCur = Tas_VarDecLevel( p, pObj );
        if LevelMax < LevelCur {
            LevelMax = LevelCur; }
    }
    for  i = pQue.iHead + 1; i < pQue.iTail; i++  {
        pQue.pData[i].fPhase = 0;
    }
    Tas_ManDeriveReason( p, LevelMax );
    return Tas_QueFinish( pQue );
}



/**Function*************************************************************

  Synopsis    [Allocates clause of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManAllocCls(p *Tas_Man_t, nSize int32) *Tas_Cls_t {
    var pCls *Tas_Cls_t;
    if p.pStore.iCur + nSize > p.pStore.nSize {
        p.pStore.nSize *= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pStore.pData {
            p.pStore.pData = ((int *) realloc((char *)(p.pStore.pData), sizeof(int) * (size_t)(p.pStore.nSize)));
        } else {
            p.pStore.pData = ((int *) malloc(sizeof(int) * (size_t)(p.pStore.nSize)));
        } 
    }
    pCls = Tas_ClsFromHandle( p, p.pStore.iCur ); p.pStore.iCur += nSize;
    memset( pCls, 0, sizeof(int) * nSize );
    p.nClauses++;
    return pCls;
}

/**Function*************************************************************

  Synopsis    [Adds one clause to the watcher list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManWatchClause(p *Tas_Man_t, pClause *Tas_Cls_t, Lit int32) {
    assert(Abc_Lit2Var(Lit) < Gia_ManObjNum(p.pAig), "Abc_Lit2Var(Lit) < Gia_ManObjNum(p.pAig)");
    assert(pClause.nLits >= 2, "pClause.nLits >= 2");
    assert(pClause.pLits[0] == Lit || pClause.pLits[1] == Lit, "pClause.pLits[0] == Lit || pClause.pLits[1] == Lit");
    if pClause.pLits[0] == Lit {
        pClause.iNext[0] = p.pWatches[Abc_LitNot(Lit)];  
    } else {
        pClause.iNext[1] = p.pWatches[Abc_LitNot(Lit)]; }  
    if p.pWatches[Abc_LitNot(Lit)] == 0 {
        Vec_IntPush( p.vWatchLits, Abc_LitNot(Lit) ); }
    p.pWatches[Abc_LitNot(Lit)] = Tas_ClsHandle( p, pClause );
}

/**Function*************************************************************

  Synopsis    [Creates clause of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManCreateCls(p *Tas_Man_t, hClause int32) *Tas_Cls_t {
    var pClause *Tas_Cls_t;
    var pQue *Tas_Que_t  = &(p.pClauses);
    var pObj *Gia_Obj_t;
    var i, nLits int32;
    assert((Tas_QueIsEmpty( pQue ) != 0), "Tas_QueIsEmpty( pQue )");
    assert(pQue.pData[hClause] != nil, "pQue.pData[hClause] != nil");
    for  i = hClause; (pObj = pQue.pData[i]); i++  {
        nLits++;
    }
    if nLits == 1 {
        return nil; }
    // create this clause
    pClause = Tas_ManAllocCls( p, nLits + 3 );
    pClause.nLits = nLits;
    for  i = hClause; (pObj = pQue.pData[i]); i++  {
        assert((Tas_VarIsAssigned( pObj ) != 0), "Tas_VarIsAssigned( pObj )");
        pClause.pLits[i-hClause] = Abc_LitNot( Tas_VarToLit(p, pObj) );
    }
    // add the clause as watched one
    if nLits >= 2 {
        Tas_ManWatchClause( p, pClause, pClause.pLits[0] );
        Tas_ManWatchClause( p, pClause, pClause.pLits[1] );
    }
    // increment activity
//    p->Pars.VarInc /= p->Pars.VarDecay;
    return pClause;
}

/**Function*************************************************************

  Synopsis    [Creates clause of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManCreateFromCls(p *Tas_Man_t, pCls *Tas_Cls_t, Level int32) int32 {
    var pQue *Tas_Que_t  = &(p.pClauses);
    var pObj *Gia_Obj_t;
    var i int32;
    assert((Tas_QueIsEmpty( pQue ) != 0), "Tas_QueIsEmpty( pQue )");
    Tas_QuePush( pQue, nil );
    for  i = 0; i < pCls.nLits; i++  {
        pObj = Gia_ManObj( p.pAig, Abc_Lit2Var(pCls.pLits[i]) );
        assert((Tas_VarIsAssigned(pObj) != 0), "Tas_VarIsAssigned(pObj)");
        assert(!(Tas_LitIsTrue( pObj, pCls.pLits[i] ) != 0), "!Tas_LitIsTrue( pObj, pCls.pLits[i] )");
        Tas_QuePush( pQue, pObj );
    }
    Tas_ManDeriveReason( p, Level );
    return Tas_QueFinish( pQue );
}

/**Function*************************************************************

  Synopsis    [Propagate one assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManPropagateWatch(p *Tas_Man_t, Level int32, Lit int32) int32 {
    var pObj *Gia_Obj_t;
    var pCur *Tas_Cls_t;
    var piPrev *int32;
    var iCur int32;
    var iTemp int32;
    var i int32;
    var LitF int32 = Abc_LitNot(Lit);
    // iterate through the clauses
    piPrev = p.pWatches + Lit;
    for  iCur = p.pWatches[Lit]; (iCur != 0); iCur = *piPrev  {
        pCur = Tas_ClsFromHandle( p, iCur );
        // make sure the false literal is in the second literal of the clause
        if pCur.pLits[0] == LitF {
            pCur.pLits[0] = pCur.pLits[1];
            pCur.pLits[1] = LitF;
            iTemp = pCur.iNext[0];
            pCur.iNext[0] = pCur.iNext[1];
            pCur.iNext[1] = iTemp;
        }
        assert(pCur.pLits[1] == LitF, "pCur.pLits[1] == LitF");

        // if the first literal is true, the clause is satisfied
//        if ( pCur->pLits[0] == p->pAssigns[Abc_Lit2Var(pCur->pLits[0])] )
        pObj = Gia_ManObj( p.pAig, Abc_Lit2Var(pCur.pLits[0]) );
        if (Tas_VarIsAssigned(pObj) != 0) && (Tas_LitIsTrue( pObj, pCur.pLits[0] ) != 0) {
            piPrev = &pCur.iNext[1];
            continue;
        }

        // look for a new literal to watch
        for  i = 2; i < int32(pCur.nLits); i++  {
            // skip the case when the literal is false
//            if ( Abc_LitNot(pCur->pLits[i]) == p->pAssigns[Abc_Lit2Var(pCur->pLits[i])] )
            pObj = Gia_ManObj( p.pAig, Abc_Lit2Var(pCur.pLits[i]) );
            if (Tas_VarIsAssigned(pObj) != 0) && !(Tas_LitIsTrue( pObj, pCur.pLits[i] ) != 0) {
                continue; }
            // the literal is either true or unassigned - watch it
            pCur.pLits[1] = pCur.pLits[i];
            pCur.pLits[i] = LitF;
            // remove this clause from the watch list of Lit
            *piPrev = pCur.iNext[1];
            // add this clause to the watch list of pCur->pLits[i] (now it is pCur->pLits[1])
            Tas_ManWatchClause( p, pCur, pCur.pLits[1] );
            break;
        }
        if i < int32(pCur.nLits) { // found new watch
            continue; }

        // clause is unit - enqueue new implication
        pObj = Gia_ManObj( p.pAig, Abc_Lit2Var(pCur.pLits[0]) );
        if !(Tas_VarIsAssigned(pObj) != 0) {
/*
            {
                int iLitLevel, iPlace;
                for ( i = 1; i < (int)pCur->nLits; i++ )
                {
                    pObj = Gia_ManObj( p->pAig, Abc_Lit2Var(pCur->pLits[i]) );
                    iLitLevel = Tas_VarDecLevel( p, pObj );
                    iPlace = pObj->Value;
                    printf( "Lit = %d. Level = %d. Place = %d.\n", pCur->pLits[i], iLitLevel, iPlace );
                    i = i;
                }
            }
*/
            Tas_ManAssign( p, Gia_ObjFromLit(p.pAig, pCur.pLits[0]), Level, nil, (Gia_Obj_t *)pCur );
            piPrev = &pCur.iNext[1];
            continue;
        }
        // conflict detected - return the conflict clause
        assert(!(Tas_LitIsTrue( pObj, pCur.pLits[0] ) != 0), "!Tas_LitIsTrue( pObj, pCur.pLits[0] )");
        return Tas_ManCreateFromCls( p, pCur, Level );
    }
    return 0;
}



/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns clause handle if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManPropagateOne(p *Tas_Man_t, pVar *Gia_Obj_t, Level int32) int32 {
    var Value0, Value1, hClause int32;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert((Tas_VarIsAssigned(pVar) != 0), "Tas_VarIsAssigned(pVar)");
    s_Counter3++;
    if hClause = Tas_ManPropagateWatch( p, Level, Tas_VarToLit(p, pVar) ) {
        return hClause; }
    if Gia_ObjIsCi(pVar) {
        return 0; }
/*
    if ( pVar->iDiff0 == 570869 && pVar->iDiff1 == 546821 && Level == 3 )
    {
        Gia_Obj_t * pFan0 = Gia_ObjFanin0(pVar);
        Gia_Obj_t * pFan1 = Gia_ObjFanin1(pVar);
        int s = 0;
    }
*/
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    Value0 = Tas_VarFanin0Value(pVar);
    Value1 = Tas_VarFanin1Value(pVar);
    if Tas_VarValue(pVar) != 0 { // value is 1
        if Value0 == 0 || Value1 == 0 { // one is 0
        
            if Value0 == 0 && Value1 != 0 {
                return Tas_ManAnalyze( p, Level, pVar, Gia_ObjFanin0(pVar), nil ); }
            if Value0 != 0 && Value1 == 0 {
                return Tas_ManAnalyze( p, Level, pVar, Gia_ObjFanin1(pVar), nil ); }
            assert(Value0 == 0 && Value1 == 0, "Value0 == 0 && Value1 == 0");
            return Tas_ManAnalyze( p, Level, pVar, Gia_ObjFanin0(pVar), Gia_ObjFanin1(pVar) );
        }
        if Value0 == 2 { // first is unassigned
            Tas_ManAssign( p, Gia_ObjChild0(pVar), Level, pVar, nil ); }
        if Value1 == 2 { // first is unassigned
            Tas_ManAssign( p, Gia_ObjChild1(pVar), Level, pVar, nil ); }
        return 0;
    }
    // value is 0
    if Value0 == 0 || Value1 == 0 { // one is 0
        return 0; }
    if Value0 == 1 && Value1 == 1 { // both are 1
        return Tas_ManAnalyze( p, Level, pVar, Gia_ObjFanin0(pVar), Gia_ObjFanin1(pVar) ); }
    if Value0 == 1 || Value1 == 1 { // one is 1 
    
        if Value0 == 2 { // first is unassigned
            Tas_ManAssign( p, Gia_Not(Gia_ObjChild0(pVar)), Level, pVar, Gia_ObjFanin1(pVar) ); }
        if Value1 == 2 { // second is unassigned
            Tas_ManAssign( p, Gia_Not(Gia_ObjChild1(pVar)), Level, pVar, Gia_ObjFanin0(pVar) ); }
        return 0;
    }
    assert((Tas_VarIsJust(pVar) != 0), "Tas_VarIsJust(pVar)");
    assert(!(Tas_QueHasNode( &p.pJust, pVar ) != 0), "!Tas_QueHasNode( &p.pJust, pVar )");
    Tas_QuePush( &p.pJust, pVar );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManPropagateTwo(p *Tas_Man_t, pVar *Gia_Obj_t, Level int32) int32 {
    var Value0, Value1 int32;
    s_Counter4++;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    assert((Tas_VarIsAssigned(pVar) != 0), "Tas_VarIsAssigned(pVar)");
    assert(!(Tas_VarValue(pVar) != 0), "!Tas_VarValue(pVar)");
    Value0 = Tas_VarFanin0Value(pVar);
    Value1 = Tas_VarFanin1Value(pVar);
    // value is 0
    if Value0 == 0 || Value1 == 0 { // one is 0
        return 0; }
    if Value0 == 1 && Value1 == 1 { // both are 1
        return Tas_ManAnalyze( p, Level, pVar, Gia_ObjFanin0(pVar), Gia_ObjFanin1(pVar) ); }
    assert(Value0 == 1 || Value1 == 1, "Value0 == 1 || Value1 == 1");
    if Value0 == 2 { // first is unassigned
        Tas_ManAssign( p, Gia_Not(Gia_ObjChild0(pVar)), Level, pVar, Gia_ObjFanin1(pVar) ); }
    if Value1 == 2 { // first is unassigned
        Tas_ManAssign( p, Gia_Not(Gia_ObjChild1(pVar)), Level, pVar, Gia_ObjFanin0(pVar) ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Propagates all variables.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManPropagate(p *Tas_Man_t, Level int32) int32 {
    var hClause int32;
    var pVar *Gia_Obj_t;
    var i, k int32;//, nIter = 0;
    for {
//        nIter++;
        // inlining of macro Tas_QueForEachEntry
        for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
            if hClause = Tas_ManPropagateOne( p, pVar, Level ) {
                return hClause; }
        }
        p.pProp.iHead = p.pProp.iTail;
        k = p.pJust.iHead;
        // inlining of macro Tas_QueForEachEntry
        for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pVar) = (p.pJust).pData[i]); i++  {
            if Tas_VarIsJust( pVar ) != 0 {
                p.pJust.pData[k++] = pVar;
            } else if hClause = Tas_ManPropagateTwo( p, pVar, Level ) {
                return hClause; }
        }
        if k == p.pJust.iTail {
            break; }
        p.pJust.iTail = k;
    }
//    printf( "%d ", nIter );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Solve the problem recursively.]

  Description [Returns learnt clause if unsat, NULL if sat or undecided.]
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Tas_ManSolve_rec(p *Tas_Man_t, Level int32) int32 { 
    var pQue *Tas_Que_t  = &(p.pClauses);
    var pVar *Gia_Obj_t = nil;
    var pDecVar *Gia_Obj_t = nil;
    var hClause, hLearn0, hLearn1 int32;
    var iPropHead, iJustHead, iJustTail int32;
    // propagate assignments
    assert(!(Tas_QueIsEmpty(&p.pProp) != 0), "!Tas_QueIsEmpty(&p.pProp)");
    if hClause = Tas_ManPropagate( p, Level ) {
        Tas_ManCreateCls( p, hClause );
        return hClause;
    }
    // check for satisfying assignment
    assert((Tas_QueIsEmpty(&p.pProp) != 0), "Tas_QueIsEmpty(&p.pProp)");
    if Tas_QueIsEmpty(&p.pJust) != 0 {
        return 0; }
    // quit using resource limits
    p.Pars.nJustThis = Abc_MaxInt( p.Pars.nJustThis, p.pJust.iTail - p.pJust.iHead );
    if Tas_ManCheckLimits( p ) != 0 {
        return 0; }
    // remember the state before branching
    iPropHead = p.pProp.iHead;
    Tas_QueStore( &p.pJust, &iJustHead, &iJustTail );
    // find the decision variable
    if p.Pars.fUseActive {
        pVar = nil;
pDecVar = Tas_ManFindActive( p );
    } else if p.Pars.fUseHighest {
//        pVar = NULL, pDecVar = Tas_ManDecideHighestFanin( p );
        pVar = Tas_ManDecideHighest( p );
    } else if p.Pars.fUseLowest {
        pVar = Tas_ManDecideLowest( p );
    } else if p.Pars.fUseMaxFF {
        pVar = Tas_ManDecideMaxFF( p );
    } else { assert(false, "0"); }
    // chose decision variable using fanout count
    if pVar != nil {
        assert((Tas_VarIsJust( pVar ) != 0), "Tas_VarIsJust( pVar )");
        if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(pVar)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(pVar)) {
            pDecVar = Gia_Not(Gia_ObjChild0(pVar));
        } else {
            pDecVar = Gia_Not(Gia_ObjChild1(pVar)); }
//        pDecVar = Gia_NotCond( pDecVar, Gia_Regular(pDecVar)->fMark1 ^ !Gia_IsComplement(pDecVar) );
    }
    // decide on first fanin
    Tas_ManAssign( p, pDecVar, Level+1, nil, nil );
    if !(hLearn0 = Tas_ManSolve_rec( p, Level+1 )) {
        return 0; }
    if pQue.pData[hLearn0] != Gia_Regular(pDecVar) {
        return hLearn0; }
    Tas_ManCancelUntil( p, iPropHead );
    Tas_QueRestore( &p.pJust, iJustHead, iJustTail );
    // decide on second fanin
    Tas_ManAssign( p, Gia_Not(pDecVar), Level+1, nil, nil );
    if !(hLearn1 = Tas_ManSolve_rec( p, Level+1 )) {
        return 0; }
    if pQue.pData[hLearn1] != Gia_Regular(pDecVar) {
        return hLearn1; }
    hClause = Tas_ManResolve( p, Level, hLearn0, hLearn1 );
    Tas_ManCreateCls( p, hClause );
//    Tas_ManPrintClauseNew( p, Level, hClause );
//    if ( Level > Tas_ClauseDecLevel(p, hClause) )
//        p->Pars.nBTThisNc++;
    p.Pars.nBTThis++;
    return hClause;
}

/**Function*************************************************************

  Synopsis    [Looking for a satisfying assignment of the node.]

  Description [Assumes that each node has flag pObj->fMark0 set to 0.
  Returns 1 if unsatisfiable, 0 if satisfiable, and -1 if undecided.
  The node may be complemented. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManSolve(p *Tas_Man_t, pObj *Gia_Obj_t, pObj2 *Gia_Obj_t) int32 {
    var i, Entry, RetValue int32;
    s_Counter2 = 0;
    Vec_IntClear( p.vModel );
    if pObj == Gia_ManConst0(p.pAig) || pObj2 == Gia_ManConst0(p.pAig) || pObj == Gia_Not(pObj2) {
        return 1; }
    if pObj == Gia_ManConst1(p.pAig) && (pObj2 == nil || pObj2 == Gia_ManConst1(p.pAig)) {
        return 0; }
    assert(!p.pProp.iHead && !p.pProp.iTail, "!p.pProp.iHead && !p.pProp.iTail");
    assert(!p.pJust.iHead && !p.pJust.iTail, "!p.pJust.iHead && !p.pJust.iTail");
    assert(p.pClauses.iHead == 1 && p.pClauses.iTail == 1, "p.pClauses.iHead == 1 && p.pClauses.iTail == 1");
    p.Pars.nBTThis = p.Pars.nJustThis = p.Pars.nBTThisNc = 0;
    Tas_ManAssign( p, pObj, 0, nil, nil );
    if (pObj2 != nil) && !(Tas_VarIsAssigned(Gia_Regular(pObj2)) != 0) {
        Tas_ManAssign( p, pObj2, 0, nil, nil ); }
    if !(Tas_ManSolve_rec(p, 0) != 0) && !(Tas_ManCheckLimits(p) != 0) {
        Tas_ManSaveModel( p, p.vModel );
    } else {
        RetValue = 1; }
    Tas_ManCancelUntil( p, 0 );
    p.pJust.iHead = p.pJust.iTail = 0;
    p.pClauses.iHead = p.pClauses.iTail = 1;
    // clauses
    if p.nClauses > 0 {
        p.pStore.iCur = 16;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vWatchLits)) && (((Entry) = Vec_IntEntry(p.vWatchLits, i)), 1); i++  {
            p.pWatches[Entry] = 0;
        }
        Vec_IntClear( p.vWatchLits );
        p.nClauses = 0;
    }
    // activity
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vActiveVars)) && (((Entry) = Vec_IntEntry(p.vActiveVars, i)), 1); i++  {
        p.pActivity[Entry] = 0.0;
    }
    Vec_IntClear( p.vActiveVars );
    // statistics
    p.Pars.nBTTotal += p.Pars.nBTThis;
    p.Pars.nJustTotal = Abc_MaxInt( p.Pars.nJustTotal, p.Pars.nJustThis );
    if Tas_ManCheckLimits( p ) != 0 {
        RetValue = -1; }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Looking for a satisfying assignment of the node.]

  Description [Assumes that each node has flag pObj->fMark0 set to 0.
  Returns 1 if unsatisfiable, 0 if satisfiable, and -1 if undecided.
  The node may be complemented. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManSolveArray(p *Tas_Man_t, vObjs *Vec_Ptr_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Entry, RetValue int32;
    s_Counter2 = 0;
    s_Counter3 = 0;
    s_Counter4 = 0;
    Vec_IntClear( p.vModel );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        if pObj == Gia_ManConst0(p.pAig) {
            return 1; }
    }
    assert(!p.pProp.iHead && !p.pProp.iTail, "!p.pProp.iHead && !p.pProp.iTail");
    assert(!p.pJust.iHead && !p.pJust.iTail, "!p.pJust.iHead && !p.pJust.iTail");
    assert(p.pClauses.iHead == 1 && p.pClauses.iTail == 1, "p.pClauses.iHead == 1 && p.pClauses.iTail == 1");
    p.Pars.nBTThis = p.Pars.nJustThis = p.Pars.nBTThisNc = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        if pObj != Gia_ManConst1(p.pAig) && !(Tas_VarIsAssigned(Gia_Regular(pObj)) != 0) {
            Tas_ManAssign( p, pObj, 0, nil, nil ); }
    }
    if !(Tas_ManSolve_rec(p, 0) != 0) && !(Tas_ManCheckLimits(p) != 0) {
        Tas_ManSaveModel( p, p.vModel );
    } else {
        RetValue = 1; }
    Tas_ManCancelUntil( p, 0 );
    p.pJust.iHead = p.pJust.iTail = 0;
    p.pClauses.iHead = p.pClauses.iTail = 1;
    // clauses
    if p.nClauses > 0 {
        p.pStore.iCur = 16;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vWatchLits)) && (((Entry) = Vec_IntEntry(p.vWatchLits, i)), 1); i++  {
            p.pWatches[Entry] = 0;
        }
        Vec_IntClear( p.vWatchLits );
        p.nClauses = 0;
    }
    // activity
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vActiveVars)) && (((Entry) = Vec_IntEntry(p.vActiveVars, i)), 1); i++  {
        p.pActivity[Entry] = 0.0;
    }
    Vec_IntClear( p.vActiveVars );
    // statistics
    p.Pars.nBTTotal += p.Pars.nBTThis;
    p.Pars.nJustTotal = Abc_MaxInt( p.Pars.nJustTotal, p.Pars.nJustThis );
    if Tas_ManCheckLimits( p ) != 0 {
        RetValue = -1; }

//    printf( "%d ", Gia_ManObjNum(p->pAig) );
//    printf( "%d ", p->Pars.nBTThis );
//    printf( "%d ", p->Pars.nJustThis );
//    printf( "%d ", s_Counter2 );
//    printf( "%d ", s_Counter3 );
//    printf( "%d  ", s_Counter4 );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Prints statistics of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManSatPrintStats(p *Tas_Man_t) {
    fmt.Printf("CO = %8d  ", Gia_ManCoNum(p.pAig));
    fmt.Printf("AND = %8d  ", Gia_ManAndNum(p.pAig));
    fmt.Printf("Conf = %6d  ", p.Pars.nBTLimit);
    fmt.Printf("JustMax = %5d  ", p.Pars.nJustLimit);
    fmt.Printf("\n");
    fmt.Printf("Unsat calls %6d  (%6.2f %%)   Ave conf = %8.1f   ", p.nSatUnsat, p.nSatTotal? 100.0*p.nSatUnsat/p.nSatTotal :0.0, p.nSatUnsat? 1.0*p.nConfUnsat/p.nSatUnsat :0.0);
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUnsat)))/(float64((p.timeTotal))) : 0.0);;
    fmt.Printf("Sat   calls %6d  (%6.2f %%)   Ave conf = %8.1f   ", p.nSatSat, p.nSatTotal? 100.0*p.nSatSat/p.nSatTotal :0.0, p.nSatSat? 1.0*p.nConfSat/p.nSatSat : 0.0);
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((p.timeTotal))) : 0.0);;
    fmt.Printf("Undef calls %6d  (%6.2f %%)   Ave conf = %8.1f   ", p.nSatUndec, p.nSatTotal? 100.0*p.nSatUndec/p.nSatTotal :0.0, p.nSatUndec? 1.0*p.nConfUndec/p.nSatUndec : 0.0);
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUndec)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUndec)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Procedure to test the new SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManSolveMiterNc(pAig *Gia_Man_t, nConfs int32, pvStatus **Vec_Str_t, fVerbose int32) *Vec_Int_t {
    extern void Gia_ManCollectTest( Gia_Man_t * pAig );
    extern void Cec_ManSatAddToStore( Vec_Int_t * vCexStore, Vec_Int_t * vCex, int Out );
    var p *Tas_Man_t; 
    var vCex, vVisit, vCexStore *Vec_Int_t;
    var vStatus *Vec_Str_t;
    var pRoot *Gia_Obj_t;//, * pRootCopy; 
//    Gia_Man_t * pAigCopy = Gia_ManDup( pAig ), * pAigTemp;

    var i, status int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    assert(Gia_ManRegNum(pAig) == 0, "Gia_ManRegNum(pAig) == 0");
//    Gia_ManCollectTest( pAig );
    // prepare AIG
    Gia_ManCreateRefs( pAig );
    Gia_ManCleanMark0( pAig );
    Gia_ManCleanMark1( pAig );
    Gia_ManFillValue( pAig ); // maps nodes into trail ids
    Gia_ManCleanPhase( pAig ); // maps nodes into trail ids
    // create logic network
    p = Tas_ManAlloc( pAig, nConfs );
    p.pAig   = pAig;
    // create resulting data-structures
    vStatus   = Vec_StrAlloc( Gia_ManPoNum(pAig) );
    vCexStore = Vec_IntAlloc( 10000 );
    vVisit    = Vec_IntAlloc( 100 );
    vCex      = Tas_ReadModel( p );
    // solve for each output
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pAig.vCos)) && ((pRoot) = Gia_ManCo(pAig, i)); i++  {
//        printf( "%d=", i );

        Vec_IntClear( vCex );
        if Gia_ObjIsConst0(Gia_ObjFanin0(pRoot)) {
            if Gia_ObjFaninC0(pRoot) {
//                printf( "Constant 1 output of SRM!!!\n" );
                Cec_ManSatAddToStore( vCexStore, vCex, i ); // trivial counter-example
                Vec_StrPush( vStatus, 0 );
            
            } else {
//                printf( "Constant 0 output of SRM!!!\n" );
                Vec_StrPush( vStatus, 1 );
            }
            continue;
        } 
        clk = Abc_Clock();
//        p->Pars.fUseActive  = 1;
        p.Pars.fUseHighest = 1;
        p.Pars.fUseLowest  = 0;
        status = Tas_ManSolve( p, Gia_ObjChild0(pRoot), nil );
//        printf( "\n" );
/*
        if ( status == -1 )
        {
            p->Pars.fUseHighest = 0;
            p->Pars.fUseLowest  = 1;
            status = Tas_ManSolve( p, Gia_ObjChild0(pRoot) );
        }
*/
        Vec_StrPush( vStatus, byte(status) );
        if status == -1 {
//            printf( "Unsolved %d.\n", i );

            p.nSatUndec++;
            p.nConfUndec += p.Pars.nBTThis;
            Cec_ManSatAddToStore( vCexStore, nil, i ); // timeout
            p.timeSatUndec += Abc_Clock() - clk;
            continue;
        }

//        pRootCopy = Gia_ManCo( pAigCopy, i );
//        pRootCopy->iDiff0 = Gia_ObjId( pAigCopy, pRootCopy );
//        pRootCopy->fCompl0 = 0;

        if status == 1 {
            p.nSatUnsat++;
            p.nConfUnsat += p.Pars.nBTThis;
            p.timeSatUnsat += Abc_Clock() - clk;
            continue;
        }
        p.nSatSat++;
        p.nConfSat += p.Pars.nBTThis;
//        Gia_SatVerifyPattern( pAig, pRoot, vCex, vVisit );
        Cec_ManSatAddToStore( vCexStore, vCex, i );
        p.timeSatSat += Abc_Clock() - clk;

//        printf( "%d ", Vec_IntSize(vCex) );
    }
//    pAigCopy = Gia_ManCleanup( pAigTemp = pAigCopy );
//    Gia_ManStop( pAigTemp );
//    Gia_DumpAiger( pAigCopy, "test", 0, 2 );
//    Gia_ManStop( pAigCopy );

    Vec_IntFree( vVisit );
    p.nSatTotal = Gia_ManPoNum(pAig);
    p.timeTotal = Abc_Clock() - clkTotal;
    if fVerbose != 0 {
        Tas_ManSatPrintStats( p ); }
//    printf( "RecCalls = %8d.  RecClause = %8d.  RecNonChro = %8d.\n", p->nRecCall, p->nRecClause, p->nRecNonChro );
    Tas_ManStop( p );
    *pvStatus = vStatus;

//    printf( "Total number of cex literals = %d. (Ave = %d)\n", 
//         Vec_IntSize(vCexStore)-2*p->nSatUndec-2*p->nSatSat, 
//        (Vec_IntSize(vCexStore)-2*p->nSatUndec-2*p->nSatSat)/p->nSatSat );
    return vCexStore;
}

/**Function*************************************************************

  Synopsis    [Packs patterns into array of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

*************************************`**********************************/
func Tas_StorePatternTry(vInfo *Vec_Ptr_t, vPres *Vec_Ptr_t, iBit int32, pLits *int32, nLits int32) int32 {
    var pInfo, pPres *uint32;
    var i int32;
    for  i = 0; i < nLits; i++  {
        pInfo = (unsigned *)Vec_PtrEntry(vInfo, Abc_Lit2Var(pLits[i]));
        pPres = (unsigned *)Vec_PtrEntry(vPres, Abc_Lit2Var(pLits[i]));
        if Abc_InfoHasBit( pPres, iBit ) && 
             Abc_InfoHasBit( pInfo, iBit ) == Abc_LitIsCompl(pLits[i]) {
             return 0; }
    }
    for  i = 0; i < nLits; i++  {
        pInfo = (unsigned *)Vec_PtrEntry(vInfo, Abc_Lit2Var(pLits[i]));
        pPres = (unsigned *)Vec_PtrEntry(vPres, Abc_Lit2Var(pLits[i]));
        Abc_InfoSetBit( pPres, iBit );
        if Abc_InfoHasBit( pInfo, iBit ) == Abc_LitIsCompl(pLits[i]) {
            Abc_InfoXorBit( pInfo, iBit ); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Procedure to test the new SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_StorePattern(vSimInfo *Vec_Ptr_t, vPres *Vec_Ptr_t, vCex *Vec_Int_t) int32 {
    var k int32;
    for  k = 1; k < 32; k++  {
        if Tas_StorePatternTry( vSimInfo, vPres, k, (int *)Vec_IntArray(vCex), Vec_IntSize(vCex) ) {
            break; }
    }
    return int32((k < 32));
}

/**Function*************************************************************

  Synopsis    [Procedure to test the new SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManSolveMiterNc2(pAig *Gia_Man_t, nConfs int32, pAigOld *Gia_Man_t, vOldRoots *Vec_Ptr_t, vSimInfo *Vec_Ptr_t) {
    var nPatMax int32  = 1000;
    var fVerbose int32  = 1;
    extern void Gia_ManCollectTest( Gia_Man_t * pAig );
    extern void Cec_ManSatAddToStore( Vec_Int_t * vCexStore, Vec_Int_t * vCex, int Out );
    var p *Tas_Man_t; 
    var vPres *Vec_Ptr_t;
    var vCex, vVisit, vCexStore *Vec_Int_t;
    var vStatus *Vec_Str_t;
    var pRoot, pOldRoot *Gia_Obj_t; 
    var i, status int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    var Tried int32;
    var Stored int32;
    var Step int32 = Gia_ManCoNum(pAig) / nPatMax;
    assert(Gia_ManRegNum(pAig) == 0, "Gia_ManRegNum(pAig) == 0");
//    Gia_ManCollectTest( pAig );
    // prepare AIG
    Gia_ManCreateRefs( pAig );
    Gia_ManCleanMark0( pAig );
    Gia_ManCleanMark1( pAig );
    Gia_ManFillValue( pAig ); // maps nodes into trail ids
    Gia_ManCleanPhase( pAig ); // maps nodes into trail ids
    // create logic network
    p = Tas_ManAlloc( pAig, nConfs );
    p.pAig   = pAig;
    // create resulting data-structures
    vStatus   = Vec_StrAlloc( Gia_ManPoNum(pAig) );
    vCexStore = Vec_IntAlloc( 10000 );
    vVisit    = Vec_IntAlloc( 100 );
    vCex      = Tas_ReadModel( p );
    // solve for each output
    vPres = Vec_PtrAllocSimInfo( Gia_ManCiNum(pAig), 1 );
    Vec_PtrCleanSimInfo( vPres, 0, 1 );
    
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pAig.vCos)) && ((pRoot) = Gia_ManCo(pAig, i)); i++  {
        assert(!Gia_ObjIsConst0(Gia_ObjFanin0(pRoot)), "!Gia_ObjIsConst0(Gia_ObjFanin0(pRoot))");
        Vec_IntClear( vCex );
        clk = Abc_Clock();
        p.Pars.fUseHighest = 1;
        p.Pars.fUseLowest  = 0;
        status = Tas_ManSolve( p, Gia_ObjChild0(pRoot), nil );
        Vec_StrPush( vStatus, byte(status) );
        if status == -1 {
            p.nSatUndec++;
            p.nConfUndec += p.Pars.nBTThis;
//            Cec_ManSatAddToStore( vCexStore, NULL, i ); // timeout
            p.timeSatUndec += Abc_Clock() - clk;

            i += Step;
            continue;
        }
        if status == 1 {
            p.nSatUnsat++;
            p.nConfUnsat += p.Pars.nBTThis;
            p.timeSatUnsat += Abc_Clock() - clk;
            // record proved
            pOldRoot = (Gia_Obj_t *)Vec_PtrEntry( vOldRoots, i );
            assert(!Gia_ObjProved( pAigOld, Gia_ObjId(pAigOld, pOldRoot) ), "!Gia_ObjProved( pAigOld, Gia_ObjId(pAigOld, pOldRoot) )");
            Gia_ObjSetProved( pAigOld, Gia_ObjId(pAigOld, pOldRoot) );

            i += Step;
            continue;
        }
        p.nSatSat++;
        p.nConfSat += p.Pars.nBTThis;
//        Gia_SatVerifyPattern( pAig, pRoot, vCex, vVisit );
//        Cec_ManSatAddToStore( vCexStore, vCex, i );

        // save pattern
        Tried++;
        Stored += Tas_StorePattern( vSimInfo, vPres, vCex );
        p.timeSatSat += Abc_Clock() - clk;
        i += Step;
    }
    fmt.Printf("Tried = %d  Stored = %d\n", Tried, Stored);
    Vec_IntFree( vVisit );
    p.nSatTotal = Gia_ManPoNum(pAig);
    p.timeTotal = Abc_Clock() - clkTotal;
    if fVerbose != 0 {
        Tas_ManSatPrintStats( p ); }
    Tas_ManStop( p );
    Vec_PtrFree( vPres );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

