package mvc

/**CFile****************************************************************

  FileName    [mvcContain.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Making the cover single-cube containment ABC_FREE.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcContain.c,v 1.4 2003/04/03 23:25:42 alanmi Exp $]

***********************************************************************/

//#include "mvc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Mvc_CoverRemoveDuplicates( Mvc_Cover_t * pCover );

// static void Mvc_CoverRemoveContained( Mvc_Cover_t * pCover );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Removes the contained cubes.]

  Description [Returns 1 if the cover has been changed.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverContain(pCover *Mvc_Cover_t) int32 {
    var nCubes int32;
    nCubes = Mvc_CoverReadCubeNum( pCover );
    if nCubes < 2 {
        return 0; }
    Mvc_CoverSetCubeSizes(pCover);
    Mvc_CoverSort( pCover, nil, Mvc_CubeCompareSizeAndInt );
    Mvc_CoverRemoveDuplicates( pCover );
    if nCubes > 1 {
        Mvc_CoverRemoveContained( pCover ); }
    return (nCubes != Mvc_CoverReadCubeNum(pCover));
}

/**Function*************************************************************

  Synopsis    [Removes adjacent duplicated cubes from the cube list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverRemoveDuplicates(pCover *Mvc_Cover_t) {
    var pPrev, pCube, pCube2 *Mvc_Cube_t;
    var fEqual int32;

    // set the first cube of the cover
    pPrev = Mvc_CoverReadCubeHead(pCover);
    // go through all the cubes after this one
    // inlining of macro Mvc_CoverForEachCubeStartSafe
    // inlining of macro Mvc_CubeReadNext
    for  pCube = ((pPrev).pNext), pCube2 = (pCube? pCube.pNext: nil);          pCube;          pCube = pCube2, pCube2 = (pCube? pCube.pNext: nil)  {
        // compare the current cube with the prev cube
        // inlining of macro Mvc_CubeBitEqual
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitEqual
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitEqual
        // inlining of macro Mvc_CubeNBitEqual
        if  (pPrev).iLast == 0     { (fEqual = (((pPrev).pData[0]) == ((pCube).pData[0]))); }    else if  (pPrev).iLast == 1{ (fEqual = ((((pPrev).pData[0]) == ((pCube).pData[0])) &&         (((pPrev).pData[1]) == ((pCube).pData[1])))); }    else                             { {    var _i_ int32; fEqual = 1;    for _i_ = (pPrev).iLast; _i_ >= 0; _i_-- {         if ((pPrev).pData[_i_]) != ((pCube).pData[_i_])              { fEqual = 0; break; } }}; };
        if fEqual != 0 { // they are equal - remove the current cube
            // inlining of macro Mvc_CoverDeleteCube
            // inlining of macro Mvc_ListDeleteCube
            {    var pList *Mvc_List_t  = &pCover.lCubes;    {    if pPrev == nil {        pList.pHead = pCube.pNext; }    else {        pPrev.pNext = pCube.pNext; }
    if pList.pTail == pCube    {        assert(pCube.pNext == nil, "pCube.pNext == nil");        pList.pTail = pPrev;    }
    pList.nItems--;};};
            Mvc_CubeFree( pCover, pCube );
            // don't change the previous cube cube
        
        } else { // they are not equal - update the previous cube
            pPrev = pCube;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Removes contained cubes from the sorted cube list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverRemoveContained(pCover *Mvc_Cover_t) {
    var pCubeBeg, pCubeEnd, pCubeLarge *Mvc_Cube_t;
    var pCube, pCube2, pPrev *Mvc_Cube_t;
    var sizeCur uint32;
    var Result int32;

    // since the cubes are sorted by size, it is sufficient 
    // to compare each cube with other cubes that have larger sizes
    // if the given cube implies a larger cube, the larger cube is removed
    pCubeBeg = Mvc_CoverReadCubeHead(pCover);
    for
    {
        // get the current cube size
        // inlining of macro Mvc_CubeReadSize
        sizeCur = ((pCubeBeg).nOnes);

        // initialize the end of the given size group
        pCubeEnd = pCubeBeg;
        // find the beginning of the next size group
        // inlining of macro Mvc_CoverForEachCubeStart
        // inlining of macro Mvc_CubeReadNext
        for  pCube = ((pCubeBeg).pNext);          pCube;          pCube = pCube.pNext  {
            // inlining of macro Mvc_CubeReadSize
            if sizeCur == ((pCube).nOnes) { 
                pCubeEnd = pCube;
            } else { // pCube is the first cube in the new size group
                break; }
        }
        // if we could not find the next size group
        // the containment check is finished
        if pCube == nil {
            break; }
        // otherwise, pCubeBeg/pCubeEnd are the first/last cubes of the group

        // go through all the cubes between pCubeBeg and pCubeEnd, inclusive,
        // and for each of them, try removing cubes after pCubeEnd
        // inlining of macro Mvc_CoverForEachCubeStart
        for  pCubeLarge = pCubeBeg;          (pCubeLarge != nil);          pCubeLarge = pCubeLarge.pNext  {
            pPrev = pCubeEnd;
            // inlining of macro Mvc_CoverForEachCubeStartSafe
            // inlining of macro Mvc_CubeReadNext
            for  pCube = ((pCubeEnd).pNext), pCube2 = (pCube? pCube.pNext: nil);          pCube;          pCube = pCube2, pCube2 = (pCube? pCube.pNext: nil)  {
                // check containment
                // inlining of macro Mvc_CubeBitNotImpl
                // inlining of macro Mvc_Cube1Words
                // inlining of macro Mvc_Cube1BitNotImpl
                // inlining of macro Mvc_Cube2Words
                // inlining of macro Mvc_Cube2BitNotImpl
                // inlining of macro Mvc_CubeNBitNotImpl
                if  (pCube).iLast == 0     { (Result = (((pCube).pData[0])  & ~((pCubeLarge).pData[0]))); }    else if  (pCube).iLast == 1{ (Result = ((((pCube).pData[0]) & ~((pCubeLarge).pData[0])) ||         (((pCube).pData[1]) & ~((pCubeLarge).pData[1])))); }    else                             { {    var _i_ int32; Result = 0;    for _i_ = (pCube).iLast; _i_ >= 0; _i_-- {        if ((pCube).pData[_i_]) & ~((pCubeLarge).pData[_i_])            { Result = 1; break; } }}         ; };
                if !(Result != 0) { // pCubeLarge implies pCube - remove pCube
                    // inlining of macro Mvc_CoverDeleteCube
                    // inlining of macro Mvc_ListDeleteCube
                    {    var pList *Mvc_List_t  = &pCover.lCubes;    {    if pPrev == nil {        pList.pHead = pCube.pNext; }    else {        pPrev.pNext = pCube.pNext; }
    if pList.pTail == pCube    {        assert(pCube.pNext == nil, "pCube.pNext == nil");        pList.pTail = pPrev;    }
    pList.nItems--;};};
                    Mvc_CubeFree( pCover, pCube );
                    // don't update the previous cube
                
                } else {   // update the previous cube
                    pPrev = pCube;
                }
            }
            // quit, if the main cube was the last one of this size
            if pCubeLarge == pCubeEnd {
                break; }
        }

        // set the beginning of the next group
        // inlining of macro Mvc_CubeReadNext
        pCubeBeg = ((pCubeEnd).pNext);
    
if !( pCubeBeg ) { break }
}
    
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

