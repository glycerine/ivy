package gia

/**CFile****************************************************************

  FileName    [giaSweeper.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Incremental SAT sweeper.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSweeper.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "base/main/main.h"
//#include "sat/bsat/satSolver.h"
//#include "proof/ssc/ssc.h"

//ABC_NAMESPACE_IMPL_START

/*

SAT sweeping/equivalence checking requires the following steps:
- Creating probes
  These APIs should be called for all internal points in the logic, which may be used as
      - nodes representing conditions to be used as constraints
      - nodes representing functions to be equivalence checked
      - nodes representing functions needed by the user at the end of SAT sweeping
  Creating new probe using  Gia_SweeperProbeCreate():   int Gia_SweeperProbeCreate( Gia_Man_t * p, int iLit );
  Delete existing probe using Gia_SweeperProbeDelete(): int Gia_SweeperProbeDelete( Gia_Man_t * p, int ProbeId );
  Update existing probe using Gia_SweeperProbeUpdate(): int Gia_SweeperProbeUpdate( Gia_Man_t * p, int ProbeId, int iLit );
  Comments:
      - a probe is identified by its 0-based ID, which is returned by above procedures
      - GIA literal of the probe is returned by       int Gia_SweeperProbeLit( Gia_Man_t * p, int ProbeId )
- Adding/removing conditions on the current path by calling Gia_SweeperCondPush() and Gia_SweeperCondPop()
      extern void Gia_SweeperCondPush( Gia_Man_t * p, int ProbeId );
      extern void Gia_SweeperCondPop( Gia_Man_t * p );
- Performing equivalence checking by calling     int Gia_SweeperCheckEquiv( Gia_Man_t * pGia, int Probe1, int Probe2 )
      (resource limits, such as the number of conflicts, will be controllable by dedicated GIA APIs)
- The resulting AIG to be returned to the user by calling Gia_SweeperExtractUserLogic()
      Gia_Man_t * Gia_SweeperExtractUserLogic( Gia_Man_t * p, Vec_Int_t * vProbeIds, Vec_Ptr_t * vOutNames )

*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sweeper cleanup.]

  Description [Returns new GIA with sweeper defined, which is the same
  as the original sweeper, with all the dangling logic removed and SAT
  solver restarted. The probe IDs are guaranteed to have the same logic
  functions as in the original manager.]

  SideEffects [The input manager is deleted inside this procedure.]

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Runs equivalence test for probes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs grafting from another manager.]

  Description [Returns the array of resulting literals in the destination manager.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs conditional sweeping of the cone.]

  Description [Returns the result as a new GIA manager with as many inputs
  as the original manager and as many outputs as there are logic cones.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Procedure to perform conditional fraig sweeping on separate logic cones.]

  Description [The procedure takes GIA with the sweeper defined. The sweeper
  is assumed to have some conditions currently pushed, which will be used
  as constraints for fraig sweeping. The second argument (vProbes) contains
  the array of probe IDs pointing to the user's logic cones to be SAT swept.
  Finally, the optional command line (pCommLine) is an ABC command line
  to be applied to the resulting GIA after SAT sweeping before it is
  grafted back into the original GIA manager. The return value is the status
  (success/failure) and the array of original probes possibly pointing to the
  new literals in the original GIA manager, corresponding to the user's
  logic cones after sweeping, synthesis and grafting.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
