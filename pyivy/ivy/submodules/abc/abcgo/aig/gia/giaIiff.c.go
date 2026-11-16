package gia

/**CFile****************************************************************

  FileName    [giaIiff.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Boolean matching.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaIiff.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// Gia_ManDeriveMatches preserves the explicit no-op stub in giaIiff.c.
// The original source returns zero without populating any outputs.
func Gia_ManDeriveMatches(pvNames **Vec_Ptr_t, pvTruths **Vec_Wrd_t, pvTt2Match4 **Vec_Int_t, pvConfigs **Vec_Int_t, pvTtMem2 []*Vec_Mem_t, pvTt2Match2 []*Vec_Int_t) int32 {
	return 0
}

// Gia_ManIiffTest likewise returns nil in the source; no matching algorithm
// or file loading is implemented here. Keep that source limitation explicit.
func Gia_ManIiffTest(pFileName string, pGia *Gia_Man_t, nLutSize int32, nNumCuts int32, fUseGates int32, fUseCells int32, fUseLuts int32, fVerbose int32) *Gia_Man_t {
	return nil
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
