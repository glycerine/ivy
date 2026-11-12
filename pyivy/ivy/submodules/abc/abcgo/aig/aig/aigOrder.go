package aig

/**CFile****************************************************************

  FileName    [aigOrder.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Dynamically updated topological order.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigOrder.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Initializes the order datastructure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManOrderStart(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Aig_ManBufNum(p) == 0, "Aig_ManBufNum(p) == 0");
    // allocate order datastructure
    assert(p.pOrderData == nil, "p.pOrderData == nil");
    p.nOrderAlloc = 2 * Aig_ManObjNumMax(p);
    if p.nOrderAlloc < (1<<12) {
        p.nOrderAlloc = (1<<12); }
    // inlining of macro ABC_ALLOC
    p.pOrderData = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(2 * p.nOrderAlloc)));
    memset( p.pOrderData, 0xFF, sizeof(unsigned) * 2 * p.nOrderAlloc );
    // add the constant node
    p.pOrderData[0] = p.pOrderData[1] = 0;
    p.iPrev = p.iNext = 0;
    // add the internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Aig_ObjOrderInsert( p, pObj.Id ); }
    }
}

/**Function*************************************************************

  Synopsis    [Deletes the order datastructure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManOrderStop(p *Aig_Man_t) {
    assert(p.pOrderData, "p.pOrderData");
    // inlining of macro ABC_FREE
    ((p.pOrderData) ? (free((char *) (p.pOrderData)), (p.pOrderData) = 0) : 0);
    p.nOrderAlloc = 0;
    p.iPrev = p.iNext = 0;
}

/**Function*************************************************************

  Synopsis    [Inserts an entry before iNext.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjOrderInsert(p *Aig_Man_t, ObjId int32) {
    var iPrev int32;
    assert(ObjId != 0, "ObjId != 0");
    assert(Aig_ObjIsNode( Aig_ManObj(p, ObjId) ), "Aig_ObjIsNode( Aig_ManObj(p, ObjId) )");
    if ObjId >= p.nOrderAlloc {
        var nOrderAlloc int32  = 2 * ObjId; 
        // inlining of macro ABC_REALLOC
        
        if p.pOrderData {
            p.pOrderData = ((unsigned *) realloc((char *)(p.pOrderData), sizeof(unsigned) * (size_t)(2 * nOrderAlloc)));
        } else {
            p.pOrderData = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(2 * nOrderAlloc)));
        }
        memset( p.pOrderData + 2 * p.nOrderAlloc, 0xFF, sizeof(unsigned) * 2 * (nOrderAlloc - p.nOrderAlloc) );
        p.nOrderAlloc = nOrderAlloc;
    }
    assert(p.pOrderData[2*ObjId] == 0xFFFFFFFF, "p.pOrderData[2*ObjId] == 0xFFFFFFFF");   // prev
    assert(p.pOrderData[2*ObjId+1] == 0xFFFFFFFF, "p.pOrderData[2*ObjId+1] == 0xFFFFFFFF"); // next
    iPrev = p.pOrderData[2*p.iNext];
    assert(p.pOrderData[2*iPrev+1] == uint32(p.iNext), "p.pOrderData[2*iPrev+1] == uint32(p.iNext)");
    p.pOrderData[2*ObjId] = iPrev;
    p.pOrderData[2*iPrev+1] = ObjId;
    p.pOrderData[2*p.iNext] = ObjId;
    p.pOrderData[2*ObjId+1] = p.iNext;
    p.nAndTotal++;
}

/**Function*************************************************************

  Synopsis    [Removes the entry.]

  Description [If iPrev is removed, it slides backward. 
  If iNext is removed, it slides forward.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjOrderRemove(p *Aig_Man_t, ObjId int32) {
    var iPrev, iNext int32;
    assert(ObjId != 0, "ObjId != 0");
    assert(Aig_ObjIsNode( Aig_ManObj(p, ObjId) ), "Aig_ObjIsNode( Aig_ManObj(p, ObjId) )");
    iPrev = p.pOrderData[2*ObjId];
    iNext = p.pOrderData[2*ObjId+1];
    p.pOrderData[2*ObjId] = 0xFFFFFFFF;
    p.pOrderData[2*ObjId+1] = 0xFFFFFFFF;
    p.pOrderData[2*iNext] = iPrev;
    p.pOrderData[2*iPrev+1] = iNext;
    if p.iPrev == ObjId {
        p.nAndPrev--;
        p.iPrev = iPrev;
    }
    if p.iNext == ObjId {
        p.iNext = iNext; }
    p.nAndTotal--;
}

/**Function*************************************************************

  Synopsis    [Advances the order forward.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjOrderAdvance(p *Aig_Man_t) {
    assert(p.pOrderData, "p.pOrderData");
    assert(p.pOrderData[2*p.iPrev+1] == uint32(p.iNext), "p.pOrderData[2*p.iPrev+1] == uint32(p.iNext)");
    p.iPrev = p.iNext;
    p.nAndPrev++;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

