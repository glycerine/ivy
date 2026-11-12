package gia

import (
	"fmt"
	"math"
	"time"
)

/**CFile****************************************************************

  FileName    [giaCof.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Cofactor estimation procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaCof.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Cof_Fan_t struct {
    iFan uint32; /* C bitfield: 31 bits; review packing and truncation */    // ID of the fanin/fanout
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */    // complemented attribute
}

   
type Cof_Obj_t struct {
    fTerm uint32; /* C bitfield: 1 bits; review packing and truncation */    // terminal node (CI/CO)
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */    // value under 000 pattern
    fMark0 uint32; /* C bitfield: 1 bits; review packing and truncation */    // first user-controlled mark
    fMark1 uint32; /* C bitfield: 1 bits; review packing and truncation */    // second user-controlled mark
    nFanins uint32; /* C bitfield: 4 bits; review packing and truncation */    // the number of fanins
    nFanouts uint32; /* C bitfield: 24 bits; review packing and truncation */    // total number of fanouts
    nFanoutsM       uint32;        // total number of MUX ctrl fanouts
    Value       uint32;            // application specific data
    Id            int32;               // ID of the node
    iNext            int32;            // next one in the linked list
    iLit            int32;             // literal of the node after rehashing
    Fanios [0]Cof_Fan_t;        // the array of fanins/fanouts
}

   
type Cof_Man_t struct {
    pGia     *Gia_Man_t;             // the original AIG manager
    vCis     *Vec_Int_t;             // the vector of CIs (PIs + LOs)
    vCos     *Vec_Int_t;             // the vector of COs (POs + LIs)
    nObjs            int32;            // the number of objects
    nNodes            int32;           // the number of nodes
    nTravIds            int32;         // traversal ID of the network
    pObjData           *int32;         // the logic network defined for the AIG
    nObjData            int32;         // the size of array to store the logic network
    pLevels           *int32;          // the linked lists of levels
    nLevels            int32;          // the max number of logic levels
}

func Gia_ObjHandle(pObj *Gia_Obj_t) uint32 { return pObj.Value;                                 } 

func Cof_ObjLevel(p *Cof_Man_t, pObj *Cof_Obj_t) int32 { return Gia_ObjLevel(p.pGia, Gia_ManObj(p.pGia,pObj.Id)); } 

func Cof_ObjHandle(p *Cof_Man_t, pObj *Cof_Obj_t) uint32 { return uint32((((int *)pObj) - p.pObjData));     } 
func Cof_ObjHandleDiff(pObj *Cof_Obj_t, pFanin *Cof_Obj_t) uint32 { return uint32((((int *)pObj) - ((int *)pFanin))); } 

func Cof_ObjIsTerm(pObj *Cof_Obj_t) int32 { return pObj.fTerm;                                 } 
func Cof_ObjIsCi(pObj *Cof_Obj_t) int32 { return pObj.fTerm && pObj.nFanins == 0;           } 
func Cof_ObjIsCo(pObj *Cof_Obj_t) int32 { return pObj.fTerm && pObj.nFanins == 1;           } 
func Cof_ObjIsNode(pObj *Cof_Obj_t) int32 { return!pObj.fTerm && pObj.nFanins > 0;            } 
func Cof_ObjIsConst0(pObj *Cof_Obj_t) int32 { return!pObj.fTerm && pObj.nFanins == 0;           } 

func Cof_ObjFaninNum(pObj *Cof_Obj_t) int32 { return pObj.nFanins;                               } 
func Cof_ObjFanoutNum(pObj *Cof_Obj_t) int32 { return pObj.nFanouts;                              } 
func Cof_ObjSize(pObj *Cof_Obj_t) int32 { return sizeof(Cof_Obj_t) / 4 + pObj.nFanins + pObj.nFanouts;  } 

func Cof_ManObj(p *Cof_Man_t, iHandle uint32) *Cof_Obj_t { return (Cof_Obj_t *)(p.pObjData + iHandle);        } 
func Cof_ObjFanin(pObj *Cof_Obj_t, i int32) *Cof_Obj_t { return (Cof_Obj_t *)(((int *)pObj) - pObj.Fanios[i].iFan);               } 
func Cof_ObjFanout(pObj *Cof_Obj_t, i int32) *Cof_Obj_t { return (Cof_Obj_t *)(((int *)pObj) + pObj.Fanios[pObj.nFanins+i].iFan); } 

func Cof_ManObjNum(p *Cof_Man_t) int32 { return p.nObjs;                                    } 
func Cof_ManNodeNum(p *Cof_Man_t) int32 { return p.nNodes;                                   } 

func Cof_ManResetTravId(p *Cof_Man_t) { extern void Cof_ManCleanValue( Cof_Man_t * p ); Cof_ManCleanValue( p ); p.nTravIds = 1;  }
func Cof_ManIncrementTravId(p *Cof_Man_t) { p.nTravIds++;                                      }
func Cof_ObjSetTravId(pObj *Cof_Obj_t, TravId int32) { pObj.Value = TravId;                               }
func Cof_ObjSetTravIdCurrent(p *Cof_Man_t, pObj *Cof_Obj_t) { pObj.Value = p.nTravIds;                          }
func Cof_ObjSetTravIdPrevious(p *Cof_Man_t, pObj *Cof_Obj_t) { pObj.Value = p.nTravIds - 1;                      }
func Cof_ObjIsTravIdCurrent(p *Cof_Man_t, pObj *Cof_Obj_t) int32 { return (int32(pObj.Value) == p.nTravIds);           }
func Cof_ObjIsTravIdPrevious(p *Cof_Man_t, pObj *Cof_Obj_t) int32 { return (int32(pObj.Value) == p.nTravIds - 1);       }

// Original macro:
// #define Cof_ManForEachObj( p, pObj, i )               \
//     for ( i = 0; (i < p->nObjData) && (pObj = Cof_ManObj(p,i)); i += Cof_ObjSize(pObj) )

// Original macro:
// #define Cof_ManForEachNode( p, pObj, i )              \
//     for ( i = 0; (i < p->nObjData) && (pObj = Cof_ManObj(p,i)); i += Cof_ObjSize(pObj) ) if ( Cof_ObjIsTerm(pObj) ) {} else

// Original macro:
// #define Cof_ObjForEachFanin( pObj, pNext, i )         \
//     for ( i = 0; (i < (int)pObj->nFanins) && (pNext = Cof_ObjFanin(pObj,i)); i++ )

// Original macro:
// #define Cof_ObjForEachFanout( pObj, pNext, i )        \
//     for ( i = 0; (i < (int)pObj->nFanouts) && (pNext = Cof_ObjFanout(pObj,i)); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManCreateLogicSimple(pGia *Gia_Man_t) *Cof_Man_t {
    
    var pObjLog, pFanLog *Cof_Obj_t;
    var pObj *Gia_Obj_t;
    var pMuxRefs *int32;
    var i, iHandle int32;
    // inlining of macro ABC_CALLOC
    p := &Cof_Man_t{};
    p.pGia = pGia;
    p.vCis = Vec_IntAlloc( Gia_ManCiNum(pGia) );
    p.vCos = Vec_IntAlloc( Gia_ManCoNum(pGia) );
    p.nObjData = (sizeof(Cof_Obj_t) / 4) * Gia_ManObjNum(pGia) + 4 * Gia_ManAndNum(pGia) + 2 * Gia_ManCoNum(pGia);
    // inlining of macro ABC_CALLOC
    p.pObjData = ((int *) calloc((size_t)(p.nObjData), sizeof(int)));
    // inlining of macro ABC_FREE
    pGia.Refs = nil;
    Gia_ManCreateRefs( pGia );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        pObj.Value = iHandle;
        pObjLog = Cof_ManObj( p, iHandle );
        pObjLog.nFanins  = 0;
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        pObjLog.Id       = i;
        pObjLog.Value    = 0;
        if Gia_ObjIsAnd(pObj) {
            pFanLog = Cof_ManObj( p, Gia_ObjHandle(Gia_ObjFanin0(pObj)) ); 
            pFanLog.Fanios[pFanLog.nFanins + pFanLog.Value++].iFan = 
                pObjLog.Fanios[pObjLog.nFanins].iFan = Cof_ObjHandleDiff( pObjLog, pFanLog );
            pObjLog.Fanios[pObjLog.nFanins++].fCompl = Gia_ObjFaninC0(pObj);

            pFanLog = Cof_ManObj( p, Gia_ObjHandle(Gia_ObjFanin1(pObj)) ); 
            pFanLog.Fanios[pFanLog.nFanins + pFanLog.Value++].iFan = 
                pObjLog.Fanios[pObjLog.nFanins].iFan = Cof_ObjHandleDiff( pObjLog, pFanLog );
            pObjLog.Fanios[pObjLog.nFanins++].fCompl = Gia_ObjFaninC1(pObj);
            p.nNodes++;
        
        } else if Gia_ObjIsCo(pObj) {
            pFanLog = Cof_ManObj( p, Gia_ObjHandle(Gia_ObjFanin0(pObj)) ); 
            pFanLog.Fanios[pFanLog.nFanins + pFanLog.Value++].iFan = 
                pObjLog.Fanios[pObjLog.nFanins].iFan = Cof_ObjHandleDiff( pObjLog, pFanLog );
            pObjLog.Fanios[pObjLog.nFanins++].fCompl = Gia_ObjFaninC0(pObj);

            pObjLog.fTerm = 1;
            Vec_IntPush( p.vCos, iHandle );
        
        } else if Gia_ObjIsCi(pObj) {
            pObjLog.fTerm = 1;
            Vec_IntPush( p.vCis, iHandle );
        }
        iHandle += Cof_ObjSize( pObjLog );
        p.nObjs++;
    }
    assert(iHandle == p.nObjData, "iHandle == p.nObjData");
    pMuxRefs = Gia_ManCreateMuxRefs( pGia );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        pObjLog = Cof_ManObj( p, Gia_ObjHandle(pObj) );
        assert(pObjLog.nFanouts == pObjLog.Value, "pObjLog.nFanouts == pObjLog.Value");
        pObjLog.nFanoutsM = pMuxRefs[i];
    }
    // inlining of macro ABC_FREE
    pMuxRefs = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManStop(p *Cof_Man_t) {
    Vec_IntFree( p.vCis );
    Vec_IntFree( p.vCos );
    // inlining of macro ABC_FREE
    p.pObjData = nil;
    // inlining of macro ABC_FREE
    p.pLevels = nil;
    // inlining of macro ABC_FREE
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManTfoSize_rec(p *Cof_Man_t, pObj *Cof_Obj_t) int32 {
    var pNext *Cof_Obj_t;
    var i, Counter uint32;
    if Cof_ObjIsTravIdCurrent(p, pObj) != 0 {
        return 0; }
    Cof_ObjSetTravIdCurrent(p, pObj);
    if Cof_ObjIsCo(pObj) != 0 {
        return 0; }
    assert((Cof_ObjIsCi(pObj) != 0) || (Cof_ObjIsNode(pObj) != 0), "Cof_ObjIsCi(pObj) || Cof_ObjIsNode(pObj)");
    // inlining of macro Cof_ObjForEachFanout
    for  i = 0; (i < int32(pObj.nFanouts)) && (pNext = Cof_ObjFanout(pObj,i)); i++  {
        Counter += Cof_ManTfoSize_rec( p, pNext );
    }
    return 1 + Counter;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManTfoSize(p *Cof_Man_t, ppObjs **Cof_Obj_t, nObjs int32) int32 {
    var i, Counter int32; 
    Cof_ManIncrementTravId( p );
    for  i = 0; i < nObjs; i++  {
        Counter += Cof_ManTfoSize_rec( p, ppObjs[i] ) - 1;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManTfiSize_rec(p *Cof_Man_t, pObj *Cof_Obj_t) int32 {
    var pNext *Cof_Obj_t;
    var i, Counter uint32;
    if Cof_ObjIsTravIdCurrent(p, pObj) != 0 {
        return 0; }
    Cof_ObjSetTravIdCurrent(p, pObj);
    if Cof_ObjIsCi(pObj) != 0 {
        return 0; }
    assert((Cof_ObjIsNode(pObj) != 0), "Cof_ObjIsNode(pObj)");
    // inlining of macro Cof_ObjForEachFanin
    for  i = 0; (i < int32(pObj.nFanins)) && (pNext = Cof_ObjFanin(pObj,i)); i++  {
        Counter += Cof_ManTfiSize_rec( p, pNext );
    }
    return 1 + Counter;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManTfiSize(p *Cof_Man_t, ppObjs **Cof_Obj_t, nObjs int32) int32 {
    var i, Counter int32; 
    Cof_ManIncrementTravId( p );
    for  i = 0; i < nObjs; i++  {
        if Cof_ObjIsCo(ppObjs[i]) != 0 {
            Counter += Cof_ManTfiSize_rec( p, Cof_ObjFanin(ppObjs[i],0) );
        } else {
            Counter += Cof_ManTfiSize_rec( p, ppObjs[i] ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManSuppSize_rec(p *Cof_Man_t, pObj *Cof_Obj_t) int32 {
    var pNext *Cof_Obj_t;
    var i, Counter uint32;
    if Cof_ObjIsTravIdCurrent(p, pObj) != 0 {
        return 0; }
    Cof_ObjSetTravIdCurrent(p, pObj);
    if Cof_ObjIsCi(pObj) != 0 {
        return 1; }
    assert((Cof_ObjIsNode(pObj) != 0), "Cof_ObjIsNode(pObj)");
    // inlining of macro Cof_ObjForEachFanin
    for  i = 0; (i < int32(pObj.nFanins)) && (pNext = Cof_ObjFanin(pObj,i)); i++  {
        Counter += Cof_ManSuppSize_rec( p, pNext );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManSuppSize(p *Cof_Man_t, ppObjs **Cof_Obj_t, nObjs int32) int32 {
    var i, Counter int32; 
    Cof_ManIncrementTravId( p );
    for  i = 0; i < nObjs; i++  {
        if Cof_ObjIsCo(ppObjs[i]) != 0 {
            Counter += Cof_ManSuppSize_rec( p, Cof_ObjFanin(ppObjs[i],0) );
        } else {
            Counter += Cof_ManSuppSize_rec( p, ppObjs[i] ); }
    }
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManCleanValue(p *Cof_Man_t) {
    var pObj *Cof_Obj_t;
    var i int32;
    // inlining of macro Cof_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Cof_ManObj(p,i)); i += Cof_ObjSize(pObj)  {
        pObj.Value = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Returns sorted array of node handles with largest fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManInsertEntry_rec(vNodes *Vec_Ptr_t, pNode *Cof_Obj_t, nNodeMax int32) {
    var pLast *Cof_Obj_t;
    if Vec_PtrSize(vNodes) == 0 {
        Vec_PtrPush(vNodes, pNode);
        return;
    }
    pLast = (Cof_Obj_t *)Vec_PtrPop(vNodes);
    if Cof_ObjFanoutNum(pLast) < Cof_ObjFanoutNum(pNode) {
        Cof_ManInsertEntry_rec( vNodes, pNode, nNodeMax );
        if Vec_PtrSize(vNodes) < nNodeMax {
            Vec_PtrPush( vNodes, pLast ); }
    
    } else {
        Vec_PtrPush( vNodes, pLast );
        if Vec_PtrSize(vNodes) < nNodeMax {
            Vec_PtrPush( vNodes, pNode ); }
    }
}

/**Function*************************************************************

  Synopsis    [Returns sorted array of node handles with largest fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManCollectHighFanout(p *Cof_Man_t, nNodes int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Cof_Obj_t;
    var i int32;
    vNodes = Vec_PtrAlloc( nNodes );
    // inlining of macro Cof_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Cof_ManObj(p,i)); i += Cof_ObjSize(pObj)  {
        if (Cof_ObjIsCi(pObj) != 0) || (Cof_ObjIsNode(pObj) != 0) {
            Cof_ManInsertEntry_rec( vNodes, pObj, nNodes ); }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns sorted array of node handles with largest fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManCountRemoved(p *Cof_Man_t, pRoot *Cof_Obj_t, fConst1 int32) int32 {
    var pNextGia *Gia_Obj_t;
    var pTemp, pNext, pFanin0, pFanin1 *Cof_Obj_t;
    var Counter, LevelStart, LevelNext int32;
    var i, k, iHandle, iLit0, iLit1, iNextNew int32;
    // restart the trav ids
    Cof_ManIncrementTravId( p );
    Cof_ObjSetTravIdCurrent( p, pRoot );
    // add the node to the queue
    LevelStart = Cof_ObjLevel(p, pRoot);
    assert(p.pLevels[LevelStart] == 0, "p.pLevels[LevelStart] == 0");
    pRoot.iNext = 0;
    p.pLevels[LevelStart] = Cof_ObjHandle( p, pRoot );
    // set the new literal
    pRoot.iLit = Abc_Var2Lit( 0, fConst1 );
    // process nodes in the levelized order
    for  i = LevelStart; i < p.nLevels; i++  {
        for  iHandle = p.pLevels[i]; 
              iHandle && (pTemp = Cof_ManObj(p, iHandle)); 
              iHandle = pTemp.iNext  {
            assert(pTemp.Id != Abc_Lit2Var(pTemp.iLit), "pTemp.Id != Abc_Lit2Var(pTemp.iLit)");
            // inlining of macro Cof_ObjForEachFanout
            for  k = 0; (k < int32(pTemp.nFanouts)) && (pNext = Cof_ObjFanout(pTemp,k)); k++  {
                if Cof_ObjIsCo(pNext) != 0 {
                    continue; }
                if Cof_ObjIsTravIdCurrent(p, pNext) != 0 {
                    continue; }
                pFanin0 = Cof_ObjFanin( pNext, 0 );
                pFanin1 = Cof_ObjFanin( pNext, 1 );
                assert(pFanin0 == pTemp || pFanin1 == pTemp, "pFanin0 == pTemp || pFanin1 == pTemp");
                pNextGia = Gia_ManObj( p.pGia, pNext.Id );
                if Cof_ObjIsTravIdCurrent(p, pFanin0) != 0 {
                    iLit0 = Abc_LitNotCond( pFanin0.iLit, Gia_ObjFaninC0(pNextGia) );
                } else {
                    iLit0 = Gia_ObjFaninLit0( pNextGia, pNext.Id ); }
                if Cof_ObjIsTravIdCurrent(p, pFanin1) != 0 {
                    iLit1 = Abc_LitNotCond( pFanin1.iLit, Gia_ObjFaninC1(pNextGia) );
                } else {
                    iLit1 = Gia_ObjFaninLit1( pNextGia, pNext.Id ); }
                iNextNew = Gia_ManHashAndTry( p.pGia, iLit0, iLit1 );
                if iNextNew == -1 {
                    continue; }
                Cof_ObjSetTravIdCurrent(p, pNext);
                // set the new literal
                pNext.iLit = iNextNew;
                // add it to be processed
                LevelNext = Cof_ObjLevel( p, pNext );
                assert(LevelNext > i && LevelNext < p.nLevels, "LevelNext > i && LevelNext < p.nLevels");
                pNext.iNext = p.pLevels[LevelNext];
                p.pLevels[LevelNext] = Cof_ObjHandle( p, pNext );
                Counter++;
            }
        }
        p.pLevels[i] = 0;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns sorted array of node handles with largest fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManPrintHighFanoutOne(p *Cof_Man_t, pObj *Cof_Obj_t) {
    fmt.Printf("%7d : ", pObj.Id);
    fmt.Printf("i/o/c =%2d %5d %5d  ", Cof_ObjFaninNum(pObj), Cof_ObjFanoutNum(pObj), 2*pObj.nFanoutsM);
    fmt.Printf("l =%4d  ", Cof_ObjLevel(p, pObj));
    fmt.Printf("s =%5d  ", Cof_ManSuppSize(p, &pObj, 1));
    fmt.Printf("TFI =%7d  ", Cof_ManTfiSize(p, &pObj, 1));
    fmt.Printf("TFO =%7d  ", Cof_ManTfoSize(p, &pObj, 1));
    fmt.Printf("C0 =%6d  ", Cof_ManCountRemoved(p, pObj, 0));
    fmt.Printf("C1 =%6d", Cof_ManCountRemoved(p, pObj, 1));
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Returns sorted array of node handles with largest fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManPrintHighFanout(p *Cof_Man_t, nNodes int32) {
    var vNodes *Vec_Ptr_t;
    var pObj *Cof_Obj_t;
    var i int32;
    vNodes = Cof_ManCollectHighFanout( p, nNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Cof_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Cof_ManPrintHighFanoutOne( p, pObj );
    }
    Vec_PtrFree( vNodes );
}


/**Function*************************************************************

  Synopsis    [Compute MFFC size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_NodeDeref_rec(pNode *Cof_Obj_t) int32 {
    if pNode.nFanins == 0 {
        return 0; }
    if --pNode.nFanouts > 0 {
        return 0; }
    return 1 + Cof_NodeDeref_rec( Cof_ObjFanin(pNode, 0) )
             + Cof_NodeDeref_rec( Cof_ObjFanin(pNode, 1) );
}
func Cof_NodeRef_rec(pNode *Cof_Obj_t) int32 {
    if pNode.nFanins == 0 {
        return 0; }
    if pNode.nFanouts++ > 0 {
        return 0; }
    return 1 + Cof_NodeRef_rec( Cof_ObjFanin(pNode, 0) )
             + Cof_NodeRef_rec( Cof_ObjFanin(pNode, 1) );
}
func Cof_ObjMffcSize(pNode *Cof_Obj_t) int32 {
    var Count1, Count2, nFanout int32;
    nFanout = pNode.nFanouts;
    pNode.nFanouts = 1;
    Count1 = Cof_NodeDeref_rec( pNode );
    Count2 = Cof_NodeRef_rec( pNode );
    pNode.nFanouts = nFanout;
    assert(Count1 == Count2, "Count1 == Count2");
    return Count1;
}

/**Function*************************************************************

  Synopsis    [Prints the distribution of fanins/fanouts in the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cof_ManPrintFanio(p *Cof_Man_t) {
    var Buffer string // was [100]byte in C

    var pNode *Cof_Obj_t;
    var vFanins, vFanouts, vMffcs *Vec_Int_t;
    var nFanins, nFanouts, nMffcs, nFaninsMax, nFanoutsMax, nMffcsMax, nFaninsAll, nFanoutsAll, nMffcsAll int32;
    var i, k, nSizeMax, nMffcNodes int32;

    // determine the largest fanin and fanout
    nFaninsMax = nFanoutsMax = nMffcsMax = 0;
    nFaninsAll = nFanoutsAll = nMffcsAll = 0;
    // inlining of macro Cof_ManForEachNode
    for  i = 0; (i < p.nObjData) && (pNode = Cof_ManObj(p,i)); i += Cof_ObjSize(pNode)  { if Cof_ObjIsTerm(pNode) != 0 {} else {
        if i == 0 { continue; }
        nFanins  = Cof_ObjFaninNum(pNode);
        nFanouts = Cof_ObjFanoutNum(pNode);
        if pNode.nFanouts > 1 {
            nMffcs = Cof_ObjMffcSize(pNode);
        } else {
            nMffcs = 0;
        }
        nFaninsAll  += nFanins;
        nFanoutsAll += nFanouts;
        nMffcsAll   += nMffcs;
        nFaninsMax   = Abc_MaxInt( nFaninsMax,  nFanins );
        nFanoutsMax  = Abc_MaxInt( nFanoutsMax, nFanouts );
        nMffcsMax    = Abc_MaxInt( nMffcsMax,   nMffcs );
    }
    }

    // allocate storage for fanin/fanout numbers
    nSizeMax = Abc_MaxInt( 10 * (Abc_Base10Log(nFaninsMax) + 1), 10 * (Abc_Base10Log(nFanoutsMax) + 1) );
    nSizeMax = Abc_MaxInt( 10 * (Abc_Base10Log(nMffcsMax) + 1),  nSizeMax );
    vFanins  = Vec_IntStart( nSizeMax );
    vFanouts = Vec_IntStart( nSizeMax );
    vMffcs   = Vec_IntStart( nSizeMax );

    // count the number of fanins and fanouts
    // inlining of macro Cof_ManForEachNode
    for  i = 0; (i < p.nObjData) && (pNode = Cof_ManObj(p,i)); i += Cof_ObjSize(pNode)  { if Cof_ObjIsTerm(pNode) != 0 {} else {
        if i == 0 { continue; }
        nFanins  = Cof_ObjFaninNum(pNode);
        nFanouts = Cof_ObjFanoutNum(pNode);
        if pNode.nFanouts > 1 {
            nMffcs = Cof_ObjMffcSize(pNode);
        } else {
            nMffcs = 0;
        }

        if nFanins < 10 {
            Vec_IntAddToEntry( vFanins, nFanins, 1 );
        } else if nFanins < 100 {
            Vec_IntAddToEntry( vFanins, 10 + nFanins/10, 1 );
        } else if nFanins < 1000 {
            Vec_IntAddToEntry( vFanins, 20 + nFanins/100, 1 );
        } else if nFanins < 10000 {
            Vec_IntAddToEntry( vFanins, 30 + nFanins/1000, 1 );
        } else if nFanins < 100000 {
            Vec_IntAddToEntry( vFanins, 40 + nFanins/10000, 1 );
        } else if nFanins < 1000000 {
            Vec_IntAddToEntry( vFanins, 50 + nFanins/100000, 1 );
        } else if nFanins < 10000000 {
            Vec_IntAddToEntry( vFanins, 60 + nFanins/1000000, 1 ); }

        if nFanouts < 10 {
            Vec_IntAddToEntry( vFanouts, nFanouts, 1 );
        } else if nFanouts < 100 {
            Vec_IntAddToEntry( vFanouts, 10 + nFanouts/10, 1 );
        } else if nFanouts < 1000 {
            Vec_IntAddToEntry( vFanouts, 20 + nFanouts/100, 1 );
        } else if nFanouts < 10000 {
            Vec_IntAddToEntry( vFanouts, 30 + nFanouts/1000, 1 );
        } else if nFanouts < 100000 {
            Vec_IntAddToEntry( vFanouts, 40 + nFanouts/10000, 1 );
        } else if nFanouts < 1000000 {
            Vec_IntAddToEntry( vFanouts, 50 + nFanouts/100000, 1 );
        } else if nFanouts < 10000000 {
            Vec_IntAddToEntry( vFanouts, 60 + nFanouts/1000000, 1 ); }
       
        if nMffcs == 0 {
            continue; }
        nMffcNodes++;

        if nMffcs < 10 {
            Vec_IntAddToEntry( vMffcs, nMffcs, 1 );
        } else if nMffcs < 100 {
            Vec_IntAddToEntry( vMffcs, 10 + nMffcs/10, 1 );
        } else if nMffcs < 1000 {
            Vec_IntAddToEntry( vMffcs, 20 + nMffcs/100, 1 );
        } else if nMffcs < 10000 {
            Vec_IntAddToEntry( vMffcs, 30 + nMffcs/1000, 1 );
        } else if nMffcs < 100000 {
            Vec_IntAddToEntry( vMffcs, 40 + nMffcs/10000, 1 );
        } else if nMffcs < 1000000 {
            Vec_IntAddToEntry( vMffcs, 50 + nMffcs/100000, 1 );
        } else if nMffcs < 10000000 {
            Vec_IntAddToEntry( vMffcs, 60 + nMffcs/1000000, 1 ); }
    }
    }

    fmt.Printf("The distribution of fanins, fanouts. and MFFCs in the network:\n");
    fmt.Printf("         Number    Nodes with fanin   Nodes with fanout   Nodes with MFFC\n");

    for  k = 0; k < nSizeMax; k++  {
        if vFanins.pArray[k] == 0 && vFanouts.pArray[k] == 0 && vMffcs.pArray[k] == 0 {
            continue; }
        if k < 10 {
            fmt.Printf("%15d : ", k);
        } else {
            Buffer = fmt.Sprintf("%d - %d", int32(math.Pow(float64(float64(10)), float64(k/10))) * (k%10), int32(math.Pow(float64(float64(10)), float64(k/10))) * (k%10+1) - 1); 
            printf( "%15s : ", Buffer );
        }
        if vFanins.pArray[k] == 0 {
            fmt.Printf("              ");
        } else {
            fmt.Printf("%11d   ", vFanins.pArray[k]); }
        fmt.Printf("    ");
        if vFanouts.pArray[k] == 0 {
            fmt.Printf("              ");
        } else {
            fmt.Printf("%12d  ", vFanouts.pArray[k]); }
        fmt.Printf("    ");
        if vMffcs.pArray[k] == 0 {
            fmt.Printf("               ");
        } else {
            fmt.Printf("  %12d  ", vMffcs.pArray[k]); }
        fmt.Printf("\n");
    }
    Vec_IntFree( vFanins );
    Vec_IntFree( vFanouts );
    Vec_IntFree( vMffcs );

    fmt.Printf("Fanins: Max = %d. Ave = %.2f.  Fanouts: Max = %d. Ave =  %.2f.  MFFCs: Max = %d. Ave =  %.2f.\n", nFaninsMax, 1.0*nFaninsAll /Cof_ManNodeNum(p), nFanoutsMax, 1.0*nFanoutsAll/Cof_ManNodeNum(p), nMffcsMax, 1.0*nMffcsAll  /nMffcNodes);
}

/**Function*************************************************************

  Synopsis    [Returns sorted array of node handles with largest fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintFanio(pGia *Gia_Man_t, nNodes int32) {
    var p *Cof_Man_t;
    var clk abctime  = Abc_Clock();
    p = Cof_ManCreateLogicSimple( pGia );
    p.nLevels = 1 + Gia_ManLevelNum( pGia );
    // inlining of macro ABC_CALLOC
    p.pLevels = ((int *) calloc((size_t)(p.nLevels), sizeof(int)));
    Cof_ManPrintFanio( p );

    if nNodes > 0 {
    Cof_ManResetTravId( p );
    Gia_ManHashStart( pGia );
    Cof_ManPrintHighFanout( p, nNodes );
    Gia_ManHashStop( pGia );
// inlining of macro ABC_PRMn
Abc_Print(1, "%s =", ("Memory for logic network"));
Abc_Print(1, "%10.3f MB  ",    1.0*(float64((4*p.nObjData)))/(1<<20));;
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }

    Cof_ManStop( p );
}


/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order while putting CIs first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupCofInt(p *Gia_Man_t, iVar int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj, pPivot *Gia_Obj_t;
    var i int32;
    var iCofVar int32 = -1;
    if !(iVar > 0 && iVar < Gia_ManObjNum(p)) {
        fmt.Printf("Gia_ManDupCof(): Variable %d is out of range (%d; %d).\n", iVar, 0, Gia_ManObjNum(p));
        return nil;
    }
    // find the cofactoring variable
    pPivot = Gia_ManObj( p, iVar );
    if !Gia_ObjIsCand(pPivot) {
        fmt.Printf("Gia_ManDupCof(): Variable %d should be a CI or an AND node.\n", iVar);
        return nil;
    }
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // compute negative cofactor
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
        if pObj == pPivot {
            iCofVar = pObj.Value;
            pObj.Value = Abc_Var2Lit( 0, 0 );
        }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        if pObj == pPivot {
            iCofVar = pObj.Value;
            pObj.Value = Abc_Var2Lit( 0, 0 );
        }
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }
    // compute the positive cofactor
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Abc_Var2Lit( Gia_ObjId(pNew, Gia_ManCi(pNew, i)), 0 );
        if pObj == pPivot {
            pObj.Value = Abc_Var2Lit( 0, 1 ); }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        if pObj == pPivot {
            pObj.Value = Abc_Var2Lit( 0, 1 ); }
    }
    }
    // create MUXes
    assert(iCofVar > 0, "iCofVar > 0");
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if pObj.Value == uint32(Gia_ObjFanin0Copy(pObj)) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ManHashMux(pNew, iCofVar, Gia_ObjFanin0Copy(pObj), pObj.Value) ); }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order while putting CIs first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupCof(p *Gia_Man_t, iVar int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    pNew = Gia_ManDupCofInt( p, iVar );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Determines variables whose fanout count is higher than this.]

  Description [Variables are returned in a reverse topological order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofVars(p *Gia_Man_t, nFanLim int32) *Vec_Int_t {
    var vVars *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro ABC_FREE
    p.Refs = nil;
    Gia_ManCreateRefs( p );
    vVars = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCand(pObj) && Gia_ObjRefNum(p, pObj) >= nFanLim {
            Vec_IntPush( vVars, i ); }
    }
    // inlining of macro ABC_FREE
    p.Refs = nil;
    return vVars;
}

/**Function*************************************************************

  Synopsis    [Transfers attributes from the original one to the final one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTransfer(pAig *Gia_Man_t, pCof *Gia_Man_t, pNew *Gia_Man_t, vSigs *Vec_Int_t) *Vec_Int_t {
    var vSigsNew *Vec_Int_t;
    var pObj, pObjF *Gia_Obj_t;
    var i int32;
    vSigsNew = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vSigs)) && ((pObj) = Gia_ManObj(pAig, Vec_IntEntry(vSigs,i))); i++  {
        assert(Gia_ObjIsCand(pObj), "Gia_ObjIsCand(pObj)");
        pObjF = Gia_ManObj( pCof, Abc_Lit2Var(pObj.Value) );
        if pObjF.Value && ~pObjF.Value {
            Vec_IntPushUnique( vSigsNew, Abc_Lit2Var(pObjF.Value) ); }
    }
    return vSigsNew;    
}

/**Function*************************************************************

  Synopsis    [Cofactors selected variables (should be in reverse topo order).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupCofAllInt(p *Gia_Man_t, vSigs *Vec_Int_t, fVerbose int32) *Gia_Man_t {
    var vSigsNew, vTemp *Vec_Int_t;
    var pAig, pCof, pNew *Gia_Man_t;
    var iVar int32;
    if fVerbose != 0 { 
        fmt.Printf("Cofactoring %d signals.\n", Vec_IntSize(vSigs));
        Gia_ManPrintStats( p, nil );
    }
    if Vec_IntSize( vSigs ) > 200 {
        fmt.Printf("Too many signals to cofactor.\n");
        return nil;
    }
    pAig = Gia_ManDup( p );
    vSigsNew = Vec_IntDup( vSigs );
    for  Vec_IntSize(vSigsNew) > 0  {
        Vec_IntSort( vSigsNew, 0 );
        iVar = Vec_IntPop( vSigsNew );
//        Gia_ManCreateRefs( pAig );
//        printf( "ref count = %d\n", Gia_ObjRefNum( pAig, Gia_ManObj(pAig, iVar) ) );
//        ABC_FREE( pAig->pRefs );
        pCof = Gia_ManDupCofInt( pAig, iVar );
        pNew = Gia_ManCleanup( pCof );
        vSigsNew = Gia_ManTransfer( pAig, pCof, pNew, vTemp = vSigsNew );
        Vec_IntFree( vTemp );
        Gia_ManStop( pAig );
        Gia_ManStop( pCof );
        pAig = pNew;
        if fVerbose != 0 {
            fmt.Printf("Cofactored variable %d.\n", iVar); }
        if fVerbose != 0 {
            Gia_ManPrintStats( pAig, nil ); }
    }
    Vec_IntFree( vSigsNew );
    return pAig;
}

/**Function*************************************************************

  Synopsis    [Cofactors all variables whose fanout is higher than this.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupCofAll(p *Gia_Man_t, nFanLim int32, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vSigs *Vec_Int_t  = Gia_ManCofVars( p, nFanLim );
    pNew = Gia_ManDupCofAllInt( p, vSigs, fVerbose );
    Vec_IntFree( vSigs );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

