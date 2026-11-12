package xsat

/**CFile****************************************************************

  FileName    [xsatBQueue.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [Bounded queue implementation.]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
   
type xSAT_BQueue_t struct {
    nSize int32;
    nCap int32;
    iFirst int32;
    iEmpty int32;
    nSum word;
    pData  *uint32;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_BQueueNew(nCap int32) *xSAT_BQueue_t {
    // inlining of macro ABC_CALLOC
    var p *xSAT_BQueue_t  = &xSAT_BQueue_t{};
    p.nCap = nCap;
    // inlining of macro ABC_CALLOC
    p.pData = ((unsigned *) calloc((size_t)(nCap), sizeof(unsigned)));
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_BQueueFree(p *xSAT_BQueue_t) {
    // inlining of macro ABC_FREE
    ((p.pData) ? (free((char *) (p.pData)), (p.pData) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_BQueuePush(p *xSAT_BQueue_t, Value uint32) {
    if p.nSize == p.nCap {
        assert(p.iFirst == p.iEmpty, "p.iFirst == p.iEmpty");
        p.nSum -= p.pData[p.iFirst];
        p.iFirst = ( p.iFirst + 1 ) % p.nCap;
    
    } else {
        p.nSize++; }

    p.nSum += Value;
    p.pData[p.iEmpty] = Value;
    if ( ++p.iEmpty ) == p.nCap {
        p.iEmpty = 0;
        p.iFirst = 0;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_BQueuePop(p *xSAT_BQueue_t) int32 {
    var RetValue int32;
    assert(p.nSize >= 1, "p.nSize >= 1");
    RetValue = p.pData[p.iFirst];
    p.nSum -= RetValue;
    p.iFirst = ( p.iFirst + 1 ) % p.nCap;
    p.nSize--;
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_BQueueAvg(p *xSAT_BQueue_t) uint32 {
    return uint32(  ( p.nSum / ( ( word ) p.nSize ) ));
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_BQueueIsValid(p *xSAT_BQueue_t) int32 {
    return ( p.nCap == p.nSize );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_BQueueClean(p *xSAT_BQueue_t) {
    p.iFirst = 0;
    p.iEmpty = 0;
    p.nSize = 0;
    p.nSum = 0;
}

//ABC_NAMESPACE_HEADER_END


