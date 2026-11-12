package mvc

/**CFile****************************************************************

  FileName    [mvcCover.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Basic procedures to manipulate unate cube covers.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcCover.c,v 1.5 2003/04/09 18:02:05 alanmi Exp $]

***********************************************************************/

//#include "mvc.h"

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
func Mvc_CoverAlloc(pMem *Mvc_Manager_t, nBits int32) *Mvc_Cover_t {
    var p *Mvc_Cover_t;
    var nBitsInUnsigned int32;

    nBitsInUnsigned  = 8 * sizeof(Mvc_CubeWord_t);
#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
    p                = (Mvc_Cover_t *)ABC_ALLOC( char, sizeof(Mvc_Cover_t) );
#else
    p                = (Mvc_Cover_t *)Extra_MmFixedEntryFetch( pMem->pManC );
#endif
    p.pMem          = pMem;
    p.nBits         = nBits;
    p.nWords        = nBits / nBitsInUnsigned + int32((nBits % nBitsInUnsigned > 0));
    p.nUnused       = p.nWords * nBitsInUnsigned - p.nBits;
    p.lCubes.nItems = 0;
    p.lCubes.pHead  = nil;
    p.lCubes.pTail  = nil;
    p.nCubesAlloc   = 0;
    p.pCubes        = nil;
    p.pMask         = nil;
    p.pLits         = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverClone(p *Mvc_Cover_t) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
    pCover                = (Mvc_Cover_t *)ABC_ALLOC( char, sizeof(Mvc_Cover_t) );
#else
    pCover                = (Mvc_Cover_t *)Extra_MmFixedEntryFetch( p->pMem->pManC );
#endif
    pCover.pMem          = p.pMem;
    pCover.nBits         = p.nBits;
    pCover.nWords        = p.nWords;
    pCover.nUnused       = p.nUnused;
    pCover.lCubes.nItems = 0;
    pCover.lCubes.pHead  = nil;
    pCover.lCubes.pTail  = nil;
    pCover.nCubesAlloc   = 0;
    pCover.pCubes        = nil;
    pCover.pMask         = nil;
    pCover.pLits         = nil;
    return pCover;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDup(p *Mvc_Cover_t) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
    var pCube, pCubeCopy *Mvc_Cube_t;
    // clone the cover
    pCover = Mvc_CoverClone( p );
    // copy the cube list
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (p).lCubes.pHead;          pCube;          pCube = pCube.pNext  {
        pCubeCopy = Mvc_CubeDup( p, pCube );
        // inlining of macro Mvc_CoverAddCubeTail
        // inlining of macro Mvc_ListAddCubeTail
        // inlining of macro Mvc_CubeSetNext
        {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
    }
    return pCover;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverFree(p *Mvc_Cover_t) {
    var pCube, pCube2 *Mvc_Cube_t;
    // recycle cube list
    // inlining of macro Mvc_CoverForEachCubeSafe
    for  pCube = (p).lCubes.pHead, pCube2 = (pCube? pCube.pNext: nil);          pCube;          pCube = pCube2, pCube2 = (pCube? pCube.pNext: nil)  {
        Mvc_CubeFree( p, pCube );
    }
    // recycle other pointers
    Mvc_CubeFree( p, p.pMask );
    // inlining of macro MEM_FREE
    // inlining of macro ABC_FREE
    if p.pCubes { ((p.pCubes) ? (free((char *) (p.pCubes)), (p.pCubes) = 0) : 0); p.pCubes = nil; };
    // inlining of macro MEM_FREE
    // inlining of macro ABC_FREE
    if p.pLits { ((p.pLits) ? (free((char *) (p.pLits)), (p.pLits) = 0) : 0); p.pLits = nil; };

#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
    ABC_FREE( p );
#else
    Extra_MmFixedEntryRecycle( p->pMem->pManC, (char *)p );
#endif
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAllocateMask(pCover *Mvc_Cover_t) { 
    if pCover.pMask == nil {
        pCover.pMask = Mvc_CubeAlloc( pCover ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAllocateArrayLits(pCover *Mvc_Cover_t) { 
    if pCover.pLits == nil {
        // inlining of macro MEM_ALLOC
        // inlining of macro ABC_ALLOC
        pCover.pLits = ((int *)((char *) malloc(sizeof(char) * (size_t)((pCover.nBits) * sizeof(int))))); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAllocateArrayCubes(pCover *Mvc_Cover_t) { 
    if pCover.nCubesAlloc < pCover.lCubes.nItems {
        if pCover.nCubesAlloc > 0 {
            // inlining of macro MEM_FREE
            // inlining of macro ABC_FREE
            if pCover.pCubes { ((pCover.pCubes) ? (free((char *) (pCover.pCubes)), (pCover.pCubes) = 0) : 0); pCover.pCubes = nil; } };
        pCover.nCubesAlloc = pCover.lCubes.nItems;
        // inlining of macro MEM_ALLOC
        // inlining of macro ABC_ALLOC
        pCover.pCubes = ((Mvc_Cube_t * *)((char *) malloc(sizeof(char) * (size_t)((pCover.nCubesAlloc) * sizeof(Mvc_Cube_t *)))));
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDeallocateMask(pCover *Mvc_Cover_t) { 
    Mvc_CubeFree( pCover, pCover.pMask );
    pCover.pMask = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDeallocateArrayLits(pCover *Mvc_Cover_t) { 
    if pCover.pLits {
        // inlining of macro MEM_FREE
        // inlining of macro ABC_FREE
        if pCover.pLits { ((pCover.pLits) ? (free((char *) (pCover.pLits)), (pCover.pLits) = 0) : 0); pCover.pLits = nil; };
        pCover.pLits = nil;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

