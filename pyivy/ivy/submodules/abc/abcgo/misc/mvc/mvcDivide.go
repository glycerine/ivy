package mvc

import "fmt"

/**CFile****************************************************************

  FileName    [mvcDivide.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Procedures for algebraic division.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcDivide.c,v 1.5 2003/04/26 20:41:36 alanmi Exp $]

***********************************************************************/

//#include "mvc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Mvc_CoverVerifyDivision( Mvc_Cover_t * pCover, Mvc_Cover_t * pDiv, Mvc_Cover_t * pQuo, Mvc_Cover_t * pRem );


var s_fVerbose int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDivide(pCover *Mvc_Cover_t, pDiv *Mvc_Cover_t, ppQuo **Mvc_Cover_t, ppRem **Mvc_Cover_t) {
    // check the number of cubes
    if Mvc_CoverReadCubeNum( pCover ) < Mvc_CoverReadCubeNum( pDiv ) {
        *ppQuo = nil;
        *ppRem = nil;
        return;
    }

    // make sure that support of pCover contains that of pDiv
    if !Mvc_CoverCheckSuppContainment( pCover, pDiv ) {
        *ppQuo = nil;
        *ppRem = nil;
        return;
    }

    // perform the general division
    Mvc_CoverDivideInternal( pCover, pDiv, ppQuo, ppRem );
}


/**Function*************************************************************

  Synopsis    [Merge the cubes inside the groups.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDivideInternal(pCover *Mvc_Cover_t, pDiv *Mvc_Cover_t, ppQuo **Mvc_Cover_t, ppRem **Mvc_Cover_t) {
    var pQuo, pRem *Mvc_Cover_t;
    var pCubeC, pCubeD, pCubeCopy *Mvc_Cube_t;
    var pCube1, pCube2 *Mvc_Cube_t;
    var pGroups *int32;
    var nGroups int32;    // the cube groups
    var nCubesC, nCubesD, nMerges, iCubeC, iCubeD int32;
    var iMerge int32  = -1; // Suppress "might be used uninitialized"
    var fSkipG, GroupSize, g, c, RetValue int32;
    var nCubes int32;

    // get cover sizes
    nCubesD = Mvc_CoverReadCubeNum( pDiv );
    nCubesC = Mvc_CoverReadCubeNum( pCover );

    // check trivial cases
    if nCubesD == 1 {
        if Mvc_CoverIsOneLiteral( pDiv ) {
            Mvc_CoverDivideByLiteral( pCover, pDiv, ppQuo, ppRem );
        } else {
            Mvc_CoverDivideByCube( pCover, pDiv, ppQuo, ppRem ); }
        return;
    }

    // create the divisor and the remainder 
    pQuo = Mvc_CoverAlloc( pCover.pMem, pCover.nBits );
    pRem = Mvc_CoverAlloc( pCover.pMem, pCover.nBits );

    // get the support of the divisor
    Mvc_CoverAllocateMask( pDiv );
    Mvc_CoverSupport( pDiv, pDiv.pMask );

    // sort the cubes of the divisor
    Mvc_CoverSort( pDiv, nil, Mvc_CubeCompareInt );
    // sort the cubes of the cover
    Mvc_CoverSort( pCover, pDiv.pMask, Mvc_CubeCompareIntOutsideAndUnderMask );

    // allocate storage for cube groups
    // inlining of macro MEM_ALLOC
    // inlining of macro ABC_ALLOC
    pGroups = ((int *)((char *) malloc(sizeof(char) * (size_t)((nCubesC + 1) * sizeof(int)))));

    // mask contains variables in the support of Div
    // split the cubes into groups using the mask
    Mvc_CoverList2Array( pCover );
    Mvc_CoverList2Array( pDiv );
    pGroups[0] = 0;
    nGroups    = 1;
    for  c = 1; c < nCubesC; c++  {
        // get the cubes
        pCube1 = pCover.pCubes[c-1];
        pCube2 = pCover.pCubes[c  ];
        // compare the cubes
        // inlining of macro Mvc_CubeBitEqualOutsideMask
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitEqualOutsideMask
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitEqualOutsideMask
        // inlining of macro Mvc_CubeNBitEqualOutsideMask
        if  (pCube1).iLast == 0     { (RetValue = ((((pCube1).pData[0]) | ((pDiv.pMask).pData[0])) == (((pCube2).pData[0]) | ((pDiv.pMask).pData[0])))); }    else if  (pCube1).iLast == 1{ (RetValue = (((((pCube1).pData[0]) | ((pDiv.pMask).pData[0])) == (((pCube2).pData[0]) | ((pDiv.pMask).pData[0])))  &&         ((((pCube1).pData[1]) | ((pDiv.pMask).pData[1])) == (((pCube2).pData[1]) | ((pDiv.pMask).pData[1]))))); }    else                             { {    var _i_ int32; RetValue = 1;    for _i_ = (pCube1).iLast; _i_ >= 0; _i_-- {        if (((pCube1).pData[_i_]) | ((pDiv.pMask).pData[_i_])) != (((pCube2).pData[_i_]) | ((pDiv.pMask).pData[_i_]))            { RetValue = 0; break; } }}              ; };
        if !(RetValue != 0) {
            pGroups[nGroups++] = c; }
    }
    // finish off the last group
    pGroups[nGroups] = nCubesC;

    // consider each group separately and decide
    // whether it can produce a quotient cube
    nCubes = 0;
    for  g = 0; g < nGroups; g++  {
        // if the group has less than nCubesD cubes, 
        // there is no way it can produce the quotient cube
        // copy the cubes to the remainder
        GroupSize = pGroups[g+1] - pGroups[g];
        if GroupSize < nCubesD {
            for  c = pGroups[g]; c < pGroups[g+1]; c++  {
                pCubeCopy = Mvc_CubeDup( pRem, pCover.pCubes[c] );
                // inlining of macro Mvc_CoverAddCubeTail
                // inlining of macro Mvc_ListAddCubeTail
                // inlining of macro Mvc_CubeSetNext
                {    var pList *Mvc_List_t  = &pRem.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
                nCubes++;
            }
            continue;
        }

        // mark the cubes as those that should be added to the remainder
        for  c = pGroups[g]; c < pGroups[g+1]; c++  {
            // inlining of macro Mvc_CubeSetSize
            ((pCover.pCubes[c]).nOnes = (1));
        }

        // go through the cubes in the group and at the same time
        // go through the cubes in the divisor
        iCubeD  = 0;
        iCubeC  = 0;
        pCubeD  = pDiv.pCubes[iCubeD]; iCubeD++;
        pCubeC  = pCover.pCubes[pGroups[g]+iCubeC++];
        fSkipG  = 0;
        nMerges = 0;

        for {
            // compare the topmost cubes in F and in D
            RetValue = Mvc_CubeCompareIntUnderMask( pCubeC, pCubeD, pDiv.pMask );
            // cube are ordered in increasing order of their int value
            if RetValue == -1 { // pCubeC is above pCubeD
              // cube in C should be added to the remainder
                // check that there is enough cubes in the group
                if GroupSize - iCubeC < nCubesD - nMerges {
                    fSkipG = 1;
                    break;
                }
                // get the next cube in the cover
                pCubeC = pCover.pCubes[pGroups[g]+iCubeC++];
                continue;
            }
            if RetValue == 1 { // pCubeD is above pCubeC
             // given cube in D does not have a corresponding cube in the cover
                fSkipG = 1;
                break;
            }
            // mark the cube as the one that should NOT be added to the remainder
            // inlining of macro Mvc_CubeSetSize
            ((pCubeC).nOnes = (0));
            // remember this merged cube
            iMerge = iCubeC-1;
            nMerges++;

            // stop if we considered the last cube of the group
            if iCubeD == nCubesD {
                break; }

            // advance the cube of the divisor
            assert(iCubeD < nCubesD, "iCubeD < nCubesD");
            pCubeD = pDiv.pCubes[iCubeD]; iCubeD++;

            // advance the cube of the group
            assert(pGroups[g]+iCubeC < nCubesC, "pGroups[g]+iCubeC < nCubesC");
            pCubeC = pCover.pCubes[pGroups[g]+iCubeC++];
        }

        if fSkipG { 
            // the group has failed, add all the cubes to the remainder
            for  c = pGroups[g]; c < pGroups[g+1]; c++  {
                pCubeCopy = Mvc_CubeDup( pRem, pCover.pCubes[c] );
                // inlining of macro Mvc_CoverAddCubeTail
                // inlining of macro Mvc_ListAddCubeTail
                // inlining of macro Mvc_CubeSetNext
                {    var pList *Mvc_List_t  = &pRem.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
                nCubes++;
            }
            continue;
        }

        // the group has worked, add left-over cubes to the remainder
        for  c = pGroups[g]; c < pGroups[g+1]; c++  {
            pCubeC = pCover.pCubes[c];
            // inlining of macro Mvc_CubeReadSize
            if  (pCubeC).nOnes {
                pCubeCopy = Mvc_CubeDup( pRem, pCubeC );
                // inlining of macro Mvc_CoverAddCubeTail
                // inlining of macro Mvc_ListAddCubeTail
                // inlining of macro Mvc_CubeSetNext
                {    var pList *Mvc_List_t  = &pRem.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
                nCubes++;
            }
        }

        // create the quotient cube
        pCube1 = Mvc_CubeAlloc( pQuo );
        // inlining of macro Mvc_CubeBitSharp
        // inlining of macro Mvc_Cube1Words
        // inlining of macro Mvc_Cube1BitSharp
        // inlining of macro Mvc_Cube2Words
        // inlining of macro Mvc_Cube2BitSharp
        if  (pCover.pCubes[pGroups[g]+iMerge]).iLast == 0     { (((pCube1).pData[0]) = ((pCover.pCubes[pGroups[g]+iMerge]).pData[0] & ~((pDiv.pMask).pData[0]))); }    else if  (pCover.pCubes[pGroups[g]+iMerge]).iLast == 1{ (((pCube1).pData[0]) = ((pCover.pCubes[pGroups[g]+iMerge]).pData[0] & ~((pDiv.pMask).pData[0])));
// inlining of macro Mvc_CubeNBitSharp
(((pCube1).pData[1]) = ((pCover.pCubes[pGroups[g]+iMerge]).pData[1] & ~((pDiv.pMask).pData[1])));; }    else                             { {    var _i_ int32;    for _i_ = (pCover.pCubes[pGroups[g]+iMerge]).iLast; _i_ >= 0; _i_-- {        (((pCube1).pData[_i_]) = ((pCover.pCubes[pGroups[g]+iMerge]).pData[_i_] & ~(pDiv.pMask).pData[_i_])); }}  ; };
        // add the cube to the quotient
        // inlining of macro Mvc_CoverAddCubeTail
        // inlining of macro Mvc_ListAddCubeTail
        // inlining of macro Mvc_CubeSetNext
        {    var pList *Mvc_List_t  = &pQuo.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCube1; }        else {            ((pList.pTail).pNext = (pCube1)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCube1;        ((pCube1).pNext = (nil));        pList.nItems++;    };};
        nCubes += nCubesD;
    }
    assert(nCubes == nCubesC, "nCubes == nCubesC");

    // deallocate the memory
    // inlining of macro MEM_FREE
    // inlining of macro ABC_FREE
    if pGroups != nil { ((pGroups) ? (free((char *) (pGroups)), (pGroups) = 0) : 0); pGroups = nil; };

    // return the results
    *ppRem = pRem;
    *ppQuo = pQuo;
//    Mvc_CoverVerifyDivision( pCover, pDiv, pQuo, pRem );
}


/**Function*************************************************************

  Synopsis    [Divides the cover by a cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDivideByCube(pCover *Mvc_Cover_t, pDiv *Mvc_Cover_t, ppQuo **Mvc_Cover_t, ppRem **Mvc_Cover_t) {
    var pQuo, pRem *Mvc_Cover_t;
    var pCubeC, pCubeD, pCubeCopy *Mvc_Cube_t;
    var CompResult int32;

    // get the only cube of D
    assert(Mvc_CoverReadCubeNum(pDiv) == 1, "Mvc_CoverReadCubeNum(pDiv) == 1");

    // start the quotient and the remainder
    pQuo = Mvc_CoverAlloc( pCover.pMem, pCover.nBits );
    pRem = Mvc_CoverAlloc( pCover.pMem, pCover.nBits );

    // get the first and only cube of the divisor
    pCubeD = Mvc_CoverReadCubeHead( pDiv );

    // iterate through the cubes in the cover
    // inlining of macro Mvc_CoverForEachCube
    for  pCubeC = (pCover).lCubes.pHead;          pCubeC;          pCubeC = pCubeC.pNext  {
        // check the containment of literals from pCubeD in pCube
        // inlining of macro Mvc_Cube2BitNotImpl
        (CompResult = ((((pCubeD).pData[0]) & ~((pCubeC).pData[0])) ||         (((pCubeD).pData[1]) & ~((pCubeC).pData[1]))));
        if !(CompResult != 0) { // this cube belongs to the quotient
            // alloc the cube
            pCubeCopy = Mvc_CubeAlloc( pQuo );
            // clean the support of D
            // inlining of macro Mvc_CubeBitSharp
            // inlining of macro Mvc_Cube1Words
            // inlining of macro Mvc_Cube1BitSharp
            // inlining of macro Mvc_Cube2Words
            // inlining of macro Mvc_Cube2BitSharp
            if  (pCubeC).iLast == 0     { (((pCubeCopy).pData[0]) = ((pCubeC).pData[0] & ~((pCubeD).pData[0]))); }    else if  (pCubeC).iLast == 1{ (((pCubeCopy).pData[0]) = ((pCubeC).pData[0] & ~((pCubeD).pData[0])));
// inlining of macro Mvc_CubeNBitSharp
(((pCubeCopy).pData[1]) = ((pCubeC).pData[1] & ~((pCubeD).pData[1])));; }    else                             { {    var _i_ int32;    for _i_ = (pCubeC).iLast; _i_ >= 0; _i_-- {        (((pCubeCopy).pData[_i_]) = ((pCubeC).pData[_i_] & ~(pCubeD).pData[_i_])); }}  ; };
            // add the cube to the quotient
            // inlining of macro Mvc_CoverAddCubeTail
            // inlining of macro Mvc_ListAddCubeTail
            // inlining of macro Mvc_CubeSetNext
            {    var pList *Mvc_List_t  = &pQuo.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
        
        } else { 
            // copy the cube
            pCubeCopy = Mvc_CubeDup( pRem, pCubeC );
            // add the cube to the remainder
            // inlining of macro Mvc_CoverAddCubeTail
            // inlining of macro Mvc_ListAddCubeTail
            // inlining of macro Mvc_CubeSetNext
            {    var pList *Mvc_List_t  = &pRem.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
        }
    }
    // return the results
    *ppRem = pRem;
    *ppQuo = pQuo;
}

/**Function*************************************************************

  Synopsis    [Divides the cover by a literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDivideByLiteral(pCover *Mvc_Cover_t, pDiv *Mvc_Cover_t, ppQuo **Mvc_Cover_t, ppRem **Mvc_Cover_t) {
    var pQuo, pRem *Mvc_Cover_t;
    var pCubeC, pCubeCopy *Mvc_Cube_t;
    var iLit int32;

    // get the only cube of D
    assert(Mvc_CoverReadCubeNum(pDiv) == 1, "Mvc_CoverReadCubeNum(pDiv) == 1");

    // start the quotient and the remainder
    pQuo = Mvc_CoverAlloc( pCover.pMem, pCover.nBits );
    pRem = Mvc_CoverAlloc( pCover.pMem, pCover.nBits );

    // get the first and only literal in the divisor cube
    iLit = Mvc_CoverFirstCubeFirstLit( pDiv );

    // iterate through the cubes in the cover
    // inlining of macro Mvc_CoverForEachCube
    for  pCubeC = (pCover).lCubes.pHead;          pCubeC;          pCubeC = pCubeC.pNext  {
        // copy the cube
        pCubeCopy = Mvc_CubeDup( pCover, pCubeC );
        // add the cube to the quotient or to the remainder depending on the literal
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        if  ((pCubeCopy).pData[((iLit) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iLit) &  BITS_PER_WORD_MINUS)))) > 0 {   // remove the literal
            // inlining of macro Mvc_CubeBitRemove
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCubeCopy).pData[((iLit) >> BITS_PER_WORD_LOG)] &= ~(((Mvc_CubeWord_t)1)<<(((iLit) &  BITS_PER_WORD_MINUS))));
            // add the cube ot the quotient
            // inlining of macro Mvc_CoverAddCubeTail
            // inlining of macro Mvc_ListAddCubeTail
            // inlining of macro Mvc_CubeSetNext
            {    var pList *Mvc_List_t  = &pQuo.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
        
        } else {   // add the cube ot the remainder
            // inlining of macro Mvc_CoverAddCubeTail
            // inlining of macro Mvc_ListAddCubeTail
            // inlining of macro Mvc_CubeSetNext
            {    var pList *Mvc_List_t  = &pRem.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeCopy; }        else {            ((pList.pTail).pNext = (pCubeCopy)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeCopy;        ((pCubeCopy).pNext = (nil));        pList.nItems++;    };};
        }
    }
    // return the results
    *ppRem = pRem;
    *ppQuo = pQuo;
}


/**Function*************************************************************

  Synopsis    [Derives the quotient of division by literal.]

  Description [Reduces the cover to be the equal to the result of
  division of the given cover by the literal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverDivideByLiteralQuo(pCover *Mvc_Cover_t, iLit int32) {
    var pCube, pCube2, pPrev *Mvc_Cube_t;
    // delete those cubes that do not have this literal
    // remove this literal from other cubes
    pPrev = nil;
    // inlining of macro Mvc_CoverForEachCubeSafe
    for  pCube = (pCover).lCubes.pHead, pCube2 = (pCube? pCube.pNext: nil);          pCube;          pCube = pCube2, pCube2 = (pCube? pCube.pNext: nil)  {
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        if  (((pCube).pData[((iLit) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iLit) &  BITS_PER_WORD_MINUS)))) > 0) == 0 { // delete the cube from the cover
            // inlining of macro Mvc_CoverDeleteCube
            // inlining of macro Mvc_ListDeleteCube
            {    var pList *Mvc_List_t  = &pCover.lCubes;    {    if pPrev == nil {        pList.pHead = pCube.pNext; }    else {        pPrev.pNext = pCube.pNext; }
    if pList.pTail == pCube    {        assert(pCube.pNext == nil, "pCube.pNext == nil");        pList.pTail = pPrev;    }
    pList.nItems--;};};
            Mvc_CubeFree( pCover, pCube );
            // don't update the previous cube
        
        } else { // delete this literal from the cube
            // inlining of macro Mvc_CubeBitRemove
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            ((pCube).pData[((iLit) >> BITS_PER_WORD_LOG)] &= ~(((Mvc_CubeWord_t)1)<<(((iLit) &  BITS_PER_WORD_MINUS))));
            // update the previous cube
            pPrev = pCube;
        }
    }
}


/**Function*************************************************************

  Synopsis    [Verifies that the result of algebraic division is correct.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverVerifyDivision(pCover *Mvc_Cover_t, pDiv *Mvc_Cover_t, pQuo *Mvc_Cover_t, pRem *Mvc_Cover_t) {   
    var pProd *Mvc_Cover_t;
    var pDiff *Mvc_Cover_t;

    pProd = Mvc_CoverAlgebraicMultiply( pDiv, pQuo );
    pDiff = Mvc_CoverAlgebraicSubtract( pCover, pProd );

    if Mvc_CoverAlgebraicEqual( pDiff, pRem ) {
        fmt.Printf("Verification OKAY!\n");
    } else {
        fmt.Printf("Verification FAILED!\n");
        fmt.Printf("pCover:\n");
        Mvc_CoverPrint( pCover );
        fmt.Printf("pDiv:\n");
        Mvc_CoverPrint( pDiv );
        fmt.Printf("pRem:\n");
        Mvc_CoverPrint( pRem );
        fmt.Printf("pQuo:\n");
        Mvc_CoverPrint( pQuo );
    }

    Mvc_CoverFree( pProd );
    Mvc_CoverFree( pDiff );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

