package gia

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


/* 
    The code is based on the paper by F. A. Aloul, I. L. Markov, and K. A. Sakallah.
    "FORCE: A Fast and Easy-To-Implement Variable-Ordering Heuristic", Proc. GLSVLSI�03.
    http://www.eecs.umich.edu/~imarkov/pubs/conf/glsvlsi03-force.pdf
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Frc_Obj_t struct {
    fCi uint32; /* C bitfield: 1 bits; review packing and truncation */    // terminal node CI
    fCo uint32; /* C bitfield: 1 bits; review packing and truncation */    // terminal node CO
    fMark0 uint32; /* C bitfield: 1 bits; review packing and truncation */    // first user-controlled mark
    fMark1 uint32; /* C bitfield: 1 bits; review packing and truncation */    // second user-controlled mark
    nFanins uint32; /* C bitfield: 28 bits; review packing and truncation */    // the number of fanins
    nFanouts       uint32;         // the number of fanouts
    iFanout       uint32;          // the current number of fanouts
    hHandle            int32;          // the handle of the node
    pPlace            int32;           // the placement of each node
    union {
    float          fEdgeCenter;      // center-of-gravity of the edge
    unsigned       iFanin; 
    };
    Fanios [0]int32;        // the array of fanins/fanouts
}

   
type Frc_Man_t struct {
    pGia     *Gia_Man_t;             // the original AIG manager
    vCis     *Vec_Int_t;             // the vector of CIs (PIs + LOs)
    vCos     *Vec_Int_t;             // the vector of COs (POs + LIs)
    nObjs            int32;            // the number of objects
    nRegs            int32;            // the number of registers
    pObjData           *int32;         // the array containing data for objects
    nObjData            int32;         // the size of array to store the logic network
    fVerbose            int32;         // verbose output flag
    nCutCur            int32;          // current cut 
    nCutMax            int32;          // max cut seen
}

func Frc_ManRegNum(p *Frc_Man_t) int32 { return p.nRegs;                                                     }
func Frc_ManCiNum(p *Frc_Man_t) int32 { return Vec_IntSize(p.vCis);                                         }
func Frc_ManCoNum(p *Frc_Man_t) int32 { return Vec_IntSize(p.vCos);                                         }
func Frc_ManPiNum(p *Frc_Man_t) int32 { return Vec_IntSize(p.vCis) - p.nRegs;                              }
func Frc_ManPoNum(p *Frc_Man_t) int32 { return Vec_IntSize(p.vCos) - p.nRegs;                              }
func Frc_ManObjNum(p *Frc_Man_t) int32 { return p.nObjs;                                                     } 
func Frc_ManNodeNum(p *Frc_Man_t) int32 { return p.nObjs - Vec_IntSize(p.vCis) - Vec_IntSize(p.vCos);       } 

func Frc_ManObj(p *Frc_Man_t, hHandle int32) *Frc_Obj_t { return (Frc_Obj_t *)(p.pObjData + hHandle);                         } 
func Frc_ManCi(p *Frc_Man_t, i int32) *Frc_Obj_t { return Frc_ManObj( p, Vec_IntEntry(p.vCis,i) );                     }
func Frc_ManCo(p *Frc_Man_t, i int32) *Frc_Obj_t { return Frc_ManObj( p, Vec_IntEntry(p.vCos,i) );                     }

func Frc_ObjIsTerm(pObj *Frc_Obj_t) int32 { return pObj.fCi || pObj.fCo;                                       } 
func Frc_ObjIsCi(pObj *Frc_Obj_t) int32 { return pObj.fCi;                                                    } 
func Frc_ObjIsCo(pObj *Frc_Obj_t) int32 { return pObj.fCo;                                                    } 
func Frc_ObjIsPi(pObj *Frc_Obj_t) int32 { return pObj.fCi && pObj.nFanins == 0;                              } 
func Frc_ObjIsPo(pObj *Frc_Obj_t) int32 { return pObj.fCo && pObj.nFanouts == 0;                             } 
func Frc_ObjIsNode(pObj *Frc_Obj_t) int32 { return!Frc_ObjIsTerm(pObj) && pObj.nFanins > 0;                     } 
func Frc_ObjIsConst0(pObj *Frc_Obj_t) int32 { return!Frc_ObjIsTerm(pObj) && pObj.nFanins == 0;                    } 

func Frc_ObjSize(pObj *Frc_Obj_t) int32 { return sizeof(Frc_Obj_t) / 4 + pObj.nFanins + pObj.nFanouts;       } 
func Frc_ObjFaninNum(pObj *Frc_Obj_t) int32 { return pObj.nFanins;                                                } 
func Frc_ObjFanoutNum(pObj *Frc_Obj_t) int32 { return pObj.nFanouts;                                               } 
func Frc_ObjFanin(pObj *Frc_Obj_t, i int32) *Frc_Obj_t { return (Frc_Obj_t *)(((int *)pObj) - pObj.Fanios[i]);               } 
func Frc_ObjFanout(pObj *Frc_Obj_t, i int32) *Frc_Obj_t { return (Frc_Obj_t *)(((int *)pObj) + pObj.Fanios[pObj.nFanins+i]); } 

// Original macro:
// #define Frc_ManForEachObj( p, pObj, i )               \
//     for ( i = 0; (i < p->nObjData) && (pObj = Frc_ManObj(p,i)); i += Frc_ObjSize(pObj) )

// Original macro:
// #define Frc_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Frc_ManObj(p, Vec_IntEntry(vVec,i))); i++ )


// Original macro:
// #define Frc_ManForEachNode( p, pObj, i )              \
//     for ( i = 0; (i < p->nObjData) && (pObj = Frc_ManObj(p,i)); i += Frc_ObjSize(pObj) ) if ( Frc_ObjIsTerm(pObj) ) {} else

// Original macro:
// #define Frc_ManForEachCi( p, pObj, i )                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCis)) && (pObj = Frc_ManObj(p,Vec_IntEntry(p->vCis,i))); i++ )

// Original macro:
// #define Frc_ManForEachCo( p, pObj, i )                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && (pObj = Frc_ManObj(p,Vec_IntEntry(p->vCos,i))); i++ )


// Original macro:
// #define Frc_ObjForEachFanin( pObj, pNext, i )         \
//     for ( i = 0; (i < (int)pObj->nFanins) && (pNext = Frc_ObjFanin(pObj,i)); i++ )

// Original macro:
// #define Frc_ObjForEachFaninReverse( pObj, pNext, i )  \
//     for ( i = (int)pObj->nFanins - 1; (i >= 0) && (pNext = Frc_ObjFanin(pObj,i)); i-- )

// Original macro:
// #define Frc_ObjForEachFanout( pObj, pNext, i )        \
//     for ( i = 0; (i < (int)pObj->nFanouts) && (pNext = Frc_ObjFanout(pObj,i)); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Creates fanin/fanout pair.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ObjAddFanin(pObj *Frc_Obj_t, pFanin *Frc_Obj_t) { 
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
func Frc_ManStartSimple(pGia *Gia_Man_t) *Frc_Man_t {
    
    var pObjLog, pFanLog *Frc_Obj_t;
    var pObj *Gia_Obj_t;//, * pObjRi, * pObjRo;
    var i, nNodes, hHandle int32;
    // prepare the AIG
    Gia_ManCreateRefs( pGia );
    // create logic network
    // inlining of macro ABC_CALLOC
    p := &Frc_Man_t{};
    p.pGia  = pGia;
    p.nRegs = Gia_ManRegNum(pGia);
    p.vCis  = Vec_IntAlloc( Gia_ManCiNum(pGia) );
    p.vCos  = Vec_IntAlloc( Gia_ManCoNum(pGia) );
    p.nObjData = (sizeof(Frc_Obj_t) / 4) * Gia_ManObjNum(pGia) + 2 * (2 * Gia_ManAndNum(pGia) + Gia_ManCoNum(pGia));
    // inlining of macro ABC_CALLOC
    p.pObjData = ((int *) calloc((size_t)(p.nObjData), sizeof(int)));
    // create constant node
    Gia_ManConst0(pGia).Value = hHandle;
    pObjLog = Frc_ManObj( p, hHandle );
    pObjLog.hHandle  = hHandle;
    pObjLog.nFanins  = 0;
    pObjLog.nFanouts = Gia_ObjRefNum( pGia, Gia_ManConst0(pGia) );
    // count objects
    hHandle += Frc_ObjSize( pObjLog );
    nNodes = 1;
    p.nObjs++;
    // create the PIs
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
        // create PI object
        pObj.Value = hHandle;
        Vec_IntPush( p.vCis, hHandle );
        pObjLog = Frc_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = 0;
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        pObjLog.fCi = 0;
        // count objects
        hHandle += Frc_ObjSize( pObjLog );
        p.nObjs++;
    }
    // create internal nodes
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(Gia_ObjRefNum( pGia, pObj ) > 0, "Gia_ObjRefNum( pGia, pObj ) > 0");
        // create node object
        pObj.Value = hHandle;
        pObjLog = Frc_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = 2;
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        // add fanins
        pFanLog = Frc_ManObj( p, Gia_ObjValue(Gia_ObjFanin0(pObj)) ); 
        Frc_ObjAddFanin( pObjLog, pFanLog );
        pFanLog = Frc_ManObj( p, Gia_ObjValue(Gia_ObjFanin1(pObj)) ); 
        Frc_ObjAddFanin( pObjLog, pFanLog );
        // count objects
        hHandle += Frc_ObjSize( pObjLog );
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
        pObjLog = Frc_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = 1;
        pObjLog.nFanouts = 0;
        pObjLog.fCo = 1;
        // add fanins
        pFanLog = Frc_ManObj( p, Gia_ObjValue(Gia_ObjFanin0(pObj)) );
        Frc_ObjAddFanin( pObjLog, pFanLog );
        // count objects
        hHandle += Frc_ObjSize( pObjLog );
        p.nObjs++;
    }
    // connect registers
//    Gia_ManForEachRiRo( pGia, pObjRi, pObjRo, i )
//        Frc_ObjAddFanin( Frc_ManObj(p,Gia_ObjValue(pObjRo)), Frc_ManObj(p,Gia_ObjValue(pObjRi)) );
    assert(nNodes  == Frc_ManNodeNum(p), "nNodes  == Frc_ManNodeNum(p)");
    assert(hHandle == p.nObjData, "hHandle == p.nObjData");
    if hHandle != p.nObjData {
        fmt.Printf("Frc_ManStartSimple(): Fatal error in internal representation.\n"); }
    // make sure the fanin/fanout counters are correct
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        if !~Gia_ObjValue(pObj) {
            continue; }
        pObjLog = Frc_ManObj( p, Gia_ObjValue(pObj) );
        assert(pObjLog.nFanins  == pObjLog.iFanin, "pObjLog.nFanins  == pObjLog.iFanin");
        assert(pObjLog.nFanouts == pObjLog.iFanout, "pObjLog.nFanouts == pObjLog.iFanout");
        pObjLog.iFanin = pObjLog.iFanout = 0;
    }
    // inlining of macro ABC_FREE
    ((pGia.pRefs) ? (free((char *) (pGia.pRefs)), (pGia.pRefs) = 0) : 0);
    return p;
}

/**Function*************************************************************

  Synopsis    [Collect the fanin IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManCollectSuper_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t, vVisit *Vec_Int_t) {
    if pObj.fMark1 {
        return; }
    pObj.fMark1 = 1;
    Vec_IntPush( vVisit, Gia_ObjId(p, pObj) );
    if pObj.fMark0 {
        Vec_IntPush( vSuper, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Frc_ManCollectSuper_rec( p, Gia_ObjFanin0(pObj), vSuper, vVisit );
    Frc_ManCollectSuper_rec( p, Gia_ObjFanin1(pObj), vSuper, vVisit );
    
}

/**Function*************************************************************

  Synopsis    [Collect the fanin IDs.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Frc_ManCollectSuper(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t, vVisit *Vec_Int_t) {
    var Entry, i int32;
    Vec_IntClear( vSuper );
    Vec_IntClear( vVisit );
    assert(pObj.fMark0 == 1, "pObj.fMark0 == 1");
    pObj.fMark0 = 0;
    Frc_ManCollectSuper_rec( p, pObj, vSuper, vVisit );
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
func Frc_ManCreateRefsSpecial(p *Gia_Man_t) {
    var pObj, pFan0, pFan1 *Gia_Obj_t;
    var pObjC, pObjD0, pObjD1 *Gia_Obj_t;
    var i int32;
    assert(p.pRefs == nil, "p.pRefs == nil");
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
func Frc_ManTransformRefs(p *Gia_Man_t, pnObjs *int32, pnFanios *int32) {
    var vSuper, vVisit *Vec_Int_t;
    var pObj, pFanin *Gia_Obj_t;
    var i, k, Counter int32;
    assert(p.pRefs != nil, "p.pRefs != nil");

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
    ((p.pRefs) ? (free((char *) (p.pRefs)), (p.pRefs) = 0) : 0);
    // inlining of macro ABC_CALLOC
    p.pRefs = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    // reference from internal nodes
    Counter = 0;
    vSuper = Vec_IntAlloc( 100 );
    vVisit = Vec_IntAlloc( 100 );
    Gia_ManCleanMark1( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if pObj.fMark0 == 0 {
            continue; }
        Frc_ManCollectSuper( p, pObj, vSuper, vVisit );
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

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManStart(pGia *Gia_Man_t) *Frc_Man_t {
    
    var pObjLog, pFanLog *Frc_Obj_t;
    var pObj, pFanin *Gia_Obj_t;//, * pObjRi, * pObjRo;
    var vSuper, vVisit *Vec_Int_t;
    var nObjs, nFanios, nNodes int32;
    var i, k, hHandle int32;
    // prepare the AIG
//    Gia_ManCreateRefs( pGia );
    Frc_ManCreateRefsSpecial( pGia );
    Frc_ManTransformRefs( pGia, &nObjs, &nFanios );
    Gia_ManFillValue( pGia );
    // create logic network
    // inlining of macro ABC_CALLOC
    p := &Frc_Man_t{};
    p.pGia  = pGia;
    p.nRegs = Gia_ManRegNum(pGia);
    p.vCis  = Vec_IntAlloc( Gia_ManCiNum(pGia) );
    p.vCos  = Vec_IntAlloc( Gia_ManCoNum(pGia) );
    p.nObjData = (sizeof(Frc_Obj_t) / 4) * nObjs + 2 * nFanios;
    // inlining of macro ABC_CALLOC
    p.pObjData = ((int *) calloc((size_t)(p.nObjData), sizeof(int)));
    // create constant node
    Gia_ManConst0(pGia).Value = hHandle;
    pObjLog = Frc_ManObj( p, hHandle );
    pObjLog.hHandle  = hHandle;
    pObjLog.nFanins  = 0;
    pObjLog.nFanouts = Gia_ObjRefNum( pGia, Gia_ManConst0(pGia) );
    // count objects
    hHandle += Frc_ObjSize( pObjLog );
    nNodes++;
    p.nObjs++;
    // create the PIs
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
        // create PI object
        pObj.Value = hHandle;
        Vec_IntPush( p.vCis, hHandle );
        pObjLog = Frc_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = 0;
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        pObjLog.fCi = 1;
        // count objects
        hHandle += Frc_ObjSize( pObjLog );
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
        Frc_ManCollectSuper( pGia, pObj, vSuper, vVisit );
        // create node object
        pObj.Value = hHandle;
        pObjLog = Frc_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = Vec_IntSize( vSuper );
        pObjLog.nFanouts = Gia_ObjRefNum( pGia, pObj );
        // add fanins
        // inlining of macro Gia_ManForEachObjVec
        for  k = 0; (k < Vec_IntSize(vSuper)) && ((pFanin) = Gia_ManObj(pGia, Vec_IntEntry(vSuper,k))); k++  {
            pFanLog = Frc_ManObj( p, Gia_ObjValue(pFanin) ); 
            Frc_ObjAddFanin( pObjLog, pFanLog );
        }
        // count objects
        hHandle += Frc_ObjSize( pObjLog );
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
        pObjLog = Frc_ManObj( p, hHandle );
        pObjLog.hHandle  = hHandle;
        pObjLog.nFanins  = 1;
        pObjLog.nFanouts = 0;
        pObjLog.fCo = 1;
        // add fanins
        pFanLog = Frc_ManObj( p, Gia_ObjValue(Gia_ObjFanin0(pObj)) );
        Frc_ObjAddFanin( pObjLog, pFanLog );
        // count objects
        hHandle += Frc_ObjSize( pObjLog );
        p.nObjs++;
    }
    // connect registers
//    Gia_ManForEachRiRo( pGia, pObjRi, pObjRo, i )
//        Frc_ObjAddFanin( Frc_ManObj(p,Gia_ObjValue(pObjRo)), Frc_ManObj(p,Gia_ObjValue(pObjRi)) );
    Gia_ManCleanMark0( pGia );
    assert(nNodes  == Frc_ManNodeNum(p), "nNodes  == Frc_ManNodeNum(p)");
    assert(nObjs   == p.nObjs, "nObjs   == p.nObjs");
    assert(hHandle == p.nObjData, "hHandle == p.nObjData");
    if hHandle != p.nObjData {
        fmt.Printf("Frc_ManStart(): Fatal error in internal representation.\n"); }
    // make sure the fanin/fanout counters are correct
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        if !~Gia_ObjValue(pObj) {
            continue; }
        pObjLog = Frc_ManObj( p, Gia_ObjValue(pObj) );
        assert(pObjLog.nFanins  == pObjLog.iFanin, "pObjLog.nFanins  == pObjLog.iFanin");
        assert(pObjLog.nFanouts == pObjLog.iFanout, "pObjLog.nFanouts == pObjLog.iFanout");
        pObjLog.iFanin = pObjLog.iFanout = 0;
    }
    // inlining of macro ABC_FREE
    ((pGia.pRefs) ? (free((char *) (pGia.pRefs)), (pGia.pRefs) = 0) : 0);
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManPrintStats(p *Frc_Man_t) {
//    if ( p->pName )
//        printf( "%8s : ", p->pName );
    fmt.Printf("i/o =%7d/%7d  ", Frc_ManPiNum(p), Frc_ManPoNum(p));
    if Frc_ManRegNum(p) != 0 {
        fmt.Printf("ff =%7d  ", Frc_ManRegNum(p)); }
    fmt.Printf("node =%8d  ", Frc_ManNodeNum(p));
    fmt.Printf("obj =%8d  ", Frc_ManObjNum(p));
//    printf( "lev =%5d  ", Frc_ManLevelNum(p) );
//    printf( "cut =%5d  ", Frc_ManCrossCut(p) );
    fmt.Printf("mem =%5.2f MB", 4.0*p.nObjData/(1<<20));
//    printf( "obj =%5d  ", Frc_ManObjNum(p) );
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Creates logic network isomorphic to the given AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManStop(p *Frc_Man_t) {
    Vec_IntFree( p.vCis );
    Vec_IntFree( p.vCos );
    // inlining of macro ABC_FREE
    ((p.pObjData) ? (free((char *) (p.pObjData)), (p.pObjData) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Computes cross cut size for the given order of POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManCrossCut_rec(p *Frc_Man_t, pObj *Frc_Obj_t) int32 {
    assert(pObj.iFanout > 0, "pObj.iFanout > 0");
    if pObj.iFanout-- == pObj.nFanouts {
        var pFanin *Frc_Obj_t;
        var i int32;
        p.nCutCur++;
        p.nCutMax = Abc_MaxInt( p.nCutMax, p.nCutCur );
        if !(Frc_ObjIsCi(pObj) != 0) {
            // inlining of macro Frc_ObjForEachFanin
            for  i = 0; (i < int32(pObj.nFanins)) && (pFanin = Frc_ObjFanin(pObj,i)); i++  {
                p.nCutCur -= Frc_ManCrossCut_rec( p, pFanin );
            } }
    }
    return pObj.iFanout == 0;
}

/**Function*************************************************************

  Synopsis    [Computes cross cut size for the given order of POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManCrossCut2_rec(p *Frc_Man_t, pObj *Frc_Obj_t) int32 {
    assert(pObj.iFanout > 0, "pObj.iFanout > 0");
    if pObj.iFanout-- == pObj.nFanouts {
        var pFanin *Frc_Obj_t;
        var i int32;
        p.nCutCur++;
        p.nCutMax = Abc_MaxInt( p.nCutMax, p.nCutCur );
        if !(Frc_ObjIsCi(pObj) != 0) {
            // inlining of macro Frc_ObjForEachFaninReverse
            for  i = int32(pObj.nFanins) - 1; (i >= 0) && (pFanin = Frc_ObjFanin(pObj,i)); i--  {
                p.nCutCur -= Frc_ManCrossCut2_rec( p, pFanin );
            } }
    }
    return pObj.iFanout == 0;
}

/**Function*************************************************************

  Synopsis    [Computes cross cut size for the given order of POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManCrossCut(p *Frc_Man_t, vOrder *Vec_Int_t, fReverse int32) int32 {
    var pObj *Frc_Obj_t;
    var i int32;
    assert(Vec_IntSize(vOrder) == Frc_ManCoNum(p), "Vec_IntSize(vOrder) == Frc_ManCoNum(p)");
    p.nCutCur = 0;
    p.nCutMax = 0;
    // inlining of macro Frc_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Frc_ManObj(p,i)); i += Frc_ObjSize(pObj)  {
        pObj.iFanout = pObj.nFanouts;
    }
    // inlining of macro Frc_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vOrder)) && ((pObj) = Frc_ManObj(p, Vec_IntEntry(vOrder,i))); i++  {
        assert((Frc_ObjIsCo(pObj) != 0), "Frc_ObjIsCo(pObj)");
        if fReverse != 0 {
            p.nCutCur -= Frc_ManCrossCut2_rec( p, Frc_ObjFanin(pObj,0) );
        } else {
            p.nCutCur -= Frc_ManCrossCut_rec( p, Frc_ObjFanin(pObj,0) ); }
    }
    assert(p.nCutCur == 0, "p.nCutCur == 0");
//    Frc_ManForEachObj( p, pObj, i )
//        assert( pObj->iFanout == 0 );
    return p.nCutMax;
}

/**Function*************************************************************

  Synopsis    [Collects CO handles.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManCollectCos(p *Frc_Man_t) *Vec_Int_t {
    var vCoOrder *Vec_Int_t;
    var pObj *Frc_Obj_t;
    var i int32;
    vCoOrder = Vec_IntAlloc( Frc_ManCoNum(p) );
    // inlining of macro Frc_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (pObj = Frc_ManObj(p,Vec_IntEntry(p.vCos,i))); i++  {
        Vec_IntPush( vCoOrder, pObj.hHandle );
    }
    return vCoOrder;
}

/**Function*************************************************************

  Synopsis    [Computes cross cut size for the given order of POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManCrossCutTest(p *Frc_Man_t, vOrderInit *Vec_Int_t) {
    var vOrder *Vec_Int_t;
//    abctime clk = Abc_Clock();
    if vOrderInit != nil {
        vOrder = vOrderInit;
    } else {
        vOrder = Frc_ManCollectCos( p );
    }
    fmt.Printf("CrossCut = %6d\n", Frc_ManCrossCut( p, vOrder, 0 ));
    fmt.Printf("CrossCut = %6d\n", Frc_ManCrossCut( p, vOrder, 1 ));
    Vec_IntReverseOrder( vOrder );
    fmt.Printf("CrossCut = %6d\n", Frc_ManCrossCut( p, vOrder, 0 ));
    fmt.Printf("CrossCut = %6d\n", Frc_ManCrossCut( p, vOrder, 1 ));
    Vec_IntReverseOrder( vOrder );
    if vOrder != vOrderInit {
        Vec_IntFree( vOrder ); }
//    ABC_PRT( "Time", Abc_Clock() - clk );
}



/**Function*************************************************************

  Synopsis    [Generates random placement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManPlaceRandom(p *Frc_Man_t) {
    var pThis *Frc_Obj_t;
    var pPlacement *int32;
    var i, h, Temp, iNext, Counter int32;
    // inlining of macro ABC_ALLOC
    pPlacement = ((int *) malloc(sizeof(int) * (size_t)(p.nObjs)));
    for  i = 0; i < p.nObjs; i++  {
        pPlacement[i] = i;
    }
    for  i = 0; i < p.nObjs; i++  {
        iNext = Gia_ManRandom( 0 ) % p.nObjs;
        Temp = pPlacement[i];
        pPlacement[i] = pPlacement[iNext];
        pPlacement[iNext] = Temp;
    }
    Counter = 0;
    // inlining of macro Frc_ManForEachObj
    for  h = 0; (h < p.nObjData) && (pThis = Frc_ManObj(p,h)); h += Frc_ObjSize(pThis)  {
        pThis.pPlace = pPlacement[Counter]; Counter++;
    }
    // inlining of macro ABC_FREE
    ((pPlacement) ? (free((char *) (pPlacement)), (pPlacement) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Shuffles array of random integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManArrayShuffle(vArray *Vec_Int_t) {
    var i, iNext, Temp int32;
    for  i = 0; i < vArray.nSize; i++  {
        iNext = Gia_ManRandom( 0 ) % vArray.nSize;
        Temp = vArray.pArray[i];
        vArray.pArray[i] = vArray.pArray[iNext];
        vArray.pArray[iNext] = Temp;
    }
}

/**Function*************************************************************

  Synopsis    [Computes cross cut size for the given order of POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManPlaceDfs_rec(p *Frc_Man_t, pObj *Frc_Obj_t, piPlace *int32) {
    assert(pObj.iFanout > 0, "pObj.iFanout > 0");
    if pObj.iFanout-- == pObj.nFanouts {
        var pFanin *Frc_Obj_t;
        var i int32;
        if !(Frc_ObjIsCi(pObj) != 0) {
            // inlining of macro Frc_ObjForEachFanin
            for  i = 0; (i < int32(pObj.nFanins)) && (pFanin = Frc_ObjFanin(pObj,i)); i++  {
                Frc_ManPlaceDfs_rec( p, pFanin, piPlace );
            } }
        pObj.pPlace = (*piPlace)++;
    }
}

/**Function*************************************************************

  Synopsis    [Generates DFS placement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManPlaceDfs(p *Frc_Man_t, vCoOrder *Vec_Int_t) {
    var pObj *Frc_Obj_t;
    var i, nPlaces int32;
    // inlining of macro Frc_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Frc_ManObj(p,i)); i += Frc_ObjSize(pObj)  {
        pObj.iFanout = pObj.nFanouts;
        if pObj.nFanouts == 0 && !(Frc_ObjIsCo(pObj) != 0) {
            pObj.pPlace = nPlaces++; }
    }
    // inlining of macro Frc_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCoOrder)) && ((pObj) = Frc_ManObj(p, Vec_IntEntry(vCoOrder,i))); i++  {
        assert((Frc_ObjIsCo(pObj) != 0), "Frc_ObjIsCo(pObj)");
        Frc_ManPlaceDfs_rec( p, Frc_ObjFanin(pObj,0), &nPlaces );
        pObj.pPlace = nPlaces++;
    }
    assert(nPlaces == p.nObjs, "nPlaces == p.nObjs");
}

/**Function*************************************************************

  Synopsis    [Generates DFS placement by trying both orders.]

  Description [Returns the cross cut size of the best order. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManPlaceDfsBoth(p *Frc_Man_t, vCoOrder *Vec_Int_t, piCutSize2 *int32) int32 {
    var nCutStart1, nCutStart2 int32;
    nCutStart1 = Frc_ManCrossCut( p, vCoOrder, 0 );
    Vec_IntReverseOrder( vCoOrder );
    nCutStart2 = Frc_ManCrossCut( p, vCoOrder, 0 );
    if nCutStart1 <= nCutStart2 {
        Vec_IntReverseOrder( vCoOrder ); // undo
        Frc_ManPlaceDfs( p, vCoOrder );
        *piCutSize2 = nCutStart2;
        return nCutStart1;
    
    } else {
        Frc_ManPlaceDfs( p, vCoOrder );
        Vec_IntReverseOrder( vCoOrder ); // undo
        *piCutSize2 = nCutStart1;
        return nCutStart2;
    }
}

/**Function*************************************************************

  Synopsis    [Performs iterative refinement of the given placement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ManPlacementRefine(p *Frc_Man_t, nIters int32, fVerbose int32) {
    var fRandomize int32;
    var vCoOrder *Vec_Int_t;
    var pThis, pNext *Frc_Obj_t;
    var CostThis, CostPrev float64;
    var pVertX *float32;
    var VertX float32;
    var pPermX, pHandles *int32;
    var k, h, Iter, iMinX, iMaxX, Counter, nCutStart, nCutCur, nCutCur2, nCutPrev int32;
    var clk abctime = Abc_Clock();
    var clk2 abctime;
    var clk2Total abctime;
    // create starting one-dimensional placement
    vCoOrder = Frc_ManCollectCos( p );
    if fRandomize != 0 {
        Frc_ManArrayShuffle( vCoOrder ); }
    nCutStart = Frc_ManPlaceDfsBoth( p, vCoOrder, &nCutCur2 );
    // refine placement
    CostPrev = 0.0;
    nCutPrev = nCutStart;
    // inlining of macro ABC_ALLOC
    pHandles = ((int *) malloc(sizeof(int) * (size_t)(p.nObjs)));
    // inlining of macro ABC_ALLOC
    pVertX   = ((float *) malloc(sizeof(float) * (size_t)(p.nObjs)));
    for  Iter = 0; Iter < nIters; Iter++  {
        // compute centers of hyperedges
        CostThis = 0.0;
        // inlining of macro Frc_ManForEachObj
        for  h = 0; (h < p.nObjData) && (pThis = Frc_ManObj(p,h)); h += Frc_ObjSize(pThis)  {
            iMinX = iMaxX = pThis.pPlace;
            // inlining of macro Frc_ObjForEachFanout
            for  k = 0; (k < int32(pThis.nFanouts)) && (pNext = Frc_ObjFanout(pThis,k)); k++  {
                iMinX = Abc_MinInt( iMinX, pNext.pPlace );
                iMaxX = Abc_MaxInt( iMaxX, pNext.pPlace );
            }
            pThis.fEdgeCenter = 0.5 * (iMaxX + iMinX);
            CostThis += (iMaxX - iMinX);
        }
        // compute new centers of objects
        Counter = 0;
        // inlining of macro Frc_ManForEachObj
        for  h = 0; (h < p.nObjData) && (pThis = Frc_ManObj(p,h)); h += Frc_ObjSize(pThis)  {
            VertX = pThis.fEdgeCenter;
            // inlining of macro Frc_ObjForEachFanin
            for  k = 0; (k < int32(pThis.nFanins)) && (pNext = Frc_ObjFanin(pThis,k)); k++  {
                VertX += pNext.fEdgeCenter;
            }
            pVertX[Counter] = VertX / (Frc_ObjFaninNum(pThis) + 1);
            pHandles[Counter++] = h;
        }
        assert(Counter == Frc_ManObjNum(p), "Counter == Frc_ManObjNum(p)");
        // sort these numbers
        clk2 = Abc_Clock();
        pPermX = Gia_SortFloats( pVertX, pHandles, p.nObjs );
        clk2Total += Abc_Clock() - clk2;
        assert(pPermX == pHandles, "pPermX == pHandles");
        Vec_IntClear( vCoOrder );
        for  k = 0; k < p.nObjs; k++  {
            pThis = Frc_ManObj( p, pPermX[k] );
            pThis.pPlace = k;
            if Frc_ObjIsCo(pThis) != 0 {
                Vec_IntPush( vCoOrder, pThis.hHandle ); }
        }
/*
        printf( "Ordering of PIs:\n" );
        Frc_ManForEachCi( p, pThis, k )
            printf( "PI number = %7d.  Object handle = %7d,  Coordinate = %7d.\n", 
                k, pThis->hHandle, pThis->pPlace );
*/
        nCutCur = Frc_ManPlaceDfsBoth( p, vCoOrder, &nCutCur2 );
        // evaluate cost
        if fVerbose != 0 {
            fmt.Printf("%2d : Span = %e  ", Iter+1, CostThis);
            fmt.Printf("Cut = %6d  (%5.2f %%)  CutR = %6d  ", nCutCur, 100.0*(nCutStart-nCutCur)/nCutStart, nCutCur2);
            // inlining of macro ABC_PRTn
            Abc_Print(1, "%s =", ("Total"));
Abc_Print(1, "%9.2f sec  ", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Sort"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((clk2Total)))/(float64(int64(time.Second))));;
//        Frc_ManCrossCutTest( p, vCoOrder );
        }
//        if ( 1.0 * nCutPrev / nCutCur < 1.001 )
//            break;
        nCutPrev = nCutCur;
    }
    // inlining of macro ABC_FREE
    ((pHandles) ? (free((char *) (pHandles)), (pHandles) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pVertX) ? (free((char *) (pVertX)), (pVertX) = 0) : 0);
    Vec_IntFree( vCoOrder );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if all fanouts are COsw.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_ObjFanoutsAreCos(pThis *Frc_Obj_t) int32 {
    var pNext *Frc_Obj_t;
    var i int32;
    // inlining of macro Frc_ObjForEachFanout
    for  i = 0; (i < int32(pThis.nFanouts)) && (pNext = Frc_ObjFanout(pThis,i)); i++  {
        if !(Frc_ObjIsCo(pNext) != 0) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes the distances from the given set of objects.]

  Description [Returns one of the most distant objects.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Frc_DumpGraphIntoFile(p *Frc_Man_t) {
    var pFile *os.File;
    var pThis, pNext *Frc_Obj_t;
    var i, k, Counter int32;
    // assign numbers to CIs and internal nodes
    // inlining of macro Frc_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pThis = Frc_ManObj(p,i)); i += Frc_ObjSize(pThis)  {
        if (i != 0) && (((Frc_ObjIsCi(pThis) != 0) && !(Frc_ObjFanoutsAreCos(pThis) != 0)) || (Frc_ObjIsNode(pThis) != 0)) {
            pThis.iFanin = Counter++;
        } else {
            pThis.iFanin = ~0; }
    }
    // assign numbers to all other nodes
    pFile = os.OpenFile("x\\large\\aig\\dg1.g", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    // inlining of macro Frc_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pThis = Frc_ManObj(p,i)); i += Frc_ObjSize(pThis)  {
        // inlining of macro Frc_ObjForEachFanout
        for  k = 0; (k < int32(pThis.nFanouts)) && (pNext = Frc_ObjFanout(pThis,k)); k++  {
            if ~pThis.iFanin && ~pNext.iFanin {
                fmt.Fprintf(pFile, "%d %d\n", pThis.iFanin, pNext.iFanin); }
        }
    }
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Experiment with the FORCE algorithm.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func For_ManExperiment(pGia *Gia_Man_t, nIters int32, fClustered int32, fVerbose int32) {
    var p *Frc_Man_t;
    Gia_ManRandom( 1 );
    if fClustered != 0 {
        p = Frc_ManStart( pGia );
    } else {
        p = Frc_ManStartSimple( pGia ); }
//    Frc_DumpGraphIntoFile( p );
    if fVerbose != 0 {
        Frc_ManPrintStats( p ); }
//    Frc_ManCrossCutTest( p, NULL );
    Frc_ManPlacementRefine( p, nIters, fVerbose );
    Frc_ManStop( p );
}

/**Function*************************************************************

  Synopsis    [Experiment with the FORCE algorithm.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func For_ManFileExperiment() {
    var pFile *os.File;
    var pBuffer *int32;
    var i int32;
    var Size int32;
    var Exp int32 = 25;
    var clk abctime  = Abc_Clock();
    var RetValue int32;

    Size = (1 << Exp);
    fmt.Printf("2^%d machine words (%d bytes).\n", Exp, int32(sizeof(int)) * Size);

    // inlining of macro ABC_ALLOC
    pBuffer = ((int *) malloc(sizeof(int) * (size_t)(Size)));
    for  i = 0; i < Size; i++  {
        pBuffer[i] = i;
    }
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Fillup"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;

clk = Abc_Clock();
    pFile = os.Open("test.txt");
    RetValue = fread( pBuffer, 1, sizeof(int) * Size, pFile );
    _ = (pFile).Close();
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Read  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;

clk = Abc_Clock();
    pFile = os.OpenFile("test.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fwrite( pBuffer, 1, sizeof(int) * Size, pFile );
    _ = (pFile).Close();
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Write "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
/*
2^25 machine words (134217728 bytes).
Fillup =    0.06 sec
Read   =    0.08 sec
Write  =    1.81 sec
*/
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

