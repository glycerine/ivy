package fxu

/**CFile****************************************************************

  FileName    [fxu.h]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [External declarations of fast extract for unate covers.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: fxu.h,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/vec/vec.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

     

// structure for the FX input/output data 
type Fxu_Data_t struct {
    // user specified parameters
    fOnlyS               int32;           // set to 1 to have only single-cube divs
    fOnlyD               int32;           // set to 1 to have only double-cube divs
    fUse0               int32;            // set to 1 to have 0-weight also extracted
    fUseCompl               int32;        // set to 1 to have complement taken into account
    fVerbose               int32;         // set to 1 to have verbose output
    fVeryVerbose               int32;     // set to 1 to have more verbose output
    nNodesExt               int32;        // the number of divisors to extract
    nSingleMax               int32;       // the max number of single-cube divisors to consider
    nPairsMax               int32;        // the max number of double-cube divisors to consider
    WeightMin               int32;        // the min weight of a divisor to extract
    LitCountMax               int32;      // the max literal count of a divisor to consider
    fCanonDivs               int32;       // use only canonical divisors (AND/XOR/MUX) 
    // the input information
    vSops        *Vec_Ptr_t;            // the SOPs for each node in the network
    vFanins        *Vec_Ptr_t;          // the fanins of each node in the network
    // output information
    vSopsNew        *Vec_Ptr_t;         // the SOPs for each node in the network after extraction
    vFaninsNew        *Vec_Ptr_t;       // the fanins of each node in the network after extraction
    // the SOP manager
    pManSop       *Mem_Flex_t;
    // statistics   
    nNodesOld               int32;        // the old number of nodes
    nNodesNew               int32;        // the number of divisors actually extracted
}

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*===== fxu.c ==========================================================*/
// extern void   Abc_NtkSetDefaultFxParams( Fxu_Data_t * p );

// extern int    Abc_NtkFastExtract( Abc_Ntk_t * pNtk, Fxu_Data_t * p );

// extern void   Abc_NtkFxuFreeInfo( Fxu_Data_t * p );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

