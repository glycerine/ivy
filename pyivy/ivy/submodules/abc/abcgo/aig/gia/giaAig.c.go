package gia

import "github.com/glycerine/abcgo/aig/aig"

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

// Higher-level DAR/DCH drivers retain their original calls while their package
// interface is integrated. Their algorithms must not be replaced by no-op hooks.

func Gia_ManPerformDch(p *Gia_Man_t, parameters any) *Gia_Man_t {
	if p.pManTime != nil && p.vLevels == nil {
		Gia_ManLevelWithBoxes(p)
	}
	// The source's local fUseMapping is always zero; its dead DSM branch is omitted.
	source := Gia_ManDup(p)
	converted := Gia_ManToAig(source, 0)
	Gia_ManStop(source)
	optimized := Dar_ManChoiceNew(converted, parameters)
	out := Gia_ManFromAigChoices(optimized)
	aig.Aig_ManStop(optimized)
	Gia_ManTransferTiming(out, p)
	return out
}
func Gia_ManSeqCleanupClasses(p *Gia_Man_t, constant, equivalent, verbose int32) {
	source := Gia_ManToAigSimple(p)
	cleaned := aig.Aig_ManScl(source, constant, equivalent, 0, -1, -1, verbose, 0)
	Gia_ManReprFromAigRepr(source, p)
	aig.Aig_ManStop(cleaned)
	aig.Aig_ManStop(source)
}
