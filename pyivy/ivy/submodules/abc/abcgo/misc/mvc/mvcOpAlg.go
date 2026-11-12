package mvc

import "fmt"

/**CFile****************************************************************

  FileName    [mvcOperAlg.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Miscellaneous operations on covers.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcOpAlg.c,v 1.4 2003/04/26 20:41:36 alanmi Exp $]

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

  Synopsis    [Multiplies two disjoint-support covers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAlgebraicMultiply(pCover1 *Mvc_Cover_t, pCover2 *Mvc_Cover_t) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
    var pCube1, pCube2, pCube *Mvc_Cube_t;
    var CompResult int32;

    // covers should be the same base
    assert(pCover1.nBits == pCover2.nBits, "pCover1.nBits == pCover2.nBits");
    // make sure that supports do not overlap
    Mvc_CoverAllocateMask( pCover1 );
    Mvc_CoverAllocateMask( pCover2 );
    Mvc_CoverSupport( pCover1, pCover1.pMask );
    Mvc_CoverSupport( pCover2, pCover2.pMask );
    // check if the cubes are bit-wise disjoint
    // inlining of macro Mvc_CubeBitDisjoint
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitDisjoint
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitDisjoint
    // inlining of macro Mvc_CubeNBitDisjoint
    if  (pCover1.pMask).iLast == 0     { (CompResult = ((((pCover1.pMask).pData[0]) &  ((pCover2.pMask).pData[0])) == 0 )); }    else if  (pCover1.pMask).iLast == 1{ (CompResult = (((((pCover1.pMask).pData[0]) &  ((pCover2.pMask).pData[0])) == 0 ) &&         ((((pCover1.pMask).pData[1]) &  ((pCover2.pMask).pData[1])) == 0 ))); }    else                             { {    var _i_ int32; CompResult = 1;    for _i_ = (pCover1.pMask).iLast; _i_ >= 0; _i_-- {        if ((pCover1.pMask).pData[_i_]) & ((pCover2.pMask).pData[_i_])             { CompResult = 0; break; } }}             ; };
    if !(CompResult != 0) {
        fmt.Printf("Mvc_CoverMultiply(): Cover supports are not disjoint!\n"); }

    // iterate through the cubes
    pCover = Mvc_CoverClone( pCover1 );
    // inlining of macro Mvc_CoverForEachCube
    for  pCube1 = (pCover1).lCubes.pHead;          pCube1;          pCube1 = pCube1.pNext  {
        // inlining of macro Mvc_CoverForEachCube
        for  pCube2 = (pCover2).lCubes.pHead;          (pCube2 != nil);          pCube2 = pCube2.pNext  { 
            // create the product cube
            pCube = Mvc_CubeAlloc( pCover );
            // set the product cube equal to the product of the two cubes
            // inlining of macro Mvc_CubeBitOr
            // inlining of macro Mvc_Cube1Words
            // inlining of macro Mvc_Cube1BitOr
            // inlining of macro Mvc_Cube2Words
            // inlining of macro Mvc_Cube2BitOr
            if  (pCube1).iLast == 0     { (((pCube).pData[0]) = ((pCube1).pData[0] | (pCube2).pData[0])); }    else if  (pCube1).iLast == 1{ (((pCube).pData[0]) = ((pCube1).pData[0] | (pCube2).pData[0]));
// inlining of macro Mvc_CubeNBitOr
(((pCube).pData[1]) = ((pCube1).pData[1] | (pCube2).pData[1]));; }    else                             { {    var _i_ int32;    for _i_ = (pCube1).iLast; _i_ >= 0; _i_-- {        (((pCube).pData[_i_]) = ((pCube1).pData[_i_] | (pCube2).pData[_i_])); }}; };
            // add the cube to the cover
            // inlining of macro Mvc_CoverAddCubeTail
            // inlining of macro Mvc_ListAddCubeTail
            // inlining of macro Mvc_CubeSetNext
            {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCube; }        else {            ((pList.pTail).pNext = (pCube)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCube;        ((pCube).pNext = (nil));        pList.nItems++;    };};
        }
    }
    return pCover;
}


/**Function*************************************************************

  Synopsis    [Subtracts the second cover from the first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAlgebraicSubtract(pCover1 *Mvc_Cover_t, pCover2 *Mvc_Cover_t) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
    var pCube1, pCube2, pCube *Mvc_Cube_t;
    var fFound int32;
    var CompResult int32;

    // covers should be the same base
    assert(pCover1.nBits == pCover2.nBits, "pCover1.nBits == pCover2.nBits");

    // iterate through the cubes
    pCover = Mvc_CoverClone( pCover1 );
    // inlining of macro Mvc_CoverForEachCube
    for  pCube1 = (pCover1).lCubes.pHead;          pCube1;          pCube1 = pCube1.pNext  {
        fFound = 0;
        // inlining of macro Mvc_CoverForEachCube
        for  pCube2 = (pCover2).lCubes.pHead;          (pCube2 != nil);          pCube2 = pCube2.pNext  {
            // inlining of macro Mvc_CubeBitEqual
            // inlining of macro Mvc_Cube1Words
            // inlining of macro Mvc_Cube1BitEqual
            // inlining of macro Mvc_Cube2Words
            // inlining of macro Mvc_Cube2BitEqual
            // inlining of macro Mvc_CubeNBitEqual
            if  (pCube1).iLast == 0     { (CompResult = (((pCube1).pData[0]) == ((pCube2).pData[0]))); }    else if  (pCube1).iLast == 1{ (CompResult = ((((pCube1).pData[0]) == ((pCube2).pData[0])) &&         (((pCube1).pData[1]) == ((pCube2).pData[1])))); }    else                             { {    var _i_ int32; CompResult = 1;    for _i_ = (pCube1).iLast; _i_ >= 0; _i_-- {         if ((pCube1).pData[_i_]) != ((pCube2).pData[_i_])              { CompResult = 0; break; } }}; };
            if CompResult != 0 {
                fFound = 1;
                break;
            }
        }
        if !(fFound != 0) { 
            // create the copy of the cube
            pCube = Mvc_CubeDup( pCover, pCube1 );
            // add the cube copy to the cover
            // inlining of macro Mvc_CoverAddCubeTail
            // inlining of macro Mvc_ListAddCubeTail
            // inlining of macro Mvc_CubeSetNext
            {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCube; }        else {            ((pList.pTail).pNext = (pCube)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCube;        ((pCube).pNext = (nil));        pList.nItems++;    };};
        }
    }
    return pCover;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAlgebraicEqual(pCover1 *Mvc_Cover_t, pCover2 *Mvc_Cover_t) int32 {
    var pCube1, pCube2 *Mvc_Cube_t;
    var fFound int32;
    var CompResult int32;

    // covers should be the same base
    assert(pCover1.nBits == pCover2.nBits, "pCover1.nBits == pCover2.nBits");
    // iterate through the cubes
    // inlining of macro Mvc_CoverForEachCube
    for  pCube1 = (pCover1).lCubes.pHead;          pCube1;          pCube1 = pCube1.pNext  {
        fFound = 0;
        // inlining of macro Mvc_CoverForEachCube
        for  pCube2 = (pCover2).lCubes.pHead;          (pCube2 != nil);          pCube2 = pCube2.pNext  {
            // inlining of macro Mvc_CubeBitEqual
            // inlining of macro Mvc_Cube1Words
            // inlining of macro Mvc_Cube1BitEqual
            // inlining of macro Mvc_Cube2Words
            // inlining of macro Mvc_Cube2BitEqual
            // inlining of macro Mvc_CubeNBitEqual
            if  (pCube1).iLast == 0     { (CompResult = (((pCube1).pData[0]) == ((pCube2).pData[0]))); }    else if  (pCube1).iLast == 1{ (CompResult = ((((pCube1).pData[0]) == ((pCube2).pData[0])) &&         (((pCube1).pData[1]) == ((pCube2).pData[1])))); }    else                             { {    var _i_ int32; CompResult = 1;    for _i_ = (pCube1).iLast; _i_ >= 0; _i_-- {         if ((pCube1).pData[_i_]) != ((pCube2).pData[_i_])              { CompResult = 0; break; } }}; };
            if CompResult != 0 {
                fFound = 1;
                break;
            }
        }
        if !(fFound != 0) {
            return 0; }
    }
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

