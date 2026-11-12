package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaFanout.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaFanout.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// 0: first iFan
// 1: prev iFan0
// 2: prev iFan1
// 3: next iFan0
// 4: next iFan1

func Gia_FanoutCreate(FanId int32, Num int32) int32 { assert(Num < 2, "Num < 2"); return (FanId << 1) | Num;   } 
func Gia_FanoutObj(pData *int32, ObjId int32) *int32 { return pData + 5*ObjId;                         }
func Gia_FanoutPrev(pData *int32, iFan int32) *int32 { return pData + 5*(iFan >> 1) + 1 + (iFan & 1);  }
func Gia_FanoutNext(pData *int32, iFan int32) *int32 { return pData + 5*(iFan >> 1) + 3 + (iFan & 1);  }

// these two procedures are only here for the use inside the iterator
func Gia_ObjFanout0Int(p *Gia_Man_t, ObjId int32) int32 { assert(ObjId < p.nFansAlloc, "ObjId < p.nFansAlloc");  return p.pFanData[5*ObjId];                         }
func Gia_ObjFanoutNext(p *Gia_Man_t, iFan int32) int32 { assert(iFan/2 < p.nFansAlloc, "iFan/2 < p.nFansAlloc"); return p.pFanData[5*(iFan >> 1) + 3 + (iFan & 1)];  }

// iterator over the fanouts
// Original macro:
// #define Gia_ObjForEachFanout( p, pObj, pFanout, iFan, i )                       \
//     for ( assert(p->pFanData), i = 0; (i < (int)(pObj)->nRefs) &&               \
//           (((iFan) = i? Gia_ObjFanoutNext(p, iFan) : Gia_ObjFanout0Int(p, Gia_ObjId(p, pObj))), 1) && \
//           (((pFanout) = Gia_ManObj(p, iFan>>1)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create fanout for all objects in the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFanoutStart(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // allocate fanout datastructure
    assert(p.pFanData == nil, "p.pFanData == nil");
    p.nFansAlloc = 2 * Gia_ManObjNum(p);
    if p.nFansAlloc < (1<<12) {
        p.nFansAlloc = (1<<12); }
    // inlining of macro ABC_ALLOC
    p.pFanData = ((int *) malloc(sizeof(int) * (size_t)(5 * p.nFansAlloc)));
    memset( p.pFanData, 0, sizeof(int) * 5 * p.nFansAlloc );
    // add fanouts for all objects
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjChild0(pObj) {
            Gia_ObjAddFanout( p, Gia_ObjFanin0(pObj), pObj ); }
        if Gia_ObjChild1(pObj) {
            Gia_ObjAddFanout( p, Gia_ObjFanin1(pObj), pObj ); }
    }
}

/**Function*************************************************************

  Synopsis    [Deletes fanout for all objects in the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFanoutStop(p *Gia_Man_t) {
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
func Gia_ObjAddFanout(p *Gia_Man_t, pObj *Gia_Obj_t, pFanout *Gia_Obj_t) {
    var iFan int32;
    var pFirst *int32;
    var pPrevC *int32;
    var pNextC *int32;
    var pPrev *int32;
    var pNext *int32;
    assert(p.pFanData, "p.pFanData");
    assert(!Gia_IsComplement(pObj) && !Gia_IsComplement(pFanout), "!Gia_IsComplement(pObj) && !Gia_IsComplement(pFanout)");
    assert(Gia_ObjId(p, pFanout) > 0, "Gia_ObjId(p, pFanout) > 0");
    if Gia_ObjId(p, pObj) >= p.nFansAlloc || Gia_ObjId(p, pFanout) >= p.nFansAlloc {
        var nFansAlloc int32  = 2 * Abc_MaxInt( Gia_ObjId(p, pObj), Gia_ObjId(p, pFanout) ); 
        // inlining of macro ABC_REALLOC
        
        if p.pFanData {
            p.pFanData = ((int *) realloc((char *)(p.pFanData), sizeof(int) * (size_t)(5 * nFansAlloc)));
        } else {
            p.pFanData = ((int *) malloc(sizeof(int) * (size_t)(5 * nFansAlloc)));
        }
        memset( p.pFanData + 5 * p.nFansAlloc, 0, sizeof(int) * 5 * (nFansAlloc - p.nFansAlloc) );
        p.nFansAlloc = nFansAlloc;
    }
    assert(Gia_ObjId(p, pObj) < p.nFansAlloc && Gia_ObjId(p, pFanout) < p.nFansAlloc, "Gia_ObjId(p, pObj) < p.nFansAlloc && Gia_ObjId(p, pFanout) < p.nFansAlloc");
    iFan   = Gia_FanoutCreate( Gia_ObjId(p, pFanout), Gia_ObjWhatFanin(p, pFanout, pObj) );
    pPrevC = Gia_FanoutPrev( p.pFanData, iFan );
    pNextC = Gia_FanoutNext( p.pFanData, iFan );
    pFirst = Gia_FanoutObj( p.pFanData, Gia_ObjId(p, pObj) );
    if *pFirst == 0 {
        *pFirst = iFan;
        *pPrevC = iFan;
        *pNextC = iFan;
    
    } else {
        pPrev = Gia_FanoutPrev( p.pFanData, *pFirst );
        pNext = Gia_FanoutNext( p.pFanData, *pPrev );
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
func Gia_ObjRemoveFanout(p *Gia_Man_t, pObj *Gia_Obj_t, pFanout *Gia_Obj_t) {
    var iFan int32;
    var pFirst *int32;
    var pPrevC *int32;
    var pNextC *int32;
    var pPrev *int32;
    var pNext *int32;
    assert(p.pFanData && Gia_ObjId(p, pObj) < p.nFansAlloc && Gia_ObjId(p, pFanout) < p.nFansAlloc, "p.pFanData && Gia_ObjId(p, pObj) < p.nFansAlloc && Gia_ObjId(p, pFanout) < p.nFansAlloc");
    assert(!Gia_IsComplement(pObj) && !Gia_IsComplement(pFanout), "!Gia_IsComplement(pObj) && !Gia_IsComplement(pFanout)");
    assert(Gia_ObjId(p, pFanout) > 0, "Gia_ObjId(p, pFanout) > 0");
    iFan   = Gia_FanoutCreate( Gia_ObjId(p, pFanout), Gia_ObjWhatFanin(p, pFanout, pObj) );
    pPrevC = Gia_FanoutPrev( p.pFanData, iFan );
    pNextC = Gia_FanoutNext( p.pFanData, iFan );
    pPrev  = Gia_FanoutPrev( p.pFanData, *pNextC );
    pNext  = Gia_FanoutNext( p.pFanData, *pPrevC );
    assert(*pPrev == iFan, "*pPrev == iFan");
    assert(*pNext == iFan, "*pNext == iFan");
    pFirst = Gia_FanoutObj( p.pFanData, Gia_ObjId(p, pObj) );
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




/**Function*************************************************************

  Synopsis    [Compute the map of all edges.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStartFanoutMap(p *Gia_Man_t, vFanoutNums *Vec_Int_t) *Vec_Int_t {
    var vEdgeMap *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, iOffset int32;
    iOffset  = Gia_ManObjNum(p);
    vEdgeMap = Vec_IntStart( iOffset + Gia_ManMuxNum(p) + 2 * Gia_ManAndNum(p) + Gia_ManCoNum(p) - Gia_ManBufNum(p) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Vec_IntWriteEntry( vEdgeMap, i, iOffset );
        iOffset += Vec_IntEntry( vFanoutNums, Gia_ObjId(p, pObj) );
    }
    assert(iOffset <= Vec_IntSize(vEdgeMap), "iOffset <= Vec_IntSize(vEdgeMap)");
    return vEdgeMap;
}

/**Function*************************************************************

  Synopsis    [Allocates static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStaticFanoutStart(p *Gia_Man_t) {
    var vCounts *Vec_Int_t;
    var pRefsOld *int32;
    var pObj, pFanin *Gia_Obj_t;
    var i, iFanout int32;
    assert(p.vFanoutNums == nil, "p.vFanoutNums == nil");
    assert(p.vFanout == nil, "p.vFanout == nil");
    // recompute reference counters
    pRefsOld = p.pRefs; p.pRefs = nil;
    Gia_ManCreateRefs(p);
    p.vFanoutNums = Vec_IntAllocArray( p.pRefs, Gia_ManObjNum(p) );
    p.pRefs = pRefsOld;
    // start the fanout maps
    p.vFanout = Gia_ManStartFanoutMap( p, p.vFanoutNums );
    // incrementally add fanouts
    vCounts = Vec_IntStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj) {
            pFanin = Gia_ObjFanin0(pObj);
            iFanout = Vec_IntEntry( vCounts, Gia_ObjId(p, pFanin) );
            Gia_ObjSetFanout( p, pFanin, iFanout, pObj );
            Vec_IntAddToEntry( vCounts, Gia_ObjId(p, pFanin), 1 );
        }
        if Gia_ObjIsAnd(pObj) && !Gia_ObjIsBuf(pObj) {
            pFanin = Gia_ObjFanin1(pObj);
            iFanout = Vec_IntEntry( vCounts, Gia_ObjId(p, pFanin) );
            Gia_ObjSetFanout( p, pFanin, iFanout, pObj );
            Vec_IntAddToEntry( vCounts, Gia_ObjId(p, pFanin), 1 );
        }
        if Gia_ObjIsMux(p, pObj) {

            pFanin = Gia_ObjFanin2(p, pObj);
            iFanout = Vec_IntEntry( vCounts, Gia_ObjId(p, pFanin) );
            Gia_ObjSetFanout( p, pFanin, iFanout, pObj );
            Vec_IntAddToEntry( vCounts, Gia_ObjId(p, pFanin), 1 );
        }
    }
    // double-check the current number of fanouts added
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        assert(Vec_IntEntry(vCounts, i) == Gia_ObjFanoutNum(p, pObj), "Vec_IntEntry(vCounts, i) == Gia_ObjFanoutNum(p, pObj)");
    }
    Vec_IntFree( vCounts );
}


/**Function*************************************************************

  Synopsis    [Compute the map of all edges.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStartMappingFanoutMap(p *Gia_Man_t, vFanoutNums *Vec_Int_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var i int32;
    var iOffset int32 = Gia_ManObjNum(p);
    var vEdgeMap *Vec_Int_t  = Vec_IntAlloc( 2 * iOffset );
    Vec_IntFill( vEdgeMap, iOffset, 0 );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Vec_IntEntry(vFanoutNums, i) == 0 {
            continue; }
        Vec_IntWriteEntry( vEdgeMap, i, iOffset );
        iOffset += Vec_IntEntry( vFanoutNums, i );
        Vec_IntFillExtra( vEdgeMap, iOffset, 0 );
    }
    //printf( "Fanout map is %.2fx larger than AIG manager.\n", 1.0*Vec_IntSize(vEdgeMap)/Gia_ManObjNum(p) );
    return vEdgeMap;
}

/**Function*************************************************************

  Synopsis    [Allocates static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStaticMappingFanoutStart(p *Gia_Man_t) {
    var vCounts *Vec_Int_t;
    var pRefsOld *int32;
    var pObj, pFanin *Gia_Obj_t;
    var i, k, iFan, iFanout int32;
    assert(p.vFanoutNums == nil, "p.vFanoutNums == nil");
    assert(p.vFanout == nil, "p.vFanout == nil");
    // recompute reference counters
    pRefsOld = p.pLutRefs; p.pLutRefs = nil;
    Gia_ManSetLutRefs(p);
    p.vFanoutNums = Vec_IntAllocArray( p.pLutRefs, Gia_ManObjNum(p) );
    p.pLutRefs = pRefsOld;
    // start the fanout maps
    p.vFanout = Gia_ManStartMappingFanoutMap( p, p.vFanoutNums );
    // incrementally add fanouts
    vCounts = Vec_IntStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
        pObj = Gia_ManObj( p, i );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++  {
            pFanin = Gia_ManObj( p, iFan );
            iFanout = Vec_IntEntry( vCounts, iFan );
            Gia_ObjSetFanout( p, pFanin, iFanout, pObj );
            Vec_IntAddToEntry( vCounts, iFan, 1 );
        }
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        iFan = Gia_ObjFaninId0p(p, pObj);
        pFanin = Gia_ManObj( p, iFan );
        iFanout = Vec_IntEntry( vCounts, iFan );
        Gia_ObjSetFanout( p, pFanin, iFanout, pObj );
        Vec_IntAddToEntry( vCounts, iFan, 1 );
    }
    // double-check the current number of fanouts added
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        assert(Vec_IntEntry(vCounts, i) == Gia_ObjFanoutNum(p, pObj), "Vec_IntEntry(vCounts, i) == Gia_ObjFanoutNum(p, pObj)");
    }
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    [Deallocates static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStaticFanoutStop(p *Gia_Man_t) {
    Vec_IntFreeP( &p.vFanoutNums );
    Vec_IntFreeP( &p.vFanout );
}


/**Function*************************************************************

  Synopsis    [Tests static fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManStaticFanoutTest(p *Gia_Man_t) {
    var pObj, pFanout *Gia_Obj_t;
    var i, k int32;
    Gia_ManStaticFanoutStart( p );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Gia_ObjPrint( p, pObj );
        fmt.Printf("   Fanouts : ");
        // inlining of macro Gia_ObjForEachFanoutStatic
        for  k = 0; (k < Gia_ObjFanoutNum(p, pObj))   && (((pFanout) = Gia_ObjFanout(p, pObj, k)), 1); k++  {
            fmt.Printf("%5d ", Gia_ObjId(p, pFanout));
        }
        fmt.Printf("\n");
    }
    Gia_ManStaticFanoutStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

