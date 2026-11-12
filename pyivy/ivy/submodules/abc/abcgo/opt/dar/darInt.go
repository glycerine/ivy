package dar

/**CFile****************************************************************

  FileName    [darInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: darInt.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "aig/aig/aig.h"
//#include "dar.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START
 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Dar_Man_t =            Dar_Man_t_;
              

// the AIG 4-cut
type Dar_Cut_t struct {  // 6 words

    uSign         uint32;          // cut signature
    uTruth uint32; /* C bitfield: 16 bits; review packing and truncation */   // the truth table of the cut function
    Value uint32; /* C bitfield: 11 bits; review packing and truncation */   // the value of the cut 
    fBest uint32; /* C bitfield: 1 bits; review packing and truncation */   // marks the best cut
    fUsed uint32; /* C bitfield: 1 bits; review packing and truncation */   // marks the cut currently in use
    nLeaves uint32; /* C bitfield: 3 bits; review packing and truncation */   // the number of leaves
    pLeaves [4]int32;     // the array of leaves
}

// the AIG manager
type Dar_Man_t_ struct {
    // input data
    pPars    *Dar_RwrPar_t;          // rewriting parameters
    pAig       *Aig_Man_t;           // AIG manager 
    // various data members
    pMemCuts   *Aig_MmFixed_t;       // memory manager for cuts
    pManCnf            any;        // CNF managers
    vCutNodes       *Vec_Ptr_t;      // the nodes with cuts allocated
    // current rewriting step
    vLeavesBest       *Vec_Ptr_t;    // the best set of leaves
    OutBest              int32;        // the best output (in the library)
    OutNumBest              int32;     // the best number of the output
    GainBest              int32;       // the best gain
    LevelBest              int32;      // the level of node with the best gain
    ClassBest              int32;      // the equivalence class of the best replacement
    // function statistics
    nTotalSubgs              int32;    // the total number of subgraphs tried
    ClassTimes [222]int32;// the runtimes for each class
    ClassGains [222]int32;// the gains for each class
    ClassSubgs [222]int32;// the graphs for each class
    nCutMemUsed              int32;    // memory used for cuts
    // rewriting statistics
    nNodesInit              int32;     // the original number of nodes
    nNodesTried              int32;    // the number of nodes attempted
    nCutsAll              int32;       // all cut pairs
    nCutsTried              int32;     // computed cuts
    nCutsUsed              int32;      // used cuts
    nCutsBad              int32;       // bad cuts due to absent fanin
    nCutsGood              int32;      // good cuts
    nCutsSkipped              int32;   // skipped bad cuts
    // timing statistics
    timeCuts          abctime;
    timeEval          abctime;
    timeOther          abctime;
    timeTotal          abctime;
    time1          abctime;
    time2          abctime;
}

func Dar_ObjCuts(pObj *Aig_Obj_t) *Dar_Cut_t { return (Dar_Cut_t *)pObj.pData;    }
func Dar_ObjSetCuts(pObj *Aig_Obj_t, pCuts *Dar_Cut_t) { assert(!Aig_ObjIsNone(pObj), "!Aig_ObjIsNone(pObj)"); pObj.pData = pCuts;   }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                          ITERATORS                               ///
////////////////////////////////////////////////////////////////////////

// iterator over all cuts of the node
// Original macro:
// #define Dar_ObjForEachCutAll( pObj, pCut, i )                                   \
//     for ( (pCut) = Dar_ObjCuts(pObj), i = 0; i < (int)(pObj)->nCuts; i++, pCut++ ) 

// Original macro:
// #define Dar_ObjForEachCut( pObj, pCut, i )                                      \
//     for ( (pCut) = Dar_ObjCuts(pObj), i = 0; i < (int)(pObj)->nCuts; i++, pCut++ ) if ( (pCut)->fUsed==0 ) {} else

// iterator over leaves of the cut
// Original macro:
// #define Dar_CutForEachLeaf( p, pCut, pLeaf, i )                                 \
//     for ( i = 0; (i < (int)(pCut)->nLeaves) && (((pLeaf) = Aig_ManObj(p, (pCut)->pLeaves[i])), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== darBalance.c ========================================================*/
/*=== darCore.c ===========================================================*/
/*=== darCut.c ============================================================*/
// extern void            Dar_ManCutsRestart( Dar_Man_t * p, Aig_Obj_t * pRoot );

// extern void            Dar_ManCutsFree( Dar_Man_t * p );

// extern Dar_Cut_t *     Dar_ObjPrepareCuts( Dar_Man_t * p, Aig_Obj_t * pObj );

// extern Dar_Cut_t *     Dar_ObjComputeCuts_rec( Dar_Man_t * p, Aig_Obj_t * pObj );

// extern Dar_Cut_t *     Dar_ObjComputeCuts( Dar_Man_t * p, Aig_Obj_t * pObj, int fSkipTtMin );

// extern void            Dar_ObjCutPrint( Aig_Man_t * p, Aig_Obj_t * pObj );

/*=== darData.c ===========================================================*/
// extern Vec_Int_t *     Dar_LibReadNodes();

// extern Vec_Int_t *     Dar_LibReadOuts();

// extern Vec_Int_t *     Dar_LibReadPrios();

/*=== darLib.c ============================================================*/
// extern void            Dar_LibStart();

// extern void            Dar_LibStop();

// extern void            Dar_LibReturnCanonicals( unsigned * pCanons );

// extern void            Dar_LibEval( Dar_Man_t * p, Aig_Obj_t * pRoot, Dar_Cut_t * pCut, int Required, int * pnMffcSize );

// extern Aig_Obj_t *     Dar_LibBuildBest( Dar_Man_t * p );

/*=== darMan.c ============================================================*/
// extern Dar_Man_t *     Dar_ManStart( Aig_Man_t * pAig, Dar_RwrPar_t * pPars );

// extern void            Dar_ManStop( Dar_Man_t * p );

// extern void            Dar_ManPrintStats( Dar_Man_t * p );

/*=== darPrec.c ============================================================*/
// extern char **         Dar_Permutations( int n );

// extern void            Dar_Truth4VarNPN( unsigned short ** puCanons, char ** puPhases, char ** puPerms, unsigned char ** puMap );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

