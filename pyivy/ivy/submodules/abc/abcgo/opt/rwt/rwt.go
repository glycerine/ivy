package rwt

/**CFile****************************************************************

  FileName    [rwt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/extra/extra.h"
//#include "misc/mem/mem.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////
 
#define RWT_LIMIT  1048576/4  // ((1 << 20) 
// Original macro:
// #define RWT_MIN(a,b)      (((a) < (b))? (a) : (b))

// Original macro:
// #define RWT_MAX(a,b)      (((a) > (b))? (a) : (b))


type Rwt_Man_t =   Rwt_Man_t_;
type Rwt_Node_t =  Rwt_Node_t_;

type Rwt_Man_t_ struct {
    // internal lookups
    nFuncs                int32;           // number of four var functions
    unsigned short *   puCanons;         // canonical forms
    pPhases              string;          // canonical phases
    pPerms              string;           // canonical permutations
    unsigned char *    pMap;             // mapping of functions into class numbers
    unsigned short *   pMapInv;          // mapping of classes into functions
    pPractical              string;       // practical NPN classes
    pPerms4             *string;          // four-var permutations
    // node space
    vForest         *Vec_Ptr_t;          // all the nodes
    pTable       **Rwt_Node_t;           // the hash table of nodes by their canonical form
    vClasses         *Vec_Vec_t;         // the nodes of the equivalence classes
    pMmNode       *Mem_Fixed_t;          // memory for nodes and cuts
    // statistical variables
    nTravIds                int32;         // the counter of traversal IDs
    nConsidered                int32;      // the number of nodes considered
    nAdded                int32;           // the number of nodes added to lists
    nClasses                int32;         // the number of NN classes
    // the result of resynthesis
    fCompl                int32;           // indicates if the output of FF should be complemented
    pCut              any;             // the decomposition tree (temporary)
    pGraph              any;           // the decomposition tree (temporary)
    pPerm              string;            // permutation used for the best cut
    vFanins         *Vec_Ptr_t;          // the fanins array (temporary)
    vFaninsCur         *Vec_Ptr_t;       // the fanins array (temporary)
    vLevNums         *Vec_Int_t;         // the array of levels (temporary)
    vNodesTemp         *Vec_Ptr_t;       // the nodes in MFFC (temporary)
    // node statistics
    nNodesConsidered                int32;
    nNodesRewritten                int32;
    nNodesGained                int32;
    nScores [222]int32;
    nCutsGood                int32;
    nCutsBad                int32;
    nSubgraphs                int32;
    // runtime statistics
    timeStart            abctime;
    timeTruth            abctime;
    timeCut            abctime;
    timeRes            abctime;
    timeEval            abctime;
    timeMffc            abctime;
    timeUpdate            abctime;
    timeTotal            abctime;
}

type Rwt_Node_t_ struct { // 24 bytes

    Id                int32;               // ID 
    TravId                int32;           // traversal ID
    uTruth uint32; /* C bitfield: 16 bits; review packing and truncation */      // truth table
    Volume uint32; /* C bitfield: 8 bits; review packing and truncation */      // volume
    Level uint32; /* C bitfield: 6 bits; review packing and truncation */      // level
    fUsed uint32; /* C bitfield: 1 bits; review packing and truncation */      // mark
    fExor uint32; /* C bitfield: 1 bits; review packing and truncation */      // mark
    p0        *Rwt_Node_t;               // first child
    p1        *Rwt_Node_t;               // second child
    pNext        *Rwt_Node_t;            // next in the table
}

// manipulation of complemented attributes
func Rwt_IsComplement(p *Rwt_Node_t) int32 { return int32((((ABC_PTRUINT_T)p) & 01));            }
func Rwt_Regular(p *Rwt_Node_t) *Rwt_Node_t { return (Rwt_Node_t *)((ABC_PTRUINT_T)(p) & ~01);  }
func Rwt_Not(p *Rwt_Node_t) *Rwt_Node_t { return (Rwt_Node_t *)((ABC_PTRUINT_T)(p) ^  01);  }
func Rwt_NotCond(p *Rwt_Node_t, c int32) *Rwt_Node_t { return (Rwt_Node_t *)((ABC_PTRUINT_T)(p) ^ (c));  }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== rwrDec.c ========================================================*/
// extern void              Rwt_ManPreprocess( Rwt_Man_t * p );

/*=== rwrMan.c ========================================================*/
// extern Rwt_Man_t *       Rwt_ManStart( int fPrecompute );

// extern void              Rwt_ManStop( Rwt_Man_t * p );

// extern void              Rwt_ManPrintStats( Rwt_Man_t * p );

// extern void              Rwt_ManPrintStatsFile( Rwt_Man_t * p );

// extern void *            Rwt_ManReadDecs( Rwt_Man_t * p );

// extern Vec_Ptr_t *       Rwt_ManReadLeaves( Rwt_Man_t * p );

// extern int               Rwt_ManReadCompl( Rwt_Man_t * p );

// extern void              Rwt_ManAddTimeCuts( Rwt_Man_t * p, abctime Time );

// extern void              Rwt_ManAddTimeUpdate( Rwt_Man_t * p, abctime Time );

// extern void              Rwt_ManAddTimeTotal( Rwt_Man_t * p, abctime Time );

/*=== rwrUtil.c ========================================================*/
// extern void              Rwt_ManLoadFromArray( Rwt_Man_t * p, int fVerbose );

// extern char *            Rwt_ManGetPractical( Rwt_Man_t * p );

// extern Rwt_Node_t *      Rwt_ManAddVar( Rwt_Man_t * p, unsigned uTruth, int fPrecompute );

// extern void              Rwt_ManIncTravId( Rwt_Man_t * p );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

