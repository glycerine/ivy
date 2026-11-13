package acb

/**CFile****************************************************************

  FileName    [acbPar.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: acbPar.h,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

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

type Acb_Par_t struct {
	nLutSize     int32 // LUT size
	nTfoLevMax   int32 // the maximum fanout levels
	nTfiLevMax   int32 // the maximum fanin levels
	nFanoutMax   int32 // the maximum number of fanouts
	nWinNodeMax  int32 // the maximum number of nodes in the window
	nGrowthLevel int32 // the maximum allowed growth in level
	nBTLimit     int32 // the maximum number of conflicts in one SAT run
	nNodesMax    int32 // the maximum number of nodes to try
	fUseAshen    int32 // user Ashenhurst decomposition
	iNodeOne     int32 // one particular node to try
	fArea        int32 // performs optimization for area
	fMoreEffort  int32 // performs high-affort minimization
	fVerbose     int32 // enable basic stats
	fVeryVerbose int32 // enable detailed stats
}

/*=== acbAbc.c =============================================================*/
// extern void   Acb_ParSetDefault( Acb_Par_t * pPars );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

// Acb_ParSetDefault resets the optimization parameters to ABC defaults.
func Acb_ParSetDefault(p *Acb_Par_t) {
	*p = Acb_Par_t{nLutSize: 4, nTfoLevMax: 2, nTfiLevMax: 3, nFanoutMax: 20, nWinNodeMax: 100, fArea: 1}
}
