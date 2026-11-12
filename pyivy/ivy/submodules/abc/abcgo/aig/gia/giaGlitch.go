package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaGlitch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Glitch simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaGlitch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Gli_Obj_t struct {
    fTerm uint32; /* C bitfield: 1 bits; review packing and truncation */    // terminal node
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */    // value under 000 pattern
    fPhase2 uint32; /* C bitfield: 1 bits; review packing and truncation */    // value under 000 pattern
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */    // user-controlled mark
    nFanins uint32; /* C bitfield: 3 bits; review packing and truncation */    // the number of fanins
    nFanouts uint32; /* C bitfield: 25 bits; review packing and truncation */    // total number of fanouts
    Handle       uint32;           // ID of the node
    pTruth          *word;           // truth table of the node
    uSimInfo       uint32;         // simulation info of the node
    union 
    {
        int        iFanin;           // the number of fanins added
        int        nSwitches;        // the number of switches
    };
    union 
    {
        int        iFanout;          // the number of fanouts added
        int        nGlitches;        // the number of glitches ( nGlitches >= nSwitches )
    };
    Fanios [0]int32;        // the array of fanins/fanouts
}

   
type Gli_Man_t struct {
    vCis     *Vec_Int_t;             // the vector of CIs (PIs + LOs)
    vCos     *Vec_Int_t;             // the vector of COs (POs + LIs)
    vCisChanged     *Vec_Int_t;      // the changed CIs
    vAffected     *Vec_Int_t;        // the affected nodes
    vFrontier     *Vec_Int_t;        // the fanouts of these nodes
    nObjs            int32;            // the number of objects
    nRegs            int32;            // the number of registers
    nTravIds            int32;         // traversal ID of the network
    iObjData            int32;         // pointer to the current data
    nObjData            int32;         // the size of array to store the logic network
    pObjData           *int32;         // the internal nodes
    pSimInfoPrev      *uint32;     // previous values of the CIs
}


func Gli_ManCiNum(p *Gli_Man_t) int32 { return Vec_IntSize(p.vCis);                                         }
func Gli_ManCoNum(p *Gli_Man_t) int32 { return Vec_IntSize(p.vCos);                                         }
func Gli_ManPiNum(p *Gli_Man_t) int32 { return Vec_IntSize(p.vCis) - p.nRegs;                              }
func Gli_ManPoNum(p *Gli_Man_t) int32 { return Vec_IntSize(p.vCos) - p.nRegs;                              }
func Gli_ManRegNum(p *Gli_Man_t) int32 { return p.nRegs;                                                     }
func Gli_ManObjNum(p *Gli_Man_t) int32 { return p.nObjs;                                                     } 
func Gli_ManNodeNum(p *Gli_Man_t) int32 { return p.nObjs - Vec_IntSize(p.vCis) - Vec_IntSize(p.vCos);       } 

func Gli_ManObj(p *Gli_Man_t, v int32) *Gli_Obj_t { return (Gli_Obj_t *)(p.pObjData + v);                               } 
func Gli_ManCi(p *Gli_Man_t, v int32) *Gli_Obj_t { return Gli_ManObj( p, Vec_IntEntry(p.vCis,v) );                     }
func Gli_ManCo(p *Gli_Man_t, v int32) *Gli_Obj_t { return Gli_ManObj( p, Vec_IntEntry(p.vCos,v) );                     }
func Gli_ManPi(p *Gli_Man_t, v int32) *Gli_Obj_t { assert(v < Gli_ManPiNum(p), "v < Gli_ManPiNum(p)");  return Gli_ManCi( p, v );            }
func Gli_ManPo(p *Gli_Man_t, v int32) *Gli_Obj_t { assert(v < Gli_ManPoNum(p), "v < Gli_ManPoNum(p)");  return Gli_ManCo( p, v );            }
func Gli_ManRo(p *Gli_Man_t, v int32) *Gli_Obj_t { assert(v < Gli_ManRegNum(p), "v < Gli_ManRegNum(p)"); return Gli_ManCi( p, Gli_ManRegNum(p)+v );      }
func Gli_ManRi(p *Gli_Man_t, v int32) *Gli_Obj_t { assert(v < Gli_ManRegNum(p), "v < Gli_ManRegNum(p)"); return Gli_ManCo( p, Gli_ManRegNum(p)+v );      }

func Gli_ObjIsTerm(pObj *Gli_Obj_t) int32 { return pObj.fTerm;                                                  } 
func Gli_ObjIsCi(pObj *Gli_Obj_t) int32 { return pObj.fTerm && pObj.nFanins == 0;                            } 
func Gli_ObjIsCo(pObj *Gli_Obj_t) int32 { return pObj.fTerm && pObj.nFanins == 1;                            } 
func Gli_ObjIsNode(pObj *Gli_Obj_t) int32 { return!pObj.fTerm;                                                  } 

func Gli_ObjFaninNum(pObj *Gli_Obj_t) int32 { return pObj.nFanins;                                                } 
func Gli_ObjFanoutNum(pObj *Gli_Obj_t) int32 { return pObj.nFanouts;                                               } 
func Gli_ObjSize(pObj *Gli_Obj_t) int32 { return sizeof(Gli_Obj_t) / 4 + pObj.nFanins + pObj.nFanouts;       } 

func Gli_ObjFanin(pObj *Gli_Obj_t, i int32) *Gli_Obj_t { return (Gli_Obj_t *)(((int *)pObj) - pObj.Fanios[i]);               } 
func Gli_ObjFanout(pObj *Gli_Obj_t, i int32) *Gli_Obj_t { return (Gli_Obj_t *)(((int *)pObj) + pObj.Fanios[pObj.nFanins+i]); } 

// Original macro:
// #define Gli_ManForEachObj( p, pObj, i )                 \
//     for ( i = 0; (i < p->nObjData) && (pObj = Gli_ManObj(p,i)); i += Gli_ObjSize(pObj) )

// Original macro:
// #define Gli_ManForEachNode( p, pObj, i )                \
//     for ( i = 0; (i < p->nObjData) && (pObj = Gli_ManObj(p,i)); i += Gli_ObjSize(pObj) ) if ( Gli_ObjIsTerm(pObj) ) {} else


// Original macro:
// #define Gli_ManForEachEntry( vVec, p, pObj, i )         \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && (pObj = Gli_ManObj(p,Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Gli_ManForEachCi( p, pObj, i )                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCis)) && (pObj = Gli_ManObj(p,Vec_IntEntry(p->vCis,i))); i++ )

// Original macro:
// #define Gli_ManForEachCo( p, pObj, i )                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && (pObj = Gli_ManObj(p,Vec_IntEntry(p->vCos,i))); i++ )


// Original macro:
// #define Gli_ManForEachPi( p, pObj, i )                                  \
//     for ( i = 0; (i < Gli_ManPiNum(p)) && ((pObj) = Gli_ManCi(p, i)); i++ )

// Original macro:
// #define Gli_ManForEachPo( p, pObj, i )                                  \
//     for ( i = 0; (i < Gli_ManPoNum(p)) && ((pObj) = Gli_ManCo(p, i)); i++ )

// Original macro:
// #define Gli_ManForEachRo( p, pObj, i )                                  \
//     for ( i = 0; (i < Gli_ManRegNum(p)) && ((pObj) = Gli_ManCi(p, Gli_ManPiNum(p)+i)); i++ )

// Original macro:
// #define Gli_ManForEachRi( p, pObj, i )                                  \
//     for ( i = 0; (i < Gli_ManRegNum(p)) && ((pObj) = Gli_ManCo(p, Gli_ManPoNum(p)+i)); i++ )

// Original macro:
// #define Gli_ManForEachRiRo( p, pObjRi, pObjRo, i )                      \
//     for ( i = 0; (i < Gli_ManRegNum(p)) && ((pObjRi) = Gli_ManCo(p, Gli_ManPoNum(p)+i)) && ((pObjRo) = Gli_ManCi(p, Gli_ManPiNum(p)+i)); i++ )


// Original macro:
// #define Gli_ObjForEachFanin( pObj, pNext, i )           \
//     for ( i = 0; (i < (int)pObj->nFanins) && (pNext = Gli_ObjFanin(pObj,i)); i++ )

// Original macro:
// #define Gli_ObjForEachFanout( pObj, pNext, i )          \
//     for ( i = 0; (i < (int)pObj->nFanouts) && (pNext = Gli_ObjFanout(pObj,i)); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates logic network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManAlloc(nObjs int32, nRegs int32, nFanioPairs int32) *Gli_Man_t {
    var p *Gli_Man_t;
    // inlining of macro ABC_CALLOC
    p = (Gli_Man_t *)((int *) calloc((size_t)((sizeof(Gli_Man_t) / 4) + (sizeof(Gli_Obj_t) / 4) * nObjs + 2 * nFanioPairs), sizeof(int)));
    p.nRegs = nRegs;
    p.vCis = Vec_IntAlloc( 1000 );
    p.vCos = Vec_IntAlloc( 1000 );
    p.vCisChanged = Vec_IntAlloc( 1000 );
    p.vAffected = Vec_IntAlloc( 1000 );
    p.vFrontier = Vec_IntAlloc( 1000 );
    p.nObjData = (sizeof(Gli_Obj_t) / 4) * nObjs + 2 * nFanioPairs;
    p.pObjData = (int *)(p + 1);
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes logic network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManStop(p *Gli_Man_t) {
    Vec_IntFree( p.vCis );
    Vec_IntFree( p.vCos );
    Vec_IntFree( p.vCisChanged );
    Vec_IntFree( p.vAffected );
    Vec_IntFree( p.vFrontier );
    // inlining of macro ABC_FREE
    ((p.pSimInfoPrev) ? (free((char *) (p.pSimInfoPrev)), (p.pSimInfoPrev) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Checks logic network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManPrintObjects(p *Gli_Man_t) {
    var pObj, pNext *Gli_Obj_t;
    var i, k int32;
    // inlining of macro Gli_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Gli_ManObj(p,i)); i += Gli_ObjSize(pObj)  {
        fmt.Printf("Node %d \n", pObj.Handle);
        fmt.Printf("Fanins: ");
        // inlining of macro Gli_ObjForEachFanin
        for  k = 0; (k < int32(pObj.nFanins)) && (pNext = Gli_ObjFanin(pObj,k)); k++  {
            fmt.Printf("%d ", pNext.Handle);
        }
        fmt.Printf("\n");
        fmt.Printf("Fanouts: ");
        // inlining of macro Gli_ObjForEachFanout
        for  k = 0; (k < int32(pObj.nFanouts)) && (pNext = Gli_ObjFanout(pObj,k)); k++  {
            fmt.Printf("%d ", pNext.Handle);
        }
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Checks logic network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManFinalize(p *Gli_Man_t) {
    var pObj *Gli_Obj_t;
    var i int32;
    assert(p.iObjData == p.nObjData, "p.iObjData == p.nObjData");
    // inlining of macro Gli_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Gli_ManObj(p,i)); i += Gli_ObjSize(pObj)  {
        assert(pObj.iFanin == int32(pObj.nFanins), "pObj.iFanin == int32(pObj.nFanins)");
        assert(pObj.iFanout == int32(pObj.nFanouts), "pObj.iFanout == int32(pObj.nFanouts)");
        pObj.iFanin = 0;
        pObj.iFanout = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Creates fanin/fanout pair.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ObjAddFanin(pObj *Gli_Obj_t, pFanin *Gli_Obj_t) { 
    assert(pObj.iFanin < int32(pObj.nFanins), "pObj.iFanin < int32(pObj.nFanins)");
    assert(pFanin.iFanout < int32(pFanin.nFanouts), "pFanin.iFanout < int32(pFanin.nFanouts)");
    pFanin.Fanios[pFanin.nFanins + pFanin.iFanout++] = 
    pObj.Fanios[pObj.iFanin++] = pObj.Handle - pFanin.Handle;
}
 
/**Function*************************************************************

  Synopsis    [Allocates object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ObjAlloc(p *Gli_Man_t, nFanins int32, nFanouts int32) *Gli_Obj_t {
    var pObj *Gli_Obj_t;
    pObj = Gli_ManObj( p, p.iObjData );
    pObj.Handle   = p.iObjData;
    pObj.nFanins  = nFanins;
    pObj.nFanouts = nFanouts;
    p.iObjData += Gli_ObjSize( pObj );
    p.nObjs++;
    return pObj;
}

/**Function*************************************************************

  Synopsis    [Creates CI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManCreateCi(p *Gli_Man_t, nFanouts int32) int32 {
    var pObj *Gli_Obj_t;
    pObj = Gli_ObjAlloc( p, 0, nFanouts );
    pObj.fTerm = 1;
    Vec_IntPush( p.vCis, pObj.Handle );
    return pObj.Handle;
}

/**Function*************************************************************

  Synopsis    [Creates CO.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManCreateCo(p *Gli_Man_t, iFanin int32) int32 {
    var pObj, pFanin *Gli_Obj_t;
    pObj = Gli_ObjAlloc( p, 1, 0 );
    pObj.fTerm = 1;
    pFanin = Gli_ManObj( p, iFanin );
    Gli_ObjAddFanin( pObj, pFanin );
    pObj.fPhase = pObj.fPhase2 = pFanin.fPhase;
    Vec_IntPush( p.vCos, pObj.Handle );
    return pObj.Handle;
}

/**Function*************************************************************

  Synopsis    [Creates node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_NodeComputeValue(pNode *Gli_Obj_t) int32 {
    var i, Phase int32;
    for  i = 0; i < int32(pNode.nFanins); i++  {
        Phase |= (Gli_ObjFanin(pNode, i).fPhase << i);
    }
    return Abc_InfoHasBit( (unsigned *)pNode.pTruth, Phase );
}

/**Function*************************************************************

  Synopsis    [Creates node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_NodeComputeValue2(pNode *Gli_Obj_t) int32 {
    var i, Phase int32;
    for  i = 0; i < int32(pNode.nFanins); i++  {
        Phase |= (Gli_ObjFanin(pNode, i).fPhase2 << i);
    }
    return Abc_InfoHasBit( (unsigned *)pNode.pTruth, Phase );
}

/**Function*************************************************************

  Synopsis    [Creates node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManCreateNode(p *Gli_Man_t, vFanins *Vec_Int_t, nFanouts int32, pGateTruth *word) int32 {
    var pObj, pFanin *Gli_Obj_t;
    var i int32;
    assert(Vec_IntSize(vFanins) <= 16, "Vec_IntSize(vFanins) <= 16");
    pObj = Gli_ObjAlloc( p, Vec_IntSize(vFanins), nFanouts );
    // inlining of macro Gli_ManForEachEntry
    for  i = 0; (i < Vec_IntSize(vFanins)) && (pFanin = Gli_ManObj(p,Vec_IntEntry(vFanins,i))); i++  {
        Gli_ObjAddFanin( pObj, pFanin );
    }
    pObj.pTruth = pGateTruth;
    pObj.fPhase = pObj.fPhase2 = Gli_NodeComputeValue( pObj );
    return pObj.Handle;
}

/**Function*************************************************************

  Synopsis    [Returns the number of switches of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ObjNumSwitches(p *Gli_Man_t, iNode int32) int32 {
    return Gli_ManObj( p, iNode ).nSwitches;
}

/**Function*************************************************************

  Synopsis    [Returns the number of glitches of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ObjNumGlitches(p *Gli_Man_t, iNode int32) int32 {
    return Gli_ManObj( p, iNode ).nGlitches;
}


/**Function*************************************************************

  Synopsis    [Sets random info at the PIs and collects changed PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManSetPiRandom(p *Gli_Man_t, PiTransProb float32) {
    var pObj *Gli_Obj_t;
    var Multi float32  = 1.0 / (1 << 16);
    var i int32;
    assert(0.0 < PiTransProb && PiTransProb < 1.0, "0.0 < PiTransProb && PiTransProb < 1.0");
    Vec_IntClear( p.vCisChanged );
    // inlining of macro Gli_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && (pObj = Gli_ManObj(p,Vec_IntEntry(p.vCis,i))); i++  {
        if Multi * (Gia_ManRandom(0) & 0xffff) < PiTransProb {
            Vec_IntPush( p.vCisChanged, pObj.Handle );
            pObj.fPhase  ^= 1;
            pObj.fPhase2 ^= 1;
            pObj.nSwitches++;
            pObj.nGlitches++;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Sets random info at the PIs and collects changed PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManSetPiFromSaved(p *Gli_Man_t, iBit int32) {
    var pObj *Gli_Obj_t;
    var i int32;
    Vec_IntClear( p.vCisChanged );
    // inlining of macro Gli_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && (pObj = Gli_ManObj(p,Vec_IntEntry(p.vCis,i))); i++  {
        if (p.pSimInfoPrev[i] ^ pObj.uSimInfo) & (1 << iBit) {
            Vec_IntPush( p.vCisChanged, pObj.Handle );
            pObj.fPhase  ^= 1;
            pObj.fPhase2 ^= 1;
            pObj.nSwitches++;
            pObj.nGlitches++;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Computes switching activity of each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManSwitching(p *Gli_Man_t) {
    var pThis *Gli_Obj_t;
    var i int32;
    // inlining of macro Gli_ManForEachNode
    for  i = 0; (i < p.nObjData) && (pThis = Gli_ManObj(p,i)); i += Gli_ObjSize(pThis)  { if Gli_ObjIsTerm(pThis) != 0 {} else {
        if (int32(pThis.fPhase)) == Gli_NodeComputeValue(pThis) {
            continue; }
        pThis.fPhase ^= 1;
        pThis.nSwitches++;
    }
    }
}

/**Function*************************************************************

  Synopsis    [Computes glitching activity of each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManGlitching(p *Gli_Man_t) {
    var pThis, pFanout *Gli_Obj_t;//, * pOther = Gli_ManObj(p, 41);
    var i, k, Handle int32;
//    Gli_ManForEachObj( p, pThis, i )
//        assert( pThis->fMark == 0 );
    // start the array of affected nodes
    Vec_IntClear( p.vAffected );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCisChanged)) && (((Handle) = Vec_IntEntry(p.vCisChanged, i)), 1); i++  {
        Vec_IntPush( p.vAffected, Handle );
    }
    // iteration propagation
    for  Vec_IntSize(p.vAffected) > 0  {
        // compute the frontier
        Vec_IntClear( p.vFrontier );
        // inlining of macro Gli_ManForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vAffected)) && (pThis = Gli_ManObj(p,Vec_IntEntry(p.vAffected,i))); i++  {
            // inlining of macro Gli_ObjForEachFanout
            for  k = 0; (k < int32(pThis.nFanouts)) && (pFanout = Gli_ObjFanout(pThis,k)); k++  {
                if Gli_ObjIsCo(pFanout) != 0 {
                    continue; }
                if pFanout.fMark {
                    continue; }
                pFanout.fMark = 1;
                Vec_IntPush( p.vFrontier, pFanout.Handle );
            }
        }
        // compute the next set of affected nodes
        Vec_IntClear( p.vAffected );
        // inlining of macro Gli_ManForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vFrontier)) && (pThis = Gli_ManObj(p,Vec_IntEntry(p.vFrontier,i))); i++  {
            pThis.fMark = 0;
            if (int32(pThis.fPhase2)) == Gli_NodeComputeValue2(pThis) {
                continue; }
            pThis.fPhase2 ^= 1;
            pThis.nGlitches++;
            Vec_IntPush( p.vAffected, pThis.Handle );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Checks that the resulting values are the same.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManVerify(p *Gli_Man_t) {
    var pObj *Gli_Obj_t;
    var i int32;
    // inlining of macro Gli_ManForEachObj
    for  i = 0; (i < p.nObjData) && (pObj = Gli_ManObj(p,i)); i += Gli_ObjSize(pObj)  {
        assert(pObj.fPhase == pObj.fPhase2, "pObj.fPhase == pObj.fPhase2");
        assert(pObj.nGlitches >= pObj.nSwitches, "pObj.nGlitches >= pObj.nSwitches");
    }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManSimulateSeqNode(p *Gli_Man_t, pNode *Gli_Obj_t) uint32 {
    var pSimInfos [6]uint32;
    var Result uint32;
    var nFanins int32  = Gli_ObjFaninNum(pNode);
    var i, k, Phase int32;
    var pFanin *Gli_Obj_t;
    assert(nFanins <= 16, "nFanins <= 16");
    // inlining of macro Gli_ObjForEachFanin
    for  i = 0; (i < int32(pNode.nFanins)) && (pFanin = Gli_ObjFanin(pNode,i)); i++  {
        pSimInfos[i] = pFanin.uSimInfo;
    }
    for  i = 0; i < 32; i++  {
        Phase = 0;
        for  k = 0; k < nFanins; k++  {
            if (pSimInfos[k] >> i) & 1 {
                Phase |= (1 << k); }
        }
        if Abc_InfoHasBit( (unsigned *)pNode.pTruth, Phase ) {
            Result |= (1 << i); }
    }
    return Result;
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManUpdateRandomInput(uInfo uint32, PiTransProb float32) uint32 {
    var Multi float32  = 1.0 / (1 << 16);
    var i int32;
    if PiTransProb == 0.5 {
        return Gia_ManRandom(0); }
    for  i = 0; i < 32; i++  {
        if Multi * (Gia_ManRandom(0) & 0xffff) < PiTransProb {
            uInfo ^= (1 << i); }
    }
    return uInfo;
}

/**Function*************************************************************

  Synopsis    [Simulates sequential network randomly for the given number of frames.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManSimulateSeqPref(p *Gli_Man_t, nPref int32) {
    var pObj, pObjRi, pObjRo *Gli_Obj_t;
    var i, f int32;
    // initialize simulation data
    // inlining of macro Gli_ManForEachPi
    for  i = 0; (i < Gli_ManPiNum(p)) && ((pObj) = Gli_ManCi(p, i)); i++  {
        pObj.uSimInfo = Gli_ManUpdateRandomInput( pObj.uSimInfo, 0.5 );
    }
    // inlining of macro Gli_ManForEachRo
    for  i = 0; (i < Gli_ManRegNum(p)) && ((pObj) = Gli_ManCi(p, Gli_ManPiNum(p)+i)); i++  {
        pObj.uSimInfo = 0;
    }
    for  f = 0; f < nPref; f++  {
        // simulate one frame
        // inlining of macro Gli_ManForEachNode
        for  i = 0; (i < p.nObjData) && (pObj = Gli_ManObj(p,i)); i += Gli_ObjSize(pObj)  { if Gli_ObjIsTerm(pObj) != 0 {} else {
            pObj.uSimInfo = Gli_ManSimulateSeqNode( p, pObj ); }
        }
        // inlining of macro Gli_ManForEachRi
        for  i = 0; (i < Gli_ManRegNum(p)) && ((pObj) = Gli_ManCo(p, Gli_ManPoNum(p)+i)); i++  {
            pObj.uSimInfo = Gli_ObjFanin(pObj, 0).uSimInfo;
        }
        // initialize the next frame
        // inlining of macro Gli_ManForEachPi
        for  i = 0; (i < Gli_ManPiNum(p)) && ((pObj) = Gli_ManCi(p, i)); i++  {
            pObj.uSimInfo = Gli_ManUpdateRandomInput( pObj.uSimInfo, 0.5 );
        }
        // inlining of macro Gli_ManForEachRiRo
        for  i = 0; (i < Gli_ManRegNum(p)) && ((pObjRi) = Gli_ManCo(p, Gli_ManPoNum(p)+i)) && ((pObjRo) = Gli_ManCi(p, Gli_ManPiNum(p)+i)); i++  {
            pObjRo.uSimInfo = pObjRi.uSimInfo;
        }
    }
    // save simulation data after nPref timeframes
    if p.pSimInfoPrev == nil {
        // inlining of macro ABC_ALLOC
        p.pSimInfoPrev = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(Gli_ManCiNum(p)))); }
    // inlining of macro Gli_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && (pObj = Gli_ManObj(p,Vec_IntEntry(p.vCis,i))); i++  {
        p.pSimInfoPrev[i] = pObj.uSimInfo;
    }
}

/**Function*************************************************************

  Synopsis    [Initialized object values to be one pattern in the saved data.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManSetDataSaved(p *Gli_Man_t, iBit int32) {
    var pObj *Gli_Obj_t;
    var i int32;
    // inlining of macro Gli_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && (pObj = Gli_ManObj(p,Vec_IntEntry(p.vCis,i))); i++  {
        pObj.fPhase = pObj.fPhase2 = ((p.pSimInfoPrev[i] >> iBit) & 1);
    }
    // inlining of macro Gli_ManForEachNode
    for  i = 0; (i < p.nObjData) && (pObj = Gli_ManObj(p,i)); i += Gli_ObjSize(pObj)  { if Gli_ObjIsTerm(pObj) != 0 {} else {
        pObj.fPhase = pObj.fPhase2 = Gli_NodeComputeValue( pObj ); }
    }
}

/**Function*************************************************************

  Synopsis    [Sets random info at the PIs and collects changed PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManSetPiRandomSeq(p *Gli_Man_t, PiTransProb float32) {
    var pObj, pObjRi *Gli_Obj_t;
    var Multi float32  = 1.0 / (1 << 16);
    var i int32;
    assert(0.0 < PiTransProb && PiTransProb < 1.0, "0.0 < PiTransProb && PiTransProb < 1.0");
    // transfer data to the COs
    // inlining of macro Gli_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (pObj = Gli_ManObj(p,Vec_IntEntry(p.vCos,i))); i++  {
        pObj.fPhase = pObj.fPhase2 = Gli_ObjFanin(pObj, 0).fPhase;
    }
    // set changed PIs
    Vec_IntClear( p.vCisChanged );
    // inlining of macro Gli_ManForEachPi
    for  i = 0; (i < Gli_ManPiNum(p)) && ((pObj) = Gli_ManCi(p, i)); i++  {
        if Multi * (Gia_ManRandom(0) & 0xffff) < PiTransProb {
            Vec_IntPush( p.vCisChanged, pObj.Handle );
            pObj.fPhase  ^= 1;
            pObj.fPhase2 ^= 1;
            pObj.nSwitches++;
            pObj.nGlitches++;
        }
    }
    // set changed ROs
    // inlining of macro Gli_ManForEachRiRo
    for  i = 0; (i < Gli_ManRegNum(p)) && ((pObjRi) = Gli_ManCo(p, Gli_ManPoNum(p)+i)) && ((pObj) = Gli_ManCi(p, Gli_ManPiNum(p)+i)); i++  {
        if pObjRi.fPhase != pObj.fPhase {
            Vec_IntPush( p.vCisChanged, pObj.Handle );
            pObj.fPhase  ^= 1;
            pObj.fPhase2 ^= 1;
            pObj.nSwitches++;
            pObj.nGlitches++;
        }
    }

}

/**Function*************************************************************

  Synopsis    [Computes glitching activity of each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gli_ManSwitchesAndGlitches(p *Gli_Man_t, nPatterns int32, PiTransProb float32, fVerbose int32) {
    var i, k int32;
    var clk abctime  = Abc_Clock();
    Gia_ManRandom( 1 );
    Gli_ManFinalize( p );
    if p.nRegs == 0 {
        for  i = 0; i < nPatterns; i++  {
            Gli_ManSetPiRandom( p, PiTransProb );
            Gli_ManSwitching( p );
            Gli_ManGlitching( p );
//            Gli_ManVerify( p );
        }
    
    } else {
        var nIters int32  = Abc_BitWordNum(nPatterns);
        Gli_ManSimulateSeqPref( p, 16 );
        for  i = 0; i < 32; i++  {
            Gli_ManSetDataSaved( p, i );
            for  k = 0; k < nIters; k++  {
                Gli_ManSetPiRandomSeq( p, PiTransProb );
                Gli_ManSwitching( p );
                Gli_ManGlitching( p );
//                Gli_ManVerify( p );
            }
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Simulated %d patterns.  Input transition probability %.2f.  ", nPatterns, PiTransProb);
        // inlining of macro ABC_PRMn
        Abc_Print(1, "%s =", ("Memory"));
Abc_Print(1, "%10.3f MB  ",    1.0*(float64((4*p.nObjData)))/(1<<20));;
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

