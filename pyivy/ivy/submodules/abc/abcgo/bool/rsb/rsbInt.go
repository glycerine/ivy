package rsb

/**CFile****************************************************************

  FileName    [rsbInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Truth-table based resubstitution.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rsbInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/util/utilTruth.h"
//#include "bool/kit/kit.h"
//#include "rsb.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// truth table computation manager
type Rsb_Man_t_ struct {
    // parameters
    nLeafMax                int32;     // the max number of leaves of a cut
    nDivMax                int32;      // the max number of divisors to collect
    nDecMax                int32;      // the max number of decompositions
    fVerbose                int32;     // verbosity level
    // decomposition
    vCexes         *Vec_Wrd_t;       // counter-examples
    vDecPats         *Vec_Int_t;     // decomposition patterns
    vFanins         *Vec_Int_t;      // the result of decomposition
    vFaninsOld         *Vec_Int_t;   // original fanins
    vTries         *Vec_Int_t;       // intermediate
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== rsbMan.c ==========================================================*/

//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

