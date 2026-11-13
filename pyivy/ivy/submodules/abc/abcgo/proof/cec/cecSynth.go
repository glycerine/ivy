package cec

/**CFile****************************************************************

  FileName    [cecSynth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Partitioned sequential synthesis.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSynth.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Populate sequential synthesis parameters.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cec_SeqSynthesisSetDefaultParams(p *Cec_ParSeq_t) {
	*p = Cec_ParSeq_t{nBTLimit: 1000, nFrames: 1, nLevelMax: -1, fConsts: 1, fEquivs: 1, nMinDomSize: 100}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cec_SeqReadMinDomSize(p *Cec_ParSeq_t) int32 {
	return p.nMinDomSize
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cec_SeqReadVerbose(p *Cec_ParSeq_t) int32 {
	return p.fVerbose
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Transfers the classes.]

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

  Synopsis    [Normalizes equivalences.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Partitioned sequential synthesis.]

  Description [Returns AIG annotated with equivalence classes.]

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
