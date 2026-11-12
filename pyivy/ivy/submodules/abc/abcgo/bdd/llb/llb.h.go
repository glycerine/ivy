package llb

/**CFile****************************************************************

  FileName    [llb.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD-based reachability.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 8, 2010.]

  Revision    [$Id: llb.h,v 1.00 2010/05/08 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Gia_ParLlb_t struct {
	nBddMax       int32 // maximum BDD size
	nIterMax      int32 // maximum iteration count
	nClusterMax   int32 // maximum cluster size
	nHintDepth    int32 // the number of times to cofactor
	HintFirst     int32 // the number of first hint to use
	fUseFlow      int32 // use flow computation
	nVolumeMax    int32 // the largest volume
	nVolumeMin    int32 // the smallest volume
	nPartValue    int32 // partitioning value
	fBackward     int32 // enable backward reachability
	fReorder      int32 // enable dynamic variable reordering
	fIndConstr    int32 // extract inductive constraints
	fUsePivots    int32 // use internal pivot variables
	fCluster      int32 // use partition clustering
	fSchedule     int32 // use cluster scheduling
	fDumpReached  int32 // dump reached states into a file
	fVerbose      int32 // print verbose information
	fVeryVerbose  int32 // print dependency matrices
	fSilent       int32 // do not print any infomation
	fSkipReach    int32 // skip reachability (preparation phase only)
	fSkipOutCheck int32 // does not check the property output
	TimeLimit     int32 // time limit for one reachability run
	TimeLimitGlo  int32 // time limit for all reachability runs
	// internal parameters
	TimeTarget abctime // the time to stop
	iFrame     int32   // explored up to this frame
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== llbCore.c ==========================================================*/
// extern void     Llb_ManSetDefaultParams( Gia_ParLlb_t * pPars );

/*=== llb4Nonlin.c ==========================================================*/
// extern int      Llb_Nonlin4CoreReach( Aig_Man_t * pAig, Gia_ParLlb_t * pPars );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
