package fraig

/**CFile****************************************************************

  FileName    [fraig.h]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [External declarations of the FRAIG package.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraig.h,v 1.18 2005/07/08 01:01:30 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

type Fraig_Man_t = Fraig_ManStruct_t_
type Fraig_Node_t = Fraig_NodeStruct_t_
type Fraig_NodeVec_t = Fraig_NodeVecStruct_t_
type Fraig_HashTable_t = Fraig_HashTableStruct_t_
type Fraig_Params_t = Fraig_ParamsStruct_t_
type Prove_Params_t = Prove_ParamsStruct_t_

type Fraig_ParamsStruct_t_ struct {
	nPatsRand  int32       // the number of words of random simulation info
	nPatsDyna  int32       // the number of words of dynamic simulation info
	nBTLimit   int32       // the max number of backtracks to perform
	nSeconds   int32       // the timeout for the final proof
	fFuncRed   int32       // performs only one level hashing
	fFeedBack  int32       // enables solver feedback
	fDist1Pats int32       // enables distance-1 patterns
	fDoSparse  int32       // performs equiv tests for sparse functions
	fChoicing  int32       // enables recording structural choices
	fTryProve  int32       // tries to solve the final miter
	fVerbose   int32       // the verbosiness flag
	fVerboseP  int32       // the verbosiness flag (for proof reporting)
	fInternal  int32       // is set to 1 for internal fraig calls
	nConfLimit int32       // the limit on the number of conflicts
	nInspLimit ABC_INT64_T // the limit on the number of inspections
}

type Prove_ParamsStruct_t_ struct {
	// general parameters
	fUseFraiging  int32 // enables fraiging
	fUseRewriting int32 // enables rewriting
	fUseBdds      int32 // enables BDD construction when other methods fail
	fVerbose      int32 // prints verbose stats
	// iterations
	nItersMax int32 // the number of iterations
	// mitering
	nMiteringLimitStart int32   // starting mitering limit
	nMiteringLimitMulti float32 // multiplicative coefficient to increase the limit in each iteration
	// rewriting
	nRewritingLimitStart int32   // the number of rewriting iterations
	nRewritingLimitMulti float32 // multiplicative coefficient to increase the limit in each iteration
	// fraiging
	nFraigingLimitStart int32   // starting backtrack(conflict) limit
	nFraigingLimitMulti float32 // multiplicative coefficient to increase the limit in each iteration
	// last-gasp BDD construction
	nBddSizeLimit int32 // the number of BDD nodes when construction is aborted
	fBddReorder   int32 // enables dynamic BDD variable reordering
	// last-gasp mitering
	nMiteringLimitLast int32 // final mitering limit
	// global SAT solver limits
	nTotalBacktrackLimit ABC_INT64_T // global limit on the number of backtracks
	nTotalInspectLimit   ABC_INT64_T // global limit on the number of clause inspects
	// global resources applied
	nTotalBacktracksMade ABC_INT64_T // the total number of backtracks made
	nTotalInspectsMade   ABC_INT64_T // the total number of inspects made
}

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

// macros working with complemented attributes of the nodes
// Original macro:
// #define Fraig_IsComplement(p)    (((int)((ABC_PTRUINT_T) (p) & 01)))

// Original macro:
// #define Fraig_Regular(p)         ((Fraig_Node_t *)((ABC_PTRUINT_T)(p) & ~01))

// Original macro:
// #define Fraig_Not(p)             ((Fraig_Node_t *)((ABC_PTRUINT_T)(p) ^ 01))

// Original macro:
// #define Fraig_NotCond(p,c)       ((Fraig_Node_t *)((ABC_PTRUINT_T)(p) ^ (c)))

// these are currently not used
// Original macro:
// #define Fraig_Ref(p)

// Original macro:
// #define Fraig_Deref(p)

// Original macro:
// #define Fraig_RecursiveDeref(p,c)

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== fraigApi.c =============================================================*/
// extern Fraig_NodeVec_t *   Fraig_ManReadVecInputs( Fraig_Man_t * p );

// extern Fraig_NodeVec_t *   Fraig_ManReadVecOutputs( Fraig_Man_t * p );

// extern Fraig_NodeVec_t *   Fraig_ManReadVecNodes( Fraig_Man_t * p );

// extern Fraig_Node_t **     Fraig_ManReadInputs ( Fraig_Man_t * p );

// extern Fraig_Node_t **     Fraig_ManReadOutputs( Fraig_Man_t * p );

// extern Fraig_Node_t **     Fraig_ManReadNodes( Fraig_Man_t * p );

// extern int                 Fraig_ManReadInputNum ( Fraig_Man_t * p );

// extern int                 Fraig_ManReadOutputNum( Fraig_Man_t * p );

// extern int                 Fraig_ManReadNodeNum( Fraig_Man_t * p );

// extern Fraig_Node_t *      Fraig_ManReadConst1 ( Fraig_Man_t * p );

// extern Fraig_Node_t *      Fraig_ManReadIthVar( Fraig_Man_t * p, int i );

// extern Fraig_Node_t *      Fraig_ManReadIthNode( Fraig_Man_t * p, int i );

// extern char **             Fraig_ManReadInputNames( Fraig_Man_t * p );

// extern char **             Fraig_ManReadOutputNames( Fraig_Man_t * p );

// extern char *              Fraig_ManReadVarsInt( Fraig_Man_t * p );

// extern char *              Fraig_ManReadSat( Fraig_Man_t * p );

// extern int                 Fraig_ManReadFuncRed( Fraig_Man_t * p );

// extern int                 Fraig_ManReadFeedBack( Fraig_Man_t * p );

// extern int                 Fraig_ManReadDoSparse( Fraig_Man_t * p );

// extern int                 Fraig_ManReadChoicing( Fraig_Man_t * p );

// extern int                 Fraig_ManReadVerbose( Fraig_Man_t * p );

// extern int *               Fraig_ManReadModel( Fraig_Man_t * p );

// extern int                 Fraig_ManReadPatternNumRandom( Fraig_Man_t * p );

// extern int                 Fraig_ManReadPatternNumDynamic( Fraig_Man_t * p );

// extern int                 Fraig_ManReadPatternNumDynamicFiltered( Fraig_Man_t * p );

// extern int                 Fraig_ManReadSatFails( Fraig_Man_t * p );

// extern int                 Fraig_ManReadConflicts( Fraig_Man_t * p );

// extern int                 Fraig_ManReadInspects( Fraig_Man_t * p );

// extern void                Fraig_ManSetFuncRed( Fraig_Man_t * p, int fFuncRed );

// extern void                Fraig_ManSetFeedBack( Fraig_Man_t * p, int fFeedBack );

// extern void                Fraig_ManSetDoSparse( Fraig_Man_t * p, int fDoSparse );

// extern void                Fraig_ManSetChoicing( Fraig_Man_t * p, int fChoicing );

// extern void                Fraig_ManSetTryProve( Fraig_Man_t * p, int fTryProve );

// extern void                Fraig_ManSetVerbose( Fraig_Man_t * p, int fVerbose );

// extern void                Fraig_ManSetOutputNames( Fraig_Man_t * p, char ** ppNames );

// extern void                Fraig_ManSetInputNames( Fraig_Man_t * p, char ** ppNames );

// extern void                Fraig_ManSetPo( Fraig_Man_t * p, Fraig_Node_t * pNode );

// extern Fraig_Node_t *      Fraig_NodeReadData0( Fraig_Node_t * p );

// extern Fraig_Node_t *      Fraig_NodeReadData1( Fraig_Node_t * p );

// extern int                 Fraig_NodeReadNum( Fraig_Node_t * p );

// extern Fraig_Node_t *      Fraig_NodeReadOne( Fraig_Node_t * p );

// extern Fraig_Node_t *      Fraig_NodeReadTwo( Fraig_Node_t * p );

// extern Fraig_Node_t *      Fraig_NodeReadNextE( Fraig_Node_t * p );

// extern Fraig_Node_t *      Fraig_NodeReadRepr( Fraig_Node_t * p );

// extern int                 Fraig_NodeReadNumRefs( Fraig_Node_t * p );

// extern int                 Fraig_NodeReadNumFanouts( Fraig_Node_t * p );

// extern int                 Fraig_NodeReadSimInv( Fraig_Node_t * p );

// extern int                 Fraig_NodeReadNumOnes( Fraig_Node_t * p );

// extern unsigned *          Fraig_NodeReadPatternsRandom( Fraig_Node_t * p );

// extern unsigned *          Fraig_NodeReadPatternsDynamic( Fraig_Node_t * p );

// extern void                Fraig_NodeSetData0( Fraig_Node_t * p, Fraig_Node_t * pData );

// extern void                Fraig_NodeSetData1( Fraig_Node_t * p, Fraig_Node_t * pData );

// extern int                 Fraig_NodeIsConst( Fraig_Node_t * p );

// extern int                 Fraig_NodeIsVar( Fraig_Node_t * p );

// extern int                 Fraig_NodeIsAnd( Fraig_Node_t * p );

// extern int                 Fraig_NodeComparePhase( Fraig_Node_t * p1, Fraig_Node_t * p2 );

// extern Fraig_Node_t *      Fraig_NodeOr( Fraig_Man_t * p, Fraig_Node_t * p1, Fraig_Node_t * p2 );

// extern Fraig_Node_t *      Fraig_NodeAnd( Fraig_Man_t * p, Fraig_Node_t * p1, Fraig_Node_t * p2 );

// extern Fraig_Node_t *      Fraig_NodeOr( Fraig_Man_t * p, Fraig_Node_t * p1, Fraig_Node_t * p2 );

// extern Fraig_Node_t *      Fraig_NodeExor( Fraig_Man_t * p, Fraig_Node_t * p1, Fraig_Node_t * p2 );

// extern Fraig_Node_t *      Fraig_NodeMux( Fraig_Man_t * p, Fraig_Node_t * pNode, Fraig_Node_t * pNodeT, Fraig_Node_t * pNodeE );

// extern void                Fraig_NodeSetChoice( Fraig_Man_t * pMan, Fraig_Node_t * pNodeOld, Fraig_Node_t * pNodeNew );

/*=== fraigMan.c =============================================================*/
// extern void                Prove_ParamsSetDefault( Prove_Params_t * pParams );

// extern void                Fraig_ParamsSetDefault( Fraig_Params_t * pParams );

// extern void                Fraig_ParamsSetDefaultFull( Fraig_Params_t * pParams );

// extern Fraig_Man_t *       Fraig_ManCreate( Fraig_Params_t * pParams );

// extern void                Fraig_ManFree( Fraig_Man_t * pMan );

// extern void                Fraig_ManPrintStats( Fraig_Man_t * p );

// extern Fraig_NodeVec_t *   Fraig_ManGetSimInfo( Fraig_Man_t * p );

// extern int                 Fraig_ManCheckClauseUsingSimInfo( Fraig_Man_t * p, Fraig_Node_t * pNode1, Fraig_Node_t * pNode2 );

// extern void                Fraig_ManAddClause( Fraig_Man_t * p, Fraig_Node_t ** ppNodes, int nNodes );

/*=== fraigDfs.c =============================================================*/
// extern Fraig_NodeVec_t *   Fraig_Dfs( Fraig_Man_t * pMan, int fEquiv );

// extern Fraig_NodeVec_t *   Fraig_DfsOne( Fraig_Man_t * pMan, Fraig_Node_t * pNode, int fEquiv );

// extern Fraig_NodeVec_t *   Fraig_DfsNodes( Fraig_Man_t * pMan, Fraig_Node_t ** ppNodes, int nNodes, int fEquiv );

// extern Fraig_NodeVec_t *   Fraig_DfsReverse( Fraig_Man_t * pMan );

// extern int                 Fraig_CountNodes( Fraig_Man_t * pMan, int fEquiv );

// extern int                 Fraig_CheckTfi( Fraig_Man_t * pMan, Fraig_Node_t * pOld, Fraig_Node_t * pNew );

// extern int                 Fraig_CountLevels( Fraig_Man_t * pMan );

/*=== fraigSat.c =============================================================*/
// extern int                 Fraig_NodesAreEqual( Fraig_Man_t * p, Fraig_Node_t * pNode1, Fraig_Node_t * pNode2, int nBTLimit, int nTimeLimit );

// extern int                 Fraig_NodeIsEquivalent( Fraig_Man_t * p, Fraig_Node_t * pOld, Fraig_Node_t * pNew, int nBTLimit, int nTimeLimit );

// extern void                Fraig_ManProveMiter( Fraig_Man_t * p );

// extern int                 Fraig_ManCheckMiter( Fraig_Man_t * p );

// extern int                 Fraig_ManCheckClauseUsingSat( Fraig_Man_t * p, Fraig_Node_t * pNode1, Fraig_Node_t * pNode2, int nBTLimit );

/*=== fraigVec.c ===============================================================*/
// extern Fraig_NodeVec_t *   Fraig_NodeVecAlloc( int nCap );

// extern void                Fraig_NodeVecFree( Fraig_NodeVec_t * p );

// extern Fraig_NodeVec_t *   Fraig_NodeVecDup( Fraig_NodeVec_t * p );

// extern Fraig_Node_t **     Fraig_NodeVecReadArray( Fraig_NodeVec_t * p );

// extern int                 Fraig_NodeVecReadSize( Fraig_NodeVec_t * p );

// extern void                Fraig_NodeVecGrow( Fraig_NodeVec_t * p, int nCapMin );

// extern void                Fraig_NodeVecShrink( Fraig_NodeVec_t * p, int nSizeNew );

// extern void                Fraig_NodeVecClear( Fraig_NodeVec_t * p );

// extern void                Fraig_NodeVecPush( Fraig_NodeVec_t * p, Fraig_Node_t * Entry );

// extern int                 Fraig_NodeVecPushUnique( Fraig_NodeVec_t * p, Fraig_Node_t * Entry );

// extern void                Fraig_NodeVecPushOrder( Fraig_NodeVec_t * p, Fraig_Node_t * pNode );

// extern int                 Fraig_NodeVecPushUniqueOrder( Fraig_NodeVec_t * p, Fraig_Node_t * pNode );

// extern void                Fraig_NodeVecPushOrderByLevel( Fraig_NodeVec_t * p, Fraig_Node_t * pNode );

// extern int                 Fraig_NodeVecPushUniqueOrderByLevel( Fraig_NodeVec_t * p, Fraig_Node_t * pNode );

// extern Fraig_Node_t *      Fraig_NodeVecPop( Fraig_NodeVec_t * p );

// extern void                Fraig_NodeVecRemove( Fraig_NodeVec_t * p, Fraig_Node_t * Entry );

// extern void                Fraig_NodeVecWriteEntry( Fraig_NodeVec_t * p, int i, Fraig_Node_t * Entry );

// extern Fraig_Node_t *      Fraig_NodeVecReadEntry( Fraig_NodeVec_t * p, int i );

// extern void                Fraig_NodeVecSortByLevel( Fraig_NodeVec_t * p, int fIncreasing );

// extern void                Fraig_NodeVecSortByNumber( Fraig_NodeVec_t * p );

/*=== fraigUtil.c ===============================================================*/
// extern void                Fraig_ManMarkRealFanouts( Fraig_Man_t * p );

// extern int                 Fraig_ManCheckConsistency( Fraig_Man_t * p );

// extern int                 Fraig_GetMaxLevel( Fraig_Man_t * pMan );

// extern void                Fraig_ManReportChoices( Fraig_Man_t * pMan );

// extern void                Fraig_MappingSetChoiceLevels( Fraig_Man_t * pMan, int fMaximum );

// extern Fraig_NodeVec_t *   Fraig_CollectSupergate( Fraig_Node_t * pNode, int fStopAtMux );

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_END
