package gia

/**CFile****************************************************************

  FileName    [giaAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Transformation between AIG manager.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "giaAig.h"
//#include "proof/fra/fra.h"
//#include "proof/dch/dch.h"
//#include "opt/dar/dar.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks integrity of choice nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Handles choices as additional combinational outputs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
*Function*************************************************************

	Synopsis    [Duplicates AIG in the DFS order.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_ManCofactorAig(p *Aig_Man_t, nFrames, nCofFanLit int32) *Aig_Man_t {
	source := Gia_ManFromAig(p)
	cofactored := Gia_ManUnrollAndCofactor(source, nFrames, nCofFanLit, 1)
	Gia_ManStop(source)
	out := Gia_ManToAig(cofactored, 0)
	Gia_ManStop(cofactored)
	return out
}
