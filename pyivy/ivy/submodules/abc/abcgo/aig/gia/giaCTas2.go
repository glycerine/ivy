package gia

/**CFile****************************************************************

  FileName    [giaCSat2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Circuit-based SAT solver.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaCSat2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Tas_Par_t struct {
    // conflict limits
    nBTLimit           int32;     // limit on the number of conflicts
    // current parameters
    nBTThis           int32;      // number of conflicts
    nBTTotal           int32;     // total number of conflicts
    // decision heuristics
    fUseHighest           int32;  // use node with the highest ID
    // other parameters
    fVerbose           int32;
}

   
type Tas_Sto_t struct {
    iCur           int32;         // currently used
    nSize           int32;        // allocated size
    pBuffer         string;      // handles of objects stored in the queue
}

   
type Tas_Que_t struct {
    iHead           int32;        // beginning of the queue
    iTail           int32;        // end of the queue
    nSize           int32;        // allocated size
    pData          *int32;        // handles of objects stored in the queue
}

   
type Tas_Var_t struct {
    fTerm uint32; /* C bitfield: 1 bits; review packing and truncation */ // terminal node
    fVal uint32; /* C bitfield: 1 bits; review packing and truncation */ // current value
    fValOld uint32; /* C bitfield: 1 bits; review packing and truncation */ // previous value
    fAssign uint32; /* C bitfield: 1 bits; review packing and truncation */ // assigned status
    fJQueue uint32; /* C bitfield: 1 bits; review packing and truncation */ // part of J-frontier
    fCompl0 uint32; /* C bitfield: 1 bits; review packing and truncation */ // complemented attribute
    fCompl1 uint32; /* C bitfield: 1 bits; review packing and truncation */ // complemented attribute
    fMark0 uint32; /* C bitfield: 1 bits; review packing and truncation */ // multi-purpose mark
    fMark1 uint32; /* C bitfield: 1 bits; review packing and truncation */ // multi-purpose mark
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */ // polarity
    Level uint32; /* C bitfield: 22 bits; review packing and truncation */ // decision level 
    Id           int32;           // unique ID of this variable
    IdAig           int32;        // original ID of this variable
    Reason0           int32;      // reason of this variable
    Reason1           int32;      // reason of this variable
    Diff0           int32;        // difference for the first fanin 
    Diff1           int32;        // difference for the second fanin 
    Watch0           int32;       // handle of first watch
    Watch1           int32;       // handle of second watch   
}

   
type Tas_Cls_t struct {
    Watch0           int32;       // next clause to watch
    Watch1           int32;       // next clause to watch
    pVars [0]int32;     // variable handles
}
 
   
type Tas_Man_t struct {
    // user data
    pAig    *Gia_Man_t;         // AIG manager
    Pars     Tas_Par_t;         // parameters
    // solver data
    pVars    *Tas_Sto_t;        // variables
    pClauses    *Tas_Sto_t;     // clauses
    // state representation
    pProp     Tas_Que_t;        // propagation queue
    pJust     Tas_Que_t;        // justification queue
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
    nConfTotal           int32;   // total conflicts
    // runtime stats
    timeSatUnsat       int64; // unsat
    timeSatSat       int64;   // sat
    timeSatUndec       int64; // undecided
    timeTotal       int64;    // total runtime
}

func Tas_VarIsAssigned(pVar *Tas_Var_t) int32 { return pVar.fAssign;                                      }
func Tas_VarAssign(pVar *Tas_Var_t) { assert(!pVar.fAssign, "!pVar.fAssign"); pVar.fAssign = 1;                 }
func Tas_VarUnassign(pVar *Tas_Var_t) { assert(pVar.fAssign, "pVar.fAssign");  pVar.fAssign = 0; pVar.fVal = 0; }
func Tas_VarValue(pVar *Tas_Var_t) int32 { assert(pVar.fAssign, "pVar.fAssign");  return pVar.fVal;                 }
func Tas_VarSetValue(pVar *Tas_Var_t, v int32) { assert(pVar.fAssign, "pVar.fAssign");  pVar.fVal = v;                    }
func Tas_VarIsJust(pVar *Tas_Var_t) int32 { return Gia_ObjIsAnd(pVar) && !Tas_VarIsAssigned(Gia_ObjFanin0(pVar)) && !Tas_VarIsAssigned(Gia_ObjFanin1(pVar)); } 
func Tas_VarFanin0Value(pVar *Tas_Var_t) int32 { if !(Tas_VarIsAssigned(Gia_ObjFanin0(pVar)) != 0) {
    return 2;
} else {
    return (Tas_VarValue(Gia_ObjFanin0(pVar)) ^ Gia_ObjFaninC0(pVar));
} }
func Tas_VarFanin1Value(pVar *Tas_Var_t) int32 { if !(Tas_VarIsAssigned(Gia_ObjFanin1(pVar)) != 0) {
    return 2;
} else {
    return (Tas_VarValue(Gia_ObjFanin1(pVar)) ^ Gia_ObjFaninC1(pVar));
} }

func Tas_VarDecLevel(p *Tas_Man_t, pVar *Tas_Var_t) int32 { assert(pVar.Value != ~0, "pVar.Value != ~0"); return Vec_IntEntry(p.vLevReas, 3*pVar.Value);          }
func Tas_VarReason0(p *Tas_Man_t, pVar *Tas_Var_t) *Tas_Var_t { assert(pVar.Value != ~0, "pVar.Value != ~0"); return pVar + Vec_IntEntry(p.vLevReas, 3*pVar.Value+1); }
func Tas_VarReason1(p *Tas_Man_t, pVar *Tas_Var_t) *Tas_Var_t { assert(pVar.Value != ~0, "pVar.Value != ~0"); return pVar + Vec_IntEntry(p.vLevReas, 3*pVar.Value+2); }
func Tas_ClauseDecLevel(p *Tas_Man_t, hClause int32) int32 { return Tas_VarDecLevel( p, p.pClauses.pData[hClause] );                               }

func Tas_ManVar(p *Tas_Man_t, h int32) *Tas_Var_t { return (Tas_Var_t *)(p.pVars.pBuffer + h);               }
func Tas_ManClause(p *Tas_Man_t, h int32) *Tas_Cls_t { return (Tas_Cls_t *)(p.pClauses.pBuffer + h);            }

// Original macro:
// #define Tas_ClaForEachVar( p, pClause, pVar, i )          \
//     for ( pVar = Tas_ManVar(p, pClause->pVars[(i=0)]); pClause->pVars[i]; pVar = (Tas_Var_t *)(((char *)pVar + pClause->pVars[++i])) )


// Original macro:
// #define Tas_QueForEachVar( p, pQue, pVar, i )             \
//     for ( pVar = Tas_ManVar(p, pQue->pVars[(i=pQue->iHead)]); i < pQue->iTail; pVar = Tas_ManVar(p, pQue->pVars[i++]) )



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Tas_ManCreateVar(p *Tas_Man_t) *Tas_Var_t {
    var pVar *Tas_Var_t;
    if p.pVars.iCur + sizeof(Tas_Var_t) > p.pVars.nSize {
        p.pVars.nSize *= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pVars.pData {
            p.pVars.pData = ((char *) realloc((char *)(p.pVars.pData), sizeof(char) * (size_t)(p.pVars.nSize)));
        } else {
            p.pVars.pData = ((char *) malloc(sizeof(char) * (size_t)(p.pVars.nSize)));
        } 
    }
    pVar = p.pVars.pData + p.pVars.iCur;
    p.pVars.iCur += sizeof(Tas_Var_t);
    *pVar = Tas_Var_t{};
    pVar.Id = pVar - ((Tas_Var_t *)p.pVars.pData);
    return pVar;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

