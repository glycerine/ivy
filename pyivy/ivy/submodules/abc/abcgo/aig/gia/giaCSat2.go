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

   
type Cbs2_Par_t struct {
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
    fUseFanout           int32;   // use node with the largest fanin fanout
    // other
    fVerbose           int32;
}

   
type Cbs2_Que_t struct {
    iHead           int32;        // beginning of the queue
    iTail           int32;        // end of the queue
    nSize           int32;        // allocated size
    pData          *int32;        // nodes stored in the queue
}
 
   
type Cbs2_Man_t struct {
    Pars    Cbs2_Par_t;         // parameters
    pAig    *Gia_Man_t;         // AIG manager
    pProp    Cbs2_Que_t;        // propagation queue
    pJust    Cbs2_Que_t;        // justification queue
    pClauses    Cbs2_Que_t;     // clause queue
    vModel    *Vec_Int_t;       // satisfying assignment
    vTemp    *Vec_Int_t;        // temporary storage
    // internal data
    vAssign     Vec_Str_t;
    vMark     Vec_Str_t;
    vLevReason     Vec_Int_t;
    vWatches     Vec_Int_t;
    vWatchUpds     Vec_Int_t;
    vFanoutN     Vec_Int_t;
    vFanout0     Vec_Int_t;
    vActivity     Vec_Int_t;
    vActStore     Vec_Int_t;
    vJStore     Vec_Int_t;
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
    timeJFront       abctime;
    timeSatUnsat       abctime; // unsat
    timeSatSat       abctime;   // sat
    timeSatUndec       abctime; // undecided
    timeTotal       abctime;    // total runtime
    // other statistics
    nPropCalls [3]int32;
    nFails [2]int32;
    nClauseConf           int32;
}

func Cbs2_VarUnused(p *Cbs2_Man_t, iVar int32) int32 { return Vec_IntEntry(&p.vLevReason, 3*iVar) == -1; }
func Cbs2_VarSetUnused(p *Cbs2_Man_t, iVar int32) { Vec_IntWriteEntry(&p.vLevReason, 3*iVar, -1);     } 

/*
static inline int   Cbs2_VarMark0( Cbs2_Man_t * p, int iVar )                      { return Vec_StrEntry(&p->vAssign, iVar);            }
static inline void  Cbs2_VarSetMark0( Cbs2_Man_t * p, int iVar, int Value )        { Vec_StrWriteEntry(&p->vAssign, iVar, (char)Value); } 

static inline int   Cbs2_VarMark1( Cbs2_Man_t * p, int iVar )                      { return Vec_StrEntry(&p->vValue, iVar);             }
static inline void  Cbs2_VarSetMark1( Cbs2_Man_t * p, int iVar, int Value )        { Vec_StrWriteEntry(&p->vValue, iVar, (char)Value);  } 

static inline int   Cbs2_VarIsAssigned( Cbs2_Man_t * p, int iVar )                 { return Cbs2_VarMark0(p, iVar);                                                                  }
static inline void  Cbs2_VarAssign( Cbs2_Man_t * p, int iVar )                     { assert(!Cbs2_VarIsAssigned(p, iVar)); Cbs2_VarSetMark0(p, iVar, 1);                             }
static inline void  Cbs2_VarUnassign( Cbs2_Man_t * p, int iVar )                   { assert( Cbs2_VarIsAssigned(p, iVar)); Cbs2_VarSetMark0(p, iVar, 0); Cbs2_VarSetUnused(p, iVar); }
static inline int   Cbs2_VarValue( Cbs2_Man_t * p, int iVar )                      { assert( Cbs2_VarIsAssigned(p, iVar)); return Cbs2_VarMark1(p, iVar);                            }
static inline void  Cbs2_VarSetValue( Cbs2_Man_t * p, int iVar, int v )            { assert( Cbs2_VarIsAssigned(p, iVar)); Cbs2_VarSetMark1(p, iVar, v);                             }

static inline int   Cbs2_VarIsJust( Cbs2_Man_t * p, Gia_Obj_t * pVar, int iVar )      { return Gia_ObjIsAnd(pVar) && !Cbs2_VarIsAssigned(p, Gia_ObjFaninId0(pVar, iVar)) && !Cbs2_VarIsAssigned(p, Gia_ObjFaninId1(pVar, iVar));  } 
static inline int   Cbs2_VarFanin0Value( Cbs2_Man_t * p, Gia_Obj_t * pVar, int iVar ) { return Cbs2_VarIsAssigned(p, Gia_ObjFaninId0(pVar, iVar)) ? (Cbs2_VarValue(p, Gia_ObjFaninId0(pVar, iVar)) ^ Gia_ObjFaninC0(pVar)) : 2;   }
static inline int   Cbs2_VarFanin1Value( Cbs2_Man_t * p, Gia_Obj_t * pVar, int iVar ) { return Cbs2_VarIsAssigned(p, Gia_ObjFaninId1(pVar, iVar)) ? (Cbs2_VarValue(p, Gia_ObjFaninId1(pVar, iVar)) ^ Gia_ObjFaninC1(pVar)) : 2;   }
*/

func Cbs2_VarMark0(p *Cbs2_Man_t, iVar int32) int32 { return Vec_StrEntry(&p.vMark, iVar);              }
func Cbs2_VarSetMark0(p *Cbs2_Man_t, iVar int32, Value int32) { Vec_StrWriteEntry(&p.vMark, iVar, byte(Value));   } 

//static inline int   Cbs2_VarMark0( Cbs2_Man_t * p, int iVar )                      { return Vec_StrEntry(&p->vAssign, iVar) >= 2;       }
//static inline void  Cbs2_VarSetMark0( Cbs2_Man_t * p, int iVar, int Value )        { Vec_StrWriteEntry(&p->vAssign, iVar, (char)(2^Vec_StrEntry(&p->vAssign, iVar)));   } 

func Cbs2_VarIsAssigned(p *Cbs2_Man_t, iVar int32) int32 { return Vec_StrEntry(&p.vAssign, iVar) < 2;        }
func Cbs2_VarUnassign(p *Cbs2_Man_t, iVar int32) { assert((Cbs2_VarIsAssigned(p, iVar) != 0), "Cbs2_VarIsAssigned(p, iVar)"); Vec_StrWriteEntry(&p.vAssign, iVar, byte((2+Vec_StrEntry(&p.vAssign, iVar)))); Cbs2_VarSetUnused(p, iVar); }

func Cbs2_VarValue(p *Cbs2_Man_t, iVar int32) int32 { return Vec_StrEntry(&p.vAssign, iVar);            }
func Cbs2_VarSetValue(p *Cbs2_Man_t, iVar int32, v int32) { assert(!(Cbs2_VarIsAssigned(p, iVar) != 0), "!Cbs2_VarIsAssigned(p, iVar)"); Vec_StrWriteEntry(&p.vAssign, iVar, byte(v));                                  }

func Cbs2_VarIsJust(p *Cbs2_Man_t, pVar *Gia_Obj_t, iVar int32) int32 { return Gia_ObjIsAnd(pVar) && !Cbs2_VarIsAssigned(p, Gia_ObjFaninId0(pVar, iVar)) && !Cbs2_VarIsAssigned(p, Gia_ObjFaninId1(pVar, iVar));  } 
func Cbs2_VarFanin0Value(p *Cbs2_Man_t, pVar *Gia_Obj_t, iVar int32) int32 { return Cbs2_VarValue(p, Gia_ObjFaninId0(pVar, iVar)) ^ Gia_ObjFaninC0(pVar);   }
func Cbs2_VarFanin1Value(p *Cbs2_Man_t, pVar *Gia_Obj_t, iVar int32) int32 { return Cbs2_VarValue(p, Gia_ObjFaninId1(pVar, iVar)) ^ Gia_ObjFaninC1(pVar);   }


func Cbs2_VarDecLevel(p *Cbs2_Man_t, iVar int32) int32 { assert(!(Cbs2_VarUnused(p, iVar) != 0), "!Cbs2_VarUnused(p, iVar)"); return Vec_IntEntry(&p.vLevReason, 3*iVar);    }
func Cbs2_VarReason0(p *Cbs2_Man_t, iVar int32) int32 { assert(!(Cbs2_VarUnused(p, iVar) != 0), "!Cbs2_VarUnused(p, iVar)"); return Vec_IntEntry(&p.vLevReason, 3*iVar+1);  }
func Cbs2_VarReason1(p *Cbs2_Man_t, iVar int32) int32 { assert(!(Cbs2_VarUnused(p, iVar) != 0), "!Cbs2_VarUnused(p, iVar)"); return Vec_IntEntry(&p.vLevReason, 3*iVar+2);  }
func Cbs2_VarReasonP(p *Cbs2_Man_t, iVar int32) *int32 { assert(!(Cbs2_VarUnused(p, iVar) != 0), "!Cbs2_VarUnused(p, iVar)"); return Vec_IntEntryP(&p.vLevReason, 3*iVar+1); }
//static inline int   Cbs2_ClauseDecLevel( Cbs2_Man_t * p, int hClause )             { return Cbs2_VarDecLevel( p, p->pClauses.pData[hClause] );                           }

func Cbs2_ClauseSize(p *Cbs2_Man_t, hClause int32) int32 { return p.pClauses.pData[hClause];                                                  }
func Cbs2_ClauseLits(p *Cbs2_Man_t, hClause int32) *int32 { return p.pClauses.pData+hClause+1;                                                 }
func Cbs2_ClauseLit(p *Cbs2_Man_t, hClause int32, i int32) int32 { return p.pClauses.pData[hClause+1+i];                                              }
func Cbs2_ClauseNext1p(p *Cbs2_Man_t, hClause int32) *int32 { return p.pClauses.pData+hClause+Cbs2_ClauseSize(p, hClause)+2;                     }

func Cbs2_ClauseSetSize(p *Cbs2_Man_t, hClause int32, x int32) { p.pClauses.pData[hClause] = x;                                                     }
func Cbs2_ClauseSetLit(p *Cbs2_Man_t, hClause int32, i int32, x int32) { p.pClauses.pData[hClause+i+1] = x;                                                 }
func Cbs2_ClauseSetNext(p *Cbs2_Man_t, hClause int32, n int32, x int32) { p.pClauses.pData[hClause+Cbs2_ClauseSize(p, hClause)+1+n] = x;                     }


// Original macro:
// #define Cbs2_QueForEachEntry( Que, iObj, i )                         \
//     for ( i = (Que).iHead; (i < (Que).iTail) && ((iObj) = (Que).pData[i]); i++ )


// Original macro:
// #define Cbs2_ClauseForEachEntry( p, hClause, iObj, i )               \
//     for ( i = 1; i <= Cbs2_ClauseSize(p, hClause) && (iObj = (p)->pClauses.pData[hClause+i]); i++ )

// Original macro:
// #define Cbs2_ClauseForEachEntry1( p, hClause, iObj, i )                   \
//     for ( i = 2; i <= Cbs2_ClauseSize(p, hClause) && (iObj = (p)->pClauses.pData[hClause+i]); i++ )


// Original macro:
// #define Cbs2_ObjForEachFanout( p, iObj, iFanLit )                      \
//     for ( iFanLit = Vec_IntEntry(&p->vFanout0, iObj); iFanLit; iFanLit = Vec_IntEntry(&p->vFanoutN, iFanLit) )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets default values of the parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_SetDefaultParams(pPars *Cbs2_Par_t) {
    *pPars = Cbs2_Par_t{};
    pPars.nBTLimit    =  1000;   // limit on the number of conflicts
    pPars.nJustLimit  =   500;   // limit on the size of justification queue
    pPars.fUseHighest =     1;   // use node with the highest ID
    pPars.fUseLowest  =     0;   // use node with the highest ID
    pPars.fUseMaxFF   =     0;   // use node with the largest fanin fanout
    pPars.fUseFanout  =     1;
    pPars.fVerbose    =     1;   // print detailed statistics
}
func Cbs2_ManSetConflictNum(p *Cbs2_Man_t, Num int32) {
    p.Pars.nBTLimit = Num;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManAlloc(pGia *Gia_Man_t) *Cbs2_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Cbs2_Man_t{};
    p.pProp.nSize = p.pJust.nSize = p.pClauses.nSize = 10000;
    // inlining of macro ABC_ALLOC
    p.pProp.pData = ((int *) malloc(sizeof(int) * (size_t)(p.pProp.nSize)));
    // inlining of macro ABC_ALLOC
    p.pJust.pData = ((int *) malloc(sizeof(int) * (size_t)(p.pJust.nSize)));
    // inlining of macro ABC_ALLOC
    p.pClauses.pData = ((int *) malloc(sizeof(int) * (size_t)(p.pClauses.nSize)));
    p.pClauses.iHead = p.pClauses.iTail = 1;
    p.vModel   = Vec_IntAlloc( 1000 );
    p.vTemp    = Vec_IntAlloc( 1000 );
    p.pAig     = pGia;
    Cbs2_SetDefaultParams( &p.Pars );
    Vec_StrFill( &p.vAssign,    Gia_ManObjNum(pGia), 2 );
    Vec_StrFill( &p.vMark,      Gia_ManObjNum(pGia), 0 );
    Vec_IntFill( &p.vLevReason, 3*Gia_ManObjNum(pGia), -1 );
    Vec_IntFill( &p.vWatches,   2*Gia_ManObjNum(pGia), 0 );
    Vec_IntFill( &p.vFanout0,   Gia_ManObjNum(pGia),   0 );
    Vec_IntFill( &p.vFanoutN,   2*Gia_ManObjNum(pGia), 0 );
    Vec_IntFill( &p.vActivity,  Gia_ManObjNum(pGia), 0 );
    Vec_IntGrow( &p.vActStore,  1000 );
    Vec_IntGrow( &p.vJStore,    1000 );
    Vec_IntGrow( &p.vWatchUpds, 1000 );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManStop(p *Cbs2_Man_t) {
    Vec_StrErase( &p.vAssign );
    Vec_StrErase( &p.vMark );
    Vec_IntErase( &p.vLevReason );
    Vec_IntErase( &p.vWatches );
    Vec_IntErase( &p.vFanout0 );
    Vec_IntErase( &p.vFanoutN );
    Vec_IntErase( &p.vActivity );
    Vec_IntErase( &p.vActStore );
    Vec_IntErase( &p.vJStore );
    Vec_IntErase( &p.vWatchUpds );
    Vec_IntFree( p.vModel );
    Vec_IntFree( p.vTemp );
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
func Cbs2_ReadModel(p *Cbs2_Man_t) *Vec_Int_t {
    return p.vModel;
}




/**Function*************************************************************

  Synopsis    [Returns 1 if the solver is out of limits.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManCheckLimits(p *Cbs2_Man_t) int32 {
    p.nFails[0] += p.Pars.nJustThis > p.Pars.nJustLimit;
    p.nFails[1] += p.Pars.nBTThis > p.Pars.nBTLimit;
    return p.Pars.nJustThis > p.Pars.nJustLimit || p.Pars.nBTThis > p.Pars.nBTLimit;
}

/**Function*************************************************************

  Synopsis    [Saves the satisfying assignment as an array of literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManSaveModel(p *Cbs2_Man_t, vCex *Vec_Int_t) {
    var i, iLit int32;
    Vec_IntClear( vCex );
    p.pProp.iHead = 0;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((iLit) = (p.pProp).pData[i]); i++  {
        var iVar int32  = Abc_Lit2Var(iLit);
        if Gia_ObjIsCi(Gia_ManObj(p.pAig, iVar)) {
            Vec_IntPush( vCex, Abc_Var2Lit(Gia_ManIdToCioId(p.pAig, iVar), !Cbs2_VarValue(p, iVar)) ); }
    }
} 
func Cbs2_ManSaveModelAll(p *Cbs2_Man_t, vCex *Vec_Int_t) {
    var i, iLit int32;
    Vec_IntClear( vCex );
    p.pProp.iHead = 0;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((iLit) = (p.pProp).pData[i]); i++  {
        var iVar int32  = Abc_Lit2Var(iLit);
        Vec_IntPush( vCex, Abc_Var2Lit(iVar, !Cbs2_VarValue(p, iVar)) );
    }
} 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_QueIsEmpty(p *Cbs2_Que_t) int32 {
    return p.iHead == p.iTail;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_QuePush(p *Cbs2_Que_t, iObj int32) {
    if p.iTail == p.nSize {
        p.nSize *= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pData {
            p.pData = ((int *) realloc((char *)(p.pData), sizeof(int) * (size_t)(p.nSize)));
        } else {
            p.pData = ((int *) malloc(sizeof(int) * (size_t)(p.nSize)));
        } 
    }
    p.pData[p.iTail++] = iObj;
}
func Cbs2_QueGrow(p *Cbs2_Que_t, Plus int32) {
    if p.iTail + Plus > p.nSize {
        p.nSize *= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pData {
            p.pData = ((int *) realloc((char *)(p.pData), sizeof(int) * (size_t)(p.nSize)));
        } else {
            p.pData = ((int *) malloc(sizeof(int) * (size_t)(p.nSize)));
        } 
    }
    assert(p.iTail + Plus <= p.nSize, "p.iTail + Plus <= p.nSize");
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the object in the queue.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_QueHasNode(p *Cbs2_Que_t, iObj int32) int32 {
    var i, iTemp int32;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (*p).iHead; (i < (*p).iTail) && ((iTemp) = (*p).pData[i]); i++  {
        if iTemp == iObj {
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
func Cbs2_QueStore(p *Cbs2_Que_t, piHeadOld *int32, piTailOld *int32) {
    var i int32;
    *piHeadOld = p.iHead;
    *piTailOld = p.iTail;
    for  i = *piHeadOld; i < *piTailOld; i++  {
        Cbs2_QuePush( p, p.pData[i] );
    }
    p.iHead = *piTailOld;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_QueRestore(p *Cbs2_Que_t, iHeadOld int32, iTailOld int32) {
    p.iHead = iHeadOld;
    p.iTail = iTailOld;
}


/**Function*************************************************************

  Synopsis    [Max number of fanins fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_VarFaninFanoutMax(p *Cbs2_Man_t, pObj *Gia_Obj_t) int32 {
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
func Cbs2_ManDecideHighest(p *Cbs2_Man_t) int32 {
    var i, iObj, iObjMax int32;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((iObj) = (p.pJust).pData[i]); i++  {
        if iObjMax == 0 || iObjMax < iObj {
            iObjMax = iObj; }
    }
    return iObjMax;
}

/**Function*************************************************************

  Synopsis    [Find variable with the lowest ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManDecideLowest(p *Cbs2_Man_t) *Gia_Obj_t {
    var i, iObj, iObjMin int32;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((iObj) = (p.pJust).pData[i]); i++  {
        if iObjMin == 0 || iObjMin > iObj {
            iObjMin = iObj; }
    }
    return Gia_ManObj(p.pAig, iObjMin);
}

/**Function*************************************************************

  Synopsis    [Find variable with the maximum number of fanin fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManDecideMaxFF(p *Cbs2_Man_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i, iMaxFF, iCurFF, iObj int32;
    assert(p.pAig.Refs != nil, "p.pAig.Refs != nil");
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((iObj) = (p.pJust).pData[i]); i++  {
        pObj = Gia_ManObj(p.pAig, iObj);
        iCurFF = Cbs2_VarFaninFanoutMax( p, pObj );
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
func Cbs2_ManCancelUntil(p *Cbs2_Man_t, iBound int32) {
    var i, iLit int32;
    assert(iBound <= p.pProp.iTail, "iBound <= p.pProp.iTail");
    p.pProp.iHead = iBound;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((iLit) = (p.pProp).pData[i]); i++  {
        Cbs2_VarUnassign( p, Abc_Lit2Var(iLit) );
    }
    p.pProp.iTail = iBound;
}

/**Function*************************************************************

  Synopsis    [Assigns the variables a value.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManAssign(p *Cbs2_Man_t, iLit int32, Level int32, iRes0 int32, iRes1 int32) {
    var iObj int32  = Abc_Lit2Var(iLit);
    assert((Cbs2_VarUnused(p, iObj) != 0), "Cbs2_VarUnused(p, iObj)");
    assert(!(Cbs2_VarIsAssigned(p, iObj) != 0), "!Cbs2_VarIsAssigned(p, iObj)");
    //Cbs2_VarAssign( p, iObj );
    Cbs2_VarSetValue( p, iObj, !Abc_LitIsCompl(iLit) );
    Cbs2_QuePush( &p.pProp, iLit );
    Vec_IntWriteEntry( &p.vLevReason, 3*iObj,   Level );
    Vec_IntWriteEntry( &p.vLevReason, 3*iObj+1, iRes0 );
    Vec_IntWriteEntry( &p.vLevReason, 3*iObj+2, iRes1 );
}









/**Function*************************************************************

  Synopsis    [Prints conflict clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManPrintClause(p *Cbs2_Man_t, Level int32, hClause int32) {
    var i, iLit int32;
    assert((Cbs2_QueIsEmpty( &p.pClauses ) != 0), "Cbs2_QueIsEmpty( &p.pClauses )");
    fmt.Printf("Level %2d : ", Level);
    // inlining of macro Cbs2_ClauseForEachEntry
    for  i = 1; i <= Cbs2_ClauseSize(p, hClause) && (iLit = (p).pClauses.pData[hClause+i]); i++  {
        fmt.Printf("%c%d ", Abc_LitIsCompl(iLit) ? '-':'+', Abc_Lit2Var(iLit));
    }
//        printf( "%d=%d(%d) ", iObj, Cbs2_VarValue(p, Abc_Lit2Var(iLit)), Cbs2_VarDecLevel(p, Abc_Lit2Var(iLit)) );
    fmt.Printf("\n");
}
func Cbs2_ManPrintCube(p *Cbs2_Man_t, Level int32, hClause int32) {
    var i, iObj int32;
    assert((Cbs2_QueIsEmpty( &p.pClauses ) != 0), "Cbs2_QueIsEmpty( &p.pClauses )");
    fmt.Printf("Level %2d : ", Level);
    // inlining of macro Cbs2_ClauseForEachEntry
    for  i = 1; i <= Cbs2_ClauseSize(p, hClause) && (iObj = (p).pClauses.pData[hClause+i]); i++  {
        fmt.Printf("%c%d ", Cbs2_VarValue(p, iObj)? '+':'-', iObj);
    }
    fmt.Printf("\n");
}
func Cbs2_ManBumpClause(p *Cbs2_Man_t, hClause int32) {
    var i, iObj int32;
    assert((Cbs2_QueIsEmpty( &p.pClauses ) != 0), "Cbs2_QueIsEmpty( &p.pClauses )");
    // inlining of macro Cbs2_ClauseForEachEntry
    for  i = 1; i <= Cbs2_ClauseSize(p, hClause) && (iObj = (p).pClauses.pData[hClause+i]); i++  {
        //if ( Vec_IntEntry(&p->vActivity, iObj) == 0 )
        //    Vec_IntPush( &p->vActStore, iObj );
        //Vec_IntAddToEntry( &p->vActivity, iObj, 1 );
    }
}
func Cbs2_ManBumpClean(p *Cbs2_Man_t) {
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&p.vActStore)) && (((iObj) = Vec_IntEntry(&p.vActStore, i)), 1); i++  {
        Vec_IntWriteEntry( &p.vActivity, iObj, 0 );
    }
}

/**Function*************************************************************

  Synopsis    [Finalized the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManCleanWatch(p *Cbs2_Man_t) {
    var i, iLit int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&p.vWatchUpds)) && (((iLit) = Vec_IntEntry(&p.vWatchUpds, i)), 1); i++  {
        Vec_IntWriteEntry( &p.vWatches, iLit, 0 );
    }
    Vec_IntClear( &p.vWatchUpds );
    //Vec_IntForEachEntry( &p->vWatches, iLit, i )
    //    assert( iLit == 0 );
}
func Cbs2_ManWatchClause(p *Cbs2_Man_t, hClause int32, Lit int32) {
    var pLits *int32  = Cbs2_ClauseLits( p, hClause );
    var pPlace *int32  = Vec_IntEntryP( &p.vWatches, Abc_LitNot(Lit) );
    if *pPlace == 0 {
        Vec_IntPush( &p.vWatchUpds, Abc_LitNot(Lit) ); }
/*
    if ( pClause->pLits[0] == Lit )
        pClause->pNext0 = p->pWatches[lit_neg(Lit)];  
    else
    {
        assert( pClause->pLits[1] == Lit );
        pClause->pNext1 = p->pWatches[lit_neg(Lit)];  
    }
    p->pWatches[lit_neg(Lit)] = pClause;
*/
    assert(Lit == pLits[0] || Lit == pLits[1], "Lit == pLits[0] || Lit == pLits[1]");
    Cbs2_ClauseSetNext( p, hClause, Lit == pLits[1], *pPlace );
    *pPlace = hClause;
}
func Cbs2_QueFinish(p *Cbs2_Man_t, Level int32) int32 {
    var pQue *Cbs2_Que_t  = &(p.pClauses); 
    var i int32;
    var iObj int32;
    var hClauseC int32;
    var hClause int32 = pQue.iHead;
    var Size int32 = pQue.iTail - pQue.iHead - 1;
    assert(pQue.iHead+1 < pQue.iTail, "pQue.iHead+1 < pQue.iTail");
    Cbs2_ClauseSetSize( p, pQue.iHead, Size );
    hClauseC = pQue.iHead = pQue.iTail;
    //printf( "Adding cube: " ); Cbs2_ManPrintCube(p, Level, hClause);
    if Size == 1 {
        return hClause; }
    // create watched clause
    pQue.iHead = hClause;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pClauses).iHead; (i < (p.pClauses).iTail) && ((iObj) = (p.pClauses).pData[i]); i++  {
        if i == hClauseC {
            break;
        } else if i == hClause { // nlits
            Cbs2_QuePush( pQue, iObj );
        } else { // literals
            Cbs2_QuePush( pQue, Abc_Var2Lit(iObj, Cbs2_VarValue(p, iObj)) ); } // complement
    }
    Cbs2_QuePush( pQue, 0 ); // next0
    Cbs2_QuePush( pQue, 0 ); // next1
    pQue.iHead = pQue.iTail;
    Cbs2_ManWatchClause( p, hClauseC, Cbs2_ClauseLit(p, hClauseC, 0) );
    Cbs2_ManWatchClause( p, hClauseC, Cbs2_ClauseLit(p, hClauseC, 1) );
    //printf( "Adding clause %d: ", hClauseC ); Cbs2_ManPrintClause(p, Level, hClauseC);
    return hClause;
}

/**Function*************************************************************

  Synopsis    [Returns conflict clause.]

  Description [Performs conflict analysis.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManDeriveReason(p *Cbs2_Man_t, Level int32) int32 {
    var pQue *Cbs2_Que_t  = &(p.pClauses);
    var i int32;
    var k int32;
    var iObj int32;
    var iLitLevel int32;
    var pReason *int32;
    assert(pQue.pData[pQue.iHead] == 0, "pQue.pData[pQue.iHead] == 0");
    assert(pQue.pData[pQue.iHead+1] == 0, "pQue.pData[pQue.iHead+1] == 0");
    assert(pQue.iHead + 2 < pQue.iTail, "pQue.iHead + 2 < pQue.iTail");
    //for ( i = pQue->iHead + 2; i < pQue->iTail; i++ )
    //    assert( !Cbs2_VarMark0(p, pQue->pData[i]) );
    // compact literals
    Vec_IntClear( p.vTemp );
    for  i = k = pQue.iHead + 2; i < pQue.iTail; i++  {
        iObj = pQue.pData[i];
        if Cbs2_VarMark0(p, iObj) != 0 { // unassigned - seen again
            continue; }
        //if ( Vec_IntEntry(&p->vActivity, iObj) == 0 )
        //    Vec_IntPush( &p->vActStore, iObj );
        //Vec_IntAddToEntry( &p->vActivity, iObj, 1 );
        // assigned - seen first time
        Cbs2_VarSetMark0(p, iObj, 1);
        Vec_IntPush( p.vTemp, iObj );
        // check decision level
        iLitLevel = Cbs2_VarDecLevel( p, iObj );
        if iLitLevel < Level {
            pQue.pData[k++] = iObj;
            continue;
        }
        assert(iLitLevel == Level, "iLitLevel == Level");
        pReason = Cbs2_VarReasonP( p, iObj );
        if pReason[0] == 0 && pReason[1] == 0 { // no reason
        
            assert(pQue.pData[pQue.iHead+1] == 0, "pQue.pData[pQue.iHead+1] == 0");
            pQue.pData[pQue.iHead+1] = iObj;
        
        } else if pReason[0] != 0 { // circuit reason
        
            Cbs2_QuePush( pQue, pReason[0] );
            if pReason[1] {
            Cbs2_QuePush( pQue, pReason[1] ); }
        
        } else { // clause reason
        
            var i int32;
            var pLits *int32;
            var nLits int32 = Cbs2_ClauseSize( p, pReason[1] );
            assert(pReason[1], "pReason[1]");
            Cbs2_QueGrow( pQue, nLits );
            pLits = Cbs2_ClauseLits( p, pReason[1] );
            assert(iObj == Abc_Lit2Var(pLits[0]), "iObj == Abc_Lit2Var(pLits[0])");
            for  i = 1; i < nLits; i++  {
                Cbs2_QuePush( pQue, Abc_Lit2Var(pLits[i]) );
            }
        }
    }
    assert(pQue.pData[pQue.iHead] == 0, "pQue.pData[pQue.iHead] == 0");
    assert(pQue.pData[pQue.iHead+1] != 0, "pQue.pData[pQue.iHead+1] != 0");
    pQue.iTail = k;
    // clear the marks
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vTemp)) && (((iObj) = Vec_IntEntry(p.vTemp, i)), 1); i++  {
        Cbs2_VarSetMark0(p, iObj, 0);
    }
    return Cbs2_QueFinish( p, Level );
}
func Cbs2_ManAnalyze(p *Cbs2_Man_t, Level int32, iVar int32, iFan0 int32, iFan1 int32) int32 {
    var pQue *Cbs2_Que_t  = &(p.pClauses); 
    assert((Cbs2_VarIsAssigned(p, iVar) != 0), "Cbs2_VarIsAssigned(p, iVar)");
    assert((Cbs2_QueIsEmpty( pQue ) != 0), "Cbs2_QueIsEmpty( pQue )");
    Cbs2_QuePush( pQue, 0 );
    Cbs2_QuePush( pQue, 0 );
    if iFan0 != 0 { // circuit conflict
    
        assert((Cbs2_VarIsAssigned(p, iFan0) != 0), "Cbs2_VarIsAssigned(p, iFan0)");
        assert(iFan1 == 0 || (Cbs2_VarIsAssigned(p, iFan1) != 0), "iFan1 == 0 || Cbs2_VarIsAssigned(p, iFan1)");
        Cbs2_QuePush( pQue, iVar );
        Cbs2_QuePush( pQue, iFan0 );
        if iFan1 != 0 {
        Cbs2_QuePush( pQue, iFan1 ); }
    
    } else { // clause conflict
    
        var i int32;
        var pLits *int32;
        var nLits int32 = Cbs2_ClauseSize( p, iFan1 );
        assert((iFan1 != 0), "iFan1");
        Cbs2_QueGrow( pQue, nLits );
        pLits = Cbs2_ClauseLits( p, iFan1 );
        assert(iVar == Abc_Lit2Var(pLits[0]), "iVar == Abc_Lit2Var(pLits[0])");
        assert(Cbs2_VarValue(p, iVar) == Abc_LitIsCompl(pLits[0]), "Cbs2_VarValue(p, iVar) == Abc_LitIsCompl(pLits[0])");
        for  i = 0; i < nLits; i++  {
            Cbs2_QuePush( pQue, Abc_Lit2Var(pLits[i]) );
        }
    }
    return Cbs2_ManDeriveReason( p, Level );
}


/**Function*************************************************************

  Synopsis    [Propagate one assignment.]

  Description [Returns handle of the conflict clause, if conflict occurs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManPropagateClauses(p *Cbs2_Man_t, Level int32, Lit int32) int32 {
    var i int32;
    var Value int32;
    var Cur int32;
    var LitF int32 = Abc_LitNot(Lit);
    var pPrev *int32  = Vec_IntEntryP( &p.vWatches, Lit );
    //for ( pCur = p->pWatches[Lit]; pCur; pCur = *ppPrev )
    for  Cur = *pPrev; (Cur != 0); Cur = *pPrev  {
        var nLits int32  = Cbs2_ClauseSize( p, Cur );
        var pLits *int32  = Cbs2_ClauseLits( p, Cur );
        p.nPropCalls[1]++;
//printf( "  Watching literal %c%d on level %d.\n", Abc_LitIsCompl(Lit) ? '-':'+', Abc_Lit2Var(Lit), Level );
        // make sure the false literal is in the second literal of the clause
        //if ( pCur->pLits[0] == LitF )
        if pLits[0] == LitF {
            //pCur->pLits[0] = pCur->pLits[1];
            pLits[0] = pLits[1];
            //pCur->pLits[1] = LitF;
            pLits[1] = LitF;
            //pTemp = pCur->pNext0;
            //pCur->pNext0 = pCur->pNext1;
            //pCur->pNext1 = pTemp;
            // inlining of macro ABC_SWAP
            { var t int32  = pLits[nLits]; pLits[nLits] = pLits[nLits+1]; pLits[nLits+1] = t; };
        }
        //assert( pCur->pLits[1] == LitF );
        assert(pLits[1] == LitF, "pLits[1] == LitF");

        // if the first literal is true, the clause is satisfied
        //if ( pCur->pLits[0] == p->pAssigns[lit_var(pCur->pLits[0])] )
        if Cbs2_VarValue(p, Abc_Lit2Var(pLits[0])) == !Abc_LitIsCompl(pLits[0]) {
            //ppPrev = &pCur->pNext1;
            pPrev = Cbs2_ClauseNext1p(p, Cur);
            continue;
        }

        // look for a new literal to watch
        for  i = 2; i < nLits; i++  {
            // skip the case when the literal is false
            //if ( lit_neg(pCur->pLits[i]) == p->pAssigns[lit_var(pCur->pLits[i])] )
            if Cbs2_VarValue(p, Abc_Lit2Var(pLits[i])) == Abc_LitIsCompl(pLits[i]) {
                continue; }
            // the literal is either true or unassigned - watch it
            //pCur->pLits[1] = pCur->pLits[i];
            //pCur->pLits[i] = LitF;
            pLits[1] = pLits[i];
            pLits[i] = LitF;
            // remove this clause from the watch list of Lit
            //*ppPrev = pCur->pNext1;
            *pPrev = *Cbs2_ClauseNext1p(p, Cur);
            // add this clause to the watch list of pCur->pLits[i] (now it is pCur->pLits[1])
            //Intb_ManWatchClause( p, pCur, pCur->pLits[1] );
            Cbs2_ManWatchClause( p, Cur, Cbs2_ClauseLit(p, Cur, 1) );
            break;
        }
        if i < nLits { // found new watch
            continue; }

        // clause is unit - enqueue new implication
        //if ( Inta_ManEnqueue(p, pCur->pLits[0], pCur) )
        //{
        //    ppPrev = &pCur->pNext1;
        //    continue;
        //}

        // clause is unit - enqueue new implication
        Value = Cbs2_VarValue(p, Abc_Lit2Var(pLits[0]));
        if Value >= 2 { // unassigned
        
            Cbs2_ManAssign( p, pLits[0], Level, 0, Cur );
            pPrev = Cbs2_ClauseNext1p(p, Cur);
            continue;
        }

        // conflict detected - return the conflict clause
        //return pCur;
        if Value == Abc_LitIsCompl(pLits[0]) {
            p.nClauseConf++;
            return Cbs2_ManAnalyze( p, Level, Abc_Lit2Var(pLits[0]), 0, Cur );
        }
    }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Performs resolution of two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManResolve(p *Cbs2_Man_t, Level int32, hClause0 int32, hClause1 int32) int32 {
    var pQue *Cbs2_Que_t  = &(p.pClauses);
    var i int32;
    var iObj int32;
    var LevelMax int32 = -1;
    var LevelCur int32;
    assert(pQue.pData[hClause0+1] != 0, "pQue.pData[hClause0+1] != 0");
    assert(pQue.pData[hClause0+1] == pQue.pData[hClause1+1], "pQue.pData[hClause0+1] == pQue.pData[hClause1+1]");
    //Cbs2_ClauseForEachEntry1( p, hClause0, iObj, i )
    //    assert( !Cbs2_VarMark0(p, iObj) );
    //Cbs2_ClauseForEachEntry1( p, hClause1, iObj, i )
    //    assert( !Cbs2_VarMark0(p, iObj) );
    assert((Cbs2_QueIsEmpty( pQue ) != 0), "Cbs2_QueIsEmpty( pQue )");
    Cbs2_QuePush( pQue, 0 );
    Cbs2_QuePush( pQue, 0 );
//    for ( i = hClause0 + 1; (iObj = pQue->pData[i]); i++ )
    // inlining of macro Cbs2_ClauseForEachEntry1
    for  i = 2; i <= Cbs2_ClauseSize(p, hClause0) && (iObj = (p).pClauses.pData[hClause0+i]); i++  {
        if Cbs2_VarMark0(p, iObj) != 0 { // unassigned - seen again
            continue; }
        //if ( Vec_IntEntry(&p->vActivity, iObj) == 0 )
        //    Vec_IntPush( &p->vActStore, iObj );
        //Vec_IntAddToEntry( &p->vActivity, iObj, 1 );
        // assigned - seen first time
        Cbs2_VarSetMark0(p, iObj, 1);
        Cbs2_QuePush( pQue, iObj );
        LevelCur = Cbs2_VarDecLevel( p, iObj );
        if LevelMax < LevelCur {
            LevelMax = LevelCur; }
    }
//    for ( i = hClause1 + 1; (iObj = pQue->pData[i]); i++ )
    // inlining of macro Cbs2_ClauseForEachEntry1
    for  i = 2; i <= Cbs2_ClauseSize(p, hClause1) && (iObj = (p).pClauses.pData[hClause1+i]); i++  {
        if Cbs2_VarMark0(p, iObj) != 0 { // unassigned - seen again
            continue; }
        //if ( Vec_IntEntry(&p->vActivity, iObj) == 0 )
        //    Vec_IntPush( &p->vActStore, iObj );
        //Vec_IntAddToEntry( &p->vActivity, iObj, 1 );
        // assigned - seen first time
        Cbs2_VarSetMark0(p, iObj, 1);
        Cbs2_QuePush( pQue, iObj );
        LevelCur = Cbs2_VarDecLevel( p, iObj );
        if LevelMax < LevelCur {
            LevelMax = LevelCur; }
    }
    for  i = pQue.iHead + 2; i < pQue.iTail; i++  {
        Cbs2_VarSetMark0(p, pQue.pData[i], 0);
    }
    return Cbs2_ManDeriveReason( p, LevelMax );
}

/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns clause handle if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManPropagateOne(p *Cbs2_Man_t, iVar int32, Level int32) int32 {
    var pVar *Gia_Obj_t  = Gia_ManObj( p.pAig, iVar ); var Value0, Value1 int32;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert((Cbs2_VarIsAssigned(p, iVar) != 0), "Cbs2_VarIsAssigned(p, iVar)");
    if Gia_ObjIsCi(pVar) {
        return 0; }
    p.nPropCalls[0]++;
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    Value0 = Cbs2_VarFanin0Value(p, pVar, iVar);
    Value1 = Cbs2_VarFanin1Value(p, pVar, iVar);
    if Cbs2_VarValue(p, iVar) == 1 { // value is 1
        if Value0 == 0 || Value1 == 0 { // one is 0
        
            if Value0 == 0 && Value1 != 0 {
                return Cbs2_ManAnalyze( p, Level, iVar, Gia_ObjFaninId0(pVar, iVar), 0 ); }
            if Value0 != 0 && Value1 == 0 {
                return Cbs2_ManAnalyze( p, Level, iVar, Gia_ObjFaninId1(pVar, iVar), 0 ); }
            assert(Value0 == 0 && Value1 == 0, "Value0 == 0 && Value1 == 0");
            return Cbs2_ManAnalyze( p, Level, iVar, Gia_ObjFaninId0(pVar, iVar), Gia_ObjFaninId1(pVar, iVar) );
        }
        if Value0 >= 2 { // first is unassigned
            Cbs2_ManAssign( p, Gia_ObjFaninLit0(pVar, iVar), Level, iVar, 0 ); }
        if Value1 >= 2 { // first is unassigned
            Cbs2_ManAssign( p, Gia_ObjFaninLit1(pVar, iVar), Level, iVar, 0 ); }
        return 0;
    }
    // value is 0
    if Value0 == 0 || Value1 == 0 { // one is 0
        return 0; }
    if Value0 == 1 && Value1 == 1 { // both are 1
        return Cbs2_ManAnalyze( p, Level, iVar, Gia_ObjFaninId0(pVar, iVar), Gia_ObjFaninId1(pVar, iVar) ); }
    if Value0 == 1 || Value1 == 1 { // one is 1 
    
        if Value0 >= 2 { // first is unassigned
            Cbs2_ManAssign( p, Abc_LitNot(Gia_ObjFaninLit0(pVar, iVar)), Level, iVar, Gia_ObjFaninId1(pVar, iVar) ); }
        if Value1 >= 2 { // second is unassigned
            Cbs2_ManAssign( p, Abc_LitNot(Gia_ObjFaninLit1(pVar, iVar)), Level, iVar, Gia_ObjFaninId0(pVar, iVar) ); }
        return 0;
    }
    assert((Cbs2_VarIsJust(p, pVar, iVar) != 0), "Cbs2_VarIsJust(p, pVar, iVar)");
    //assert( !Cbs2_QueHasNode( &p->pJust, iVar ) );
    if !p.Pars.fUseFanout {
        Cbs2_QuePush( &p.pJust, iVar ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManPropagateTwo(p *Cbs2_Man_t, iVar int32, Level int32) int32 {
    var pVar *Gia_Obj_t  = Gia_ManObj( p.pAig, iVar ); var Value0, Value1 int32;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    assert((Cbs2_VarIsAssigned(p, iVar) != 0), "Cbs2_VarIsAssigned(p, iVar)");
    assert(Cbs2_VarValue(p, iVar) == 0, "Cbs2_VarValue(p, iVar) == 0");
    Value0 = Cbs2_VarFanin0Value(p, pVar, iVar);
    Value1 = Cbs2_VarFanin1Value(p, pVar, iVar);
    p.nPropCalls[1]++;
    // value is 0
    if Value0 == 0 || Value1 == 0 { // one is 0
        return 0; }
    if Value0 == 1 && Value1 == 1 { // both are 1
        return Cbs2_ManAnalyze( p, Level, iVar, Gia_ObjFaninId0(pVar, iVar), Gia_ObjFaninId1(pVar, iVar) ); }
    assert(Value0 == 1 || Value1 == 1, "Value0 == 1 || Value1 == 1");
    if Value0 >= 2 { // first is unassigned
        Cbs2_ManAssign( p, Abc_LitNot(Gia_ObjFaninLit0(pVar, iVar)), Level, iVar, Gia_ObjFaninId1(pVar, iVar) ); }
    if Value1 >= 2 { // first is unassigned
        Cbs2_ManAssign( p, Abc_LitNot(Gia_ObjFaninLit1(pVar, iVar)), Level, iVar, Gia_ObjFaninId0(pVar, iVar) ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Propagates a variable.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManPropagateUnassigned(p *Cbs2_Man_t, iVar int32, Level int32) {
    var pVar *Gia_Obj_t  = Gia_ManObj( p.pAig, iVar ); var Value0, Value1 int32;
    assert(!Gia_IsComplement(pVar), "!Gia_IsComplement(pVar)");
    assert(Gia_ObjIsAnd(pVar), "Gia_ObjIsAnd(pVar)");
    assert(!(Cbs2_VarIsAssigned(p, iVar) != 0), "!Cbs2_VarIsAssigned(p, iVar)");
    Value0 = Cbs2_VarFanin0Value(p, pVar, iVar);
    Value1 = Cbs2_VarFanin1Value(p, pVar, iVar);
    p.nPropCalls[2]++;
    if Value0 == 0 && Value1 == 0 { // the output becomes 1
        Cbs2_ManAssign( p, Abc_Var2Lit(iVar, 1), Level, Gia_ObjFaninId0(pVar, iVar), Gia_ObjFaninId1(pVar, iVar) );
    } else if Value0 == 0 { // the output becomes 0
        Cbs2_ManAssign( p, Abc_Var2Lit(iVar, 1), Level, Gia_ObjFaninId0(pVar, iVar), 0 );
    } else if Value1 == 0 { // the output becomes 0
        Cbs2_ManAssign( p, Abc_Var2Lit(iVar, 1), Level, Gia_ObjFaninId1(pVar, iVar), 0 );
    } else if Value0 == 1 && Value1 == 1 { // the output becomes 1
        Cbs2_ManAssign( p, Abc_Var2Lit(iVar, 0), Level, Gia_ObjFaninId0(pVar, iVar), Gia_ObjFaninId1(pVar, iVar) ); }
}

/**Function*************************************************************

  Synopsis    [Propagates all variables.]

  Description [Returns 1 if conflict; 0 if no conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManPropagate(p *Cbs2_Man_t, Level int32) int32 {
    for {
        var i, k, iVar, iLit, hClause int32;
        // inlining of macro Cbs2_QueForEachEntry
        for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((iLit) = (p.pProp).pData[i]); i++  {
            if hClause = Cbs2_ManPropagateOne( p, Abc_Lit2Var(iLit), Level ) {
                return hClause; }
        }
        p.pProp.iHead = p.pProp.iTail;
        k = p.pJust.iHead;
        // inlining of macro Cbs2_QueForEachEntry
        for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((iVar) = (p.pJust).pData[i]); i++  {
            if Cbs2_VarIsJust(p, Gia_ManObj(p.pAig, iVar), iVar) != 0 {
                p.pJust.pData[k++] = iVar;
            } else if hClause = Cbs2_ManPropagateTwo( p, iVar, Level ) {
                return hClause; }
        }
        if k == p.pJust.iTail {
            break; }
        p.pJust.iTail = k;
    }
    return 0;
}


func Cbs2_ManPropagate2(p *Cbs2_Man_t, Level int32) int32 {
    var i, iLit, iFan, hClause int32;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((iLit) = (p.pProp).pData[i]); i++  {
        if hClause = Cbs2_ManPropagateClauses(p, Level, iLit) {
            return hClause; }
        // inlining of macro Cbs2_ObjForEachFanout
        for  iFan = Vec_IntEntry(&p.vFanout0, Abc_Lit2Var(iLit)); (iFan != 0); iFan = Vec_IntEntry(&p.vFanoutN, iFan)  {
            var iFanout int32  = Abc_Lit2Var(iFan);
            if !(Cbs2_VarIsAssigned(p, iFanout) != 0) {
                Cbs2_ManPropagateUnassigned( p, iFanout, Level );
            } else if hClause = Cbs2_ManPropagateOne(p, iFanout, Level) {
                return hClause; }
        }
        if hClause = Cbs2_ManPropagateOne( p, Abc_Lit2Var(iLit), Level ) {
            return hClause; }
    }
    p.pProp.iHead = p.pProp.iTail;
    return 0;
}


/**Function*************************************************************

  Synopsis    [Updates J-frontier.]

  Description [Returns 1 if found SAT; 0 if continues solving.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManUpdateDecVar2(p *Cbs2_Man_t, iObj int32, iDecLit int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pAig, iObj); var iFan int32;

    iFan = Gia_ObjFaninId0(pObj, iObj);
    if iDecLit == -1 || Gia_ObjLevelId(p.pAig, Abc_Lit2Var(iDecLit)) < Gia_ObjLevelId(p.pAig, iFan) {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit0(pObj, iObj)); }

    iFan = Gia_ObjFaninId1(pObj, iObj);
    if iDecLit == -1 || Gia_ObjLevelId(p.pAig, Abc_Lit2Var(iDecLit)) < Gia_ObjLevelId(p.pAig, iFan) {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit1(pObj, iObj)); }

    return iDecLit;
}
func Cbs2_ManUpdateDecVar3(p *Cbs2_Man_t, iObj int32, iDecLit int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pAig, iObj); var iFan int32;

    iFan = Gia_ObjFaninId0(pObj, iObj);
    if iDecLit == -1 || Vec_IntEntry(&p.vActivity, Abc_Lit2Var(iDecLit)) < Vec_IntEntry(&p.vActivity, iFan) {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit0(pObj, iObj)); }

    iFan = Gia_ObjFaninId1(pObj, iObj);
    if iDecLit == -1 || Vec_IntEntry(&p.vActivity, Abc_Lit2Var(iDecLit)) < Vec_IntEntry(&p.vActivity, iFan) {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit1(pObj, iObj)); }

    return iDecLit;
}
func Cbs2_ManUpdateDecVar(p *Cbs2_Man_t, iObj int32, iDecLit int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pAig, iObj); var iFan int32;

    iFan = Gia_ObjFaninId0(pObj, iObj);
    if iDecLit == -1 || Gia_ObjRefNumId(p.pAig, Abc_Lit2Var(iDecLit)) < Gia_ObjRefNumId(p.pAig, iFan) {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit0(pObj, iObj)); }

    iFan = Gia_ObjFaninId1(pObj, iObj);
    if iDecLit == -1 || Gia_ObjRefNumId(p.pAig, Abc_Lit2Var(iDecLit)) < Gia_ObjRefNumId(p.pAig, iFan) {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit1(pObj, iObj)); }

    return iDecLit;
}
func Cbs2_ManUpdateFrontier(p *Cbs2_Man_t, iPropHeadOld int32, piDecLit *int32) int32 {
    var clk abctime  = Abc_Clock();
    var i int32;
    var iVar int32;
    var iLit int32;
    var iJustTailOld int32 = p.pJust.iTail;
    *piDecLit = -1;
    assert((Cbs2_QueIsEmpty(&p.pProp) != 0), "Cbs2_QueIsEmpty(&p.pProp)");
    // visit old frontier nodes
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pJust).iHead; (i < (p.pJust).iTail) && ((iVar) = (p.pJust).pData[i]); i++  {
        if i == iJustTailOld {
            break;
        } else if Cbs2_VarIsJust(p, Gia_ManObj(p.pAig, iVar), iVar) != 0 {
            Cbs2_QuePush( &p.pJust, iVar );
            //*piDecLit = Cbs2_ManUpdateDecVar( p, iVar, *piDecLit );
        }
    }
    // append new nodes
    p.pProp.iHead = iPropHeadOld;
    // inlining of macro Cbs2_QueForEachEntry
    for  i = (p.pProp).iHead; (i < (p.pProp).iTail) && ((iLit) = (p.pProp).pData[i]); i++  {
        iVar = Abc_Lit2Var(iLit);
        if Cbs2_VarIsJust(p, Gia_ManObj(p.pAig, iVar), iVar) != 0 {
            Cbs2_QuePush( &p.pJust, iVar );
            //*piDecLit = Cbs2_ManUpdateDecVar( p, iVar, *piDecLit );
        }
    }
    p.pProp.iHead = p.pProp.iTail;
    // update the head of the frontier
    p.pJust.iHead = iJustTailOld;
    // return 1 if the queue is empty
    p.timeJFront += Abc_Clock() - clk;
//printf( "%d ", p->pJust.iTail - p->pJust.iHead );
    return Cbs2_QueIsEmpty(&p.pJust);
}

/**Function*************************************************************

  Synopsis    [Solve the problem recursively.]

  Description [Returns learnt clause if unsat, NULL if sat or undecided.]
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/

func Cbs2_ManSolve1_rec(p *Cbs2_Man_t, Level int32) int32 { 
    var pVar *Gia_Obj_t;
    var pQue *Cbs2_Que_t  = &(p.pClauses);
    var iPropHead, iJustHead, iJustTail int32;
    var hClause, hLearn0, hLearn1, iVar, iDecLit int32;
    // propagate assignments
    assert(!(Cbs2_QueIsEmpty(&p.pProp) != 0), "!Cbs2_QueIsEmpty(&p.pProp)");
    if hClause = Cbs2_ManPropagate( p, Level ) {
        return hClause; }
    // check for satisfying assignment
    assert((Cbs2_QueIsEmpty(&p.pProp) != 0), "Cbs2_QueIsEmpty(&p.pProp)");
    if Cbs2_QueIsEmpty(&p.pJust) != 0 {
        return 0; }
    // quit using resource limits
    p.Pars.nJustThis = Abc_MaxInt( p.Pars.nJustThis, p.pJust.iTail - p.pJust.iHead );
    if Cbs2_ManCheckLimits( p ) != 0 {
        return 0; }
    // remember the state before branching
    iPropHead = p.pProp.iHead;
    Cbs2_QueStore( &p.pJust, &iJustHead, &iJustTail );
    // find the decision variable
    assert(p.Pars.fUseHighest, "p.Pars.fUseHighest");
    iVar = Cbs2_ManDecideHighest( p );
    pVar = Gia_ManObj( p.pAig, iVar );
    assert((Cbs2_VarIsJust(p, pVar, iVar) != 0), "Cbs2_VarIsJust(p, pVar, iVar)");
    // chose decision variable using fanout count
    if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(pVar)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(pVar)) {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit0(pVar, iVar));
    } else {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit1(pVar, iVar)); }
    // decide on first fanin
    Cbs2_ManAssign( p, iDecLit, Level+1, 0, 0 );
    if !(hLearn0 = Cbs2_ManSolve1_rec( p, Level+1 )) {
        return 0; }
    if pQue.pData[hLearn0+1] != Abc_Lit2Var(iDecLit) {
        return hLearn0; }
    Cbs2_ManCancelUntil( p, iPropHead );
    Cbs2_QueRestore( &p.pJust, iJustHead, iJustTail );
    // decide on second fanin
    Cbs2_ManAssign( p, Abc_LitNot(iDecLit), Level+1, 0, 0 );
    if !(hLearn1 = Cbs2_ManSolve1_rec( p, Level+1 )) {
        return 0; }
    if pQue.pData[hLearn1+1] != Abc_Lit2Var(iDecLit) {
        return hLearn1; }
    hClause = Cbs2_ManResolve( p, Level, hLearn0, hLearn1 );
    Cbs2_ManBumpClause( p, hClause );
//    Cbs2_ManPrintCube( p, Level, hClause );
//    if ( Level > Cbs2_ClauseDecLevel(p, hClause) )
//        p->Pars.nBTThisNc++;
    p.Pars.nBTThis++;
    return hClause;
}

func Cbs2_ManSolve2_rec(p *Cbs2_Man_t, Level int32) int32 { 
    var pVar *Gia_Obj_t;
    var pQue *Cbs2_Que_t  = &(p.pClauses);
    var iPropHead, iJustHead, iJustTail int32;
    var hClause, hLearn0, hLearn1, iVar, iDecLit, iDecLit2 int32;
    var iPropHeadOld int32  = p.pProp.iHead;
    // propagate assignments
    assert(!(Cbs2_QueIsEmpty(&p.pProp) != 0), "!Cbs2_QueIsEmpty(&p.pProp)");
    if hClause = Cbs2_ManPropagate2( p, Level ) {
        return hClause; }
    // check for satisfying assignment
    assert((Cbs2_QueIsEmpty(&p.pProp) != 0), "Cbs2_QueIsEmpty(&p.pProp)");
//    if ( Cbs2_QueIsEmpty(&p->pJust) )
//        return 0;
    if Cbs2_ManUpdateFrontier(p, iPropHeadOld, &iDecLit2) != 0 {
        return 0; }
    // quit using resource limits
    p.Pars.nJustThis = Abc_MaxInt( p.Pars.nJustThis, p.pJust.iTail - p.pJust.iHead );
    if Cbs2_ManCheckLimits( p ) != 0 {
        return 0; }
    // remember the state before branching
    iPropHead = p.pProp.iHead;
//    Cbs2_QueStore( &p->pJust, &iJustHead, &iJustTail );
    iJustHead = p.pJust.iHead;
    iJustTail = p.pJust.iTail;
    // find the decision variable

    assert(p.Pars.fUseHighest, "p.Pars.fUseHighest");
    iVar = Cbs2_ManDecideHighest( p );
    pVar = Gia_ManObj( p.pAig, iVar );
    assert((Cbs2_VarIsJust(p, pVar, iVar) != 0), "Cbs2_VarIsJust(p, pVar, iVar)");
    // chose decision variable using fanout count
    if Gia_ObjRefNum(p.pAig, Gia_ObjFanin0(pVar)) > Gia_ObjRefNum(p.pAig, Gia_ObjFanin1(pVar)) {
//    if ( Vec_IntEntry(&p->vActivity, Gia_ObjFaninId0(pVar, iVar)) > Vec_IntEntry(&p->vActivity, Gia_ObjFaninId1(pVar, iVar)) )
        iDecLit = Abc_LitNot(Gia_ObjFaninLit0(pVar, iVar));
    } else {
        iDecLit = Abc_LitNot(Gia_ObjFaninLit1(pVar, iVar)); }

    //iDecLit = iDecLit2;

    // decide on first fanin
    Cbs2_ManAssign( p, iDecLit, Level+1, 0, 0 );
    if !(hLearn0 = Cbs2_ManSolve2_rec( p, Level+1 )) {
        return 0; }
    if pQue.pData[hLearn0+1] != Abc_Lit2Var(iDecLit) {
        return hLearn0; }
    Cbs2_ManCancelUntil( p, iPropHead );
    Cbs2_QueRestore( &p.pJust, iJustHead, iJustTail );
    // decide on second fanin
    Cbs2_ManAssign( p, Abc_LitNot(iDecLit), Level+1, 0, 0 );
    if !(hLearn1 = Cbs2_ManSolve2_rec( p, Level+1 )) {
        return 0; }
    if pQue.pData[hLearn1+1] != Abc_Lit2Var(iDecLit) {
        return hLearn1; }
    hClause = Cbs2_ManResolve( p, Level, hLearn0, hLearn1 );
    Cbs2_ManBumpClause( p, hClause );
    //Cbs2_ManPrintCube( p, Level, hClause );
//    if ( Level > Cbs2_ClauseDecLevel(p, hClause) )
//        p->Pars.nBTThisNc++;
    p.Pars.nBTThis++;
    return hClause;
}

func Cbs2_ManSolve_rec(p *Cbs2_Man_t, Level int32) int32 { 
    if p.Pars.fUseFanout {
        return Cbs2_ManSolve2_rec(p, Level);
    } else {
        return Cbs2_ManSolve1_rec(p, Level);
    }
}

/**Function*************************************************************

  Synopsis    [Looking for a satisfying assignment of the node.]

  Description [Assumes that each node has flag pObj->fMark0 set to 0.
  Returns 1 if unsatisfiable, 0 if satisfiable, and -1 if undecided.
  The node may be complemented. ]
               
  SideEffects [The two procedures differ in the CEX format.]

  SeeAlso     []

***********************************************************************/
func Cbs2_ManSolve(p *Cbs2_Man_t, iLit int32) int32 {
    var RetValue int32;
    assert(!p.pProp.iHead && !p.pProp.iTail, "!p.pProp.iHead && !p.pProp.iTail");
    assert(!p.pJust.iHead && !p.pJust.iTail, "!p.pJust.iHead && !p.pJust.iTail");
    assert(p.pClauses.iHead == 1 && p.pClauses.iTail == 1, "p.pClauses.iHead == 1 && p.pClauses.iTail == 1");
    p.Pars.nBTThis = p.Pars.nJustThis = p.Pars.nBTThisNc = 0;
    Cbs2_ManAssign( p, iLit, 0, 0, 0 );
    if !(Cbs2_ManSolve_rec(p, 0) != 0) && !(Cbs2_ManCheckLimits(p) != 0) {
        Cbs2_ManSaveModel( p, p.vModel );
    } else {
        RetValue = 1; }
    Cbs2_ManCancelUntil( p, 0 );
    Cbs2_ManCleanWatch( p );
    Cbs2_ManBumpClean( p );
    p.pJust.iHead = p.pJust.iTail = 0;
    p.pClauses.iHead = p.pClauses.iTail = 1;
    p.Pars.nBTTotal += p.Pars.nBTThis;
    p.Pars.nJustTotal = Abc_MaxInt( p.Pars.nJustTotal, p.Pars.nJustThis );
    if Cbs2_ManCheckLimits( p ) != 0 {
        RetValue = -1; }
    return RetValue;
}
func Cbs2_ManSolve2(p *Cbs2_Man_t, iLit int32, iLit2 int32) int32 {
    var RetValue int32;
    assert(!p.pProp.iHead && !p.pProp.iTail, "!p.pProp.iHead && !p.pProp.iTail");
    assert(!p.pJust.iHead && !p.pJust.iTail, "!p.pJust.iHead && !p.pJust.iTail");
    assert(p.pClauses.iHead == 1 && p.pClauses.iTail == 1, "p.pClauses.iHead == 1 && p.pClauses.iTail == 1");
    p.Pars.nBTThis = p.Pars.nJustThis = p.Pars.nBTThisNc = 0;
    Cbs2_ManAssign( p, iLit, 0, 0, 0 );
    if iLit2 != 0 { 
    Cbs2_ManAssign( p, iLit2, 0, 0, 0 ); }
    if !(Cbs2_ManSolve_rec(p, 0) != 0) && !(Cbs2_ManCheckLimits(p) != 0) {
        Cbs2_ManSaveModelAll( p, p.vModel );
    } else {
        RetValue = 1; }
    Cbs2_ManCancelUntil( p, 0 );
    Cbs2_ManCleanWatch( p );
    Cbs2_ManBumpClean( p );
    p.pJust.iHead = p.pJust.iTail = 0;
    p.pClauses.iHead = p.pClauses.iTail = 1;
    p.Pars.nBTTotal += p.Pars.nBTThis;
    p.Pars.nJustTotal = Abc_MaxInt( p.Pars.nJustTotal, p.Pars.nJustThis );
    if Cbs2_ManCheckLimits( p ) != 0 {
        RetValue = -1; }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Prints statistics of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManSatPrintStats(p *Cbs2_Man_t) {
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

  Synopsis    [Create fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ObjPrintFanouts(p *Cbs2_Man_t, iObj int32) {
    var iFanLit int32;
    fmt.Printf("Fanouts of node %d: ", iObj);
    // inlining of macro Cbs2_ObjForEachFanout
    for  iFanLit = Vec_IntEntry(&p.vFanout0, iObj); (iFanLit != 0); iFanLit = Vec_IntEntry(&p.vFanoutN, iFanLit)  {
        fmt.Printf("%d ", Abc_Lit2Var(iFanLit));
    }
    fmt.Printf("\n");
}

func Cbs2_ManPrintFanouts(p *Cbs2_Man_t) {
    var pObj *Gia_Obj_t; var iObj int32;
    // inlining of macro Gia_ManForEachObj
    for  iObj = 0; (iObj < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, iObj)); iObj++  {
        if Vec_IntEntry(&p.vFanout0, iObj) {
            Cbs2_ObjPrintFanouts( p, iObj ); }
    }
}
func Cbs2_ObjCreateFanout(p *Cbs2_Man_t, iObj int32, iFan0 int32, iFan1 int32) {
    Vec_IntWriteEntry( &p.vFanoutN, Abc_Var2Lit(iObj, 0), Vec_IntEntry(&p.vFanout0, iFan0) );
    Vec_IntWriteEntry( &p.vFanoutN, Abc_Var2Lit(iObj, 1), Vec_IntEntry(&p.vFanout0, iFan1) );
    Vec_IntWriteEntry( &p.vFanout0, iFan0, Abc_Var2Lit(iObj, 0) );
    Vec_IntWriteEntry( &p.vFanout0, iFan1, Abc_Var2Lit(iObj, 1) );
}
func Cbs2_ObjDeleteFanout(p *Cbs2_Man_t, iObj int32) {
    Vec_IntWriteEntry( &p.vFanoutN, Abc_Var2Lit(iObj, 0), 0 );
    Vec_IntWriteEntry( &p.vFanoutN, Abc_Var2Lit(iObj, 1), 0 );
    Vec_IntWriteEntry( &p.vFanout0, iObj, 0 );
}
func Cbs2_ManCreateFanout_rec(p *Cbs2_Man_t, iObj int32) {
    var pObj *Gia_Obj_t; var iFan0, iFan1 int32;
    if !(iObj != 0) || Gia_ObjIsTravIdCurrentId(p.pAig, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p.pAig, iObj);
    pObj = Gia_ManObj(p.pAig, iObj);
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    iFan0 = Gia_ObjFaninId0(pObj, iObj);
    iFan1 = Gia_ObjFaninId1(pObj, iObj);
    Cbs2_ManCreateFanout_rec( p, iFan0 );
    Cbs2_ManCreateFanout_rec( p, iFan1 );
    Cbs2_ObjCreateFanout( p, iObj, iFan0, iFan1 ); 
}
func Cbs2_ManDeleteFanout_rec(p *Cbs2_Man_t, iObj int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pAig, iObj); var iFan0, iFan1 int32;
    Cbs2_ObjDeleteFanout( p, iObj ); 
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    iFan0 = Gia_ObjFaninId0(pObj, iObj);
    iFan1 = Gia_ObjFaninId1(pObj, iObj);
    if Vec_IntEntry(&p.vFanout0, iFan0) {  Cbs2_ManDeleteFanout_rec( p, iFan0 ); }
    if Vec_IntEntry(&p.vFanout0, iFan1) {  Cbs2_ManDeleteFanout_rec( p, iFan1 ); }
}
func Cbs2_ManCheckFanouts(p *Cbs2_Man_t) {
    var pObj *Gia_Obj_t;
    var iObj int32;
    // inlining of macro Gia_ManForEachObj
    for  iObj = 0; (iObj < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, iObj)); iObj++  {
        assert(Vec_IntEntry(&p.vFanout0, iObj) == 0, "Vec_IntEntry(&p.vFanout0, iObj) == 0");
        assert(Vec_IntEntry(&p.vFanoutN, Abc_Var2Lit(iObj, 0)) == 0, "Vec_IntEntry(&p.vFanoutN, Abc_Var2Lit(iObj, 0)) == 0");
        assert(Vec_IntEntry(&p.vFanoutN, Abc_Var2Lit(iObj, 1)) == 0, "Vec_IntEntry(&p.vFanoutN, Abc_Var2Lit(iObj, 1)) == 0");
    }
}

/**Function*************************************************************

  Synopsis    [Procedure to test the new SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cbs2_ManSolveMiterNc(pAig *Gia_Man_t, nConfs int32, pvStatus **Vec_Str_t, fVerbose int32) *Vec_Int_t {
    extern void Gia_ManCollectTest( Gia_Man_t * pAig );
    extern void Cec_ManSatAddToStore( Vec_Int_t * vCexStore, Vec_Int_t * vCex, int Out );
    var p *Cbs2_Man_t; 
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
    //Gia_ManLevelNum( pAig );
    //Gia_ManCleanMark0( pAig );
    //Gia_ManCleanMark1( pAig );
    //Gia_ManFillValue( pAig ); // maps nodes into trail ids
    //Gia_ManSetPhase( pAig ); // maps nodes into trail ids
    // create logic network
    p = Cbs2_ManAlloc( pAig );
    p.Pars.nBTLimit = nConfs;
    // create resulting data-structures
    vStatus   = Vec_StrAlloc( Gia_ManPoNum(pAig) );
    vCexStore = Vec_IntAlloc( 10000 );
    vVisit    = Vec_IntAlloc( 100 );
    vCex      = Cbs2_ReadModel( p );
    // solve for each output
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pAig.vCos)) && ((pRoot) = Gia_ManCo(pAig, i)); i++  {
        //printf( "\nOutput %d\n", i );

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

        Gia_ManIncrementTravId( pAig );
        Cbs2_ManCreateFanout_rec( p, Gia_ObjFaninId0p(pAig, pRoot) );
        //Cbs2_ManPrintFanouts( p );

        status = Cbs2_ManSolve( p, Gia_ObjFaninLit0p(pAig, pRoot) );
        //printf( "\n" );

        Cbs2_ManDeleteFanout_rec( p, Gia_ObjFaninId0p(pAig, pRoot) );
        //Cbs2_ManCheckFanouts( p );

/*
        if ( status == -1 )
        {
            p->Pars.fUseHighest = 0;
            p->Pars.fUseLowest  = 1;
            status = Cbs2_ManSolve( p, Gia_ObjChild0(pRoot) );
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
        Cbs2_ManSatPrintStats( p ); }
    if fVerbose != 0 {
//    printf( "RecCalls = %8d.  RecClause = %8d.  RecNonChro = %8d.\n", p->nRecCall, p->nRecClause, p->nRecNonChro );
    fmt.Printf("Prop1 = %d.  Prop2 = %d.  Prop3 = %d.  ClaConf = %d.   FailJ = %d.  FailC = %d.   ", p.nPropCalls[0], p.nPropCalls[1], p.nPropCalls[2], p.nClauseConf, p.nFails[0], p.nFails[1]);
    Abc_PrintTime( 1, "JFront", p.timeJFront );
    }

    Cbs2_ManStop( p );
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

