package bsat

/**CFile****************************************************************

  FileName    [satStore.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Proof recording.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: pr.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

// Stored clauses own their literal slices. Links remain explicit because proof
// and interpolation algorithms traverse the clause and watch chains directly.
type Sto_Cls_t struct {
	pNext, pNext0, pNext1 *Sto_Cls_t
	Id                    int32
	fA, fRoot, fVisit     uint32
	nLits                 int32
	pLits                 []lit
}
type Sto_Man_t struct {
	nVars, nRoots, nClauses, nClausesA int32
	pHead, pTail, pEmpty               *Sto_Cls_t
	nChunkSize, nChunkUsed             int32
	pChunkLast                         []byte
	chunks                             [][]byte
}
