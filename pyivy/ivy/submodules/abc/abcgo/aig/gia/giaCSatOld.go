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


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Cbs0_Par_t struct {
    // conflict limits
    nBTLimit           int32;     // limit on the number of conflicts
    nJustLimit           int32;   // limit on the size of justification queue
    // current parameters
    nBTThis           int32;      // number of conflicts
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

   
type Cbs0_Que_t struct {
    iHead           int32;        // beginning of the queue
    iTail           int32;        // end of the queue
    nSize           int32;        // allocated size
    pData   **Gia_Obj_t;        // nodes stored in the queue
}
 
   
type Cbs0_Man_t struct {
    Pars    Cbs0_Par_t;         // parameters
    pAig    *Gia_Man_t;         // AIG manager
    pProp    Cbs0_Que_t;        // propagation queue
    pJust    Cbs0_Que_t;        // justification queue
    vModel    *Vec_Int_t;       // satisfying assignment
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

func Cbs0_VarIsAssigned(pVar *Gia_Obj_t) int32 { return pVar.fMark0;                        }
func Cbs0_VarAssign(pVar *Gia_Obj_t) { assert(!pVar.fMark0, "!pVar.fMark0"); pVar.fMark0 = 1;    }
func Cbs0_VarUnassign(pVar *Gia_Obj_t) { assert(pVar.fMark0, "pVar.fMark0");  pVar.fMark0 = 0; pVar.fMark1 = 0;         }
func Cbs0_VarValue(pVar *Gia_Obj_t) int32 { assert(pVar.fMark0, "pVar.fMark0");  return pVar.fMark1; }
func Cbs0_VarSetValue(pVar *Gia_Obj_t, v int32) { assert(pVar.fMark0, "pVar.fMark0");  pVar.fMark1 = v;    }
func Cbs0_VarIsJust(pVar *Gia_Obj_t) int32 { return Gia_ObjIsAnd(pVar) && !Cbs0_VarIsAssigned(Gia_ObjFanin0(pVar)) && !Cbs0_VarIsAssigned(Gia_ObjFanin1(pVar)); } 
func Cbs0_VarFanin0Value(pVar *Gia_Obj_t) int32 { if !(Cbs0_VarIsAssigned(Gia_ObjFanin0(pVar)) != 0) {
    return 2;
} else {
    return (Cbs0_VarValue(Gia_ObjFanin0(pVar)) ^ Gia_ObjFaninC0(pVar));
} }
func Cbs0_VarFanin1Value(pVar *Gia_Obj_t) int32 { if !(Cbs0_VarIsAssigned(Gia_ObjFanin1(pVar)) != 0) {
    return 2;
} else {
    return (Cbs0_VarValue(Gia_ObjFanin1(pVar)) ^ Gia_ObjFaninC1(pVar));
} }

// Original macro:
// #define Cbs0_QueForEachEntry( Que, pObj, i )                    \
//     for ( i = (Que).iHead; (i < (Que).iTail) && ((pObj) = (Que).pData[i]); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets default values of the parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_SetDefaultParams(pPars *Cbs0_Par_t) {
    *pPars = Cbs0_Par_t{};
    pPars.nBTLimit    =  1000;   // limit on the number of conflicts
    pPars.nJustLimit  =   100;   // limit on the size of justification queue
    pPars.fUseHighest =     1;   // use node with the highest ID
    pPars.fUseLowest  =     0;   // use node with the highest ID
    pPars.fUseMaxFF   =     0;   // use node with the largest fanin fanout
    pPars.fVerbose    =     1;   // print detailed statistics
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManAlloc() *Cbs0_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Cbs0_Man_t{};
    p.pProp.nSize = p.pJust.nSize = 10000;
    // inlining of macro ABC_ALLOC
    p.pProp.pData = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.pProp.nSize)));
    // inlining of macro ABC_ALLOC
    p.pJust.pData = ((Gia_Obj_t * *) malloc(sizeof(Gia_Obj_t *) * (size_t)(p.pJust.nSize)));
    p.vModel = Vec_IntAlloc( 1000 );
    Cbs0_SetDefaultParams( &p.Pars );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManStop(p *Cbs0_Man_t) {
    Vec_IntFree( p.vModel );
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
func Cbs0_ReadModel(p *Cbs0_Man_t) *Vec_Int_t {
    return p.vModel;
}




/**Function*************************************************************

  Synopsis    [Returns 1 if the solver is out of limits.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManCheckLimits(p *Cbs0_Man_t) int32 {
    return p.Pars.nJustThis > p.Pars.nJustLimit || p.Pars.nBTThis > p.Pars.nBTLimit;
}

/**Function*************************************************************

  Synopsis    [Saves the satisfying assignment as an array of literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManSaveModel(p *Cbs0_Man_t, vCex *Vec_Int_t) {
    var pVar *Gia_Obj_t;
    var i int32;
    Vec_IntClear( vCex );
    p.pProp.iHead = 0;
    // inlining of macro Cbs0_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
        if Gia_ObjIsCi(pVar) {
//            Vec_IntPush( vCex, Abc_Var2Lit(Gia_ObjId(p->pAig,pVar), !Cbs0_VarValue(pVar)) );
            Vec_IntPush( vCex, Abc_Var2Lit(Gia_ObjCioId(pVar), !Cbs0_VarValue(pVar)) ); }
    }
} 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_QueIsEmpty(p *Cbs0_Que_t) int32 {
    return p.iHead == p.iTail;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_QuePush(p *Cbs0_Que_t, pObj *Gia_Obj_t) {
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
func Cbs0_QueHasNode(p *Cbs0_Que_t, pObj *Gia_Obj_t) int32 {
    var pTemp *Gia_Obj_t;
    var i int32;
    // inlining of macro Cbs0_QueForEachEntry
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
func Cbs0_QueStore(p *Cbs0_Que_t, piHeadOld *int32, piTailOld *int32) {
    var i int32;
    *piHeadOld = p.iHead;
    *piTailOld = p.iTail;
    for  i = *piHeadOld; i < *piTailOld; i++  {
        Cbs0_QuePush( p, p.pData[i] );
    }
    p.iHead = *piTailOld;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_QueRestore(p *Cbs0_Que_t, iHeadOld int32, iTailOld int32) {
    p.iHead = iHeadOld;
    p.iTail = iTailOld;
}


/**Function*************************************************************

  Synopsis    [Max number of fanins fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_VarFaninFanoutMax(p *Cbs0_Man_t, pObj *Gia_Obj_t) int32 {
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
func Cbs0_ManDecideHighest(p *Cbs0_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i int32;
    // inlining of macro Cbs0_QueForEachEntry
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
func Cbs0_ManDecideLowest(p *Cbs0_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMin *Gia_Obj_t = nil;
    var i int32;
    // inlining of macro Cbs0_QueForEachEntry
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
func Cbs0_ManDecideMaxFF(p *Cbs0_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i, iMaxFF, iCurFF int32;
    assert(p.pAig.Refs != nil, "p.pAig.Refs != nil");
    // inlining of macro Cbs0_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pObj) = (p.pJust).pData[i]); i++  {
        iCurFF = Cbs0_VarFaninFanoutMax( p, pObj );
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
func Cbs0_ManCancelUntil(p *Cbs0_Man_t, iBound int32) {
    var pVar *Gia_Obj_t;
    var i int32;
    assert(iBound <= p.pProp.iTail, "iBound <= p.pProp.iTail");
    p.pProp.iHead = iBound;
    // inlining of macro Cbs0_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
        Cbs0_VarUnassign( pVar );
    }
    p.pProp.iTail = iBound;
}

/**Function*************************************************************

  Synopsis    [Assigns the variables a value.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManAssign(p *Cbs0_Man_t, pObj *Gia_Obj_t) {
    var pObjR *Gia_Obj_t  = Gia_Regular(pObj);
    assert(Gia_ObjIsCand(pObjR), "Gia_ObjIsCand(pObjR)");
    assert(!(Cbs0_VarIsAssigned(pObjR) != 0), "!Cbs0_VarIsAssigned(pObjR)");
    Cbs0_VarAssign( pObjR );
    Cbs0_VarSetValue( pObjR, !Gia_IsComplement(pObj) );
    Cbs0_QuePush( &p.pProp, pObjR );
}

/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManPropagateOne(p *Cbs0_Man_t, pVar *Gia_Obj_t) int32 {
    var Value0, Value1 int32;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert((Cbs0_VarIsAssigned(pVar) != 0), "Cbs0_VarIsAssigned(pVar)");
    if Gia_ObjIsCi(pVar) {
        return 0; }
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    Value0 = Cbs0_VarFanin0Value(pVar);
    Value1 = Cbs0_VarFanin1Value(pVar);
    if Cbs0_VarValue(pVar) != 0 { // value is 1
        if Value0 == 0 || Value1 == 0 { // one is 0
            return 1; }
        if Value0 == 2 { // first is unassigned
            Cbs0_ManAssign( p, Gia_ObjChild0(pVar) ); }
        if Value1 == 2 { // first is unassigned
            Cbs0_ManAssign( p, Gia_ObjChild1(pVar) ); }
        return 0;
    }
    // value is 0
    if Value0 == 0 || Value1 == 0 { // one is 0
        return 0; }
    if Value0 == 1 && Value1 == 1 { // both are 1
        return 1; }
    if Value0 == 1 || Value1 == 1 { // one is 1 
    
        if Value0 == 2 { // first is unassigned
            Cbs0_ManAssign( p, Gia_Not(Gia_ObjChild0(pVar)) ); }
        if Value1 == 2 { // first is unassigned
            Cbs0_ManAssign( p, Gia_Not(Gia_ObjChild1(pVar)) ); }
        return 0;
    }
    assert((Cbs0_VarIsJust(pVar) != 0), "Cbs0_VarIsJust(pVar)");
    assert(!(Cbs0_QueHasNode( &p.pJust, pVar ) != 0), "!Cbs0_QueHasNode( &p.pJust, pVar )");
    Cbs0_QuePush( &p.pJust, pVar );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManPropagateTwo(p *Cbs0_Man_t, pVar *Gia_Obj_t) int32 {
    var Value0, Value1 int32;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    assert((Cbs0_VarIsAssigned(pVar) != 0), "Cbs0_VarIsAssigned(pVar)");
    assert(!(Cbs0_VarValue(pVar) != 0), "!Cbs0_VarValue(pVar)");
    Value0 = Cbs0_VarFanin0Value(pVar);
    Value1 = Cbs0_VarFanin1Value(pVar);
    // value is 0
    if Value0 == 0 || Value1 == 0 { // one is 0
        return 0; }
    if Value0 == 1 && Value1 == 1 { // both are 1
        return 1; }
    assert(Value0 == 1 || Value1 == 1, "Value0 == 1 || Value1 == 1");
    if Value0 == 2 { // first is unassigned
        Cbs0_ManAssign( p, Gia_Not(Gia_ObjChild0(pVar)) ); }
    if Value1 == 2 { // first is unassigned
        Cbs0_ManAssign( p, Gia_Not(Gia_ObjChild1(pVar)) ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Propagates all variables.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManPropagate(p *Cbs0_Man_t) int32 {
    var pVar *Gia_Obj_t;
    var i, k int32;
    for {
        // inlining of macro Cbs0_QueForEachEntry
        for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((pVar) = (p.pProp).pData[i]); i++  {
            if Cbs0_ManPropagateOne( p, pVar ) != 0 {
                return 1; }
        }
        p.pProp.iHead = p.pProp.iTail;
        k = p.pJust.iHead;
        // inlining of macro Cbs0_QueForEachEntry
        for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((pVar) = (p.pJust).pData[i]); i++  {
            if Cbs0_VarIsJust( pVar ) != 0 {
                p.pJust.pData[k++] = pVar;
            } else if Cbs0_ManPropagateTwo( p, pVar ) != 0 {
                return 1; }
        }
        if k == p.pJust.iTail {
            break; }
        p.pJust.iTail = k;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Solve the problem recursively.]

  Description [Returns 1 if unsat or undecided; 0 if satisfiable.]
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Cbs0_ManSolve_rec(p *Cbs0_Man_t) int32 {
    var pVar *Gia_Obj_t = nil;
    var pDecVar *Gia_Obj_t;
    var iPropHead, iJustHead, iJustTail int32;
    // propagate assignments
    assert(!(Cbs0_QueIsEmpty(&p.pProp) != 0), "!Cbs0_QueIsEmpty(&p.pProp)");
    if Cbs0_ManPropagate( p ) != 0 {
        return 1; }
    // check for satisfying assignment
    assert((Cbs0_QueIsEmpty(&p.pProp) != 0), "Cbs0_QueIsEmpty(&p.pProp)");
    if Cbs0_QueIsEmpty(&p.pJust) != 0 {
        return 0; }
    // quit using resource limits
    p.Pars.nJustThis = Abc_MaxInt( p.Pars.nJustThis, p.pJust.iTail - p.pJust.iHead );
    if Cbs0_ManCheckLimits( p ) != 0 {
        return 0; }
    // remember the state before branching
    iPropHead = p.pProp.iHead;
    Cbs0_QueStore( &p.pJust, &iJustHead, &iJustTail );
    // find the decision variable
    if p.Pars.fUseHighest {
        pVar = Cbs0_ManDecideHighest( p );
    } else if p.Pars.fUseLowest {
        pVar = Cbs0_ManDecideLowest( p );
    } else if p.Pars.fUseMaxFF {
        pVar = Cbs0_ManDecideMaxFF( p );
    } else { assert(false, "0"); }
    assert((Cbs0_VarIsJust( pVar ) != 0), "Cbs0_VarIsJust( pVar )");
    // chose decision variable using fanout count
    if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(pVar)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(pVar)) {
        pDecVar = Gia_Not(Gia_ObjChild0(pVar));
    } else {
        pDecVar = Gia_Not(Gia_ObjChild1(pVar)); }
    // decide on first fanin
    Cbs0_ManAssign( p, pDecVar );
    if !(Cbs0_ManSolve_rec( p ) != 0) {
        return 0; }
    Cbs0_ManCancelUntil( p, iPropHead );
    Cbs0_QueRestore( &p.pJust, iJustHead, iJustTail );
    // decide on second fanin
    Cbs0_ManAssign( p, Gia_Not(pDecVar) );
    if !(Cbs0_ManSolve_rec( p ) != 0) {
        return 0; }
    p.Pars.nBTThis++;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Looking for a satisfying assignment of the node.]

  Description [Assumes that each node has flag pObj->fMark0 set to 0.
  Returns 1 if unsatisfiable, 0 if satisfiable, and -1 if undecided.
  The node may be complemented. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManSolve(p *Cbs0_Man_t, pObj *Gia_Obj_t) int32 {
    var RetValue int32;
    assert(!p.pProp.iHead && !p.pProp.iTail, "!p.pProp.iHead && !p.pProp.iTail");
    assert(!p.pJust.iHead && !p.pJust.iTail, "!p.pJust.iHead && !p.pJust.iTail");
    p.Pars.nBTThis = p.Pars.nJustThis = 0;
    Cbs0_ManAssign( p, pObj );
    RetValue = Cbs0_ManSolve_rec( p );
    if RetValue == 0 && !(Cbs0_ManCheckLimits(p) != 0) {
        Cbs0_ManSaveModel( p, p.vModel ); }
    Cbs0_ManCancelUntil( p, 0 );
    p.pJust.iHead = p.pJust.iTail = 0;
    p.Pars.nBTTotal += p.Pars.nBTThis;
    p.Pars.nJustTotal = Abc_MaxInt( p.Pars.nJustTotal, p.Pars.nJustThis );
    if Cbs0_ManCheckLimits( p ) != 0 {
        RetValue = -1; }
//    printf( "Outcome = %2d.  Confs = %6d.  Decision level max = %3d.\n", 
//        RetValue, p->Pars.nBTThis, p->DecLevelMax );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Prints statistics of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs0_ManSatPrintStats(p *Cbs0_Man_t) {
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
func Cbs_ManSolveMiter(pAig *Gia_Man_t, nConfs int32, pvStatus **Vec_Str_t, fVerbose int32) *Vec_Int_t {
    extern void Cec_ManSatAddToStore( Vec_Int_t * vCexStore, Vec_Int_t * vCex, int Out );
    var p *Cbs0_Man_t; 
    var vCex, vVisit, vCexStore *Vec_Int_t;
    var vStatus *Vec_Str_t;
    var pRoot *Gia_Obj_t; 
    var i, status int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    assert(Gia_ManRegNum(pAig) == 0, "Gia_ManRegNum(pAig) == 0");
    // prepare AIG
    Gia_ManCreateRefs( pAig );
    Gia_ManCleanMark0( pAig );
    Gia_ManCleanMark1( pAig );
    // create logic network
    p = Cbs0_ManAlloc();
    p.Pars.nBTLimit = nConfs;
    p.pAig   = pAig;
    // create resulting data-structures
    vStatus   = Vec_StrAlloc( Gia_ManPoNum(pAig) );
    vCexStore = Vec_IntAlloc( 10000 );
    vVisit    = Vec_IntAlloc( 100 );
    vCex      = Cbs0_ReadModel( p );
    // solve for each output
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pAig.vCos)) && ((pRoot) = Gia_ManCo(pAig, i)); i++  {
        Vec_IntClear( vCex );
        if Gia_ObjIsConst0(Gia_ObjFanin0(pRoot)) {
            if Gia_ObjFaninC0(pRoot) {
                fmt.Printf("Constant 1 output of SRM!!!\n");
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
        status = Cbs0_ManSolve( p, Gia_ObjChild0(pRoot) );
/*
        if ( status == -1 )
        {
            p->Pars.fUseHighest = 0;
            p->Pars.fUseLowest  = 1;
            status = Cbs0_ManSolve( p, Gia_ObjChild0(pRoot) );
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
        Cbs0_ManSatPrintStats( p ); }
    Cbs0_ManStop( p );
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

