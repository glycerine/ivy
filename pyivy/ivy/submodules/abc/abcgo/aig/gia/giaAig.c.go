package gia

import "fmt"

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


/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofactorAig(p *Aig_Man_t, nFrames int32, nCofFanLit int32) *Aig_Man_t {
    var pMan *Aig_Man_t;
    var pGia, pTemp *Gia_Man_t;
    pGia = Gia_ManFromAig( p );
    pGia = Gia_ManUnrollAndCofactor( pTemp = pGia, nFrames, nCofFanLit, 1 );
    Gia_ManStop( pTemp );
    pMan = Gia_ManToAig( pGia, 0 );
    Gia_ManStop( pGia );
    return pMan;
}


/**Function*************************************************************

  Synopsis    [Transfers representatives from pGia to pAig.]

  Description [Assumes that pGia was created from pAig.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Transfers representatives from pAig to pGia.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Applies DC2 to the GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCompress2(p *Gia_Man_t, fUpdateLevel int32, fVerbose int32) *Gia_Man_t {
    var pGia *Gia_Man_t;
    var pNew, pTemp *Aig_Man_t;
    if p.pManTime && p.vLevels == nil {
        Gia_ManLevelWithBoxes( p ); }
    pNew = Gia_ManToAig( p, 0 );
    pNew = Dar_ManCompress2( pTemp = pNew, 1, fUpdateLevel, 1, 0, fVerbose );
    Aig_ManStop( pTemp );
    pGia = Gia_ManFromAig( pNew );
    Aig_ManStop( pNew );
    Gia_ManTransferTiming( pGia, p );
    return pGia;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformDch(p *Gia_Man_t, pPars any) *Gia_Man_t {
    var fUseMapping int32;
    var pGia, pGia1 *Gia_Man_t;
    var pNew *Aig_Man_t;
    if p.pManTime && p.vLevels == nil {
        Gia_ManLevelWithBoxes( p ); }
    if (fUseMapping != 0) && Gia_ManHasMapping(p) {
        pGia1 = (Gia_Man_t *)Dsm_ManDeriveGia( p, 0 );
    } else {
        pGia1 = Gia_ManDup( p ); }
    pNew = Gia_ManToAig( pGia1, 0 );
    Gia_ManStop( pGia1 );
    pNew = Dar_ManChoiceNew( pNew, (Dch_Pars_t *)pPars );
//    pGia = Gia_ManFromAig( pNew );
    pGia = Gia_ManFromAigChoices( pNew );
    Aig_ManStop( pNew );
    Gia_ManTransferTiming( pGia, p );
    return pGia;
}

/**Function*************************************************************

  Synopsis    [Computes equivalences after structural sequential cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSeqCleanupClasses(p *Gia_Man_t, fConst int32, fEquiv int32, fVerbose int32) {
    var pNew, pTemp *Aig_Man_t;
    pNew  = Gia_ManToAigSimple( p );
    pTemp = Aig_ManScl( pNew, fConst, fEquiv, 0, -1, -1, fVerbose, 0 );
    Gia_ManReprFromAigRepr( pNew, p );
    Aig_ManStop( pTemp );
    Aig_ManStop( pNew );
}

/**Function*************************************************************

  Synopsis    [Solves SAT problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

