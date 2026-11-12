package csw

/**CFile****************************************************************

  FileName    [cswInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Cut sweeping.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 11, 2007.]

  Revision    [$Id: cswInt.h,v 1.00 2007/07/11 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "aig/aig/aig.h"
//#include "opt/dar/dar.h"
//#include "bool/kit/kit.h"
//#include "csw.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START
 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Csw_Man_t =            Csw_Man_t_;
              

// the cut used to represent node in the AIG
type Csw_Cut_t struct {
    pNext      *Csw_Cut_t;           // the next cut in the table 
    Cost             int32;            // the cost of the cut
//    float           Cost;            // the cost of the cut
    uSign        uint32;           // cut signature
    iNode             int32;           // the node, for which it is the cut
    nCutSize           int16;        // the number of bytes in the cut
    nLeafMax            byte;        // the maximum number of fanins
    nFanins            byte;         // the current number of fanins
    pFanins [0]int32;      // the fanins (followed by the truth table)
}

// the CNF computation manager
type Csw_Man_t_ struct {
    // AIG manager
    pManAig      *Aig_Man_t;         // the input AIG manager
    pManRes      *Aig_Man_t;         // the output AIG manager
    pEquiv     **Aig_Obj_t;          // the equivalent nodes in the resulting manager
    pCuts     **Csw_Cut_t;           // the cuts for each node in the output manager
    pnRefs            *int32;          // the number of references of each new node
    // hash table for cuts
    pTable     **Csw_Cut_t;          // the table composed of cuts 
    nTableSize             int32;      // the size of hash table
    // parameters
    nCutsMax             int32;        // the max number of cuts at the node
    nLeafMax             int32;        // the max number of leaves of a cut
    fVerbose             int32;        // enables verbose output
    // internal variables
    nCutSize             int32;        // the number of bytes needed to store one cut
    nTruthWords             int32;     // the number of truth table words
    pMemCuts  *Aig_MmFixed_t;        // memory manager for cuts
    puTemp [4]*uint32;       // used for the truth table computation
    // statistics
    nNodesTriv0             int32;     // the number of trivial nodes
    nNodesTriv1             int32;     // the number of trivial nodes
    nNodesTriv2             int32;     // the number of trivial nodes
    nNodesCuts             int32;      // the number of rewritten nodes
    nNodesTried             int32;     // the number of nodes tried
    timeCuts         abctime;        // time to compute the cut and its truth table
    timeHash         abctime;        // time for hashing cuts
    timeOther         abctime;       // other time
    timeTotal         abctime;       // total time
}

func Csw_CutLeaveNum(pCut *Csw_Cut_t) int32 { return pCut.nFanins;                                   }
func Csw_CutLeaves(pCut *Csw_Cut_t) *int32 { return pCut.pFanins;                                   }
func Csw_CutTruth(pCut *Csw_Cut_t) *uint32 { return (unsigned *)(pCut.pFanins + pCut.nLeafMax);    }
func Csw_CutNext(pCut *Csw_Cut_t) *Csw_Cut_t { return (Csw_Cut_t *)(((char *)pCut) + pCut.nCutSize);  }

func Csw_ObjRefs(p *Csw_Man_t, pObj *Aig_Obj_t) int32 { return p.pnRefs[pObj.Id];   }
func Csw_ObjAddRefs(p *Csw_Man_t, pObj *Aig_Obj_t, nRefs int32) { p.pnRefs[pObj.Id] += nRefs; }

func Csw_ObjCuts(p *Csw_Man_t, pObj *Aig_Obj_t) *Csw_Cut_t { return p.pCuts[pObj.Id];    }
func Csw_ObjSetCuts(p *Csw_Man_t, pObj *Aig_Obj_t, pCuts *Csw_Cut_t) { p.pCuts[pObj.Id] = pCuts;   }

func Csw_ObjEquiv(p *Csw_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { return p.pEquiv[pObj.Id];   }
func Csw_ObjSetEquiv(p *Csw_Man_t, pObj *Aig_Obj_t, pEquiv *Aig_Obj_t) { p.pEquiv[pObj.Id] = pEquiv; }

func Csw_ObjChild0Equiv(p *Csw_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin0(pObj) {
    return Aig_NotCond(Csw_ObjEquiv(p, Aig_ObjFanin0(pObj)), Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Csw_ObjChild1Equiv(p *Csw_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) {
    return Aig_NotCond(Csw_ObjEquiv(p, Aig_ObjFanin1(pObj)), Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                           ITERATORS                              ///
////////////////////////////////////////////////////////////////////////

// iterator over cuts of the node
// Original macro:
// #define Csw_ObjForEachCut( p, pObj, pCut, i )                           \
//     for ( i = 0, pCut = Csw_ObjCuts(p, pObj); i < p->nCutsMax; i++, pCut = Csw_CutNext(pCut) ) 

// iterator over leaves of the cut
// Original macro:
// #define Csw_CutForEachLeaf( p, pCut, pLeaf, i )                         \
//     for ( i = 0; (i < (int)(pCut)->nFanins) && ((pLeaf) = Aig_ManObj(p, (pCut)->pFanins[i])); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cnfCut.c ========================================================*/
// extern Csw_Cut_t *    Csw_ObjPrepareCuts( Csw_Man_t * p, Aig_Obj_t * pObj, int fTriv );

// extern Aig_Obj_t *    Csw_ObjSweep( Csw_Man_t * p, Aig_Obj_t * pObj, int fTriv );

/*=== cnfMan.c ========================================================*/
// extern Csw_Man_t *    Csw_ManStart( Aig_Man_t * pMan, int nCutsMax, int nLeafMax, int fVerbose );

// extern void           Csw_ManStop( Csw_Man_t * p );

/*=== cnfTable.c ========================================================*/
// extern int            Csw_TableCountCuts( Csw_Man_t * p );

// extern void           Csw_TableCutInsert( Csw_Man_t * p, Csw_Cut_t * pCut );

// extern Aig_Obj_t *    Csw_TableCutLookup( Csw_Man_t * p, Csw_Cut_t * pCut );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

