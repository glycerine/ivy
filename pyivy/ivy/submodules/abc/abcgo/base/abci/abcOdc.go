package abci

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [abcOdc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Scalable computation of observability don't-cares.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcOdc.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

#define ABC_DC_MAX_NODES   (1<<15)

typedef unsigned short Odc_Lit_t;

        // 16 bytes
type Odc_Obj_t struct {
    iFan0               Odc_Lit_t;       // first fanin
    iFan1               Odc_Lit_t;       // second fanin
    iNext               Odc_Lit_t;       // next node in the hash table
    unsigned short          TravId;      // the traversal ID
    uData                uint32;       // the computed data
    uMask                uint32;       // the variable mask 
}

type Odc_Man_t_ struct {
    // dont'-care parameters
    nVarsMax                     int32;    // the max number of cut variables
    nLevels                     int32;     // the number of ODC levels
    fVerbose                     int32;    // the verbosiness flag
    fVeryVerbose                     int32;// the verbosiness flag to print per-node stats
    nPercCutoff                     int32; // cutoff percentage

    // windowing
    pNode              *Abc_Obj_t;       // the node for windowing
    vLeaves              *Vec_Ptr_t;     // the number of the cut
    vRoots              *Vec_Ptr_t;      // the roots of the cut
    vBranches              *Vec_Ptr_t;   // additional inputs 

    // internal AIG package
    // objects
    nPis                     int32;        // number of PIs (nVarsMax + 32)
    nObjs                     int32;       // number of objects (Const1, PIs, ANDs)
    nObjsAlloc                     int32;  // number of objects allocated
    pObjs              *Odc_Obj_t;       // objects 
    iRoot               Odc_Lit_t;       // the root object
    unsigned short          nTravIds;    // the number of travIDs
    // structural hashing
    pTable              *Odc_Lit_t;      // hash table
    nTableSize                     int32;  // hash table size
    vUsedSpots              *Vec_Int_t;  // the used spots

    // truth tables
    nBits                     int32;       // the number of bits
    nWords                     int32;      // the number of words 
    vTruths              *Vec_Ptr_t;     // truth tables for each node
    vTruthsElem              *Vec_Ptr_t; // elementary truth tables for the PIs
    puTruth               *uint32;     // the place where the resulting truth table does

    // statistics
    nWins                     int32;       // the number of windows processed
    nWinsEmpty                     int32;  // the number of empty windows
    nSimsEmpty                     int32;  // the number of empty simulation infos
    nQuantsOver                     int32; // the number of quantification overflows
    nWinsFinish                     int32; // the number of windows that finished
    nTotalDcs                     int32;   // total percentage of DCs

    // runtime
    timeClean                 abctime;   // windowing
    timeWin                 abctime;     // windowing
    timeMiter                 abctime;   // computing the miter
    timeSim                 abctime;     // simulation
    timeQuant                 abctime;   // quantification
    timeTruth                 abctime;   // truth table
    timeTotal                 abctime;   // useful runtime
    timeAbort                 abctime;   // aborted runtime
}


// quantity of different objects
func Odc_PiNum(p *Odc_Man_t) int32 { return p.nPis;                       }
func Odc_NodeNum(p *Odc_Man_t) int32 { return p.nObjs - p.nPis - 1;        }
func Odc_ObjNum(p *Odc_Man_t) int32 { return p.nObjs;                      }

// complemented attributes of objects
func Odc_IsComplement(Lit Odc_Lit_t) int32 { return Lit &  (Odc_Lit_t)1;           }
func Odc_Regular(Lit Odc_Lit_t) Odc_Lit_t { return Lit & ~(Odc_Lit_t)1;           }
func Odc_Not(Lit Odc_Lit_t) Odc_Lit_t { return Lit ^  (Odc_Lit_t)1;           }
func Odc_NotCond(Lit Odc_Lit_t, c int32) Odc_Lit_t { return Lit ^  (Odc_Lit_t)(c!=0);      }

// specialized Literals
func Odc_Const0() Odc_Lit_t { return 1;                             }
func Odc_Const1() Odc_Lit_t { return 0;                             }
func Odc_Var(p *Odc_Man_t, i int32) Odc_Lit_t { assert(i >= 0 && i < p.nPis, "i >= 0 && i < p.nPis"); return (i+1) << 1;  }
func Odc_IsConst(Lit Odc_Lit_t) int32 { return Lit <  (Odc_Lit_t)2;           }
func Odc_IsTerm(p *Odc_Man_t, Lit Odc_Lit_t) int32 { return int32((Lit>>1)) <= p.nPis;      }

// accessing internal storage
func Odc_ObjNew(p *Odc_Man_t) *Odc_Obj_t { assert(p.nObjs < p.nObjsAlloc, "p.nObjs < p.nObjsAlloc"); return p.pObjs + p.nObjs++;        }
func Odc_Obj2Lit(p *Odc_Man_t, pObj *Odc_Obj_t) Odc_Lit_t { assert(pObj, "pObj"); return (pObj - p.pObjs) << 1;                           }
func Odc_Lit2Obj(p *Odc_Man_t, Lit Odc_Lit_t) *Odc_Obj_t { assert(!(Lit & 1) && int32((Lit>>1)) < p.nObjs, "!(Lit & 1) && int32((Lit>>1)) < p.nObjs"); return p.pObjs + (Lit>>1); }

// fanins and their complements
func Odc_ObjChild0(pObj *Odc_Obj_t) Odc_Lit_t { return pObj.iFan0;                   }
func Odc_ObjChild1(pObj *Odc_Obj_t) Odc_Lit_t { return pObj.iFan1;                   }
func Odc_ObjFanin0(pObj *Odc_Obj_t) Odc_Lit_t { return Odc_Regular(pObj.iFan0);      }
func Odc_ObjFanin1(pObj *Odc_Obj_t) Odc_Lit_t { return Odc_Regular(pObj.iFan1);      }
func Odc_ObjFaninC0(pObj *Odc_Obj_t) int32 { return Odc_IsComplement(pObj.iFan0); }
func Odc_ObjFaninC1(pObj *Odc_Obj_t) int32 { return Odc_IsComplement(pObj.iFan1); }

// traversal IDs
func Odc_ManIncrementTravId(p *Odc_Man_t) { p.nTravIds++;                                    }
func Odc_ObjSetTravIdCurrent(p *Odc_Man_t, pObj *Odc_Obj_t) { pObj.TravId = p.nTravIds;                       }
func Odc_ObjIsTravIdCurrent(p *Odc_Man_t, pObj *Odc_Obj_t) int32 { return int32( (int32(pObj.TravId) == p.nTravIds));  }

// truth tables
func Odc_ObjTruth(p *Odc_Man_t, Lit Odc_Lit_t) *uint32 { assert(!(Lit & 1), "!(Lit & 1)"); return (unsigned *) Vec_PtrEntry(p.vTruths, Lit >> 1);  }

// iterators 
// Original macro:
// #define Odc_ForEachPi( p, Lit, i )                                                 \
//     for ( i = 0; (i < Odc_PiNum(p)) && (((Lit) = Odc_Var(p, i)), 1); i++ )

// Original macro:
// #define Odc_ForEachAnd( p, pObj, i )                                               \
//     for ( i = 1 + Odc_CiNum(p); (i < Odc_ObjNum(p)) && ((pObj) = (p)->pObjs + i); i++ )



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the don't-care manager.]

  Description [The parameters are the max number of cut variables, 
  the number of fanout levels used for the ODC computation, and verbosiness.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareAlloc(nVarsMax int32, nLevels int32, fVerbose int32, fVeryVerbose int32) *Odc_Man_t {
    var p *Odc_Man_t;
    var pData *uint32;
    var i, k int32;
    // inlining of macro ABC_ALLOC
    p = ((Odc_Man_t *) malloc(sizeof(Odc_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Odc_Man_t) );
    assert(nVarsMax > 4 && nVarsMax < 16, "nVarsMax > 4 && nVarsMax < 16");
    assert(nLevels > 0 && nLevels < 10, "nLevels > 0 && nLevels < 10");

    srand( 0xABC );

    // dont'-care parameters
    p.nVarsMax     = nVarsMax;
    p.nLevels      = nLevels;
    p.fVerbose     = fVerbose;
    p.fVeryVerbose = fVeryVerbose;
    p.nPercCutoff  = 10;

    // windowing
    p.vRoots    = Vec_PtrAlloc( 128 );
    p.vBranches = Vec_PtrAlloc( 128 );

    // internal AIG package
    // allocate room for objects
    p.nObjsAlloc = ABC_DC_MAX_NODES; 
    // inlining of macro ABC_ALLOC
    p.pObjs = ((Odc_Obj_t *) malloc(sizeof(Odc_Obj_t) * (size_t)(p.nObjsAlloc * sizeof(Odc_Obj_t))));
    p.nPis  = nVarsMax + 32;
    p.nObjs = 1 + p.nPis;
    memset( p.pObjs, 0, p.nObjs * sizeof(Odc_Obj_t) );
    // set the PI masks
    for  i = 0; i < 32; i++  {
        p.pObjs[1 + p.nVarsMax + i].uMask = (1 << i);
    }
    // allocate hash table
    p.nTableSize = p.nObjsAlloc/3 + 1;
    // inlining of macro ABC_ALLOC
    p.pTable = ((Odc_Lit_t *) malloc(sizeof(Odc_Lit_t) * (size_t)(p.nTableSize * sizeof(Odc_Lit_t))));
    memset( p.pTable, 0, p.nTableSize * sizeof(Odc_Lit_t) );
    p.vUsedSpots = Vec_IntAlloc( 1000 );

    // truth tables
    p.nWords = Abc_TruthWordNum( p.nVarsMax );
    p.nBits = p.nWords * 8 * sizeof(unsigned);
    p.vTruths = Vec_PtrAllocSimInfo( p.nObjsAlloc, p.nWords );
    p.vTruthsElem = Vec_PtrAllocSimInfo( p.nVarsMax, p.nWords );

    // set elementary truth tables
    Abc_InfoFill( (unsigned *)Vec_PtrEntry(p.vTruths, 0), p.nWords );
    for  k = 0; k < p.nVarsMax; k++  {
//        pData = Odc_ObjTruth( p, Odc_Var(p, k) );
        pData = (unsigned *)Vec_PtrEntry( p.vTruthsElem, k );
        Abc_InfoClear( pData, p.nWords );
        for  i = 0; i < p.nBits; i++  {
            if i & (1 << k) {
                pData[i>>5] |= (1 << (i&31)); }
        }
    }

    // set random truth table for the additional inputs
    for  k = p.nVarsMax; k < p.nPis; k++  {
        pData = Odc_ObjTruth( p, Odc_Var(p, k) );
        Abc_InfoRandom( pData, p.nWords );
    }

    // set the miter to the unused value
    p.iRoot = 0xffff;
    return p;
}

/**Function*************************************************************

  Synopsis    [Clears the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareClear(p *Odc_Man_t) {
    var clk abctime  = Abc_Clock();
    // clean the structural hashing table
    if Vec_IntSize(p.vUsedSpots) > p.nTableSize/3 { // more than one third
        memset( p.pTable, 0, sizeof(Odc_Lit_t) * p.nTableSize );
    } else {
        var iSpot, i int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vUsedSpots)) && (((iSpot) = Vec_IntEntry(p.vUsedSpots, i)), 1); i++  {
            p.pTable[iSpot] = 0;
        }
    }
    Vec_IntClear( p.vUsedSpots ); 
    // reset the number of nodes
    p.nObjs = 1 + p.nPis;
    // reset the root node
    p.iRoot = 0xffff;

p.timeClean += Abc_Clock() - clk;
}

/**Function*************************************************************

  Synopsis    [Frees the don't-care manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareFree(p *Odc_Man_t) {
    if p.fVerbose {
        fmt.Printf("Wins = %5d. Empty = %5d. SimsEmpty = %5d. QuantOver = %5d. WinsFinish = %5d.\n", p.nWins, p.nWinsEmpty, p.nSimsEmpty, p.nQuantsOver, p.nWinsFinish);
        fmt.Printf("Ave DCs per window = %6.2f %%. Ave DCs per finished window = %6.2f %%.\n", 1.0*p.nTotalDcs/p.nWins, 1.0*p.nTotalDcs/p.nWinsFinish);
        fmt.Printf("Runtime stats of the ODC manager:\n");
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Cleaning    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeClean)))/(float64(int64(time.Second))));;
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Windowing   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeWin)))/(float64(int64(time.Second))));;
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Miter       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMiter)))/(float64(int64(time.Second))));;
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Simulation  "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSim)))/(float64(int64(time.Second))));;
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Quantifying "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeQuant)))/(float64(int64(time.Second))));;
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Truth table "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTruth)))/(float64(int64(time.Second))));;
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("TOTAL       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Aborted     "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeAbort)))/(float64(int64(time.Second))));;
    }
    Vec_PtrFree( p.vRoots );
    Vec_PtrFree( p.vBranches );
    Vec_PtrFree( p.vTruths );
    Vec_PtrFree( p.vTruthsElem );
    Vec_IntFree( p.vUsedSpots );
    // inlining of macro ABC_FREE
    ((p.pObjs) ? (free((char *) (p.pObjs)), (p.pObjs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTable) ? (free((char *) (p.pTable)), (p.pTable) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}



/**Function*************************************************************

  Synopsis    [Marks the TFO of the collected nodes up to the given level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareWinSweepLeafTfo_rec(pObj *Abc_Obj_t, nLevelLimit int32, pNode *Abc_Obj_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    if Abc_ObjIsCo(pObj) || int32(pObj.Level) > nLevelLimit || pObj == pNode {
        return; }
    if Abc_NodeIsTravIdCurrent(pObj) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    ////////////////////////////////////////
    // try to reduce the runtime
    if Abc_ObjFanoutNum(pObj) > 100 {
        return; }
    ////////////////////////////////////////
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Abc_NtkDontCareWinSweepLeafTfo_rec( pFanout, nLevelLimit, pNode );
    }
}

/**Function*************************************************************

  Synopsis    [Marks the TFO of the collected nodes up to the given level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareWinSweepLeafTfo(p *Odc_Man_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_NtkIncrementTravId( p.pNode.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, i)), 1); i++  {
        Abc_NtkDontCareWinSweepLeafTfo_rec( pObj, p.pNode.Level + p.nLevels, p.pNode );
    }
}

/**Function*************************************************************

  Synopsis    [Recursively collects the roots.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareWinCollectRoots_rec(pObj *Abc_Obj_t, vRoots *Vec_Ptr_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    assert(Abc_NodeIsTravIdCurrent(pObj), "Abc_NodeIsTravIdCurrent(pObj)");
    // check if the node has all fanouts marked
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if !Abc_NodeIsTravIdCurrent(pFanout) {
            break; }
    }
    // if some of the fanouts are unmarked, add the node to the root
    if i < Abc_ObjFanoutNum(pObj) {
        Vec_PtrPushUnique( vRoots, pObj );
        return;
    }
    // otherwise, call recursively
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Abc_NtkDontCareWinCollectRoots_rec( pFanout, vRoots );
    }
}

/**Function*************************************************************

  Synopsis    [Collects the roots of the window.]

  Description [Roots of the window are the nodes that have at least
  one fanout that it not in the TFO of the leaves.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareWinCollectRoots(p *Odc_Man_t) {
    assert(!Abc_NodeIsTravIdCurrent(p.pNode), "!Abc_NodeIsTravIdCurrent(p.pNode)");
    // mark the node with the old traversal ID
    Abc_NodeSetTravIdCurrent( p.pNode ); 
    // collect the roots
    Vec_PtrClear( p.vRoots );
    Abc_NtkDontCareWinCollectRoots_rec( p.pNode, p.vRoots );
}
 
/**Function*************************************************************

  Synopsis    [Recursively adds missing nodes and leaves.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Abc_NtkDontCareWinAddMissing_rec(p *Odc_Man_t, pObj *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    // skip the already collected leaves and branches
    if Abc_NodeIsTravIdCurrent(pObj) {
        return 1; }
    // if this is not an internal node - make it a new branch
    if !Abc_NodeIsTravIdPrevious(pObj) || Abc_ObjIsCi(pObj) { //|| (int)pObj->Level <= p->nLevLeaves )
    
        Abc_NodeSetTravIdCurrent( pObj );
        Vec_PtrPush( p.vBranches, pObj );
        return Vec_PtrSize(p.vBranches) <= 32;
    }
    // visit the fanins of the node
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if !(Abc_NtkDontCareWinAddMissing_rec( p, pFanin ) != 0) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds to the window nodes and leaves in the TFI of the roots.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareWinAddMissing(p *Odc_Man_t) int32 {
    var pObj *Abc_Obj_t;
    var i int32;
    // set the leaves
    Abc_NtkIncrementTravId( p.pNode.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pObj );
    }        
    // explore from the roots
    Vec_PtrClear( p.vBranches );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vRoots, i)), 1); i++  {
        if !(Abc_NtkDontCareWinAddMissing_rec( p, pObj ) != 0) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes window for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareWindow(p *Odc_Man_t) int32 {
    // mark the TFO of the collected nodes up to the given level (p->pNode->Level + p->nWinTfoMax)
    Abc_NtkDontCareWinSweepLeafTfo( p );
    // find the roots of the window
    Abc_NtkDontCareWinCollectRoots( p );
    if Vec_PtrSize(p.vRoots) == 1 && Vec_PtrEntry(p.vRoots, 0) == p.pNode {
//        printf( "Empty window\n" );
        return 0;
    }
    // add the nodes in the TFI of the roots that are not yet in the window
    if !(Abc_NtkDontCareWinAddMissing( p ) != 0) {
//        printf( "Too many branches (%d)\n", Vec_PtrSize(p->vBranches) );
        return 0;
    }
    return 1;
}





/**Function*************************************************************

  Synopsis    [Performing hashing of two AIG Literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Odc_HashKey(iFan0 Odc_Lit_t, iFan1 Odc_Lit_t, TableSize int32) uint32 {
    var Key uint32;
    Key ^= Odc_Regular(iFan0) * 7937;
    Key ^= Odc_Regular(iFan1) * 2971;
    Key ^= Odc_IsComplement(iFan0) * 911;
    Key ^= Odc_IsComplement(iFan1) * 353;
    return Key % TableSize;
}

/**Function*************************************************************

  Synopsis    [Checks if the given name node already exists in the table.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Odc_HashLookup(p *Odc_Man_t, iFan0 Odc_Lit_t, iFan1 Odc_Lit_t) *Odc_Lit_t {
    var pObj *Odc_Obj_t;
    var pEntry *Odc_Lit_t;
    var uHashKey uint32;
    assert(iFan0 < iFan1, "iFan0 < iFan1");
    // get the hash key for this node
    uHashKey = Odc_HashKey( iFan0, iFan1, p.nTableSize );
    // remember the spot in the hash table that will be used
    if p.pTable[uHashKey] == 0 {
        Vec_IntPush( p.vUsedSpots, uHashKey ); }
    // find the entry
    for  pEntry = p.pTable + uHashKey; *pEntry; pEntry = &pObj.iNext  {
        pObj = Odc_Lit2Obj( p, *pEntry );
        if pObj.iFan0 == iFan0 && pObj.iFan1 == iFan1 {
            return pEntry; }
    }
    return pEntry;
}

/**Function*************************************************************

  Synopsis    [Finds node by structural hashing or creates a new node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Odc_And(p *Odc_Man_t, iFan0 Odc_Lit_t, iFan1 Odc_Lit_t) Odc_Lit_t {
    var pObj *Odc_Obj_t;
    var pEntry *Odc_Lit_t;
    var uMask0, uMask1 uint32;
    var Temp int32;
    // consider trivial cases
    if iFan0 == iFan1 {
        return iFan0; }
    if iFan0 == Odc_Not(iFan1) {
        return Odc_Const0(); }
    if Odc_Regular(iFan0) == Odc_Const1() {
        if iFan0 == Odc_Const1() {
            return iFan1;
        } else {
            return Odc_Const0();
        } }
    if Odc_Regular(iFan1) == Odc_Const1() {
        if iFan1 == Odc_Const1() {
            return iFan0;
        } else {
            return Odc_Const0();
        } }
    // canonicize the fanin order
    if iFan0 > iFan1 {
        Temp = iFan0;
iFan0 = iFan1;
iFan1 = Temp; }
    // check if a node with these fanins exists
    pEntry = Odc_HashLookup( p, iFan0, iFan1 );
    if *pEntry {
        return *pEntry; }
    // create a new node
    pObj = Odc_ObjNew( p );
    pObj.iFan0 = iFan0;
    pObj.iFan1 = iFan1;
    pObj.iNext = 0;
    pObj.TravId = 0;
    // set the mask
    uMask0 = Odc_Lit2Obj(p, Odc_Regular(iFan0)).uMask;
    uMask1 = Odc_Lit2Obj(p, Odc_Regular(iFan1)).uMask;
    pObj.uMask = uMask0 | uMask1;
    // add to the table
    *pEntry = Odc_Obj2Lit( p, pObj );
    return *pEntry;
}

/**Function*************************************************************

  Synopsis    [Boolean OR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Odc_Or(p *Odc_Man_t, iFan0 Odc_Lit_t, iFan1 Odc_Lit_t) Odc_Lit_t {
    return Odc_Not( Odc_And(p, Odc_Not(iFan0), Odc_Not(iFan1)) );
}

/**Function*************************************************************

  Synopsis    [Boolean XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Odc_Xor(p *Odc_Man_t, iFan0 Odc_Lit_t, iFan1 Odc_Lit_t) Odc_Lit_t {
    return Odc_Or( p, Odc_And(p, iFan0, Odc_Not(iFan1)), Odc_And(p, Odc_Not(iFan0), iFan1) );
}





/**Function*************************************************************

  Synopsis    [Transfers the window into the AIG package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareTransfer_rec(p *Odc_Man_t, pNode *Abc_Obj_t, pPivot *Abc_Obj_t) any {
    var uData0, uData1 uint32;
    var uLit0, uLit1, uRes0, uRes1 Odc_Lit_t;
    assert(!Abc_ObjIsComplement(pNode), "!Abc_ObjIsComplement(pNode)");
    // skip visited objects
    if Abc_NodeIsTravIdCurrent(pNode) {
        return pNode.pCopy; }
    Abc_NodeSetTravIdCurrent(pNode);
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // consider the case when the node is the pivot
    if pNode == pPivot {
        return pNode.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)((Odc_Const1() << 16) | Odc_Const0()); }
    // compute the cofactors
    uData0 = uint32((ABC_PTRUINT_T))Abc_NtkDontCareTransfer_rec( p, Abc_ObjFanin0(pNode), pPivot );
    uData1 = uint32((ABC_PTRUINT_T))Abc_NtkDontCareTransfer_rec( p, Abc_ObjFanin1(pNode), pPivot );
    // find the 0-cofactor
    uLit0 = Odc_NotCond( (Odc_Lit_t)(uData0 & 0xffff), Abc_ObjFaninC0(pNode) );
    uLit1 = Odc_NotCond( (Odc_Lit_t)(uData1 & 0xffff), Abc_ObjFaninC1(pNode) );
    uRes0 = Odc_And( p, uLit0, uLit1 );
    // find the 1-cofactor
    uLit0 = Odc_NotCond( (Odc_Lit_t)(uData0 >> 16), Abc_ObjFaninC0(pNode) );
    uLit1 = Odc_NotCond( (Odc_Lit_t)(uData1 >> 16), Abc_ObjFaninC1(pNode) );
    uRes1 = Odc_And( p, uLit0, uLit1 );
    // find the result
    return pNode.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)((uRes1 << 16) | uRes0);
}

/**Function*************************************************************

  Synopsis    [Transfers the window into the AIG package.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareTransfer(p *Odc_Man_t) int32 {
    var pObj *Abc_Obj_t;
    var uRes0, uRes1 Odc_Lit_t;
    var uLit Odc_Lit_t;
    var uData uint32;
    var i int32;
    Abc_NtkIncrementTravId( p.pNode.pNtk );
    // set elementary variables at the leaves 
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, i)), 1); i++  {
        uLit = Odc_Var( p, i );
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)((uLit << 16) | uLit);
        Abc_NodeSetTravIdCurrent(pObj);
    }
    // set elementary variables at the branched 
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vBranches)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vBranches, i)), 1); i++  {
        uLit = Odc_Var( p, i+p.nVarsMax );
        pObj.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)((uLit << 16) | uLit);
        Abc_NodeSetTravIdCurrent(pObj);
    }
    // compute the AIG for the window
    p.iRoot = Odc_Const0();
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(p.vRoots, i)), 1); i++  {
        uData = uint32((ABC_PTRUINT_T))Abc_NtkDontCareTransfer_rec( p, pObj, p.pNode );
        // get the cofactors
        uRes0 = uData & 0xffff;
        uRes1 = uData >> 16;
        // compute the miter
//        assert( uRes0 != uRes1 ); // may be false if the node is redundant w.r.t. this root
        uLit = Odc_Xor( p, uRes0, uRes1 );
        p.iRoot = Odc_Or( p, p.iRoot, uLit );
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Recursively computes the pair of cofactors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareCofactors_rec(p *Odc_Man_t, Lit Odc_Lit_t, uMask uint32) uint32 {
    var pObj *Odc_Obj_t;
    var uData0, uData1 uint32;
    var uLit0, uLit1, uRes0, uRes1 Odc_Lit_t;
    assert(!(Odc_IsComplement(Lit) != 0), "!Odc_IsComplement(Lit)");
    // skip visited objects
    pObj = Odc_Lit2Obj( p, Lit );
    if Odc_ObjIsTravIdCurrent(p, pObj) != 0 {
        return pObj.uData; }
    Odc_ObjSetTravIdCurrent(p, pObj);
    // skip objects out of the cone
    if (pObj.uMask & uMask) == 0 {
        return pObj.uData = ((Lit << 16) | Lit); }
    // consider the case when the node is the var
    if pObj.uMask == uMask && (Odc_IsTerm(p, Lit) != 0) {
        return pObj.uData = ((Odc_Const1() << 16) | Odc_Const0()); }
    // compute the cofactors
    uData0 = Abc_NtkDontCareCofactors_rec( p, Odc_ObjFanin0(pObj), uMask );
    uData1 = Abc_NtkDontCareCofactors_rec( p, Odc_ObjFanin1(pObj), uMask );
    // find the 0-cofactor
    uLit0 = Odc_NotCond( (Odc_Lit_t)(uData0 & 0xffff), Odc_ObjFaninC0(pObj) );
    uLit1 = Odc_NotCond( (Odc_Lit_t)(uData1 & 0xffff), Odc_ObjFaninC1(pObj) );
    uRes0 = Odc_And( p, uLit0, uLit1 );
    // find the 1-cofactor
    uLit0 = Odc_NotCond( (Odc_Lit_t)(uData0 >> 16), Odc_ObjFaninC0(pObj) );
    uLit1 = Odc_NotCond( (Odc_Lit_t)(uData1 >> 16), Odc_ObjFaninC1(pObj) );
    uRes1 = Odc_And( p, uLit0, uLit1 );
    // find the result
    return pObj.uData = ((uRes1 << 16) | uRes0);
}

/**Function*************************************************************

  Synopsis    [Quantifies the branch variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareQuantify(p *Odc_Man_t) int32 {   
    var uRes0, uRes1 Odc_Lit_t;
    var uData uint32;
    var i int32;
    assert(p.iRoot < 0xffff, "p.iRoot < 0xffff");
    assert(Vec_PtrSize(p.vBranches) <= 32, "Vec_PtrSize(p.vBranches) <= 32"); // the mask size
    for  i = 0; i < Vec_PtrSize(p.vBranches); i++  {
        // compute the cofactors w.r.t. this variable
        Odc_ManIncrementTravId( p );
        uData = Abc_NtkDontCareCofactors_rec( p, Odc_Regular(p.iRoot), (1 << i) );
        uRes0 = Odc_NotCond( (Odc_Lit_t)(uData & 0xffff), Odc_IsComplement(p.iRoot) );
        uRes1 = Odc_NotCond( (Odc_Lit_t)(uData >> 16),    Odc_IsComplement(p.iRoot) );
        // quantify this variable existentially
        p.iRoot = Odc_Or( p, uRes0, uRes1 );
        // check the limit
        if Odc_ObjNum(p) > ABC_DC_MAX_NODES/2 {
            return 0; }
    }
    assert(p.nObjs <= p.nObjsAlloc, "p.nObjs <= p.nObjsAlloc");
    return 1;
}



/**Function*************************************************************

  Synopsis    [Set elementary truth tables for PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareSimulateSetElem2(p *Odc_Man_t) {
    var pData *uint32;
    var i, k int32;
    for  k = 0; k < p.nVarsMax; k++  {
        pData = Odc_ObjTruth( p, Odc_Var(p, k) );
        Abc_InfoClear( pData, p.nWords );
        for  i = 0; i < p.nBits; i++  {
            if i & (1 << k) {
                pData[i>>5] |= (1 << (i&31)); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Set elementary truth tables for PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareSimulateSetElem(p *Odc_Man_t) {
    var pData, pData2 *uint32;
    var k int32;
    for  k = 0; k < p.nVarsMax; k++  {
        pData = Odc_ObjTruth( p, Odc_Var(p, k) );
        pData2 = (unsigned *)Vec_PtrEntry( p.vTruthsElem, k );
        Abc_InfoCopy( pData, pData2, p.nWords );
    }
}

/**Function*************************************************************

  Synopsis    [Set random simulation words for PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareSimulateSetRand(p *Odc_Man_t) {
    var pData *uint32;
    var w, k, Number int32;
    for  w = 0; w < p.nWords; w++  {
        Number = rand();
        for  k = 0; k < p.nVarsMax; k++  {
            pData = Odc_ObjTruth( p, Odc_Var(p, k) );
            pData[w] = (Number & (1<<k)) ? ~0 : 0;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Set random simulation words for PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareCountMintsWord(p *Odc_Man_t, puTruth *uint32) int32 {
    var w, Counter int32;
    for  w = 0; w < p.nWords; w++  {
        if puTruth[w] {
            Counter++; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareTruthOne(p *Odc_Man_t, Lit Odc_Lit_t) {
    var pObj *Odc_Obj_t;
    var pInfo, pInfo1, pInfo2 *uint32;
    var k, fComp1, fComp2 int32;
    assert(!(Odc_IsComplement( Lit ) != 0), "!Odc_IsComplement( Lit )");
    assert(!(Odc_IsTerm( p, Lit ) != 0), "!Odc_IsTerm( p, Lit )");
    // get the truth tables
    pObj   = Odc_Lit2Obj( p, Lit );
    pInfo  = Odc_ObjTruth( p, Lit );
    pInfo1 = Odc_ObjTruth( p, Odc_ObjFanin0(pObj) );
    pInfo2 = Odc_ObjTruth( p, Odc_ObjFanin1(pObj) );
    fComp1 = Odc_ObjFaninC0( pObj );
    fComp2 = Odc_ObjFaninC1( pObj );
    // simulate
    if fComp1 && (fComp2 != 0) {
        for  k = 0; k < p.nWords; k++  {
            pInfo[k] = ~pInfo1[k] & ~pInfo2[k];
        }
    } else if fComp1 && !(fComp2 != 0) {
        for  k = 0; k < p.nWords; k++  {
            pInfo[k] = ~pInfo1[k] &  pInfo2[k];
        }
    } else if !fComp1 && (fComp2 != 0) {
        for  k = 0; k < p.nWords; k++  {
            pInfo[k] =  pInfo1[k] & ~pInfo2[k];
        }
    } else { // if ( fComp1 && fComp2 )
        for  k = 0; k < p.nWords; k++  {
            pInfo[k] =  pInfo1[k] &  pInfo2[k];
        } }
}

/**Function*************************************************************

  Synopsis    [Computes the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareSimulate_rec(p *Odc_Man_t, Lit Odc_Lit_t) {
    var pObj *Odc_Obj_t;
    assert(!(Odc_IsComplement(Lit) != 0), "!Odc_IsComplement(Lit)");
    // skip terminals
    if Odc_IsTerm(p, Lit) != 0 {
        return; }
    // skip visited objects
    pObj = Odc_Lit2Obj( p, Lit );
    if Odc_ObjIsTravIdCurrent(p, pObj) != 0 {
        return; }
    Odc_ObjSetTravIdCurrent(p, pObj);
    // call recursively
    Abc_NtkDontCareSimulate_rec( p, Odc_ObjFanin0(pObj) );
    Abc_NtkDontCareSimulate_rec( p, Odc_ObjFanin1(pObj) );
    // construct the truth table
    Abc_NtkDontCareTruthOne( p, Lit );
}

/**Function*************************************************************

  Synopsis    [Computes the truth table of the care set.]

  Description [Returns the number of ones in the simulation info.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareSimulate(p *Odc_Man_t, puTruth *uint32) int32 {
    Odc_ManIncrementTravId( p );
    Abc_NtkDontCareSimulate_rec( p, Odc_Regular(p.iRoot) );
    Abc_InfoCopy( puTruth, Odc_ObjTruth(p, Odc_Regular(p.iRoot)), p.nWords );
    if Odc_IsComplement(p.iRoot) != 0 {
        Abc_InfoNot( puTruth, p.nWords ); }
    return Extra_TruthCountOnes( puTruth, p.nVarsMax );
}

/**Function*************************************************************

  Synopsis    [Computes the truth table of the care set.]

  Description [Returns the number of ones in the simulation info.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareSimulateBefore(p *Odc_Man_t, puTruth *uint32) int32 {
    var nIters int32  = 2;
    var nRounds, Counter, r int32;
    // decide how many rounds to simulate
    nRounds = p.nBits / p.nWords;
    Counter = 0;
    for  r = 0; r < nIters; r++  {
        Abc_NtkDontCareSimulateSetRand( p );
        Abc_NtkDontCareSimulate( p, puTruth );
        Counter += Abc_NtkDontCareCountMintsWord( p, puTruth );
    }
    // normalize
    Counter = Counter * nRounds / nIters;
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Computes ODCs for the node in terms of the cut variables.]

  Description [Returns the number of don't care minterms in the truth table.
  In particular, this procedure returns 0 if there is no don't-cares.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDontCareCompute(p *Odc_Man_t, pNode *Abc_Obj_t, vLeaves *Vec_Ptr_t, puTruth *uint32) int32 {
    var nMints, RetValue int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();

    p.nWins++;
    
    // set the parameters
    assert(!Abc_ObjIsComplement(pNode), "!Abc_ObjIsComplement(pNode)");
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    assert(Vec_PtrSize(vLeaves) <= p.nVarsMax, "Vec_PtrSize(vLeaves) <= p.nVarsMax");
    p.vLeaves = vLeaves;
    p.pNode = pNode;

    // compute the window
clk = Abc_Clock();
    RetValue = Abc_NtkDontCareWindow( p );
p.timeWin += Abc_Clock() - clk;
    if !(RetValue != 0) {
p.timeAbort += Abc_Clock() - clkTotal;
        Abc_InfoFill( puTruth, p.nWords );
        p.nWinsEmpty++;        
        return 0;
    }

    if p.fVeryVerbose {
        fmt.Printf(" %5d : ", pNode.Id);
        fmt.Printf("Leaf = %2d ", Vec_PtrSize(p.vLeaves));
        fmt.Printf("Root = %2d ", Vec_PtrSize(p.vRoots));
        fmt.Printf("Bran = %2d ", Vec_PtrSize(p.vBranches));
        fmt.Printf(" |  ");
    }

    // transfer the window into the AIG package
clk = Abc_Clock();
    Abc_NtkDontCareTransfer( p );
p.timeMiter += Abc_Clock() - clk;

    // simulate to estimate the amount of don't-cares
clk = Abc_Clock();
    nMints = Abc_NtkDontCareSimulateBefore( p, puTruth );
p.timeSim += Abc_Clock() - clk;
    if p.fVeryVerbose {
        fmt.Printf("AIG = %5d ", Odc_NodeNum(p));
        fmt.Printf("%6.2f %%  ", 100.0 * (p.nBits - nMints) / p.nBits);
    }

    // if there is less then the given percentage of don't-cares, skip
    if 100.0 * (p.nBits - nMints) / p.nBits < 1.0 * p.nPercCutoff {
p.timeAbort += Abc_Clock() - clkTotal;
        if p.fVeryVerbose {
            fmt.Printf("Simulation cutoff.\n"); }
        Abc_InfoFill( puTruth, p.nWords );
        p.nSimsEmpty++;
        return 0;
    }

    // quantify external variables
clk = Abc_Clock();
    RetValue = Abc_NtkDontCareQuantify( p );
p.timeQuant += Abc_Clock() - clk;
    if !(RetValue != 0) {
p.timeAbort += Abc_Clock() - clkTotal;
        if p.fVeryVerbose {
            fmt.Printf("=== Overflow! ===\n"); }
        Abc_InfoFill( puTruth, p.nWords );
        p.nQuantsOver++;
        return 0;
    }

    // get the truth table
clk = Abc_Clock();
    Abc_NtkDontCareSimulateSetElem( p );
    nMints = Abc_NtkDontCareSimulate( p, puTruth );
p.timeTruth += Abc_Clock() - clk;
    if p.fVeryVerbose {
        fmt.Printf("AIG = %5d ", Odc_NodeNum(p));
        fmt.Printf("%6.2f %%  ", 100.0 * (p.nBits - nMints) / p.nBits);
        fmt.Printf("\n");
    }
p.timeTotal += Abc_Clock() - clkTotal;
    p.nWinsFinish++;
    p.nTotalDcs += int32((100.0 * (p.nBits - nMints) / p.nBits));
    return nMints;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

