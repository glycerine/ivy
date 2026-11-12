package mvc

import "fmt"

/**CFile****************************************************************

  FileName    [mvcUtils.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Various cover handling utilities.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcUtils.c,v 1.7 2003/04/26 20:41:36 alanmi Exp $]

***********************************************************************/

//#include "mvc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var bit_count [256]int32 = [256]int32{
  0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
};


// static void Mvc_CoverCopyColumn( Mvc_Cover_t * pCoverOld, Mvc_Cover_t * pCoverNew, int iColOld, int iColNew );



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSupport(pCover *Mvc_Cover_t, pSupp *Mvc_Cube_t) {
    var pCube *Mvc_Cube_t;
    // clean the support
    // inlining of macro Mvc_CubeBitClean
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitClean
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitClean
    if  (pSupp).iLast == 0      { ((pSupp).pData[0] = 0); }    else if  (pSupp).iLast == 1 { ((pSupp).pData[0] = 0);
// inlining of macro Mvc_CubeNBitClean
((pSupp).pData[1] = 0);; }    else                             { {    var _i_ int32;    for  _i_ = (pSupp).iLast; _i_ >= 0; _i_-- {           (pSupp).pData[_i_] = 0; }}; };
    // collect the support
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        // inlining of macro Mvc_CubeBitOr
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitOr
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitOr
        if  (pSupp).iLast == 0     { (((pSupp).pData[0]) = ((pSupp).pData[0] | (pCube).pData[0])); }    else if  (pSupp).iLast == 1{ (((pSupp).pData[0]) = ((pSupp).pData[0] | (pCube).pData[0]));
// inlining of macro Mvc_CubeNBitOr
(((pSupp).pData[1]) = ((pSupp).pData[1] | (pCube).pData[1]));; }    else                             { {    var _i_ int32;    for _i_ = (pSupp).iLast; _i_ >= 0; _i_-- {        (((pSupp).pData[_i_]) = ((pSupp).pData[_i_] | (pCube).pData[_i_])); }}; }
    };
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSupportAnd(pCover *Mvc_Cover_t, pSupp *Mvc_Cube_t) {
    var pCube *Mvc_Cube_t;
    // clean the support
    // inlining of macro Mvc_CubeBitFill
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitFill
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitFill
    if  (pSupp).iLast == 0      { (pSupp).pData[0] = (BITS_FULL >> (pSupp).nUnused);; }    else if  (pSupp).iLast == 1 { ((pSupp).pData[0] =  BITS_FULL);
// inlining of macro Mvc_CubeNBitFill
((pSupp).pData[1] = (BITS_FULL >> (pSupp).nUnused)); ; }    else                             { {    var _i_ int32;    (pSupp).pData[(pSupp).iLast] = (BITS_FULL >> (pSupp).nUnused);    for  _i_ = (pSupp).iLast - 1; _i_ >= 0; _i_--  {        (pSupp).pData[_i_] =  BITS_FULL; }}; };
    // collect the support
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        // inlining of macro Mvc_CubeBitAnd
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitAnd
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitAnd
        if  (pSupp).iLast == 0     { (((pSupp).pData[0]) = ((pSupp).pData[0] & (pCube).pData[0])); }    else if  (pSupp).iLast == 1{ (((pSupp).pData[0]) = ((pSupp).pData[0] & (pCube).pData[0]));
// inlining of macro Mvc_CubeNBitAnd
(((pSupp).pData[1]) = ((pSupp).pData[1] & (pCube).pData[1]));; }    else                             { {    var _i_ int32;    for _i_ = (pSupp).iLast; _i_ >= 0; _i_-- {        (((pSupp).pData[_i_]) = ((pSupp).pData[_i_] & (pCube).pData[_i_])); }} ; }
    };
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSupportSizeBinary(pCover *Mvc_Cover_t) int32 {
    var pSupp *Mvc_Cube_t;
    var Counter, i, v0, v1 int32;
    // compute the support
    pSupp = Mvc_CubeAlloc( pCover );
    Mvc_CoverSupportAnd( pCover, pSupp );
    Counter = pCover.nBits/2;
    for  i = 0; i < pCover.nBits/2; i++  {
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        v0 = (((pSupp).pData[((2*i) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((2*i) &  BITS_PER_WORD_MINUS)))) > 0);
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        v1 = (((pSupp).pData[((2*i+1) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((2*i+1) &  BITS_PER_WORD_MINUS)))) > 0);
        if v0 && (v1 != 0) {
            Counter--; }
    }
    Mvc_CubeFree( pCover, pSupp );
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSupportVarBelongs(pCover *Mvc_Cover_t, iVar int32) int32 {
    var pSupp *Mvc_Cube_t;
    var RetValue, v0, v1 int32;
    // compute the support
    pSupp = Mvc_CubeAlloc( pCover );
    Mvc_CoverSupportAnd( pCover, pSupp );
    // inlining of macro Mvc_CubeBitValue
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    v0 = (((pSupp).pData[((2*iVar) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((2*iVar) &  BITS_PER_WORD_MINUS)))) > 0);
    // inlining of macro Mvc_CubeBitValue
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    v1 = (((pSupp).pData[((2*iVar+1) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((2*iVar+1) &  BITS_PER_WORD_MINUS)))) > 0);
    RetValue = int32(( !v0 || !v1 ));
    Mvc_CubeFree( pCover, pSupp );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverCommonCube(pCover *Mvc_Cover_t, pComCube *Mvc_Cube_t) {
    var pCube *Mvc_Cube_t;
    // clean the support
    // inlining of macro Mvc_CubeBitFill
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitFill
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitFill
    if  (pComCube).iLast == 0      { (pComCube).pData[0] = (BITS_FULL >> (pComCube).nUnused);; }    else if  (pComCube).iLast == 1 { ((pComCube).pData[0] =  BITS_FULL);
// inlining of macro Mvc_CubeNBitFill
((pComCube).pData[1] = (BITS_FULL >> (pComCube).nUnused)); ; }    else                             { {    var _i_ int32;    (pComCube).pData[(pComCube).iLast] = (BITS_FULL >> (pComCube).nUnused);    for  _i_ = (pComCube).iLast - 1; _i_ >= 0; _i_--  {        (pComCube).pData[_i_] =  BITS_FULL; }}; };
    // collect the support
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        // inlining of macro Mvc_CubeBitAnd
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitAnd
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitAnd
        if  (pComCube).iLast == 0     { (((pComCube).pData[0]) = ((pComCube).pData[0] & (pCube).pData[0])); }    else if  (pComCube).iLast == 1{ (((pComCube).pData[0]) = ((pComCube).pData[0] & (pCube).pData[0]));
// inlining of macro Mvc_CubeNBitAnd
(((pComCube).pData[1]) = ((pComCube).pData[1] & (pCube).pData[1]));; }    else                             { {    var _i_ int32;    for _i_ = (pComCube).iLast; _i_ >= 0; _i_-- {        (((pComCube).pData[_i_]) = ((pComCube).pData[_i_] & (pCube).pData[_i_])); }} ; }
    };
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverIsCubeFree(pCover *Mvc_Cover_t) int32 {
    var Result int32;
    // get the common cube
    Mvc_CoverAllocateMask( pCover );
    Mvc_CoverCommonCube( pCover, pCover.pMask );
    // check whether the common cube is empty
    // inlining of macro Mvc_CubeBitEmpty
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitEmpty
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitEmpty
    // inlining of macro Mvc_CubeNBitEmpty
    if  (pCover.pMask).iLast == 0      { (Result = ((pCover.pMask).pData[0] == 0)); }    else if  (pCover.pMask).iLast == 1 { (Result = ((pCover.pMask).pData[0] == 0 && (pCover.pMask).pData[1] == 0)); }    else                             { {    var _i_ int32; Result = 1;    for _i_ = (pCover.pMask).iLast; _i_ >= 0; _i_-- {        if (pCover.pMask).pData[_i_]           { Result = 0; break; } }}; };
    return Result;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverMakeCubeFree(pCover *Mvc_Cover_t) {
    var pCube *Mvc_Cube_t;
    // get the common cube
    Mvc_CoverAllocateMask( pCover );
    Mvc_CoverCommonCube( pCover, pCover.pMask );
    // remove this cube from the cubes in the cover
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        // inlining of macro Mvc_CubeBitSharp
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitSharp
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitSharp
        if  (pCube).iLast == 0     { (((pCube).pData[0]) = ((pCube).pData[0] & ~((pCover.pMask).pData[0]))); }    else if  (pCube).iLast == 1{ (((pCube).pData[0]) = ((pCube).pData[0] & ~((pCover.pMask).pData[0])));
// inlining of macro Mvc_CubeNBitSharp
(((pCube).pData[1]) = ((pCube).pData[1] & ~((pCover.pMask).pData[1])));; }    else                             { {    var _i_ int32;    for _i_ = (pCube).iLast; _i_ >= 0; _i_-- {        (((pCube).pData[_i_]) = ((pCube).pData[_i_] & ~(pCover.pMask).pData[_i_])); }}  ; }
    };
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverCommonCubeCover(pCover *Mvc_Cover_t) *Mvc_Cover_t {
    var pRes *Mvc_Cover_t;
    var pCube *Mvc_Cube_t;
    // create the new cover
    pRes = Mvc_CoverClone( pCover );
    // get the new cube
    pCube = Mvc_CubeAlloc( pRes );
    // get the common cube
    Mvc_CoverCommonCube( pCover, pCube );
    // add the cube to the cover
    // inlining of macro Mvc_CoverAddCubeTail
    // inlining of macro Mvc_ListAddCubeTail
    // inlining of macro Mvc_CubeSetNext
    {    var pList *Mvc_List_t  = &pRes.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCube; }        else {            ((pList.pTail).pNext = (pCube)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCube;        ((pCube).pNext = (nil));        pList.nItems++;    };};
    return pRes;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if the support of cover2 is contained in the support of cover1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverCheckSuppContainment(pCover1 *Mvc_Cover_t, pCover2 *Mvc_Cover_t) int32 {
    var Result int32;
    assert(pCover1.nBits == pCover2.nBits, "pCover1.nBits == pCover2.nBits");
    // set the supports
    Mvc_CoverAllocateMask( pCover1 );
    Mvc_CoverSupport( pCover1, pCover1.pMask );
    Mvc_CoverAllocateMask( pCover2 );
    Mvc_CoverSupport( pCover2, pCover2.pMask );
    // check the containment
    // inlining of macro Mvc_CubeBitNotImpl
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitNotImpl
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitNotImpl
    // inlining of macro Mvc_CubeNBitNotImpl
    if  (pCover2.pMask).iLast == 0     { (Result = (((pCover2.pMask).pData[0])  & ~((pCover1.pMask).pData[0]))); }    else if  (pCover2.pMask).iLast == 1{ (Result = ((((pCover2.pMask).pData[0]) & ~((pCover1.pMask).pData[0])) ||         (((pCover2.pMask).pData[1]) & ~((pCover1.pMask).pData[1])))); }    else                             { {    var _i_ int32; Result = 0;    for _i_ = (pCover2.pMask).iLast; _i_ >= 0; _i_-- {        if ((pCover2.pMask).pData[_i_]) & ~((pCover1.pMask).pData[_i_])            { Result = 1; break; } }}         ; };
    return !Result;
}

/**Function*************************************************************

  Synopsis    [Counts the cube sizes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverSetCubeSizes(pCover *Mvc_Cover_t) int32 {
    var pCube *Mvc_Cube_t;
    unsigned char * pByte, * pByteStart, * pByteStop;
    var nBytes, nOnes int32;

    // get the number of unsigned chars in the cube's bit strings
    nBytes = pCover.nBits / (8 * sizeof(unsigned char)) + int32((pCover.nBits % (8 * sizeof(unsigned char)) > 0));
    // iterate through the cubes
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        // clean the counter of ones
        nOnes = 0;
        // set the starting and stopping positions
        pByteStart = (unsigned char *)pCube.pData;
        pByteStop  = pByteStart + nBytes;
        // iterate through the positions
        for  pByte = pByteStart; pByte < pByteStop; pByte++  {
            nOnes += bit_count[*pByte];
        }
        // set the nOnes
        // inlining of macro Mvc_CubeSetSize
        ((pCube).nOnes = (nOnes));
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Counts the cube sizes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverGetCubeSize(pCube *Mvc_Cube_t) int32 {
    unsigned char * pByte, * pByteStart, * pByteStop;
    var nOnes, nBytes, nBits int32;
    // get the number of unsigned chars in the cube's bit strings
    nBits = (pCube.iLast + 1) * sizeof(Mvc_CubeWord_t) * 8 - pCube.nUnused;
    nBytes = nBits / (8 * sizeof(unsigned char)) + int32((nBits % (8 * sizeof(unsigned char)) > 0));
    // clean the counter of ones
    nOnes = 0;
    // set the starting and stopping positions
    pByteStart = (unsigned char *)pCube.pData;
    pByteStop  = pByteStart + nBytes;
    // iterate through the positions
    for  pByte = pByteStart; pByte < pByteStop; pByte++  {
        nOnes += bit_count[*pByte];
    }
    return nOnes;
}

/**Function*************************************************************

  Synopsis    [Counts the differences in each cube pair in the cover.]

  Description [Takes the cover (pCover) and the array where the
  diff counters go (pDiffs). The array pDiffs should have as many
  entries as there are different pairs of cubes in the cover: n(n-1)/2.
  Fills out the array pDiffs with the following info: For each cube
  pair, included in the array is the number of literals in both cubes
  after they are made cube ABC_FREE.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Mvc_CoverCountCubePairDiffs( Mvc_Cover_t * pCover, unsigned char pDiffs[] )
{
    var pCube1 *Mvc_Cube_t;
    var pCube2 *Mvc_Cube_t;
    var pMask *Mvc_Cube_t;
    unsigned char * pByte, * pByteStart, * pByteStop;
    var nBytes, nOnes int32;
    var nCubePairs int32;

    // allocate a temporary mask
    pMask = Mvc_CubeAlloc( pCover );
    // get the number of unsigned chars in the cube's bit strings
    nBytes = pCover.nBits / (8 * sizeof(unsigned char)) + int32((pCover.nBits % (8 * sizeof(unsigned char)) > 0));
    // iterate through the cubes
    nCubePairs = 0;
    // inlining of macro Mvc_CoverForEachCube
    for  pCube1 = (pCover).lCubes.pHead;          (pCube1 != nil);          pCube1 = pCube1.pNext  {
        // inlining of macro Mvc_CoverForEachCubeStart
        // inlining of macro Mvc_CubeReadNext
        for  pCube2 = ((pCube1).pNext);          (pCube2 != nil);          pCube2 = pCube2.pNext  {
            // find the bit-wise exor of cubes
            // inlining of macro Mvc_CubeBitExor
            // inlining of macro Mvc_Cube1Words
            // inlining of macro Mvc_Cube1BitExor
            // inlining of macro Mvc_Cube2Words
            // inlining of macro Mvc_Cube2BitExor
            if  (pCube1).iLast == 0     { (((pMask).pData[0]) = ((pCube1).pData[0] ^ (pCube2).pData[0])); }    else if  (pCube1).iLast == 1{ (((pMask).pData[0]) = ((pCube1).pData[0] ^ (pCube2).pData[0]));
// inlining of macro Mvc_CubeNBitExor
(((pMask).pData[1]) = ((pCube1).pData[1] ^ (pCube2).pData[1]));; }    else                             { {    var _i_ int32;    for _i_ = (pCube1).iLast; _i_ >= 0; _i_-- {        (((pMask).pData[_i_]) = ((pCube1).pData[_i_] ^ (pCube2).pData[_i_])); }}; };
            // set the starting and stopping positions
            pByteStart = (unsigned char *)pMask.pData;
            pByteStop  = pByteStart + nBytes;
            // clean the counter of ones
            nOnes = 0;
            // iterate through the positions
            for  pByte = pByteStart; pByte < pByteStop; pByte++  {
                nOnes += bit_count[*pByte];
            }
            // set the nOnes
            pDiffs[nCubePairs++] = nOnes;
        }
    }
    // deallocate the mask
    Mvc_CubeFree( pCover, pMask );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Creates a new cover containing some literals of the old cover.]

  Description [Creates the new cover containing the given number (nVarsRem)
  literals of the old cover. All the bits of the new cover are initialized
  to "1". The selected bits from the old cover are copied on top. The numbers
  of the selected bits to copy are given in the array pVarsRem. The i-set
  entry in this array is the index of the bit in the old cover which goes
  to the i-th place in the new cover. If the i-th entry in pVarsRem is -1, 
  it means that the i-th bit does not change (remains composed of all 1's).
  This is a useful feature to speed up remapping covers, which are known
  to depend only on a subset of input variables.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverRemap(p *Mvc_Cover_t, pVarsRem *int32, nVarsRem int32) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
    var pCube, pCubeCopy *Mvc_Cube_t;
    var i int32;
    // clone the cover
    pCover = Mvc_CoverAlloc( p.pMem, nVarsRem );
    // copy the cube list
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (p).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        pCubeCopy = Mvc_CubeAlloc( pCover );
        //Mvc_CubeBitClean( pCubeCopy );   //changed by wjiang
        // inlining of macro Mvc_CubeBitFill
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitFill
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitFill
        if  (pCubeCopy).iLast == 0      { (pCubeCopy).pData[0] = (BITS_FULL >> (pCubeCopy).nUnused);; }    else if  (pCubeCopy).iLast == 1 { ((pCubeCopy).pData[0] =  BITS_FULL);
// inlining of macro Mvc_CubeNBitFill
((pCubeCopy).pData[1] = (BITS_FULL >> (pCubeCopy).nUnused)); ; }    else                             { {    var _i_ int32;    (pCubeCopy).pData[(pCubeCopy).iLast] = (BITS_FULL >> (pCubeCopy).nUnused);    for  _i_ = (pCubeCopy).iLast - 1; _i_ >= 0; _i_--  {        (pCubeCopy).pData[_i_] =  BITS_FULL; }}; };      //changed by wjiang
        // inlining of macro Mvc_CoverAddCubeTail
        // inlining of macro Mvc_ListAddCubeTail
        // inlining of macro Mvc_CubeSetNext
        {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
    }
    // copy the corresponding columns
    for  i = 0; i < nVarsRem; i++  {
        if pVarsRem[i] < 0 { 
            continue; }     //added by wjiang
        assert(pVarsRem[i] >= 0 && pVarsRem[i] < p.nBits, "pVarsRem[i] >= 0 && pVarsRem[i] < p.nBits");
        Mvc_CoverCopyColumn( p, pCover, pVarsRem[i], i );
    }
    return pCover;
}

/**Function*************************************************************

  Synopsis    [Copies a column from the old cover to the new cover.]

  Description [Copies the column (iColOld) of the old cover (pCoverOld)
  into the column (iColNew) of the new cover (pCoverNew). Assumes that 
  the number of cubes is the same in both covers. Makes no assuptions 
  about the current contents of the column in the new cover.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverCopyColumn(pCoverOld *Mvc_Cover_t, pCoverNew *Mvc_Cover_t, iColOld int32, iColNew int32) {
    var pCubeOld, pCubeNew *Mvc_Cube_t;
    var iWordOld, iWordNew, iBitOld, iBitNew int32;

    assert(Mvc_CoverReadCubeNum(pCoverOld) == Mvc_CoverReadCubeNum(pCoverNew), "Mvc_CoverReadCubeNum(pCoverOld) == Mvc_CoverReadCubeNum(pCoverNew)");

    // get the place of the old and new columns
    // inlining of macro Mvc_CubeWhichWord
    iWordOld = ((iColOld) >> BITS_PER_WORD_LOG);
    // inlining of macro Mvc_CubeWhichBit
    iBitOld  = ((iColOld) &  BITS_PER_WORD_MINUS);
    // inlining of macro Mvc_CubeWhichWord
    iWordNew = ((iColNew) >> BITS_PER_WORD_LOG);
    // inlining of macro Mvc_CubeWhichBit
    iBitNew  = ((iColNew) &  BITS_PER_WORD_MINUS);

    // go through the cubes of both covers
    pCubeNew = Mvc_CoverReadCubeHead(pCoverNew);
    // inlining of macro Mvc_CoverForEachCube
    for  pCubeOld = (pCoverOld).lCubes.pHead;          pCubeOld;          pCubeOld = pCubeOld.pNext  {
        if pCubeOld.pData[iWordOld] & (1<<iBitOld) {
            pCubeNew.pData[iWordNew] |= (1<<iBitNew);
        } else {
            pCubeNew.pData[iWordNew] &= ~(1<<iBitNew); }  // added by wjiang
        // inlining of macro Mvc_CubeReadNext
        pCubeNew = ((pCubeNew).pNext);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverInverse(pCover *Mvc_Cover_t) {
    var pCube *Mvc_Cube_t;
    // complement the cubes
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        // inlining of macro Mvc_CubeBitNot
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitNot
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitNot
        if  (pCube).iLast == 0      { ((pCube).pData[0] ^= (BITS_FULL >> (pCube).nUnused)); }    else if  (pCube).iLast == 1 { ((pCube).pData[0] ^=  BITS_FULL);
// inlining of macro Mvc_CubeNBitNot
((pCube).pData[1] ^= (BITS_FULL >> (pCube).nUnused));; }    else                             { {    var _i_ int32;    (pCube).pData[(pCube).iLast] ^= (BITS_FULL >> (pCube).nUnused);    for  _i_ = (pCube).iLast - 1; _i_ >= 0; _i_--  {        (pCube).pData[_i_] ^=  BITS_FULL; }}; }
    };
}

/**Function*************************************************************

  Synopsis    [This function dups the cover and removes DC literals from cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverRemoveDontCareLits(pCover *Mvc_Cover_t) *Mvc_Cover_t {
    var pCoverNew *Mvc_Cover_t;
    var pCube *Mvc_Cube_t;

    pCoverNew = Mvc_CoverDup( pCover );
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCoverNew).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        Mvc_CubeBitRemoveDcs( pCube );
    }
    return pCoverNew;
}

/**Function*************************************************************

  Synopsis    [Returns the cofactor w.r.t. to a binary var.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverCofactor(p *Mvc_Cover_t, iValue int32, iValueOther int32) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
    var pCube, pCubeCopy *Mvc_Cube_t;
    // clone the cover
    pCover = Mvc_CoverClone( p );
    // copy the cube list
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (p).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        if  ((pCube).pData[((iValue) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iValue) &  BITS_PER_WORD_MINUS)))) > 0 {
            pCubeCopy = Mvc_CubeDup( pCover, pCube );
            // inlining of macro Mvc_CoverAddCubeTail
            // inlining of macro Mvc_ListAddCubeTail
            // inlining of macro Mvc_CubeSetNext
            {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
            // inlining of macro Mvc_CubeBitInsert
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValueOther) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iValueOther) &  BITS_PER_WORD_MINUS))));
        }
    }
    return pCover;
}

/**Function*************************************************************

  Synopsis    [Returns the cover, in which the binary var is complemented.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverFlipVar(p *Mvc_Cover_t, iValue0 int32, iValue1 int32) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
    var pCube, pCubeCopy *Mvc_Cube_t;
    var Value0, Value1, Temp int32;

    assert(iValue0 + 1 == iValue1, "iValue0 + 1 == iValue1"); // should be adjacent

    // clone the cover
    pCover = Mvc_CoverClone( p );
    // copy the cube list
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (p).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        pCubeCopy = Mvc_CubeDup( pCover, pCube );
        // inlining of macro Mvc_CoverAddCubeTail
        // inlining of macro Mvc_ListAddCubeTail
        // inlining of macro Mvc_CubeSetNext
        {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};

        // get the bits
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        Value0 = (((pCubeCopy).pData[((iValue0) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iValue0) &  BITS_PER_WORD_MINUS)))) > 0);
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        Value1 = (((pCubeCopy).pData[((iValue1) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iValue1) &  BITS_PER_WORD_MINUS)))) > 0);

        // if both bits are one, nothing to swap
        if Value0 && Value1 {
            continue; }
        // cannot be both zero because they belong to the same var
        assert(Value0 || Value1, "Value0 || Value1"); 

        // swap the bits
        Temp   = Value0;
        Value0 = Value1;
        Value1 = Temp;

        // set the bits after the swap
        if Value0 {
            // inlining of macro Mvc_CubeBitInsert
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValue0) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iValue0) &  BITS_PER_WORD_MINUS))));
        } else {
            // inlining of macro Mvc_CubeBitRemove
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValue0) >> BITS_PER_WORD_LOG)] &= ~(((Mvc_CubeWord_t)1)<<(((iValue0) &  BITS_PER_WORD_MINUS)))); }

        if Value1 {
            // inlining of macro Mvc_CubeBitInsert
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValue1) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iValue1) &  BITS_PER_WORD_MINUS))));
        } else {
            // inlining of macro Mvc_CubeBitRemove
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValue1) >> BITS_PER_WORD_LOG)] &= ~(((Mvc_CubeWord_t)1)<<(((iValue1) &  BITS_PER_WORD_MINUS)))); }
    }
    return pCover;
}

/**Function*************************************************************

  Synopsis    [Returns the cover derived by universal quantification.]

  Description [Returns the cover computed by universal quantification
  as follows: CoverNew = Univ(B) [Cover & (A==B)]. Removes the second 
  binary var from the support (given by values iValueB0 and iValueB1). 
  Leaves the first binary variable (given by values iValueA0 and iValueA1) 
  in the support.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverUnivQuantify(p *Mvc_Cover_t, iValueA0 int32, iValueA1 int32, iValueB0 int32, iValueB1 int32) *Mvc_Cover_t {
    var pCover *Mvc_Cover_t;
    var pCube, pCubeCopy *Mvc_Cube_t;
    var ValueA0, ValueA1, ValueB0, ValueB1 int32;

    // clone the cover
    pCover = Mvc_CoverClone( p );
    // copy the cube list
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (p).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        // get the bits
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        ValueA0 = (((pCube).pData[((iValueA0) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iValueA0) &  BITS_PER_WORD_MINUS)))) > 0);
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        ValueA1 = (((pCube).pData[((iValueA1) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iValueA1) &  BITS_PER_WORD_MINUS)))) > 0);
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        ValueB0 = (((pCube).pData[((iValueB0) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iValueB0) &  BITS_PER_WORD_MINUS)))) > 0);
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        ValueB1 = (((pCube).pData[((iValueB1) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iValueB1) &  BITS_PER_WORD_MINUS)))) > 0);

        // cannot be both zero because they belong to the same var
        assert(ValueA0 || ValueA1, "ValueA0 || ValueA1"); 
        assert(ValueB0 || (ValueB1 != 0), "ValueB0 || ValueB1"); 

        // if the values of this var are different, do not add the cube
        if ValueA0 != ValueB0 && ValueA1 != ValueB1 {
            continue; }

        // create the cube
        pCubeCopy = Mvc_CubeDup( pCover, pCube );
        // inlining of macro Mvc_CoverAddCubeTail
        // inlining of macro Mvc_ListAddCubeTail
        // inlining of macro Mvc_CubeSetNext
        {    var pList *Mvc_List_t  = &pCover.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};

        // insert 1's into for the first var, if both have this value
        if ValueA0 && ValueB0 {
            // inlining of macro Mvc_CubeBitInsert
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValueA0) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iValueA0) &  BITS_PER_WORD_MINUS))));
        } else {
            // inlining of macro Mvc_CubeBitRemove
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValueA0) >> BITS_PER_WORD_LOG)] &= ~(((Mvc_CubeWord_t)1)<<(((iValueA0) &  BITS_PER_WORD_MINUS)))); }

        if ValueA1 && (ValueB1 != 0) {
            // inlining of macro Mvc_CubeBitInsert
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValueA1) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iValueA1) &  BITS_PER_WORD_MINUS))));
        } else {
            // inlining of macro Mvc_CubeBitRemove
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iValueA1) >> BITS_PER_WORD_LOG)] &= ~(((Mvc_CubeWord_t)1)<<(((iValueA1) &  BITS_PER_WORD_MINUS)))); }
            
        // insert 1's into for the second var (the cover does not depend on it)
        // inlining of macro Mvc_CubeBitInsert
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        ((pCubeCopy).pData[((iValueB0) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iValueB0) &  BITS_PER_WORD_MINUS))));
        // inlining of macro Mvc_CubeBitInsert
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        ((pCubeCopy).pData[((iValueB1) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iValueB1) &  BITS_PER_WORD_MINUS))));
    }
    return pCover;
}

     

                                                                        

                                                   

                                                                     
                                       
               
                

                

                                                                        
                                                                                       
 
                          
                                   
                                     

                                     
                                                             
                                                    
                                                     
                                    
                                             

                           
                                         
     
                                                                      
                                                                                            
                  
         
                                                  
                                                              
                     
         

                                               
                                       
                                                             
             
                                                      
                                                                          
                                                            
             
     
                  
 

                                                                        

                                                                               

                                                                            
                                                                     
                                                                         
                                                      
               
                

                

                                                                        
                                                                                                 
 
                       
                                  

                
                                                               
                                         
     
                                                         
                                                                                            
                  
                     
                                                                         
                                                                   
     
                   
 

                                                                        

                                           

                                                                       
                                                                     
                                                                        
                                                                    
                 
               
                

                

                                                                        
                                                                                   
 
                            
                                   
                           

                           
                                                                                 

                                          
                                       
                                         
                                                        
                                                 

                                                       
                                                        
                                                           

                                        
                                         
     
                                              
                                     
                                                                  
                                                         
                                           
                                           
                                              
                             
                                                              
                                        
                                                    
     
                     
 

      

/**Function*************************************************************

  Synopsis    [Transposes the cube cover.]

  Description [Returns the cube cover that looks like a transposed
  matrix, compared to the matrix derived from the original cover.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverTranspose(pCover *Mvc_Cover_t) *Mvc_Cover_t {
    var pRes *Mvc_Cover_t;
    var pCubeRes, pCube *Mvc_Cube_t;
    var nWord, nBit, i, iCube int32;

    pRes = Mvc_CoverAlloc( pCover.pMem, Mvc_CoverReadCubeNum(pCover) );
    for  i = 0; i < pCover.nBits; i++  {
        // get the word and bit of this literal
        // inlining of macro Mvc_CubeWhichWord
        nWord = ((i) >> BITS_PER_WORD_LOG);
        // inlining of macro Mvc_CubeWhichBit
        nBit  = ((i) &  BITS_PER_WORD_MINUS);
        // get the transposed cube
        pCubeRes = Mvc_CubeAlloc( pRes );
        // inlining of macro Mvc_CubeBitClean
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitClean
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitClean
        if  (pCubeRes).iLast == 0      { ((pCubeRes).pData[0] = 0); }    else if  (pCubeRes).iLast == 1 { ((pCubeRes).pData[0] = 0);
// inlining of macro Mvc_CubeNBitClean
((pCubeRes).pData[1] = 0);; }    else                             { {    var _i_ int32;    for  _i_ = (pCubeRes).iLast; _i_ >= 0; _i_-- {           (pCubeRes).pData[_i_] = 0; }}; };
        iCube = 0;
        // inlining of macro Mvc_CoverForEachCube
        for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
            if pCube.pData[nWord] & (1<<nBit) {
                // inlining of macro Mvc_CubeBitInsert
                // inlining of macro Mvc_CubeWhichWord
                // inlining of macro Mvc_CubeWhichBit
                ((pCubeRes).pData[((iCube) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iCube) &  BITS_PER_WORD_MINUS)))); }
            iCube++;
        }
        // inlining of macro Mvc_CoverAddCubeTail
        // inlining of macro Mvc_ListAddCubeTail
        // inlining of macro Mvc_CubeSetNext
        {    var pList *Mvc_List_t  = &pRes.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeRes; }        else {            ((pList.pTail).pNext = (pCubeRes)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeRes;        ((pCubeRes).pNext = (nil));        pList.nItems++;    };}; 
    }
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Checks that the cubes of the cover have 0's in unused bits.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_UtilsCheckUnusedZeros(pCover *Mvc_Cover_t) int32 {
    var Unsigned uint32;
    var pCube *Mvc_Cube_t;
    var nCubes int32;

    nCubes = 0;
    // inlining of macro Mvc_CoverForEachCube
    for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
        if pCube.nUnused == 0 {
            continue; }

        Unsigned = ( pCube.pData[pCube.iLast] & 
                    (BITS_FULL << (32-pCube.nUnused)) );
        if Unsigned != 0 {
            fmt.Printf("Cube %2d out of %2d contains dirty bits.\n", nCubes, Mvc_CoverReadCubeNum(pCover));
        }
        nCubes++;
    }
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

