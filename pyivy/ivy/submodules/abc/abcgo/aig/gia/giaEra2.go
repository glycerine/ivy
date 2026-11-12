package gia

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "giaAig.h"

//ABC_NAMESPACE_IMPL_START

/*
   Limitations of this package:
   - no more than (1<<31)-1 state cubes and internal nodes
   - no more than MAX_VARS_NUM state variables
   - no more than MAX_CALL_NUM transitions from a state
   - cube list rebalancing happens when cube count reaches MAX_CUBE_NUM
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
