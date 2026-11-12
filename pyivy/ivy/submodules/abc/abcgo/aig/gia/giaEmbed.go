package gia

import (
	"fmt"
	"math"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [giaEmbed.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Logic network derived from AIG.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaEmbed.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "gia.h"
//#include "aig/ioa/ioa.h"

//ABC_NAMESPACE_IMPL_START


/* 
    The code is based on the paper by D. Harel and Y. Koren, 
    "Graph drawing by high-dimensional embedding", 
    J. Graph Algs & Apps, Vol 8(2), pp. 195-217 (2004).
    http://www.emis.de/journals/JGAA/accepted/2004/HarelKoren2004.8.2.pdf 

    Iterative refinement is described in the paper: F. A. Aloul, I. L. Markov, and K. A. Sakallah.
    "FORCE: A Fast and Easy-To-Implement Variable-Ordering Heuristic", Proc. GLSVLSI_03.
    http://www.eecs.umich.edu/~imarkov/pubs/conf/glsvlsi03-force.pdf
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const GIA_PLACE_SIZE = 0x7fff 
// objects will be placed in box [0, GIA_PLACE_SIZE] x [0, GIA_PLACE_SIZE]

typedef float  Emb_Dat_t;

   
type Emb_Obj_t struct {
    fCi uint32; /* C bitfield: 1 bits; review packing and truncation */    // terminal node CI
    fCo uint32; /* C bitfield: 1 bits; review packing and truncation */    // terminal node CO
    fMark0 uint32; /* C bitfield: 1 bits; review packing and truncation */    // first user-controlled mark
    fMark1 uint32; /* C bitfield: 1 bits; review packing and truncation */    // second user-controlled mark
    nFanins uint32; /* C bitfield: 28 bits; review packing and truncation */    // the number of fanins
    nFanouts       uint32;         // the number of fanouts
    hHandle            int32;          // the handle of the node
    union {
    unsigned       TravId;           // user-specified value
    unsigned       iFanin; 
    };
    union {
    unsigned       Value;            // user-specified value
    unsigned       iFanout; 
    };
    Fanios [0]int32;        // the array of fanins/fanouts
}

   
type Emb_Man_t struct {
    pGia     *Gia_Man_t;             // the original AIG manager
    vCis     *Vec_Int_t;             // the vector of CIs (PIs + LOs)
    vCos     *Vec_Int_t;             // the vector of COs (POs + LIs)
    nObjs            int32;            // the number of objects
    nRegs            int32;            // the number of registers
    nTravIds            int32;         // traversal ID of the network
    pObjData           *int32;         // the array containing data for objects
    nObjData            int32;         // the size of array to store the logic network
    fVerbose            int32;         // verbose output flag
    pVecs     *Emb_Dat_t;            // array of vectors of size nObjs * nDims
    nReached            int32;         // the number of nodes reachable from the pivot
    nDistMax            int32;         // the maximum distance from the node
    pMatr        **float32;            // covariance matrix nDims * nDims
    pEigen        **float32;           // the first several eigen values of the matrix
    pSols         *float32;            // solutions to the problem nObjs * nSols;
    unsigned short*pPlacement;       // (x,y) coordinates for each cell
}

func Emb_ManRegNum(p *Emb_Man_t) int32 { return p.nRegs;                                    }
func Emb_ManCiNum(p *Emb_Man_t) int32 { return Vec_IntSize(p.vCis);                        }
func Emb_ManCoNum(p *Emb_Man_t) int32 { return Vec_IntSize(p.vCos);                        }
func Emb_ManPiNum(p *Emb_Man_t) int32 { return Vec_IntSize(p.vCis) - p.nRegs;             }
func Emb_ManPoNum(p *Emb_Man_t) int32 { return Vec_IntSize(p.vCos) - p.nRegs;             }
func Emb_ManObjNum(p *Emb_Man_t) int32 { return p.nObjs;                                    } 
func Emb_ManNodeNum(p *Emb_Man_t) int32 { return p.nObjs - Vec_IntSize(p.vCis) - Vec_IntSize(p.vCos); } 

func Emb_ManObj(p *Emb_Man_t, hHandle uint32) *Emb_Obj_t { return (Emb_Obj_t *)(p.pObjData + hHandle);        } 
func Emb_ManCi(p *Emb_Man_t, i int32) *Emb_Obj_t { return Emb_ManObj( p, Vec_IntEntry(p.vCis,i) );    }
func Emb_ManCo(p *Emb_Man_t, i int32) *Emb_Obj_t { return Emb_ManObj( p, Vec_IntEntry(p.vCos,i) );    }

func Emb_ObjIsTerm(pObj *Emb_Obj_t) int32 { return pObj.fCi || pObj.fCo;                      } 
func Emb_ObjIsCi(pObj *Emb_Obj_t) int32 { return pObj.fCi;                                   } 
func Emb_ObjIsCo(pObj *Emb_Obj_t) int32 { return pObj.fCo;                                   } 
//static inline int         Emb_ObjIsPi( Emb_Obj_t * pObj )                             { return pObj->fCi && pObj->nFanins == 0;             } 
//static inline int         Emb_ObjIsPo( Emb_Obj_t * pObj )                             { return pObj->fCo && pObj->nFanouts == 0;            } 
func Emb_ObjIsNode(pObj *Emb_Obj_t) int32 { return!Emb_ObjIsTerm(pObj) && pObj.nFanins > 0;    } 
//static inline int         Emb_ObjIsConst0( Emb_Obj_t * pObj )                         { return!Emb_ObjIsTerm(pObj) && pObj->nFanins == 0;   } 

func Emb_ObjSize(pObj *Emb_Obj_t) int32 { return sizeof(Emb_Obj_t) / 4 + pObj.nFanins + pObj.nFanouts;  } 
func Emb_ObjFaninNum(pObj *Emb_Obj_t) int32 { return pObj.nFanins;                               } 
func Emb_ObjFanoutNum(pObj *Emb_Obj_t) int32 { return pObj.nFanouts;                              } 
func Emb_ObjFanin(pObj *Emb_Obj_t, i int32) *Emb_Obj_t { return (Emb_Obj_t *)(((int *)pObj) - pObj.Fanios[i]);               } 
func Emb_ObjFanout(pObj *Emb_Obj_t, i int32) *Emb_Obj_t { return (Emb_Obj_t *)(((int *)pObj) + pObj.Fanios[pObj.nFanins+i]); } 

func Emb_ManResetTravId(p *Emb_Man_t) { extern void Emb_ManCleanTravId( Emb_Man_t * p ); Emb_ManCleanTravId( p ); p.nTravIds = 1;  }
func Emb_ManIncrementTravId(p *Emb_Man_t) { p.nTravIds++;                                      }
func Emb_ObjSetTravId(pObj *Emb_Obj_t, TravId int32) { pObj.TravId = TravId;                              }
func Emb_ObjSetTravIdCurrent(p *Emb_Man_t, pObj *Emb_Obj_t) { pObj.TravId = p.nTravIds;                         }
func Emb_ObjSetTravIdPrevious(p *Emb_Man_t, pObj *Emb_Obj_t) { pObj.TravId = p.nTravIds - 1;                     }
func Emb_ObjIsTravIdCurrent(p *Emb_Man_t, pObj *Emb_Obj_t) int32 { return (int32(pObj.TravId) == p.nTravIds);          }
func Emb_ObjIsTravIdPrevious(p *Emb_Man_t, pObj *Emb_Obj_t) int32 { return (int32(pObj.TravId) == p.nTravIds - 1);      }

func Emb_ManVec(p *Emb_Man_t, v int32) *Emb_Dat_t { return p.pVecs + v * p.nObjs;                     }
func Emb_ManSol(p *Emb_Man_t, v int32) *float32 { return p.pSols + v * p.nObjs;                     }

// Original macro:
// #define Emb_ManForEachObj( p, pObj, i )               \
//     for ( i = 0; (i < p->nObjData) && (pObj = Emb_ManObj(p,i)); i += Emb_ObjSize(pObj) )

// Original macro:
// #define Emb_ManForEachNode( p, pObj, i )              \
//     for ( i = 0; (i < p->nObjData) && (pObj = Emb_ManObj(p,i)); i += Emb_ObjSize(pObj) ) if ( Emb_ObjIsTerm(pObj) ) {} else

// Original macro:
// #define Emb_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Emb_ManObj(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Emb_ObjForEachFanin( pObj, pNext, i )         \
//     for ( i = 0; (i < (int)pObj->nFanins) && (pNext = Emb_ObjFanin(pObj,i)); i++ )

// Original macro:
// #define Emb_ObjForEachFanout( pObj, pNext, i )        \
//     for ( i = 0; (i < (int)pObj->nFanouts) && (pNext = Emb_ObjFanout(pObj,i)); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates fanin/fanout pair.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ObjAddFanin(pObj *Emb_Obj_t, pFanin *Emb_Obj_t) { 
    assert(pObj.iFanin < pObj.nFanins, "pObj.iFanin < pObj.nFanins");
    assert(pFanin.iFanout < pFanin.nFanouts, "pFanin.iFanout < pFanin.nFanouts");
    pFanin.Fanios[pFanin.nFanins + pFanin.iFanout++] = 
    pObj.Fanios[pObj.iFanin++] = pObj.hHandle - pFanin.hHandle;
}

/**Function*************************************************************

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManStartSimple(pGia *Gia_Man_t) *Emb_Man_t {
    
    var pObjLog, pFanLog *Emb_Obj_t;
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var i, nNodes, hHandle int32;
    // prepare the AIG
    Gia_ManCreateRefs( pGia );
    // create logic network
    // inlining of macro ABC_CALLOC
    p := &Emb_Man_t{};
    p.pGia  = pGia;
    p.nRegs = Gia_ManRegNum(pGia);
    p.vCis  = Vec_IntAlloc( Gia_ManCiNum(pGia) );
    p.vCos  = Vec_IntAlloc( Gia_ManCoNum(pGia) );
    p.nObjData = (sizeof(Emb_Obj_t) / 4) * Gia_ManObjNum(pGia) + 2 * (2 * Gia_ManAndNum(pGia) + Gia_ManCoNum(pGia) + Gia_ManRegNum(pGia) + Gia_ManCoNum(pGia));
    // inlining of macro ABC_CALLOC
    p.pObjData = ((int *) calloc((size_t)(p.nObjData), sizeof(int)));
    // create constant node
    Gia_ManConst0(pGia).Value = hHandle;
    pObjLog = Emb_ManObj( p, hHandle );
    pObjLog.hHandle  = hHandle;
    pObjLog.nFanins  = Gia_ManCoNum(pGia);  //0;
    pObjLog.nFanouts = Gia_ObjRefNum( pGia, Gia_ManConst0(pGia) );
    // count objects
    hHandle += Emb_ObjSize( pObjLog );
    nNodes = 1;
    p.nObjs++;
    // create the PIs
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
        // create PI object
        pObj.Value = hHandle;
        Vec_IntPush( p.vCis, hHandle );
        pObjLog = Emb_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = Gia_ObjIsRo( pGia, pObj );
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        pObjLog.fCi = 1;
        // count objects
        hHandle += Emb_ObjSize( pObjLog );
        p.nObjs++;
    }
    // create internal nodes
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(Gia_ObjRefNum( pGia, pObj ) > 0, "Gia_ObjRefNum( pGia, pObj ) > 0");
        // create node object
        pObj.Value = hHandle;
        pObjLog = Emb_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = 2;
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        // add fanins
        pFanLog = Emb_ManObj( p, Gia_ObjValue(Gia_ObjFanin0(pObj)) ); 
        Emb_ObjAddFanin( pObjLog, pFanLog );
        pFanLog = Emb_ManObj( p, Gia_ObjValue(Gia_ObjFanin1(pObj)) ); 
        Emb_ObjAddFanin( pObjLog, pFanLog );
        // count objects
        hHandle += Emb_ObjSize( pObjLog );
        nNodes++;
        p.nObjs++;
    }
    }
    // create the POs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
        // create PO object
        pObj.Value = hHandle;
        Vec_IntPush( p.vCos, hHandle );
        pObjLog = Emb_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = 1;
        pObjLog.nFanouts = 1 + Gia_ObjIsRi( pGia, pObj );
        pObjLog.fCo = 1;
        // add fanins
        pFanLog = Emb_ManObj( p, Gia_ObjValue(Gia_ObjFanin0(pObj)) );
        Emb_ObjAddFanin( pObjLog, pFanLog );
        // count objects
        hHandle += Emb_ObjSize( pObjLog );
        p.nObjs++;
    }
    // connect registers
    // inlining of macro Gia_ManForEachRiRo
    for  i = 0; (i < Gia_ManRegNum(pGia)) && ((pObjRi) = Gia_ManCo(pGia, Gia_ManPoNum(pGia)+i)) && ((pObjRo) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+i)); i++  {
        Emb_ObjAddFanin( Emb_ManObj(p,Gia_ObjValue(pObjRo)), Emb_ManObj(p,Gia_ObjValue(pObjRi)) );
    }
    assert(nNodes  == Emb_ManNodeNum(p), "nNodes  == Emb_ManNodeNum(p)");
    assert(hHandle == p.nObjData, "hHandle == p.nObjData");
    assert(p.nObjs == Gia_ManObjNum(pGia), "p.nObjs == Gia_ManObjNum(pGia)");
    if hHandle != p.nObjData {
        fmt.Printf("Emb_ManStartSimple(): Fatal error in internal representation.\n"); }
    // make sure the fanin/fanout counters are correct
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        if !~Gia_ObjValue(pObj) {
            continue; }
        pObjLog = Emb_ManObj( p, Gia_ObjValue(pObj) );
        assert(pObjLog.nFanins  == pObjLog.iFanin || Gia_ObjIsConst0(pObj), "pObjLog.nFanins  == pObjLog.iFanin || Gia_ObjIsConst0(pObj)");
        assert(pObjLog.nFanouts == pObjLog.iFanout || Gia_ObjIsCo(pObj), "pObjLog.nFanouts == pObjLog.iFanout || Gia_ObjIsCo(pObj)");
        pObjLog.iFanin = pObjLog.iFanout = 0;
    }
    // inlining of macro ABC_FREE
    pGia.Refs = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    [Collect the fanin IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManCollectSuper_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t, vVisit *Vec_Int_t) {
    if pObj.fMark1 {
        return; }
    pObj.fMark1 = 1;
    Vec_IntPush( vVisit, Gia_ObjId(p, pObj) );
    if pObj.fMark0 {
        Vec_IntPush( vSuper, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Emb_ManCollectSuper_rec( p, Gia_ObjFanin0(pObj), vSuper, vVisit );
    Emb_ManCollectSuper_rec( p, Gia_ObjFanin1(pObj), vSuper, vVisit );
    
}

/**Function*************************************************************

  Synopsis    [Collect the fanin IDs.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Emb_ManCollectSuper(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t, vVisit *Vec_Int_t) {
    var Entry, i int32;
    Vec_IntClear( vSuper );
    Vec_IntClear( vVisit );
    assert(pObj.fMark0 == 1, "pObj.fMark0 == 1");
    pObj.fMark0 = 0;
    Emb_ManCollectSuper_rec( p, pObj, vSuper, vVisit );
    pObj.fMark0 = 1;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVisit)) && (((Entry) = Vec_IntEntry(vVisit, i)), 1); i++  {
        Gia_ManObj(p, Entry).fMark1 = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Assigns references while removing the MUX/XOR ones.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManCreateRefsSpecial(p *Gia_Man_t) {
    var pObj, pFan0, pFan1 *Gia_Obj_t;
    var pObjC, pObjD0, pObjD1 *Gia_Obj_t;
    var i int32;
    assert(p.Refs == nil, "p.Refs == nil");
    Gia_ManCleanMark0( p );
    Gia_ManCreateRefs( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
        pFan0 = Gia_ObjFanin0(pObj);
        pFan1 = Gia_ObjFanin1(pObj);
        // skip nodes whose fanins are PIs or are already marked
        if Gia_ObjIsCi(pFan0) || pFan0.fMark0 || 
             Gia_ObjIsCi(pFan1) || pFan1.fMark0 {
             continue; }
        // skip nodes that are not MUX type
        if !Gia_ObjIsMuxType(pObj) {
            continue; }
        // the node is MUX type, mark it and its fanins
        pObj.fMark0  = 1;
        pFan0.fMark0 = 1;
        pFan1.fMark0 = 1;
        // deref the control 
        pObjC = Gia_ObjRecognizeMux( pObj, &pObjD1, &pObjD0 );
        Gia_ObjRefDec( p, Gia_Regular(pObjC) );
        if Gia_Regular(pObjD0) == Gia_Regular(pObjD1) {
            Gia_ObjRefDec( p, Gia_Regular(pObjD0) ); }
    }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(Gia_ObjRefNum(p, pObj) > 0, "Gia_ObjRefNum(p, pObj) > 0"); }
    }
    Gia_ManCleanMark0( p );
}

/**Function*************************************************************

  Synopsis    [Assigns references while removing the MUX/XOR ones.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManTransformRefs(p *Gia_Man_t, pnObjs *int32, pnFanios *int32) {
    var vSuper, vVisit *Vec_Int_t;
    var pObj, pFanin *Gia_Obj_t;
    var i, k, Counter int32;
    assert(p.Refs != nil, "p.Refs != nil");

    // mark nodes to be used in the logic network
    Gia_ManCleanMark0( p );
    Gia_ManConst0(p).fMark0 = 1;
    // mark the inputs
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fMark0 = 1;
    }
    // mark those nodes that have ref count more than 1
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.fMark0 = (Gia_ObjRefNum(p, pObj) > 1); }
    }
    // mark the output drivers
    // inlining of macro Gia_ManForEachCoDriver
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ObjFanin0(Gia_ManCo(p, i))); i++  {
        pObj.fMark0 = 1;
    }

    // count the number of nodes
    Counter = 0;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Counter += pObj.fMark0;
    }
    *pnObjs = Counter + Gia_ManCoNum(p);

    // reset the references
    // inlining of macro ABC_FREE
    p.Refs = nil;
    // inlining of macro ABC_CALLOC
    p.Refs = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    // reference from internal nodes
    Counter = 0;
    vSuper = Vec_IntAlloc( 100 );
    vVisit = Vec_IntAlloc( 100 );
    Gia_ManCleanMark1( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if pObj.fMark0 == 0 {
            continue; }
        Emb_ManCollectSuper( p, pObj, vSuper, vVisit );
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vSuper)) && ((pFanin) = Gia_ManObj(p, Vec_IntEntry(vSuper,k))); k++  {
            assert(pFanin.fMark0, "pFanin.fMark0");
            Gia_ObjRefInc( p, pFanin );
        }
        Counter += Vec_IntSize( vSuper );
    }
    }
    Gia_ManCheckMark1( p );
    Vec_IntFree( vSuper );
    Vec_IntFree( vVisit );
    // reference from outputs
    // inlining of macro Gia_ManForEachCoDriver
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ObjFanin0(Gia_ManCo(p, i))); i++  {
        assert(pObj.fMark0, "pObj.fMark0");
        Gia_ObjRefInc( p, pObj );
    }
    *pnFanios = Counter + Gia_ManCoNum(p);
}

/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManCleanTravId(p *Emb_Man_t) {
    var pObj *Emb_Obj_t;
    var i int32;
    // inlining of macro Emb_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Emb_ManObj(p,i)); i += Emb_ObjSize(pObj)  {
        pObj.TravId = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Cleans the value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManSetValue(p *Emb_Man_t) {
    var pObj *Emb_Obj_t;
    var i, Counter int32;
    // inlining of macro Emb_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Emb_ManObj(p,i)); i += Emb_ObjSize(pObj)  {
        pObj.Value = Counter++;
//        if ( pObj->fCi && pObj->nFanins == 0 ) 
//            printf( "CI:  Handle = %8d.  Value = %6d. Fanins = %d.\n", pObj->hHandle, pObj->Value, pObj->nFanins );
    }
}

/**Function*************************************************************

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManStart(pGia *Gia_Man_t) *Emb_Man_t {
    
    var pObjLog, pFanLog *Emb_Obj_t;
    var pObj, pObjRi, pObjRo, pFanin *Gia_Obj_t;
    var vSuper, vVisit *Vec_Int_t;
    var nObjs, nFanios, nNodes int32;
    var i, k, hHandle int32;
    // prepare the AIG
//    Gia_ManCreateRefs( pGia );
    Emb_ManCreateRefsSpecial( pGia );
    Emb_ManTransformRefs( pGia, &nObjs, &nFanios );
    Gia_ManFillValue( pGia );
    // create logic network
    // inlining of macro ABC_CALLOC
    p := &Emb_Man_t{};
    p.pGia  = pGia;
    p.nRegs = Gia_ManRegNum(pGia);
    p.vCis  = Vec_IntAlloc( Gia_ManCiNum(pGia) );
    p.vCos  = Vec_IntAlloc( Gia_ManCoNum(pGia) );
    p.nObjData = (sizeof(Emb_Obj_t) / 4) * nObjs + 2 * (nFanios + Gia_ManRegNum(pGia) + Gia_ManCoNum(pGia));
    // inlining of macro ABC_CALLOC
    p.pObjData = ((int *) calloc((size_t)(p.nObjData), sizeof(int)));
    // create constant node
    Gia_ManConst0(pGia).Value = hHandle;
    pObjLog = Emb_ManObj( p, hHandle );
    pObjLog.hHandle  = hHandle;
    pObjLog.nFanins  = Gia_ManCoNum(pGia);  //0;
    pObjLog.nFanouts = Gia_ObjRefNum( pGia, Gia_ManConst0(pGia) );
    // count objects
    hHandle += Emb_ObjSize( pObjLog );
    nNodes++;
    p.nObjs++;
    // create the PIs
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
        // create PI object
        pObj.Value = hHandle;
        Vec_IntPush( p.vCis, hHandle );
        pObjLog = Emb_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = Gia_ObjIsRo( pGia, pObj );
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        pObjLog.fCi = 1;
        // count objects
        hHandle += Emb_ObjSize( pObjLog );
        p.nObjs++;
    }
    // create internal nodes
    vSuper = Vec_IntAlloc( 100 );
    vVisit = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if pObj.fMark0 == 0 {
            assert(Gia_ObjRefNum( pGia, pObj ) == 0, "Gia_ObjRefNum( pGia, pObj ) == 0");
            continue;
        }
        assert(Gia_ObjRefNum( pGia, pObj ) > 0, "Gia_ObjRefNum( pGia, pObj ) > 0");
        Emb_ManCollectSuper( pGia, pObj, vSuper, vVisit );
        // create node object
        pObj.Value = hHandle;
        pObjLog = Emb_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = Vec_IntSize( vSuper );
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        // add fanins
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vSuper)) && ((pFanin) = Gia_ManObj(pGia, Vec_IntEntry(vSuper,k))); k++  {
            pFanLog = Emb_ManObj( p, Gia_ObjValue(pFanin) ); 
            Emb_ObjAddFanin( pObjLog, pFanLog );
        }
        // count objects
        hHandle += Emb_ObjSize( pObjLog );
        nNodes++;
        p.nObjs++;
    }
    }
    Vec_IntFree( vSuper );
    Vec_IntFree( vVisit );
    // create the POs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
        // create PO object
        pObj.Value = hHandle;
        Vec_IntPush( p.vCos, hHandle );
        pObjLog = Emb_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = 1;
        pObjLog.nFanouts = 1 + Gia_ObjIsRi( pGia, pObj );
        pObjLog.fCo = 1;
        // add fanins
        pFanLog = Emb_ManObj( p, Gia_ObjValue(Gia_ObjFanin0(pObj)) );
        Emb_ObjAddFanin( pObjLog, pFanLog );
        // count objects
        hHandle += Emb_ObjSize( pObjLog );
        p.nObjs++;
    }
    // connect registers
    // inlining of macro Gia_ManForEachRiRo
    for  i = 0; (i < Gia_ManRegNum(pGia)) && ((pObjRi) = Gia_ManCo(pGia, Gia_ManPoNum(pGia)+i)) && ((pObjRo) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+i)); i++  {
        Emb_ObjAddFanin( Emb_ManObj(p,Gia_ObjValue(pObjRo)), Emb_ManObj(p,Gia_ObjValue(pObjRi)) );
    }
    Gia_ManCleanMark0( pGia );
    assert(nNodes  == Emb_ManNodeNum(p), "nNodes  == Emb_ManNodeNum(p)");
    assert(nObjs   == p.nObjs, "nObjs   == p.nObjs");
    assert(hHandle == p.nObjData, "hHandle == p.nObjData");
    if hHandle != p.nObjData {
        fmt.Printf("Emb_ManStart(): Fatal error in internal representation.\n"); }
    // make sure the fanin/fanout counters are correct
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        if !~Gia_ObjValue(pObj) {
            continue; }
        pObjLog = Emb_ManObj( p, Gia_ObjValue(pObj) );
        assert(pObjLog.nFanins  == pObjLog.iFanin || Gia_ObjIsConst0(pObj), "pObjLog.nFanins  == pObjLog.iFanin || Gia_ObjIsConst0(pObj)");
        assert(pObjLog.nFanouts == pObjLog.iFanout || Gia_ObjIsCo(pObj), "pObjLog.nFanouts == pObjLog.iFanout || Gia_ObjIsCo(pObj)");
        pObjLog.iFanin = pObjLog.iFanout = 0;
    }
    // inlining of macro ABC_FREE
    pGia.Refs = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManPrintStats(p *Emb_Man_t) {
//    if ( p->pName )
//        printf( "%8s : ", p->pName );
    fmt.Printf("i/o =%7d/%7d  ", Emb_ManPiNum(p), Emb_ManPoNum(p));
    if Emb_ManRegNum(p) != 0 {
        fmt.Printf("ff =%7d  ", Emb_ManRegNum(p)); }
    fmt.Printf("node =%8d  ", Emb_ManNodeNum(p));
    fmt.Printf("obj =%8d  ", Emb_ManObjNum(p));
//    printf( "lev =%5d  ", Emb_ManLevelNum(p) );
//    printf( "cut =%5d  ", Emb_ManCrossCut(p) );
    fmt.Printf("mem =%5.2f MB", 4.0*p.nObjData/(1<<20));
//    printf( "obj =%5d  ", Emb_ManObjNum(p) );
    fmt.Printf("\n");

//    Emb_ManSatExperiment( p );
}

/**Function*************************************************************

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManStop(p *Emb_Man_t) {
    Vec_IntFree( p.vCis );
    Vec_IntFree( p.vCos );
    // inlining of macro ABC_FREE
    p.pPlacement = nil;
    // inlining of macro ABC_FREE
    p.pVecs = nil;
    // inlining of macro ABC_FREE
    p.pSols = nil;
    // inlining of macro ABC_FREE
    p.pMatr = nil;
    // inlining of macro ABC_FREE
    p.pEigen = nil;
    // inlining of macro ABC_FREE
    p.pObjData = nil;
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Prints the distribution of fanins/fanouts in the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManPrintFanio(p *Emb_Man_t) {
    var Buffer string // was [100]byte in C

    var pNode *Emb_Obj_t;
    var vFanins, vFanouts *Vec_Int_t;
    var nFanins, nFanouts, nFaninsMax, nFanoutsMax, nFaninsAll, nFanoutsAll int32;
    var i, k, nSizeMax int32;

    // determine the largest fanin and fanout
    nFaninsMax = nFanoutsMax = 0;
    nFaninsAll = nFanoutsAll = 0;
    // inlining of macro Emb_ManForEachNode
    for  i = 0; (i < p.nObjData) && (pNode = Emb_ManObj(p,i)); i += Emb_ObjSize(pNode)  { if Emb_ObjIsTerm(pNode) != 0 {} else {
        if i == 0 { continue; } // skip const 0 obj
        nFanins  = Emb_ObjFaninNum(pNode);
        nFanouts = Emb_ObjFanoutNum(pNode);
        nFaninsAll  += nFanins;
        nFanoutsAll += nFanouts;
        nFaninsMax   = Abc_MaxInt( nFaninsMax, nFanins );
        nFanoutsMax  = Abc_MaxInt( nFanoutsMax, nFanouts );
    }
    }

    // allocate storage for fanin/fanout numbers
    nSizeMax = Abc_MaxInt( 10 * (Abc_Base10Log(nFaninsMax) + 1), 10 * (Abc_Base10Log(nFanoutsMax) + 1) );
    vFanins  = Vec_IntStart( nSizeMax );
    vFanouts = Vec_IntStart( nSizeMax );

    // count the number of fanins and fanouts
    // inlining of macro Emb_ManForEachNode
    for  i = 0; (i < p.nObjData) && (pNode = Emb_ManObj(p,i)); i += Emb_ObjSize(pNode)  { if Emb_ObjIsTerm(pNode) != 0 {} else {
        if i == 0 { continue; } // skip const 0 obj
        nFanins  = Emb_ObjFaninNum(pNode);
        nFanouts = Emb_ObjFanoutNum(pNode);

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
    }
    }

    fmt.Printf("The distribution of fanins and fanouts in the network:\n");
    fmt.Printf("         Number   Nodes with fanin  Nodes with fanout\n");
    for  k = 0; k < nSizeMax; k++  {
        if vFanins.pArray[k] == 0 && vFanouts.pArray[k] == 0 {
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
            fmt.Printf("%12d  ", vFanins.pArray[k]); }
        fmt.Printf("    ");
        if vFanouts.pArray[k] == 0 {
            fmt.Printf("              ");
        } else {
            fmt.Printf("%12d  ", vFanouts.pArray[k]); }
        fmt.Printf("\n");
    }
    Vec_IntFree( vFanins );
    Vec_IntFree( vFanouts );

    fmt.Printf("Fanins: Max = %d. Ave = %.2f.  Fanouts: Max = %d. Ave =  %.2f.\n", nFaninsMax, 1.0*nFaninsAll/Emb_ManNodeNum(p), nFanoutsMax, 1.0*nFanoutsAll/Emb_ManNodeNum(p));
}

/**Function*************************************************************

  Synopsis    [Computes the distance from the given object]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManComputeDistance_old(p *Emb_Man_t, pPivot *Emb_Obj_t) int32 {
    var vThis, vNext, vTemp *Vec_Int_t;
    var pThis, pNext *Emb_Obj_t;
    var i, k, d, nVisited int32;
//    assert( Emb_ObjIsTerm(pPivot) );
    vThis = Vec_IntAlloc( 1000 );
    vNext = Vec_IntAlloc( 1000 );
    Emb_ManIncrementTravId( p );
    Emb_ObjSetTravIdCurrent( p, pPivot );
    Vec_IntPush( vThis, pPivot.hHandle );
    for  d = 0; Vec_IntSize(vThis) > 0; d++  {
        nVisited += Vec_IntSize(vThis);
        Vec_IntClear( vNext );
        // inlining of macro Emb_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vThis)) && ((pThis) = Emb_ManObj(p, Vec_IntEntry(vThis,i))); i++  {
            // inlining of macro Emb_ObjForEachFanin
            for  k = 0; (k < int32(pThis.nFanins)) && (pNext = Emb_ObjFanin(pThis,k)); k++  {
                if Emb_ObjIsTravIdCurrent(p, pNext) != 0 {
                    continue; }
                Emb_ObjSetTravIdCurrent(p, pNext);
                Vec_IntPush( vNext, pNext.hHandle );
                nVisited += !Emb_ObjIsTerm(pNext);
            }
            // inlining of macro Emb_ObjForEachFanout
            for  k = 0; (k < int32(pThis.nFanouts)) && (pNext = Emb_ObjFanout(pThis,k)); k++  {
                if Emb_ObjIsTravIdCurrent(p, pNext) != 0 {
                    continue; }
                Emb_ObjSetTravIdCurrent(p, pNext);
                Vec_IntPush( vNext, pNext.hHandle );
                nVisited += !Emb_ObjIsTerm(pNext);
            }
        }
        vTemp = vThis; vThis = vNext; vNext = vTemp;
    }
    Vec_IntFree( vThis );
    Vec_IntFree( vNext );
    // check if there are several strongly connected components
//    if ( nVisited < Emb_ManNodeNum(p) )
//        printf( "Visited less nodes (%d) than present (%d).\n", nVisited, Emb_ManNodeNum(p) );
    return d;
}

/**Function*************************************************************

  Synopsis    [Traverses from the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Gia_ManTestDistanceInternal(p *Emb_Man_t) {
    var nAttempts int32  = 20;
    var i, iNode, Dist int32;
    var clk abctime;
    var pPivot, pNext *Emb_Obj_t;
    Gia_ManRandom( 1 );
    Emb_ManResetTravId( p );
    // compute distances from several randomly selected PIs
    clk = Abc_Clock();
    fmt.Printf("From inputs: ");
    for  i = 0; i < nAttempts; i++  {
        iNode = Gia_ManRandom( 0 ) % Emb_ManCiNum(p);
        pPivot = Emb_ManCi( p, iNode );
        if Emb_ObjFanoutNum(pPivot) == 0 { i--; continue; }
        pNext = Emb_ObjFanout( pPivot, 0 );
        if !(Emb_ObjIsNode(pNext) != 0) { i--; continue; }
        Dist = Emb_ManComputeDistance_old( p, pPivot );
        fmt.Printf("%d ", Dist);
    }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // compute distances from several randomly selected POs
    clk = Abc_Clock();
    fmt.Printf("From outputs: ");
    for  i = 0; i < nAttempts; i++  {
        iNode = Gia_ManRandom( 0 ) % Emb_ManCoNum(p);
        pPivot = Emb_ManCo( p, iNode );
        pNext = Emb_ObjFanin( pPivot, 0 );
        if !(Emb_ObjIsNode(pNext) != 0) { i--; continue; }
        Dist = Emb_ManComputeDistance_old( p, pPivot );
        fmt.Printf("%d ", Dist);
    }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // compute distances from several randomly selected nodes
    clk = Abc_Clock();
    fmt.Printf("From nodes: ");
    for  i = 0; i < nAttempts; i++  {
        iNode = Gia_ManRandom( 0 ) % Gia_ManObjNum(p.pGia);
        if !~Gia_ManObj(p.pGia, iNode).Value { i--; continue; }
        pPivot = Emb_ManObj( p, Gia_ManObj(p.pGia, iNode).Value );
        if !(Emb_ObjIsNode(pPivot) != 0) { i--; continue; }
        Dist = Emb_ManComputeDistance_old( p, pPivot );
        fmt.Printf("%d ", Dist);
    }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Returns sorted array of node handles with largest fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTestDistance(pGia *Gia_Man_t) {
    var p *Emb_Man_t;
    var clk abctime  = Abc_Clock();
    p = Emb_ManStart( pGia );
//    Emb_ManPrintFanio( p );
    Emb_ManPrintStats( p );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    Gia_ManTestDistanceInternal( p );
    Emb_ManStop( p );
}




/**Function*************************************************************

  Synopsis    [Perform BFS from the set of nodes.]

  Description [Returns one of the most distant objects.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManPerformBfs(p *Emb_Man_t, vThis *Vec_Int_t, vNext *Vec_Int_t, pDist *Emb_Dat_t) *Emb_Obj_t {
    var vTemp *Vec_Int_t;
    var pThis, pNext, pResult *Emb_Obj_t;
    var i, k int32;
    assert(Vec_IntSize(vThis) > 0, "Vec_IntSize(vThis) > 0");
    for  p.nDistMax = 0; Vec_IntSize(vThis) > 0; p.nDistMax++  {
        p.nReached += Vec_IntSize(vThis);
        Vec_IntClear( vNext );
        // inlining of macro Emb_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vThis)) && ((pThis) = Emb_ManObj(p, Vec_IntEntry(vThis,i))); i++  {
            if pDist != nil { pDist[pThis.Value] = p.nDistMax; }
            // inlining of macro Emb_ObjForEachFanin
            for  k = 0; (k < int32(pThis.nFanins)) && (pNext = Emb_ObjFanin(pThis,k)); k++  {
                if Emb_ObjIsTravIdCurrent(p, pNext) != 0 {
                    continue; }
                Emb_ObjSetTravIdCurrent(p, pNext);
                Vec_IntPush( vNext, pNext.hHandle );
            }
            // inlining of macro Emb_ObjForEachFanout
            for  k = 0; (k < int32(pThis.nFanouts)) && (pNext = Emb_ObjFanout(pThis,k)); k++  {
                if Emb_ObjIsTravIdCurrent(p, pNext) != 0 {
                    continue; }
                Emb_ObjSetTravIdCurrent(p, pNext);
                Vec_IntPush( vNext, pNext.hHandle );
            }
        }
        vTemp = vThis; vThis = vNext; vNext = vTemp;
    }
    assert(Vec_IntSize(vNext) > 0, "Vec_IntSize(vNext) > 0");
    pResult = Emb_ManObj( p, Vec_IntEntry(vNext, 0) );
    assert(pDist == nil || pDist[pResult.Value] == p.nDistMax - 1, "pDist == nil || pDist[pResult.Value] == p.nDistMax - 1");
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Computes the distances from the given set of objects.]

  Description [Returns one of the most distant objects.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManConnectedComponents(p *Emb_Man_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var vThis, vNext, vResult *Vec_Int_t;
    var pThis *Emb_Obj_t;
    var i int32;
    vResult = Vec_IntAlloc( 1000 );
    vThis   = Vec_IntAlloc( 1000 );
    vNext   = Vec_IntAlloc( 1000 );
    p.nReached = 0;
    Emb_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        pThis = Emb_ManObj( p, Gia_ObjValue(pObj) );
        if Emb_ObjIsTravIdCurrent(p, pThis) != 0 {
            continue; }
        Emb_ObjSetTravIdCurrent( p, pThis );
        Vec_IntPush( vResult, pThis.hHandle );
        // perform BFS from this node
        Vec_IntClear( vThis );
        Vec_IntPush( vThis, pThis.hHandle );
        Emb_ManPerformBfs( p, vThis, vNext, nil ); 
    }
    Vec_IntFree( vThis );
    Vec_IntFree( vNext );
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Computes the distances from the given set of objects.]

  Description [Returns one of the most distant objects.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManFindDistances(p *Emb_Man_t, vStart *Vec_Int_t, pDist *Emb_Dat_t) *Emb_Obj_t {
    var vThis, vNext *Vec_Int_t;
    var pThis, pResult *Emb_Obj_t;
    var i int32;
    p.nReached = p.nDistMax = 0;
    vThis = Vec_IntAlloc( 1000 );
    vNext = Vec_IntAlloc( 1000 );
    Emb_ManIncrementTravId( p );
    // inlining of macro Emb_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vStart)) && ((pThis) = Emb_ManObj(p, Vec_IntEntry(vStart,i))); i++  {
        Emb_ObjSetTravIdCurrent( p, pThis );
        Vec_IntPush( vThis, pThis.hHandle );
    }
    pResult = Emb_ManPerformBfs( p, vThis, vNext, pDist );
    Vec_IntFree( vThis );
    Vec_IntFree( vNext );
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Traverses from the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Emb_ManRandomVertex(p *Emb_Man_t) *Emb_Obj_t {
    var pPivot *Emb_Obj_t;
    for {
        var iNode int32  = (911 * Gia_ManRandom(0)) % Gia_ManObjNum(p.pGia);
        if ~Gia_ManObj(p.pGia, iNode).Value {
            pPivot = Emb_ManObj( p, Gia_ManObj(p.pGia, iNode).Value );
        } else {
            pPivot = nil; }
    
if !( pPivot == nil || !(Emb_ObjIsNode(pPivot) != 0) ) { break }
}
    
    return pPivot;
}

/**Function*************************************************************

  Synopsis    [Computes the distances from the given set of objects.]

  Description [Returns one of the most distant objects.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_DumpGraphIntoFile(p *Emb_Man_t) {
    var pFile *os.File;
    var pThis, pNext *Emb_Obj_t;
    var i, k int32;
    pFile = os.OpenFile("1.g", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    // inlining of macro Emb_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pThis = Emb_ManObj(p,i)); i += Emb_ObjSize(pThis)  {
        if !(Emb_ObjIsTravIdCurrent(p, pThis) != 0) {
            continue; }
        // inlining of macro Emb_ObjForEachFanout
        for  k = 0; (k < int32(pThis.nFanouts)) && (pNext = Emb_ObjFanout(pThis,k)); k++  {
            assert((Emb_ObjIsTravIdCurrent(p, pNext) != 0), "Emb_ObjIsTravIdCurrent(p, pNext)");
            fmt.Fprintf(pFile, "%d %d\n", pThis.Value, pNext.Value);
        }
    }
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Computes dimentions of the graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManComputeDimensions(p *Emb_Man_t, nDims int32) {
    var pRandom, pPivot *Emb_Obj_t;
    var vStart, vComps *Vec_Int_t;
    var d, nReached int32;
    var i int32;//, Counter;
    // connect unconnected components
    vComps = Emb_ManConnectedComponents( p );
//    printf( "Components = %d. Considered %d objects (out of %d).\n", Vec_IntSize(vComps), p->nReached, Emb_ManObjNum(p) );
    if Vec_IntSize(vComps) > 1 {
        var pFanin *Emb_Obj_t;
        var pObj *Emb_Obj_t = Emb_ManObj( p, 0 );
        // inlining of macro Emb_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vComps)) && ((pFanin) = Emb_ManObj(p, Vec_IntEntry(vComps,i))); i++  {
            assert((Emb_ObjIsCo(pFanin) != 0), "Emb_ObjIsCo(pFanin)");
            pFanin.Fanios[pFanin.nFanins + pFanin.nFanouts-1] = 
                pObj.Fanios[i] = pObj.hHandle - pFanin.hHandle;
        }
    }
    Vec_IntFree( vComps );
    // allocate memory for vectors
    assert(p.pVecs == nil, "p.pVecs == nil");
    // inlining of macro ABC_CALLOC
    p.pVecs = ((Emb_Dat_t *) calloc((size_t)(p.nObjs * nDims), sizeof(Emb_Dat_t)));
//    for ( i = 0; i < p->nObjs * nDims; i++ )
//        p->pVecs[i] = ABC_INFINITY;
    vStart = Vec_IntAlloc( nDims );
    // get the pivot vertex
    pRandom = Emb_ManRandomVertex( p );
    Vec_IntPush( vStart, pRandom.hHandle );
    // get the most distant vertex from the pivot
    pPivot = Emb_ManFindDistances( p, vStart, nil );
//    Emb_DumpGraphIntoFile( p );
    nReached = p.nReached;
    if nReached < Emb_ManObjNum(p) {
//        printf( "Considering a connected component with %d objects (out of %d).\n", p->nReached, Emb_ManObjNum(p) );
    }
    // start dimensions with this vertex
    Vec_IntClear( vStart );
    for  d = 0; d < nDims; d++  {
//        printf( "%3d : Adding vertex %7d with distance %3d.\n", d+1, pPivot->Value, p->nDistMax ); 
        Vec_IntPush( vStart, pPivot.hHandle );
        if d+1 == nReached {
            break; }
        pPivot = Emb_ManFindDistances( p, vStart, Emb_ManVec(p, d) );
        assert(nReached == p.nReached, "nReached == p.nReached");
    }
    Vec_IntFree( vStart );
    // make sure the number of reached objects is correct
//    Counter = 0;
//    for ( i = 0; i < p->nObjs; i++ )
//        if ( p->pVecs[i] < ABC_INFINITY )
//            Counter++;
//    assert( Counter == nReached );
}

/**Function*************************************************************

  Synopsis    [Allocated square matrix of floats.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManMatrAlloc(nDims int32) **float32 {
    var i int32;
    // inlining of macro ABC_ALLOC
    var pMatr **float32  = (float **)((char *) malloc(sizeof(char) * (size_t)(sizeof(float *) * nDims + sizeof(float) * nDims * nDims)));
    pMatr[0] = (float *)(pMatr + nDims);
    for  i = 1; i < nDims; i++  {
        pMatr[i] = pMatr[i-1] + nDims;
    }
    return pMatr;
}

/**Function*************************************************************

  Synopsis    [Computes covariance matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManComputeCovariance(p *Emb_Man_t, nDims int32) {
    var pOne, pTwo *Emb_Dat_t;
    var Ave float64;
    var pRow *float32;
    var d, i, k, v int32;
    // average vectors
    for  d = 0; d < nDims; d++  {
        // compute average
        Ave = 0.0;
        pOne = Emb_ManVec( p, d );
        for  v = 0; v < p.nObjs; v++  {
            if pOne[v] < ABC_INFINITY {
                Ave += pOne[v]; }
        }
        Ave /= p.nReached;
        // update the vector
        for  v = 0; v < p.nObjs; v++  {
            if pOne[v] < ABC_INFINITY {
                pOne[v] -= Ave;
            } else {
                pOne[v] = 0.0; }
        }        
    }
    // compute the matrix
    assert(p.pMatr == nil, "p.pMatr == nil");
    assert(p.pEigen == nil, "p.pEigen == nil");
    p.pMatr = Emb_ManMatrAlloc( nDims );
    p.pEigen = Emb_ManMatrAlloc( nDims );
    for  i = 0; i < nDims; i++  {
        pOne = Emb_ManVec( p, i );
        pRow = p.pMatr[i];
        for  k = 0; k < nDims; k++  {
            pTwo = Emb_ManVec( p, k );
            pRow[k] = 0.0;
            for  v = 0; v < p.nObjs; v++  {
                pRow[k] += pOne[v]*pTwo[v];
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Returns random vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManVecRandom(pVec *float32, nDims int32) {
    var i int32;
    for  i = 0; i < nDims; i++  {
        pVec[i] = Gia_ManRandom( 0 );
    }
}

/**Function*************************************************************

  Synopsis    [Returns normalized vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManVecNormal(pVec *float32, nDims int32) {
    var i int32;
    var Norm float64  = 0.0;
    for  i = 0; i < nDims; i++  {
        Norm += pVec[i] * pVec[i];
    }
    Norm = math.Pow(float64(Norm), float64(0.5));
    for  i = 0; i < nDims; i++  {
        pVec[i] /= Norm;
    }
}

/**Function*************************************************************

  Synopsis    [Multiplies vector by vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManVecMultiplyOne(pVec0 *float32, pVec1 *float32, nDims int32) float32 {
    var Res float32  = 0.0;
    var i int32;
    for  i = 0; i < nDims; i++  {
        Res += pVec0[i] * pVec1[i];
    }
    return Res;
}

/**Function*************************************************************

  Synopsis    [Copies the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManVecCopyOne(pVecDest *float32, pVecSour *float32, nDims int32) {
    var i int32;
    for  i = 0; i < nDims; i++  {
        pVecDest[i] = pVecSour[i];
    }
}

/**Function*************************************************************

  Synopsis    [Multiplies matrix by vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManVecMultiply(pMatr **float32, pVec *float32, nDims int32, pRes *float32) {
    var k int32;
    for  k = 0; k < nDims; k++  {
        pRes[k] = Emb_ManVecMultiplyOne( pMatr[k], pVec, nDims );
    }
}

/**Function*************************************************************

  Synopsis    [Multiplies vector by matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManVecOrthogonolizeOne(pEigen *float32, pVecI *float32, nDims int32, pVecRes *float32) {
    var k int32;
    for  k = 0; k < nDims; k++  {
        pVecRes[k] = pVecI[k] - pEigen[k] * Emb_ManVecMultiplyOne( pVecI, pEigen, nDims );
    }
}

/**Function*************************************************************

  Synopsis    [Computes the first nSols eigen-vectors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManComputeEigenvectors(p *Emb_Man_t, nDims int32, nSols int32) {
    var pVecUiHat, pVecUi *float32;
    var i, j, k int32;
    assert(nSols < nDims, "nSols < nDims");
    pVecUiHat = p.pEigen[nSols];
    for  i = 0; i < nSols; i++  {
        pVecUi = p.pEigen[i];
        Emb_ManVecRandom( pVecUiHat, nDims );
        Emb_ManVecNormal( pVecUiHat, nDims );
        k = 0;
        for {
            k++;
            Emb_ManVecCopyOne( pVecUi, pVecUiHat, nDims );
            for  j = 0; j < i; j++  {
                Emb_ManVecOrthogonolizeOne( p.pEigen[j], pVecUi, nDims, pVecUiHat );
                Emb_ManVecCopyOne( pVecUi, pVecUiHat, nDims );
            }
            Emb_ManVecMultiply( p.pMatr, pVecUi, nDims, pVecUiHat );
            Emb_ManVecNormal( pVecUiHat, nDims );
        
if !( Emb_ManVecMultiplyOne( pVecUiHat, pVecUi, nDims ) < 0.999 && k < 100 ) { break }
} 
        Emb_ManVecCopyOne( pVecUi, pVecUiHat, nDims );
//        printf( "Converged after %d iterations.\n", k );
    }
}

/**Function*************************************************************

  Synopsis    [Derives solutions from original vectors and eigenvectors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManComputeSolutions(p *Emb_Man_t, nDims int32, nSols int32) {
    var pX *Emb_Dat_t;
    var pY *float32;
    var i, j, k int32;
    assert(p.pSols == nil, "p.pSols == nil");
    // inlining of macro ABC_CALLOC
    p.pSols = ((float *) calloc((size_t)(p.nObjs * nSols), sizeof(float)));
    for  i = 0; i < nDims; i++  {
        pX = Emb_ManVec( p, i );
        for  j = 0; j < nSols; j++  {
            pY = Emb_ManSol( p, j );
            for  k = 0; k < p.nObjs; k++  {
                pY[k] += pX[k] * p.pEigen[j][i];
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Projects into square of size [0;GIA_PLACE_SIZE] x [0;GIA_PLACE_SIZE].]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManDerivePlacement(p *Emb_Man_t, nSols int32) {
    var pY0 *float32;
    var pY1 *float32;
    var Max0 float32;
    var Max1 float32;
    var Min0 float32;
    var Min1 float32;
    var Str0 float32;
    var Str1 float32;
    var pPerm0, pPerm1 *int32;
    var k int32;
    if nSols != 2 {
        return; }
    // compute intervals
    Min0 =  ABC_INFINITY;
    Max0 = -ABC_INFINITY;
    pY0 = Emb_ManSol( p, 0 );
    for  k = 0; k < p.nObjs; k++  {
        Min0 = Abc_MinInt( Min0, pY0[k] );
        Max0 = Abc_MaxInt( Max0, pY0[k] );
    }
    Str0 = 1.0*GIA_PLACE_SIZE/(Max0 - Min0);
    // update the coordinates
    for  k = 0; k < p.nObjs; k++  {
        pY0[k] = (pY0[k] != 0.0) ? ((pY0[k] - Min0) * Str0) : 0.0;
    }

    // compute intervals
    Min1 =  ABC_INFINITY;
    Max1 = -ABC_INFINITY;
    pY1 = Emb_ManSol( p, 1 );
    for  k = 0; k < p.nObjs; k++  {
        Min1 = Abc_MinInt( Min1, pY1[k] );
        Max1 = Abc_MaxInt( Max1, pY1[k] );
    }
    Str1 = 1.0*GIA_PLACE_SIZE/(Max1 - Min1);
    // update the coordinates
    for  k = 0; k < p.nObjs; k++  {
        pY1[k] = (pY1[k] != 0.0) ? ((pY1[k] - Min1) * Str1) : 0.0;
    }

    // derive the order of these numbers
    pPerm0 = Gia_SortFloats( pY0, nil, p.nObjs );
    pPerm1 = Gia_SortFloats( pY1, nil, p.nObjs );

    // average solutions and project them into square [0;GIA_PLACE_SIZE] x [0;GIA_PLACE_SIZE]
    // inlining of macro ABC_ALLOC
    p.pPlacement = ((unsigned short *) malloc(sizeof(unsigned short) * (size_t)(2 * p.nObjs)));
    for  k = 0; k < p.nObjs; k++  {
        p.pPlacement[2*pPerm0[k]+0] = (unsigned short)int32((1.0 * k * GIA_PLACE_SIZE / p.nObjs));
        p.pPlacement[2*pPerm1[k]+1] = (unsigned short)int32((1.0 * k * GIA_PLACE_SIZE / p.nObjs));
    }
    // inlining of macro ABC_FREE
    pPerm0 = nil;
    // inlining of macro ABC_FREE
    pPerm1 = nil;
}


/**Function*************************************************************

  Synopsis    [Computes wire-length.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManComputeHPWL(p *Emb_Man_t) float64 {
    var Result float64  = 0.0;
    var pThis, pNext *Emb_Obj_t;
    var i, k, iMinX, iMaxX, iMinY, iMaxY int32;
    if p.pPlacement == nil {
        return 0.0; }
    // inlining of macro Emb_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pThis = Emb_ManObj(p,i)); i += Emb_ObjSize(pThis)  {
        iMinX = iMaxX = p.pPlacement[2*pThis.Value+0];
        iMinY = iMaxY = p.pPlacement[2*pThis.Value+1];
        // inlining of macro Emb_ObjForEachFanout
        for  k = 0; (k < int32(pThis.nFanouts)) && (pNext = Emb_ObjFanout(pThis,k)); k++  {
            iMinX = Abc_MinInt( iMinX, p.pPlacement[2*pNext.Value+0] );
            iMaxX = Abc_MaxInt( iMaxX, p.pPlacement[2*pNext.Value+0] );
            iMinY = Abc_MinInt( iMinY, p.pPlacement[2*pNext.Value+1] );
            iMaxY = Abc_MaxInt( iMaxY, p.pPlacement[2*pNext.Value+1] );
        }
        Result += (iMaxX - iMinX) + (iMaxY - iMinY);
    }
    return Result;
}


/**Function*************************************************************

  Synopsis    [Performs iterative refinement of the given placement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManPlacementRefine(p *Emb_Man_t, nIters int32, fVerbose int32) {
    var pThis, pNext *Emb_Obj_t;
    var CostThis, CostPrev float64;
    var pEdgeX, pEdgeY *float32;
    var pVertX, pVertY *float32;
    var VertX, VertY float32;
    var pPermX, pPermY *int32;
    var i, k, Iter, iMinX, iMaxX, iMinY, iMaxY int32;
    var clk abctime  = Abc_Clock();
    if p.pPlacement == nil {
        return; }
    // inlining of macro ABC_ALLOC
    pEdgeX = ((float *) malloc(sizeof(float) * (size_t)(p.nObjs)));
    // inlining of macro ABC_ALLOC
    pEdgeY = ((float *) malloc(sizeof(float) * (size_t)(p.nObjs)));
    // inlining of macro ABC_ALLOC
    pVertX = ((float *) malloc(sizeof(float) * (size_t)(p.nObjs)));
    // inlining of macro ABC_ALLOC
    pVertY = ((float *) malloc(sizeof(float) * (size_t)(p.nObjs)));
    // refine placement
    CostPrev = 0.0;
    for  Iter = 0; Iter < nIters; Iter++  {
        // compute centers of hyperedges
        CostThis = 0.0;
        // inlining of macro Emb_ManForEachObj
        for  i = 0; (i < p.nObjData) && (pThis = Emb_ManObj(p,i)); i += Emb_ObjSize(pThis)  {
            iMinX = iMaxX = p.pPlacement[2*pThis.Value+0];
            iMinY = iMaxY = p.pPlacement[2*pThis.Value+1];
            // inlining of macro Emb_ObjForEachFanout
            for  k = 0; (k < int32(pThis.nFanouts)) && (pNext = Emb_ObjFanout(pThis,k)); k++  {
                iMinX = Abc_MinInt( iMinX, p.pPlacement[2*pNext.Value+0] );
                iMaxX = Abc_MaxInt( iMaxX, p.pPlacement[2*pNext.Value+0] );
                iMinY = Abc_MinInt( iMinY, p.pPlacement[2*pNext.Value+1] );
                iMaxY = Abc_MaxInt( iMaxY, p.pPlacement[2*pNext.Value+1] );
            }
            pEdgeX[pThis.Value] = 0.5 * (iMaxX + iMinX);
            pEdgeY[pThis.Value] = 0.5 * (iMaxY + iMinY);
            CostThis += (iMaxX - iMinX) + (iMaxY - iMinY);
        }
        // compute new centers of objects
        // inlining of macro Emb_ManForEachObj
        for  i = 0; (i < p.nObjData) && (pThis = Emb_ManObj(p,i)); i += Emb_ObjSize(pThis)  {
            VertX = pEdgeX[pThis.Value];
            VertY = pEdgeY[pThis.Value];
            // inlining of macro Emb_ObjForEachFanin
            for  k = 0; (k < int32(pThis.nFanins)) && (pNext = Emb_ObjFanin(pThis,k)); k++  {
                VertX += pEdgeX[pNext.Value];
                VertY += pEdgeY[pNext.Value];
            }
            pVertX[pThis.Value] = VertX / (Emb_ObjFaninNum(pThis) + 1);
            pVertY[pThis.Value] = VertY / (Emb_ObjFaninNum(pThis) + 1);
        }
        // sort these numbers
        pPermX = Gia_SortFloats( pVertX, nil, p.nObjs );
        pPermY = Gia_SortFloats( pVertY, nil, p.nObjs );
        for  k = 0; k < p.nObjs; k++  {
            p.pPlacement[2*pPermX[k]+0] = (unsigned short)int32((1.0 * k * GIA_PLACE_SIZE / p.nObjs));
            p.pPlacement[2*pPermY[k]+1] = (unsigned short)int32((1.0 * k * GIA_PLACE_SIZE / p.nObjs));
        }
        // inlining of macro ABC_FREE
        pPermX = nil;
        // inlining of macro ABC_FREE
        pPermY = nil;
        // evaluate cost
        if fVerbose != 0 {
        fmt.Printf("%2d : HPWL = %e  ", Iter+1, CostThis);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
    }
    // inlining of macro ABC_FREE
    pEdgeX = nil;
    // inlining of macro ABC_FREE
    pEdgeY = nil;
    // inlining of macro ABC_FREE
    pVertX = nil;
    // inlining of macro ABC_FREE
    pVertY = nil;
}


/**Function*************************************************************

  Synopsis    [Derives solutions from original vectors and eigenvectors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManPrintSolutions(p *Emb_Man_t, nSols int32) {
    var pSol *float32;
    var i, k int32;
    for  i = 0; i < nSols; i++  {
        pSol = Emb_ManSol( p, i );
        for  k = 0; k < p.nObjs; k++  {
            fmt.Printf("%4d ", int32((100 * pSol[k])));
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Prepares image for dumping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManDumpGnuplotPrepare(p *Emb_Man_t) *Vec_Int_t {
//    int nRows = 496;
//    int nCols = 710;
    var nRows int32  = 500;
    var nCols int32  = 700;
    var vLines *Vec_Int_t;
    var pThis *Emb_Obj_t;
    var pBuffer string;
    var ppRows *string;
    var i, k, placeX, placeY int32;
    var fStart int32;
    // alloc memory
    // inlining of macro ABC_CALLOC
    pBuffer = ((char *) calloc((size_t)(nRows * (nCols+1)), sizeof(char)));
    // inlining of macro ABC_ALLOC
    ppRows  = ((char * *) malloc(sizeof(char *) * (size_t)(nRows)));
    for  i = 0; i < nRows; i++  {
        ppRows[i] = pBuffer + i*(nCols+1);
    }
    // put data into them
    // inlining of macro Emb_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pThis = Emb_ManObj(p,i)); i += Emb_ObjSize(pThis)  {
        placeX = p.pPlacement[2*pThis.Value+0] * nCols / (1<<16);
        placeY = p.pPlacement[2*pThis.Value+1] * nRows / (1<<16);
        assert(placeX < nCols && placeY < nRows, "placeX < nCols && placeY < nRows");
        ppRows[placeY][placeX] = 1;
    }
    // select lines
    vLines = Vec_IntAlloc( 1000 );
    for  i = 0; i < nRows; i++  {
        fStart = 0;
        for  k = 0; k <= nCols; k++  {
            if ppRows[i][k] && !(fStart != 0) {
                Vec_IntPush( vLines, k );
                Vec_IntPush( vLines, i );
                fStart = 1;
            }
            if !ppRows[i][k] && (fStart != 0) {
                Vec_IntPush( vLines, k-1 );
                Vec_IntPush( vLines, i );
                fStart = 0;
            }
        }
        assert(fStart == 0, "fStart == 0");
    }
    // inlining of macro ABC_FREE
    pBuffer = "";
    // inlining of macro ABC_FREE
    ppRows = nil;
    return vLines;
}

/**Function*************************************************************

  Synopsis    [Derives solutions from original vectors and eigenvectors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Emb_ManDumpGnuplot(p *Emb_Man_t, pName string, fDumpLarge int32, fShowImage int32) {
    extern void Gia_ManGnuplotShow( char * pPlotFileName );
//    char * pDirectory = "place\\";
    var pDirectory string  = "";
//    extern char * Ioa_TimeStamp();
    var pFile *os.File;
    var Buffer string // was [1000]byte in C

    var pThis, pNext *Emb_Obj_t;
    var i, k int32;
    if p.pPlacement == nil {
        fmt.Printf("Emb_ManDumpGnuplot(): Placement is not available.\n");
        return;
    }
    Buffer = fmt.Sprintf("%s%s", pDirectory, Gia_FileNameGenericAppend(pName, ".plt")); 
    pFile = fopen( Buffer, "w" );
    fmt.Fprintf(pFile, "# This Gnuplot file was produced by ABC on %s\n", Ioa_TimeStamp());
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "set nokey\n");
    fmt.Fprintf(pFile, "\n");
    if !(fShowImage != 0) {
//    fprintf( pFile, "set terminal postscript\n" );
    fmt.Fprintf(pFile, "set terminal gif font \'arial\' 10 size 800,600 xffffff x000000 x000000 x000000\n");
    fmt.Fprintf(pFile, "set output \'%s\'\n", Gia_FileNameGenericAppend(pName, ".gif"));
    fmt.Fprintf(pFile, "\n");
    }
    fmt.Fprintf(pFile, "set title \"%s :  PI = %d   PO = %d   FF = %d   Node = %d   Obj = %d  HPWL = %.2e\\n", pName, Emb_ManPiNum(p), Emb_ManPoNum(p), Emb_ManRegNum(p), Emb_ManNodeNum(p), Emb_ManObjNum(p), Emb_ManComputeHPWL(p));
    fmt.Fprintf(pFile, "(image generated by ABC and Gnuplot on %s)\"", Ioa_TimeStamp());
    fmt.Fprintf(pFile, "font \"Times, 12\"\n");
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "plot [:] '-' w l\n");
    fmt.Fprintf(pFile, "\n");
    if fDumpLarge != 0 {
        var begX, begY, endX, endY int32;
        var vLines *Vec_Int_t  = Emb_ManDumpGnuplotPrepare( p );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLines)) && (((begX) = Vec_IntEntry(vLines, i)), 1); i++  {
            begY = Vec_IntEntry( vLines, i+1 );
            endX = Vec_IntEntry( vLines, i+2 );
            endY = Vec_IntEntry( vLines, i+3 );
            i += 3;
            fmt.Fprintf(pFile, "%5d %5d\n", begX, begY);
            fmt.Fprintf(pFile, "%5d %5d\n", endX, endY);
            fmt.Fprintf(pFile, "\n");
        }
        Vec_IntFree( vLines );
    
    } else {
        // inlining of macro Emb_ManForEachObj
        for  i = 0; (i < p.nObjData) && (pThis = Emb_ManObj(p,i)); i += Emb_ObjSize(pThis)  {
            if !(Emb_ObjIsTravIdCurrent(p, pThis) != 0) {
                continue; }
            // inlining of macro Emb_ObjForEachFanout
            for  k = 0; (k < int32(pThis.nFanouts)) && (pNext = Emb_ObjFanout(pThis,k)); k++  {
                assert((Emb_ObjIsTravIdCurrent(p, pNext) != 0), "Emb_ObjIsTravIdCurrent(p, pNext)");
                fmt.Fprintf(pFile, "%5d %5d\n", p.pPlacement[2*pThis.Value+0], p.pPlacement[2*pThis.Value+1]);
                fmt.Fprintf(pFile, "%5d %5d\n", p.pPlacement[2*pNext.Value+0], p.pPlacement[2*pNext.Value+1]);
                fmt.Fprintf(pFile, "\n");
            }
        }
    }
    fmt.Fprintf(pFile, "EOF\n");
    fmt.Fprintf(pFile, "\n");
    if fShowImage != 0 {
    fmt.Fprintf(pFile, "pause -1 \"Close window\"\n");  // Hit return to continue
    fmt.Fprintf(pFile, "reset\n");
    fmt.Fprintf(pFile, "\n");
    
    } else {
    fmt.Fprintf(pFile, "# pause -1 \"Close window\"\n");  // Hit return to continue
    fmt.Fprintf(pFile, "# reset\n");
    fmt.Fprintf(pFile, "\n");
    }
    _ = (pFile).Close();
    if fShowImage != 0 {
        Gia_ManGnuplotShow( Buffer ); }
}

/**Function*************************************************************

  Synopsis    [Computes dimentions of the graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSolveProblem(pGia *Gia_Man_t, pPars *Emb_Par_t) {
    var p *Emb_Man_t;
    var i int32;
    var clkSetup abctime;
    var clk abctime;
//   Gia_ManTestDistance( pGia );

    // transform AIG into internal data-structure
clk = Abc_Clock();
    if pPars.fCluster {
        p = Emb_ManStart( pGia );
        if pPars.fVerbose {
            fmt.Printf("Clustered: ");
            Emb_ManPrintStats( p );
        }
    
    } else {
        p = Emb_ManStartSimple( pGia ); }
    p.fVerbose = pPars.fVerbose;
//    Emb_ManPrintFanio( p );

    // prepare data-structure
    Gia_ManRandom( 1 );  // reset random numbers for deterministic behavior
    Emb_ManResetTravId( p );
    Emb_ManSetValue( p );
clkSetup = Abc_Clock() - clk;

clk = Abc_Clock();
    Emb_ManComputeDimensions( p, pPars.nDims );
if pPars.fVerbose {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Setup     "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((clkSetup)))/(float64(int64(time.Second))));; }
if pPars.fVerbose {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Dimensions"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }

clk = Abc_Clock();
    Emb_ManComputeCovariance( p, pPars.nDims );
if pPars.fVerbose {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Matrix    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }

clk = Abc_Clock();
    Emb_ManComputeEigenvectors( p, pPars.nDims, pPars.nSols );
    Emb_ManComputeSolutions( p, pPars.nDims, pPars.nSols );
    Emb_ManDerivePlacement( p, pPars.nSols );
if pPars.fVerbose {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Eigenvecs "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }

    if pPars.fRefine {
clk = Abc_Clock();
    Emb_ManPlacementRefine( p, pPars.nIters, pPars.fVerbose );
if pPars.fVerbose {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Refinement"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
    }

    if (pPars.fDump || pPars.fDumpLarge) && pPars.nSols == 2 {
clk = Abc_Clock();
        Emb_ManDumpGnuplot( p, pGia.pName, pPars.fDumpLarge, pPars.fShowImage );
if pPars.fVerbose {
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Image dump"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));; }
    }

    // transfer placement
    if Gia_ManObjNum(pGia) == p.nObjs {
        // assuming normalized ordering of the AIG
        // inlining of macro ABC_CALLOC
        pGia.pPlacement = ((Gia_Plc_t *) calloc((size_t)(p.nObjs), sizeof(Gia_Plc_t)));
        for  i = 0; i < p.nObjs; i++  {
            pGia.pPlacement[i].xCoord = p.pPlacement[2*i+0];
            pGia.pPlacement[i].yCoord = p.pPlacement[2*i+1];
        }
    }
    Emb_ManStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

