package abcgo

/**CFile****************************************************************

  FileName    [.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName []

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: .c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/aig/aig.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// parameter structure 
   
type Xyz_ParTry_t struct {
    Par                int32;
}
 
// operation manager
   
type Xyz_ManTry_t struct {
    pPar      *Xyz_ParTry_t;           // parameters
    pAig         *Aig_Man_t;           // user's AIG 
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Xyz_ManTryAlloc(pAig *Aig_Man_t, pPar *Xyz_ParTry_t) *Xyz_ManTry_t {
    
    // inlining of macro ABC_CALLOC
    p := &Xyz_ManTry_t{};
    p.pAig    = pAig;
    p.pPar    = pPar;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Xyz_ManTryFree(p *Xyz_ManTry_t) {
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Xyz_ManPerform(pAig *Aig_Man_t, pPar *Xyz_ParTry_t) int32 {
    var p *Xyz_ManTry_t;
    var RetValue int32;
    p = Xyz_ManTryAlloc( pAig, pPar );
    RetValue = 1;
    Xyz_ManTryFree( p );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

