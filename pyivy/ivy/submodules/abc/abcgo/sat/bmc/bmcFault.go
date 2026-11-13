package bmc

/**CFile****************************************************************

  FileName    [bmcFault.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Checking for functional faults.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcFault.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "aig/gia/giaAig.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [This procedure sets default parameters.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_ParFfSetDefault(p *Bmc_ParFf_t) {
	*p = Bmc_ParFf_t{}
}

/**Function*************************************************************

  Synopsis    [Adds a general cardinality constraint in terms of vVars.]

  Description [Strict is "exactly K out of N". Non-strict is
  "less or equal than K out of N".]

  SideEffects []

  SeeAlso     []

***********************************************************************/
