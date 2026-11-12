package gia

import (
	"fmt"
	"time"
)

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

   
type Cbs_Par_t struct {
    // conflict limits
    nBTLimit           int32;     // limit on the number of conflicts
    nJustLimit           int32;   // limit on the size of justification queue
    // current parameters
    nBTThis           int32;      // number of conflicts
    nBTThisNc           int32;    // number of conflicts
    nJustThis           int32;    // max size of the frontier
    nBTTotal           int32;     // total number of conflicts
    nJustTotal           int32;   // total size of the frontier
    // decision heuristics
    fUseHighest           int32;  // use node with the highest ID
    fUseLowest           int32;   // use node with the highest ID
    fUseMaxFF           int32;    // use node with the largest fanin fanout
    // other
    fVerbose           int32;
}

   
type Cbs_Que_t struct {
    iHead           int32;        // beginning of the queue
    iTail           int32;        // end of the queue
    nSize           int32;        // allocated size
    pData   **Gia_Obj_t;        // nodes stored in the queue
}
 
//typedef struct Cbs_Man_t_ Cbs_Man_t;
type Cbs_Man_t_ struct {
    Pars     Cbs_Par_t;         // parameters
    pAig    *Gia_Man_t;         // AIG manager
    pProp     Cbs_Que_t;        // propagation queue
    pJust     Cbs_Que_t;        // justification queue
    pClauses     Cbs_Que_t;     // clause queue
    pIter   **Gia_Obj_t;        // iterator through clause vars
    vLevReas    *Vec_Int_t;     // levels and decisions
    vModel    *Vec_Int_t;       // satisfying assignment
    vTemp    *Vec_Ptr_t;        // temporary storage
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

func Cbs_VarIsAssigned(pVar *Gia_Obj_t) int32 { return pVar.fMark0;                        }
func Cbs_VarAssign(pVar *Gia_Obj_t) { assert(!pVar.fMark0, "!pVar.fMark0"); pVar.fMark0 = 1;    }
func Cbs_VarUnassign(pVar *Gia_Obj_t) { assert(pVar.fMark0, "pVar.fMark0");  pVar.fMark0 = 0; pVar.fMark1 = 0; pVar.Value = ~0; }
func Cbs_VarValue(pVar *Gia_Obj_t) int32 { assert(pVar.fMark0, "pVar.fMark0");  return pVar.fMark1; }
func Cbs_VarSetValue(pVar *Gia_Obj_t, v int32) { assert(pVar.fMark0, "pVar.fMark0");  pVar.fMark1 = v;    }
func Cbs_VarIsJust(pVar *Gia_Obj_t) int32 { return Gia_ObjIsAnd(pVar) && !Cbs_VarIsAssigned(Gia_ObjFanin0(pVar)) && !Cbs_VarIsAssigned(Gia_ObjFanin1(pVar)); } 
func Cbs_VarFanin0Value(pVar *Gia_Obj_t) int32 { if !(Cbs_VarIsAssigned(Gia_ObjFanin0(pVar)) != 0) {
    return 2;
} else {
    return (Cbs_VarValue(Gia_ObjFanin0(pVar)) ^ Gia_ObjFaninC0(pVar));
} }
func Cbs_VarFanin1Value(pVar *Gia_Obj_t) int32 { if !(Cbs_VarIsAssigned(Gia_ObjFanin1(pVar)) != 0) {
    return 2;
} else {
    return (Cbs_VarValue(Gia_ObjFanin1(pVar)) ^ Gia_ObjFaninC1(pVar));
} }

func Cbs_VarDecLevel(p *Cbs_Man_t, pVar *Gia_Obj_t) int32 { assert(pVar.Value != ~0, "pVar.Value != ~0"); return Vec_IntEntry(p.vLevReas, 3*pVar.Value);          }
func Cbs_VarReason0(p *Cbs_Man_t, pVar *Gia_Obj_t) *Gia_Obj_t { assert(pVar.Value != ~0, "pVar.Value != ~0"); return pVar + Vec_IntEntry(p.vLevReas, 3*pVar.Value+1); }
func Cbs_VarReason1(p *Cbs_Man_t, pVar *Gia_Obj_t) *Gia_Obj_t { assert(pVar.Value != ~0, "pVar.Value != ~0"); return pVar + Vec_IntEntry(p.vLevReas, 3*pVar.Value+2); }
func Cbs_ClauseDecLevel(p *Cbs_Man_t, hClause int32) int32 { return Cbs_VarDecLevel( p, p.pClauses.pData[hClause] );                               }

// Original macro:
// #define Cbs_QueForEachEntry( Que, pObj, i )                         \
//     for ( i = (Que).iHead; (i < (Que).iTail) && ((pObj) = (Que).pData[i]); i++ )


// Original macro:
// #define Cbs_ClauseForEachVar( p, hClause, pObj )                    \
//     for ( (p)->pIter = (p)->pClauses.pData + hClause; (pObj = *pIter); (p)->pIter++ )

// Original macro:
// #define Cbs_ClauseForEachVar1( p, hClause, pObj )                   \
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
func Cbs_SetDefaultParams(pPars *Cbs_Par_t) {
    *pPars = Cbs_Par_t{};
    pPars.nBTLimit    =  1000;   // limit on the number of conflicts
    pPars.nJustLimit  =   100;   // limit on the size of justification queue
    pPars.fUseHighest =     1;   // use node with the highest ID
    pPars.fUseLowest  =     0;   // use node with the highest ID
    pPars.fUseMaxFF   =     0;   // use node with the largest fanin fanout
    pPars.fVerbose    =     1;   // print detailed statistics
}
func Cbs_ManSetConflictNum(p *Cbs_Man_t, Num int32) {
    p.Pars.nBTLimit = Num;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManAlloc(pGia *Gia_Man_t) *Cbs_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Cbs_Man_t{};
    p.pProp.nSize = p.pJust.nSize = p.pClauses.nSize = 10000;
    // inlining of macro ABC_ALLOC
    p.pProp.pData = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.pProp.nSize)));
    // inlining of macro ABC_ALLOC
    p.pJust.pData = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.pJust.nSize)));
    // inlining of macro ABC_ALLOC
    p.pClauses.pData = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.pClauses.nSize)));
    p.pClauses.iHead = p.pClauses.iTail = 1;
    p.vModel   = Vec_IntAlloc( 1000 );
    p.vLevReas = Vec_IntAlloc( 1000 );
    p.vTemp    = Vec_PtrAlloc( 1000 );
    p.pAig     = pGia;
    Cbs_SetDefaultParams( &p.Pars );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManStop(p *Cbs_Man_t) {
    Vec_IntFree( p.vLevReas );
    Vec_IntFree( p.vModel );
    Vec_PtrFree( p.vTemp );
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
func Cbs_ReadModel(p *Cbs_Man_t) *Vec_Int_t {
    return p.vModel;
}




/**Function*************************************************************

  Synopsis    [Returns 1 if the solver is out of limits.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManCheckLimits(p *Cbs_Man_t) int32 {
    return p.Pars.nJustThis > p.Pars.nJustLimit || p.Pars.nBTThis > p.Pars.nBTLimit;
}

/**Function*************************************************************

  Synopsis    [Saves the satisfying assignment as an array of literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManSaveModel(p *Cbs_Man_t, vCex *Vec_Int_t) {
    var pVar *Gia_Obj_t;
    var i int32;
    Vec_IntClear( vCex );
    p.pProp.iHead = 0;
    // inlining of macro Cbs_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
        if Gia_ObjIsCi(pVar) {
//            Vec_IntPush( vCex, Abc_Var2Lit(Gia_ObjId(p->pAig,pVar), !Cbs_VarValue(pVar)) );
            Vec_IntPush( vCex, Abc_Var2Lit(Gia_ObjCioId(pVar), !Cbs_VarValue(pVar)) ); }
    }
} 
func Cbs_ManSaveModelAll(p *Cbs_Man_t, vCex *Vec_Int_t) {
    var pVar *Gia_Obj_t;
    var i int32;
    Vec_IntClear( vCex );
    p.pProp.iHead = 0;
    // inlining of macro Cbs_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
        Vec_IntPush( vCex, Abc_Var2Lit(Gia_ObjId(p.pAig,pVar), !Cbs_VarValue(pVar)) );
    }
} 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_QueIsEmpty(p *Cbs_Que_t) int32 {
    return p.iHead == p.iTail;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_QuePush(p *Cbs_Que_t, pObj *Gia_Obj_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
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
func Cbs_QueHasNode(p *Cbs_Que_t, pObj *Gia_Obj_t) int32 {
    var pTemp *Gia_Obj_t;
    var i int32;
    // inlining of macro Cbs_QueForEachEntry
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
func Cbs_QueStore(p *Cbs_Que_t, piHeadOld *int32, piTailOld *int32) {
    var i int32;
    *piHeadOld = p.iHead;
    *piTailOld = p.iTail;
    for  i = *piHeadOld; i < *piTailOld; i++  {
        Cbs_QuePush( p, p.pData[i] );
    }
    p.iHead = *piTailOld;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_QueRestore(p *Cbs_Que_t, iHeadOld int32, iTailOld int32) {
    p.iHead = iHeadOld;
    p.iTail = iTailOld;
}

/**Function*************************************************************

  Synopsis    [Finalized the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_QueFinish(p *Cbs_Que_t) int32 {
    var iHeadOld int32  = p.iHead;
    assert(p.iHead < p.iTail, "p.iHead < p.iTail");
    Cbs_QuePush( p, nil );
    p.iHead = p.iTail;
    return iHeadOld;
}


/**Function*************************************************************

  Synopsis    [Max number of fanins fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_VarFaninFanoutMax(p *Cbs_Man_t, pObj *Gia_Obj_t) int32 {
    var Count0, Count1 int32;
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Count0 = Gia_ObjRefNum( p.pAig, Gia_ObjFanin0(pObj) );
    Count1 = Gia_ObjRefNum( p.pAig, Gia_ObjFanin1(pObj) );
    return Abc_MaxInt( Count0, Count1 );
}

/**Function*************************************************************

  Synopsis    [Find variable with the highest ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManDecideHighest(p *Cbs_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i int32;
    // inlining of macro Cbs_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
        if pObjMax == nil || pObjMax < pObj {
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
func Cbs_ManDecideLowest(p *Cbs_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMin *Gia_Obj_t = nil;
    var i int32;
    // inlining of macro Cbs_QueForEachEntry
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
func Cbs_ManDecideMaxFF(p *Cbs_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i, iMaxFF, iCurFF int32;
    assert(p.pAig.Refs != nil, "p.pAig.Refs != nil");
    // inlining of macro Cbs_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
        iCurFF = Cbs_VarFaninFanoutMax( p, pObj );
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
func Cbs_ManCancelUntil(p *Cbs_Man_t, iBound int32) {
    var pVar *Gia_Obj_t;
    var i int32;
    assert(iBound <= p.pProp.iTail, "iBound <= p.pProp.iTail");
    p.pProp.iHead = iBound;
    // inlining of macro Cbs_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
        Cbs_VarUnassign( pVar );
    }
    p.pProp.iTail = iBound;
    Vec_IntShrink( p.vLevReas, 3*iBound );
}

var s_Counter int32;

/**Function*************************************************************

  Synopsis    [Assigns the variables a value.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManAssign(p *Cbs_Man_t, pObj *Gia_Obj_t, Level int32, pRes0 *Gia_Obj_t, pRes1 *Gia_Obj_t) {
    var pObjR *Gia_Obj_t  = Gia_Regular(pObj);
    assert(Gia_ObjIsCand(pObjR), "Gia_ObjIsCand(pObjR)");
    assert(!(Cbs_VarIsAssigned(pObjR) != 0), "!Cbs_VarIsAssigned(pObjR)");
    Cbs_VarAssign( pObjR );
    Cbs_VarSetValue( pObjR, !Gia_IsComplement(pObj) );
    assert(pObjR.Value == ~0, "pObjR.Value == ~0");
    pObjR.Value = p.pProp.iTail;
    Cbs_QuePush( &p.pProp, pObjR );
    Vec_IntPush( p.vLevReas, Level );
    Vec_IntPush( p.vLevReas, pRes0 ? pRes0-pObjR : 0 );
    Vec_IntPush( p.vLevReas, pRes1 ? pRes1-pObjR : 0 );
    assert(Vec_IntSize(p.vLevReas) == 3 * p.pProp.iTail, "Vec_IntSize(p.vLevReas) == 3 * p.pProp.iTail");
//    s_Counter++;
//    s_Counter = Abc_MaxIntInt( s_Counter, Vec_IntSize(p->vLevReas)/3 );
}


/**Function*************************************************************

  Synopsis    [Returns clause size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManClauseSize(p *Cbs_Man_t, hClause int32) int32 {
    var pQue *Cbs_Que_t  = &(p.pClauses);
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
func Cbs_ManPrintClause(p *Cbs_Man_t, Level int32, hClause int32) {
    var pQue *Cbs_Que_t  = &(p.pClauses);
    var pObj *Gia_Obj_t;
    var i int32;
    assert((Cbs_QueIsEmpty( pQue ) != 0), "Cbs_QueIsEmpty( pQue )");
    fmt.Printf("Level %2d : ", Level);
    for  i = hClause; (pObj = pQue.pData[i]); i++  {
        fmt.Printf("%d=%d(%d) ", Gia_ObjId(p.pAig, pObj), Cbs_VarValue(pObj), Cbs_VarDecLevel(p, pObj));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Prints conflict clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManPrintClauseNew(p *Cbs_Man_t, Level int32, hClause int32) {
    var pQue *Cbs_Que_t  = &(p.pClauses);
    var pObj *Gia_Obj_t;
    var i int32;
    assert((Cbs_QueIsEmpty( pQue ) != 0), "Cbs_QueIsEmpty( pQue )");
    fmt.Printf("Level %2d : ", Level);
    for  i = hClause; (pObj = pQue.pData[i]); i++  {
        fmt.Printf("%c%d ", Cbs_VarValue(pObj)? '+':'-', Gia_ObjId(p.pAig, pObj));
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Returns conflict clause.]

  Description [Performs conflict analysis.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManDeriveReason(p *Cbs_Man_t, Level int32) {
    var pQue *Cbs_Que_t  = &(p.pClauses);
    var pObj, pReason *Gia_Obj_t;
    var i, k, iLitLevel int32;
    assert(pQue.pData[pQue.iHead] == nil, "pQue.pData[pQue.iHead] == nil");
    assert(pQue.iHead + 1 < pQue.iTail, "pQue.iHead + 1 < pQue.iTail");
/*
    for ( i = pQue->iHead + 1; i < pQue->iTail; i++ )
    {
        pObj = pQue->pData[i];
        assert( pObj->fMark0 == 1 );
    }
*/
    // compact literals
    Vec_PtrClear( p.vTemp );
    for  i = k = pQue.iHead + 1; i < pQue.iTail; i++  {
        pObj = pQue.pData[i];
        if !pObj.fMark0 { // unassigned - seen again
            continue; }
        // assigned - seen first time
        pObj.fMark0 = 0;
        Vec_PtrPush( p.vTemp, pObj );
        // check decision level
        iLitLevel = Cbs_VarDecLevel( p, pObj );
        if iLitLevel < Level {
            pQue.pData[k++] = pObj;
            continue;
        }
        assert(iLitLevel == Level, "iLitLevel == Level");
        pReason = Cbs_VarReason0( p, pObj );
        if pReason == pObj { // no reason
        
            //assert( pQue->pData[pQue->iHead] == NULL );
            pQue.pData[pQue.iHead] = pObj;
            continue;
        }
        Cbs_QuePush( pQue, pReason );
        pReason = Cbs_VarReason1( p, pObj );
        if pReason != pObj { // second reason
            Cbs_QuePush( pQue, pReason ); }
    }
    assert(pQue.pData[pQue.iHead] != nil, "pQue.pData[pQue.iHead] != nil");
    pQue.iTail = k;
    // clear the marks
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTemp)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(p.vTemp, i)), 1); i++  {
        pObj.fMark0 = 1;
    }
}

/**Function*************************************************************

  Synopsis    [Returns conflict clause.]

  Description [Performs conflict analysis.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManAnalyze(p *Cbs_Man_t, Level int32, pVar *Gia_Obj_t, pFan0 *Gia_Obj_t, pFan1 *Gia_Obj_t) int32 {
    var pQue *Cbs_Que_t  = &(p.pClauses);
    assert((Cbs_VarIsAssigned(pVar) != 0), "Cbs_VarIsAssigned(pVar)");
    assert((Cbs_VarIsAssigned(pFan0) != 0), "Cbs_VarIsAssigned(pFan0)");
    assert(pFan1 == nil || (Cbs_VarIsAssigned(pFan1) != 0), "pFan1 == nil || Cbs_VarIsAssigned(pFan1)");
    assert((Cbs_QueIsEmpty( pQue ) != 0), "Cbs_QueIsEmpty( pQue )");
    Cbs_QuePush( pQue, nil );
    Cbs_QuePush( pQue, pVar );
    Cbs_QuePush( pQue, pFan0 );
    if pFan1 != nil {
        Cbs_QuePush( pQue, pFan1 ); }
    Cbs_ManDeriveReason( p, Level );
    return Cbs_QueFinish( pQue );
}


/**Function*************************************************************

  Synopsis    [Performs resolution of two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManResolve(p *Cbs_Man_t, Level int32, hClause0 int32, hClause1 int32) int32 {
    var pQue *Cbs_Que_t  = &(p.pClauses);
    var pObj *Gia_Obj_t;
    var i int32;
    var LevelMax int32 = -1;
    var LevelCur int32;
    assert(pQue.pData[hClause0] != nil, "pQue.pData[hClause0] != nil");
    assert(pQue.pData[hClause0] == pQue.pData[hClause1], "pQue.pData[hClause0] == pQue.pData[hClause1]");
/*
    for ( i = hClause0 + 1; (pObj = pQue->pData[i]); i++ )
        assert( pObj->fMark0 == 1 );
    for ( i = hClause1 + 1; (pObj = pQue->pData[i]); i++ )
        assert( pObj->fMark0 == 1 );
*/
    assert((Cbs_QueIsEmpty( pQue ) != 0), "Cbs_QueIsEmpty( pQue )");
    Cbs_QuePush( pQue, nil );
    for  i = hClause0 + 1; (pObj = pQue.pData[i]); i++  {
        if !pObj.fMark0 { // unassigned - seen again
            continue; }
        // assigned - seen first time
        pObj.fMark0 = 0;
        Cbs_QuePush( pQue, pObj );
        LevelCur = Cbs_VarDecLevel( p, pObj );
        if LevelMax < LevelCur {
            LevelMax = LevelCur; }
    }
    for  i = hClause1 + 1; (pObj = pQue.pData[i]); i++  {
        if !pObj.fMark0 { // unassigned - seen again
            continue; }
        // assigned - seen first time
        pObj.fMark0 = 0;
        Cbs_QuePush( pQue, pObj );
        LevelCur = Cbs_VarDecLevel( p, pObj );
        if LevelMax < LevelCur {
            LevelMax = LevelCur; }
    }
    for  i = pQue.iHead + 1; i < pQue.iTail; i++  {
        pQue.pData[i].fMark0 = 1;
    }
    Cbs_ManDeriveReason( p, LevelMax );
    return Cbs_QueFinish( pQue );
}

/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns clause handle if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManPropagateOne(p *Cbs_Man_t, pVar *Gia_Obj_t, Level int32) int32 {
    var Value0, Value1 int32;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert((Cbs_VarIsAssigned(pVar) != 0), "Cbs_VarIsAssigned(pVar)");
    if Gia_ObjIsCi(pVar) {
        return 0; }
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    Value0 = Cbs_VarFanin0Value(pVar);
    Value1 = Cbs_VarFanin1Value(pVar);
    if Cbs_VarValue(pVar) != 0 { // value is 1
        if Value0 == 0 || Value1 == 0 { // one is 0
        
            if Value0 == 0 && Value1 != 0 {
                return Cbs_ManAnalyze( p, Level, pVar, Gia_ObjFanin0(pVar), nil ); }
            if Value0 != 0 && Value1 == 0 {
                return Cbs_ManAnalyze( p, Level, pVar, Gia_ObjFanin1(pVar), nil ); }
            assert(Value0 == 0 && Value1 == 0, "Value0 == 0 && Value1 == 0");
            return Cbs_ManAnalyze( p, Level, pVar, Gia_ObjFanin0(pVar), Gia_ObjFanin1(pVar) );
        }
        if Value0 == 2 { // first is unassigned
            Cbs_ManAssign( p, Gia_ObjChild0(pVar), Level, pVar, nil ); }
        if Value1 == 2 { // first is unassigned
            Cbs_ManAssign( p, Gia_ObjChild1(pVar), Level, pVar, nil ); }
        return 0;
    }
    // value is 0
    if Value0 == 0 || Value1 == 0 { // one is 0
        return 0; }
    if Value0 == 1 && Value1 == 1 { // both are 1
        return Cbs_ManAnalyze( p, Level, pVar, Gia_ObjFanin0(pVar), Gia_ObjFanin1(pVar) ); }
    if Value0 == 1 || Value1 == 1 { // one is 1 
    
        if Value0 == 2 { // first is unassigned
            Cbs_ManAssign( p, Gia_Not(Gia_ObjChild0(pVar)), Level, pVar, Gia_ObjFanin1(pVar) ); }
        if Value1 == 2 { // second is unassigned
            Cbs_ManAssign( p, Gia_Not(Gia_ObjChild1(pVar)), Level, pVar, Gia_ObjFanin0(pVar) ); }
        return 0;
    }
    assert((Cbs_VarIsJust(pVar) != 0), "Cbs_VarIsJust(pVar)");
    assert(!(Cbs_QueHasNode( &p.pJust, pVar ) != 0), "!Cbs_QueHasNode( &p.pJust, pVar )");
    Cbs_QuePush( &p.pJust, pVar );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManPropagateTwo(p *Cbs_Man_t, pVar *Gia_Obj_t, Level int32) int32 {
    var Value0, Value1 int32;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    assert((Cbs_VarIsAssigned(pVar) != 0), "Cbs_VarIsAssigned(pVar)");
    assert(!(Cbs_VarValue(pVar) != 0), "!Cbs_VarValue(pVar)");
    Value0 = Cbs_VarFanin0Value(pVar);
    Value1 = Cbs_VarFanin1Value(pVar);
    // value is 0
    if Value0 == 0 || Value1 == 0 { // one is 0
        return 0; }
    if Value0 == 1 && Value1 == 1 { // both are 1
        return Cbs_ManAnalyze( p, Level, pVar, Gia_ObjFanin0(pVar), Gia_ObjFanin1(pVar) ); }
    assert(Value0 == 1 || Value1 == 1, "Value0 == 1 || Value1 == 1");
    if Value0 == 2 { // first is unassigned
        Cbs_ManAssign( p, Gia_Not(Gia_ObjChild0(pVar)), Level, pVar, Gia_ObjFanin1(pVar) ); }
    if Value1 == 2 { // first is unassigned
        Cbs_ManAssign( p, Gia_Not(Gia_ObjChild1(pVar)), Level, pVar, Gia_ObjFanin0(pVar) ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Propagates all variables.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManPropagate(p *Cbs_Man_t, Level int32) int32 {
    var hClause int32;
    var pVar *Gia_Obj_t;
    var i, k int32;
    for {
        // inlining of macro Cbs_QueForEachEntry
        for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
            if hClause = Cbs_ManPropagateOne( p, pVar, Level ) {
                return hClause; }
        }
        p.pProp.iHead = p.pProp.iTail;
        k = p.pJust.iHead;
        // inlining of macro Cbs_QueForEachEntry
        for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pVar) = (p.pJust).pData[i]); i++  {
            if Cbs_VarIsJust( pVar ) != 0 {
                p.pJust.pData[k++] = pVar;
            } else if hClause = Cbs_ManPropagateTwo( p, pVar, Level ) {
                return hClause; }
        }
        if k == p.pJust.iTail {
            break; }
        p.pJust.iTail = k;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Solve the problem recursively.]

  Description [Returns learnt clause if unsat, NULL if sat or undecided.]
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Cbs_ManSolve_rec(p *Cbs_Man_t, Level int32) int32 { 
    var pQue *Cbs_Que_t  = &(p.pClauses);
    var pVar *Gia_Obj_t = nil;
    var pDecVar *Gia_Obj_t;
    var hClause, hLearn0, hLearn1 int32;
    var iPropHead, iJustHead, iJustTail int32;
    // propagate assignments
    assert(!(Cbs_QueIsEmpty(&p.pProp) != 0), "!Cbs_QueIsEmpty(&p.pProp)");
    if hClause = Cbs_ManPropagate( p, Level ) {
        return hClause; }
    // check for satisfying assignment
    assert((Cbs_QueIsEmpty(&p.pProp) != 0), "Cbs_QueIsEmpty(&p.pProp)");
    if Cbs_QueIsEmpty(&p.pJust) != 0 {
        return 0; }
    // quit using resource limits
    p.Pars.nJustThis = Abc_MaxInt( p.Pars.nJustThis, p.pJust.iTail - p.pJust.iHead );
    if Cbs_ManCheckLimits( p ) != 0 {
        return 0; }
    // remember the state before branching
    iPropHead = p.pProp.iHead;
    Cbs_QueStore( &p.pJust, &iJustHead, &iJustTail );
    // find the decision variable
    if p.Pars.fUseHighest {
        pVar = Cbs_ManDecideHighest( p );
    } else if p.Pars.fUseLowest {
        pVar = Cbs_ManDecideLowest( p );
    } else if p.Pars.fUseMaxFF {
        pVar = Cbs_ManDecideMaxFF( p );
    } else { assert(false, "0"); }
    assert((Cbs_VarIsJust( pVar ) != 0), "Cbs_VarIsJust( pVar )");
    // chose decision variable using fanout count
    if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(pVar)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(pVar)) {
        pDecVar = Gia_Not(Gia_ObjChild0(pVar));
    } else {
        pDecVar = Gia_Not(Gia_ObjChild1(pVar)); }
//    pDecVar = Gia_NotCond( Gia_Regular(pDecVar), Gia_Regular(pDecVar)->fPhase );
//    pDecVar = Gia_Not(pDecVar);
    // decide on first fanin
    Cbs_ManAssign( p, pDecVar, Level+1, nil, nil );
    if !(hLearn0 = Cbs_ManSolve_rec( p, Level+1 )) {
        return 0; }
    if pQue.pData[hLearn0] != Gia_Regular(pDecVar) {
        return hLearn0; }
    Cbs_ManCancelUntil( p, iPropHead );
    Cbs_QueRestore( &p.pJust, iJustHead, iJustTail );
    // decide on second fanin
    Cbs_ManAssign( p, Gia_Not(pDecVar), Level+1, nil, nil );
    if !(hLearn1 = Cbs_ManSolve_rec( p, Level+1 )) {
        return 0; }
    if pQue.pData[hLearn1] != Gia_Regular(pDecVar) {
        return hLearn1; }
    hClause = Cbs_ManResolve( p, Level, hLearn0, hLearn1 );
//    Cbs_ManPrintClauseNew( p, Level, hClause );
//    if ( Level > Cbs_ClauseDecLevel(p, hClause) )
//        p->Pars.nBTThisNc++;
    p.Pars.nBTThis++;
    return hClause;
}

/**Function*************************************************************

  Synopsis    [Looking for a satisfying assignment of the node.]

  Description [Assumes that each node has flag pObj->fMark0 set to 0.
  Returns 1 if unsatisfiable, 0 if satisfiable, and -1 if undecided.
  The node may be complemented. ]
               
  SideEffects [The two procedures differ in the CEX format.]

  SeeAlso     []

***********************************************************************/
func Cbs_ManSolve(p *Cbs_Man_t, pObj *Gia_Obj_t) int32 {
    var RetValue int32;
    s_Counter = 0;
    assert(!p.pProp.iHead && !p.pProp.iTail, "!p.pProp.iHead && !p.pProp.iTail");
    assert(!p.pJust.iHead && !p.pJust.iTail, "!p.pJust.iHead && !p.pJust.iTail");
    assert(p.pClauses.iHead == 1 && p.pClauses.iTail == 1, "p.pClauses.iHead == 1 && p.pClauses.iTail == 1");
    p.Pars.nBTThis = p.Pars.nJustThis = p.Pars.nBTThisNc = 0;
    Cbs_ManAssign( p, pObj, 0, nil, nil );
    if !(Cbs_ManSolve_rec(p, 0) != 0) && !(Cbs_ManCheckLimits(p) != 0) {
        Cbs_ManSaveModel( p, p.vModel );
    } else {
        RetValue = 1; }
    Cbs_ManCancelUntil( p, 0 );
    p.pJust.iHead = p.pJust.iTail = 0;
    p.pClauses.iHead = p.pClauses.iTail = 1;
    p.Pars.nBTTotal += p.Pars.nBTThis;
    p.Pars.nJustTotal = Abc_MaxInt( p.Pars.nJustTotal, p.Pars.nJustThis );
    if Cbs_ManCheckLimits( p ) != 0 {
        RetValue = -1; }
//    printf( "%d ", s_Counter );
    return RetValue;
}
func Cbs_ManSolve2(p *Cbs_Man_t, pObj *Gia_Obj_t, pObj2 *Gia_Obj_t) int32 {
    var RetValue int32;
    s_Counter = 0;
    assert(!p.pProp.iHead && !p.pProp.iTail, "!p.pProp.iHead && !p.pProp.iTail");
    assert(!p.pJust.iHead && !p.pJust.iTail, "!p.pJust.iHead && !p.pJust.iTail");
    assert(p.pClauses.iHead == 1 && p.pClauses.iTail == 1, "p.pClauses.iHead == 1 && p.pClauses.iTail == 1");
    p.Pars.nBTThis = p.Pars.nJustThis = p.Pars.nBTThisNc = 0;
    Cbs_ManAssign( p, pObj, 0, nil, nil );
    if pObj2 != nil { 
    Cbs_ManAssign( p, pObj2, 0, nil, nil ); }
    if !(Cbs_ManSolve_rec(p, 0) != 0) && !(Cbs_ManCheckLimits(p) != 0) {
        Cbs_ManSaveModelAll( p, p.vModel );
    } else {
        RetValue = 1; }
    Cbs_ManCancelUntil( p, 0 );
    p.pJust.iHead = p.pJust.iTail = 0;
    p.pClauses.iHead = p.pClauses.iTail = 1;
    p.Pars.nBTTotal += p.Pars.nBTThis;
    p.Pars.nJustTotal = Abc_MaxInt( p.Pars.nJustTotal, p.Pars.nJustThis );
    if Cbs_ManCheckLimits( p ) != 0 {
        RetValue = -1; }
//    printf( "%d ", s_Counter );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Prints statistics of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs_ManSatPrintStats(p *Cbs_Man_t) {
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
func Cbs_ManSolveMiterNc(pAig *Gia_Man_t, nConfs int32, pvStatus **Vec_Str_t, f0Proved int32, fVerbose int32) *Vec_Int_t {
    extern void Gia_ManCollectTest( Gia_Man_t * pAig );
    extern void Cec_ManSatAddToStore( Vec_Int_t * vCexStore, Vec_Int_t * vCex, int Out );
    var p *Cbs_Man_t; 
    var vCex, vVisit, vCexStore *Vec_Int_t;
    var vStatus *Vec_Str_t;
    var pRoot *Gia_Obj_t; 
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
    Gia_ManSetPhase( pAig ); // maps nodes into trail ids
    // create logic network
    p = Cbs_ManAlloc( pAig );
    p.Pars.nBTLimit = nConfs;
    // create resulting data-structures
    vStatus   = Vec_StrAlloc( Gia_ManPoNum(pAig) );
    vCexStore = Vec_IntAlloc( 10000 );
    vVisit    = Vec_IntAlloc( 100 );
    vCex      = Cbs_ReadModel( p );
    // solve for each output
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pAig.vCos)) && ((pRoot) = Gia_ManCo(pAig, i)); i++  {
//        printf( "\n" );

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
        p.Pars.fUseHighest = 1;
        p.Pars.fUseLowest  = 0;
        status = Cbs_ManSolve( p, Gia_ObjChild0(pRoot) );
//        printf( "\n" );
/*
        if ( status == -1 )
        {
            p->Pars.fUseHighest = 0;
            p->Pars.fUseLowest  = 1;
            status = Cbs_ManSolve( p, Gia_ObjChild0(pRoot) );
        }
*/
        Vec_StrPush( vStatus, byte(status) );
        if status == -1 {
            p.nSatUndec++;
            p.nConfUndec += p.Pars.nBTThis;
            Cec_ManSatAddToStore( vCexStore, nil, i ); // timeout
            p.timeSatUndec += Abc_Clock() - clk;
            continue;
        }
        if status == 1 {
            if f0Proved != 0 {
                Gia_ManPatchCoDriver( pAig, i, 0 ); }
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
    }
    Vec_IntFree( vVisit );
    p.nSatTotal = Gia_ManPoNum(pAig);
    p.timeTotal = Abc_Clock() - clkTotal;
    if fVerbose != 0 {
        Cbs_ManSatPrintStats( p ); }
//    printf( "RecCalls = %8d.  RecClause = %8d.  RecNonChro = %8d.\n", p->nRecCall, p->nRecClause, p->nRecNonChro );
    Cbs_ManStop( p );
    *pvStatus = vStatus;

//    printf( "Total number of cex literals = %d. (Ave = %d)\n", 
//         Vec_IntSize(vCexStore)-2*p->nSatUndec-2*p->nSatSat, 
//        (Vec_IntSize(vCexStore)-2*p->nSatUndec-2*p->nSatSat)/p->nSatSat );
    return vCexStore;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

