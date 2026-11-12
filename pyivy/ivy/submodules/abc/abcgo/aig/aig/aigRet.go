package aig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [aigRet.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Retiming of AIGs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigRet.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// init values
 type Rtm_Init_t int32
const ( 
    RTM_VAL_NONE Rtm_Init_t = iota
                    // 0: non-existent value
    RTM_VAL_ZERO
                    // 1: initial value 0
    RTM_VAL_ONE
                     // 2: initial value 1
    RTM_VAL_VOID                     // 3: unused value

)

 

      
type Rtm_Man_t struct {
    // network representation
    vObjs       *Vec_Ptr_t;          // retiming objects
    vPis       *Vec_Ptr_t;           // PIs only
    vPos       *Vec_Ptr_t;           // POs only
    pMem    *Aig_MmFlex_t;           // the memory manager
    // autonomous components after cutting off
    // storage for overflow latches
    pExtra        *uint32;   
    nExtraCur              int32;
    nExtraAlloc              int32;
}

      
type Rtm_Edg_t struct {
    unsigned long    nLats   :  12;  // the number of latches
    unsigned long    LData   :  20;  // the latches themselves
}

      
type Rtm_Obj_t struct {
    pCopy            any;          // the copy of this object
    unsigned long    Type    :  3;   // object type
    unsigned long    fMark   :  1;   // multipurpose mark
    unsigned long    fAuto   :  1;   // this object belongs to an autonomous component
    unsigned long    fCompl0 :  1;   // complemented attribute of the first edge
    unsigned long    fCompl1 :  1;   // complemented attribute of the second edge
    unsigned long    nFanins :  8;   // the number of fanins
    Num uint32; /* C bitfield: 17 bits; review packing and truncation */   // the retiming number of this node
    Id              int32;             // ID of this object
    Temp              int32;           // temporary usage
    nFanouts              int32;       // the number of fanouts
    void *           pFanio[0];      // fanins and their edges (followed by fanouts and pointers to their edges)
}

func Rtm_ObjFanin(pObj *Rtm_Obj_t, i int32) *Rtm_Obj_t { return (Rtm_Obj_t *)pObj.pFanio[2*i];                     }
func Rtm_ObjFanout(pObj *Rtm_Obj_t, i int32) *Rtm_Obj_t { return (Rtm_Obj_t *)pObj.pFanio[2*(pObj.nFanins+i)];     }
func Rtm_ObjEdge(pObj *Rtm_Obj_t, i int32) *Rtm_Edg_t { return (Rtm_Edg_t *)(pObj.pFanio + 2*i + 1);              }
func Rtm_ObjFanoutEdge(pObj *Rtm_Obj_t, i int32) *Rtm_Edg_t { return (Rtm_Edg_t *)pObj.pFanio[2*(pObj.nFanins+i) + 1]; }

func Rtm_InitNot(Val Rtm_Init_t) Rtm_Init_t { if Val == RTM_VAL_ZERO { return (Rtm_Init_t)RTM_VAL_ONE; }
 if Val == RTM_VAL_ONE { return (Rtm_Init_t)RTM_VAL_ZERO; }
 assert(false, "0"); return (Rtm_Init_t)-1; }
func Rtm_InitNotCond(Val Rtm_Init_t, c int32) Rtm_Init_t { if c != 0 {
    return Rtm_InitNot(Val);
} else {
    return Val;
}                         }
func Rtm_InitAnd(ValA Rtm_Init_t, ValB Rtm_Init_t) Rtm_Init_t { if ValA == RTM_VAL_ONE && ValB == RTM_VAL_ONE { return (Rtm_Init_t)RTM_VAL_ONE; }
  if ValA == RTM_VAL_ZERO || ValB == RTM_VAL_ZERO { return (Rtm_Init_t)RTM_VAL_ZERO; }
 assert(false, "0"); return (Rtm_Init_t)-1;   }

func Rtm_InitWordsNum(nLats int32) int32 { return (nLats >> 4) + ((nLats & 15) > 0); }
func Rtm_InitGetTwo(p *uint32, i int32) int32 { return (p[i>>4] >> ((i & 15)<<1)) & 3;    }
func Rtm_InitSetTwo(p *uint32, i int32, val int32) { p[i>>4] |= (val << ((i & 15)<<1));        }
func Rtm_InitXorTwo(p *uint32, i int32, val int32) { p[i>>4] ^= (val << ((i & 15)<<1));        }

func Rtm_ObjGetFirst1(pEdge *Rtm_Edg_t) Rtm_Init_t { return (Rtm_Init_t)(pEdge.LData & 3);                                    }
func Rtm_ObjGetLast1(pEdge *Rtm_Edg_t) Rtm_Init_t { return (Rtm_Init_t)((pEdge.LData >> ((pEdge.nLats-1)<<1)) & 3);         }
func Rtm_ObjGetOne1(pEdge *Rtm_Edg_t, i int32) Rtm_Init_t { assert(i < int32(pEdge.nLats), "i < int32(pEdge.nLats)"); return (Rtm_Init_t)((pEdge.LData >> (i << 1)) & 3);  }
func Rtm_ObjRemFirst1(pEdge *Rtm_Edg_t) Rtm_Init_t { var Val int32  = pEdge.LData & 3; pEdge.LData >>= 2; assert(pEdge.nLats > 0, "pEdge.nLats > 0"); pEdge.nLats--; return (Rtm_Init_t)Val;  }
func Rtm_ObjRemLast1(pEdge *Rtm_Edg_t) Rtm_Init_t { var Val int32  = (pEdge.LData >> ((pEdge.nLats-1)<<1)) & 3; pEdge.LData ^= Val << ((pEdge.nLats-1)<<1); assert(pEdge.nLats > 0, "pEdge.nLats > 0"); pEdge.nLats--; return (Rtm_Init_t)Val;  }
func Rtm_ObjAddFirst1(pEdge *Rtm_Edg_t, Val Rtm_Init_t) { assert(Val > 0, "Val > 0"); pEdge.LData = (pEdge.LData << 2) | Val;  pEdge.nLats++;   }
func Rtm_ObjAddLast1(pEdge *Rtm_Edg_t, Val Rtm_Init_t) { assert(Val > 0, "Val > 0"); pEdge.LData |= Val << (pEdge.nLats<<1);  pEdge.nLats++;   }

func Rtm_ObjGetFirst2(p *Rtm_Man_t, pEdge *Rtm_Edg_t) Rtm_Init_t { return (Rtm_Init_t)Rtm_InitGetTwo( p.pExtra + pEdge.LData, 0 );                }
func Rtm_ObjGetLast2(p *Rtm_Man_t, pEdge *Rtm_Edg_t) Rtm_Init_t { return (Rtm_Init_t)Rtm_InitGetTwo( p.pExtra + pEdge.LData, pEdge.nLats - 1 ); }
func Rtm_ObjGetOne2(p *Rtm_Man_t, pEdge *Rtm_Edg_t, i int32) Rtm_Init_t { return (Rtm_Init_t)Rtm_InitGetTwo( p.pExtra + pEdge.LData, i );                }
// static        Rtm_Init_t  Rtm_ObjRemFirst2( Rtm_Man_t * p, Rtm_Edg_t * pEdge );

func Rtm_ObjRemLast2(p *Rtm_Man_t, pEdge *Rtm_Edg_t) Rtm_Init_t { var Val Rtm_Init_t  = Rtm_ObjGetLast2( p, pEdge );    Rtm_InitXorTwo( p.pExtra + pEdge.LData, pEdge.nLats - 1, Val ); pEdge.nLats--; return (Rtm_Init_t)Val; }
// static        void        Rtm_ObjAddFirst2( Rtm_Man_t * p, Rtm_Edg_t * pEdge, Rtm_Init_t Val );

func Rtm_ObjAddLast2(p *Rtm_Man_t, pEdge *Rtm_Edg_t, Val Rtm_Init_t) { Rtm_InitSetTwo( p.pExtra + pEdge.LData, pEdge.nLats, Val );  pEdge.nLats++;  }

// static        void        Rtm_ObjTransferToSmall( Rtm_Man_t * p, Rtm_Edg_t * pEdge );

// static        void        Rtm_ObjTransferToBig( Rtm_Man_t * p, Rtm_Edg_t * pEdge );
 
// static        void        Rtm_ObjTransferToBigger( Rtm_Man_t * p, Rtm_Edg_t * pEdge );


func Rtm_ObjGetFirst(p *Rtm_Man_t, pEdge *Rtm_Edg_t) Rtm_Init_t { if pEdge.nLats > 10 {
    return Rtm_ObjGetFirst2(p, pEdge);
} else {
    return Rtm_ObjGetFirst1(pEdge);
}   }
func Rtm_ObjGetLast(p *Rtm_Man_t, pEdge *Rtm_Edg_t) Rtm_Init_t { if pEdge.nLats > 10 {
    return Rtm_ObjGetLast2(p, pEdge);
} else {
    return Rtm_ObjGetLast1(pEdge);
}    }
func Rtm_ObjGetOne(p *Rtm_Man_t, pEdge *Rtm_Edg_t, i int32) Rtm_Init_t { if pEdge.nLats > 10 {
    return Rtm_ObjGetOne2(p, pEdge, i);
} else {
    return Rtm_ObjGetOne1(pEdge, i);
}  }
func Rtm_ObjRemFirst(p *Rtm_Man_t, pEdge *Rtm_Edg_t) Rtm_Init_t { var Res Rtm_Init_t;
if pEdge.nLats > 10 {
    Res = Rtm_ObjRemFirst2(p, pEdge);
} else {
    Res = Rtm_ObjRemFirst1(pEdge);
}
 if pEdge.nLats == 10 { Rtm_ObjTransferToSmall(p, pEdge); }
 return Res; }
func Rtm_ObjRemLast(p *Rtm_Man_t, pEdge *Rtm_Edg_t) Rtm_Init_t { var Res Rtm_Init_t;
if pEdge.nLats > 10 {
    Res = Rtm_ObjRemLast2(p, pEdge);
} else {
    Res = Rtm_ObjRemLast1(pEdge);
}
  if pEdge.nLats == 10 { Rtm_ObjTransferToSmall(p, pEdge); }
 return Res; }
func Rtm_ObjAddFirst(p *Rtm_Man_t, pEdge *Rtm_Edg_t, Val Rtm_Init_t) { if pEdge.nLats == 10 { Rtm_ObjTransferToBig(p, pEdge); } else if (pEdge.nLats & 15) == 15 { Rtm_ObjTransferToBigger(p, pEdge); }
 if pEdge.nLats >= 10 { Rtm_ObjAddFirst2(p, pEdge, Val); } else { Rtm_ObjAddFirst1(pEdge, Val); } }
func Rtm_ObjAddLast(p *Rtm_Man_t, pEdge *Rtm_Edg_t, Val Rtm_Init_t) { if pEdge.nLats == 10 { Rtm_ObjTransferToBig(p, pEdge); } else if (pEdge.nLats & 15) == 15 { Rtm_ObjTransferToBigger(p, pEdge); }
 if pEdge.nLats >= 10 { Rtm_ObjAddLast2(p, pEdge, Val); }  else { Rtm_ObjAddLast1(pEdge, Val); }  }


// iterator over the primary inputs
// Original macro:
// #define Rtm_ManForEachPi( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Rtm_Obj_t *, p->vPis, pObj, i )

// iterator over the primary outputs
// Original macro:
// #define Rtm_ManForEachPo( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( Rtm_Obj_t *, p->vPos, pObj, i )

// iterator over all objects, including those currently not used
// Original macro:
// #define Rtm_ManForEachObj( p, pObj, i )                                         \
//     Vec_PtrForEachEntry( Rtm_Obj_t *, p->vObjs, pObj, i ) 

// iterate through the fanins
// Original macro:
// #define Rtm_ObjForEachFanin( pObj, pFanin, i )                                  \
//     for ( i = 0; i < (int)(pObj)->nFanins && ((pFanin = Rtm_ObjFanin(pObj, i)), 1); i++ )

// iterate through the fanouts
// Original macro:
// #define Rtm_ObjForEachFanout( pObj, pFanout, i )                                \
//     for ( i = 0; i < (int)(pObj)->nFanouts && ((pFanout = Rtm_ObjFanout(pObj, i)), 1); i++ )

// iterate through the fanin edges
// Original macro:
// #define Rtm_ObjForEachFaninEdge( pObj, pEdge, i )                               \
//     for ( i = 0; i < (int)(pObj)->nFanins && ((pEdge = Rtm_ObjEdge(pObj, i)), 1); i++ )

// iterate through the fanout edges
// Original macro:
// #define Rtm_ObjForEachFanoutEdge( pObj, pEdge, i )                              \
//     for ( i = 0; i < (int)(pObj)->nFanouts && ((pEdge = Rtm_ObjFanoutEdge(pObj, i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transfers from big to small storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjTransferToSmall(p *Rtm_Man_t, pEdge *Rtm_Edg_t) {
    assert(pEdge.nLats == 10, "pEdge.nLats == 10");
    pEdge.LData = p.pExtra[pEdge.LData];
}

/**Function*************************************************************

  Synopsis    [Transfers from small to big storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjTransferToBig(p *Rtm_Man_t, pEdge *Rtm_Edg_t) {
    assert(pEdge.nLats == 10, "pEdge.nLats == 10");
    if p.nExtraCur + 1 > p.nExtraAlloc {
        var nExtraAllocNew int32  = Abc_MaxInt( 2 * p.nExtraAlloc, 1024 );
        // inlining of macro ABC_REALLOC
        
        if p.pExtra {
            p.pExtra = ((unsigned *) realloc((char *)(p.pExtra), sizeof(unsigned) * (size_t)(nExtraAllocNew)));
        } else {
            p.pExtra = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nExtraAllocNew)));
        }
        p.nExtraAlloc = nExtraAllocNew;
    }
    p.pExtra[p.nExtraCur] = pEdge.LData;
    pEdge.LData = p.nExtraCur++;
}

/**Function*************************************************************

  Synopsis    [Transfers to bigger storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjTransferToBigger(p *Rtm_Man_t, pEdge *Rtm_Edg_t) {
    var nWords int32;
    assert((pEdge.nLats & 15) == 15, "(pEdge.nLats & 15) == 15");
    nWords = (pEdge.nLats + 1) >> 4;
    if p.nExtraCur + nWords + 1 > p.nExtraAlloc {
        var nExtraAllocNew int32  = Abc_MaxInt( 2 * p.nExtraAlloc, 1024 );
        // inlining of macro ABC_REALLOC
        
        if p.pExtra {
            p.pExtra = ((unsigned *) realloc((char *)(p.pExtra), sizeof(unsigned) * (size_t)(nExtraAllocNew)));
        } else {
            p.pExtra = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nExtraAllocNew)));
        }
        p.nExtraAlloc = nExtraAllocNew;
    }
    memcpy( p.pExtra + p.nExtraCur, p.pExtra + pEdge.LData, sizeof(unsigned) * nWords );
    p.pExtra[p.nExtraCur + nWords] = 0;
    pEdge.LData = p.nExtraCur;
    p.nExtraCur += nWords + 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjRemFirst2(p *Rtm_Man_t, pEdge *Rtm_Edg_t) Rtm_Init_t { 
    var Val Rtm_Init_t = (Rtm_Init_t)0;
    var Temp Rtm_Init_t;   
    var pB *uint32 = p.pExtra + pEdge.LData;
    var pE *uint32 = pB + Rtm_InitWordsNum( pEdge.nLats-- ) - 1;  
    for  pE >= pB  {
        Temp = (Rtm_Init_t)(*pE & 3);
        *pE = (*pE >> 2) | (Val << 30);
        Val = Temp;
        pE--;
    }
    assert(Val != 0, "Val != 0");
    return Val;  
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjAddFirst2(p *Rtm_Man_t, pEdge *Rtm_Edg_t, Val Rtm_Init_t) {
    var pB *uint32 = p.pExtra + pEdge.LData;
    var pE *uint32 = pB + Rtm_InitWordsNum( ++pEdge.nLats ); 
    var Temp Rtm_Init_t;
    assert(Val != 0, "Val != 0");
    for  pB < pE  {
        Temp = (Rtm_Init_t)(*pB >> 30);
        *pB = (*pB << 2) | Val;
        Val = Temp;
        pB++;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_PrintEdge(p *Rtm_Man_t, pEdge *Rtm_Edg_t) {
//    unsigned LData = pEdge->LData;
    fmt.Printf("%d : ", int32(pEdge.nLats));
/*
    if ( pEdge->nLats > 10 )
        Extra_PrintBinary( stdout, p->pExtra + pEdge->LData, 2*(pEdge->nLats+1) );
    else
        Extra_PrintBinary( stdout, &LData, 2*(pEdge->nLats+1) );
*/
    fmt.Printf("\n");
}


/**Function*************************************************************

  Synopsis    [Allocates the retiming manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManAlloc(p *Aig_Man_t) *Rtm_Man_t {
    var pRtm *Rtm_Man_t;
    // start the manager
    // inlining of macro ABC_ALLOC
    pRtm = ((Rtm_Man_t *) malloc(sizeof(Rtm_Man_t) * (size_t)(1)));
    *pRtm = Rtm_Man_t{};
    // perform initializations
    pRtm.vObjs = Vec_PtrAlloc( Aig_ManObjNum(p) );
    pRtm.vPis  = Vec_PtrAlloc( Aig_ManCiNum(p) );
    pRtm.vPos  = Vec_PtrAlloc( Aig_ManCoNum(p) );
    pRtm.pMem  = Aig_MmFlexStart();
    return pRtm;
}

/**Function*************************************************************

  Synopsis    [Allocates the retiming manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManFree(p *Rtm_Man_t) {
    Vec_PtrFree( p.vObjs );
    Vec_PtrFree( p.vPis );
    Vec_PtrFree( p.vPos );
    Aig_MmFlexStop( p.pMem, 0 );
    // inlining of macro ABC_FREE
    ((p.pExtra) ? (free((char *) (p.pExtra)), (p.pExtra) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Counts the maximum number of latches on an edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManLatchMax(p *Rtm_Man_t) int32 {
    var pObj *Rtm_Obj_t;
    var pEdge *Rtm_Edg_t;
    var nLatchMax, i, k int32;//, c, Val;
    // inlining of macro Rtm_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Rtm_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { 
    // inlining of macro Rtm_ObjForEachFaninEdge
    for  k = 0; k < int32((pObj).nFanins) && ((pEdge = Rtm_ObjEdge(pObj, k)), 1); k++  {
/*
        for ( c = 0; c < (int)pEdge->nLats; c++ )
        {
            Val = Rtm_ObjGetOne( p, pEdge, c );
            assert( Val == 1 || Val == 2 );
        }
*/
        nLatchMax = Abc_MaxInt( nLatchMax, int32(pEdge.nLats) );
    }
    }
    return nLatchMax;
}

/**Function*************************************************************

  Synopsis    [Allocates the retiming object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjAlloc(pRtm *Rtm_Man_t, nFanins int32, nFanouts int32) *Rtm_Obj_t {
    var pObj *Rtm_Obj_t;
    var Size int32  = sizeof(Rtm_Obj_t) + sizeof(Rtm_Obj_t *) * (nFanins + nFanouts) * 2;
    pObj = (Rtm_Obj_t *)Aig_MmFlexEntryFetch( pRtm.pMem, Size );
    *pObj = Rtm_Obj_t{};
    pObj.Type = int32((nFanins == 1 && nFanouts == 0)); // mark PO
    pObj.Num  = nFanins;  // temporary
    pObj.Temp = nFanouts;
    pObj.Id = Vec_PtrSize(pRtm.vObjs);
    Vec_PtrPush( pRtm.vObjs, pObj );
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Allocates the retiming object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjAddFanin(pObj *Rtm_Obj_t, pFanin *Rtm_Obj_t, fCompl int32) {
    pObj.pFanio[ 2*pObj.nFanins ] = pFanin;
    pObj.pFanio[ 2*pObj.nFanins + 1 ] = nil;
    pFanin.pFanio[ 2*(pFanin.Num + pFanin.nFanouts) ] = pObj;
    pFanin.pFanio[ 2*(pFanin.Num + pFanin.nFanouts) + 1 ] = pObj.pFanio + 2*pObj.nFanins + 1;
    if pObj.nFanins == 0 {
        pObj.fCompl0 = fCompl;
    } else if pObj.nFanins == 1 {
        pObj.fCompl1 = fCompl;
    } else {
        assert(false, "0"); }
    pObj.nFanins++;
    pFanin.nFanouts++;
    assert(pObj.nFanins <= pObj.Num, "pObj.nFanins <= pObj.Num");
    assert(pFanin.nFanouts <= pFanin.Temp, "pFanin.nFanouts <= pFanin.Temp");
}

/**Function*************************************************************

  Synopsis    [Check the possibility of forward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjCheckRetimeFwd(pObj *Rtm_Obj_t) int32 {
    var pEdge *Rtm_Edg_t;
    var i int32;
    // inlining of macro Rtm_ObjForEachFaninEdge
    for  i = 0; i < int32((pObj).nFanins) && ((pEdge = Rtm_ObjEdge(pObj, i)), 1); i++  {
        if pEdge.nLats == 0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Check the possibility of forward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjCheckRetimeBwd(pObj *Rtm_Obj_t) int32 {
    var pEdge *Rtm_Edg_t;
    var i int32;
    // inlining of macro Rtm_ObjForEachFanoutEdge
    for  i = 0; i < int32((pObj).nFanouts) && ((pEdge = Rtm_ObjFanoutEdge(pObj, i)), 1); i++  {
        if pEdge.nLats == 0 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Check the possibility of forward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjGetDegreeFwd(pObj *Rtm_Obj_t) int32 {
    var pFanin *Rtm_Obj_t;
    var i, Degree int32;
    // inlining of macro Rtm_ObjForEachFanin
    for  i = 0; i < int32((pObj).nFanins) && ((pFanin = Rtm_ObjFanin(pObj, i)), 1); i++  {
        Degree = Abc_MaxInt( Degree, int32(pFanin.Num) );
    }
    return Degree + 1;
}

/**Function*************************************************************

  Synopsis    [Check the possibility of forward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjGetDegreeBwd(pObj *Rtm_Obj_t) int32 {
    var pFanout *Rtm_Obj_t;
    var i, Degree int32;
    // inlining of macro Rtm_ObjForEachFanout
    for  i = 0; i < int32((pObj).nFanouts) && ((pFanout = Rtm_ObjFanout(pObj, i)), 1); i++  {
        Degree = Abc_MaxInt( Degree, int32(pFanout.Num) );
    }
    return Degree + 1;
}

/**Function*************************************************************

  Synopsis    [Performs forward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjRetimeFwd(pRtm *Rtm_Man_t, pObj *Rtm_Obj_t) {
    var ValTotal, ValCur Rtm_Init_t;
    var pEdge *Rtm_Edg_t;
    var i int32;
    assert((Rtm_ObjCheckRetimeFwd(pObj) != 0), "Rtm_ObjCheckRetimeFwd(pObj)");
    // extract values and compute the result
    ValTotal = RTM_VAL_ONE;
    // inlining of macro Rtm_ObjForEachFaninEdge
    for  i = 0; i < int32((pObj).nFanins) && ((pEdge = Rtm_ObjEdge(pObj, i)), 1); i++  {
        ValCur = Rtm_ObjRemFirst( pRtm, pEdge );
        ValCur = Rtm_InitNotCond( ValCur, i? pObj.fCompl1 : pObj.fCompl0 );
        ValTotal = Rtm_InitAnd( ValTotal, ValCur );
    }
    // insert the result in the fanout values
    // inlining of macro Rtm_ObjForEachFanoutEdge
    for  i = 0; i < int32((pObj).nFanouts) && ((pEdge = Rtm_ObjFanoutEdge(pObj, i)), 1); i++  {
        Rtm_ObjAddLast( pRtm, pEdge, ValTotal );
    }
}

/**Function*************************************************************

  Synopsis    [Performs forward retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjRetimeBwd(pRtm *Rtm_Man_t, pObj *Rtm_Obj_t) {
    var pEdge *Rtm_Edg_t;
    var i int32;
    assert((Rtm_ObjCheckRetimeBwd(pObj) != 0), "Rtm_ObjCheckRetimeBwd(pObj)");
    // extract values and compute the result
    // inlining of macro Rtm_ObjForEachFanoutEdge
    for  i = 0; i < int32((pObj).nFanouts) && ((pEdge = Rtm_ObjFanoutEdge(pObj, i)), 1); i++  {
        Rtm_ObjRemLast( pRtm, pEdge );
    }
    // insert the result in the fanout values
    // inlining of macro Rtm_ObjForEachFaninEdge
    for  i = 0; i < int32((pObj).nFanins) && ((pEdge = Rtm_ObjEdge(pObj, i)), 1); i++  {
        Rtm_ObjAddFirst( pRtm, pEdge, RTM_VAL_VOID );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjMarkAutoFwd_rec(pObj *Rtm_Obj_t) {
    var pFanout *Rtm_Obj_t;
    var i int32;
    if pObj.fAuto {
        return; }
    pObj.fAuto = 1;
    // inlining of macro Rtm_ObjForEachFanout
    for  i = 0; i < int32((pObj).nFanouts) && ((pFanout = Rtm_ObjFanout(pObj, i)), 1); i++  {
        Rtm_ObjMarkAutoFwd_rec( pFanout );
    }
}

/**Function*************************************************************

  Synopsis    [Marks the nodes unreachable from the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManMarkAutoFwd(pRtm *Rtm_Man_t) int32 {
    var pObjRtm *Rtm_Obj_t;
    var i, Counter int32;
    // mark nodes reachable from the PIs
    pObjRtm = (Rtm_Obj_t *)Vec_PtrEntry( pRtm.vObjs, 0 );
    Rtm_ObjMarkAutoFwd_rec( pObjRtm );
    // inlining of macro Rtm_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vPis)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vPis, i)), 1); i++  {
        Rtm_ObjMarkAutoFwd_rec( pObjRtm );
    }
    // count the number of autonomous nodes
    // inlining of macro Rtm_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vObjs)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vObjs, i)), 1); i++  {
        pObjRtm.fAuto = !pObjRtm.fAuto;
        Counter += pObjRtm.fAuto;
    }
    // mark the fanins of the autonomous nodes
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ObjMarkAutoBwd_rec(pObj *Rtm_Obj_t) {
    var pFanin *Rtm_Obj_t;
    var i int32;
    if pObj.fAuto {
        return; }
    pObj.fAuto = 1;
    // inlining of macro Rtm_ObjForEachFanin
    for  i = 0; i < int32((pObj).nFanins) && ((pFanin = Rtm_ObjFanin(pObj, i)), 1); i++  {
        Rtm_ObjMarkAutoBwd_rec( pFanin );
    }
}

/**Function*************************************************************

  Synopsis    [Marks the nodes unreachable from the POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManMarkAutoBwd(pRtm *Rtm_Man_t) int32 {
    var pObjRtm *Rtm_Obj_t;
    var i, Counter int32;
    // mark nodes reachable from the PIs
    pObjRtm = (Rtm_Obj_t *)Vec_PtrEntry( pRtm.vObjs, 0 );
    pObjRtm.fAuto = 1;
    // inlining of macro Rtm_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vPis)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vPis, i)), 1); i++  {
        pObjRtm.fAuto = 1;
    }
    // inlining of macro Rtm_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vPos)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vPos, i)), 1); i++  {
        Rtm_ObjMarkAutoBwd_rec( pObjRtm );
    }
    // count the number of autonomous nodes
    // inlining of macro Rtm_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vObjs)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vObjs, i)), 1); i++  {
        pObjRtm.fAuto = !pObjRtm.fAuto;
        Counter += pObjRtm.fAuto;
    }
    // mark the fanins of the autonomous nodes
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Derive retiming manager from the given AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManFromAig(p *Aig_Man_t) *Rtm_Man_t {
    var pRtm *Rtm_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    assert(Aig_ManBufNum(p) == 0, "Aig_ManBufNum(p) == 0");
    // allocate the manager
    pRtm = Rtm_ManAlloc( p );
    // allocate objects
    pObj = Aig_ManConst1(p);
    pObj.pData = Rtm_ObjAlloc( pRtm, 0, pObj.nRefs );
    // inlining of macro Aig_ManForEachPiSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Rtm_ObjAlloc( pRtm, 0, pObj.nRefs );
        Vec_PtrPush( pRtm.vPis, pObj.pData );
    }
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.pData = Rtm_ObjAlloc( pRtm, 1, 0 );
        Vec_PtrPush( pRtm.vPos, pObj.pData );
    }
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.pData = Rtm_ObjAlloc( pRtm, 1, pObj.nRefs );
    }
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.pData = Rtm_ObjAlloc( pRtm, 1, 1 );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Rtm_ObjAlloc( pRtm, 2, pObj.nRefs ); }
    }
    // connect objects
    // inlining of macro Aig_ManForEachPoSeq
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Aig_ManCoNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Rtm_ObjAddFanin( (Rtm_Obj_t *)pObj.pData, (Rtm_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0(pObj) );
    }
    // inlining of macro Aig_ManForEachLiSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Rtm_ObjAddFanin( (Rtm_Obj_t *)pObj.pData, (Rtm_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0(pObj) );
    }
    // inlining of macro Aig_ManForEachLiLoSeq
    for  i = 0; (i < Aig_ManRegNum(p)) && (((pObjLi) = Aig_ManLi(p, i)), 1)            && (((pObjLo)=Aig_ManLo(p, i)), 1); i++  {
        Rtm_ObjAddFanin( (Rtm_Obj_t *)pObjLo.pData, (Rtm_Obj_t *)pObjLi.pData, 0 );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Rtm_ObjAddFanin( (Rtm_Obj_t *)pObj.pData, (Rtm_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0(pObj) );
        Rtm_ObjAddFanin( (Rtm_Obj_t *)pObj.pData, (Rtm_Obj_t *)Aig_ObjFanin1(pObj).pData, Aig_ObjFaninC1(pObj) );
    }
    }
    return pRtm;
}

/**Function*************************************************************

  Synopsis    [Derive AIG manager after retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManToAig_rec(pNew *Aig_Man_t, pRtm *Rtm_Man_t, pObjRtm *Rtm_Obj_t, pLatches *int32) *Aig_Obj_t {
    var pEdge *Rtm_Edg_t;
    var pRes, pFanin *Aig_Obj_t;
    var k, Val int32;
    if pObjRtm.pCopy {
        return (Aig_Obj_t *)pObjRtm.pCopy; }
    // get the inputs
    pRes = Aig_ManConst1( pNew );
    // inlining of macro Rtm_ObjForEachFaninEdge
    for  k = 0; k < int32((pObjRtm).nFanins) && ((pEdge = Rtm_ObjEdge(pObjRtm, k)), 1); k++  {
        if pEdge.nLats == 0 {
            pFanin = Rtm_ManToAig_rec( pNew, pRtm, Rtm_ObjFanin(pObjRtm, k), pLatches );
        } else {
            Val = Rtm_ObjGetFirst( pRtm, pEdge );
            pFanin = Aig_ManCi( pNew, pLatches[2*pObjRtm.Id + k] + pEdge.nLats - 1 );
            pFanin = Aig_NotCond( pFanin, Val == RTM_VAL_ONE );
        }
        pFanin = Aig_NotCond( pFanin, k ? pObjRtm.fCompl1 : pObjRtm.fCompl0 );
        pRes = Aig_And( pNew, pRes, pFanin );
    }
    return (Aig_Obj_t *)(pObjRtm.pCopy = pRes);
}

/**Function*************************************************************

  Synopsis    [Derive AIG manager after retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManToAig(pRtm *Rtm_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObjNew *Aig_Obj_t;
    var pObjRtm *Rtm_Obj_t;
    var pEdge *Rtm_Edg_t;
    var i int32;
    var k int32;
    var m int32;
    var Val int32;
    var nLatches int32;
    var pLatches *int32;
    // count latches and mark the first latch on each edge
    // inlining of macro ABC_ALLOC
    pLatches = ((int *) malloc(sizeof(int) * (size_t)(2 * Vec_PtrSize(pRtm.vObjs))));
    nLatches = 0;
    // inlining of macro Rtm_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vObjs)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vObjs, i)), 1); i++  { 
    // inlining of macro Rtm_ObjForEachFaninEdge
    for  k = 0; k < int32((pObjRtm).nFanins) && ((pEdge = Rtm_ObjEdge(pObjRtm, k)), 1); k++  {
        pLatches[2*pObjRtm.Id + k] = Vec_PtrSize(pRtm.vPis) + nLatches;
        nLatches += pEdge.nLats;
    }
    }
    // create the new manager
    pNew = Aig_ManStart( Vec_PtrSize(pRtm.vObjs) + nLatches );
    // create PIs/POs and latches
    pObjRtm = (Rtm_Obj_t *)Vec_PtrEntry( pRtm.vObjs, 0 );
    pObjRtm.pCopy = Aig_ManConst1(pNew);
    // inlining of macro Rtm_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vPis)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vPis, i)), 1); i++  {
        pObjRtm.pCopy = Aig_ObjCreateCi(pNew);
    }
    for  i = 0; i < nLatches; i++  {
        Aig_ObjCreateCi(pNew);
    }
    // create internal nodes
    // inlining of macro Rtm_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vObjs)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vObjs, i)), 1); i++  { 
        Rtm_ManToAig_rec( pNew, pRtm, pObjRtm, pLatches );
    }
    // create POs
    // inlining of macro Rtm_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vPos)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vPos, i)), 1); i++  {
        Aig_ObjCreateCo( pNew, (Aig_Obj_t *)pObjRtm.pCopy );
    }
    // connect latches 
    // inlining of macro Rtm_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vObjs)) && (((pObjRtm) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vObjs, i)), 1); i++  { 
    // inlining of macro Rtm_ObjForEachFaninEdge
    for  k = 0; k < int32((pObjRtm).nFanins) && ((pEdge = Rtm_ObjEdge(pObjRtm, k)), 1); k++  {
        if pEdge.nLats == 0 {
            continue; }
        pObjNew = (Aig_Obj_t *)Rtm_ObjFanin( pObjRtm, k ).pCopy;
        for  m = 0; m < int32(pEdge.nLats); m++  {
            Val = Rtm_ObjGetOne( pRtm, pEdge, pEdge.nLats - 1 - m );
            assert(Val == RTM_VAL_ZERO || Val == RTM_VAL_ONE || Val == RTM_VAL_VOID, "Val == RTM_VAL_ZERO || Val == RTM_VAL_ONE || Val == RTM_VAL_VOID");
            pObjNew = Aig_NotCond( pObjNew, Val == RTM_VAL_ONE );
            Aig_ObjCreateCo( pNew, pObjNew );
            pObjNew = Aig_ManCi( pNew, pLatches[2*pObjRtm.Id + k] + m );
            pObjNew = Aig_NotCond( pObjNew, Val == RTM_VAL_ONE );
        }
//        assert( Aig_Regular(pObjNew)->nRefs > 0 );
    }
    }
    // inlining of macro ABC_FREE
    ((pLatches) ? (free((char *) (pLatches)), (pLatches) = 0) : 0);
    Aig_ManSetRegNum( pNew, nLatches );
    // remove useless nodes
    Aig_ManCleanup( pNew );
    if !Aig_ManCheck( pNew ) {
        fmt.Printf("Rtm_ManToAig: The network check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Performs forward retiming with the given limit on depth.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rtm_ManRetime(p *Aig_Man_t, fForward int32, nStepsMax int32, fVerbose int32) *Aig_Man_t {
    var vQueue *Vec_Ptr_t;
    var pNew *Aig_Man_t;
    var pRtm *Rtm_Man_t;
    var pObj, pNext *Rtm_Obj_t;
    var pObjAig *Aig_Obj_t;
    var i, k, nAutos, Degree, DegreeMax int32; 
    var clk abctime;

    // create the retiming manager
clk = Abc_Clock();
    pRtm = Rtm_ManFromAig( p );
    // set registers
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObjAig) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Rtm_ObjAddFirst( pRtm, Rtm_ObjEdge((Rtm_Obj_t *)pObjAig.pData, 0), fForward? RTM_VAL_ZERO : RTM_VAL_VOID );
    }
    // detect and mark the autonomous components
    if fForward != 0 {
        nAutos = Rtm_ManMarkAutoFwd( pRtm );
    } else {
        nAutos = Rtm_ManMarkAutoBwd( pRtm ); }
    if fVerbose != 0 {
        fmt.Printf("Detected %d autonomous objects. ", nAutos);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

    // set the current retiming number
    // inlining of macro Rtm_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pRtm.vObjs)) && (((pObj) = (Rtm_Obj_t *)Vec_PtrEntry(pRtm.vObjs, i)), 1); i++  {
        assert(pObj.nFanins == pObj.Num, "pObj.nFanins == pObj.Num");
        assert(pObj.nFanouts == pObj.Temp, "pObj.nFanouts == pObj.Temp");
        pObj.Num = 0;
    }

clk = Abc_Clock();
    // put the LOs on the queue
    vQueue = Vec_PtrAlloc( 1000 );
    if fForward != 0 {
        // inlining of macro Aig_ManForEachLoSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObjAig) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pObj = (Rtm_Obj_t *)pObjAig.pData;
            if pObj.fAuto {
                continue; }
            pObj.fMark = 1;
            Vec_PtrPush( vQueue, pObj );
        }
    
    } else {
        // inlining of macro Aig_ManForEachLiSeq
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = Aig_ManCoNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCos)) && (((pObjAig) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            pObj = (Rtm_Obj_t *)pObjAig.pData;
            if pObj.fAuto {
                continue; }
            pObj.fMark = 1;
            Vec_PtrPush( vQueue, pObj );
        }
    }
    // perform retiming 
    DegreeMax = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vQueue)) && (((pObj) = (Rtm_Obj_t *)Vec_PtrEntry(vQueue, i)), 1); i++  {
        pObj.fMark = 0;
        // retime the node 
        if fForward != 0 {
            Rtm_ObjRetimeFwd( pRtm, pObj );
            // check if its fanouts should be retimed
            // inlining of macro Rtm_ObjForEachFanout
            for  k = 0; k < int32((pObj).nFanouts) && ((pNext = Rtm_ObjFanout(pObj, k)), 1); k++  {
                if pNext.fMark { // skip aleady scheduled
                    continue; }
                if pNext.Type { // skip POs
                    continue; }
                if !(Rtm_ObjCheckRetimeFwd( pNext ) != 0) { // skip non-retimable
                    continue; }
                Degree = Rtm_ObjGetDegreeFwd( pNext );
                DegreeMax = Abc_MaxInt( DegreeMax, Degree );
                if Degree > nStepsMax { // skip nodes with high degree
                    continue; }
                pNext.fMark = 1;
                pNext.Num = Degree;
                Vec_PtrPush( vQueue, pNext );
            }
        
        } else {
            Rtm_ObjRetimeBwd( pRtm, pObj );
            // check if its fanouts should be retimed
            // inlining of macro Rtm_ObjForEachFanin
            for  k = 0; k < int32((pObj).nFanins) && ((pNext = Rtm_ObjFanin(pObj, k)), 1); k++  {
                if pNext.fMark { // skip aleady scheduled
                    continue; }
                if pNext.nFanins == 0 { // skip PIs
                    continue; }
                if !(Rtm_ObjCheckRetimeBwd( pNext ) != 0) { // skip non-retimable
                    continue; }
                Degree = Rtm_ObjGetDegreeBwd( pNext );
                DegreeMax = Abc_MaxInt( DegreeMax, Degree );
                if Degree > nStepsMax { // skip nodes with high degree
                    continue; }
                pNext.fMark = 1;
                pNext.Num = Degree;
                Vec_PtrPush( vQueue, pNext );
            }
        }
    }

    if fVerbose != 0 {
        fmt.Printf("Performed %d %s latch moves of max depth %d and max latch count %d.\n", Vec_PtrSize(vQueue), fForward? "fwd":"bwd", DegreeMax, Rtm_ManLatchMax(pRtm));
        fmt.Printf("Memory usage = %d.  ", pRtm.nExtraCur);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    Vec_PtrFree( vQueue );

    // get the new manager
    pNew = Rtm_ManToAig( pRtm );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Rtm_ManFree( pRtm );
    // group the registers
clk = Abc_Clock();
    pNew = Aig_ManReduceLaches( pNew, fVerbose );
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Register sharing time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    return pNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

