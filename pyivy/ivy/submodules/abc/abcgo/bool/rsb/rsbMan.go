package rsb

/**CFile****************************************************************

  FileName    [rsbMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Truth-table based resubstitution.]

  Synopsis    [Manager maintenance.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rsbMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rsbInt.h"

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
func Rsb_ManAlloc(nLeafMax int32, nDivMax int32, nDecMax int32, fVerbose int32) *Rsb_Man_t {
    var p *Rsb_Man_t;
    assert(nLeafMax <= 20, "nLeafMax <= 20");
    assert(nDivMax  <= 200, "nDivMax  <= 200");
    p = new(Rsb_Man_t);
    p.nLeafMax   = nLeafMax;
    p.nDivMax    = nDivMax;
    p.nDecMax    = nDecMax;
    p.fVerbose   = fVerbose;
    // decomposition
    p.vCexes     = Vec_WrdAlloc( nDivMax + 150 );
    p.vDecPats   = Vec_IntAlloc( Abc_TtWordNum(nLeafMax) );
    p.vFanins    = Vec_IntAlloc( 10 );
    p.vFaninsOld = Vec_IntAlloc( 10 );
    p.vTries     = Vec_IntAlloc( 10 );
    return p;
}
func Rsb_ManFree(p *Rsb_Man_t) {
    Vec_WrdFree( p.vCexes );
    Vec_IntFree( p.vDecPats );
    Vec_IntFree( p.vFanins );
    Vec_IntFree( p.vFaninsOld );
    Vec_IntFree( p.vTries );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rsb_ManGetFanins(p *Rsb_Man_t) *Vec_Int_t {
    return p.vFanins;
}
func Rsb_ManGetFaninsOld(p *Rsb_Man_t) *Vec_Int_t {
    return p.vFaninsOld;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

