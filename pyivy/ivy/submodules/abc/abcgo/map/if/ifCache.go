package if_

import "fmt"

/**CFile****************************************************************

  FileName    [ifCache.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifCache.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManCacheRecord(p *If_Man_t, iDsd0 int32, iDsd1 int32, nShared int32, iDsd int32) {
	assert(nShared >= 0 && nShared <= p.pPars.NLutSize, "nShared >= 0 && nShared <= p.pPars.nLutSize")
	if p.vCutData == nil {
		p.vCutData = Vec_IntAlloc(10000)
	}
	if iDsd0 > iDsd1 {
		// inlining of macro ABC_SWAP
		var t int32 = iDsd0
		iDsd0 = iDsd1
		iDsd1 = t
	}
	Vec_IntPush(p.vCutData, iDsd0)
	Vec_IntPush(p.vCutData, iDsd1)
	Vec_IntPush(p.vCutData, nShared)
	Vec_IntPush(p.vCutData, iDsd)
	// printf( "%6d %6d %6d %6d\n", iDsd0, iDsd1, nShared, iDsd );
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManCacheAnalize(p *If_Man_t) {
	var vRes *Vec_Int_t
	var vTest [32]*Vec_Int_t
	var uUnique int
	vRes = Hsh_IntManHashArray(p.vCutData, 4)
	for i := 0; i <= int(p.pPars.NLutSize); i++ {
		vTest[i] = Vec_IntAlloc(1000)
	}
	// inlining of macro Vec_IntForEachEntry
	data := Vec_IntArray(vRes)
	cutData := Vec_IntArray(p.vCutData)
	for i, entry := range data {
		lutSize := cutData[4*i+2]
		Vec_IntPush(vTest[lutSize], entry)
	}
	for i := 0; i <= int(p.pPars.NLutSize); i++ {
		uUnique = Vec_IntCountUnique(vTest[i])
		entries := Vec_IntSize(vTest[i])
		fmt.Printf("%2d-var entries = %8d. (%6.2f %%)  Unique entries = %8d. (%6.2f %%)\n", i, entries, 100.0*float64(entries)/float64(max(1, Vec_IntSize(vRes))), uUnique, 100.0*float64(uUnique)/float64(max(1, entries)))
	}
	for i := 0; i <= int(p.pPars.NLutSize); i++ {
		Vec_IntFree(vTest[i])
	}
	uUnique = Vec_IntCountUnique(vRes)
	total := Vec_IntSize(p.vCutData) / 4
	fmt.Printf("Total  entries = %8d. (%6.2f %%)  Unique entries = %8d. (%6.2f %%)\n", total, 100.0, uUnique, 100.0*float64(uUnique)/float64(max(1, total)))
	Vec_IntFree(vRes)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
