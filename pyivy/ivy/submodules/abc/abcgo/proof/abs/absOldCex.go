package abs

/**CFile****************************************************************

  FileName    [saigAbsCba.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [CEX-based abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigAbsCba.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// local manager
   


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Selects the best flops from the given array.]

  Description [Selects the best 'nFfsToSelect' flops among the array 
  'vAbsFfsToAdd' of flops that should be added to the abstraction.
  To this end, this procedure simulates the original AIG (pAig) using
  the given CEX (pAbsCex), which was detected for the abstraction.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Duplicate with literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Maps array of frame PI IDs into array of additional PI IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns reasons for the property to fail.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Collect nodes in the unrolled timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Creates refinement manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Destroys refinement manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [SAT-based refinement of the counter-example.]

  Description [The first parameter (nInputs) indicates how many first 
  primary inputs to skip without considering as care candidates.]
               

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Checks the abstracted model for a counter-example.]

  Description [Returns the array of abstracted flops that should be added
  to the abstraction.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

