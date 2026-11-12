package xsat

/**CFile****************************************************************

  FileName    [xsatMemory.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [Memory management implementation.]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include "misc/util/abc_global.h"

//#include "xsatClause.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
   
type xSAT_Mem_t struct {
    nSize   uint32;
    nCap   uint32;
    nWasted   uint32;
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
func xSAT_MemClauseHand(p *xSAT_Mem_t, h int32) *xSAT_Clause_t {
    if h != 0xFFFFFFFF {
        return ( xSAT_Clause_t * )( p.pData + h );
    } else {
        return nil;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_MemGrow(p *xSAT_Mem_t, nCap uint32) {
    var nPrevCap uint32  = p.nCap;
    if p.nCap >= nCap {
        return; }
    for p.nCap < nCap {
        var delta uint32  = ( ( p.nCap >> 1 ) + ( p.nCap >> 3 ) + 2 ) & ~1;
        p.nCap += delta;
        assert(p.nCap >= nPrevCap, "p.nCap >= nPrevCap");
    }
    assert(p.nCap > 0, "p.nCap > 0");
    // inlining of macro ABC_REALLOC
    
    if p.pData {
        p.pData = ((unsigned *) realloc((char *)(p.pData), sizeof(unsigned) * (size_t)(p.nCap)));
    } else {
        p.pData = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nCap)));
    }
}

/**Function*************************************************************

  Synopsis    [Allocating vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_MemAlloc(nCap int32) *xSAT_Mem_t {
    
    // inlining of macro ABC_CALLOC
    p := &xSAT_Mem_t{};
    if nCap <= 0 {
        nCap = 1024*1024; }

    xSAT_MemGrow(p, nCap);
    return p;
}

/**Function*************************************************************

  Synopsis    [Resetting vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_MemRestart(p *xSAT_Mem_t) {
    p.nSize = 0;
    p.nWasted = 0;
}

/**Function*************************************************************

  Synopsis    [Freeing vector.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_MemFree(p *xSAT_Mem_t) {
    // inlining of macro ABC_FREE
    ((p.pData) ? (free((char *) (p.pData)), (p.pData) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Creates new clause.]

  Description [The resulting clause is fully initialized.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_MemAppend(p *xSAT_Mem_t, nSize int32) uint32 {
    var nPrevSize uint32;
    assert(nSize > 0, "nSize > 0");
    xSAT_MemGrow( p, p.nSize + nSize );
    nPrevSize = p.nSize;
    p.nSize += nSize;
    assert(p.nSize > nPrevSize, "p.nSize > nPrevSize");
    return nPrevSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_MemCRef(p *xSAT_Mem_t, pC *uint32) uint32 {
    return uint32(  ( pC - &(p.pData[0]) ));
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_MemCap(p *xSAT_Mem_t) uint32 {
    return p.nCap;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_MemWastedCap(p *xSAT_Mem_t) uint32 {
    return p.nWasted;
}

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
