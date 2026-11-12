package aig

/**CFile****************************************************************

  FileName    [aigFanout.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Fanout manipulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigFanout.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


// 0: first iFan
// 1: prev iFan0
// 2: prev iFan1
// 3: next iFan0
// 4: next iFan1

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Aig_FanoutCreate(FanId int32, Num int32) int32 { assert(Num < 2, "Num < 2"); return (FanId << 1) | Num;   } 
func Aig_FanoutObj(pData *int32, ObjId int32) *int32 { return pData + 5*ObjId;                         }
func Aig_FanoutPrev(pData *int32, iFan int32) *int32 { return pData + 5*(iFan >> 1) + 1 + (iFan & 1);  }
func Aig_FanoutNext(pData *int32, iFan int32) *int32 { return pData + 5*(iFan >> 1) + 3 + (iFan & 1);  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create fanout for all objects in the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFanoutStart(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Aig_ManBufNum(p) == 0, "Aig_ManBufNum(p) == 0");
    // allocate fanout datastructure
    assert(p.pFanData == nil, "p.pFanData == nil");
    p.nFansAlloc = 2 * Aig_ManObjNumMax(p);
    if p.nFansAlloc < (1<<12) {
        p.nFansAlloc = (1<<12); }
    // inlining of macro ABC_ALLOC
    p.pFanData = ((int *) malloc(sizeof(int) * (size_t)(5 * p.nFansAlloc)));
    memset( p.pFanData, 0, sizeof(int) * 5 * p.nFansAlloc );
    // add fanouts for all objects
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjChild0(pObj) {
            Aig_ObjAddFanout( p, Aig_ObjFanin0(pObj), pObj ); }
        if Aig_ObjChild1(pObj) {
            Aig_ObjAddFanout( p, Aig_ObjFanin1(pObj), pObj ); }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Deletes fanout for all objects in the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFanoutStop(p *Aig_Man_t) {
    assert(p.pFanData != nil, "p.pFanData != nil");
    // inlining of macro ABC_FREE
    ((p.pFanData) ? (free((char *) (p.pFanData)), (p.pFanData) = 0) : 0);
    p.nFansAlloc = 0;
}

/**Function*************************************************************

  Synopsis    [Adds fanout (pFanout) of node (pObj).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjAddFanout(p *Aig_Man_t, pObj *Aig_Obj_t, pFanout *Aig_Obj_t) {
    var iFan int32;
    var pFirst *int32;
    var pPrevC *int32;
    var pNextC *int32;
    var pPrev *int32;
    var pNext *int32;
    assert(p.pFanData, "p.pFanData");
    assert(!Aig_IsComplement(pObj) && !Aig_IsComplement(pFanout), "!Aig_IsComplement(pObj) && !Aig_IsComplement(pFanout)");
    assert(pFanout.Id > 0, "pFanout.Id > 0");
    if pObj.Id >= p.nFansAlloc || pFanout.Id >= p.nFansAlloc {
        var nFansAlloc int32  = 2 * Abc_MaxInt( pObj.Id, pFanout.Id ); 
        // inlining of macro ABC_REALLOC
        
        if p.pFanData {
            p.pFanData = ((int *) realloc((char *)(p.pFanData), sizeof(int) * (size_t)(5 * nFansAlloc)));
        } else {
            p.pFanData = ((int *) malloc(sizeof(int) * (size_t)(5 * nFansAlloc)));
        }
        memset( p.pFanData + 5 * p.nFansAlloc, 0, sizeof(int) * 5 * (nFansAlloc - p.nFansAlloc) );
        p.nFansAlloc = nFansAlloc;
    }
    assert(pObj.Id < p.nFansAlloc && pFanout.Id < p.nFansAlloc, "pObj.Id < p.nFansAlloc && pFanout.Id < p.nFansAlloc");
    iFan   = Aig_FanoutCreate( pFanout.Id, Aig_ObjWhatFanin(pFanout, pObj) );
    pPrevC = Aig_FanoutPrev( p.pFanData, iFan );
    pNextC = Aig_FanoutNext( p.pFanData, iFan );
    pFirst = Aig_FanoutObj( p.pFanData, pObj.Id );
    if *pFirst == 0 {
        *pFirst = iFan;
        *pPrevC = iFan;
        *pNextC = iFan;
    
    } else {
        pPrev = Aig_FanoutPrev( p.pFanData, *pFirst );
        pNext = Aig_FanoutNext( p.pFanData, *pPrev );
        assert(*pNext == *pFirst, "*pNext == *pFirst");
        *pPrevC = *pPrev;
        *pNextC = *pFirst;
        *pPrev  = iFan;
        *pNext  = iFan;
    }
}

/**Function*************************************************************

  Synopsis    [Removes fanout (pFanout) of node (pObj).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjRemoveFanout(p *Aig_Man_t, pObj *Aig_Obj_t, pFanout *Aig_Obj_t) {
    var iFan int32;
    var pFirst *int32;
    var pPrevC *int32;
    var pNextC *int32;
    var pPrev *int32;
    var pNext *int32;
    assert(p.pFanData && pObj.Id < p.nFansAlloc && pFanout.Id < p.nFansAlloc, "p.pFanData && pObj.Id < p.nFansAlloc && pFanout.Id < p.nFansAlloc");
    assert(!Aig_IsComplement(pObj) && !Aig_IsComplement(pFanout), "!Aig_IsComplement(pObj) && !Aig_IsComplement(pFanout)");
    assert(pFanout.Id > 0, "pFanout.Id > 0");
    iFan   = Aig_FanoutCreate( pFanout.Id, Aig_ObjWhatFanin(pFanout, pObj) );
    pPrevC = Aig_FanoutPrev( p.pFanData, iFan );
    pNextC = Aig_FanoutNext( p.pFanData, iFan );
    pPrev  = Aig_FanoutPrev( p.pFanData, *pNextC );
    pNext  = Aig_FanoutNext( p.pFanData, *pPrevC );
    assert(*pPrev == iFan, "*pPrev == iFan");
    assert(*pNext == iFan, "*pNext == iFan");
    pFirst = Aig_FanoutObj( p.pFanData, pObj.Id );
    assert(*pFirst > 0, "*pFirst > 0");
    if *pFirst == iFan {
        if *pNextC == iFan {
            *pFirst = 0;
            *pPrev  = 0;
            *pNext  = 0;
            *pPrevC = 0;
            *pNextC = 0;
            return;
        }
        *pFirst = *pNextC;
    }
    *pPrev  = *pPrevC;
    *pNext  = *pNextC;
    *pPrevC = 0;
    *pNextC = 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

