package mvc

/**CFile****************************************************************

  FileName    [mvcCube.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Manipulating unate cubes.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcCube.c,v 1.4 2003/04/03 06:31:49 alanmi Exp $]

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
func Mvc_CubeAlloc(pCover *Mvc_Cover_t) *Mvc_Cube_t {
    var pCube *Mvc_Cube_t;

    assert(pCover.nWords >= 0, "pCover.nWords >= 0");
    // allocate the cube
#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
    if ( pCover->nWords == 0 )
        pCube = (Mvc_Cube_t *)ABC_ALLOC( char, sizeof(Mvc_Cube_t) );
    else
        pCube = (Mvc_Cube_t *)ABC_ALLOC( char,  sizeof(Mvc_Cube_t) + sizeof(Mvc_CubeWord_t) * (pCover->nWords - 1) );
#else
    switch( pCover->nWords )
    {
    case 0:
    case 1:
        pCube = (Mvc_Cube_t *)Extra_MmFixedEntryFetch( pCover->pMem->pMan1 );
        break;
    case 2:
        pCube = (Mvc_Cube_t *)Extra_MmFixedEntryFetch( pCover->pMem->pMan2 );
        break;
    case 3:
    case 4:
        pCube = (Mvc_Cube_t *)Extra_MmFixedEntryFetch( pCover->pMem->pMan4 );
        break;
    default:
        pCube = (Mvc_Cube_t *)ABC_ALLOC( char, sizeof(Mvc_Cube_t) + sizeof(Mvc_CubeWord_t) * (pCover->nWords - 1) );
        break;
    }
#endif
    // set the parameters charactering this cube
    if pCover.nWords == 0 {
        pCube.iLast   = pCover.nWords;
    } else {
        pCube.iLast   = pCover.nWords - 1; }
    pCube.nUnused = pCover.nUnused;
    return pCube;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CubeDup(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) *Mvc_Cube_t {
    var pCubeCopy *Mvc_Cube_t;
    pCubeCopy = Mvc_CubeAlloc( pCover );
    // inlining of macro Mvc_CubeBitCopy
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitCopy
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitCopy
    if  (pCubeCopy).iLast == 0     { (((pCubeCopy).pData[0]) = ((pCube).pData[0])); }    else if  (pCubeCopy).iLast == 1{ (((pCubeCopy).pData[0]) = ((pCube).pData[0]));
// inlining of macro Mvc_CubeNBitCopy
(((pCubeCopy).pData[1])=  ((pCube).pData[1]));; }    else                             { {    var _i_ int32;    for _i_ = (pCubeCopy).iLast; _i_ >= 0; _i_-- {        ((pCubeCopy).pData[_i_]) = ((pCube).pData[_i_]); }}; };
    return pCubeCopy;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CubeFree(pCover *Mvc_Cover_t, pCube *Mvc_Cube_t) {
    if pCube == nil {
        return; }

    // verify the parameters charactering this cube
    assert(pCube.iLast == 0 || (int32(pCube.iLast)) == pCover.nWords - 1, "pCube.iLast == 0 || (int32(pCube.iLast)) == pCover.nWords - 1");
    assert((int32(pCube.nUnused)) == pCover.nUnused, "(int32(pCube.nUnused)) == pCover.nUnused");

    // deallocate the cube
#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
    ABC_FREE( pCube );
#else
    switch( pCover->nWords )
    {
    case 0:
    case 1:
        Extra_MmFixedEntryRecycle( pCover->pMem->pMan1, (char *)pCube );
        break;
    case 2:
        Extra_MmFixedEntryRecycle( pCover->pMem->pMan2, (char *)pCube );
        break;
    case 3:
    case 4:
        Extra_MmFixedEntryRecycle( pCover->pMem->pMan4, (char *)pCube );
        break;
    default:
        ABC_FREE( pCube );
        break;
    }
#endif
}


/**Function*************************************************************

  Synopsis    [Removes the don't-care variable from the cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CubeBitRemoveDcs(pCube *Mvc_Cube_t) {
    var Mask uint32;
    var i int32;
    // inlining of macro Mvc_CubeReadLast
    for  i = ((pCube).iLast); i >= 0; i--  {
        // detect those variables that are different (not DCs)
        Mask = (pCube.pData[i] ^ (pCube.pData[i] >> 1)) & BITS_DISJOINT; 
        // create the mask of all that are different
        Mask |= (Mask << 1);
        // remove other bits from the set
        pCube.pData[i] &= Mask;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

