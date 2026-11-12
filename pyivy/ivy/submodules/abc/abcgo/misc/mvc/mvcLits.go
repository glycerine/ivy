package mvc

/**CFile****************************************************************

  FileName    [mvcLits.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Literal counting/updating procedures.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: mvcLits.c,v 1.4 2003/04/03 06:31:50 alanmi Exp $]

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

  Synopsis    [Find the any literal that occurs more than once.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverAnyLiteral(pCover *Mvc_Cover_t, pMask *Mvc_Cube_t) int32 {
    var pCube *Mvc_Cube_t;
    var nWord, nBit, i int32;
    var nLitsCur int32;
    var fUseFirst int32;

    // go through each literal
    if fUseFirst != 0 {
        for  i = 0; i < pCover.nBits; i++  {
            // inlining of macro Mvc_CubeBitValue
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            if !pMask || (((pMask).pData[((i) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((i) &  BITS_PER_WORD_MINUS)))) > 0) {
                // get the word and bit of this literal
                // inlining of macro Mvc_CubeWhichWord
                nWord = ((i) >> BITS_PER_WORD_LOG);
                // inlining of macro Mvc_CubeWhichBit
                nBit  = ((i) &  BITS_PER_WORD_MINUS);
                // go through all the cubes
                nLitsCur = 0;
                // inlining of macro Mvc_CoverForEachCube
                for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
                    if pCube.pData[nWord] & (1<<nBit) {
                        nLitsCur++;
                        if nLitsCur > 1 {
                            return i; }
                    }
                }
            }
        }
    
    } else {
        for  i = pCover.nBits - 1; i >=0; i--  {
            // inlining of macro Mvc_CubeBitValue
            // inlining of macro Mvc_CubeWhichWord
            // inlining of macro Mvc_CubeWhichBit
            if !pMask || (((pMask).pData[((i) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((i) &  BITS_PER_WORD_MINUS)))) > 0) {
                // get the word and bit of this literal
                // inlining of macro Mvc_CubeWhichWord
                nWord = ((i) >> BITS_PER_WORD_LOG);
                // inlining of macro Mvc_CubeWhichBit
                nBit  = ((i) &  BITS_PER_WORD_MINUS);
                // go through all the cubes
                nLitsCur = 0;
                // inlining of macro Mvc_CoverForEachCube
                for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
                    if pCube.pData[nWord] & (1<<nBit) {
                        nLitsCur++;
                        if nLitsCur > 1 {
                            return i; }
                    }
                }
            }
        }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Find the most often occurring literal.]

  Description [Find the most often occurring literal among those
  that occur more than once.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverBestLiteral(pCover *Mvc_Cover_t, pMask *Mvc_Cube_t) int32 {
    var pCube *Mvc_Cube_t;
    var nWord, nBit int32;
    var i, iMax, nLitsMax, nLitsCur int32;
    var fUseFirst int32  = 1;

    // go through each literal
    iMax = -1;
    nLitsMax = -1;
    for  i = 0; i < pCover.nBits; i++  {
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        if !pMask || (((pMask).pData[((i) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((i) &  BITS_PER_WORD_MINUS)))) > 0) {
            // get the word and bit of this literal
            // inlining of macro Mvc_CubeWhichWord
            nWord = ((i) >> BITS_PER_WORD_LOG);
            // inlining of macro Mvc_CubeWhichBit
            nBit  = ((i) &  BITS_PER_WORD_MINUS);
            // go through all the cubes
            nLitsCur = 0;
            // inlining of macro Mvc_CoverForEachCube
            for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
                if pCube.pData[nWord] & (1<<nBit) {
                    nLitsCur++; }
            }

            // check if this is the best literal
            if fUseFirst != 0 {
                if nLitsMax < nLitsCur {
                    nLitsMax = nLitsCur;
                    iMax = i;
                }
            
            } else {
                if nLitsMax <= nLitsCur {
                    nLitsMax = nLitsCur;
                    iMax = i;
                }
            }
        }
    }

    if nLitsMax > 1 {
        return iMax; }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Find the most often occurring literal.]

  Description [Find the most often occurring literal among those
  that occur more than once.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverWorstLiteral(pCover *Mvc_Cover_t, pMask *Mvc_Cube_t) int32 {
    var pCube *Mvc_Cube_t;
    var nWord, nBit int32;
    var i, iMin, nLitsMin, nLitsCur int32;
    var fUseFirst int32  = 1;

    // go through each literal
    iMin = -1;
    nLitsMin = 1000000;
    for  i = 0; i < pCover.nBits; i++  {
        // inlining of macro Mvc_CubeBitValue
        // inlining of macro Mvc_CubeWhichWord
        // inlining of macro Mvc_CubeWhichBit
        if !pMask || (((pMask).pData[((i) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((i) &  BITS_PER_WORD_MINUS)))) > 0) {
            // get the word and bit of this literal
            // inlining of macro Mvc_CubeWhichWord
            nWord = ((i) >> BITS_PER_WORD_LOG);
            // inlining of macro Mvc_CubeWhichBit
            nBit  = ((i) &  BITS_PER_WORD_MINUS);
            // go through all the cubes
            nLitsCur = 0;
            // inlining of macro Mvc_CoverForEachCube
            for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
                if pCube.pData[nWord] & (1<<nBit) {
                    nLitsCur++; }
            }

            // skip the literal that does not occur or occurs once
            if nLitsCur < 2 {
                continue; }

            // check if this is the best literal
            if fUseFirst != 0 {
                if nLitsMin > nLitsCur {
                    nLitsMin = nLitsCur;
                    iMin = i;
                }
            
            } else {
                if nLitsMin >= nLitsCur {
                    nLitsMin = nLitsCur;
                    iMin = i;
                }
            }
        }
    }

    if nLitsMin < 1000000 {
        return iMin; }
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverBestLiteralCover(pCover *Mvc_Cover_t, pSimple *Mvc_Cover_t) *Mvc_Cover_t {
    var pCoverNew *Mvc_Cover_t;
    var pCubeNew *Mvc_Cube_t;
    var pCubeS *Mvc_Cube_t;
    var iLitBest int32;

    // create the new cover
    pCoverNew = Mvc_CoverClone( pCover );
    // get the new cube
    pCubeNew = Mvc_CubeAlloc( pCoverNew );
    // clean the cube
    // inlining of macro Mvc_CubeBitClean
    // inlining of macro Mvc_Cube1Words
    // inlining of macro Mvc_Cube1BitClean
    // inlining of macro Mvc_Cube2Words
    // inlining of macro Mvc_Cube2BitClean
    if  (pCubeNew).iLast == 0      { ((pCubeNew).pData[0] = 0); }    else if  (pCubeNew).iLast == 1 { ((pCubeNew).pData[0] = 0);
// inlining of macro Mvc_CubeNBitClean
((pCubeNew).pData[1] = 0);; }    else                             { {    var _i_ int32;    for  _i_ = (pCubeNew).iLast; _i_ >= 0; _i_-- {           (pCubeNew).pData[_i_] = 0; }}; };

    // get the first cube of pSimple
    assert(Mvc_CoverReadCubeNum(pSimple) == 1, "Mvc_CoverReadCubeNum(pSimple) == 1");
    pCubeS = Mvc_CoverReadCubeHead( pSimple );
    // find the best literal among those of pCubeS
    iLitBest = Mvc_CoverBestLiteral( pCover, pCubeS );

    // insert this literal into the cube
    // inlining of macro Mvc_CubeBitInsert
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    ((pCubeNew).pData[((iLitBest) >> BITS_PER_WORD_LOG)] |=  (((Mvc_CubeWord_t)1)<<(((iLitBest) &  BITS_PER_WORD_MINUS))));
    // add the cube to the cover
    // inlining of macro Mvc_CoverAddCubeTail
    // inlining of macro Mvc_ListAddCubeTail
    // inlining of macro Mvc_CubeSetNext
    {    var pList *Mvc_List_t  = &pCoverNew.lCubes;    {        if pList.pHead == nil {            pList.pHead = pCubeNew; }        else {            ((pList.pTail).pNext = (pCubeNew)); }
        // inlining of macro Mvc_CubeSetNext
        pList.pTail    = pCubeNew;        ((pCubeNew).pNext = (nil));        pList.nItems++;    };};
    return pCoverNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverFirstCubeFirstLit(pCover *Mvc_Cover_t) int32 {
    var pCube *Mvc_Cube_t;
    var iBit, Value int32;

    // get the first cube
    pCube = Mvc_CoverReadCubeHead( pCover );
    // get the first literal
    // inlining of macro Mvc_CubeForEachBit
    // inlining of macro Mvc_CubeBitValue
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    for  iBit = 0;          iBit < pCover.nBits && ((Value = (((pCube).pData[((iBit) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iBit) &  BITS_PER_WORD_MINUS)))) > 0)), 1);          iBit++  {
        if Value != 0 {
            return iBit; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Returns the number of literals in the cover.]

  Description [Allocates storage for literal counters and fills it up
  using the current information.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverCountLiterals(pCover *Mvc_Cover_t) int32 {
    var pCube *Mvc_Cube_t;
    var nWord, nBit int32;
    var i, CounterTot, CounterCur int32;

    // allocate/clean the storage for literals
//    Mvc_CoverAllocateArrayLits( pCover );
//    memset( pCover->pLits, 0, pCover->nBits * sizeof(int) );
    // go through each literal
    CounterTot = 0;
    for  i = 0; i < pCover.nBits; i++  {
        // get the word and bit of this literal
        // inlining of macro Mvc_CubeWhichWord
        nWord = ((i) >> BITS_PER_WORD_LOG);
        // inlining of macro Mvc_CubeWhichBit
        nBit  = ((i) &  BITS_PER_WORD_MINUS);
        // go through all the cubes
        CounterCur = 0;
        // inlining of macro Mvc_CoverForEachCube
        for  pCube = (pCover).lCubes.pHead;          (pCube != nil);          pCube = pCube.pNext  {
            if pCube.pData[nWord] & (1<<nBit) {
                CounterCur++; }
        }
        CounterTot += CounterCur;
    }
    return CounterTot;
}

/**Function*************************************************************

  Synopsis    [Returns the number of literals in the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mvc_CoverIsOneLiteral(pCover *Mvc_Cover_t) int32 {
    var pCube *Mvc_Cube_t;
    var iBit, Counter, Value int32;
    if Mvc_CoverReadCubeNum(pCover) != 1 {
        return 0; }
    pCube = Mvc_CoverReadCubeHead(pCover);
    // count literals
    Counter = 0;
    // inlining of macro Mvc_CubeForEachBit
    // inlining of macro Mvc_CubeBitValue
    // inlining of macro Mvc_CubeWhichWord
    // inlining of macro Mvc_CubeWhichBit
    for  iBit = 0;          iBit < pCover.nBits && ((Value = (((pCube).pData[((iBit) >> BITS_PER_WORD_LOG)] &   (((Mvc_CubeWord_t)1)<<(((iBit) &  BITS_PER_WORD_MINUS)))) > 0)), 1);          iBit++  {
        if Value != 0 {
            if Counter++ {
                return 0; }
        }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

