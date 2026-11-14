package lpk

/**CFile****************************************************************

  FileName    [lpk.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpk.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

type Lpk_Par_t struct {
	// user-controlled parameters
	NLutsMax     int32 // (N) the maximum number of LUTs in the structure
	NLutsOver    int32 // (Q) the maximum number of LUTs not in the MFFC
	NVarsShared  int32 // (S) the maximum number of shared variables (crossbars)
	NGrowthLevel int32 // (L) the maximum increase in the node level after resynthesis
	FSatur       int32 // iterate till saturation
	FZeroCost    int32 // accept zero-cost replacements
	FFirst       int32 // use root node and first cut only
	FOldAlgo     int32 // use old algorithm
	FVerbose     int32 // the verbosiness flag
	FVeryVerbose int32 // additional verbose info printout
	// internal parameters
	NLutSize int32 // (K) the LUT size (determined by the input network)
	NVarsMax int32 // (V) the largest number of variables: V = N * (K-1) + 1
}

// Lpk_AbcAPI supplies the abci helpers that LPK needs without making the
// optimization package depend back on base/abci.
type Lpk_AbcAPI struct {
	NtkLutmin     func(*Abc_Ntk_t, int32, int32) *Abc_Ntk_t
	NodeFromIfRec func(*Abc_Ntk_t, *If_Man_t, *If_Obj_t, *Vec_Int_t) *Abc_Obj_t
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                           ITERATORS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== lpkCore.c ========================================================*/
// extern int     Lpk_Resynthesize( Abc_Ntk_t * pNtk, Lpk_Par_t * pPars );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
