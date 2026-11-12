package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaSweep.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Sweeping of GIA manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSweep.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "giaAig.h"
//#include "proof/dch/dch.h"
//#include "misc/tim/tim.h"
//#include "proof/cec/cec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Mark GIA nodes that feed into POs.]

  Description [Returns the array of classes of remaining registers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Compute the set of CIs representing carry-outs of boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Checks integriting of complex flops and carry-chains.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes representatives in terms of the original objects.]

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

  Synopsis    [Computes equivalences for one clock domain.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSweepWithBoxesAndDomains(p *Gia_Man_t, pParsS any, fConst int32, fEquiv int32, fVerbose int32, fVerbEquivs int32) *Gia_Man_t { 
    var pClp, pNew, pTemp *Gia_Man_t;
    var nDoms int32  = Vec_IntFindMax(p.vRegClasses);
    var pReprs *int32;
    var iDom int32;
    var pFlopTypes [3]int32 ;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(p.pAigExtra != nil, "p.pAigExtra != nil");
    assert(nDoms > 1, "nDoms > 1");
    // order AIG objects
    pNew = Gia_ManDupUnnormalize( p );
    if pNew == nil {
        return nil; }
    Gia_ManTransferTiming( pNew, p );
    // iterate over domains
    for  iDom = 1; iDom <= nDoms; iDom++  {
        var nFlopsNew int32;
        var nFlops int32 = Vec_IntCountEntry(pNew.vRegClasses, iDom);
        if nFlops < 2 {
            continue; }
        // find global equivalences
        pClp = Gia_ManDupCollapse( pNew, pNew.pAigExtra, nil, 1 );
        //Gia_DumpAiger( pClp, p->pSpec, iDom, 2 );
        //Gia_ManPrintStats( pClp, NULL );
        // compute equivalences
        Gia_ManSweepComputeOneDomainEquivs( pClp, pNew.vRegClasses, iDom, pParsS, fConst, fEquiv, fVerbose );
        // transfer equivalences
        pReprs = Gia_ManFraigSelectReprs( pNew, pClp, fVerbose, pFlopTypes );
        Gia_ManStop( pClp );
        // reduce AIG
        Gia_ManTransferTiming( p, pNew );
        pNew = Gia_ManFraigReduceGia( pTemp = pNew, pReprs );
        Gia_ManTransferTiming( pNew, p );
        Gia_ManStop( pTemp );
        // inlining of macro ABC_FREE
        pReprs = nil;
        // derive new AIG
        pNew = Gia_ManDupWithBoxes( pTemp = pNew, 1 );
        Gia_ManStop( pTemp );
        // report
        nFlopsNew = Vec_IntCountEntry(pNew.vRegClasses, iDom);
        pFlopTypes[2] = nFlops - nFlopsNew - (pFlopTypes[0] + pFlopTypes[1]);
        if fVerbEquivs != 0 {
            fmt.Printf("Domain %2d : %5d -> %5d :  ", iDom, nFlops, nFlopsNew);
            fmt.Printf("EqConst =%4d.  EqFlop =%4d.  Dangling =%4d.  Unused =%4d.\n", pFlopTypes[0], pFlopTypes[1], Abc_MaxInt(0, pFlopTypes[2]), Abc_MaxInt(0, -pFlopTypes[2]));
            //Gia_ManPrintStats( pNew, NULL );
        }
    }
    // normalize the result
    pNew = Gia_ManDupNormalize( pTemp = pNew, 0 );
    Gia_ManTransferTiming( pNew, pTemp );
    Gia_ManStop( pTemp );
    // check integrity
    //Gia_ManCheckIntegrityWithBoxes( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reduces root model with scorr.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSweepWithBoxes(p *Gia_Man_t, pParsC any, pParsS any, fConst int32, fEquiv int32, fVerbose int32, fVerbEquivs int32) *Gia_Man_t { 
    var pClp, pNew, pTemp *Gia_Man_t;
    var pReprs *int32;
    var pFlopTypes [3]int32 ;
    var nFlopsNew, nFlops int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(p.pAigExtra != nil, "p.pAigExtra != nil");
    // consider seq synthesis with multiple clock domains
    if pParsC == nil && Gia_ManClockDomainNum(p) > 1 {
        return Gia_ManSweepWithBoxesAndDomains( p, pParsS, fConst, fEquiv, fVerbose, fVerbEquivs ); }
    // order AIG objects
    pNew = Gia_ManDupUnnormalize( p );
    if pNew == nil {
        return nil; }
    Gia_ManTransferTiming( pNew, p );
    nFlops = Vec_IntCountEntry(pNew.vRegClasses, 1);
    // find global equivalences
    pClp = Gia_ManDupCollapse( pNew, pNew.pAigExtra, nil, pParsC ? 0 : 1 );
    //Gia_DumpAiger( pClp, p->pSpec, 1, 1 );
    // compute equivalences
    if pParsC {
        Gia_ManFraigSweepPerform( pClp, pParsC );
    } else if pParsS {
        Cec_ManLSCorrespondenceClasses( pClp, (Cec_ParCor_t *)pParsS );
    } else { 
        Gia_ManSeqCleanupClasses( pClp, fConst, fEquiv, fVerbose ); }
    // transfer equivalences
    pReprs = Gia_ManFraigSelectReprs( pNew, pClp, fVerbose, pFlopTypes );
    Gia_ManStop( pClp );
    // reduce AIG
    Gia_ManTransferTiming( p, pNew );
    pNew = Gia_ManFraigReduceGia( pTemp = pNew, pReprs );
    Gia_ManTransferTiming( pNew, p );
    Gia_ManStop( pTemp );
    // inlining of macro ABC_FREE
    pReprs = nil;
    // derive new AIG
    pNew = Gia_ManDupWithBoxes( pTemp = pNew, pParsC ? 0 : 1 );
    Gia_ManStop( pTemp );
    // report
    nFlopsNew = Vec_IntCountEntry(pNew.vRegClasses, 1);
    pFlopTypes[2] = nFlops - nFlopsNew - (pFlopTypes[0] + pFlopTypes[1]);
    if fVerbEquivs != 0 {
        fmt.Printf("Domain %2d : %5d -> %5d :  ", 1, nFlops, nFlopsNew);
        fmt.Printf("EqConst =%4d.  EqFlop =%4d.  Dangling =%4d.  Unused =%4d.\n", pFlopTypes[0], pFlopTypes[1], Abc_MaxInt(0, pFlopTypes[2]), Abc_MaxInt(0, -pFlopTypes[2]));
    }
    // normalize the result
    pNew = Gia_ManDupNormalize( pTemp = pNew, 0 );
    Gia_ManTransferTiming( pNew, pTemp );
    Gia_ManStop( pTemp );
    // check integrity
    //Gia_ManCheckIntegrityWithBoxes( pNew );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

