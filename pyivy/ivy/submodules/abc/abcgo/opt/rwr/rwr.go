package rwr

/**CFile****************************************************************

  FileName    [rwr.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwr.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "base/abc/abc.h"
//#include "opt/cut/cut.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////
 
const RWR_LIMIT = 1048576 / 4

type Rwr_Man_t =   Rwr_Man_t_;
type Rwr_Node_t =  Rwr_Node_t_;

type Rwr_Man_t_ struct {
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
    pTable       **Rwr_Node_t;           // the hash table of nodes by their canonical form
    vClasses         *Vec_Vec_t;         // the nodes of the equivalence classes
    pMmNode   *Extra_MmFixed_t;          // memory for nodes and cuts
    // statistical variables
    nTravIds                int32;         // the counter of traversal IDs
    nConsidered                int32;      // the number of nodes considered
    nAdded                int32;           // the number of nodes added to lists
    nClasses                int32;         // the number of NN classes
    // the result of resynthesis
    fCompl                int32;           // indicates if the output of FF should be complemented
    pGraph              any;           // the decomposition tree (temporary)
    vFanins         *Vec_Ptr_t;          // the fanins array (temporary)
    vFaninsCur         *Vec_Ptr_t;       // the fanins array (temporary)
    vLevNums         *Vec_Int_t;         // the array of levels (temporary)
    vNodesTemp         *Vec_Ptr_t;       // the nodes in MFFC (temporary)
    // node statistics
    nNodesConsidered                int32;
    nNodesRewritten                int32;
    nNodesGained                int32;
    nNodesBeg                int32;
    nNodesEnd                int32;
    nScores [222]int32;
    nCutsGood                int32;
    nCutsBad                int32;
    nSubgraphs                int32;
    // runtime statistics
    timeStart            abctime;
    timeCut            abctime;
    timeRes            abctime;
    timeEval            abctime;
    timeMffc            abctime;
    timeUpdate            abctime;
    timeTotal            abctime;
}

type Rwr_Node_t_ struct { // 24 bytes

    Id                int32;               // ID 
    TravId                int32;           // traversal ID
    nScore              int16;
    nGain              int16;
    nAdded              int16;
    uTruth uint32; /* C bitfield: 16 bits; review packing and truncation */      // truth table
    Volume uint32; /* C bitfield: 8 bits; review packing and truncation */      // volume
    Level uint32; /* C bitfield: 6 bits; review packing and truncation */      // level
    fUsed uint32; /* C bitfield: 1 bits; review packing and truncation */      // mark
    fExor uint32; /* C bitfield: 1 bits; review packing and truncation */      // mark
    p0        *Rwr_Node_t;               // first child
    p1        *Rwr_Node_t;               // second child
    pNext        *Rwr_Node_t;            // next in the table
}

// manipulation of complemented attributes
func Rwr_IsComplement(p *Rwr_Node_t) int32 { return int32( (((ABC_PTRUINT_T)p) & 01));           }
func Rwr_Regular(p *Rwr_Node_t) *Rwr_Node_t { return (Rwr_Node_t *)((ABC_PTRUINT_T)(p) & ~01);  }
func Rwr_Not(p *Rwr_Node_t) *Rwr_Node_t { return (Rwr_Node_t *)((ABC_PTRUINT_T)(p) ^  01);  }
func Rwr_NotCond(p *Rwr_Node_t, c int32) *Rwr_Node_t { return (Rwr_Node_t *)((ABC_PTRUINT_T)(p) ^ (c));  }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== rwrDec.c ========================================================*/
// extern void              Rwr_ManPreprocess( Rwr_Man_t * p );

/*=== rwrEva.c ========================================================*/
// extern int               Rwr_NodeRewrite( Rwr_Man_t * p, Cut_Man_t * pManCut, Abc_Obj_t * pNode, int fUpdateLevel, int fUseZeros, int fPlaceEnable );

// extern void              Rwr_ScoresClean( Rwr_Man_t * p );

// extern void              Rwr_ScoresReport( Rwr_Man_t * p );

/*=== rwrLib.c ========================================================*/
// extern void              Rwr_ManPrecompute( Rwr_Man_t * p );

// extern Rwr_Node_t *      Rwr_ManAddVar( Rwr_Man_t * p, unsigned uTruth, int fPrecompute );

// extern Rwr_Node_t *      Rwr_ManAddNode( Rwr_Man_t * p, Rwr_Node_t * p0, Rwr_Node_t * p1, int fExor, int Level, int Volume );

// extern int               Rwr_ManNodeVolume( Rwr_Man_t * p, Rwr_Node_t * p0, Rwr_Node_t * p1 );

// extern void              Rwr_ManIncTravId( Rwr_Man_t * p );

/*=== rwrMan.c ========================================================*/
// extern Rwr_Man_t *       Rwr_ManStart( int  fPrecompute );

// extern void              Rwr_ManStop( Rwr_Man_t * p );

// extern void              Rwr_ManPrintStats( Rwr_Man_t * p );

// extern void              Rwr_ManPrintStatsFile( Rwr_Man_t * p );

// extern void *            Rwr_ManReadDecs( Rwr_Man_t * p );

// extern Vec_Ptr_t *       Rwr_ManReadLeaves( Rwr_Man_t * p );

// extern int               Rwr_ManReadCompl( Rwr_Man_t * p );

// extern void              Rwr_ManAddTimeCuts( Rwr_Man_t * p, abctime Time );

// extern void              Rwr_ManAddTimeUpdate( Rwr_Man_t * p, abctime Time );

// extern void              Rwr_ManAddTimeTotal( Rwr_Man_t * p, abctime Time );

/*=== rwrPrint.c ========================================================*/
// extern void              Rwr_ManPrint( Rwr_Man_t * p );

/*=== rwrUtil.c ========================================================*/
// extern void              Rwr_ManWriteToArray( Rwr_Man_t * p );

// extern void              Rwr_ManLoadFromArray( Rwr_Man_t * p, int fVerbose );

// extern void              Rwr_ManWriteToFile( Rwr_Man_t * p, char * pFileName );

// extern void              Rwr_ManLoadFromFile( Rwr_Man_t * p, char * pFileName );

// extern void              Rwr_ListAddToTail( Rwr_Node_t ** ppList, Rwr_Node_t * pNode );

// extern char *            Rwr_ManGetPractical( Rwr_Man_t * p );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

