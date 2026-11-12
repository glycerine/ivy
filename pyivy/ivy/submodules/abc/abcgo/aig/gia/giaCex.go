package gia

/**CFile****************************************************************

  FileName    [giaAbs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Counter-example-guided abstraction refinement.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAbs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//#include "sat/bsat/satSolver.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Determines the failed PO when its exact frame is not known.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Starts the process of returning values for internal nodes.]

  Description [Should be called when pCex is available, before probing
  any object for its value using Gia_ManCounterExampleValueLookup().]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stops the process of returning values for internal nodes.]

  Description [Should be called when probing is no longer needed]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns the value of the given object in the given timeframe.]

  Description [Should be called to probe the value of an object with
  the given ID (iFrame is a 0-based number of a time frame - should not
  exceed the number of timeframes in the original counter-example).]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Procedure to test the above code.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns CEX containing PI+CS values for each timeframe.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns CEX containing all object valuess for each timeframe.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
